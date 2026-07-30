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
		evalCase := pack.Cases[index]
		if evalCase.ModelTier == frontierTier {
			result.Cases[index].Model = "frontier-test"
		} else {
			result.Cases[index].Model = "oss-test"
		}
		if evalCase.ModelTier == ossTier && evalCase.Dimension == roleDimension {
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
		"F frontier-test · O oss-test · 7/12 pass · 86/96 points",
		"| engineer | 32/32✓ | 8✓ | 8✓ | 24/32× | 8✓ | 6× | 86/96 |",
		"`F` frontier · `O` OSS · `R` role · `P` personality · `A` adjacent-role discrimination",
	} {
		if !strings.Contains(string(first), want) {
			t.Errorf("scorecard omitted %q:\n%s", want, first)
		}
	}
}

func TestMarkdownScorecardRequiresSelectedResults(t *testing.T) {
	t.Parallel()
	if _, err := MarkdownScorecard(t.TempDir(), "codex"); err == nil ||
		!strings.Contains(err.Error(), "no results") {
		t.Fatalf("empty scorecard error = %v", err)
	}
}

func TestMarkdownHistoricalScorecardInfersV2ScenarioShape(t *testing.T) {
	t.Parallel()
	pack, err := Build("engineer", "codex")
	if err != nil {
		t.Fatal(err)
	}
	result := passingResult(pack)
	result.Format = ResultFormatV2
	result.Provenance.PackDigest = "sha256:historical-fixture"
	raw, err := marshalYAML(result)
	if err != nil {
		t.Fatal(err)
	}
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "engineer-codex.yaml"), raw, 0o600); err != nil {
		t.Fatal(err)
	}

	scorecard, err := MarkdownHistoricalScorecard(root, "codex")
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{
		"| role | FR | FP | FA | OR | OP | OA | Σ |",
		"| engineer | 32/32✓ | 8✓ | 8✓ | 32/32✓ | 8✓ | 8✓ | 96/96 |",
	} {
		if !strings.Contains(string(scorecard), want) {
			t.Errorf("historical scorecard omitted %q:\n%s", want, scorecard)
		}
	}
}
