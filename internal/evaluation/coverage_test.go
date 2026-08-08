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
		"evidence acquisition without the meld": func(candidate *Pack) {
			var kept []MeldContext
			for _, meld := range candidate.Melds {
				if meld.Name != evidenceMeld {
					kept = append(kept, meld)
				}
			}
			candidate.Melds = kept
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

func TestValidateCorePackRejectsCommunicationDriftForBoundRoles(t *testing.T) {
	// The engineer pack no longer declares comms, so communication coverage is
	// exercised against a declaring role and against the counterpart.
	declaring, err := Build("ops", "codex")
	if err != nil {
		t.Fatal(err)
	}
	counterpart, err := Build("creator", "codex")
	if err != nil {
		t.Fatal(err)
	}
	for name, test := range map[string]struct {
		base   *Pack
		mutate func(*Pack)
	}{
		"declaring role missing the boundary": {base: declaring, mutate: func(candidate *Pack) {
			var kept []Case
			for _, evalCase := range candidate.Cases {
				if evalCase.ScenarioKind != ScenarioHumanCommunication {
					kept = append(kept, evalCase)
				}
			}
			candidate.Cases = kept
		}},
		"counterpart missing the boundary": {base: counterpart, mutate: func(candidate *Pack) {
			var kept []Case
			for _, evalCase := range candidate.Cases {
				if evalCase.ScenarioKind != ScenarioHumanCommunication {
					kept = append(kept, evalCase)
				}
			}
			candidate.Cases = kept
		}},
		"counterpart relationship dropped": {base: counterpart, mutate: func(candidate *Pack) {
			candidate.CounterpartMelds = nil
		}},
		"boundary without the meld": {base: declaring, mutate: func(candidate *Pack) {
			var kept []MeldContext
			for _, meld := range candidate.Melds {
				if meld.Name != commsMeld {
					kept = append(kept, meld)
				}
			}
			candidate.Melds = kept
		}},
		"boundary is not a hard fail": {base: declaring, mutate: func(candidate *Pack) {
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
		}},
	} {
		t.Run(name, func(t *testing.T) {
			candidate := *test.base
			candidate.Cases = append([]Case(nil), test.base.Cases...)
			candidate.Melds = append([]MeldContext(nil), test.base.Melds...)
			candidate.CounterpartMelds = append([]string(nil), test.base.CounterpartMelds...)
			for index := range candidate.Cases {
				candidate.Cases[index].Rubric = append(
					[]Criterion(nil),
					test.base.Cases[index].Rubric...,
				)
			}
			test.mutate(&candidate)
			if err := ValidateCorePack(&candidate); err == nil {
				t.Fatal("invalid communication coverage passed")
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
