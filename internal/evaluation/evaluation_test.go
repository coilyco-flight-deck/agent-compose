package evaluation

import (
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"forgejo.coilysiren.me/coilyco-flight-deck/agent-compose/internal/person"
	"forgejo.coilysiren.me/coilyco-flight-deck/agent-compose/internal/schema"
	"gopkg.in/yaml.v3"
)

func TestBuildEmitsFourCaseFrontierOSSMatrix(t *testing.T) {
	pack, err := Build("engineer", "codex")
	if err != nil {
		t.Fatal(err)
	}
	if pack.Format != Format || pack.Role != "engineer" ||
		pack.Seat.Harness != "codex" || pack.Seat.Name == "" {
		t.Fatalf("evaluation identity = %+v", pack)
	}
	if len(pack.Cases) != 4 {
		t.Fatalf("evaluation cases = %d, want 4", len(pack.Cases))
	}
	want := []struct {
		id         string
		tier       string
		modelClass string
		dimension  string
	}{
		{"frontier-role-understanding", "frontier", schema.ModelClassFrontier, "role-understanding"},
		{"frontier-personality-expression", "frontier", schema.ModelClassFrontier, "personality-expression"},
		{"oss-role-understanding", "oss", schema.ModelClassLowContext, "role-understanding"},
		{"oss-personality-expression", "oss", schema.ModelClassLowContext, "personality-expression"},
	}
	for i, expected := range want {
		got := pack.Cases[i]
		if got.ID != expected.id || got.ModelTier != expected.tier ||
			got.BundleModelClass != expected.modelClass || got.Dimension != expected.dimension {
			t.Errorf("case %d = %+v, want %+v", i, got, expected)
		}
		if len(got.Rubric) != 4 || !strings.Contains(got.Prompt, "Do not name") {
			t.Errorf("case %q is not review-ready: %+v", got.ID, got)
		}
	}
}

func TestBuildUsesPortfolioNativeEmbeddedCases(t *testing.T) {
	p, err := person.Load()
	if err != nil {
		t.Fatal(err)
	}
	for _, roleName := range p.RoleOrder {
		role := p.Roles[roleName]
		pack, buildErr := Build(roleName, role.Seats[0].Harness)
		if buildErr != nil {
			t.Fatalf("build %q: %v", roleName, buildErr)
		}
		rolePrompt := pack.Cases[0].Prompt
		if rolePrompt == "" || strings.Contains(rolePrompt, "improve the company") {
			t.Errorf("%q role prompt is not portfolio-native: %q", roleName, rolePrompt)
		}
		if strings.Contains(pack.Cases[1].Prompt, "The team has") {
			t.Errorf("%q personality prompt retained fictional-team framing", roleName)
		}
		for _, personality := range pack.Personalities {
			if strings.Contains(personality.Definition, "The agent") ||
				strings.Contains(personality.Definition, "the agent") {
				t.Errorf("%q personality %q retains third-person self-instruction", roleName, personality.Name)
			}
		}
	}
}

func TestBuildForExternalPersonUsesOnlySelectedPackage(t *testing.T) {
	p, err := person.LoadDirectory(filepath.Join(
		"..", "..", "testdata", "contracts", "person-independent",
	))
	if err != nil {
		t.Fatal(err)
	}
	pack, err := BuildFor(p, "builder", "codex")
	if err != nil {
		t.Fatal(err)
	}
	if pack.Person != "workbench" || pack.Role != "builder" ||
		pack.Seat.Name != "workbench builder" {
		t.Fatalf("external evaluation identity = %+v", pack)
	}
	if !strings.Contains(pack.Invariant, "Workbench invariant") ||
		strings.Contains(pack.Invariant, "Personality invariant") {
		t.Fatalf("external evaluation invariant crossed source boundaries: %q", pack.Invariant)
	}
	if len(pack.Personalities) != 2 ||
		pack.Personalities[0].Name != "bright" ||
		pack.Personalities[1].Name != "steady" {
		t.Fatalf("external evaluation personalities = %+v", pack.Personalities)
	}
	if strings.Contains(pack.Cases[0].Prompt, "Kai") ||
		strings.Contains(pack.Cases[0].Prompt, "the company") {
		t.Fatalf("external evaluation inherited embedded domain framing: %q", pack.Cases[0].Prompt)
	}
}

