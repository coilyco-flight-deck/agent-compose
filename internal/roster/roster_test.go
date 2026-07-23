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
				Personalities: []string{"bright", "pending"},
				Seats: []person.Seat{
					{Harness: "claude", Name: "opal builder", Pronouns: "she"},
					{Harness: "codex", Name: "terran builder", Pronouns: "he"},
				},
			},
			"seatless": {
				Purpose:       "Remain seatless.",
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
		"If you are claude running the builder role: your name is opal builder (pronouns: she).",
		"If you are codex running the builder role: your name is terran builder (pronouns: he).",
		"bright (favorite color #c87945), defined in [bright](personalities/bright.md)",
		"## builder - Build the fixture.",
		"Melded personalities: bright (favorite color #c87945), defined in [bright](personalities/bright.md); pending (favorite color #7d9fd3), definition pending.",
		"Melded favorite color: " + melded,
	} {
		if !strings.Contains(table, want) {
			t.Fatalf("table missing %q:\n%s", want, table)
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
