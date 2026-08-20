package person

import "testing"

// The scoped axis is additive, so the shipped roster must keep loading with no
// role declaring it. Everything below drives the axis through fixtures.
func TestShippedRosterLoadsWithoutScopedBoundaries(t *testing.T) {
	p, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	for _, roleName := range p.RoleOrder {
		if got := len(p.Roles[roleName].ScopedBoundaries); got != 0 {
			t.Errorf("role %q declares %d scoped boundaries, want 0 before the axis is used", roleName, got)
		}
	}
}

func TestScopedBoundaryJoinsTheActiveSet(t *testing.T) {
	p := &Person{
		RoleOrder:  []string{"builder"},
		Roles:      map[string]Role{"builder": {ScopedBoundaries: []ScopedBoundary{{Name: "modify-live-backend", Scope: "local worlds only"}}}},
		Boundaries: map[string]Boundary{"modify-live-backend": {Owner: "operator", Skill: "boundary-modify-live-backend"}},
	}
	active := p.RoleActiveBoundaries("builder")
	if len(active) != 1 || active[0] != "modify-live-backend" {
		t.Fatalf("active = %v, want the scoped boundary to be delivered", active)
	}
}

func TestScopedBoundaryCountsAsUse(t *testing.T) {
	p := &Person{
		RoleOrder:  []string{"builder"},
		Roles:      map[string]Role{"builder": {ScopedBoundaries: []ScopedBoundary{{Name: "modify-live-backend", Scope: "local worlds only"}}}},
		Boundaries: map[string]Boundary{"modify-live-backend": {Owner: "operator"}},
	}
	if err := validateNoUnusedBoundaries(p); err != nil {
		t.Fatalf("a scoped declaration should count as a use: %v", err)
	}
}

// An owner already receives the body by owning it. Scoping its own boundary
// would hand it two contradictory sides.
func TestOwnerMayNotScopeItsOwnBoundary(t *testing.T) {
	p := &Person{
		RoleOrder:  []string{"operator"},
		Roles:      map[string]Role{"operator": {ScopedBoundaries: []ScopedBoundary{{Name: "modify-live-backend", Scope: "anything"}}}},
		Boundaries: map[string]Boundary{"modify-live-backend": {Owner: "operator"}},
	}
	if err := validateBoundaryOwners(p); err == nil {
		t.Fatal("owner scoping its own boundary should fail")
	}
}

func TestScopedBoundaryBodyOrdering(t *testing.T) {
	for name, body := range map[string]string{
		"scoped before own":  boundaryScopedHeading + "\na\n" + boundaryOwnHeading + "\nb\n" + boundaryDeferHeading + "\nc\n",
		"scoped after defer": boundaryOwnHeading + "\na\n" + boundaryDeferHeading + "\nb\n" + boundaryScopedHeading + "\nc\n",
	} {
		if err := validateBoundaryBodySides("b", body); err == nil {
			t.Errorf("%s: want an ordering failure", name)
		}
	}
	ok := boundaryOwnHeading + "\na\n" + boundaryScopedHeading + "\nb\n" + boundaryDeferHeading + "\nc\n"
	if err := validateBoundaryBodySides("b", ok); err != nil {
		t.Errorf("own..scoped..defer should validate: %v", err)
	}
}

// A two-sided body is still valid. Only a boundary somebody scopes needs three.
func TestTwoSidedBodyStillValidates(t *testing.T) {
	body := boundaryOwnHeading + "\na\n" + boundaryDeferHeading + "\nb\n"
	if err := validateBoundaryBodySides("b", body); err != nil {
		t.Fatalf("a boundary nobody scopes needs no scoped section: %v", err)
	}
}

func TestRoleMayNotBothDeferAndScope(t *testing.T) {
	p, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	for _, roleName := range p.RoleOrder {
		role := p.Roles[roleName]
		for _, scoped := range role.ScopedBoundaries {
			for _, declared := range role.Boundaries {
				if declared == scoped.Name {
					t.Errorf("role %q both defers and scopes %q", roleName, scoped.Name)
				}
			}
		}
	}
}
