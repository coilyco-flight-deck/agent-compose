package skillmount

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func makeSkill(t *testing.T, repo, name string) string {
	t.Helper()
	path := filepath.Join(repo, ".agents", "skills", name)
	if err := os.MkdirAll(path, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(path, "SKILL.md"), []byte(name), 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

func writeEligibility(t *testing.T, path string, defaults []string, harnesses map[string][]string) {
	t.Helper()
	raw, err := json.Marshal(eligibility{Defaults: defaults, Harnesses: harnesses})
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, raw, 0o644); err != nil {
		t.Fatal(err)
	}
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
	manifest := filepath.Join(dir, "mount-eligibility.json")
	writeEligibility(t, manifest, []string{public, private}, map[string][]string{})

	result, err := Apply(manifest, loadPoints, filepath.Join(dir, "state"))
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
	manifest := filepath.Join(dir, "mount-eligibility.json")
	writeEligibility(t, manifest, []string{root}, map[string][]string{})
	if _, err := Apply(manifest, points, state); err != nil {
		t.Fatal(err)
	}

	if err := os.RemoveAll(stale); err != nil {
		t.Fatal(err)
	}
	foreign := filepath.Join(destination, "foreign")
	if err := os.WriteFile(foreign, []byte("mine"), 0o644); err != nil {
		t.Fatal(err)
	}
	result, err := Apply(manifest, points, state)
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

func TestApplyDoesNotMutateWhenManifestIsMissing(t *testing.T) {
	dir := t.TempDir()
	destination := filepath.Join(dir, "skills")
	if err := os.MkdirAll(destination, 0o755); err != nil {
		t.Fatal(err)
	}
	foreign := filepath.Join(destination, "foreign")
	if err := os.WriteFile(foreign, []byte("mine"), 0o644); err != nil {
		t.Fatal(err)
	}
	_, err := Apply(filepath.Join(dir, "missing.json"), map[string]string{"codex": destination}, filepath.Join(dir, "state"))
	if err == nil {
		t.Fatal("missing root must fail")
	}
	if raw, readErr := os.ReadFile(foreign); readErr != nil || string(raw) != "mine" {
		t.Fatal("validation failure must not mutate the destination")
	}
}

func TestApplyWithCatalogsOverlaysLocalSkillsForEveryLoadPoint(t *testing.T) {
	dir := t.TempDir()
	local := filepath.Join(dir, "local")
	remote := filepath.Join(dir, "remote-skills")
	makeSkill(t, local, "local-only")
	makeSkill(t, local, "shared")
	remoteShared := filepath.Join(remote, "shared")
	if err := os.MkdirAll(remoteShared, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(remoteShared, "SKILL.md"), []byte("remote"), 0o644); err != nil {
		t.Fatal(err)
	}
	remoteOnly := filepath.Join(remote, "remote-only")
	if err := os.MkdirAll(remoteOnly, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(remoteOnly, "SKILL.md"), []byte("remote"), 0o644); err != nil {
		t.Fatal(err)
	}

	loadPoints := map[string]string{
		"claude": filepath.Join(dir, "claude"),
		"codex":  filepath.Join(dir, "codex"),
	}
	manifest := filepath.Join(dir, "mount-eligibility.json")
	writeEligibility(t, manifest, []string{local}, map[string][]string{})
	result, err := ApplyWithCatalogs(
		manifest,
		loadPoints,
		filepath.Join(dir, "state"),
		[]Catalog{{Path: remote}},
	)
	if err != nil {
		t.Fatal(err)
	}
	if result.Linked != 6 {
		t.Fatalf("linked %d, want 6", result.Linked)
	}
	for harness, destination := range loadPoints {
		target, err := os.Readlink(filepath.Join(destination, "shared"))
		if err != nil || target != remoteShared {
			t.Fatalf("remote catalog must overlay local for %s: target=%q err=%v", harness, target, err)
		}
		if _, err := os.Readlink(filepath.Join(destination, "remote-only")); err != nil {
			t.Fatalf("remote-only skill missing from %s: %v", harness, err)
		}
	}
}
