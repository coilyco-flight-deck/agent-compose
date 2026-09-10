package person

import (
	"strings"
	"testing"
)

func withCarried(c *Carried) *Person {
	return &Person{
		Name:      "test",
		RoleOrder: []string{"vera"},
		Roles:     map[string]Role{"vera": {Carried: c}},
	}
}

// The binary lands ahead of its data, which is the safe direction: a roster
// ahead of its binary is the failure agent-compose#7212 names.
func TestCarriedWarningsSilentOnShippedRoster(t *testing.T) {
	p, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if got := p.CarriedWarnings(); len(got) != 0 {
		t.Fatalf("shipped roster warns: %v", got)
	}
	for name, role := range p.Roles {
		if role.Carried != nil {
			t.Fatalf("role %q already declares carried; this test's premise is stale", name)
		}
	}
}

func TestCarriedWarningsCatchEachIncompleteShape(t *testing.T) {
	cases := []struct {
		name    string
		carried *Carried
		want    string
	}{
		{"no clause", &Carried{Clause: "  "}, "has no clause"},
		{
			"nothing protects it",
			&Carried{Clause: "a stack of cargo containers"},
			"no material declared",
		},
		{
			"noun absent from its own clause",
			&Carried{
				Clause:    "a stack of cargo containers roped to its shell",
				Materials: []Material{{Noun: "helm wheel", Material: "iron"}},
			},
			"does not appear in its own carried clause",
		},
		{
			"half a pair",
			&Carried{
				Clause:    "a stack of cargo containers",
				Materials: []Material{{Noun: "cargo containers", Material: ""}},
			},
			"incomplete",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := withCarried(tc.carried).CarriedWarnings()
			if len(got) == 0 {
				t.Fatalf("no warning for %q", tc.name)
			}
			if !strings.Contains(got[0], tc.want) {
				t.Fatalf("warning %q does not mention %q", got[0], tc.want)
			}
			if !strings.Contains(got[0], "vera") {
				t.Fatalf("warning %q does not name the role", got[0])
			}
		})
	}
}

// The positive control the four cases above need: a complete declaration is
// silent, so the checks are discriminating rather than always firing.
func TestCarriedWarningsSilentOnACompleteDeclaration(t *testing.T) {
	complete := &Carried{
		Clause: "a seven-spoked iron ship's helm wheel mounted upright on the side of its " +
			"shell, a neat stack of square timber cargo containers roped down on top",
		Stance: "planted square on both feet",
		Only:   false,
		Materials: []Material{
			{Noun: "helm wheel", Material: "iron"},
			{Noun: "cargo containers", Material: "timber"},
		},
	}
	if got := withCarried(complete).CarriedWarnings(); len(got) != 0 {
		t.Fatalf("complete declaration warns: %v", got)
	}
}

// Absence is ordinary: a role that carries nothing draws empty-handed.
func TestCarriedAbsenceIsNotAWarning(t *testing.T) {
	if got := withCarried(nil).CarriedWarnings(); len(got) != 0 {
		t.Fatalf("a role with no carried object warns: %v", got)
	}
}
