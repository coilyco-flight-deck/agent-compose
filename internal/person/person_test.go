package person

import (
	"strings"
	"testing"
)

func TestLoadEmbeddedRoster(t *testing.T) {
	p, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	for _, roleName := range p.RoleOrder {
		role := p.Roles[roleName]
		if got := strings.Count(role.Briefing, "\n\n"); got != 1 {
			t.Errorf("role %q briefing has %d paragraph breaks, want 1", roleName, got)
		}
		for _, name := range role.Personalities {
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
        briefing """
            You are a builder. Build the fixture from repository evidence.

            Finish validation and hand back a complete result.
            """
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
	wantBriefing := "You are a builder. Build the fixture from repository evidence.\n\n" +
		"Finish validation and hand back a complete result."
	if got := p.Roles["builder"].Briefing; got != wantBriefing {
		t.Fatalf("role briefing = %q, want %q", got, wantBriefing)
	}
}

func TestParseRejectsInvalidRoleBriefing(t *testing.T) {
	cases := map[string]struct {
		body string
		want string
	}{
		"missing": {
			body: `person "fixture" {
    role "builder" {
        purpose "Build."
        personality "bright" "steady"
    }
    personality "bright" skill="personality-bright" color="#d98e48"
    personality "steady" skill="personality-steady" color="#5fa87a"
}`,
			want: "needs a briefing",
		},
		"empty": {
			body: `person "fixture" {
    role "builder" {
        purpose "Build."
        briefing "   "
        personality "bright" "steady"
    }
    personality "bright" skill="personality-bright" color="#d98e48"
    personality "steady" skill="personality-steady" color="#5fa87a"
}`,
			want: "briefing must not be empty",
		},
		"duplicate": {
			body: `person "fixture" {
    role "builder" {
        purpose "Build."
        briefing "First."
        briefing "Second."
        personality "bright" "steady"
    }
    personality "bright" skill="personality-bright" color="#d98e48"
    personality "steady" skill="personality-steady" color="#5fa87a"
}`,
			want: "duplicate briefing",
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			if _, err := parse([]byte(tc.body)); err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("invalid role briefing error = %v, want %q", err, tc.want)
			}
		})
	}
}

func TestParseRejectsIllegibleColor(t *testing.T) {
	body := `person "fixture" {
    role "builder" {
        purpose "Build."
        briefing "Build independently."
        personality "bright" "steady"
    }
    personality "bright" skill="personality-bright" color="#111111"
    personality "steady" skill="personality-steady" color="#5fa87a"
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
        briefing "Build independently."
        agent "claude" name="first builder"
        agent "claude" name="another"
    }
}`,
		"seat without name": `person "fixture" {
    role "builder" {
        purpose "Build."
        briefing "Build independently."
        agent "claude"
    }
}`,
		"unknown role child": `person "fixture" {
    role "builder" {
        briefing "Build independently."
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
        briefing "Build independently."
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
        briefing "Build independently."
        personality "bright"
    }
}`,
		"four personalities": `person "fixture" {
    role "builder" {
        briefing "Build independently."
        personality "one" "two" "three" "four"
    }
}`,
		"duplicate personality": `person "fixture" {
    role "builder" {
        briefing "Build independently."
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
