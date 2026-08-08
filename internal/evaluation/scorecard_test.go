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
		switch evalCase.ModelTier {
		case frontierTier:
			result.Cases[index].Model = "frontier-test"
		case commodityTier:
			result.Cases[index].Model = "commodity-test"
		case ossTier:
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
		"F frontier-test // C commodity-test // O oss-test // 19/27 pass // 218/234 points",
		"| engineer | 62/62✓ | 8✓ | 8✓ | 62/62✓ | 8✓ | 8✓ | 48/62× | 8✓ | 6× | 218/234 |",
		"`F` frontier // `C` commodity // `O` OSS // `R` role // `P` personality // `A` adjacent-role discrimination",
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

func TestMarkdownScorecardMarksDisabledNonFrontierTiers(t *testing.T) {
	t.Parallel()
	pack, err := Build("engineer", "codex")
	if err != nil {
		t.Fatal(err)
	}
	result := passingResult(pack)
	result.Provenance.EvaluatedAt = "2026-08-04"
	kept := result.Cases[:0]
	for _, scored := range result.Cases {
		if !strings.HasPrefix(scored.ID, frontierTier+"-") {
			continue
		}
		scored.Model = "frontier-test"
		kept = append(kept, scored)
	}
	result.Cases = kept
	raw, err := MarshalResult(result, pack)
	if err != nil {
		t.Fatal(err)
	}
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "engineer-codex.yaml"), raw, 0o600); err != nil {
		t.Fatal(err)
	}

	scorecard, err := MarkdownScorecard(root, "codex")
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{
		"F frontier-test // C disabled // O disabled // 9/9 pass // 78/78 points",
		"| engineer | 62/62✓ | 8✓ | 8✓ | - | - | - | - | - | - | 78/78 |",
	} {
		if !strings.Contains(string(scorecard), want) {
			t.Errorf("disabled-tier scorecard omitted %q:\n%s", want, scorecard)
		}
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
	kept := result.Cases[:0]
	for _, scored := range result.Cases {
		if strings.HasPrefix(scored.ID, commodityTier+"-") {
			continue
		}
		kept = append(kept, scored)
	}
	result.Cases = kept
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
		"F fixture-model // C disabled // O fixture-model // 18/18 pass // 156/156 points",
		"| role | FR | FP | FA | CR | CP | CA | OR | OP | OA | Σ |",
		"| engineer | 62/62✓ | 8✓ | 8✓ | - | - | - | 62/62✓ | 8✓ | 8✓ | 156/156 |",
	} {
		if !strings.Contains(string(scorecard), want) {
			t.Errorf("historical scorecard omitted %q:\n%s", want, scorecard)
		}
	}
}
