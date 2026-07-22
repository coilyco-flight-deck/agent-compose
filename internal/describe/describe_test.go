package describe

import (
	"path/filepath"
	"strings"
	"testing"

	"forgejo.coilysiren.me/coilyco-flight-deck/agent-compose/internal/compose"
	"forgejo.coilysiren.me/coilyco-flight-deck/agent-compose/internal/resolver"
)

func composeFixture(t *testing.T, name string) string {
	t.Helper()
	result, err := compose.Run(filepath.Join("..", "..", "testdata", "contracts", name), t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	return result.Bundle.Dir
}

func TestBundleRendersSections(t *testing.T) {
	dir := composeFixture(t, "native-full.kdl")
	out, err := Bundle(dir, Options{})
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{
		"engineer/curious",
		"favorite #d98e48",
		"\nprofile\n", "\nsources\n", "\nselection\n", "\ndelivery\n",
		"role engineer", "person \"kai\" defines this role",
		"personality curious", "compatible set: curious, grounded, meticulous",
		"✓ person:kai",
		"✓ aos-public",
		"✓ skill personality-curious",
		"✗ skill fixture-review",
		"→ content/skills",
		"machine-readable trace: trace.json",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("describe output missing %q:\n%s", want, out)
		}
	}
	if strings.Contains(out, "\x1b") {
		t.Fatal("plain output must carry no ANSI codes")
	}
}

func TestCollapseFoldsLargeExclusionGroups(t *testing.T) {
	var excluded []resolver.Decision
	for range 5 {
		excluded = append(excluded, resolver.Decision{
			Subject: "skill:x", Kind: "skill",
			Outcome: resolver.OutcomeExcluded, Reason: "not bound",
		})
	}
	lines := collapse(excluded, Options{})
	if len(lines) != 1 || !strings.Contains(lines[0], "5 skills excluded") || !strings.Contains(lines[0], "--all") {
		t.Fatalf("expected one collapsed line, got %v", lines)
	}
	if got := collapse(excluded[:2], Options{}); len(got) != 2 {
		t.Fatalf("small groups must stay expanded, got %v", got)
	}
}

func TestWhyFollowsOneItem(t *testing.T) {
	dir := composeFixture(t, "native-full.kdl")

	out, err := Why(dir, "skill:personality-grounded", Options{})
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{
		"outcome: excluded",
		"considered: declared by aos-public",
		`would select under: personality "grounded" (compatible with role "engineer")`,
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("why output missing %q:\n%s", want, out)
		}
	}

	out, err = Why(dir, "personality-curious", Options{})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "outcome: selected") {
		t.Fatalf("bare-name lookup failed:\n%s", out)
	}

	if _, err := Why(dir, "skill:nonexistent", Options{}); err == nil {
		t.Fatal("expected unknown subject to error with guidance")
	}
}

func TestDiffReportsSemanticChanges(t *testing.T) {
	full := composeFixture(t, "native-full.kdl")
	brief := composeFixture(t, "native-brief.kdl")

	out, err := Diff(full, brief)
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{
		"engineer/curious", "engineer/meticulous",
		"~ skill personality-curious", "selected → excluded",
		"~ skill personality-meticulous",
		"decisions unchanged",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("diff output missing %q:\n%s", want, out)
		}
	}

	same, err := Diff(full, full)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(same, "no semantic differences") {
		t.Fatalf("identical bundles must diff clean:\n%s", same)
	}
}
