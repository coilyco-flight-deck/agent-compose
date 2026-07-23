package person

import (
	"reflect"
	"testing"
)

func TestLoadEmbeddedRoster(t *testing.T) {
	p, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	wantPersonalities := map[string][]string{
		"director":         {"bold", "grounded", "diplomatic"},
		"advisor":          {"reflective", "curious", "candid"},
		"pm":               {"warm", "meticulous", "curious"},
		"designer":         {"imaginative", "playful", "warm"},
		"engineer":         {"curious", "grounded", "meticulous"},
		"qa":               {"meticulous", "candid", "playful"},
		"ops":              {"protective", "grounded", "reflective"},
		"sales":            {"charming", "energetic", "warm"},
		"social":           {"quirky", "playful", "optimistic"},
		"customer-success": {"nurturing", "diplomatic", "optimistic"},
	}
	if len(p.Roles) != len(wantPersonalities) {
		t.Fatalf("expected %d roles, got %d", len(wantPersonalities), len(p.Roles))
	}
	for name, want := range wantPersonalities {
		role, ok := p.Roles[name]
		if !ok {
			t.Fatalf("missing role %q", name)
		}
		if !reflect.DeepEqual(role.Personalities, want) {
			t.Errorf("role %q personalities = %q, want %q", name, role.Personalities, want)
		}
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

	for _, stub := range []string{"designer", "social", "sales", "customer-success"} {
		role, ok := p.Roles[stub]
		if !ok || role.Purpose == "" || len(role.Seats) != 0 {
			t.Fatalf("seatless role %q must exist with purpose and no seats: %+v", stub, role)
		}
	}
	if got := p.Roles["designer"].Purpose; got != "Product shaping." {
		t.Fatalf("designer purpose = %q", got)
	}
	if got := p.Roles["customer-success"].Purpose; got != "Onboarding, support, retention, customer research, and feeding recurring customer pain back into product work." {
		t.Fatalf("customer-success purpose = %q", got)
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
