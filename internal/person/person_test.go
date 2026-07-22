package person

import "testing"

func TestLoadEmbeddedRoster(t *testing.T) {
	p, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if len(p.Roles) != 8 {
		t.Fatalf("expected 8 roles, got %d", len(p.Roles))
	}
	var seats int
	for _, role := range p.Roles {
		seats += len(role.Seats)
	}
	if seats != 12 {
		t.Fatalf("expected 12 named seats, got %d", seats)
	}

	engineer := p.Roles["engineer"]
	if len(engineer.Personalities) != 3 || len(engineer.Seats) != 2 {
		t.Fatalf("unexpected engineer role: %+v", engineer)
	}
	for _, seat := range engineer.Seats {
		switch seat.Harness {
		case "claude":
			if seat.Name != "opal engineer" || seat.Pronouns != "she" {
				t.Fatalf("unexpected claude engineer seat: %+v", seat)
			}
		case "codex":
			if seat.Name != "terran engineer" || seat.Pronouns != "he" {
				t.Fatalf("unexpected codex engineer seat: %+v", seat)
			}
		default:
			t.Fatalf("unexpected harness %q", seat.Harness)
		}
	}

	for _, stub := range []string{"social", "sales"} {
		role, ok := p.Roles[stub]
		if !ok || role.Purpose == "" || len(role.Seats) != 0 {
			t.Fatalf("stub role %q must exist with purpose and no seats: %+v", stub, role)
		}
	}
	if len(p.Personalities) != 3 {
		t.Fatalf("expected 3 bound personalities, got %d", len(p.Personalities))
	}
	for name, binding := range p.Personalities {
		if binding.Color == "" {
			t.Fatalf("personality %q must carry a favorite color", name)
		}
	}
}

func TestParseRejectsIllegibleColor(t *testing.T) {
	body := `person "kai" {
    role "engineer" {
        purpose "Build."
        personality "curious"
    }
    personality "curious" skill="personality-curious" color="#111111"
}`
	if _, err := parse([]byte(body)); err == nil {
		t.Fatal("an illegible color must fail the parse-time gate")
	}
}

func TestParseSeatValidation(t *testing.T) {
	cases := map[string]string{
		"duplicate seat": `person "kai" {
    role "engineer" {
        purpose "Build."
        agent "claude" name="opal engineer"
        agent "claude" name="another"
    }
}`,
		"seat without name": `person "kai" {
    role "engineer" {
        purpose "Build."
        agent "claude"
    }
}`,
		"unknown role child": `person "kai" {
    role "engineer" {
        guardfile "guardfile.aws.kdl"
    }
}`,
	}
	for name, body := range cases {
		t.Run(name, func(t *testing.T) {
			if _, err := parse([]byte(body)); err == nil {
				t.Fatal("expected parse failure")
			}
		})
	}
}
