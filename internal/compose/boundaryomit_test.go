package compose

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"forgejo.coilysiren.me/coilyco-flight-deck/agent-compose/internal/person"
	"forgejo.coilysiren.me/coilyco-flight-deck/agent-compose/internal/resolver"
	"forgejo.coilysiren.me/coilyco-flight-deck/agent-compose/internal/schema"
)

// A defer-side boundary routes work to the seat that owns it. A deployment
// with no such seat composes without it. See docs/ownership.md.

func TestBoundaryOmitLeavesNoTraceInTheBundle(t *testing.T) {
	out := t.TempDir()
	result, err := Run(fixture(t, "boundary-omit.kdl"), out)
	if err != nil {
		t.Fatalf("compose: %v", err)
	}
	if m := readManifest(t, result.Bundle.Dir); len(m.Boundaries) != 0 {
		t.Fatalf("manifest still claims boundaries: %v", m.Boundaries)
	}
	instructions, err := os.ReadFile(filepath.Join(result.Bundle.Dir, "content", "instructions.md"))
	if err != nil {
		t.Fatal(err)
	}
	// Naming a boundary whose body is absent is the failure this replaces, so
	// the card has to lose the name and not merely the prose.
	for _, absent := range []string{
		"boundary-modify-live-system",
		"boundary-seek-external-validation",
		"## Boundaries",
	} {
		if strings.Contains(string(instructions), absent) {
			t.Errorf("identity card still carries %q", absent)
		}
	}
	skills := filepath.Join(result.Bundle.Dir, "content", "skills")
	matches, err := filepath.Glob(filepath.Join(skills, "*", "boundary-*"))
	if err != nil {
		t.Fatal(err)
	}
	if len(matches) != 0 {
		t.Fatalf("bundle still ships boundary skills: %v", matches)
	}
}

// A bundle that quietly lacks a boundary is worse than one that never had it,
// because the review surface stops telling the truth.
func TestBoundaryOmitIsRecordedInTheDecisionTrace(t *testing.T) {
	out := t.TempDir()
	result, err := Run(fixture(t, "boundary-omit.kdl"), out)
	if err != nil {
		t.Fatalf("compose: %v", err)
	}
	excluded := map[string]bool{}
	for _, d := range result.Resolution.Decisions {
		if d.Kind == "profile" && d.Outcome == resolver.OutcomeExcluded {
			excluded[d.Subject] = true
		}
	}
	for _, want := range []string{
		"boundary:modify-live-system",
		"boundary:seek-external-validation",
	} {
		if !excluded[want] {
			t.Errorf("decision trace does not record %s as excluded", want)
		}
	}
}

func TestBoundaryOmitRefusesWhatItCannotMean(t *testing.T) {
	p, err := person.Load()
	if err != nil {
		t.Fatal(err)
	}
	cases := map[string]struct {
		role, omit, want string
	}{
		"unknown boundary": {"engineer", "no-such-boundary", "unknown boundary"},
		// An owner losing its own boundary is a larger claim than a deferrer
		// losing one, and this knob is not allowed to make it.
		"owned by the role": {"ops", "modify-live-system", "is owned by role"},
		"not active for it": {"engineer", "suggest-human-comms", "is not active for role"},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			req := &schema.Request{
				Role:              tc.role,
				Delivery:          "native-skills",
				BoundaryOmissions: []string{tc.omit},
			}
			_, err := resolver.Resolve(req, p, nil, nil)
			if err == nil {
				t.Fatalf("omitting %q for role %q was accepted", tc.omit, tc.role)
			}
			if !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("error %q does not mention %q", err, tc.want)
			}
		})
	}
}
