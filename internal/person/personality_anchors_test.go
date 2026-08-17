package person

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

// anchorsPath is repo-root relative; this test runs in the package directory.
const anchorsPath = "../../evaluations/personality-anchors.yaml"

type personalityAnchors struct {
	Version int `yaml:"version"`
	Scale   struct {
		Fit        string `yaml:"fit"`
		Undecided  string `yaml:"undecided"`
		DoesNotFit string `yaml:"does_not_fit"`
	} `yaml:"scale"`
	UniversalDeductions map[string]string `yaml:"universal_deductions"`
	Traits              map[string]struct {
		Fit         string `yaml:"fit"`
		Deduct      string `yaml:"deduct"`
		Distinguish string `yaml:"distinguish"`
	} `yaml:"traits"`
}

func loadAnchors(t *testing.T) personalityAnchors {
	t.Helper()
	raw, err := os.ReadFile(filepath.Clean(anchorsPath))
	if err != nil {
		t.Fatalf("read personality anchors: %v", err)
	}
	var anchors personalityAnchors
	if err := yaml.Unmarshal(raw, &anchors); err != nil {
		t.Fatalf("parse personality anchors: %v", err)
	}
	return anchors
}

// The personality tier bypasses item analysis, so these anchors are the only
// thing holding its scale still. A personality without one cannot be graded.
func TestEveryRosterPersonalityHasAGradingAnchor(t *testing.T) {
	p, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	anchors := loadAnchors(t)
	used := map[string]bool{}
	for _, roleName := range p.RoleOrder {
		for _, name := range p.Roles[roleName].Personalities {
			used[name] = true
			if _, ok := anchors.Traits[name]; !ok {
				t.Errorf("role %q melds %q, which has no grading anchor", roleName, name)
			}
		}
	}
	for name := range anchors.Traits {
		if !used[name] {
			t.Errorf("anchor %q matches no personality in any role meld", name)
		}
	}
}

func TestAnchorsAreObservableBehaviours(t *testing.T) {
	anchors := loadAnchors(t)
	if anchors.Version == 0 {
		t.Fatal("anchors need a version the graded records can reference")
	}
	if anchors.Scale.Fit == "" || anchors.Scale.Undecided == "" || anchors.Scale.DoesNotFit == "" {
		t.Fatal("all three scale points need an anchor")
	}
	if len(anchors.UniversalDeductions) == 0 {
		t.Fatal("the universal deduction patterns are missing")
	}
	for name, trait := range anchors.Traits {
		if trait.Fit == "" || trait.Deduct == "" || trait.Distinguish == "" {
			t.Errorf("anchor %q is incomplete", name)
			continue
		}
		// A fit anchor naming only the trait back at the grader is the defect
		// these anchors exist to avoid, so reject the tautology.
		if strings.HasPrefix(strings.ToLower(trait.Fit), "is "+name) ||
			strings.EqualFold(strings.TrimSpace(trait.Fit), name) {
			t.Errorf("anchor %q restates the trait instead of naming a behaviour", name)
		}
		if !strings.Contains(trait.Distinguish, ".") {
			t.Errorf("anchor %q must name the neighbour it is not", name)
		}
	}
}
