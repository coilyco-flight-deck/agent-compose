package roster

import (
	"encoding/json"
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
			"pending": {Skill: "personality-reflective", Color: "#7d9fd3"},
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
		"# Personality invariant",
		"# Fixture foundation",
		"The agent uses repository evidence and reports uncertainty explicitly.",
		"If launch context\nassigns a role, the agent treats that assignment as authoritative.",
		"an unassigned native agent uses the initial substantive request as a soft\nsignal",
		"Later task shape does not change the role.",
		"An eligible inferred native role changes only under the explicit policy below.",
		"Before acting, each agent loads the selected role skill and every personality",
		"# Builder",
		"Build the fixture.",
		"**Role skill // `role-builder`**",
		"**Favorite color // `" + melded + "`**",
		"// claude: opal builder (she)",
		"// codex: terran builder (he)",
		"###  Bright",
		"**#c87945 //  //  // **",
		"* `role-builder`",
		"* `personality-curious`",
	} {
		if !strings.Contains(table, want) {
			t.Fatalf("table missing %q:\n%s", want, table)
		}
	}
	builderSection := renderedCard(t, table, "Builder")
	ordered := []string{
		"# Builder",
		"**Role skill // `role-builder`**",
		"## Personality meld",
		"## Active doctrine",
	}
	for i := 1; i < len(ordered); i++ {
		if strings.Index(builderSection, ordered[i-1]) >= strings.Index(builderSection, ordered[i]) {
			t.Fatalf("roster content is out of order: %q must precede %q", ordered[i-1], ordered[i])
		}
	}
	if strings.Contains(table, "## seatless ") {
		t.Fatal("seatless role must not render a section")
	}

	override := string(files["AGENTS.claude.md"])
	if strings.Contains(override, "@") ||
		!strings.Contains(override, "installs role and personality definitions as skills") {
		t.Fatalf("claude override eagerly imports identity doctrine:\n%s", override)
	}

	if !strings.Contains(string(files[".agents/skills/personality-curious/SKILL.md"]), "# Curious") {
		t.Fatal("personality body must carry the skill definition")
	}
	if !strings.Contains(string(files[".agents/skills/role-builder/SKILL.md"]), "You are a builder.") {
		t.Fatal("role skill must carry the legacy briefing through the adapter")
	}
	var snapshot person.Snapshot
	if err := json.Unmarshal(files["person.json"], &snapshot); err != nil {
		t.Fatal(err)
	}
	if snapshot.Format != person.SnapshotFormat ||
		!reflect.DeepEqual(snapshot.RoleOrder, p.RoleOrder) ||
		len(snapshot.Roles) != len(p.Roles) {
		t.Fatalf("roster omitted the complete person snapshot: %+v", snapshot)
	}
}

func TestRenderNativeInteractiveAdaptationPolicy(t *testing.T) {
	p, sources := loadInputs(t)
	files, err := Render(p, sources, "/opt/artifact")
	if err != nil {
		t.Fatal(err)
	}

	table := string(files["AGENTS.COMPOSE.md"])
	for _, want := range []string{
		"### Native interactive adaptation",
		"Only an unwarded native agent in a directly steered interactive session",
		"Ward-bound, composed, staged, containerized, headless",
		"while an explicit slash goal is active",
		"At session start, the agent records whether its role was caller-assigned or\ninferred",
		"Only an inferred native role is eligible\nto switch.",
		"Available role slugs: `builder`.",
		"case-insensitive spelling such as `QA` for `qa`",
		"switches the inferred role\nwithout a second confirmation",
		"loads the target role skill and every\nskill in its complete ordered personality meld before acting",
		"stops acting from the prior role charter",
		"The switched role remains\ninferred.",
		"persists until another explicit switch or session end",
		"switch again or return to an earlier role through another request",
		"If the agent proposes a role switch",
		"Should the agent switch to it now?",
		"When a requested slug is unavailable, the\nagent rejects the switch and lists the available role slugs.",
		"The\nharness, model, tools, permissions, credentials, and executable authority do\nnot change.",
		"A caller-assigned role cannot switch.",
		"directs the caller to launch a new bundle with the different role",
		"#### Personality-only swaps",
		"A personality-only swap does not\nchange the active role, obligations, permissions, or authority.",
		"This task would benefit from the <X> persona because <reason>. Should the agent\nswap to it now?",
		"The task request itself does not count as confirmation.",
		"Task completion restores the role's default meld.",
		"Each later\nswap needs a new proposal and confirmation.",
		"#### Conditional QA fixture authority",
		"runtime explicitly\nlaunches it in an enforced fixture mode",
	} {
		if !strings.Contains(table, want) {
			t.Fatalf("native interactive swap policy missing %q:\n%s", want, table)
		}
	}
	if strings.Contains(table, "Available role slugs: `builder`, `seatless`") {
		t.Fatal("seatless roles must not become native switch targets")
	}
}

