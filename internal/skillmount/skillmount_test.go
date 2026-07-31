package skillmount

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
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
	raw, err := json.Marshal(Eligibility{
		ProjectsRoot: filepath.Dir(path),
		Defaults:     defaults,
		Harnesses:    harnesses,
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, raw, 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestEligibilityRoleProviderOrderingAndStrictLoading(t *testing.T) {
	dir := t.TempDir()
	defaultProvider := filepath.Join(dir, "default")
	harnessProvider := filepath.Join(dir, "harness")
	roleOne := filepath.Join(dir, "role-one")
	roleTwo := filepath.Join(dir, "role-two")
	manifestPath := filepath.Join(dir, "mount-eligibility.json")
	raw, err := json.Marshal(Eligibility{
		ProjectsRoot: dir,
		Defaults:     []string{defaultProvider},
		Harnesses:    map[string][]string{"codex": {harnessProvider}},
		RoleProviders: map[string][]RoleProvider{
			"ops": {
				{Path: roleOne, Required: true},
				{Path: roleTwo},
			},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(manifestPath, raw, 0o644); err != nil {
		t.Fatal(err)
	}
	manifest, err := LoadEligibility(manifestPath)
	if err != nil {
		t.Fatal(err)
	}
	providers := manifest.Providers("codex", "ops")
	want := []string{defaultProvider, harnessProvider, roleOne, roleTwo}
	if len(providers) != len(want) {
		t.Fatalf("providers = %+v, want %v", providers, want)
	}
	for index, path := range want {
		if providers[index].Path != path {
			t.Fatalf("provider order = %+v, want %v", providers, want)
		}
	}
	if got := manifest.Repositories("codex"); len(got) != 2 || got[0] != defaultProvider || got[1] != harnessProvider {
		t.Fatalf("bare repositories leaked role providers: %v", got)
	}
	makeSkill(t, defaultProvider, "global")
	makeSkill(t, roleOne, "role-only")
	destination := filepath.Join(dir, "native-skills")
	if _, err := Apply(
		manifestPath,
		map[string]string{"codex": destination},
		filepath.Join(dir, "state"),
	); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(destination, "global")); err != nil {
		t.Fatalf("bare convergence omitted global skill: %v", err)
	}
	if _, err := os.Lstat(filepath.Join(destination, "role-only")); !os.IsNotExist(err) {
		t.Fatalf("bare convergence leaked role-only skill: %v", err)
	}

	for name, body := range map[string]string{
		"top-level": `{"projects_root":"/tmp","defaults":[],"harnesses":{},"extra":true}`,
		"nested":    `{"projects_root":"/tmp","defaults":[],"harnesses":{},"role_providers":{"ops":[{"path":"/tmp/provider","required":true,"extra":true}]}}`,
		"trailing":  `{"projects_root":"/tmp","defaults":[],"harnesses":{}} {}`,
	} {
		t.Run(name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "manifest.json")
			if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
				t.Fatal(err)
			}
			if _, err := LoadEligibility(path); err == nil {
				t.Fatal("invalid eligibility passed strict loading")
			}
		})
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

func TestApplyWarnsAndRemovesOwnedLinkWhenSkillTargetDisappears(t *testing.T) {
	dir := t.TempDir()
	root := filepath.Join(dir, "root")
	skillRoot := filepath.Join(root, ".agents", "skills")
	makeSkill(t, root, "kept")
	target := filepath.Join(dir, "fragile-target")
	if err := os.MkdirAll(target, 0o755); err != nil {
		t.Fatal(err)
	}
	fragile := filepath.Join(skillRoot, "fragile")
	if err := os.Symlink(target, fragile); err != nil {
		t.Fatal(err)
	}
	destination := filepath.Join(dir, "skills")
	state := filepath.Join(dir, "state")
	points := map[string]string{"codex": destination}
	manifest := filepath.Join(dir, "mount-eligibility.json")
	writeEligibility(t, manifest, []string{root}, map[string][]string{})

	if _, err := Apply(manifest, points, state); err != nil {
		t.Fatal(err)
	}
	if err := os.RemoveAll(target); err != nil {
		t.Fatal(err)
	}

	result, err := Apply(manifest, points, state)
	if err != nil {
		t.Fatalf("vanished skill target must not fail convergence: %v", err)
	}
	if result.Removed != 1 || result.Verified != 1 {
		t.Fatalf("result = %+v, want one stale removal and one verified skill", result)
	}
	if len(result.Warnings) != 1 ||
		!strings.Contains(result.Warnings[0], fragile) ||
		!strings.Contains(result.Warnings[0], "no such file") {
		t.Fatalf("warnings = %q, want vanished skill path and cause", result.Warnings)
	}
	if _, err := os.Lstat(filepath.Join(destination, "fragile")); !os.IsNotExist(err) {
		t.Fatal("owned projection for vanished skill must be removed")
	}
	if _, err := os.Stat(filepath.Join(destination, "kept")); err != nil {
		t.Fatalf("available skill must remain projected: %v", err)
	}
}

func TestApplyStillFailsOnNonMissingSkillInspectionError(t *testing.T) {
	dir := t.TempDir()
	root := filepath.Join(dir, "root")
	skillRoot := filepath.Join(root, ".agents", "skills")
	if err := os.MkdirAll(skillRoot, 0o755); err != nil {
		t.Fatal(err)
	}
	loop := filepath.Join(skillRoot, "loop")
	if err := os.Symlink(loop, loop); err != nil {
		t.Fatal(err)
	}
	manifest := filepath.Join(dir, "mount-eligibility.json")
	writeEligibility(t, manifest, []string{root}, map[string][]string{})

	_, err := Apply(
		manifest,
		map[string]string{"codex": filepath.Join(dir, "skills")},
		filepath.Join(dir, "state"),
	)
	if err == nil || !strings.Contains(err.Error(), "inspect skill") {
		t.Fatalf("symlink loop must remain fatal, got %v", err)
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
	additional := filepath.Join(dir, "additional-skills")
	makeSkill(t, local, "local-only")
	makeSkill(t, local, "shared")
	additionalShared := filepath.Join(additional, "shared")
	if err := os.MkdirAll(additionalShared, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(additionalShared, "SKILL.md"), []byte("additional"), 0o644); err != nil {
		t.Fatal(err)
	}
	additionalOnly := filepath.Join(additional, "additional-only")
	if err := os.MkdirAll(additionalOnly, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(additionalOnly, "SKILL.md"), []byte("additional"), 0o644); err != nil {
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
		[]Catalog{{Path: additional}},
	)
	if err != nil {
		t.Fatal(err)
	}
	if result.Linked != 6 {
		t.Fatalf("linked %d, want 6", result.Linked)
	}
	for harness, destination := range loadPoints {
		target, err := os.Readlink(filepath.Join(destination, "shared"))
		if err != nil || target != additionalShared {
			t.Fatalf("additional catalog must overlay local for %s: target=%q err=%v", harness, target, err)
		}
		if _, err := os.Readlink(filepath.Join(destination, "additional-only")); err != nil {
			t.Fatalf("additional-only skill missing from %s: %v", harness, err)
		}
	}
}
