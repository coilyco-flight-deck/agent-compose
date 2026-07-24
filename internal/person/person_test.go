package person

import (
	"io/fs"
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
		if got := strings.Count(role.Briefing, "\n\n"); got < 1 {
			t.Errorf("role %q briefing has no paragraph break", roleName)
		}
		if _, ok := p.Inspirations[role.Inspiration.ID]; !ok {
			t.Errorf("role %q inspiration %q has no catalog entry", roleName, role.Inspiration.ID)
		}
		for _, name := range role.Personalities {
			binding, ok := p.Personalities[name]
			if !ok {
				t.Fatalf("role %q personality %q has no catalog binding", roleName, name)
			}
			if want := "personality-" + name; binding.Skill != want {
				t.Errorf("personality %q skill = %q, want %q", name, binding.Skill, want)
			}
			if _, ok := p.Inspirations[binding.Inspiration.ID]; !ok {
				t.Errorf("personality %q inspiration %q has no catalog entry", name, binding.Inspiration.ID)
			}
		}
	}
	for _, id := range p.InspirationOrder {
		inspiration := p.Inspirations[id]
		if inspiration.Name == "" || inspiration.Achievement == "" ||
			inspiration.ImpactMode == "" || inspiration.ImpactFit == "" ||
			inspiration.ProfileCitation == "" {
			t.Errorf("inspiration %q is incomplete: %+v", id, inspiration)
		}
		if inspiration.Appearance.ID == "" || inspiration.Appearance.Summary == "" ||
			len(inspiration.Appearance.Citations) == 0 {
			t.Errorf("inspiration %q appearance is incomplete: %+v", id, inspiration.Appearance)
		}
	}
}

func TestEmbeddedPersonSourceOwnsEveryPersonalityDefinition(t *testing.T) {
	p, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	src, err := Source(p)
	if err != nil {
		t.Fatal(err)
	}
	if src.ID != "person:kai" {
		t.Fatalf("source id = %q, want person:kai", src.ID)
	}
	if len(src.Instructions) != 1 || src.Instructions[0].ID != "personality-invariant" {
		t.Fatalf("unexpected embedded instructions: %+v", src.Instructions)
	}
	if len(src.Skills) != len(p.Personalities) {
		t.Fatalf("embedded skills = %d, catalog bindings = %d", len(src.Skills), len(p.Personalities))
	}
	for _, skill := range src.Skills {
		if _, err := fs.Stat(src.FileSystem(), skill.Path+"/SKILL.md"); err != nil {
			t.Fatalf("definition %q is not embedded: %v", skill.ID, err)
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
        inspiration "fixture-builder" {
            fit "The fixture is a useful builder archetype."
        }
    }
    personality "bright" skill="personality-bright" color="#d98e48" {
        inspiration "fixture-builder" {
            fit "The fixture demonstrates brightness."
        }
    }
    personality "steady" skill="personality-steady" color="#5fa87a" {
        inspiration "fixture-builder" {
            fit "The fixture demonstrates steadiness."
        }
    }
    inspiration "fixture-builder" name="Fixture Builder" profile-citation="fixture-builder-profile" impact-mode="fixture-building" {
        achievement "The fixture builder made the parser test concrete."
        impact-fit "The fixture builder creates impact by keeping the successful parse path complete."
        appearance "fixture-talk" title="Building Fixtures" event="Fixture Conference" year="2026" format="keynote" {
            summary "The fixture builder explains how a complete person source stays internally consistent."
            citation "fixture-builder-talk"
        }
    }
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

func TestParseRejectsBrokenInspirationRelationships(t *testing.T) {
	valid := inspirationFixture()
	cases := map[string]struct {
		body string
		want string
	}{
		"role missing inspiration": {
			body: strings.Replace(valid, `
        inspiration "fixture-builder" {
            fit "The fixture is a useful builder archetype."
        }`, "", 1),
			want: `role "builder" needs an inspiration`,
		},
		"unknown role inspiration": {
			body: strings.Replace(valid, `inspiration "fixture-builder" {`, `inspiration "missing-builder" {`, 1),
			want: `role "builder": inspiration "missing-builder" has no catalog entry`,
		},
		"personality missing inspiration": {
			body: strings.Replace(valid, `
    personality "bright" skill="personality-bright" color="#d98e48" {
        inspiration "fixture-builder" {
            fit "The fixture demonstrates brightness."
        }
    }`, `
    personality "bright" skill="personality-bright" color="#d98e48"`, 1),
			want: `personality "bright" needs an inspiration`,
		},
		"appearance missing citation": {
			body: strings.Replace(valid, `
            citation "fixture-builder-talk"`, "", 1),
			want: `appearance "fixture-talk" needs at least one citation`,
		},
		"unreferenced inspiration": {
			body: strings.TrimSuffix(valid, "\n}") + `
    inspiration "unused" name="Unused" profile-citation="unused-profile" impact-mode="unused" {
        achievement "Unused achievement."
        impact-fit "Unused impact."
        appearance "unused-talk" title="Unused" event="Fixture Conference" year="2026" format="keynote" {
            summary "Unused summary."
            citation "unused-talk"
        }
    }
}`,
			want: `inspiration "unused" is not used by a role or personality`,
		},
		"duplicate credited person": {
			body: strings.TrimSuffix(valid, "\n}") + `
    inspiration "duplicate" name="Fixture Builder" profile-citation="duplicate-profile" impact-mode="duplicate" {
        achievement "Duplicate achievement."
        impact-fit "Duplicate impact."
        appearance "duplicate-talk" title="Duplicate" event="Fixture Conference" year="2026" format="keynote" {
            summary "Duplicate summary."
            citation "duplicate-talk"
        }
    }
}`,
			want: `inspirations "fixture-builder" and "duplicate" name the same person`,
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			if _, err := parse([]byte(tc.body)); err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("parse error = %v, want substring %q", err, tc.want)
			}
		})
	}
}

func inspirationFixture() string {
	return `person "fixture" {
    role "builder" {
        purpose "Build."
        briefing "Build independently."
        personality "bright" "steady"
        inspiration "fixture-builder" {
            fit "The fixture is a useful builder archetype."
        }
    }
    personality "bright" skill="personality-bright" color="#d98e48" {
        inspiration "fixture-builder" {
            fit "The fixture demonstrates brightness."
        }
    }
    personality "steady" skill="personality-steady" color="#5fa87a" {
        inspiration "fixture-builder" {
            fit "The fixture demonstrates steadiness."
        }
    }
    inspiration "fixture-builder" name="Fixture Builder" profile-citation="fixture-builder-profile" impact-mode="fixture-building" {
        achievement "The fixture builder made the parser test concrete."
        impact-fit "The fixture builder creates impact by keeping the successful parse path complete."
        appearance "fixture-talk" title="Building Fixtures" event="Fixture Conference" year="2026" format="keynote" {
            summary "The fixture builder explains how a complete person source stays internally consistent."
            citation "fixture-builder-talk"
        }
    }
}`
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
