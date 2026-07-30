package compose

import (
	"encoding/json"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"forgejo.coilysiren.me/coilyco-flight-deck/agent-compose/internal/bundle"
	"forgejo.coilysiren.me/coilyco-flight-deck/agent-compose/internal/color"
	"forgejo.coilysiren.me/coilyco-flight-deck/agent-compose/internal/person"
	"forgejo.coilysiren.me/coilyco-flight-deck/agent-compose/internal/personpolicy"
	"forgejo.coilysiren.me/coilyco-flight-deck/agent-compose/internal/resolver"
	"forgejo.coilysiren.me/coilyco-flight-deck/agent-compose/internal/schema"
)

func fixture(t *testing.T, name string) string {
	t.Helper()
	return filepath.Join("..", "..", "testdata", "contracts", name)
}

func readManifest(t *testing.T, dir string) bundle.Manifest {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join(dir, "manifest.json"))
	if err != nil {
		t.Fatal(err)
	}
	var m bundle.Manifest
	if err := json.Unmarshal(raw, &m); err != nil {
		t.Fatal(err)
	}
	return m
}

func TestComposeAllFixtures(t *testing.T) {
	p, err := person.Load()
	if err != nil {
		t.Fatal(err)
	}
	wantPersonalities := p.Roles["engineer"].Personalities
	var componentColors []string
	for _, personalityName := range wantPersonalities {
		componentColors = append(componentColors, p.Personalities[personalityName].Color)
	}
	wantColor, err := color.Favorite(componentColors)
	if err != nil {
		t.Fatal(err)
	}
	cases := map[string]string{
		"native.kdl":   "native-skills",
		"compiled.kdl": "compiled",
	}
	for name, want := range cases {
		t.Run(name, func(t *testing.T) {
			out := t.TempDir()
			result, err := Run(fixture(t, name), out)
			if err != nil {
				t.Fatal(err)
			}
			m := readManifest(t, result.Bundle.Dir)
			if m.Format != "agent-compose.bundle" || m.Role != "engineer" ||
				m.ModelClass != schema.ModelClassFrontier ||
				!slices.Equal(m.Personalities, wantPersonalities) ||
				m.Delivery.Mode != want {
				t.Fatalf("unexpected manifest: %+v", m)
			}
			if len(m.Sources) != 2 || m.Sources[0] != "roster:core" || m.Sources[1] != "aos-public" {
				t.Fatalf("unexpected sources: %+v", m.Sources)
			}
			if m.Color != wantColor {
				t.Fatalf("manifest color = %q, want melded %q", m.Color, wantColor)
			}
			instructionsPath := filepath.Join(
				result.Bundle.Dir,
				"content",
				"instructions.md",
			)
			instructions, err := os.ReadFile(instructionsPath)
			if err != nil {
				t.Fatal(err)
			}
			instructionText := string(instructions)
			wantCard, err := p.RenderRoleIdentityCard("engineer", wantColor)
			if err != nil {
				t.Fatal(err)
			}
			for _, selected := range []string{
				"# Role instructions",
				"Agent-compose assigned the `engineer` role from the caller's compose request.",
				"The agent treats this caller-assigned role as authoritative and fixed for the session.",
				"The agent does not activate, blend, or adopt another role's briefing or personality set.",
				"If the user requests a role switch, the agent rejects the request and directs the caller " +
					"to launch a new bundle with the different role.",
				wantCard,
				"**Role skill // `role-engineer`**",
				"# Fixture foundation",
			} {
				if !strings.Contains(instructionText, selected) {
					t.Fatalf("instructions missing %q:\n%s", selected, instructionText)
				}
			}
			if strings.Contains(instructionText, p.Roles["engineer"].Briefing) {
				t.Fatalf("native startup instructions eagerly embedded the role skill body:\n%s", instructionText)
			}
			for _, roleName := range p.RoleOrder {
				if roleName == "engineer" {
					continue
				}
				if strings.Contains(instructionText, p.Roles[roleName].Briefing) {
					t.Fatalf("inactive role %q briefing entered the bundle:\n%s", roleName, instructionText)
				}
			}
			mustExist(t, result.Bundle.Dir, "trace.json")
			mustExist(t, result.Bundle.Dir, "content/skills/roster%3Acore/role-engineer/SKILL.md")
			for _, personalityName := range wantPersonalities {
				skillPath := "content/skills/roster%3Acore/personality-" + personalityName + "/SKILL.md"
				mustExist(t, result.Bundle.Dir, skillPath)
			}
			if want == "compiled" {
				if m.Delivery.CompiledContext != "delivery/compiled.md" || m.Delivery.SkillsRoot != "" {
					t.Fatalf("unexpected compiled delivery: %+v", m.Delivery)
				}
				mustExist(t, result.Bundle.Dir, "delivery/compiled.md")
				compiled, err := os.ReadFile(filepath.Join(result.Bundle.Dir, "delivery", "compiled.md"))
				if err != nil {
					t.Fatal(err)
				}
				if !strings.Contains(string(compiled), p.Roles["engineer"].Briefing) ||
					!strings.Contains(string(compiled), wantCard) {
					t.Fatalf("compiled context omitted role card or skill body:\n%s", compiled)
				}
			} else {
				if m.Delivery.SkillsRoot != "content/skills" || m.Delivery.CompiledContext != "" {
					t.Fatalf("unexpected native delivery: %+v", m.Delivery)
				}
			}
		})
	}
}

