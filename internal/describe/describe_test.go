package describe

import (
	"path/filepath"
	"strings"
	"testing"

	"forgejo.coilysiren.me/coilyco-flight-deck/agent-compose/internal/compose"
	"forgejo.coilysiren.me/coilyco-flight-deck/agent-compose/internal/person"
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
	p, err := person.Load()
	if err != nil {
		t.Fatal(err)
	}
	personalitySet := strings.Join(p.Roles["engineer"].Personalities, "+")
	dir := composeFixture(t, "native.kdl")
	out, err := Bundle(dir, Options{})
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{
		"engineer/" + personalitySet,
		"melded #",
		"\nprofile\n", "\nsources\n", "\nselection\n", "\ndelivery\n",
		"role engineer", "person \"kai\" defines this role",
		"personality curious", "activates its full personality set",
		"✓ person:kai",
		"✓ aos-public",
		"✓ skill personality-curious",
		"✓ skill personality-grounded",
		"✓ skill personality-meticulous",
		"✓ skill fixture-review",
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
	dir := composeFixture(t, "native.kdl")

	out, err := Why(dir, "skill:fixture-review", Options{})
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{
		"outcome: selected",
		"considered: declared by aos-public",
		"ordinary provider skills are discoverable for every role",
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
	p, err := person.Load()
	if err != nil {
		t.Fatal(err)
	}
	identity := "engineer/" + strings.Join(p.Roles["engineer"].Personalities, "+")
	native := composeFixture(t, "native.kdl")
	compiled := composeFixture(t, "compiled.kdl")

	out, err := Diff(native, compiled)
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{
		identity,
		"native-skills",
		"compiled",
		"+ delivery/compiled.md",
		"decisions unchanged",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("diff output missing %q:\n%s", want, out)
		}
	}

	same, err := Diff(native, native)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(same, "no semantic differences") {
		t.Fatalf("identical bundles must diff clean:\n%s", same)
	}
}
