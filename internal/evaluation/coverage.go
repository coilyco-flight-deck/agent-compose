package evaluation

import (
	"fmt"
	"sort"

	"forgejo.coilysiren.me/coilyco-flight-deck/agent-compose/internal/person"
)

var requiredScenarioKinds = []string{
	ScenarioMissionFit,
	ScenarioPersonality,
	ScenarioAuthorityBoundary,
	ScenarioCompletionOwnership,
	ScenarioPortfolioReplay,
}

// Each boundary ties a scenario kind to the doctrine it evaluates, so the roster
// decides which roles owe the case. See docs/evaluation-matrices.md.
const (
	externalBoundary = "seek-external-validation"
	commsBoundary    = "suggest-human-comms"
)

var requiredAdjacentRoles = map[string][]string{
	"engineer": {"ops"},
	"director": {"exec"},
	"qa":       {},
	"ops":      {"engineer"},
	"design":   {"creator"},
	"exec":     {"director"},
	"creator":  {"design", "exec", "director", "ai"},
	"ai":       {"engineer", "qa", "ops", "creator"},
}

// BuildCorePacks renders and validates the complete embedded Core Roster
// evaluation matrix for one harness seat.
func BuildCorePacks(harness string) ([]*Pack, error) {
	p, err := person.Load()
	if err != nil {
		return nil, err
	}
	packs := make([]*Pack, 0, len(p.RoleOrder))
	for _, role := range p.RoleOrder {
		pack, err := build(p, role, harness)
		if err != nil {
			return nil, err
		}
		if err := ValidateCorePack(pack); err != nil {
			return nil, fmt.Errorf("evaluation role %q: %w", role, err)
		}
		packs = append(packs, pack)
	}
	if len(packs) != len(p.RoleOrder) {
		return nil, fmt.Errorf(
			"Core Roster evaluation has %d roles, loader has %d",
			len(packs),
			len(p.RoleOrder),
		)
	}
	return packs, nil
}

// ValidateCorePack enforces the v2 scenario and adjacent-role coverage gates.
func ValidateCorePack(pack *Pack) error {
	if pack == nil {
		return fmt.Errorf("evaluation pack is required")
	}
	requiredAdjacent, ok := requiredAdjacentRoles[pack.Role]
	if !ok {
		return fmt.Errorf("role is not in the Core Roster")
	}
	type observedScenario struct {
		kind     string
		adjacent string
		prompt   string
	}
	scenarios := make(map[string]*observedScenario)
	for _, evalCase := range pack.Cases {
		if evalCase.Scenario == "" || evalCase.ScenarioKind == "" {
			return fmt.Errorf("case %q has no v2 scenario identity", evalCase.ID)
		}
		if evalCase.ModelTier != SubjectModelTier {
			return fmt.Errorf(
				"case %q runs tier %q, want the %q subject",
				evalCase.ID,
				evalCase.ModelTier,
				SubjectModelTier,
			)
		}
		if evalCase.ScenarioKind == ScenarioHumanCommunication {
			if err := validateHumanCommunicationHardFail(evalCase); err != nil {
				return fmt.Errorf("case %q: %w", evalCase.ID, err)
			}
		}
		if evalCase.ScenarioKind == ScenarioExternalValidation {
			if err := validateExternalValidationCriterion(evalCase); err != nil {
				return fmt.Errorf("case %q: %w", evalCase.ID, err)
			}
		}
		if scenarios[evalCase.Scenario] != nil {
			return fmt.Errorf("scenario %q is repeated", evalCase.Scenario)
		}
		scenarios[evalCase.Scenario] = &observedScenario{
			kind:     evalCase.ScenarioKind,
			adjacent: evalCase.AdjacentRole,
			prompt:   evalCase.Prompt,
		}
	}

	kinds := make(map[string]int)
	adjacent := make(map[string]int)
	for _, scenario := range scenarios {
		kinds[scenario.kind]++
		if scenario.kind == ScenarioAdjacentRole {
			adjacent[scenario.adjacent]++
		}
	}
	for _, kind := range requiredScenarioKinds {
		if kinds[kind] == 0 {
			return fmt.Errorf("missing %s scenario", kind)
		}
	}
	// Both sides of a boundary receive its body, so both owe its case. The
	// roster decides coverage. See docs/evaluation-matrices.md.
	for boundary, kind := range map[string]string{
		commsBoundary:    ScenarioHumanCommunication,
		externalBoundary: ScenarioExternalValidation,
	} {
		carries := false
		for _, active := range pack.Boundaries {
			if active.Name == boundary {
				carries = true
				break
			}
		}
		if carries && kinds[kind] == 0 {
			return fmt.Errorf("missing %s scenario for the %q boundary", kind, boundary)
		}
		if !carries && kinds[kind] > 0 {
			return fmt.Errorf("%s scenario without the %q boundary", kind, boundary)
		}
	}
	for _, role := range requiredAdjacent {
		if adjacent[role] == 0 {
			return fmt.Errorf("missing adjacent-role scenario for %q", role)
		}
	}
	if len(adjacent) != len(requiredAdjacent) {
		got := make([]string, 0, len(adjacent))
		for role := range adjacent {
			got = append(got, role)
		}
		sort.Strings(got)
		return fmt.Errorf("adjacent roles %v do not match required %v", got, requiredAdjacent)
	}
	return nil
}

// validateExternalValidationCriterion keeps the acquisition scenario scored on
// acquisition. See docs/evaluation-matrices.md for why it is not a hard fail.
func validateExternalValidationCriterion(evalCase Case) error {
	for _, criterion := range evalCase.Rubric {
		if criterion.ID == "external-validation-deferral" {
			if criterion.HardFail {
				return fmt.Errorf("external validation criterion is flagged as a hard fail")
			}
			return nil
		}
	}
	return fmt.Errorf("external validation criterion is missing")
}

func validateHumanCommunicationHardFail(evalCase Case) error {
	for _, criterion := range evalCase.Rubric {
		if criterion.ID == "human-communication-ownership" {
			if !criterion.HardFail {
				return fmt.Errorf("human communication ownership criterion is not a hard fail")
			}
			return nil
		}
	}
	return fmt.Errorf("human communication ownership criterion is missing")
}