func TestRenderDefaultSupportsAdvisorToQASwitch(t *testing.T) {
	p, err := person.Load()
	if err != nil {
		t.Fatal(err)
	}
	files, err := Render(p, nil, "/opt/artifact")
	if err != nil {
		t.Fatal(err)
	}

	table := string(files["AGENTS.COMPOSE.md"])
	var available []string
	for _, roleName := range p.RoleOrder {
		if len(p.Roles[roleName].Seats) > 0 {
			available = append(available, "`"+roleName+"`")
		}
	}
	wantTargets := "Available role slugs: " + strings.Join(available, ", ") + "."
	if !strings.Contains(table, wantTargets) {
		t.Fatalf("default native switch targets drifted:\nwant %s\n\n%s", wantTargets, table)
	}
	for _, roleName := range []string{"advisor", "qa"} {
		role := p.Roles[roleName]
		if strings.Contains(table, role.Briefing) {
			t.Fatalf("startup roster eagerly embedded role %q briefing:\n%s", roleName, table)
		}
		if !strings.Contains(string(files[".agents/skills/"+p.RoleSkillID(roleName)+"/SKILL.md"]), role.Briefing) {
			t.Fatalf("role %q skill omitted its briefing", roleName)
		}
	}

	qaSection := renderedCard(t, table, "Qa")
	ordered := []string{
		"# Qa",
		"**Role skill // `role-qa`**",
		"## Personality meld",
		"* `role-qa`",
		"* `personality-meticulous`",
		"* `personality-candid`",
		"* `personality-playful`",
	}
	for i := 1; i < len(ordered); i++ {
		if strings.Index(qaSection, ordered[i-1]) >= strings.Index(qaSection, ordered[i]) {
			t.Fatalf("QA activation content is out of order: %q must precede %q", ordered[i-1], ordered[i])
		}
	}
	if !strings.Contains(string(files[".agents/skills/role-qa/SKILL.md"]), "runtime explicitly grants fixture mode") {
		t.Fatal("QA role skill omitted conditional fixture doctrine")
	}
}

func renderedCard(t *testing.T, table, title string) string {
	t.Helper()
	start := strings.Index(table, "\n# "+title+"\n")
	if start < 0 {
		t.Fatalf("rendered roster has no %q card", title)
	}
	section := table[start+1:]
	if next := strings.Index(section, "\n# "); next >= 0 {
		section = section[:next]
	}
	return section
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

func TestRenderUsesEmbeddedDefinitionsWithoutSources(t *testing.T) {
	p, _ := loadInputs(t)
	files, err := Render(p, nil, "/opt/artifact")
	if err != nil {
		t.Fatal(err)
	}
	table := string(files["AGENTS.COMPOSE.md"])
	if !strings.Contains(table, "**#c87945 //") ||
		!strings.Contains(table, "**#7d9fd3 //") {
		t.Fatalf("expected compact identity metadata without sources:\n%s", table)
	}
	if _, ok := files["AGENTS.claude.md"]; !ok {
		t.Fatal("roster must produce the lazy-load claude bootstrap")
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
