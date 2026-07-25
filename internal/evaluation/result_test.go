package evaluation

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLatestScoredResultsValidateAgainstCurrentPacks(t *testing.T) {
	t.Parallel()
	root := filepath.Join("..", "..", "evaluations", "latest")
	files, err := filepath.Glob(filepath.Join(root, "*.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	if len(files) == 0 {
		t.Fatal("evaluations/latest has no scored results")
	}
	legacy, err := filepath.Glob(filepath.Join(root, "*.json"))
	if err != nil {
		t.Fatal(err)
	}
	if len(legacy) != 0 {
		t.Fatalf("evaluations/latest retains JSON results: %v", legacy)
	}
	for _, path := range files {
		path := path
		t.Run(filepath.Base(path), func(t *testing.T) {
			t.Parallel()
			raw, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			result, err := DecodeResult(raw)
			if err != nil {
				t.Fatal(err)
			}
			pack, err := Build(result.Role, result.Seat)
			if err != nil {
				t.Fatal(err)
			}
			if err := ValidateResult(result, pack); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestMarshalResultWritesDeterministicYAML(t *testing.T) {
	t.Parallel()
	pack, err := Build("engineer", "codex")
	if err != nil {
		t.Fatal(err)
	}
	result := passingResult(pack)
	result.Cases[0].RawResponse = "first line\nsecond line"
	first, err := MarshalResult(result, pack)
	if err != nil {
		t.Fatal(err)
	}
	second, err := MarshalResult(result, pack)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(first, second) {
		t.Fatal("scored evaluation YAML is not deterministic")
	}
	if !strings.Contains(string(first), "raw_response: |-\n") {
		t.Fatalf("multiline response did not use a YAML block scalar:\n%s", first)
	}
	decoded, err := DecodeResult(first)
	if err != nil {
		t.Fatal(err)
	}
	if err := ValidateResult(decoded, pack); err != nil {
		t.Fatal(err)
	}
}

func TestValidateResultDerivesTotalsAndVerdictsFromPack(t *testing.T) {
	t.Parallel()
	pack, err := Build("engineer", "codex")
	if err != nil {
		t.Fatal(err)
	}
	result := passingResult(pack)
	if err := ValidateResult(result, pack); err != nil {
		t.Fatalf("valid scored result failed: %v", err)
	}

	result.Cases[0].Total--
	if err := ValidateResult(result, pack); err == nil {
		t.Fatal("inconsistent total passed validation")
	}
	result.Cases[0].Total++
	result.Cases[0].Passed = false
	if err := ValidateResult(result, pack); err == nil {
		t.Fatal("inconsistent verdict passed validation")
	}
}

func passingResult(pack *Pack) *ScoredResult {
	result := &ScoredResult{
		Format: ResultFormat,
		Role:   pack.Role,
		Seat:   pack.Seat.Harness,
		Provenance: ResultProvenance{
			EvaluatedAt:    "2026-01-01",
			SourceIssue:    "https://example.com/issues/1",
			SourceRevision: "example#1",
			Reviewer:       "fixture reviewer",
		},
	}
	for _, evalCase := range pack.Cases {
		scored := ScoredCase{
			ID:          evalCase.ID,
			Model:       "fixture-model",
			RawResponse: "fixture response",
			Passed:      true,
		}
		for _, criterion := range evalCase.Rubric {
			scored.Scores = append(scored.Scores, CriterionScore{
				Criterion: criterion.ID,
				Score:     2,
				Evidence:  "fixture evidence",
			})
			scored.Total += 2
		}
		result.Cases = append(result.Cases, scored)
	}
	return result
}
