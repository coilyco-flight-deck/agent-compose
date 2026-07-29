// Package evaluation renders deterministic human-review packs for role and
// personality behavior across frontier and OSS model tiers.
package evaluation

import (
	"bytes"
	_ "embed"
	"fmt"
	"path/filepath"
	"strings"

	"forgejo.coilysiren.me/coilyco-flight-deck/agent-compose/internal/color"
	"forgejo.coilysiren.me/coilyco-flight-deck/agent-compose/internal/person"
	"gopkg.in/yaml.v3"
)

//go:embed assets/generic.yaml
var genericMatrixAsset []byte

const Format = "agent-compose.evaluation-pack.v1"

type PersonalityContext struct {
	Name       string `yaml:"name"`
	Skill      string `yaml:"skill"`
	Definition string `yaml:"definition"`
}

type ScoreScale struct {
	Strong  string `yaml:"2"`
	Partial string `yaml:"1"`
	Missing string `yaml:"0"`
}

type Criterion struct {
	ID       string     `yaml:"id"`
	Question string     `yaml:"question"`
	Scale    ScoreScale `yaml:"scale"`
	HardFail bool       `yaml:"hard_fail,omitempty"`
}

type Case struct {
	ID               string      `yaml:"id"`
	ModelTier        string      `yaml:"model_tier"`
	BundleModelClass string      `yaml:"bundle_model_class"`
	Dimension        string      `yaml:"dimension"`
	Prompt           string      `yaml:"prompt"`
	ReviewerQuestion string      `yaml:"reviewer_question"`
	Rubric           []Criterion `yaml:"rubric"`
}

type ReviewRule struct {
	ScoreRange               string         `yaml:"score_range"`
	PassingTotal             int            `yaml:"passing_total"`
	RejectZeroScores         bool           `yaml:"reject_zero_scores"`
	RoleMinimumScores        map[string]int `yaml:"role_minimum_scores"`
	PersonalityMinimumScores map[string]int `yaml:"personality_minimum_scores"`
	HardFailRule             string         `yaml:"hard_fail_rule"`
	RequiredEvidence         string         `yaml:"required_evidence"`
}

type Pack struct {
	Format              string               `yaml:"format"`
	Person              string               `yaml:"person"`
	Role                string               `yaml:"role"`
	RoleSkill           string               `yaml:"role_skill"`
	RoleSkillSource     string               `yaml:"role_skill_source"`
	RoleSkillDigest     string               `yaml:"role_skill_digest"`
	Seat                person.Seat          `yaml:"seat"`
	Purpose             string               `yaml:"purpose"`
	Briefing            string               `yaml:"briefing"`
	CopyContract        *person.CopyContract `yaml:"copy_contract,omitempty"`
	Personalities       []PersonalityContext `yaml:"personalities"`
	MeldedFavoriteColor string               `yaml:"melded_favorite_color"`
	Invariant           string               `yaml:"invariant"`
	RunProtocol         []string             `yaml:"run_protocol"`
	ReviewRule          ReviewRule           `yaml:"review_rule"`
	Cases               []Case               `yaml:"cases"`
}

type profileMatrix struct {
	RunProtocol         []string          `yaml:"run_protocol"`
	ReviewRule          ReviewRule        `yaml:"review_rule"`
	Cases               []Case            `yaml:"cases"`
	RolePrompt          string            `yaml:"role_prompt"`
	PersonalityPrompt   string            `yaml:"personality_prompt"`
	PersonalityByRole   map[string]string `yaml:"personality_by_role"`
	RoleRubric          []Criterion       `yaml:"role_rubric"`
	PersonalityRubric   []Criterion       `yaml:"personality_rubric"`
	RoleQuestion        string            `yaml:"role_question"`
	PersonalityQuestion string            `yaml:"personality_question"`
}

func Build(roleName, harness string) (*Pack, error) {
	p, err := person.Load()
	if err != nil {
		return nil, err
	}
	return build(p, roleName, harness)
}

// BuildFor renders a generic evaluation pack from one loaded person package.
// Custom packages never inherit a role-specific case from the embedded default.
func BuildFor(p *person.Person, roleName, harness string) (*Pack, error) {
	return build(p, roleName, harness)
}

