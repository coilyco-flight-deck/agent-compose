package project

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"forgejo.coilysiren.me/coilyco-flight-deck/agent-compose/internal/compose"
	"forgejo.coilysiren.me/coilyco-flight-deck/agent-compose/internal/person"
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
		"claude":   {"CLAUDE.md", ".claude/skills"},
		"codex":    {"AGENTS.md", ".agents/skills"},
		"goose":    {".goosehints", ".agents/skills"},
		"opencode": {"AGENTS.md", ".agents/skills"},
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
			for _, identity := range activePersonalitySkills(t, "engineer") {
				skill := readTarget(t, target, want.skillsDir+"/"+identity+"/SKILL.md")
				if !strings.Contains(skill, skillHeading(identity)) {
					t.Fatalf("skill load point %s has wrong content:\n%s", identity, skill)
				}
			}
			if len(result.Files) < 2 {
				t.Fatalf("expected instructions and personality files, got %v", result.Files)
			}
		})
	}
}

func TestProjectCompiledLayouts(t *testing.T) {
	bundleDir := composeFixture(t, "compiled-full.kdl")
	cases := map[string]string{
		"claude":   "CLAUDE.md",
		"codex":    "AGENTS.md",
		"goose":    ".goosehints",
		"opencode": "AGENTS.md",
	}
	for layout, instructions := range cases {
		t.Run(layout, func(t *testing.T) {
			target := t.TempDir()
			result, err := Project(bundleDir, layout, target)
			if err != nil {
				t.Fatal(err)
			}
			compiled := readTarget(t, target, instructions)
			for _, identity := range activePersonalitySkills(t, "engineer") {
				if !strings.Contains(compiled, skillHeading(identity)) {
					t.Fatalf("%s load point missing compiled prose for %s", layout, identity)
				}
			}
			if len(result.Files) != 1 {
				t.Fatalf("compiled projection must place one file, got %v", result.Files)
			}
		})
	}

	briefBundle := composeFixture(t, "compiled-brief.kdl")
	target := t.TempDir()
	if _, err := Project(briefBundle, "opencode", target); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(readTarget(t, target, "AGENTS.md"), "Grounded: calm") {
		t.Fatal("opencode load point missing brief compiled prose")
	}
}

