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
	"forgejo.coilysiren.me/coilyco-flight-deck/agent-compose/internal/resolver"
	"forgejo.coilysiren.me/coilyco-flight-deck/agent-compose/internal/schema"
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

func providerReport(t *testing.T, result *Result, source string) resolver.ProviderReport {
	t.Helper()
	for _, provider := range result.Composition.Resolution.Providers {
		if provider.Source == source {
			return provider
		}
	}
	t.Fatalf("provider report omitted %s: %+v", source, result.Composition.Resolution.Providers)
	return resolver.ProviderReport{}
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
			if result.ModelTier != schema.ModelTierFrontier {
				t.Fatalf("default model tier = %q, want frontier", result.ModelTier)
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

func TestRefreshDefaultsToFrontierModelTierAndClass(t *testing.T) {
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
	if result.ModelTier != schema.ModelTierFrontier {
		t.Fatalf("default model tier = %q, want frontier", result.ModelTier)
	}
}

func TestRefreshAcceptsCanonicalModelTierAndRejectsUnknownTier(t *testing.T) {
	projects := filepath.Join(t.TempDir(), "projects")
	provider := filepath.Join(projects, "example", "provider")
	writeProvider(t, provider, true)
	manifest := filepath.Join(t.TempDir(), "mount-eligibility.json")
	writeManifest(t, manifest, projects, provider)
	options := Options{
		Role:         "design",
		Harness:      "codex",
		ModelTier:    schema.ModelTierCommodity,
		CWD:          projects,
		TargetDir:    t.TempDir(),
		ManifestPath: manifest,
		OutDir:       filepath.Join(t.TempDir(), "bundles"),
	}

	result, err := Refresh(options)
	if err != nil {
		t.Fatal(err)
	}
	if result.ModelTier != schema.ModelTierCommodity ||
		result.Composition.Resolution.Request.ModelTier != schema.ModelTierCommodity {
		t.Fatalf("commodity model tier not retained: %+v", result)
	}

	options.ModelTier = "budget"
	_, err = Refresh(options)
	if err == nil || !strings.Contains(err.Error(), "unsupported native model tier") {
		t.Fatalf("unknown model-tier error = %v", err)
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
	writeOrdinarySkill(t, base, "repo-infrastructure")
	writeOrdinarySkill(t, base, "repo-deploy")
	writeOrdinarySkill(t, infrastructure, "infrastructure-ops")
	writeOrdinarySkill(t, infrastructure, "repo-infrastructure")
	writeOrdinarySkill(t, deploy, "deploy-ops")
	writeOrdinarySkill(t, deploy, "repo-deploy")
	manifestPath := filepath.Join(t.TempDir(), "mount-eligibility.json")
	writeEligibilityManifest(t, manifestPath, skillmount.Eligibility{
		ProjectsRoot: projects,
		Defaults:     []string{base},
		Harnesses:    map[string][]string{},
		RoleProviders: map[string][]skillmount.RoleProvider{
			"ops": {
				{Path: infrastructure, Required: true, Name: "infrastructure", DeclaredBy: "example/aosk"},
				{Path: deploy, Required: true, Name: "deploy", DeclaredBy: "example/aosk"},
				{Path: missingOptional, Name: "optional", DeclaredBy: "example/aosk"},
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

	for _, role := range []string{"ops", "engineer"} {
		staged := t.TempDir()
		if _, err := project.ProjectScoped(results[role].BundleDir, "claude", staged, project.ScopeHome); err != nil {
			t.Fatal(err)
		}
		for _, skill := range []string{"infrastructure-ops", "deploy-ops"} {
			_, err := os.Stat(filepath.Join(staged, ".claude", "skills", skill, "SKILL.md"))
			if role == "ops" && err != nil {
				t.Errorf("staged Ops home omitted %s: %v", skill, err)
			}
			if role == "engineer" && !os.IsNotExist(err) {
				t.Errorf("staged Engineer home leaked %s: %v", skill, err)
			}
		}
		nativeSkills := skillNames(t, filepath.Join(targets[role], ".agents", "skills"))
		stagedSkills := skillNames(t, filepath.Join(staged, ".claude", "skills"))
		if !slices.Equal(nativeSkills, stagedSkills) {
			t.Fatalf("native and staged %s skills differ: native=%v staged=%v", role, nativeSkills, stagedSkills)
		}
	}

	for _, result := range results {
		baseReport := providerReport(t, result, "example--base")
		if baseReport.Category != resolver.ProviderCategoryCatalogue ||
			baseReport.Scope != "default" ||
			baseReport.Outcome != resolver.OutcomeSelected ||
			baseReport.Skills == 0 || baseReport.ContextBytes == 0 {
			t.Fatalf("ordinary catalogue report = %+v", baseReport)
		}
	}
	for _, source := range []string{"example--infrastructure", "example--deploy"} {
		selected := providerReport(t, results["ops"], source)
		if selected.Category != resolver.ProviderCategoryRole ||
			selected.Scope != "role" ||
			selected.Outcome != resolver.OutcomeSelected ||
			selected.Skills != 1 || selected.ContextBytes == 0 || selected.ApproximateTokens == 0 {
			t.Fatalf("selected role provider %s report = %+v", source, selected)
		}
		excluded := providerReport(t, results["engineer"], source)
		if excluded.Category != resolver.ProviderCategoryRole ||
			excluded.Scope != "role" ||
			excluded.Outcome != resolver.OutcomeExcluded ||
			excluded.Skills != 0 || excluded.ContextBytes != 0 || excluded.ApproximateTokens != 0 {
			t.Fatalf("excluded role provider %s report = %+v", source, excluded)
		}
	}

	selectedWhy, err := describe.Why(results["ops"].BundleDir, "skill:infrastructure-ops", describe.Options{})
	if err != nil || !strings.Contains(
		selectedWhy,
		"role \"ops\" -> provider \"infrastructure\" declared by example/aosk -> selected catalogue",
	) {
		t.Fatalf("selected role-provider why = %q, err=%v", selectedWhy, err)
	}
	excludedWhy, err := describe.Why(results["engineer"].BundleDir, "skill:infrastructure-ops", describe.Options{})
	if err != nil || !strings.Contains(excludedWhy, "not selected role \"engineer\"") {
		t.Fatalf("excluded role-provider why = %q, err=%v", excludedWhy, err)
	}
	if !strings.Contains(excludedWhy, "provider: role-provider/role") ||
		!strings.Contains(excludedWhy, "context: 0 skills, 0 bytes, approximately 0 tokens") {
		t.Fatalf("excluded role-provider budget missing from why output: %q", excludedWhy)
	}
	optionalWhy, err := describe.Why(results["ops"].BundleDir, "source:example--optional", describe.Options{})
	if err != nil || !strings.Contains(optionalWhy, "optional role provider") {
		t.Fatalf("optional missing provider why = %q, err=%v", optionalWhy, err)
	}
	pointerWhy, err := describe.Why(
		results["ops"].BundleDir,
		"skill:repo-infrastructure",
		describe.Options{},
	)
	if err != nil ||
		!strings.Contains(pointerWhy, "provider: catalogue/default") ||
		!strings.Contains(pointerWhy, "provider: role-provider/role") ||
		!strings.Contains(pointerWhy, "outcome: shadowed") {
		t.Fatalf("ordinary pointer and role-provider provenance = %q, err=%v", pointerWhy, err)
	}
	described, err := describe.Bundle(results["engineer"].BundleDir, describe.Options{All: true})
	if err != nil ||
		!strings.Contains(described, "example--infrastructure") ||
		!strings.Contains(described, "(role-provider/role)") ||
		!strings.Contains(described, "0 skills · 0 bytes · ~0 tokens") {
		t.Fatalf("provider-aware describe output = %q, err=%v", described, err)
	}
}

func TestRoleProviderSelectorsMatchAcrossNativeAndStagedHarnesses(t *testing.T) {
	projects := filepath.Join(t.TempDir(), "projects")
	base := filepath.Join(projects, "example", "base")
	hardware := filepath.Join(projects, "example", "hardware")
	writeProvider(t, base, true)
	writeProvider(t, hardware, false)
	for _, skill := range []string{
		"compute-stack",
		"home-power-strip",
		"machine-alpha",
		"machine-beta",
	} {
		writeOrdinarySkill(t, hardware, skill)
	}
	manifestPath := filepath.Join(t.TempDir(), "mount-eligibility.json")
	writeEligibilityManifest(t, manifestPath, skillmount.Eligibility{
		ProjectsRoot: projects,
		Defaults:     []string{base},
		Harnesses:    map[string][]string{},
		RoleProviders: map[string][]skillmount.RoleProvider{
			"design": {{
				Path:     hardware,
				Required: true,
				Skills:   []string{"compute-stack", "machine-*"},
			}},
		},
	})

	cases := map[string]struct {
		modelClass   string
		nativeSkills string
		stagedSkills string
	}{
		"claude":   {"frontier", ".claude/skills", ".claude/skills"},
		"codex":    {"frontier", ".agents/skills", ".agents/skills"},
		"goose":    {"low-context", ".agents/skills", ".agents/skills"},
		"opencode": {"low-context", ".agents/skills", ".agents/skills"},
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
				ManifestPath: manifestPath,
				OutDir:       filepath.Join(t.TempDir(), "bundles"),
			})
			if err != nil {
				t.Fatal(err)
			}
			staged := t.TempDir()
			if _, err := project.ProjectScoped(result.BundleDir, harness, staged, project.ScopeHome); err != nil {
				t.Fatal(err)
			}
			for _, root := range []string{
				filepath.Join(target, tc.nativeSkills),
				filepath.Join(staged, tc.stagedSkills),
			} {
				for _, skill := range []string{"compute-stack", "machine-alpha", "machine-beta"} {
					if _, err := os.Stat(filepath.Join(root, skill, "SKILL.md")); err != nil {
						t.Errorf("%s omitted selected %s: %v", root, skill, err)
					}
				}
				for _, skill := range []string{"home-power-strip"} {
					if _, err := os.Stat(filepath.Join(root, skill, "SKILL.md")); !os.IsNotExist(err) {
						t.Errorf("%s leaked excluded %s: %v", root, skill, err)
					}
				}
			}
			report := providerReport(t, result, "example--hardware")
			if report.Skills != 3 || report.ContextBytes == 0 ||
				!strings.Contains(report.SelectorReason, "admitted 3 of 5 catalogue skills") {
				t.Fatalf("selected-slice provider report = %+v", report)
			}
			why, err := describe.Why(result.BundleDir, "skill:home-power-strip", describe.Options{})
			if err != nil || !strings.Contains(why, "matched no configured selector pattern") ||
				!strings.Contains(why, "selector: ordinary skill selector") ||
				!strings.Contains(why, "context: 3 skills") {
				t.Fatalf("selector exclusion why = %q, err=%v", why, err)
			}
		})
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
