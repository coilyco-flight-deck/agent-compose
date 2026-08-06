package evaluation

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestV1ScoredResultsRemainReadableHistoricalEvidence(t *testing.T) {
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
			raw, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			result, err := DecodeResult(raw)
			if err != nil {
				t.Fatal(err)
			}
			if result.Format != ResultFormatV2 {
				t.Fatalf("historical result uses %q, want %q", result.Format, ResultFormatV2)
			}
			if result.Role == "" || result.Seat == "" || len(result.Cases) == 0 {
				t.Fatalf("historical result lost identity or evidence: %+v", result)
			}
			for _, scoredCase := range result.Cases {
				if strings.TrimSpace(scoredCase.RawResponse) == "" &&
					strings.TrimSpace(scoredCase.FinishReason) == "" {
					t.Fatalf(
						"historical case %q lost response evidence",
						scoredCase.ID,
					)
				}
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

func TestMarshalResultWritesV2AndRejectsPackDrift(t *testing.T) {
	pack, err := Build("engineer", "codex")
	if err != nil {
		t.Fatal(err)
	}
	result := passingResult(pack)
	result.Format = ""
	raw, err := MarshalResult(result, pack)
	if err != nil {
		t.Fatal(err)
	}
	decoded, err := DecodeResult(raw)
	if err != nil {
		t.Fatal(err)
	}
	if decoded.Format != ResultFormatV2 ||
		!strings.HasPrefix(decoded.Provenance.PackDigest, "sha256:") ||
		decoded.Provenance.RetryProvenance == nil {
		t.Fatalf("new result omitted v2 pack provenance: %+v", decoded.Provenance)
	}
	mutated := *pack
	mutated.Briefing += "\n\nA prose-only role change."
	if err := ValidateResult(decoded, &mutated); err == nil ||
		!strings.Contains(err.Error(), "pack digest") {
		t.Fatalf("role prose drift passed v2 validation: %v", err)
	}
	mutated = *pack
	mutated.Personalities = append([]PersonalityContext(nil), pack.Personalities...)
	mutated.Personalities[0].Definition += "\nA prose-only personality change."
	if err := ValidateResult(decoded, &mutated); err == nil ||
		!strings.Contains(err.Error(), "pack digest") {
		t.Fatalf("personality prose drift passed v2 validation: %v", err)
	}
	if err := ValidateResult(passingResult(pack), pack); err != nil {
		t.Fatalf("v1 compatibility failed: %v", err)
	}
}

func TestValidateResultRequiresCompleteV2RetryProvenance(t *testing.T) {
	t.Parallel()
	pack, err := Build("engineer", "codex")
	if err != nil {
		t.Fatal(err)
	}
	result := passingResult(pack)
	result.Format = ResultFormatV2
	result.Provenance.PromptAuthor = "fixture prompt author"
	digest, err := PackDigest(pack)
	if err != nil {
		t.Fatal(err)
	}
	result.Provenance.PackDigest = digest
	if err := ValidateResult(result, pack); err == nil ||
		!strings.Contains(err.Error(), "retry provenance is required") {
		t.Fatalf("missing retry provenance error = %v", err)
	}

	result.Provenance.RetryProvenance = &RetryProvenance{
		Attempts: []RetryAttempt{{
			Case:    caseForScenarioKind(t, pack, ossTier, ScenarioMissionFit).ID,
			Attempt: 1,
			Outcome: "transport-timeout",
			Reason:  "The model endpoint returned no bytes before the deadline.",
		}},
	}
	if err := ValidateResult(result, pack); err != nil {
		t.Fatalf("complete retry provenance failed: %v", err)
	}
	result.Provenance.RetryProvenance.Attempts[0].Case = "unknown"
	if err := ValidateResult(result, pack); err == nil ||
		!strings.Contains(err.Error(), "unknown case") {
		t.Fatalf("unknown retry case error = %v", err)
	}
}

func TestValidateResultRequiresIndependentV2ReviewerAndExactRevision(t *testing.T) {
	t.Parallel()
	pack, err := Build("engineer", "codex")
	if err != nil {
		t.Fatal(err)
	}
	result := passingResult(pack)
	result.Format = ResultFormatV2
	result.Provenance.PackDigest, err = PackDigest(pack)
	if err != nil {
		t.Fatal(err)
	}
	result.Provenance.RetryProvenance = &RetryProvenance{Attempts: []RetryAttempt{}}
	result.Provenance.PromptAuthor = ""
	if err := ValidateResult(result, pack); err == nil ||
		!strings.Contains(err.Error(), "prompt author") {
		t.Fatalf("missing prompt author error = %v", err)
	}
	result.Provenance.PromptAuthor = result.Provenance.Reviewer
	if err := ValidateResult(result, pack); err == nil ||
		!strings.Contains(err.Error(), "independent") {
		t.Fatalf("same reviewer error = %v", err)
	}
	result.Provenance.PromptAuthor = "fixture prompt author"
	result.Provenance.SourceRevision = "short"
	if err := ValidateResult(result, pack); err == nil ||
		!strings.Contains(err.Error(), "full Git object id") {
		t.Fatalf("short revision error = %v", err)
	}
	result.Provenance.SourceRevision = strings.Repeat("a", 40)
	if err := ValidateResult(result, pack); err != nil {
		t.Fatalf("independent provenance failed: %v", err)
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

func TestValidateResultAcceptsExplicitEmptyFailedResponse(t *testing.T) {
	t.Parallel()
	pack, err := Build("engineer", "codex")
	if err != nil {
		t.Fatal(err)
	}
	result := passingResult(pack)
	failed := &result.Cases[0]
	failed.RawResponse = ""
	failed.FinishReason = "length"
	failed.Total = 0
	failed.Passed = false
	for index := range failed.Scores {
		failed.Scores[index].Score = 0
		failed.Scores[index].Evidence = "The model returned no content."
	}
	if err := ValidateResult(result, pack); err != nil {
		t.Fatalf("explicit empty response failure was rejected: %v", err)
	}

	failed.FinishReason = "stop"
	if err := ValidateResult(result, pack); err == nil ||
		!strings.Contains(err.Error(), "non-success finish reason") {
		t.Fatalf("successful empty response error = %v", err)
	}
	failed.FinishReason = "length"
	failed.Scores[0].Score = 1
	failed.Total = 1
	if err := ValidateResult(result, pack); err == nil ||
		!strings.Contains(err.Error(), "score zero") {
		t.Fatalf("nonzero empty response error = %v", err)
	}
}

func TestValidateResultAcceptsCompleteMultipleModelsPerTier(t *testing.T) {
	t.Parallel()
	pack, err := Build("creator", "discord")
	if err != nil {
		t.Fatal(err)
	}
	result := passingResult(pack)
	for _, scored := range result.Cases {
		if strings.HasPrefix(scored.ID, "oss-") {
			candidate := scored
			candidate.Model = "second-oss-model"
			result.Cases = append(result.Cases, candidate)
		}
	}
	if err := ValidateResult(result, pack); err != nil {
		t.Fatalf("complete second OSS model failed validation: %v", err)
	}
}

func TestValidateResultAllowsDisabledTierToBeOmitted(t *testing.T) {
	t.Parallel()
	pack, err := Build("engineer", "codex")
	if err != nil {
		t.Fatal(err)
	}
	result := passingResult(pack)
	kept := result.Cases[:0]
	for _, scored := range result.Cases {
		if !strings.HasPrefix(scored.ID, ossTier+"-") {
			kept = append(kept, scored)
		}
	}
	result.Cases = kept
	if err := ValidateResult(result, pack); err != nil {
		t.Fatalf("frontier-only result failed while OSS is disabled: %v", err)
	}

	result.Cases = result.Cases[1:]
	if err := ValidateResult(result, pack); err == nil ||
		!strings.Contains(err.Error(), "missing frontier case") {
		t.Fatalf("incomplete active tier error = %v", err)
	}
}

func TestValidateResultRejectsIncompleteOrRepeatedModelCases(t *testing.T) {
	t.Parallel()
	pack, err := Build("creator", "discord")
	if err != nil {
		t.Fatal(err)
	}
	result := passingResult(pack)
	var candidate ScoredCase
	for _, scored := range result.Cases {
		if scored.ID == caseForScenarioKind(t, pack, ossTier, ScenarioMissionFit).ID {
			candidate = scored
			candidate.Model = "incomplete-oss-model"
			break
		}
	}
	result.Cases = append(result.Cases, candidate)
	if err := ValidateResult(result, pack); err == nil ||
		!strings.Contains(err.Error(), "is missing oss case") {
		t.Fatalf("incomplete OSS model error = %v", err)
	}

	result = passingResult(pack)
	result.Cases = append(result.Cases, result.Cases[2])
	if err := ValidateResult(result, pack); err == nil ||
		!strings.Contains(err.Error(), "repeats case") {
		t.Fatalf("repeated model case error = %v", err)
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
			SourceRevision: strings.Repeat("a", 40),
			PromptAuthor:   "fixture prompt author",
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