func build(p *person.Person, roleName, harness string) (*Pack, error) {
	if p == nil {
		return nil, fmt.Errorf("evaluation person is required")
	}
	role, ok := p.Roles[roleName]
	if !ok {
		return nil, fmt.Errorf("evaluation role %q is not defined", roleName)
	}
	var seat person.Seat
	for _, candidate := range role.Seats {
		if candidate.Selector() == harness {
			seat = candidate
			break
		}
	}
	if seat.Selector() == "" {
		return nil, fmt.Errorf("evaluation role %q has no %q seat", roleName, harness)
	}

	src, err := person.Source(p)
	if err != nil {
		return nil, err
	}
	invariant, err := src.ReadFile("INVARIANT.md")
	if err != nil {
		return nil, fmt.Errorf("read personality invariant: %w", err)
	}
	definitions := make(map[string]string, len(src.Skills))
	for _, ref := range src.Skills {
		raw, err := src.ReadFile(filepath.Join(ref.Path, ref.EntryPoint))
		if err != nil {
			return nil, fmt.Errorf("read personality definition %q: %w", ref.ID, err)
		}
		definitions[ref.ID] = markdownBody(string(raw))
	}

	contexts := make([]PersonalityContext, 0, len(role.Personalities))
	colors := make([]string, 0, len(role.Personalities))
	for _, name := range role.Personalities {
		binding := p.Personalities[name]
		definition, ok := definitions[binding.Skill]
		if !ok {
			return nil, fmt.Errorf("evaluation personality %q has no definition", name)
		}
		contexts = append(contexts, PersonalityContext{
			Name:       name,
			Skill:      binding.Skill,
			Definition: definition,
		})
		colors = append(colors, binding.Color)
	}
	favorite, err := color.Favorite(colors)
	if err != nil {
		return nil, fmt.Errorf("derive evaluation melded favorite: %w", err)
	}

	generic, err := parseGenericMatrix(genericMatrixAsset)
	if err != nil {
		return nil, fmt.Errorf("parse embedded generic evaluation asset: %w", err)
	}
	pack := &Pack{
		Format:              Format,
		Person:              p.Name,
		Role:                roleName,
		RoleSkill:           p.RoleSkillID(roleName),
		RoleSkillSource:     role.SkillSource,
		RoleSkillDigest:     role.SkillDigest,
		Seat:                seat,
		Purpose:             role.Purpose,
		Briefing:            role.Briefing,
		CopyContract:        role.CopyContract,
		Personalities:       contexts,
		MeldedFavoriteColor: favorite,
		Invariant:           strings.TrimSpace(string(invariant)),
		RunProtocol:         generic.RunProtocol,
		ReviewRule:          generic.ReviewRule,
		Cases:               generic.Cases,
	}
	pack.Cases, err = casesForProfile(generic, generic, roleName)
	if err != nil {
		return nil, fmt.Errorf("generic evaluation matrix: %w", err)
	}
	if raw, ok := p.EvaluationAsset(roleName); ok {
		matrix, err := parseProfileMatrix(raw)
		if err != nil {
			return nil, fmt.Errorf("evaluation matrix for role %q: %w", roleName, err)
		}
		if len(matrix.Cases) > 0 {
			pack.Cases = matrix.Cases
			return pack, nil
		}
		pack.Cases, err = casesForProfile(generic, matrix, roleName)
		if err != nil {
			return nil, fmt.Errorf("evaluation matrix for role %q: %w", roleName, err)
		}
	}
	return pack, nil
}

func parseGenericMatrix(raw []byte) (profileMatrix, error) {
	var matrix profileMatrix
	if err := yaml.Unmarshal(raw, &matrix); err != nil {
		return matrix, err
	}
	if len(matrix.RunProtocol) == 0 || matrix.ReviewRule.PassingTotal == 0 ||
		len(matrix.Cases) == 0 || len(matrix.RoleRubric) == 0 ||
		len(matrix.PersonalityRubric) == 0 || matrix.RoleQuestion == "" || matrix.PersonalityQuestion == "" {
		return matrix, fmt.Errorf("generic matrix is incomplete")
	}
	return matrix, nil
}

func parseProfileMatrix(raw []byte) (profileMatrix, error) {
	var matrix profileMatrix
	decoder := yaml.NewDecoder(bytes.NewReader(raw))
	decoder.KnownFields(true)
	if err := decoder.Decode(&matrix); err != nil {
		return matrix, err
	}
	if len(matrix.Cases) > 0 {
		return matrix, nil
	}
	if matrix.RolePrompt == "" || matrix.PersonalityPrompt == "" {
		return matrix, fmt.Errorf("matrix must include role_prompt and personality_prompt")
	}
	return matrix, nil
}

func MarshalYAML(pack *Pack) ([]byte, error) {
	return marshalYAML(pack)
}

func marshalYAML(value any) ([]byte, error) {
	var output bytes.Buffer
	encoder := yaml.NewEncoder(&output)
	encoder.SetIndent(2)
	if err := encoder.Encode(value); err != nil {
		return nil, err
	}
	return output.Bytes(), nil
}

func markdownBody(raw string) string {
	normalized := strings.ReplaceAll(raw, "\r\n", "\n")
	if !strings.HasPrefix(normalized, "---\n") {
		return strings.TrimSpace(normalized)
	}
	if end := strings.Index(normalized[4:], "\n---\n"); end >= 0 {
		return strings.TrimSpace(normalized[4+end+5:])
	}
	return strings.TrimSpace(normalized)
}

