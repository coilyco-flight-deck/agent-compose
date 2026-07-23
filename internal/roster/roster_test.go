package roster

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"forgejo.coilysiren.me/coilyco-flight-deck/agent-compose/internal/person"
	"forgejo.coilysiren.me/coilyco-flight-deck/agent-compose/internal/project"
	"forgejo.coilysiren.me/coilyco-flight-deck/agent-compose/internal/schema"
)

func loadInputs(t *testing.T) (*person.Person, []*schema.Source) {
	t.Helper()
	p, err := person.Load()
	if err != nil {
		t.Fatal(err)
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
	for _, want := range []string{
		"If you are claude running the engineer role: your name is opal engineer (pronouns: she).",
		"If you are codex running the director role: your name is solar director (pronouns: he).",
		"curious (favorite color #d98e48), defined in [curious](personalities/curious.md)",
		"## director - Pair with the human on high level goals.",
		"Compatible personalities: bold, definition pending; grounded (favorite color #5fa87a), defined in [grounded](personalities/grounded.md); diplomatic, definition pending.",
	} {
		if !strings.Contains(table, want) {
			t.Fatalf("table missing %q:\n%s", want, table)
		}
	}
	for _, role := range []string{"designer", "social", "sales", "customer-success"} {
		if strings.Contains(table, "## "+role+" ") {
			t.Fatalf("seatless role %q must not render a section", role)
		}
	}

	override := string(files["AGENTS.claude.md"])
	if !strings.Contains(override, "@/opt/artifact/personalities/curious.md") {
		t.Fatalf("claude override missing mechanical import:\n%s", override)
	}

	if !strings.Contains(string(files["personalities/curious.md"]), "# Curious") {
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
	if !strings.Contains(table, "curious (favorite color #d98e48), definition pending") {
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
	if _, err := project.ApplyOwned(target, files, "roster", "person:kai"); err != nil {
		t.Fatal(err)
	}
	if _, err := project.ApplyOwned(target, files, "roster", "person:kai"); err != nil {
		t.Fatalf("re-apply over owned files must succeed: %v", err)
	}

	foreign := t.TempDir()
	if err := os.WriteFile(filepath.Join(foreign, "AGENTS.COMPOSE.md"), []byte("mine\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := project.ApplyOwned(foreign, files, "roster", "person:kai"); err == nil || !strings.Contains(err.Error(), "refusing to overwrite") {
		t.Fatalf("expected ownership refusal, got %v", err)
	}
}
