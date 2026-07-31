package nativelaunch

import (
	"encoding/json"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"strings"
	"testing"

	"forgejo.coilysiren.me/coilyco-flight-deck/agent-compose/internal/describe"
	"forgejo.coilysiren.me/coilyco-flight-deck/agent-compose/internal/person"
	"forgejo.coilysiren.me/coilyco-flight-deck/agent-compose/internal/project"
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
		"roles {\n    role design {\n        composed-skill design-method\n    }\n}\n",
	)
}

func writeOrdinarySkill(t *testing.T, root, name string) {
	t.Helper()
	writeFile(
		t,
		filepath.Join(root, ".agents", "skills", name, "SKILL.md"),
		"---\nname: "+name+"\ndescription: Role-only fixture skill.\n---\n\n# "+name+"\n",
	)
}

func writeEligibilityManifest(t *testing.T, path string, manifest skillmount.Eligibility) {
	t.Helper()
	raw, err := json.Marshal(manifest)
	if err != nil {
		t.Fatal(err)
	}
	writeFile(t, path, string(raw))
}

func skillNames(t *testing.T, root string) []string {
	t.Helper()
	entries, err := os.ReadDir(root)
	if err != nil {
		t.Fatal(err)
	}
	names := make([]string, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() {
			names = append(names, entry.Name())
		}
	}
	sort.Strings(names)
	return names
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
	designer := profile.Roles["design"]

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
				Role:         "design",
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
			if result.Composition == nil ||
				result.Composition.Resolution.Request.Role != "design" {
				t.Fatalf("composition result does not retain the assigned role: %+v", result.Composition)
			}
			instructions, err := os.ReadFile(filepath.Join(target, tc.instructions))
			if err != nil {
				t.Fatal(err)
			}
			if !strings.Contains(string(instructions), "assigned the `design` role") {
				t.Fatalf("assigned role missing from instructions:\n%s", instructions)
			}
			skills := []string{
				"ordinary",
				"role-design",
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
		Role:         "design",
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
		Role:         "design",
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

func TestRoleProvidersStayScopedAcrossNativeAndStagedHomes(t *testing.T) {
	projects := filepath.Join(t.TempDir(), "projects")
	base := filepath.Join(projects, "example", "base")
	infrastructure := filepath.Join(projects, "example", "infrastructure")
	deploy := filepath.Join(projects, "example", "deploy")
	missingOptional := filepath.Join(projects, "example", "optional")
	writeProvider(t, base, true)
	writeOrdinarySkill(t, infrastructure, "infrastructure-ops")
	writeOrdinarySkill(t, deploy, "deploy-ops")
	manifestPath := filepath.Join(t.TempDir(), "mount-eligibility.json")
	writeEligibilityManifest(t, manifestPath, skillmount.Eligibility{
		ProjectsRoot: projects,
		Defaults:     []string{base},
		Harnesses:    map[string][]string{},
		RoleProviders: map[string][]skillmount.RoleProvider{
			"ops": {
				{Path: infrastructure, Required: true},
				{Path: deploy, Required: true},
				{Path: missingOptional},
			},
		},
	})
	out := filepath.Join(t.TempDir(), "bundles")
	results := map[string]*Result{}
	targets := map[string]string{}
	for _, role := range []string{"ops", "engineer"} {
		target := t.TempDir()
		targets[role] = target
		result, err := Refresh(Options{
			Role:         role,
			Harness:      "codex",
			CWD:          projects,
			TargetDir:    target,
			ManifestPath: manifestPath,
			OutDir:       out,
		})
		if err != nil {
			t.Fatalf("refresh %s: %v", role, err)
		}
		results[role] = result
		for _, skill := range []string{"infrastructure-ops", "deploy-ops"} {
			_, err := os.Stat(filepath.Join(target, ".agents", "skills", skill, "SKILL.md"))
			if role == "ops" && err != nil {
				t.Errorf("ops native bundle omitted %s: %v", skill, err)
			}
			if role == "engineer" && !os.IsNotExist(err) {
				t.Errorf("engineer native bundle leaked %s: %v", skill, err)
			}
		}
	}

	staged := t.TempDir()
	if _, err := project.ProjectScoped(results["ops"].BundleDir, "claude", staged, project.ScopeHome); err != nil {
		t.Fatal(err)
	}
	for _, skill := range []string{"infrastructure-ops", "deploy-ops"} {
		if _, err := os.Stat(filepath.Join(staged, ".claude", "skills", skill, "SKILL.md")); err != nil {
			t.Errorf("staged Ops home omitted %s: %v", skill, err)
		}
	}
	nativeSkills := skillNames(t, filepath.Join(targets["ops"], ".agents", "skills"))
	stagedSkills := skillNames(t, filepath.Join(staged, ".claude", "skills"))
	if !slices.Equal(nativeSkills, stagedSkills) {
		t.Fatalf("native and staged Ops skills differ: native=%v staged=%v", nativeSkills, stagedSkills)
	}

	selectedWhy, err := describe.Why(results["ops"].BundleDir, "skill:infrastructure-ops", describe.Options{})
	if err != nil || !strings.Contains(selectedWhy, "role provider selected because role \"ops\" requests it") {
		t.Fatalf("selected role-provider why = %q, err=%v", selectedWhy, err)
	}
	excludedWhy, err := describe.Why(results["engineer"].BundleDir, "skill:infrastructure-ops", describe.Options{})
	if err != nil || !strings.Contains(excludedWhy, "not selected role \"engineer\"") {
		t.Fatalf("excluded role-provider why = %q, err=%v", excludedWhy, err)
	}
	optionalWhy, err := describe.Why(results["ops"].BundleDir, "source:example--optional", describe.Options{})
	if err != nil || !strings.Contains(optionalWhy, "optional role provider") {
		t.Fatalf("optional missing provider why = %q, err=%v", optionalWhy, err)
	}
}

func TestMissingRequiredRoleProviderFailsExplicitly(t *testing.T) {
	projects := filepath.Join(t.TempDir(), "projects")
	base := filepath.Join(projects, "example", "base")
	missing := filepath.Join(projects, "example", "required")
	writeProvider(t, base, true)
	manifestPath := filepath.Join(t.TempDir(), "mount-eligibility.json")
	writeEligibilityManifest(t, manifestPath, skillmount.Eligibility{
		ProjectsRoot: projects,
		Defaults:     []string{base},
		Harnesses:    map[string][]string{},
		RoleProviders: map[string][]skillmount.RoleProvider{
			"ops": {{Path: missing, Required: true}},
		},
	})
	_, err := Refresh(Options{
		Role:         "ops",
		Harness:      "codex",
		CWD:          projects,
		TargetDir:    t.TempDir(),
		ManifestPath: manifestPath,
		OutDir:       filepath.Join(t.TempDir(), "bundles"),
	})
	if err == nil || !strings.Contains(err.Error(), "required role provider example/required") {
		t.Fatalf("required missing provider error = %v", err)
	}
}
