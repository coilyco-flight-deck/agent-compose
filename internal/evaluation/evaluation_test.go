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

func TestBuildEmitsCoreV2ThreeTierMatrix(t *testing.T) {
	pack, err := Build("engineer", "codex")
	if err != nil {
		t.Fatal(err)
	}
	if pack.Format != Format || pack.Role != "engineer" ||
		pack.Seat.Harness != "codex" || pack.Seat.Name == "" {
		t.Fatalf("evaluation identity = %+v", pack)
	}
	if len(pack.Cases) != 21 {
		t.Fatalf("evaluation cases = %d, want 21", len(pack.Cases))
	}
	if !reflect.DeepEqual(pack.DisabledModelTiers, []string{commodityTier, ossTier}) {
		t.Fatalf("disabled model tiers = %v, want [%s %s]", pack.DisabledModelTiers, commodityTier, ossTier)
	}
	if err := ValidateCorePack(pack); err != nil {
		t.Fatal(err)
	}
	for _, evalCase := range pack.Cases {
		wantClass := schema.ModelClassFrontier
		if evalCase.ModelTier == ossTier {
			wantClass = schema.ModelClassLowContext
		}
		if evalCase.BundleModelClass != wantClass ||
			evalCase.Scenario == "" ||
			evalCase.ScenarioKind == "" {
			t.Errorf("case %q has incomplete v2 identity: %+v", evalCase.ID, evalCase)
		}
		wantRubric := 4
		if evalCase.ScenarioKind == ScenarioHumanCommunication {
			wantRubric = 5
		}
		if len(evalCase.Rubric) != wantRubric || !strings.Contains(evalCase.Prompt, "Do not name") {
			t.Errorf("case %q is not review-ready: %+v", evalCase.ID, evalCase)
		}
	}
}

func TestBuildCorePacksValidatesEveryRoleAndAdjacentPair(t *testing.T) {
	packs, err := BuildCorePacks("codex")
	if err != nil {
		t.Fatal(err)
	}
	if len(packs) != 8 {
		t.Fatalf("Core Roster packs = %d, want 8", len(packs))
	}
	for _, pack := range packs {
		communication := caseForScenarioKind(t, pack, frontierTier, ScenarioHumanCommunication)
		if err := validateHumanCommunicationHardFail(communication); err != nil {
			t.Errorf("role %q communication case: %v", pack.Role, err)
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
	if len(generic.Cases) != 6 || generic.Cases[0].ID != "frontier-role-understanding" {
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
	if len(pack.Personalities) == 0 || pack.Invariant == "" || pack.MeldedFavoriteColor == "" {
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
		if !strings.Contains(evalCase.Prompt, "Do not") {
			t.Errorf("community case %q omits its evidence boundary: %q", evalCase.ID, evalCase.Prompt)
		}
	}
	rolePrompt := caseForScenarioKind(t, pack, frontierTier, ScenarioMissionFit).Prompt
	for _, want := range []string{
		"#welcome",
		"verified orientation",
		"community-state record",
		"factual handoff",
	} {
		if !strings.Contains(rolePrompt, want) {
			t.Errorf("community role prompt omitted %q: %q", want, rolePrompt)
		}
	}
	personalityPrompt := caseForScenarioKind(t, pack, frontierTier, ScenarioPersonality).Prompt
	for _, want := range []string{
		"welcomed newcomers",
		"may be outdated",
		"smallest verification step",
		"factual handoff",
	} {
		if !strings.Contains(personalityPrompt, want) {
			t.Errorf("community personality prompt omitted %q: %q", want, personalityPrompt)
		}
	}
}

func TestBuildUsesDesignerPageExperienceBoundaryCases(t *testing.T) {
	pack, err := Build("design", "codex")
	if err != nil {
		t.Fatal(err)
	}
	rolePrompt := caseForScenarioKind(t, pack, frontierTier, ScenarioAuthorityBoundary).Prompt
	for _, want := range []string{
		"/orgs and three static organization detail routes",
		"places ./orgs in its top navigation",
		"static public catalog",
		"graphical React view while consuming unchanged props and state",
		"authenticated dashboard that fetches account data at runtime",
		"stateful form workflow with validation and submission",
		"terminal UI",
		"procedural game surface generated by simulation logic",
		"visual-only and page-experience effect tests",
		"focused routes and navigation can qualify while routing-system architecture cannot",
		"local verification and resolved delivery follow-through",
	} {
		if !strings.Contains(rolePrompt, want) {
			t.Errorf("Designer role prompt omitted %q: %q", want, rolePrompt)
		}
	}
	if caseForScenarioKind(t, pack, ossTier, ScenarioAuthorityBoundary).Prompt != rolePrompt {
		t.Fatal("Designer frontier and OSS role cases exercise different boundaries")
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
		"Disabled model tiers: `commodity`, `oss`",
		"## Scenario matrix (21 cases)",
		"### frontier-mission-repository-proof",
		"### oss-personality-small-inconsistency",
	} {
		if !strings.Contains(markdown, want) {
			t.Errorf("Markdown evaluation omitted %q", want)
		}
	}
}

func caseForScenarioKind(t *testing.T, pack *Pack, tier, kind string) Case {
	t.Helper()
	for _, evalCase := range pack.Cases {
		if evalCase.ModelTier == tier && evalCase.ScenarioKind == kind {
			return evalCase
		}
	}
	t.Fatalf("pack %q has no %s/%s case", pack.Role, tier, kind)
	return Case{}
}
