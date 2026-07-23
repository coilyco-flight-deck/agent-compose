package roster

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"forgejo.coilysiren.me/coilyco-flight-deck/agent-compose/internal/color"
	"forgejo.coilysiren.me/coilyco-flight-deck/agent-compose/internal/person"
	"forgejo.coilysiren.me/coilyco-flight-deck/agent-compose/internal/project"
	"forgejo.coilysiren.me/coilyco-flight-deck/agent-compose/internal/schema"
)

func loadInputs(t *testing.T) (*person.Person, []*schema.Source) {
	t.Helper()
	p := &person.Person{
		Name: "fixture",
		Roles: map[string]person.Role{
			"builder": {
				Purpose:       "Build the fixture.",
				Briefing:      "You are a builder. Build the fixture from repository evidence.\n\nFinish validation and return a complete result.",
				Personalities: []string{"bright", "pending"},
				Seats: []person.Seat{
					{Harness: "claude", Name: "opal builder", Pronouns: "she"},
					{Harness: "codex", Name: "terran builder", Pronouns: "he"},
				},
			},
			"seatless": {
				Purpose:       "Remain seatless.",
				Briefing:      "You are seatless.\n\nRemain outside the rendered dispatch table.",
				Personalities: []string{"bright"},
			},
		},
		RoleOrder: []string{"builder", "seatless"},
		Personalities: map[string]person.Personality{
			"bright":  {Skill: "personality-curious", Color: "#c87945"},
			"pending": {Skill: "personality-pending", Color: "#7d9fd3"},
		},
	}
	src, err := schema.LoadSource(filepath.Join("..", "..", "testdata", "contracts", "source-public.kdl"))
	if err != nil {
		t.Fatal(err)
	}
	return p, []*schema.Source{src}
}

func TestRenderDispatchTable(t *testing.T) {
	p, sources := loadInputs(t)
	files, err := Render(p, sources, "/opt/artifact")
	if err != nil {
		t.Fatal(err)
	}

	table := string(files["AGENTS.COMPOSE.md"])
	melded, err := color.Favorite([]string{
		p.Personalities["bright"].Color,
		p.Personalities["pending"].Color,
	})
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{
		"# Fixture foundation",
		"The agent uses repository evidence and reports uncertainty explicitly.",
		"Each agent loads every linked definition on that role's Melded personalities",
		"If you are claude running the builder role: your name is opal builder (pronouns: she).",
		"If you are codex running the builder role: your name is terran builder (pronouns: he).",
		"bright (favorite color #c87945), defined in [bright](personalities/bright.md)",
		"## builder - Build the fixture.",
		"You are a builder. Build the fixture from repository evidence.\n\nFinish validation and return a complete result.",
		"Melded personalities: bright (favorite color #c87945), defined in [bright](personalities/bright.md); pending (favorite color #7d9fd3), definition pending.",
		"Melded favorite color: " + melded,
	} {
		if !strings.Contains(table, want) {
			t.Fatalf("table missing %q:\n%s", want, table)
		}
	}
	ordered := []string{
		"## builder - Build the fixture.",
		"You are a builder.",
		"If you are claude running the builder role",
		"Melded personalities:",
	}
	for i := 1; i < len(ordered); i++ {
		if strings.Index(table, ordered[i-1]) >= strings.Index(table, ordered[i]) {
			t.Fatalf("roster content is out of order: %q must precede %q", ordered[i-1], ordered[i])
		}
	}
	if strings.Contains(table, "## seatless ") {
		t.Fatal("seatless role must not render a section")
	}

	override := string(files["AGENTS.claude.md"])
	if !strings.Contains(override, "@/opt/artifact/personalities/bright.md") {
		t.Fatalf("claude override missing mechanical import:\n%s", override)
	}

	if !strings.Contains(string(files["personalities/bright.md"]), "# Curious") {
		t.Fatal("personality body must carry the skill definition")
	}
}

func TestRenderRejectsMissingInstruction(t *testing.T) {
	p, _ := loadInputs(t)
	sources := []*schema.Source{{
		ID:   "broken",
		Root: t.TempDir(),
		Instructions: []schema.ContentRef{{
			ID:   "personality-invariant",
			Path: "missing.md",
		}},
	}}
	if _, err := Render(p, sources, "/opt/artifact"); err == nil ||
		!strings.Contains(err.Error(), `instruction "personality-invariant"`) {
		t.Fatalf("missing provider instruction must fail clearly, got %v", err)
	}
}

func TestRenderIsDeterministic(t *testing.T) {
	p, sources := loadInputs(t)
	first, err := Render(p, sources, "/opt/artifact")
	if err != nil {
		t.Fatal(err)
	}
	second, err := Render(p, sources, "/opt/artifact")
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(first, second) {
		t.Fatal("render must be deterministic")
	}
}

func TestRenderDegradesWithoutSources(t *testing.T) {
	p, _ := loadInputs(t)
	files, err := Render(p, nil, "/opt/artifact")
	if err != nil {
		t.Fatal(err)
	}
	table := string(files["AGENTS.COMPOSE.md"])
	if !strings.Contains(table, "bright (favorite color #c87945), definition pending") {
		t.Fatalf("expected pending definitions without sources:\n%s", table)
	}
	if _, ok := files["AGENTS.claude.md"]; ok {
		t.Fatal("no bodies means no claude import override")
	}
}

func TestApplyOwnedProtectsForeignFiles(t *testing.T) {
	p, sources := loadInputs(t)
	target := t.TempDir()
	files, err := Render(p, sources, target)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := project.ApplyOwned(target, files, "roster", "person:fixture"); err != nil {
		t.Fatal(err)
	}
	if _, err := project.ApplyOwned(target, files, "roster", "person:fixture"); err != nil {
		t.Fatalf("re-apply over owned files must succeed: %v", err)
	}

	foreign := t.TempDir()
	if err := os.WriteFile(filepath.Join(foreign, "AGENTS.COMPOSE.md"), []byte("mine\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := project.ApplyOwned(foreign, files, "roster", "person:fixture"); err == nil || !strings.Contains(err.Error(), "refusing to overwrite") {
		t.Fatalf("expected ownership refusal, got %v", err)
	}
}
