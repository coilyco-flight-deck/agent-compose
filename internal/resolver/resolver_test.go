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
			"engineer": {
				Purpose:       "Build.",
				Briefing:      "You are an engineer.\n\nFinish the complete repository workflow.",
				Personalities: []string{"curious", "grounded"},
			},
			"writer": {
				Purpose:       "Write.",
				Briefing:      "You are a writer.\n\nDeliver the finished text.",
				Personalities: []string{"grounded"},
			},
		},
		Personalities: map[string]person.Personality{
			"curious":  {Skill: "personality-curious", Color: "#d98e48"},
			"grounded": {Skill: "personality-grounded", Color: "#5fa87a"},
		},
		Raw: []byte("person \"kai\"\n"),
	}
}

func testRequest(delivery string) *schema.Request {
	return &schema.Request{
		Role: "engineer", Delivery: delivery, ModelClass: schema.ModelClassFrontier,
	}
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
		"fixture-review": "# Review\n",
	})
	res, err := Resolve(testRequest(schema.DeliveryNativeSkills), testPerson(), []*schema.Source{src}, nil)
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
	if res.RoleBriefing != testPerson().Roles["engineer"].Briefing {
		t.Fatalf("role briefing = %q", res.RoleBriefing)
	}
	var selected, briefingSelected bool
	for _, d := range res.Decisions {
		if d.Subject == "skill:fixture-review" && d.Outcome == OutcomeSelected {
			selected = true
		}
		if d.Subject == "instruction:role-briefing" &&
			d.Source == "person:kai" &&
			d.Outcome == OutcomeSelected {
			briefingSelected = true
		}
	}
	if !selected {
		t.Fatalf("expected fixture-review selected, decisions: %+v", res.Decisions)
	}
	if !briefingSelected {
		t.Fatalf("expected role briefing selected, decisions: %+v", res.Decisions)
	}
}

func TestResolveLetsSkillsOptOutOfLowContextModels(t *testing.T) {
	src := makeSource(t, "aos", map[string]string{
		"coding-python": `---
name: coding-python
description: Core Python guidance.
low-context: required
---

# Python
`,
		"architecture": `---
name: architecture
description: High-end architecture guidance.
low-context: optional
---

# Architecture
`,
	})
	low := testRequest(schema.DeliveryNativeSkills)
	low.ModelClass = schema.ModelClassLowContext
	res, err := Resolve(low, testPerson(), []*schema.Source{src}, nil)
	if err != nil {
		t.Fatal(err)
	}
	selected := map[string]bool{}
	for _, skill := range res.Skills {
		selected[skill.ID] = true
	}
	if !selected["coding-python"] || selected["architecture"] {
		t.Fatalf("low-context selection = %v", selected)
	}
	var excluded bool
	for _, decision := range res.Decisions {
		if decision.Subject == "skill:architecture" &&
			decision.Outcome == OutcomeExcluded &&
			strings.Contains(decision.Reason, "optional for low-context") {
			excluded = true
		}
	}
	if !excluded {
		t.Fatalf("trace omitted low-context exclusion: %+v", res.Decisions)
	}

	frontier, err := Resolve(
		testRequest(schema.DeliveryNativeSkills),
		testPerson(),
		[]*schema.Source{src},
		nil,
	)
	if err != nil {
		t.Fatal(err)
	}
	selected = map[string]bool{}
	for _, skill := range frontier.Skills {
		selected[skill.ID] = true
	}
	if !selected["coding-python"] || !selected["architecture"] {
		t.Fatalf("frontier selection = %v", selected)
	}
}

func TestResolveRejectsUnsupportedRoleModelClass(t *testing.T) {
	p, err := person.Load()
	if err != nil {
		t.Fatal(err)
	}
	lowContext := &schema.Request{
		Role:       "ceo",
		Delivery:   schema.DeliveryNativeSkills,
		ModelClass: schema.ModelClassLowContext,
	}
	if _, err := Resolve(lowContext, p, nil, nil); err == nil ||
		err.Error() != `role "ceo" requires a frontier model` {
		t.Fatalf("low-context CEO error = %v", err)
	}

	frontier := *lowContext
	frontier.ModelClass = schema.ModelClassFrontier
	if _, err := Resolve(&frontier, p, nil, nil); err != nil {
		t.Fatalf("frontier CEO failed: %v", err)
	}
}

