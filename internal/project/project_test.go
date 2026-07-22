package project

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"forgejo.coilysiren.me/coilyco-flight-deck/agent-compose/internal/compose"
)

func composeFixture(t *testing.T, name string) string {
	t.Helper()
	result, err := compose.Run(filepath.Join("..", "..", "testdata", "contracts", name), t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	return result.Bundle.Dir
}

func readTarget(t *testing.T, target, rel string) string {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join(target, filepath.FromSlash(rel)))
	if err != nil {
		t.Fatalf("expected %s at load point: %v", rel, err)
	}
	return string(raw)
}

func TestProjectNativeLayouts(t *testing.T) {
	bundleDir := composeFixture(t, "native-full.kdl")
	cases := map[string]struct{ instructions, skillsDir string }{
		"claude": {"CLAUDE.md", ".claude/skills"},
		"codex":  {"AGENTS.md", ".agents/skills"},
	}
	for layout, want := range cases {
		t.Run(layout, func(t *testing.T) {
			target := t.TempDir()
			result, err := Project(bundleDir, layout, target)
			if err != nil {
				t.Fatal(err)
			}
			if !strings.Contains(readTarget(t, target, want.instructions), "Fixture foundation") {
				t.Fatal("instructions load point missing foundation content")
			}
			skill := readTarget(t, target, want.skillsDir+"/personality-curious/SKILL.md")
			if !strings.Contains(skill, "# Curious") {
				t.Fatalf("skill load point wrong content:\n%s", skill)
			}
			if len(result.Files) != 2 {
				t.Fatalf("expected 2 projected files, got %v", result.Files)
			}
		})
	}
}

func TestProjectCompiledLayouts(t *testing.T) {
	bundleDir := composeFixture(t, "compiled-full.kdl")
	target := t.TempDir()
	if _, err := Project(bundleDir, "goose", target); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(readTarget(t, target, ".goosehints"), "# Grounded") {
		t.Fatal("goose load point missing compiled prose")
	}

	briefBundle := composeFixture(t, "compiled-brief.kdl")
	target = t.TempDir()
	if _, err := Project(briefBundle, "opencode", target); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(readTarget(t, target, "AGENTS.md"), "Grounded: calm") {
		t.Fatal("opencode load point missing brief compiled prose")
	}
}

func TestProjectRejectsUnknownLayoutAndModeMismatch(t *testing.T) {
	native := composeFixture(t, "native-full.kdl")
	if _, err := Project(native, "emacs", t.TempDir()); err == nil || !strings.Contains(err.Error(), "v0.1 layouts") {
		t.Fatalf("expected unknown-layout diagnostic, got %v", err)
	}
	if _, err := Project(native, "goose", t.TempDir()); err == nil || !strings.Contains(err.Error(), "delivers native-skills") {
		t.Fatalf("expected mode-mismatch diagnostic, got %v", err)
	}
}

func TestProjectRefusesForeignFiles(t *testing.T) {
	bundleDir := composeFixture(t, "native-full.kdl")
	target := t.TempDir()
	handAuthored := filepath.Join(target, "CLAUDE.md")
	if err := os.WriteFile(handAuthored, []byte("mine\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := Project(bundleDir, "claude", target); err == nil || !strings.Contains(err.Error(), "refusing to overwrite") {
		t.Fatalf("expected ownership refusal, got %v", err)
	}
	if raw, _ := os.ReadFile(handAuthored); string(raw) != "mine\n" {
		t.Fatalf("foreign file must stay untouched, got %q", raw)
	}
}

func TestReprojectionReplacesOwnFilesOnly(t *testing.T) {
	target := t.TempDir()
	curious := composeFixture(t, "native-full.kdl")
	if _, err := Project(curious, "claude", target); err != nil {
		t.Fatal(err)
	}
	foreign := filepath.Join(target, "notes.md")
	if err := os.WriteFile(foreign, []byte("keep me\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	meticulous := composeFixture(t, "native-brief.kdl")
	result, err := Project(meticulous, "claude", target)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(readTarget(t, target, ".claude/skills/personality-meticulous/SKILL.md"), "# Meticulous") {
		t.Fatal("new skill missing after re-projection")
	}
	if _, err := os.Stat(filepath.Join(target, ".claude", "skills", "personality-curious")); !os.IsNotExist(err) {
		t.Fatal("stale skill must be removed on re-projection")
	}
	if raw, _ := os.ReadFile(foreign); string(raw) != "keep me\n" {
		t.Fatal("foreign file must survive re-projection")
	}
	for _, rel := range result.Files {
		if strings.Contains(rel, "curious") {
			t.Fatalf("sidecar still lists stale file %s", rel)
		}
	}
}