func TestBuildForCustomRoleReplacesCompleteMatrix(t *testing.T) {
	p, err := person.LoadDirectoryWithLibraries(
		filepath.Join("..", "..", "examples", "person-profile"),
		filepath.Join("..", "..", "examples", "shared-personality-library"),
	)
	if err != nil {
		t.Fatal(err)
	}
	custom, err := BuildFor(p, "bulk-captioner", "chatbot-sonnet-low")
	if err != nil {
		t.Fatal(err)
	}
	if len(custom.RunProtocol) != 2 ||
		custom.ReviewRule.PassingTotal != 2 ||
		len(custom.Cases) != 1 ||
		custom.Cases[0].ID != "chatbot-guidance" {
		t.Fatalf("custom matrix was field-merged with generic data: %+v", custom)
	}
	if custom.Seat.Selector() != "chatbot-sonnet-low" ||
		custom.Seat.Pronouns != "they" ||
		custom.CopyContract == nil ||
		custom.CopyContract.Source == "" ||
		!strings.HasPrefix(custom.CopyContract.Digest, "sha256:") {
		t.Fatalf("custom evaluation context is incomplete: %+v", custom)
	}

	generic, err := BuildFor(p, "caption-review", "chatbot-sonnet-low")
	if err != nil {
		t.Fatal(err)
	}
	if len(generic.Cases) != 4 || generic.Cases[0].ID != "frontier-role-understanding" {
		t.Fatalf("role without a custom matrix did not receive generic fallback: %+v", generic.Cases)
	}
}

func TestParseProfileMatrixRejectsIncompleteAndUnknownAssets(t *testing.T) {
	for name, raw := range map[string]string{
		"missing review rule": `
run_protocol: [Run.]
cases:
  - id: fixture
    model_tier: frontier
    bundle_model_class: frontier
    dimension: role
    prompt: Respond.
    reviewer_question: Is it sound?
    rubric:
      - id: sound
        question: Is it sound?
        scale: {"2": Yes., "1": Partly., "0": No.}
`,
		"unknown field": "role_prompt: One.\npersonality_prompt: Two.\nunknown: true\n",
	} {
		t.Run(name, func(t *testing.T) {
			if _, err := parseProfileMatrix([]byte(raw)); err == nil {
				t.Fatal("malformed profile matrix passed")
			}
		})
	}
}

func TestBuildCarriesSelfContainedReviewContext(t *testing.T) {
	pack, err := Build("ops", "codex")
	if err != nil {
		t.Fatal(err)
	}
	if paragraphs := strings.Count(pack.Briefing, "\n\n") + 1; paragraphs < 3 {
		t.Fatalf("ops briefing has %d paragraphs", paragraphs)
	}
	if len(pack.Personalities) != 3 || pack.Invariant == "" || pack.MeldedFavoriteColor == "" {
		t.Fatalf("evaluation context is incomplete: %+v", pack)
	}
	if !strings.Contains(pack.Invariant, "use first person for your own actions") {
		t.Fatalf("evaluation invariant omits first-person self-reference: %q", pack.Invariant)
	}
	for _, evalCase := range pack.Cases {
		if !strings.Contains(evalCase.Prompt, "first person for your own actions") {
			t.Errorf("case %q omits first-person self-reference: %q", evalCase.ID, evalCase.Prompt)
		}
	}
	for _, personality := range pack.Personalities {
		if personality.Definition == "" || strings.HasPrefix(personality.Definition, "---") {
			t.Errorf("personality definition was not normalized: %+v", personality)
		}
	}
}

func TestBuildUsesDiscordNativeCommunityCases(t *testing.T) {
	pack, err := Build("community", "discord")
	if err != nil {
		t.Fatal(err)
	}
	if pack.Seat.Name != "siren community host" {
		t.Fatalf("community seat = %+v", pack.Seat)
	}
	for _, evalCase := range pack.Cases {
		if !strings.Contains(evalCase.Prompt, "public") {
			t.Errorf("community case %q is not member-facing: %q", evalCase.ID, evalCase.Prompt)
		}
		if !strings.Contains(evalCase.Prompt, "Do not") {
			t.Errorf("community case %q omits its evidence boundary: %q", evalCase.ID, evalCase.Prompt)
		}
	}
	rolePrompt := pack.Cases[0].Prompt
	for _, want := range []string{
		"#welcome",
		"private plan",
		"text-only planning surface",
		"member's guess as confirmation",
		"no event schedule, resolving source, or staff contact",
	} {
		if !strings.Contains(rolePrompt, want) {
			t.Errorf("community role prompt omitted %q: %q", want, rolePrompt)
		}
	}
	personalityPrompt := pack.Cases[1].Prompt
	for _, want := range []string{
		"welcoming newcomers",
		"may be outdated",
		"account handle",
		"propose a check, not a change",
		"Do not offer or promise a later edit",
	} {
		if !strings.Contains(personalityPrompt, want) {
			t.Errorf("community personality prompt omitted %q: %q", want, personalityPrompt)
		}
	}
}

