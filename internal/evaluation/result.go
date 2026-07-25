package evaluation

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"
)

const ResultFormat = "agent-compose.evaluation-result.v1"

type ResultProvenance struct {
	EvaluatedAt    string `json:"evaluated_at"`
	SourceIssue    string `json:"source_issue"`
	SourceRevision string `json:"source_revision"`
	Reviewer       string `json:"reviewer"`
}

type CriterionScore struct {
	Criterion string `json:"criterion"`
	Score     int    `json:"score"`
	Evidence  string `json:"evidence"`
}

type ScoredCase struct {
	ID          string           `json:"id"`
	Model       string           `json:"model"`
	RawResponse string           `json:"raw_response"`
	Scores      []CriterionScore `json:"scores"`
	Total       int              `json:"total"`
	Passed      bool             `json:"passed"`
}

type ScoredResult struct {
	Format     string           `json:"format"`
	Role       string           `json:"role"`
	Seat       string           `json:"seat"`
	Provenance ResultProvenance `json:"provenance"`
	Cases      []ScoredCase     `json:"cases"`
}

func DecodeResult(raw []byte) (*ScoredResult, error) {
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	var result ScoredResult
	if err := decoder.Decode(&result); err != nil {
		return nil, fmt.Errorf("decode scored evaluation: %w", err)
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		return nil, fmt.Errorf("decode scored evaluation: trailing content")
	}
	return &result, nil
}

func ValidateResult(result *ScoredResult, pack *Pack) error {
	if result == nil || pack == nil {
		return fmt.Errorf("scored evaluation and pack are required")
	}
	if result.Format != ResultFormat {
		return fmt.Errorf("result format %q, want %q", result.Format, ResultFormat)
	}
	if result.Role != pack.Role || result.Seat != pack.Seat.Harness {
		return fmt.Errorf(
			"result identity %s/%s, want %s/%s",
			result.Role,
			result.Seat,
			pack.Role,
			pack.Seat.Harness,
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
	if len(result.Cases) != len(pack.Cases) {
		return fmt.Errorf("result has %d cases, want %d", len(result.Cases), len(pack.Cases))
	}

	expected := make(map[string]Case, len(pack.Cases))
	for _, evalCase := range pack.Cases {
		expected[evalCase.ID] = evalCase
	}
	seen := make(map[string]bool, len(result.Cases))
	for _, scored := range result.Cases {
		evalCase, ok := expected[scored.ID]
		if !ok {
			return fmt.Errorf("result contains unknown case %q", scored.ID)
		}
		if seen[scored.ID] {
			return fmt.Errorf("result repeats case %q", scored.ID)
		}
		seen[scored.ID] = true
		if err := validateScoredCase(scored, evalCase, pack.ReviewRule); err != nil {
			return fmt.Errorf("case %q: %w", scored.ID, err)
		}
	}
	return nil
}

func validateScoredCase(scored ScoredCase, evalCase Case, rule ReviewRule) error {
	if strings.TrimSpace(scored.Model) == "" || strings.TrimSpace(scored.RawResponse) == "" {
		return fmt.Errorf("model and raw response are required")
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