func TestProjectRejectsUnknownLayoutAndUnsupportedMode(t *testing.T) {
	native := composeFixture(t, "native-full.kdl")
	if _, err := Project(native, "emacs", t.TempDir()); err == nil || !strings.Contains(err.Error(), "v0.1 layouts") {
		t.Fatalf("expected unknown-layout diagnostic, got %v", err)
	}

	Registry["compiled-only"] = Layout{Compiled: &LoadPoints{Instructions: "CONTEXT.md"}}
	defer delete(Registry, "compiled-only")
	if _, err := Project(native, "compiled-only", t.TempDir()); err == nil || !strings.Contains(err.Error(), "does not support bundles") {
		t.Fatalf("expected unsupported-mode diagnostic, got %v", err)
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

func TestHomeScopeProjection(t *testing.T) {
	native := composeFixture(t, "native-full.kdl")
	compiled := composeFixture(t, "compiled-full.kdl")
	nativeBefore := treeFingerprint(t, native)
	compiledBefore := treeFingerprint(t, compiled)
	expectedIdentities := activePersonalitySkills(t, "engineer")
	cases := map[string]struct{ instructions, skillsDir string }{
		"claude":   {".claude/CLAUDE.md", ".claude/skills"},
		"codex":    {".codex/AGENTS.md", ".agents/skills"},
		"goose":    {".config/goose/.goosehints", ".agents/skills"},
		"opencode": {".config/opencode/AGENTS.md", ".agents/skills"},
	}
	for layout, want := range cases {
		t.Run(layout, func(t *testing.T) {
			home := t.TempDir()
			if _, err := ProjectScoped(native, layout, home, ScopeHome); err != nil {
				t.Fatal(err)
			}
			if !strings.Contains(readTarget(t, home, want.instructions), "Fixture foundation") {
				t.Fatalf("home instructions load point %s missing content", want.instructions)
			}
			for _, identity := range expectedIdentities {
				readTarget(t, home, want.skillsDir+"/"+identity+"/SKILL.md")
			}
			identities := projectedIdentities(t, home)
			slices.Sort(identities)
			if !slices.Equal(identities, expectedIdentities) {
				t.Fatalf("native home identities = %v, want %v", identities, expectedIdentities)
			}

			home = t.TempDir()
			if _, err := ProjectScoped(compiled, layout, home, ScopeHome); err != nil {
				t.Fatal(err)
			}
			compiledContext := readTarget(t, home, want.instructions)
			for _, identity := range expectedIdentities {
				if !strings.Contains(compiledContext, skillHeading(identity)) {
					t.Fatalf("home compiled load point missing %s", identity)
				}
			}
			if identities := projectedIdentities(t, home); len(identities) != 0 {
				t.Fatalf("compiled home unexpectedly mounted identity trees: %v", identities)
			}
		})
	}

	if _, err := ProjectScoped(native, "claude", t.TempDir(), "galaxy"); err == nil || !strings.Contains(err.Error(), "unknown scope") {
		t.Fatalf("expected unknown-scope diagnostic, got %v", err)
	}
	if after := treeFingerprint(t, native); after != nativeBefore {
		t.Fatal("home projection mutated the native input bundle")
	}
	if after := treeFingerprint(t, compiled); after != compiledBefore {
		t.Fatal("home projection mutated the compiled input bundle")
	}
}

func TestReprojectionPreservesEquivalentPersonalitySetAndForeignFiles(t *testing.T) {
	target := t.TempDir()
	full := composeFixture(t, "native-full.kdl")
	if _, err := Project(full, "claude", target); err != nil {
		t.Fatal(err)
	}
	foreign := filepath.Join(target, "notes.md")
	if err := os.WriteFile(foreign, []byte("keep me\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	brief := composeFixture(t, "native-brief.kdl")
	result, err := Project(brief, "claude", target)
	if err != nil {
		t.Fatal(err)
	}
	for _, identity := range activePersonalitySkills(t, "engineer") {
		if _, err := os.Stat(filepath.Join(target, ".claude", "skills", identity)); err != nil {
			t.Fatalf("active identity %s missing after re-projection: %v", identity, err)
		}
	}
	if raw, _ := os.ReadFile(foreign); string(raw) != "keep me\n" {
		t.Fatal("foreign file must survive re-projection")
	}
	for _, rel := range result.Files {
		if strings.Contains(rel, "fixture-review") {
			t.Fatalf("sidecar lists an excluded skill %s", rel)
		}
	}
}

func TestProjectionRollsBackAfterWriteFailure(t *testing.T) {
	target := t.TempDir()
	initial := map[string][]byte{"one.txt": []byte("old\n")}
	if _, err := ApplyOwned(target, initial, "fixture", "old-bundle"); err != nil {
		t.Fatal(err)
	}
	sidecarPath := filepath.Join(target, filepath.FromSlash(sidecarRel))
	sidecarBefore, err := os.ReadFile(sidecarPath)
	if err != nil {
		t.Fatal(err)
	}

	writeFile := func(path string, content []byte, mode fs.FileMode) error {
		if filepath.Base(path) == "two.txt" {
			return errors.New("injected write failure")
		}
		return os.WriteFile(path, content, mode)
	}
	next := map[string][]byte{
		"one.txt": []byte("new\n"),
		"two.txt": []byte("new\n"),
	}
	if _, err := applyOwnedWithWriter(target, next, "fixture", "new-bundle", writeFile); err == nil ||
		!strings.Contains(err.Error(), "previous projection restored") {
		t.Fatalf("expected restored projection failure, got %v", err)
	}
	if got := readTarget(t, target, "one.txt"); got != "old\n" {
		t.Fatalf("owned file was not restored, got %q", got)
	}
	if _, err := os.Stat(filepath.Join(target, "two.txt")); !os.IsNotExist(err) {
		t.Fatalf("partially written file survived rollback: %v", err)
	}
	sidecarAfter, err := os.ReadFile(sidecarPath)
	if err != nil {
		t.Fatal(err)
	}
	if string(sidecarAfter) != string(sidecarBefore) {
		t.Fatal("projection sidecar changed despite rollback")
	}
}

func TestInvalidBundleDoesNotTouchProjectionTarget(t *testing.T) {
	bundleDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(bundleDir, "manifest.json"), []byte("{bad"), 0o644); err != nil {
		t.Fatal(err)
	}
	target := t.TempDir()
	marker := filepath.Join(target, "keep.txt")
	if err := os.WriteFile(marker, []byte("unchanged\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	before := treeFingerprint(t, target)
	if _, err := ProjectScoped(bundleDir, "claude", target, ScopeHome); err == nil {
		t.Fatal("invalid bundle unexpectedly projected")
	}
	if after := treeFingerprint(t, target); after != before {
		t.Fatal("bundle verification failure changed the projection target")
	}
}

func TestProjectionCannotTargetItsInputBundle(t *testing.T) {
	bundleDir := composeFixture(t, "native-full.kdl")
	before := treeFingerprint(t, bundleDir)
	if _, err := ProjectScoped(bundleDir, "claude", bundleDir, ScopeHome); err == nil ||
		!strings.Contains(err.Error(), "must not be the bundle") {
		t.Fatalf("expected input-bundle target failure, got %v", err)
	}
	if after := treeFingerprint(t, bundleDir); after != before {
		t.Fatal("rejected projection changed its input bundle")
	}
}

func TestApplyOwnedRejectsUnsafeTargetPath(t *testing.T) {
	target := t.TempDir()
	if _, err := ApplyOwned(target, map[string][]byte{"../outside": []byte("bad")}, "fixture", "bundle"); err == nil ||
		!strings.Contains(err.Error(), "safe relative path") {
		t.Fatalf("expected unsafe target-path failure, got %v", err)
	}
	if _, err := os.Stat(filepath.Join(filepath.Dir(target), "outside")); !os.IsNotExist(err) {
		t.Fatalf("unsafe target path escaped projection root: %v", err)
	}
}

func TestApplyOwnedRejectsSymlinkParent(t *testing.T) {
	target := t.TempDir()
	external := t.TempDir()
	if err := os.Symlink(external, filepath.Join(target, "linked")); err != nil {
		t.Skipf("symlinks unavailable: %v", err)
	}
	if _, err := ApplyOwned(target, map[string][]byte{"linked/context.md": []byte("bad")}, "fixture", "bundle"); err == nil ||
		!strings.Contains(err.Error(), "is a symlink") {
		t.Fatalf("expected symlink-parent failure, got %v", err)
	}
	if _, err := os.Stat(filepath.Join(external, "context.md")); !os.IsNotExist(err) {
		t.Fatalf("projection escaped through symlink parent: %v", err)
	}
}

func projectedIdentities(t *testing.T, root string) []string {
	t.Helper()
	var identities []string
	err := filepath.WalkDir(root, func(_ string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() && strings.HasPrefix(entry.Name(), "personality-") {
			identities = append(identities, entry.Name())
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	return identities
}

func activePersonalitySkills(t *testing.T, roleName string) []string {
	t.Helper()
	p, err := person.Load()
	if err != nil {
		t.Fatal(err)
	}
	role, ok := p.Roles[roleName]
	if !ok {
		t.Fatalf("fixture role %q is absent", roleName)
	}
	skills := make([]string, 0, len(role.Personalities))
	for _, personalityName := range role.Personalities {
		skills = append(skills, p.Personalities[personalityName].Skill)
	}
	slices.Sort(skills)
	return skills
}

func skillHeading(skillID string) string {
	name := strings.TrimPrefix(skillID, "personality-")
	return "# " + strings.ToUpper(name[:1]) + name[1:]
}

func treeFingerprint(t *testing.T, root string) string {
	t.Helper()
	hash := sha256.New()
	err := filepath.WalkDir(root, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		info, err := entry.Info()
		if err != nil {
			return err
		}
		fmt.Fprintf(hash, "%s\x00%s\x00", filepath.ToSlash(rel), info.Mode())
		if info.Mode().IsRegular() {
			raw, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			hash.Write(raw)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	return fmt.Sprintf("%x", hash.Sum(nil))
}
