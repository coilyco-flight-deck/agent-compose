package resolver

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"forgejo.coilysiren.me/coilyco-flight-deck/agent-compose/internal/person"
	"forgejo.coilysiren.me/coilyco-flight-deck/agent-compose/internal/schema"
)

func testPerson() *person.Person {
	return &person.Person{
		Name: "kai",
		Roles: map[string]person.Role{
			"engineer": {Purpose: "Build.", Personalities: []string{"curious", "grounded"}},
			"writer":   {Purpose: "Write.", Personalities: []string{"grounded"}},
		},
		Personalities: map[string]person.Personality{
			"curious":  {Skill: "personality-curious", Color: "#d98e48"},
			"grounded": {Skill: "personality-grounded", Color: "#5fa87a"},
		},
		Raw: []byte("person \"kai\"\n"),
	}
}

func testRequest(delivery, density string) *schema.Request {
	return &schema.Request{Role: "engineer", Delivery: delivery, Density: density}
}

// makeSource lays out a synthetic source on disk: one instruction plus the
// named skills, each with a SKILL.md carrying the given body.
func makeSource(t *testing.T, id string, skillBodies map[string]string) *schema.Source {
	t.Helper()
	root := t.TempDir()
	src := &schema.Source{ID: id, Root: root, Declaration: []byte(id)}
	instr := filepath.Join(root, "foundation.md")
	if err := os.WriteFile(instr, []byte("# Foundation from "+id+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	src.Instructions = []schema.ContentRef{{ID: "foundation-" + id, Path: "foundation.md"}}
	for name, body := range skillBodies {
		dir := filepath.Join(root, name)
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dir, "SKILL.md"), []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
		src.Skills = append(src.Skills, schema.ContentRef{ID: name, Path: name})
	}
	return src
}

func TestResolveSelectsPersonalityAndOrdinarySkills(t *testing.T) {
	src := makeSource(t, "aos", map[string]string{
		"personality-curious":  "# Curious\n",
		"personality-grounded": "# Grounded\n",
		"fixture-review":       "# Review\n",
	})
	res, err := Resolve(testRequest(schema.DeliveryNativeSkills, schema.DensityFull), testPerson(), []*schema.Source{src}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Skills) != 3 ||
		res.Skills[0].ID != "personality-curious" ||
		res.Skills[1].ID != "personality-grounded" ||
		res.Skills[2].ID != "fixture-review" {
		t.Fatalf("expected role personalities followed by ordinary skills, got %+v", res.Skills)
	}
	if res.FavoriteColor == "" {
		t.Fatal("expected a melded favorite color")
	}
	var selected bool
	for _, d := range res.Decisions {
		if d.Subject == "skill:fixture-review" && d.Outcome == OutcomeSelected {
			selected = true
		}
	}
	if !selected {
		t.Fatalf("expected fixture-review selected, decisions: %+v", res.Decisions)
	}
}

func TestResolveComposesOnlyTheActiveRolesSkills(t *testing.T) {
	src := makeSource(t, "aos", map[string]string{
		"personality-curious":  "# Curious\n",
		"personality-grounded": "# Grounded\n",
	})
	for _, name := range []string{"coding-shape-cli", "html-a11y"} {
		dir := filepath.Join(src.Root, name)
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dir, "COMPOSED.md"), []byte("# "+name+"\n"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	src.RoleSkills = map[string][]schema.ContentRef{
		"engineer": {{
			ID: "coding-shape-cli", Path: "coding-shape-cli", EntryPoint: "COMPOSED.md",
		}},
		"writer": {{
			ID: "html-a11y", Path: "html-a11y", EntryPoint: "COMPOSED.md",
		}},
	}

	res, err := Resolve(
		testRequest(schema.DeliveryNativeSkills, schema.DensityFull),
		testPerson(),
		[]*schema.Source{src},
		nil,
	)
	if err != nil {
		t.Fatal(err)
	}
	var found bool
	for _, skill := range res.Skills {
		if skill.ID == "html-a11y" {
			t.Fatal("an inactive role's composed skill entered the resolution")
		}
		if skill.ID == "coding-shape-cli" {
			found = true
			if skill.EntryPoint != "COMPOSED.md" {
				t.Fatalf("composed entry point = %q", skill.EntryPoint)
			}
		}
	}
	if !found {
		t.Fatalf("active role composed skill missing from %+v", res.Skills)
	}
	for _, decision := range res.Decisions {
		if decision.Subject == "skill:html-a11y" {
			t.Fatalf("inactive role skill leaked into the decision trace: %+v", decision)
		}
	}
}

func TestEmbeddedRolePersonalitiesSelectBoundSkills(t *testing.T) {
	p, err := person.Load()
	if err != nil {
		t.Fatal(err)
	}
	skillBodies := make(map[string]string, len(p.Personalities))
	for _, binding := range p.Personalities {
		skillBodies[binding.Skill] = "# Fixture\n"
	}
	src := makeSource(t, "aos", skillBodies)

	for _, roleName := range p.RoleOrder {
		t.Run(roleName, func(t *testing.T) {
			res, err := Resolve(&schema.Request{
				Role:     roleName,
				Delivery: schema.DeliveryNativeSkills,
				Density:  schema.DensityFull,
			}, p, []*schema.Source{src}, nil)
			if err != nil {
				t.Fatal(err)
			}
			rolePersonalities := p.Roles[roleName].Personalities
			if len(res.Skills) != len(rolePersonalities) {
				t.Fatalf("selected %d skills, want %d: %+v", len(res.Skills), len(rolePersonalities), res.Skills)
			}
			for i, personalityName := range rolePersonalities {
				if want := p.Personalities[personalityName].Skill; res.Skills[i].ID != want {
					t.Fatalf("selected skill %d = %q, want %q", i, res.Skills[i].ID, want)
				}
			}
		})
	}
}

func TestResolveValidationFailures(t *testing.T) {
	src := makeSource(t, "aos", map[string]string{
		"personality-curious":  "# Curious\n",
		"personality-grounded": "# Grounded\n",
	})
	p := testPerson()

	if _, err := Resolve(&schema.Request{Role: "pilot", Delivery: schema.DeliveryNativeSkills, Density: schema.DensityFull}, p, []*schema.Source{src}, nil); err == nil || !strings.Contains(err.Error(), "not defined") {
		t.Fatalf("expected unknown role failure, got %v", err)
	}
	broken := testPerson()
	broken.Roles["engineer"] = person.Role{Purpose: "Build.", Personalities: []string{"missing"}}
	if _, err := Resolve(testRequest(schema.DeliveryNativeSkills, schema.DensityFull), broken, []*schema.Source{src}, nil); err == nil || !strings.Contains(err.Error(), "without a catalog binding") {
		t.Fatalf("expected missing catalog binding failure, got %v", err)
	}
	empty := &schema.Source{ID: "empty", Root: t.TempDir()}
	if _, err := Resolve(testRequest(schema.DeliveryNativeSkills, schema.DensityFull), p, []*schema.Source{empty}, nil); err == nil || !strings.Contains(err.Error(), "no admitted source provides it") {
		t.Fatalf("expected missing bound skill failure, got %v", err)
	}
}

func TestResolveShadowsIdenticalAndFailsConflicts(t *testing.T) {
	base := map[string]string{
		"personality-curious":  "# Curious\n",
		"personality-grounded": "# Grounded\n",
	}
	a := makeSource(t, "overlay", base)
	b := makeSource(t, "aos", base)
	res, err := Resolve(testRequest(schema.DeliveryNativeSkills, schema.DensityFull), testPerson(), []*schema.Source{a, b}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if res.Skills[0].Source != "overlay" || res.Skills[1].Source != "overlay" {
		t.Fatalf("expected the higher-precedence copies selected, got %+v", res.Skills)
	}
	var shadowed bool
	for _, d := range res.Decisions {
		if d.Outcome == OutcomeShadowed && d.Source == "aos" {
			shadowed = true
		}
	}
	if !shadowed {
		t.Fatalf("expected the aos copy shadowed, decisions: %+v", res.Decisions)
	}

	c := makeSource(t, "aos", map[string]string{
		"personality-curious":  "# Different\n",
		"personality-grounded": "# Grounded\n",
	})
	if _, err := Resolve(testRequest(schema.DeliveryNativeSkills, schema.DensityFull), testPerson(), []*schema.Source{a, c}, nil); err == nil || !strings.Contains(err.Error(), "conflicts") {
		t.Fatalf("expected non-identical collision failure, got %v", err)
	}
}

func TestCompiledDensityBriefFallsBackWithoutBrief(t *testing.T) {
	src := makeSource(t, "aos", map[string]string{
		"personality-curious":  "# Curious\n",
		"personality-grounded": "# Grounded\n",
	})
	res, err := Resolve(testRequest(schema.DeliveryCompiled, schema.DensityBrief), testPerson(), []*schema.Source{src}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(res.CompiledBodies) != 2 ||
		filepath.Base(res.CompiledBodies[0]) != "SKILL.md" ||
		filepath.Base(res.CompiledBodies[1]) != "SKILL.md" {
		t.Fatalf("expected every personality to fall back to SKILL.md, got %v", res.CompiledBodies)
	}
	var fallback bool
	for _, d := range res.Decisions {
		if d.Outcome == OutcomeFallback {
			fallback = true
		}
	}
	if !fallback {
		t.Fatalf("expected a fallback decision, decisions: %+v", res.Decisions)
	}
}

func TestCompiledDensityBriefPrefersBrief(t *testing.T) {
	src := makeSource(t, "aos", map[string]string{
		"personality-curious":  "# Curious\n",
		"personality-grounded": "# Grounded\n",
	})
	curiousBrief := filepath.Join(src.Root, "personality-curious", "BRIEF.md")
	if err := os.WriteFile(curiousBrief, []byte("Curious, briefly.\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	brief := filepath.Join(src.Root, "personality-grounded", "BRIEF.md")
	if err := os.WriteFile(brief, []byte("Grounded, briefly.\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	res, err := Resolve(testRequest(schema.DeliveryCompiled, schema.DensityBrief), testPerson(), []*schema.Source{src}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(res.CompiledBodies) != 2 ||
		filepath.Base(res.CompiledBodies[0]) != "BRIEF.md" ||
		filepath.Base(res.CompiledBodies[1]) != "BRIEF.md" {
		t.Fatalf("expected BRIEF.md preferred for every personality, got %v", res.CompiledBodies)
	}
}

func TestMissingOptionalSourceLandsInTrace(t *testing.T) {
	src := makeSource(t, "aos", map[string]string{
		"personality-curious":  "# Curious\n",
		"personality-grounded": "# Grounded\n",
	})
	res, err := Resolve(testRequest(schema.DeliveryNativeSkills, schema.DensityFull), testPerson(), []*schema.Source{src},
		[]schema.MissingSource{{ID: "overlay", Reason: "optional source declaration overlay.kdl is absent"}})
	if err != nil {
		t.Fatal(err)
	}
	var noted bool
	for _, d := range res.Decisions {
		if d.Subject == "source:overlay" && d.Outcome == OutcomeExcluded {
			noted = true
		}
	}
	if !noted {
		t.Fatalf("expected the missing optional source excluded in trace, decisions: %+v", res.Decisions)
	}
}
