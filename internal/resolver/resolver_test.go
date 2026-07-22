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
		Personalities: map[string]string{
			"curious":  "personality-curious",
			"grounded": "personality-grounded",
		},
		Raw: []byte("person \"kai\"\n"),
	}
}

func testRequest(personality, delivery, density string) *schema.Request {
	return &schema.Request{Role: "engineer", Personality: personality, Delivery: delivery, Density: density}
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

func TestResolveSelectsBoundSkillAndExcludesOthers(t *testing.T) {
	src := makeSource(t, "aos", map[string]string{
		"personality-curious": "# Curious\n",
		"fixture-review":      "# Review\n",
	})
	res, err := Resolve(testRequest("curious", schema.DeliveryNativeSkills, schema.DensityFull), testPerson(), []*schema.Source{src}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if res.Skill.ID != "personality-curious" {
		t.Fatalf("expected bound skill selected, got %+v", res.Skill)
	}
	var excluded bool
	for _, d := range res.Decisions {
		if d.Subject == "skill:fixture-review" && d.Outcome == OutcomeExcluded {
			excluded = true
		}
	}
	if !excluded {
		t.Fatalf("expected fixture-review excluded, decisions: %+v", res.Decisions)
	}
}

func TestResolveValidationFailures(t *testing.T) {
	src := makeSource(t, "aos", map[string]string{"personality-curious": "# Curious\n"})
	p := testPerson()

	if _, err := Resolve(&schema.Request{Role: "pilot", Personality: "curious", Delivery: schema.DeliveryNativeSkills, Density: schema.DensityFull}, p, []*schema.Source{src}, nil); err == nil || !strings.Contains(err.Error(), "not defined") {
		t.Fatalf("expected unknown role failure, got %v", err)
	}
	if _, err := Resolve(testRequest("energetic", schema.DeliveryNativeSkills, schema.DensityFull), p, []*schema.Source{src}, nil); err == nil || !strings.Contains(err.Error(), "not defined") {
		t.Fatalf("expected unknown personality failure, got %v", err)
	}
	if _, err := Resolve(&schema.Request{Role: "writer", Personality: "curious", Delivery: schema.DeliveryNativeSkills, Density: schema.DensityFull}, p, []*schema.Source{src}, nil); err == nil || !strings.Contains(err.Error(), "does not pair") {
		t.Fatalf("expected pairing failure, got %v", err)
	}
	empty := &schema.Source{ID: "empty", Root: t.TempDir()}
	if _, err := Resolve(testRequest("curious", schema.DeliveryNativeSkills, schema.DensityFull), p, []*schema.Source{empty}, nil); err == nil || !strings.Contains(err.Error(), "no admitted source provides it") {
		t.Fatalf("expected missing bound skill failure, got %v", err)
	}
}

func TestResolveShadowsIdenticalAndFailsConflicts(t *testing.T) {
	a := makeSource(t, "overlay", map[string]string{"personality-curious": "# Curious\n"})
	b := makeSource(t, "aos", map[string]string{"personality-curious": "# Curious\n"})
	res, err := Resolve(testRequest("curious", schema.DeliveryNativeSkills, schema.DensityFull), testPerson(), []*schema.Source{a, b}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if res.Skill.Source != "overlay" {
		t.Fatalf("expected the higher-precedence copy selected, got %+v", res.Skill)
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

	c := makeSource(t, "aos", map[string]string{"personality-curious": "# Different\n"})
	if _, err := Resolve(testRequest("curious", schema.DeliveryNativeSkills, schema.DensityFull), testPerson(), []*schema.Source{a, c}, nil); err == nil || !strings.Contains(err.Error(), "conflicts") {
		t.Fatalf("expected non-identical collision failure, got %v", err)
	}
}

func TestCompiledDensityBriefFallsBackWithoutBrief(t *testing.T) {
	src := makeSource(t, "aos", map[string]string{"personality-curious": "# Curious\n"})
	res, err := Resolve(testRequest("curious", schema.DeliveryCompiled, schema.DensityBrief), testPerson(), []*schema.Source{src}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if filepath.Base(res.CompiledBody) != "SKILL.md" {
		t.Fatalf("expected SKILL.md fallback, got %s", res.CompiledBody)
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
	src := makeSource(t, "aos", map[string]string{"personality-grounded": "# Grounded\n"})
	brief := filepath.Join(src.Root, "personality-grounded", "BRIEF.md")
	if err := os.WriteFile(brief, []byte("Grounded, briefly.\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	res, err := Resolve(testRequest("grounded", schema.DeliveryCompiled, schema.DensityBrief), testPerson(), []*schema.Source{src}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if filepath.Base(res.CompiledBody) != "BRIEF.md" {
		t.Fatalf("expected BRIEF.md preferred, got %s", res.CompiledBody)
	}
}

func TestMissingOptionalSourceLandsInTrace(t *testing.T) {
	src := makeSource(t, "aos", map[string]string{"personality-curious": "# Curious\n"})
	res, err := Resolve(testRequest("curious", schema.DeliveryNativeSkills, schema.DensityFull), testPerson(), []*schema.Source{src},
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
