package roster

import (
	"fmt"
	"strings"
	"testing"

	"forgejo.coilysiren.me/coilyco-flight-deck/agent-compose/internal/person"
	"forgejo.coilysiren.me/coilyco-flight-deck/agent-compose/internal/schema"
)

func TestEmbeddedRosterRendersEveryCanonicalSeat(t *testing.T) {
	p, err := person.Load()
	if err != nil {
		t.Fatal(err)
	}
	source, err := person.Source(p)
	if err != nil {
		t.Fatal(err)
	}
	files, err := Render(p, []*schema.Source{source}, "/opt/artifact")
	if err != nil {
		t.Fatal(err)
	}
	table := string(files["AGENTS.COMPOSE.md"])
	for _, roleName := range p.RoleOrder {
		role := p.Roles[roleName]
		if !strings.Contains(table, "## "+roleName+" - "+role.Purpose) {
			t.Errorf("roster omitted role %q", roleName)
		}
		for _, seat := range role.Seats {
			want := fmt.Sprintf("If you are %s running the %s role: your name is %s",
				seat.Harness, roleName, seat.Name)
			if !strings.Contains(table, want) {
				t.Errorf("roster omitted seat %q for role %q", seat.Harness, roleName)
			}
		}
	}
}
