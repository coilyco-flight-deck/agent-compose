// Command evaluation-record writes canonical v3 review records.
//
// It joins one driver run with its independent review and emits the same
// compact YAML the evaluation package validates, so a re-earned baseline is
// written by the owning marshaller rather than hand-assembled.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"forgejo.coilysiren.me/coilyco-flight-deck/agent-compose/internal/evaluation"
)

type driverRun struct {
	Arm            string `json:"arm"`
	Model          string `json:"model"`
	SourceRevision string `json:"source_revision"`
	Cases          []struct {
		ID             string   `json:"id"`
		Role           string   `json:"role"`
		Answer         string   `json:"answer"`
		FinishReason   string   `json:"finish_reason"`
		Succeeded      bool     `json:"succeeded"`
		ResolvedModels []string `json:"resolved_models"`
		Retries        []struct {
			Attempt int    `json:"attempt"`
			Outcome string `json:"outcome"`
			Reason  string `json:"reason"`
		} `json:"retries"`
	} `json:"cases"`
}

type reviewFile struct {
	ReviewerModel  string `json:"reviewer_model"`
	ReviewerEffort string `json:"reviewer_effort"`
	Reviews        []struct {
		ID       string `json:"id"`
		Role     string `json:"role"`
		Arm      string `json:"arm"`
		Reviewed bool   `json:"reviewed"`
		Reason   string `json:"reason"`
		Scores   []struct {
			Criterion string `json:"criterion"`
			Score     int    `json:"score"`
			Evidence  string `json:"evidence"`
		} `json:"scores"`
	} `json:"reviews"`
}

func main() {
	run := flag.String("run", "", "driver run JSON")
	review := flag.String("review", "", "independent review JSON")
	outDir := flag.String("out", "", "directory receiving the records")
	seat := flag.String("seat", "claude", "seat selector the packs were rendered for")
	issue := flag.String("issue", "", "source issue URL")
	author := flag.String("prompt-author", "agent-compose Core Roster authors", "prompt author")
	evaluatedAt := flag.String("evaluated-at", "", "YYYY-MM-DD")
	flag.Parse()

	if err := write(*run, *review, *outDir, *seat, *issue, *author, *evaluatedAt); err != nil {
		fmt.Fprintf(os.Stderr, "evaluation-record: %v\n", err)
		os.Exit(1)
	}
}

