package skillmount

import (
	"os"
	"path/filepath"
	"testing"
)

func makeSkill(t *testing.T, root, name string) string {
	t.Helper()
	path := filepath.Join(root, name)
	if err := os.MkdirAll(path, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(path, "SKILL.md"), []byte(name), 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestApplyOverlaysAndMultipleLoadPoints(t *testing.T) {
	dir := t.TempDir()
	public := filepath.Join(dir, "public")
	private := filepath.Join(dir, "private")
	makeSkill(t, public, "coding-go")
	makeSkill(t, public, "shared")
	privateShared := makeSkill(t, private, "shared")
	makeSkill(t, private, "kai-local")
	loadPoints := map[string]string{
		"claude": filepath.Join(dir, "claude"),
		"codex":  filepath.Join(dir, "codex"),
	}

	result, err := Apply([]string{public, private}, loadPoints, filepath.Join(dir, "state"))
	if err != nil {
		t.Fatal(err)
	}
	if result.Linked != 6 {
		t.Fatalf("linked %d, want 6", result.Linked)
	}
	for _, destination := range loadPoints {
		target, err := os.Readlink(filepath.Join(destination, "shared"))
		if err != nil {
			t.Fatal(err)
		}
		if target != privateShared {
			t.Fatalf("later root must win: %s", target)
		}
	}
}

func TestApplyRemovesStaleOwnedLinksAndPreservesForeignEntries(t *testing.T) {
	dir := t.TempDir()
	root := filepath.Join(dir, "root")
	stale := makeSkill(t, root, "stale")
	makeSkill(t, root, "kept")
	destination := filepath.Join(dir, "skills")
	state := filepath.Join(dir, "state")
	points := map[string]string{"codex": destination}
	if _, err := Apply([]string{root}, points, state); err != nil {
		t.Fatal(err)
	}

	if err := os.RemoveAll(stale); err != nil {
		t.Fatal(err)
	}
	foreign := filepath.Join(destination, "foreign")
	if err := os.WriteFile(foreign, []byte("mine"), 0o644); err != nil {
		t.Fatal(err)
	}
	result, err := Apply([]string{root}, points, state)
	if err != nil {
		t.Fatal(err)
	}
	if result.Removed != 1 {
		t.Fatalf("removed %d, want 1", result.Removed)
	}
	if _, err := os.Lstat(filepath.Join(destination, "stale")); !os.IsNotExist(err) {
		t.Fatal("stale owned link must be removed")
	}
	if raw, err := os.ReadFile(foreign); err != nil || string(raw) != "mine" {
		t.Fatal("foreign entry must survive")
	}
}

func TestApplyDoesNotMutateWhenRootIsMissing(t *testing.T) {
	dir := t.TempDir()
	destination := filepath.Join(dir, "skills")
	if err := os.MkdirAll(destination, 0o755); err != nil {
		t.Fatal(err)
	}
	foreign := filepath.Join(destination, "foreign")
	if err := os.WriteFile(foreign, []byte("mine"), 0o644); err != nil {
		t.Fatal(err)
	}
	_, err := Apply([]string{filepath.Join(dir, "missing")}, map[string]string{"codex": destination}, filepath.Join(dir, "state"))
	if err == nil {
		t.Fatal("missing root must fail")
	}
	if raw, readErr := os.ReadFile(foreign); readErr != nil || string(raw) != "mine" {
		t.Fatal("validation failure must not mutate the destination")
	}
}
