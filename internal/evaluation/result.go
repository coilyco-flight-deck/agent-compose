package evaluation

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

const (
	ResultFormatV1 = "agent-compose.evaluation-result.v1"
	ResultFormatV2 = "agent-compose.evaluation-result.v2"
	// ResultFormat remains the v1 fixture compatibility marker. New callers
	// select ResultFormatV2 and include PackDigest.
	ResultFormat = ResultFormatV1
)

type ResultProvenance struct {
	EvaluatedAt     string           `yaml:"evaluated_at"`
	SourceIssue     string           `yaml:"source_issue"`
	SourceRevision  string           `yaml:"source_revision"`
	PromptAuthor    string           `yaml:"prompt_author,omitempty"`
	Reviewer        string           `yaml:"reviewer"`
	PackDigest      string           `yaml:"pack_digest,omitempty"`
	RetryProvenance *RetryProvenance `yaml:"retry_provenance,omitempty"`
}

type RetryProvenance struct {
	Attempts []RetryAttempt `yaml:"attempts"`
}

type RetryAttempt struct {
	Case    string `yaml:"case"`
	Attempt int    `yaml:"attempt"`
	Outcome string `yaml:"outcome"`
	Reason  string `yaml:"reason"`
}

type CriterionScore struct {
	Criterion string `yaml:"criterion"`
	Score     int    `yaml:"score"`
	Evidence  string `yaml:"evidence"`
}

type ScoredCase struct {
	ID           string           `yaml:"id"`
	Model        string           `yaml:"model"`
	RawResponse  string           `yaml:"raw_response"`
	FinishReason string           `yaml:"finish_reason,omitempty"`
	Scores       []CriterionScore `yaml:"scores"`
	Total        int              `yaml:"total"`
	Passed       bool             `yaml:"passed"`
}

type ScoredResult struct {
	Format     string           `yaml:"format"`
	Role       string           `yaml:"role"`
	Seat       string           `yaml:"seat"`
	Provenance ResultProvenance `yaml:"provenance"`
	Cases      []ScoredCase     `yaml:"cases"`
}

func DecodeResult(raw []byte) (*ScoredResult, error) {
	decoder := yaml.NewDecoder(bytes.NewReader(raw))
	decoder.KnownFields(true)
	var result ScoredResult
	if err := decoder.Decode(&result); err != nil {
		return nil, fmt.Errorf("decode scored evaluation YAML: %w", err)
	}
	var trailing any
	if err := decoder.Decode(&trailing); err != io.EOF {
		return nil, fmt.Errorf("decode scored evaluation YAML: trailing content")
	}
	return &result, nil
}

func MarshalResult(result *ScoredResult, pack *Pack) ([]byte, error) {
	if result == nil {
		return nil, fmt.Errorf("scored evaluation is required")
	}
	if result.Format == "" {
		result.Format = ResultFormatV2
		digest, err := PackDigest(pack)
		if err != nil {
			return nil, err
		}
		result.Provenance.PackDigest = digest
		if result.Provenance.RetryProvenance == nil {
			result.Provenance.RetryProvenance = &RetryProvenance{
				Attempts: []RetryAttempt{},
			}
		}
	}
	if result.Format == ResultFormatV2 &&
		result.Provenance.RetryProvenance == nil {
		result.Provenance.RetryProvenance = &RetryProvenance{
			Attempts: []RetryAttempt{},
		}
	}
	if err := ValidateResult(result, pack); err != nil {
		return nil, err
	}
	return marshalYAML(result)
}

