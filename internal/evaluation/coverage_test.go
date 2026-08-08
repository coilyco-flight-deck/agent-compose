package evaluation

import (
	"strings"
	"testing"
)

func TestValidateCorePackRejectsCoverageAndPairDrift(t *testing.T) {
	pack, err := Build("engineer", "codex")
	if err != nil {
		t.Fatal(err)
	}
	for name, mutate := range map[string]func(*Pack){
		"missing tier": func(candidate *Pack) {
			var kept []Case
			for _, evalCase := range candidate.Cases {
				if evalCase.ModelTier == ossTier &&
					evalCase.ScenarioKind == ScenarioMissionFit {
					continue
				}
				kept = append(kept, evalCase)
			}
			candidate.Cases = kept
		},
		"tier prompt drift": func(candidate *Pack) {
			for index := range candidate.Cases {
				if candidate.Cases[index].ModelTier == ossTier {
					candidate.Cases[index].Prompt += " drift"
					return
				}
			}
		},
		"missing adjacent role": func(candidate *Pack) {
			var kept []Case
			for _, evalCase := range candidate.Cases {
				if evalCase.ScenarioKind != ScenarioAdjacentRole {
					kept = append(kept, evalCase)
				}
			}
			candidate.Cases = kept
		},
		"missing human communication boundary": func(candidate *Pack) {
			var kept []Case
			for _, evalCase := range candidate.Cases {
				if evalCase.ScenarioKind != ScenarioHumanCommunication {
					kept = append(kept, evalCase)
				}
			}
			candidate.Cases = kept
		},
		"missing evidence acquisition": func(candidate *Pack) {
			var kept []Case
			for _, evalCase := range candidate.Cases {
				if evalCase.ScenarioKind != ScenarioEvidenceAcquisition {
					kept = append(kept, evalCase)
				}
			}
			candidate.Cases = kept
		},
		"evidence acquisition criterion dropped": func(candidate *Pack) {
			for caseIndex := range candidate.Cases {
				if candidate.Cases[caseIndex].ScenarioKind != ScenarioEvidenceAcquisition {
					continue
				}
				var kept []Criterion
				for _, criterion := range candidate.Cases[caseIndex].Rubric {
					if criterion.ID != "evidence-acquisition" {
						kept = append(kept, criterion)
					}
				}
				candidate.Cases[caseIndex].Rubric = kept
				return
			}
		},
		"human communication boundary is not a hard fail": func(candidate *Pack) {
			for caseIndex := range candidate.Cases {
				if candidate.Cases[caseIndex].ScenarioKind != ScenarioHumanCommunication {
					continue
				}
				for criterionIndex := range candidate.Cases[caseIndex].Rubric {
					criterion := &candidate.Cases[caseIndex].Rubric[criterionIndex]
					if criterion.ID == "human-communication-ownership" {
						criterion.HardFail = false
						return
					}
				}
			}
		},
	} {
		t.Run(name, func(t *testing.T) {
			candidate := *pack
			candidate.Cases = append([]Case(nil), pack.Cases...)
			for index := range candidate.Cases {
				candidate.Cases[index].Rubric = append(
					[]Criterion(nil),
					pack.Cases[index].Rubric...,
				)
			}
			mutate(&candidate)
			if err := ValidateCorePack(&candidate); err == nil {
				t.Fatal("invalid Core Roster evaluation coverage passed")
			}
		})
	}
}

func TestParseProfileMatrixRejectsInvalidScenarioAssets(t *testing.T) {
	for name, raw := range map[string]string{
		"unknown kind": `
scenarios:
  - id: fixture
    kind: mystery
    prompt: Speak in first person.
`,
		"adjacent without target": `
scenarios:
  - id: fixture
    kind: adjacent-role-discrimination
    prompt: Speak in first person.
`,
		"duplicate id": `
scenarios:
  - id: fixture
    kind: mission-fit
    prompt: Speak in first person.
  - id: fixture
    kind: personality-expression
    prompt: Speak in first person.
`,
		"mixed ownership": `
run_protocol: [Run.]
scenarios:
  - id: fixture
    kind: mission-fit
    prompt: Speak in first person.
`,
	} {
		t.Run(name, func(t *testing.T) {
			if _, err := parseProfileMatrix([]byte(raw)); err == nil ||
				strings.TrimSpace(err.Error()) == "" {
				t.Fatalf("invalid scenario matrix error = %v", err)
			}
		})
	}
}