func TestResolveComposesOnlyTheActiveRolesSkills(t *testing.T) {
	src := makeSource(t, "aos", nil)
	for _, name := range []string{"coding-shape-cli", "design-system"} {
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
			ID: "design-system", Path: "design-system", EntryPoint: "COMPOSED.md",
		}},
	}

	res, err := Resolve(
		testRequest(schema.DeliveryNativeSkills),
		testPerson(),
		[]*schema.Source{src},
		nil,
	)
	if err != nil {
		t.Fatal(err)
	}
	var found bool
	for _, skill := range res.Skills {
		if skill.ID == "design-system" {
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
		if decision.Subject == "skill:design-system" {
			t.Fatalf("inactive role skill leaked into the decision trace: %+v", decision)
		}
	}
}

func TestEmbeddedRolePersonalitiesSelectBoundSkills(t *testing.T) {
	p, err := person.Load()
	if err != nil {
		t.Fatal(err)
	}
	for _, roleName := range p.RoleOrder {
		t.Run(roleName, func(t *testing.T) {
			res, err := Resolve(&schema.Request{
				Role:     roleName,
				Delivery: schema.DeliveryNativeSkills,
			}, p, nil, nil)
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
	src := makeSource(t, "aos", nil)
	p := testPerson()

	if _, err := Resolve(&schema.Request{Role: "pilot", Delivery: schema.DeliveryNativeSkills}, p, []*schema.Source{src}, nil); err == nil || !strings.Contains(err.Error(), "not defined") {
		t.Fatalf("expected unknown role failure, got %v", err)
	}
	broken := testPerson()
	broken.Roles["engineer"] = person.Role{Purpose: "Build.", Personalities: []string{"missing"}}
	if _, err := Resolve(testRequest(schema.DeliveryNativeSkills), broken, []*schema.Source{src}, nil); err == nil || !strings.Contains(err.Error(), "without a catalog binding") {
		t.Fatalf("expected missing catalog binding failure, got %v", err)
	}
	missingDefinition := testPerson()
	missingDefinition.Personalities["curious"] = person.Personality{
		Skill: "personality-absent", Color: "#d98e48",
	}
	if _, err := Resolve(testRequest(schema.DeliveryNativeSkills), missingDefinition, nil, nil); err == nil ||
		!strings.Contains(err.Error(), `binds non-canonical skill "personality-absent"`) {
		t.Fatalf("expected missing embedded definition failure, got %v", err)
	}
}

func TestResolveShadowsIdenticalAndFailsConflicts(t *testing.T) {
	legacy, err := person.Source(testPerson())
	if err != nil {
		t.Fatal(err)
	}
	legacy.ID = "aos"
	res, err := Resolve(testRequest(schema.DeliveryNativeSkills), testPerson(), []*schema.Source{legacy}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if res.Skills[0].Source != "person:kai" || res.Skills[1].Source != "person:kai" {
		t.Fatalf("expected embedded copies selected, got %+v", res.Skills)
	}
	var shadowed bool
	for _, d := range res.Decisions {
		if d.Outcome == OutcomeShadowed && d.Source == legacy.ID {
			shadowed = true
		}
	}
	if !shadowed {
		t.Fatalf("expected the aos copy shadowed, decisions: %+v", res.Decisions)
	}

	conflict := makeSource(t, "aos", map[string]string{
		"personality-curious": "# Different\n",
	})
	if _, err := Resolve(testRequest(schema.DeliveryNativeSkills), testPerson(), []*schema.Source{conflict}, nil); err == nil || !strings.Contains(err.Error(), "conflicts") {
		t.Fatalf("expected non-identical collision failure, got %v", err)
	}
}

func TestCompiledDeliveryUsesCanonicalSkillBodies(t *testing.T) {
	res, err := Resolve(testRequest(schema.DeliveryCompiled), testPerson(), nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(res.CompiledBodies) != 2 ||
		res.CompiledBodies[0].EntryPoint != "SKILL.md" ||
		res.CompiledBodies[1].EntryPoint != "SKILL.md" {
		t.Fatalf("expected canonical SKILL.md for every personality, got %v", res.CompiledBodies)
	}
}

func TestMissingOptionalSourceLandsInTrace(t *testing.T) {
	src := makeSource(t, "aos", nil)
	res, err := Resolve(testRequest(schema.DeliveryNativeSkills), testPerson(), []*schema.Source{src},
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