func ValidateResult(result *ScoredResult, pack *Pack) error {
	if result == nil || pack == nil {
		return fmt.Errorf("scored evaluation and pack are required")
	}
	if result.Format != ResultFormatV1 && result.Format != ResultFormatV2 {
		return fmt.Errorf("result format %q is unsupported", result.Format)
	}
	if result.Role != pack.Role || result.Seat != pack.Seat.Selector() {
		return fmt.Errorf(
			"result identity %s/%s, want %s/%s",
			result.Role,
			result.Seat,
			pack.Role,
			pack.Seat.Selector(),
		)
	}
	if _, err := time.Parse(time.DateOnly, result.Provenance.EvaluatedAt); err != nil {
		return fmt.Errorf("evaluated_at must use YYYY-MM-DD: %w", err)
	}
	if strings.TrimSpace(result.Provenance.SourceIssue) == "" ||
		strings.TrimSpace(result.Provenance.SourceRevision) == "" ||
		strings.TrimSpace(result.Provenance.Reviewer) == "" {
		return fmt.Errorf("result provenance is incomplete")
	}
	if result.Format == ResultFormatV2 {
		if strings.TrimSpace(result.Provenance.PromptAuthor) == "" {
			return fmt.Errorf("v2 result prompt author is required")
		}
		if strings.EqualFold(
			strings.TrimSpace(result.Provenance.PromptAuthor),
			strings.TrimSpace(result.Provenance.Reviewer),
		) {
			return fmt.Errorf("v2 result reviewer must be independent from the prompt author")
		}
		if !isFullGitRevision(result.Provenance.SourceRevision) {
			return fmt.Errorf("v2 result source_revision must be a full Git object id")
		}
		if result.Provenance.RetryProvenance == nil {
			return fmt.Errorf("v2 result retry provenance is required")
		}
		digest, err := PackDigest(pack)
		if err != nil {
			return err
		}
		if result.Provenance.PackDigest != digest {
			return fmt.Errorf("result pack digest %q does not match %q", result.Provenance.PackDigest, digest)
		}
	}
	expected := make(map[string]Case, len(pack.Cases))
	expectedByTier := make(map[string]map[string]bool)
	for _, evalCase := range pack.Cases {
		expected[evalCase.ID] = evalCase
		if expectedByTier[evalCase.ModelTier] == nil {
			expectedByTier[evalCase.ModelTier] = make(map[string]bool)
		}
		expectedByTier[evalCase.ModelTier][evalCase.ID] = true
	}
	if retry := result.Provenance.RetryProvenance; retry != nil {
		seenRetries := make(map[string]bool, len(retry.Attempts))
		for _, attempt := range retry.Attempts {
			if _, ok := expected[attempt.Case]; !ok {
				return fmt.Errorf("retry provenance contains unknown case %q", attempt.Case)
			}
			if attempt.Attempt < 1 ||
				strings.TrimSpace(attempt.Outcome) == "" ||
				strings.TrimSpace(attempt.Reason) == "" {
				return fmt.Errorf("retry provenance for case %q is incomplete", attempt.Case)
			}
			key := fmt.Sprintf("%s\x00%d", attempt.Case, attempt.Attempt)
			if seenRetries[key] {
				return fmt.Errorf(
					"retry provenance repeats case %q attempt %d",
					attempt.Case,
					attempt.Attempt,
				)
			}
			seenRetries[key] = true
		}
	}
	seen := make(map[string]bool, len(result.Cases))
	observed := make(map[string]map[string]map[string]bool)
	for _, scored := range result.Cases {
		evalCase, ok := expected[scored.ID]
		if !ok {
			return fmt.Errorf("result contains unknown case %q", scored.ID)
		}
		key := evalCase.ModelTier + "\x00" + scored.Model + "\x00" + scored.ID
		if seen[key] {
			return fmt.Errorf("result repeats case %q for model %q", scored.ID, scored.Model)
		}
		seen[key] = true
		if err := validateScoredCase(scored, evalCase, pack.ReviewRule); err != nil {
			return fmt.Errorf("case %q: %w", scored.ID, err)
		}
		if observed[evalCase.ModelTier] == nil {
			observed[evalCase.ModelTier] = make(map[string]map[string]bool)
		}
		if observed[evalCase.ModelTier][scored.Model] == nil {
			observed[evalCase.ModelTier][scored.Model] = make(map[string]bool)
		}
		observed[evalCase.ModelTier][scored.Model][scored.ID] = true
	}
	for tier, expectedCases := range expectedByTier {
		if len(observed[tier]) == 0 {
			if pack.modelTierDisabled(tier) {
				continue
			}
			return fmt.Errorf("result has no %s model", tier)
		}
		for model, modelCases := range observed[tier] {
			for id := range expectedCases {
				if !modelCases[id] {
					return fmt.Errorf(
						"model %q is missing %s case %q",
						model,
						tier,
						id,
					)
				}
			}
		}
	}
	return nil
}

func isFullGitRevision(revision string) bool {
	if len(revision) != 40 && len(revision) != 64 {
		return false
	}
	_, err := hex.DecodeString(revision)
	return err == nil
}

// PackDigest binds a v2 result to the complete rendered review contract.
func PackDigest(pack *Pack) (string, error) {
	if pack == nil {
		return "", fmt.Errorf("evaluation pack is required")
	}
	raw, err := json.Marshal(pack)
	if err != nil {
		return "", fmt.Errorf("marshal evaluation pack for digest: %w", err)
	}
	digest := sha256.Sum256(raw)
	return fmt.Sprintf("sha256:%x", digest), nil
}

func validateScoredCase(scored ScoredCase, evalCase Case, rule ReviewRule) error {
	if strings.TrimSpace(scored.Model) == "" {
		return fmt.Errorf("model is required")
	}
	emptyResponse := strings.TrimSpace(scored.RawResponse) == ""
	if emptyResponse {
		finishReason := strings.TrimSpace(scored.FinishReason)
		if finishReason == "" || strings.EqualFold(finishReason, "stop") {
			return fmt.Errorf(
				"empty raw response requires a non-success finish reason",
			)
		}
	}
	if len(scored.Scores) != len(evalCase.Rubric) {
		return fmt.Errorf("has %d scores, want %d", len(scored.Scores), len(evalCase.Rubric))
	}
	expected := make(map[string]bool, len(evalCase.Rubric))
	for _, criterion := range evalCase.Rubric {
		expected[criterion.ID] = true
	}
	values := make(map[string]int, len(scored.Scores))
	total := 0
	for _, score := range scored.Scores {
		if !expected[score.Criterion] {
			return fmt.Errorf("unknown criterion %q", score.Criterion)
		}
		if _, duplicate := values[score.Criterion]; duplicate {
			return fmt.Errorf("criterion %q is repeated", score.Criterion)
		}
		if score.Score < 0 || score.Score > 2 {
			return fmt.Errorf("criterion %q score %d is outside 0..2", score.Criterion, score.Score)
		}
		if strings.TrimSpace(score.Evidence) == "" {
			return fmt.Errorf("criterion %q has no evidence", score.Criterion)
		}
		values[score.Criterion] = score.Score
		total += score.Score
	}
	if emptyResponse && total != 0 {
		return fmt.Errorf("empty raw response must score zero on every criterion")
	}
	if total != scored.Total {
		return fmt.Errorf("total %d, want %d from criterion scores", scored.Total, total)
	}
	passed := total >= rule.PassingTotal
	if rule.RejectZeroScores {
		for _, value := range values {
			passed = passed && value != 0
		}
	}
	minimums := rule.RoleMinimumScores
	if evalCase.Dimension == "personality-expression" {
		minimums = rule.PersonalityMinimumScores
	}
	for criterion, minimum := range minimums {
		passed = passed && values[criterion] >= minimum
	}
	if scored.Passed != passed {
		return fmt.Errorf("passed is %t, want %t from the review rule", scored.Passed, passed)
	}
	return nil
}