func write(runPath, reviewPath, outDir, seat, issue, author, evaluatedAt string) error {
	if runPath == "" || reviewPath == "" || outDir == "" || issue == "" || evaluatedAt == "" {
		return fmt.Errorf("--run, --review, --out, --issue, and --evaluated-at are required")
	}

	var run driverRun
	if err := readJSON(runPath, &run); err != nil {
		return err
	}
	var reviews reviewFile
	if err := readJSON(reviewPath, &reviews); err != nil {
		return err
	}

	scored := make(map[string][]struct {
		Criterion string `json:"criterion"`
		Score     int    `json:"score"`
		Evidence  string `json:"evidence"`
	}, len(reviews.Reviews))
	// One review file may cover several driver arms. Take only the entries
	// belonging to this run, so a two-arm comparison cannot cross-join.
	for _, entry := range reviews.Reviews {
		if entry.Arm != "" && entry.Arm != run.Arm {
			continue
		}
		if !entry.Reviewed {
			return fmt.Errorf("case %q was not reviewed: %s", entry.ID, entry.Reason)
		}
		scored[entry.Role+"\x00"+entry.ID] = entry.Scores
	}

	reviewer := fmt.Sprintf(
		"%s %s independent isolated reviewer",
		reviews.ReviewerModel,
		reviews.ReviewerEffort,
	)

	byRole := map[string]*evaluation.ScoredResult{}
	roles := []string{}
	for _, record := range run.Cases {
		result, ok := byRole[record.Role]
		if !ok {
			result = &evaluation.ScoredResult{
				Format: evaluation.ResultFormatV3,
				Role:   record.Role,
				Seat:   seat,
				Provenance: evaluation.ResultProvenance{
					EvaluatedAt:     evaluatedAt,
					SourceIssue:     issue,
					SourceRevision:  run.SourceRevision,
					PromptAuthor:    author,
					Reviewer:        reviewer,
					RetryProvenance: &evaluation.RetryProvenance{},
				},
			}
			byRole[record.Role] = result
			roles = append(roles, record.Role)
		}

		model := run.Model
		if len(record.ResolvedModels) == 1 {
			model = record.ResolvedModels[0]
		}
		scoredCase := evaluation.ScoredCase{
			ID:           record.ID,
			Model:        model,
			RawResponse:  record.Answer,
			FinishReason: record.FinishReason,
		}
		entries, reviewed := scored[record.Role+"\x00"+record.ID]
		// The reviewer skips cases the driver could not complete. Any other
		// gap means the join lost a case, which must not become a silent zero.
		if !reviewed && record.Succeeded {
			return fmt.Errorf("no review for succeeded case %s/%s", record.Role, record.ID)
		}
		for _, entry := range entries {
			scoredCase.Scores = append(scoredCase.Scores, evaluation.CriterionScore{
				Criterion: entry.Criterion,
				Score:     entry.Score,
				Evidence:  entry.Evidence,
			})
		}
		sort.Slice(scoredCase.Scores, func(i, j int) bool {
			return scoredCase.Scores[i].Criterion < scoredCase.Scores[j].Criterion
		})
		result.Cases = append(result.Cases, scoredCase)

		for _, retry := range record.Retries {
			// The driver leaves reason empty on the attempt that worked, and
			// provenance validation requires one, so name the fact instead.
			reason := retry.Reason
			if strings.TrimSpace(reason) == "" {
				reason = "no failure recorded"
			}
			result.Provenance.RetryProvenance.Attempts = append(
				result.Provenance.RetryProvenance.Attempts,
				evaluation.RetryAttempt{
					Case:    record.ID,
					Attempt: retry.Attempt,
					Outcome: retry.Outcome,
					Reason:  reason,
				},
			)
		}
	}

	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return err
	}
	sort.Strings(roles)
	for _, role := range roles {
		result := byRole[role]
		pack, err := evaluation.Build(role, seat)
		if err != nil {
			return fmt.Errorf("build %s pack: %w", role, err)
		}
		// MarshalResult stamps the digest only when the format arrives empty,
		// and this writer sets v3 up front, so bind the record to its pack here.
		digest, err := evaluation.PackDigest(pack)
		if err != nil {
			return fmt.Errorf("digest %s pack: %w", role, err)
		}
		result.Provenance.PackDigest = digest
		applyRule(result, pack)
		raw, err := evaluation.MarshalResult(result, pack)
		if err != nil {
			return fmt.Errorf("marshal %s record: %w", role, err)
		}
		target := filepath.Join(outDir, fmt.Sprintf("%s-%s.yaml", role, seat))
		if err := os.WriteFile(target, raw, 0o644); err != nil {
			return err
		}
		fmt.Fprintf(os.Stderr, "wrote %s\n", target)
	}
	return nil
}

// applyRule derives every total and verdict from the pack review rule so the
// reviewer supplies judgement only and never an arithmetic claim.
func applyRule(result *evaluation.ScoredResult, pack *evaluation.Pack) {
	dimensions := make(map[string]string, len(pack.Cases))
	for _, evalCase := range pack.Cases {
		dimensions[evalCase.ID] = evalCase.Dimension
	}
	rule := pack.ReviewRule
	for index := range result.Cases {
		scored := &result.Cases[index]
		total := 0
		values := make(map[string]int, len(scored.Scores))
		for _, score := range scored.Scores {
			total += score.Score
			values[score.Criterion] = score.Score
		}
		scored.Total = total
		passed := total >= rule.PassingTotal
		if rule.RejectZeroScores {
			for _, value := range values {
				passed = passed && value != 0
			}
		}
		minimums := rule.RoleMinimumScores
		if dimensions[scored.ID] == "personality-expression" {
			minimums = rule.PersonalityMinimumScores
		}
		for criterion, minimum := range minimums {
			passed = passed && values[criterion] >= minimum
		}
		scored.Passed = passed
	}
}

func readJSON(path string, target any) error {
	raw, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(raw, target); err != nil {
		return fmt.Errorf("decode %s: %w", path, err)
	}
	return nil
}