func TestComposeExternalPersonDoesNotInheritEmbeddedRoster(t *testing.T) {
	result, err := Run(fixture(t, "custom-person.kdl"), t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	if !result.ExternalOnly {
		t.Fatal("portable external-only request lost its policy")
	}
	manifest := readManifest(t, result.Bundle.Dir)
	if manifest.Role != "builder" ||
		!slices.Equal(manifest.Personalities, []string{"bright", "steady"}) {
		t.Fatalf("external manifest identity = %+v", manifest)
	}
	if !slices.Equal(manifest.Sources, []string{"person:workbench", "aos-public"}) {
		t.Fatalf("external manifest sources = %+v", manifest.Sources)
	}
	if slices.Contains(manifest.Sources, "roster:core") {
		t.Fatal("external composition inherited roster:core")
	}
	for _, rel := range []string{
		"content/skills/person%3Aworkbench/personality-bright/SKILL.md",
		"content/skills/person%3Aworkbench/personality-steady/SKILL.md",
	} {
		mustExist(t, result.Bundle.Dir, rel)
	}
	instructions, err := os.ReadFile(filepath.Join(result.Bundle.Dir, "content", "instructions.md"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(instructions), "# Workbench invariant") ||
		strings.Contains(string(instructions), "# Personality invariant") {
		t.Fatalf("external instructions crossed person boundaries:\n%s", instructions)
	}
}

func TestComposeHostExternalOnlySelection(t *testing.T) {
	personRoot, err := filepath.Abs(fixture(t, "person-independent"))
	if err != nil {
		t.Fatal(err)
	}
	result, err := RunWithOptions(
		fixture(t, "custom-person-inherited.kdl"),
		t.TempDir(),
		Options{
			PersonPolicy: personpolicy.ExternalOnly,
			PersonSource: personRoot,
		},
	)
	if err != nil {
		t.Fatal(err)
	}
	manifest := readManifest(t, result.Bundle.Dir)
	if !result.ExternalOnly ||
		!slices.Equal(manifest.Sources, []string{"person:workbench", "aos-public"}) {
		t.Fatalf("guarded host selection = %+v / %+v", result, manifest.Sources)
	}
}

func TestComposeHostExternalOnlyRequiresSource(t *testing.T) {
	_, err := RunWithOptions(
		fixture(t, "native.kdl"),
		t.TempDir(),
		Options{PersonPolicy: personpolicy.ExternalOnly},
	)
	if err == nil || !IsExternalOnlyError(err) ||
		!strings.Contains(err.Error(), "requires person_source") {
		t.Fatalf("external-only missing source error = %v", err)
	}
}

func TestComposeLowContextPrunesOnlyOptedOutSkills(t *testing.T) {
	result, err := Run(fixture(t, "low-context.kdl"), t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	manifest := readManifest(t, result.Bundle.Dir)
	if manifest.ModelClass != schema.ModelClassLowContext {
		t.Fatalf("model class = %q", manifest.ModelClass)
	}
	ordinary := filepath.Join(
		result.Bundle.Dir,
		"content",
		"skills",
		"aos-public",
		"fixture-review",
	)
	if _, err := os.Stat(ordinary); !os.IsNotExist(err) {
		t.Fatalf("low-context optional skill entered the bundle: %v", err)
	}
	for _, personalityName := range manifest.Personalities {
		mustExist(
			t,
			result.Bundle.Dir,
			"content/skills/roster%3Acore/personality-"+personalityName+"/SKILL.md",
		)
	}
	var excluded bool
	for _, decision := range result.Resolution.Decisions {
		if decision.Subject == "skill:fixture-review" &&
			decision.Outcome == resolver.OutcomeExcluded {
			excluded = true
		}
	}
	if !excluded {
		t.Fatalf("low-context exclusion missing from trace: %+v", result.Resolution.Decisions)
	}
}

func TestComposeInferredProviderRoot(t *testing.T) {
	p, err := person.Load()
	if err != nil {
		t.Fatal(err)
	}
	root := t.TempDir()
	ordinary := filepath.Join(root, ".agents", "skills", "coding-go")
	if err := os.MkdirAll(ordinary, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(ordinary, "SKILL.md"), []byte("# Go\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	for name, body := range map[string]string{
		"coding-shape-cli": "# CLI foundation\n",
		"design-system":    "# Design system\n",
	} {
		dir := filepath.Join(root, ".agents", "composed", name)
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dir, "COMPOSED.md"), []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.WriteFile(filepath.Join(root, ".agents", "roles.kdl"), []byte(`roles {
    role "engineer" {
        composed-skill "coding-shape-cli"
    }
    role "design" {
        composed-skill "design-system"
    }
}
`), 0o644); err != nil {
		t.Fatal(err)
	}
	request := filepath.Join(root, "request.kdl")
	if err := os.WriteFile(request, []byte(`compose {
    role "engineer"
    delivery "native-skills"
    source "aos-public" root="." required=#true
}
`), 0o644); err != nil {
		t.Fatal(err)
	}

	result, err := Run(request, t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	manifest := readManifest(t, result.Bundle.Dir)
	if len(manifest.Personalities) != len(p.Roles["engineer"].Personalities) {
		t.Fatalf("inferred provider selected the wrong personalities: %+v", manifest.Personalities)
	}
	instructions, err := os.ReadFile(filepath.Join(result.Bundle.Dir, "content", "instructions.md"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(instructions), "# Personality invariant") {
		t.Fatalf("composition omitted the embedded personality invariant:\n%s", instructions)
	}
	for _, rel := range []string{
		"content/skills/roster%3Acore/personality-curious/SKILL.md",
		"content/skills/aos-public/coding-go/SKILL.md",
		"content/skills/aos-public/coding-shape-cli/SKILL.md",
	} {
		mustExist(t, result.Bundle.Dir, rel)
	}
	composedSource := filepath.Join(
		result.Bundle.Dir,
		"content",
		"skills",
		"aos-public",
		"coding-shape-cli",
		"COMPOSED.md",
	)
	if _, err := os.Stat(composedSource); !os.IsNotExist(err) {
		t.Fatalf("source-only COMPOSED.md leaked into the bundle: %v", err)
	}
	inactive := filepath.Join(result.Bundle.Dir, "content", "skills", "aos-public", "design-system")
	if _, err := os.Stat(inactive); !os.IsNotExist(err) {
		t.Fatalf("inactive role skill leaked into the bundle: %v", err)
	}
}

func TestCompiledDeliveryUsesCanonicalProse(t *testing.T) {
	out := t.TempDir()
	result, err := Run(fixture(t, "compiled.kdl"), out)
	if err != nil {
		t.Fatal(err)
	}
	body, err := os.ReadFile(filepath.Join(result.Bundle.Dir, "delivery", "compiled.md"))
	if err != nil {
		t.Fatal(err)
	}
	for _, heading := range []string{"# Curious", "# Grounded", "# Meticulous"} {
		if !strings.Contains(string(body), heading) {
			t.Fatalf("compiled prose missing %q:\n%s", heading, body)
		}
	}
}

func TestDesignerPageExperienceBoundaryMatchesNativeAndCompiledDelivery(t *testing.T) {
	provider := t.TempDir()
	skillDir := filepath.Join(provider, ".agents", "skills", "fixture")
	if err := os.MkdirAll(skillDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(
		filepath.Join(skillDir, "SKILL.md"),
		[]byte("---\nname: fixture\ndescription: Fixture capability.\n---\n\n# Fixture\n"),
		0o644,
	); err != nil {
		t.Fatal(err)
	}
	p, err := person.Load()
	if err != nil {
		t.Fatal(err)
	}
	want := p.Roles["design"].Briefing
	for _, delivery := range []string{
		schema.DeliveryNativeSkills,
		schema.DeliveryCompiled,
	} {
		t.Run(delivery, func(t *testing.T) {
			result, err := RunRoots(
				&schema.Request{
					Role:       "design",
					Delivery:   delivery,
					ModelClass: schema.ModelClassFrontier,
				},
				[]RootSource{{ID: "fixture", Root: provider}},
				t.TempDir(),
				Options{},
			)
			if err != nil {
				t.Fatal(err)
			}
			var path string
			if delivery == schema.DeliveryCompiled {
				path = filepath.Join(result.Bundle.Dir, "delivery", "compiled.md")
			} else {
				path = filepath.Join(
					result.Bundle.Dir,
					"content",
					"skills",
					"roster%3Acore",
					"role-design",
					"SKILL.md",
				)
			}
			raw, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			if !strings.Contains(string(raw), want) {
				t.Fatalf("%s delivery omitted the canonical Designer page-experience boundary:\n%s",
					delivery, raw)
			}
		})
	}
}

func TestRepeatedRunsReuseWithoutRewriting(t *testing.T) {
	out := t.TempDir()
	first, err := Run(fixture(t, "native.kdl"), out)
	if err != nil {
		t.Fatal(err)
	}
	if first.Bundle.Reused {
		t.Fatal("first run must materialize, not reuse")
	}
	manifest := filepath.Join(first.Bundle.Dir, "manifest.json")
	before, err := os.Stat(manifest)
	if err != nil {
		t.Fatal(err)
	}

	second, err := Run(fixture(t, "native.kdl"), out)
	if err != nil {
		t.Fatal(err)
	}
	if !second.Bundle.Reused || second.Bundle.Dir != first.Bundle.Dir || second.Bundle.Key != first.Bundle.Key {
		t.Fatalf("expected cache reuse, got %+v then %+v", first.Bundle, second.Bundle)
	}
	after, err := os.Stat(manifest)
	if err != nil {
		t.Fatal(err)
	}
	if !before.ModTime().Equal(after.ModTime()) {
		t.Fatal("reuse must not rewrite bundle files")
	}
}

func TestDifferentDeliveriesGetDifferentBundles(t *testing.T) {
	out := t.TempDir()
	a, err := Run(fixture(t, "native.kdl"), out)
	if err != nil {
		t.Fatal(err)
	}
	b, err := Run(fixture(t, "compiled.kdl"), out)
	if err != nil {
		t.Fatal(err)
	}
	if a.Bundle.Key == b.Bundle.Key {
		t.Fatal("different deliveries must produce different bundle keys")
	}
}

func TestFailedFinalizeLeavesNoPartialBundle(t *testing.T) {
	out := t.TempDir()
	probe, err := Run(fixture(t, "native.kdl"), t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	blocker := filepath.Join(out, probe.Bundle.Key)
	if err := os.WriteFile(blocker, []byte("occupied"), 0o644); err != nil {
		t.Fatal(err)
	}

	if _, err := Run(fixture(t, "native.kdl"), out); err == nil {
		t.Fatal("expected finalize to fail when the target path is blocked")
	}
	entries, err := os.ReadDir(out)
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range entries {
		if strings.HasPrefix(e.Name(), ".stage-") {
			t.Fatalf("staging directory leaked: %s", e.Name())
		}
	}
	raw, err := os.ReadFile(blocker)
	if err != nil || string(raw) != "occupied" {
		t.Fatalf("blocked target must be untouched, got %q / %v", raw, err)
	}
}

func TestTraceRecordsDecisions(t *testing.T) {
	out := t.TempDir()
	result, err := Run(fixture(t, "native.kdl"), out)
	if err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile(filepath.Join(result.Bundle.Dir, "trace.json"))
	if err != nil {
		t.Fatal(err)
	}
	var trace bundle.Trace
	if err := json.Unmarshal(raw, &trace); err != nil {
		t.Fatal(err)
	}
	if trace.Format != "agent-compose.trace" || len(trace.Decisions) == 0 {
		t.Fatalf("unexpected trace: %+v", trace)
	}
	outcomes := map[string]bool{}
	for _, d := range trace.Decisions {
		if d.Reason == "" {
			t.Fatalf("decision without a reason: %+v", d)
		}
		outcomes[d.Outcome] = true
	}
	for _, want := range []string{"selected", "delivered"} {
		if !outcomes[want] {
			t.Fatalf("trace missing %q outcome: %+v", want, trace.Decisions)
		}
	}
}

func mustExist(t *testing.T, root, rel string) {
	t.Helper()
	if _, err := os.Stat(filepath.Join(root, rel)); err != nil {
		t.Fatalf("expected %s in bundle: %v", rel, err)
	}
}
