package evaluation

import (
	"encoding/json"
	"strings"
	"testing"

	"forgejo.coilysiren.me/coilyco-flight-deck/agent-compose/internal/schema"
)

func TestBuildEmitsFourCaseFrontierOSSMatrix(t *testing.T) {
	pack, err := Build("engineer", "codex")
	if err != nil {
		t.Fatal(err)
	}
	if pack.Format != Format || pack.Role != "engineer" ||
		pack.Seat.Harness != "codex" || pack.Seat.Name == "" {
		t.Fatalf("evaluation identity = %+v", pack)
	}
	if len(pack.Cases) != 4 {
		t.Fatalf("evaluation cases = %d, want 4", len(pack.Cases))
	}
	want := []struct {
		id         string
		tier       string
		modelClass string
		dimension  string
	}{
		{"frontier-role-understanding", "frontier", schema.ModelClassFrontier, "role-understanding"},
		{"frontier-personality-expression", "frontier", schema.ModelClassFrontier, "personality-expression"},
		{"oss-role-understanding", "oss", schema.ModelClassLowContext, "role-understanding"},
		{"oss-personality-expression", "oss", schema.ModelClassLowContext, "personality-expression"},
	}
	for i, expected := range want {
		got := pack.Cases[i]
		if got.ID != expected.id || got.ModelTier != expected.tier ||
			got.BundleModelClass != expected.modelClass || got.Dimension != expected.dimension {
			t.Errorf("case %d = %+v, want %+v", i, got, expected)
		}
		if len(got.Rubric) != 4 || !strings.Contains(got.Prompt, "Do not name") {
			t.Errorf("case %q is not review-ready: %+v", got.ID, got)
		}
	}
}

func TestBuildCarriesSelfContainedReviewContext(t *testing.T) {
	pack, err := Build("ops", "codex")
	if err != nil {
		t.Fatal(err)
	}
	if paragraphs := strings.Count(pack.Briefing, "\n\n") + 1; paragraphs < 3 {
		t.Fatalf("ops briefing has %d paragraphs", paragraphs)
	}
	if len(pack.Personalities) != 3 || pack.Invariant == "" || pack.MeldedFavoriteColor == "" {
		t.Fatalf("evaluation context is incomplete: %+v", pack)
	}
	for _, personality := range pack.Personalities {
		if personality.Definition == "" || strings.HasPrefix(personality.Definition, "---") {
			t.Errorf("personality definition was not normalized: %+v", personality)
		}
	}
}

func TestBuildReviewMinimumsReferenceCaseCriteria(t *testing.T) {
	pack, err := Build("ops", "codex")
	if err != nil {
		t.Fatal(err)
	}
	if pack.ReviewRule.PassingTotal <= 0 || pack.ReviewRule.PassingTotal > 8 {
		t.Fatalf("passing total is outside the rubric range: %+v", pack.ReviewRule)
	}
	if !pack.ReviewRule.RejectZeroScores {
		t.Fatal("strict review must reject a missing criterion")
	}
	for _, evalCase := range pack.Cases {
		minimums := pack.ReviewRule.RoleMinimumScores
		if evalCase.Dimension == "personality-expression" {
			minimums = pack.ReviewRule.PersonalityMinimumScores
		}
		if len(minimums) == 0 {
			t.Fatalf("case %q has no dimension minimums", evalCase.ID)
		}
		criteria := make(map[string]bool, len(evalCase.Rubric))
		for _, criterion := range evalCase.Rubric {
			criteria[criterion.ID] = true
		}
		for criterion, minimum := range minimums {
			if !criteria[criterion] {
				t.Errorf("case %q minimum names unknown criterion %q", evalCase.ID, criterion)
			}
			if minimum < 0 || minimum > 2 {
				t.Errorf("case %q minimum for %q is outside score range: %d", evalCase.ID, criterion, minimum)
			}
		}
	}
}

func TestBuildRejectsUnknownRoleOrSeat(t *testing.T) {
	if _, err := Build("missing", "codex"); err == nil || !strings.Contains(err.Error(), "is not defined") {
		t.Fatalf("unknown role error = %v", err)
	}
	if _, err := Build("engineer", "missing"); err == nil || !strings.Contains(err.Error(), "has no") {
		t.Fatalf("unknown seat error = %v", err)
	}
}

func TestJSONAndMarkdownAreDeterministic(t *testing.T) {
	pack, err := Build("engineer", "codex")
	if err != nil {
		t.Fatal(err)
	}
	first, err := Marshal(pack)
	if err != nil {
		t.Fatal(err)
	}
	second, err := Marshal(pack)
	if err != nil {
		t.Fatal(err)
	}
	if string(first) != string(second) {
		t.Fatal("JSON evaluation pack is not deterministic")
	}
	var decoded Pack
	if err := json.Unmarshal(first, &decoded); err != nil {
		t.Fatalf("decode evaluation JSON: %v", err)
	}
	markdown := string(Markdown(pack))
	for _, want := range []string{
		"# Agent-compose behavior evaluation",
		"## Four-case matrix",
		"### frontier-role-understanding",
		"### oss-personality-expression",
	} {
		if !strings.Contains(markdown, want) {
			t.Errorf("Markdown evaluation omitted %q", want)
		}
	}
}
