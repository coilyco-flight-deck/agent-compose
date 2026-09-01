package person

import "testing"

// These bound the roster this repository ships, not the mount contract: a
// library-merged package legitimately fails four of them. #336.
func TestShippedRosterMeetsWholeRosterGates(t *testing.T) {
	p, err := Load()
	if err != nil {
		t.Fatalf("load shipped roster: %v", err)
	}
	for _, gate := range []struct {
		name string
		fn   func(*Person) error
	}{
		{"unused personalities", validateNoUnusedPersonalities},
		{"unused boundaries", validateNoUnusedBoundaries},
		{"scoped boundary bodies", validateScopedBoundaryBodies},
		{"boundary owners", validateBoundaryOwners},
		{"act coverage", validateActCoverage},
		{"role adjacents", validateRoleAdjacents},
		{"personality melds", validateCorePersonalityMelds},
	} {
		if err := gate.fn(p); err != nil {
			t.Errorf("%s: %v", gate.name, err)
		}
	}
	if err := validateRosterProseFloors(p.source, p); err != nil {
		t.Errorf("prose floors: %v", err)
	}
}