func TestBuildReviewMinimumsReferenceCaseCriteria(t *testing.T) {
	pack, err := Build("ops", "codex")
	if err != nil {
		t.Fatal(err)
	}
	if pack.ReviewRule.PassingTotal <= 0 || pack.ReviewRule.PassingTotal > 8 {
		t.Fatalf("passing total is outside the rubric range: %+v", pack.ReviewRule)
	}
	if !pack.ReviewRule.RejectZeroScores {
		t.Fatal("strict review must reject a missing criterion")
	}
	for _, evalCase := range pack.Cases {
		minimums := pack.ReviewRule.RoleMinimumScores
		if evalCase.Dimension == "personality-expression" {
			minimums = pack.ReviewRule.PersonalityMinimumScores
		}
		if len(minimums) == 0 {
			t.Fatalf("case %q has no dimension minimums", evalCase.ID)
		}
		criteria := make(map[string]bool, len(evalCase.Rubric))
		for _, criterion := range evalCase.Rubric {
			criteria[criterion.ID] = true
		}
		for criterion, minimum := range minimums {
			if !criteria[criterion] {
				t.Errorf("case %q minimum names unknown criterion %q", evalCase.ID, criterion)
			}
			if minimum < 0 || minimum > 2 {
				t.Errorf("case %q minimum for %q is outside score range: %d", evalCase.ID, criterion, minimum)
			}
		}
	}
}

func TestSeatSelectionChangesIdentityNotDoctrineOrAuthority(t *testing.T) {
	claude, err := Build("engineer", "claude")
	if err != nil {
		t.Fatal(err)
	}
	codex, err := Build("engineer", "codex")
	if err != nil {
		t.Fatal(err)
	}
	if claude.Seat.Selector() == codex.Seat.Selector() ||
		claude.Seat.Name == codex.Seat.Name {
		t.Fatalf("fixture seats are not distinct: %+v %+v", claude.Seat, codex.Seat)
	}
	if claude.RoleSkill != codex.RoleSkill ||
		claude.RoleSkillDigest != codex.RoleSkillDigest ||
		claude.Briefing != codex.Briefing ||
		!reflect.DeepEqual(claude.Personalities, codex.Personalities) ||
		!reflect.DeepEqual(claude.RunProtocol, codex.RunProtocol) ||
		!reflect.DeepEqual(claude.ReviewRule, codex.ReviewRule) ||
		!reflect.DeepEqual(claude.Cases, codex.Cases) {
		t.Fatal("seat selection changed role doctrine or evaluation authority boundary")
	}
}

func TestBuildRejectsUnknownRoleOrSeat(t *testing.T) {
	if _, err := Build("missing", "codex"); err == nil || !strings.Contains(err.Error(), "is not defined") {
		t.Fatalf("unknown role error = %v", err)
	}
	if _, err := Build("engineer", "missing"); err == nil || !strings.Contains(err.Error(), "has no") {
		t.Fatalf("unknown seat error = %v", err)
	}
}

func TestYAMLAndMarkdownAreDeterministic(t *testing.T) {
	pack, err := Build("engineer", "codex")
	if err != nil {
		t.Fatal(err)
	}
	first, err := MarshalYAML(pack)
	if err != nil {
		t.Fatal(err)
	}
	second, err := MarshalYAML(pack)
	if err != nil {
		t.Fatal(err)
	}
	if string(first) != string(second) {
		t.Fatal("YAML evaluation pack is not deterministic")
	}
	var decoded Pack
	if err := yaml.Unmarshal(first, &decoded); err != nil {
		t.Fatalf("decode evaluation YAML: %v", err)
	}
	markdown := string(Markdown(pack))
	for _, want := range []string{
		"# Agent-compose behavior evaluation",
		"## Four-case matrix",
		"### frontier-role-understanding",
		"### oss-personality-expression",
	} {
		if !strings.Contains(markdown, want) {
			t.Errorf("Markdown evaluation omitted %q", want)
		}
	}
}
