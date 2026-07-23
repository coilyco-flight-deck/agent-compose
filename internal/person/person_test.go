package person

import "testing"

func TestLoadEmbeddedRoster(t *testing.T) {
	p, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	for _, roleName := range p.RoleOrder {
		for _, name := range p.Roles[roleName].Personalities {
			binding, ok := p.Personalities[name]
			if !ok {
				t.Fatalf("role %q personality %q has no catalog binding", roleName, name)
			}
			if want := "personality-" + name; binding.Skill != want {
				t.Errorf("personality %q skill = %q, want %q", name, binding.Skill, want)
			}
		}
	}
}

func TestParsePreservesRolePersonalities(t *testing.T) {
	body := `person "fixture" {
    role "builder" {
        purpose "Build."
        personality "bright" "steady"
    }
    personality "bright" skill="personality-bright" color="#d98e48"
    personality "steady" skill="personality-steady" color="#5fa87a"
}`
	p, err := parse([]byte(body))
	if err != nil {
		t.Fatal(err)
	}
	got := p.Roles["builder"].Personalities
	if len(got) != 2 || got[0] != "bright" || got[1] != "steady" {
		t.Fatalf("role personalities = %q", got)
	}
}

func TestParseRejectsIllegibleColor(t *testing.T) {
	body := `person "fixture" {
    role "builder" {
        purpose "Build."
        personality "bright"
    }
    personality "bright" skill="personality-bright" color="#111111"
}`
	if _, err := parse([]byte(body)); err == nil {
		t.Fatal("an illegible color must fail the parse-time gate")
	}
}

func TestParseRejectsMissingColor(t *testing.T) {
	body := `person "fixture" {
    personality "bright" skill="personality-bright"
}`
	if _, err := parse([]byte(body)); err == nil {
		t.Fatal("a personality without a color must fail the parse-time gate")
	}
}

func TestParseSeatValidation(t *testing.T) {
	cases := map[string]string{
		"duplicate seat": `person "fixture" {
    role "builder" {
        purpose "Build."
        agent "claude" name="first builder"
        agent "claude" name="another"
    }
}`,
		"seat without name": `person "fixture" {
    role "builder" {
        purpose "Build."
        agent "claude"
    }
}`,
		"unknown role child": `person "fixture" {
    role "builder" {
        guardfile "fixture.kdl"
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

func TestParseRejectsRolePersonalityWithoutCatalogBinding(t *testing.T) {
	body := `person "fixture" {
    role "builder" {
        purpose "Build."
        personality "bright" "steady"
    }
}`
	if _, err := parse([]byte(body)); err == nil {
		t.Fatal("role personality without a catalog binding must fail")
	}
}

func TestParseRejectsRolePersonalityCardinalityAndDuplicates(t *testing.T) {
	cases := map[string]string{
		"one personality": `person "fixture" {
    role "builder" {
        personality "bright"
    }
}`,
		"four personalities": `person "fixture" {
    role "builder" {
        personality "one" "two" "three" "four"
    }
}`,
		"duplicate personality": `person "fixture" {
    role "builder" {
        personality "bright" "bright"
    }
}`,
	}
	for name, body := range cases {
		t.Run(name, func(t *testing.T) {
			if _, err := parse([]byte(body)); err == nil {
				t.Fatal("invalid role personality set must fail")
			}
		})
	}
}
