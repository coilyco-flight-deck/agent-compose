package evaluation

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestMarkdownScorecardRendersValidatedResults(t *testing.T) {
	t.Parallel()
	pack, err := Build("engineer", "codex")
	if err != nil {
		t.Fatal(err)
	}
	result := passingResult(pack)
	result.Provenance.EvaluatedAt = "2026-07-30"
	for index := range result.Cases {
		result.Cases[index].Model = "subject-test"
		if pack.Cases[index].Dimension == roleDimension {
			result.Cases[index].Scores[0].Score = 0
			result.Cases[index].Total -= 2
			result.Cases[index].Passed = false
		}
	}
	raw, err := MarshalResult(result, pack)
	if err != nil {
		t.Fatal(err)
	}
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "engineer-codex.yaml"), raw, 0o600); err != nil {
		t.Fatal(err)
	}

	first, err := MarkdownScorecard(root, "codex")
	if err != nil {
		t.Fatal(err)
	}
	second, err := MarkdownScorecard(root, "codex")
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(first, second) {
		t.Fatal("scorecard output is not deterministic")
	}
	for _, want := range []string{
		"# Evaluation scorecard",
		"2026-07-30 // seat codex // " + SubjectModelTier + " subject-test",
		"| role | R | P | A | Σ |",
		"| engineer |",
		"`R` role // `P` personality // `A` adjacent-role discrimination",
	} {
		if !strings.Contains(string(first), want) {
			t.Errorf("scorecard omitted %q:\n%s", want, first)
		}
	}
	if strings.Contains(string(first), "frontier") || strings.Contains(string(first), "OSS") {
		t.Errorf("scorecard still renders a retired lane:\n%s", first)
	}
	if !strings.Contains(string(first), "×") {
		t.Errorf("scorecard lost the failing role cell:\n%s", first)
	}
}

func TestMarkdownScorecardRequiresSelectedResults(t *testing.T) {
	t.Parallel()
	if _, err := MarkdownScorecard(t.TempDir(), "codex"); err == nil ||
		!strings.Contains(err.Error(), "no results") {
		t.Fatalf("empty scorecard error = %v", err)
	}
}

// A record earned on one subject cannot silently carry a case from another, or
// the single lane would sum two different measurements.
func TestMarkdownScorecardRejectsMixedTiers(t *testing.T) {
	t.Parallel()
	root := historicalFixture(t, func(result *ScoredResult) {
		_, scenario, _ := strings.Cut(result.Cases[0].ID, "-")
		result.Cases[0].ID = "frontier-" + scenario
	})
	if _, err := MarkdownHistoricalScorecard(root, "codex"); err == nil ||
		!strings.Contains(err.Error(), "mixes tiers") {
		t.Fatalf("mixed-tier scorecard error = %v", err)
	}
}

func TestMarkdownHistoricalScorecardInfersV2ScenarioShape(t *testing.T) {
	t.Parallel()
	root := historicalFixture(t, nil)
	scorecard, err := MarkdownHistoricalScorecard(root, "codex")
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{
		"| role | R | P | A | Σ |",
		"| engineer |",
		SubjectModelTier + " fixture-model",
	} {
		if !strings.Contains(string(scorecard), want) {
			t.Errorf("historical scorecard omitted %q:\n%s", want, scorecard)
		}
	}
}

func historicalFixture(t *testing.T, mutate func(*ScoredResult)) string {
	t.Helper()
	pack, err := Build("engineer", "codex")
	if err != nil {
		t.Fatal(err)
	}
	result := passingResult(pack)
	result.Format = ResultFormatV2
	result.Provenance.PackDigest = "sha256:historical-fixture"
	if mutate != nil {
		mutate(result)
	}
	raw, err := marshalYAML(result)
	if err != nil {
		t.Fatal(err)
	}
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "engineer-codex.yaml"), raw, 0o600); err != nil {
		t.Fatal(err)
	}
	return root
}
