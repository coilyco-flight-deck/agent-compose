package nativelaunch

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"forgejo.coilysiren.me/coilyco-flight-deck/agent-compose/internal/person"
	"forgejo.coilysiren.me/coilyco-flight-deck/agent-compose/internal/skillmount"
)

func writeFile(t *testing.T, path, body string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

func writeProvider(t *testing.T, root string, withRole bool) {
	t.Helper()
	writeFile(
		t,
		filepath.Join(root, ".agents", "skills", "ordinary", "SKILL.md"),
		"---\nname: ordinary\ndescription: Ordinary test skill.\n---\n\n# Ordinary\n",
	)
	if !withRole {
		return
	}
	writeFile(
		t,
		filepath.Join(root, ".agents", "composed", "design-method", "COMPOSED.md"),
		"---\nname: design-method\ndescription: Role-composed design method.\n---\n\n# Design method\n",
	)
	writeFile(
		t,
		filepath.Join(root, ".agents", "roles.kdl"),
		"roles {\n    role designer {\n        composed-skill design-method\n    }\n}\n",
	)
}

func writeManifest(t *testing.T, path, projectsRoot, provider string) {
	t.Helper()
	raw, err := json.Marshal(skillmount.Eligibility{
		ProjectsRoot: projectsRoot,
		Defaults:     []string{provider},
		Harnesses:    map[string][]string{},
	})
	if err != nil {
		t.Fatal(err)
	}
	writeFile(t, path, string(raw))
}

func TestRefreshProjectsAssignedRoleBundleForEveryNativeHarness(t *testing.T) {
	projects := filepath.Join(t.TempDir(), "projects")
	provider := filepath.Join(projects, "coilyco-flight-deck", "agentic-os")
	writeProvider(t, provider, true)
	manifest := filepath.Join(t.TempDir(), "mount-eligibility.json")
	writeManifest(t, manifest, projects, provider)
	out := filepath.Join(t.TempDir(), "bundles")
	profile, err := person.Load()
	if err != nil {
		t.Fatal(err)
	}
	designer := profile.Roles["designer"]

	cases := map[string]struct {
		instructions string
		skills       string
		modelClass   string
	}{
		"claude":   {"CLAUDE.md", ".claude/skills", "frontier"},
		"codex":    {"AGENTS.md", ".agents/skills", "frontier"},
		"goose":    {".goosehints", ".agents/skills", "low-context"},
		"opencode": {"AGENTS.md", ".agents/skills", "low-context"},
	}
	for harness, tc := range cases {
		t.Run(harness, func(t *testing.T) {
			target := t.TempDir()
			result, err := Refresh(Options{
				Role:         "designer",
				Harness:      harness,
				ModelClass:   tc.modelClass,
				CWD:          projects,
				TargetDir:    target,
				ManifestPath: manifest,
				OutDir:       out,
			})
			if err != nil {
				t.Fatal(err)
			}
			if result.ModelClass != tc.modelClass {
				t.Fatalf("model class = %q, want %q", result.ModelClass, tc.modelClass)
			}
			instructions, err := os.ReadFile(filepath.Join(target, tc.instructions))
			if err != nil {
				t.Fatal(err)
			}
			if !strings.Contains(string(instructions), "assigned the `designer` role") {
				t.Fatalf("assigned role missing from instructions:\n%s", instructions)
			}
			skills := []string{
				"ordinary",
				"role-designer",
				"design-method",
			}
			for _, name := range designer.Personalities {
				skills = append(skills, profile.Personalities[name].Skill)
			}
			for _, skill := range skills {
				if _, err := os.Stat(filepath.Join(target, tc.skills, skill, "SKILL.md")); err != nil {
					t.Errorf("selected skill %s missing: %v", skill, err)
				}
			}
			if _, err := os.Stat(filepath.Join(
				target,
				tc.skills,
				"role-engineer",
				"SKILL.md",
			)); !os.IsNotExist(err) {
				t.Errorf("inactive role-engineer entered the bundle: %v", err)
			}
		})
	}
}

func TestRefreshDefaultsToFrontierModelClass(t *testing.T) {
	projects := filepath.Join(t.TempDir(), "projects")
	provider := filepath.Join(projects, "example", "provider")
	writeProvider(t, provider, true)
	manifest := filepath.Join(t.TempDir(), "mount-eligibility.json")
	writeManifest(t, manifest, projects, provider)

	result, err := Refresh(Options{
		Role:         "designer",
		Harness:      "goose",
		CWD:          projects,
		TargetDir:    t.TempDir(),
		ManifestPath: manifest,
		OutDir:       filepath.Join(t.TempDir(), "bundles"),
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.ModelClass != "frontier" {
		t.Fatalf("default model class = %q, want frontier", result.ModelClass)
	}
}

func TestRefreshRequiresRoleComposedProvider(t *testing.T) {
	projects := filepath.Join(t.TempDir(), "projects")
	provider := filepath.Join(projects, "example", "ordinary")
	writeProvider(t, provider, false)
	manifest := filepath.Join(t.TempDir(), "mount-eligibility.json")
	writeManifest(t, manifest, projects, provider)

	_, err := Refresh(Options{
		Role:         "designer",
		Harness:      "codex",
		CWD:          projects,
		TargetDir:    t.TempDir(),
		ManifestPath: manifest,
		OutDir:       filepath.Join(t.TempDir(), "bundles"),
	})
	if err == nil || !strings.Contains(err.Error(), "needs an eligible provider") {
		t.Fatalf("role-provider error = %v", err)
	}
}