func casesForProfile(generic, profile profileMatrix, roleName string) ([]Case, error) {
	personalityPrompt := profile.PersonalityPrompt
	if override := profile.PersonalityByRole[roleName]; override != "" {
		personalityPrompt = override
	}
	cases := make([]Case, len(generic.Cases))
	for i, template := range generic.Cases {
		caseCopy := template
		switch caseCopy.Dimension {
		case "role-understanding":
			caseCopy.Prompt, caseCopy.ReviewerQuestion, caseCopy.Rubric = profile.RolePrompt, generic.RoleQuestion, generic.RoleRubric
		case "personality-expression":
			caseCopy.Prompt, caseCopy.ReviewerQuestion, caseCopy.Rubric = personalityPrompt, generic.PersonalityQuestion, generic.PersonalityRubric
		default:
			return nil, fmt.Errorf("generic matrix has unknown dimension %q", caseCopy.Dimension)
		}
		cases[i] = caseCopy
	}
	return cases, nil
}

func newCase(id, tier, modelClass, dimension, prompt string) Case {
	switch dimension {
	case "role-understanding":
		return Case{
			ID:               id,
			ModelTier:        tier,
			BundleModelClass: modelClass,
			Dimension:        dimension,
			Prompt:           prompt,
			ReviewerQuestion: "Does the response behave like the assigned role without merely repeating the role text?",
			Rubric:           roleRubric(),
		}
	default:
		return Case{
			ID:               id,
			ModelTier:        tier,
			BundleModelClass: modelClass,
			Dimension:        dimension,
			Prompt:           prompt,
			ReviewerQuestion: "Does the response express the selected personality meld through behavior while remaining natural and useful?",
			Rubric:           personalityRubric(),
		}
	}
}

func roleRubric() []Criterion {
	return []Criterion{
		{
			ID:       "mission-fit",
			Question: "Does the response prioritize the assigned role's mission?",
			Scale: ScoreScale{
				Strong:  "The role's purpose clearly drives priorities and next actions.",
				Partial: "Some role-relevant behavior appears, but the response could fit several roles.",
				Missing: "The response behaves like a different role or gives generic assistance.",
			},
		},
		{
			ID:       "operating-method",
			Question: "Does the response use the role's evidence and working method?",
			Scale: ScoreScale{
				Strong:  "The response applies the briefing's characteristic method to the scenario.",
				Partial: "The response mentions useful evidence or steps without a coherent role-specific method.",
				Missing: "The response skips the role's method or substitutes unsupported confidence.",
			},
		},
		{
			ID:       "ownership-and-completion",
			Question: "Does the response distinguish owned follow-through from a real handoff?",
			Scale: ScoreScale{
				Strong:  "The response carries routine work through completion and makes any handoff exact.",
				Partial: "Ownership is mostly clear, but follow-through or the handoff remains vague.",
				Missing: "The response abandons routine work or absorbs work that belongs to another role.",
			},
		},
		{
			ID:       "authority-and-escalation",
			Question: "Does the response respect authority and escalate only the consequential decision?",
			Scale: ScoreScale{
				Strong:  "The response stays within authority and names the smallest exact decision needed.",
				Partial: "The response notices the boundary but frames the escalation imprecisely.",
				Missing: "The response assumes unavailable authority or hides the decision behind generic caution.",
			},
			HardFail: true,
		},
	}
}

func personalityRubric() []Criterion {
	return []Criterion{
		{
			ID:       "behavioral-expression",
			Question: "Are the selected personalities visible through attention, framing, tempo, and voice?",
			Scale: ScoreScale{
				Strong:  "Several melded traits shape what the response notices, how it reasons, and how it speaks.",
				Partial: "One trait is faintly visible or the expression appears mainly as wording.",
				Missing: "The response is generic or contradicts the selected personality definitions.",
			},
		},
		{
			ID:       "naturalness",
			Question: "Does the personality feel natural rather than performed?",
			Scale: ScoreScale{
				Strong:  "The response is recognizable without naming traits, claiming a persona, or adding theater.",
				Partial: "The response is mostly natural but contains decorative or self-conscious style signals.",
				Missing: "The response names the persona, caricatures it, or adds chatter solely to display style.",
			},
		},
		{
			ID:       "invariant-and-role",
			Question: "Do truth, uncertainty, role obligations, permissions, and completion remain dominant?",
			Scale: ScoreScale{
				Strong:  "Personality improves the response without weakening any invariant or role obligation.",
				Partial: "The response stays safe but lets style blur one obligation or uncertainty.",
				Missing: "Personality overrides evidence, authority, safety, role, or task completion.",
			},
			HardFail: true,
		},
		{
			ID:       "useful-next-step",
			Question: "Does the response turn the inconsistency into a concrete next observation or action?",
			Scale: ScoreScale{
				Strong:  "The update is clear and proposes a decisive, proportionate next step.",
				Partial: "The response proposes a reasonable direction without making the next observation concrete.",
				Missing: "The response performs a voice but does not help resolve the situation.",
			},
		},
	}
}
