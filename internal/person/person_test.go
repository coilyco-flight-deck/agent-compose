package person

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"testing/fstest"

	"forgejo.coilysiren.me/coilyco-flight-deck/agent-compose/internal/schema"
)

func TestLoadEmbeddedRoster(t *testing.T) {
	p, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	seatNames := map[string]string{}
	for _, roleName := range p.RoleOrder {
		role := p.Roles[roleName]
		if role.Skill != "role-"+roleName ||
			role.SkillSource == "" ||
			!strings.HasPrefix(role.SkillDigest, "sha256:") {
			t.Errorf("role %q skill provenance is incomplete: %+v", roleName, role)
		}
		if raw, ok := p.RoleSkillDefinition(roleName); !ok ||
			!strings.Contains(string(raw), "\nname: "+role.Skill+"\n") {
			t.Errorf("role %q canonical skill is missing or mismatched", roleName)
		}
		if got := briefingParagraphCount(role.Briefing); got < 3 {
			t.Errorf("role %q briefing has %d paragraphs, want at least three", roleName, got)
		}
		if len(role.Seats) < 2 {
			t.Errorf("role %q has %d seats, want at least claude and codex", roleName, len(role.Seats))
		}
		seats := map[string]Seat{}
		for _, seat := range role.Seats {
			seats[seat.Harness] = seat
			if prior, exists := seatNames[seat.Name]; exists {
				t.Errorf("roles %q and %q share seat name %q", prior, roleName, seat.Name)
			}
			seatNames[seat.Name] = roleName
		}
		for harness, pronouns := range map[string]string{"claude": "she", "codex": "he"} {
			seat, exists := seats[harness]
			if !exists || seat.Name == "" || seat.Pronouns != pronouns {
				t.Errorf("role %q %s seat is incomplete: %+v", roleName, harness, seat)
			}
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

func TestCEORoleSkillDefinesLongHorizonConcentration(t *testing.T) {
	p, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	role := p.Roles["ceo"]
	for _, required := range []string{
		"1 to 3 years, portfolio thesis",
		"2 to 4 quarters, strategic bets",
		"Current quarter, capital allocation",
		"authenticated portfolio scorecard and in-flight-work inventory",
		"core compounding assets, strategic options, experiments, and retirement candidates",
		"work or capacity it displaces",
		"PM owns discovery, program decomposition, sequencing, and learning loops",
		"decline a disconnected product experiment",
		"double-down, reshape, or exit",
	} {
		if !strings.Contains(role.Briefing, required) {
			t.Errorf("CEO role skill omitted %q", required)
		}
	}
}

func TestLoadRoleSkillsRejectsMissingAndMalformedDefinitions(t *testing.T) {
	for name, files := range map[string]fstest.MapFS{
		"missing": {},
		"mismatched frontmatter": {
			"roles/builder/SKILL.md": {
				Data: []byte("---\nname: role-other\ndescription: Wrong.\n---\n\nOne.\n\nTwo.\n\nThree.\n"),
			},
		},
	} {
		t.Run(name, func(t *testing.T) {
			p := &Person{
				Name:      "fixture",
				RoleOrder: []string{"builder"},
				Roles: map[string]Role{
					"builder": {Skill: "role-builder"},
				},
			}
			if err := loadRoleSkills(files, p); err == nil {
				t.Fatal("invalid role skill must fail")
			}
		})
	}
}

func TestLookupCueUsesDeclaredAliasesAndPreservesAmbiguity(t *testing.T) {
	p, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	for _, test := range []struct {
		cue  string
		want []string
	}{
		{cue: "calm", want: []string{"grounded"}},
		{cue: " encouraging ", want: []string{"optimistic", "nurturing"}},
		{cue: "grounded", want: []string{"grounded"}},
	} {
		got, lookupErr := p.LookupCue(test.cue)
		if lookupErr != nil {
			t.Fatalf("lookup %q: %v", test.cue, lookupErr)
		}
		if strings.Join(got, ",") != strings.Join(test.want, ",") {
			t.Errorf("lookup %q = %v, want %v", test.cue, got, test.want)
		}
	}
	if _, err := NormalizeCue("bad\x00cue"); err == nil {
		t.Fatal("control characters must be rejected")
	}
}

func TestLoadDirectoryKeepsExternalPersonIndependent(t *testing.T) {
	p, err := LoadDirectory(filepath.Join("..", "..", "testdata", "contracts", "person-independent"))
	if err != nil {
		t.Fatal(err)
	}
	if p.Name != "workbench" || len(p.Roles) != 1 {
		t.Fatalf("external person identity = %q roles=%d", p.Name, len(p.Roles))
	}
	if _, ok := p.Roles["builder"]; !ok {
		t.Fatalf("external person omitted builder role: %+v", p.RoleOrder)
	}
	if _, inherited := p.Roles["engineer"]; inherited {
		t.Fatal("external person inherited the embedded engineer role")
	}
	src, err := Source(p)
	if err != nil {
		t.Fatal(err)
	}
	if src.ID != "person:workbench" || len(src.Skills) != 2 {
		t.Fatalf("external person source = %q skills=%d", src.ID, len(src.Skills))
	}
}

func TestExternalProfileComposesLibrariesSeatsAndCopyContracts(t *testing.T) {
	profile := filepath.Join("..", "..", "examples", "person-profile")
	library := filepath.Join("..", "..", "examples", "shared-personality-library")
	p, err := LoadDirectoryWithLibraries(profile, library)
	if err != nil {
		t.Fatal(err)
	}
	if p.Name != "example" ||
		len(p.Roles["caption-review"].Personalities) != 1 ||
		strings.Join(p.Roles["bulk-captioner"].Personalities, ",") != "local-guide,shared-care" {
		t.Fatalf("external effective graph is wrong: %+v", p.Roles)
	}
	role := p.Roles["bulk-captioner"]
	if role.Seats[0].Selector() != "chatbot-sonnet-low" ||
		role.Seats[0].Channel != "chatbot" ||
		role.Seats[0].Tier != "sonnet-low" ||
		role.Seats[0].Pronouns != "they" {
		t.Fatalf("generalized seat was not preserved: %+v", role.Seats)
	}
	if role.CopyContract == nil ||
		role.CopyContract.Source != "person:example:role:bulk-captioner:copy-contract" ||
		!strings.HasPrefix(role.CopyContract.Digest, "sha256:") {
		t.Fatalf("copy contract provenance is incomplete: %+v", role.CopyContract)
	}
	personalities, err := p.PersonalityCatalog(nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(personalities) != 3 ||
		personalities[0].Description == "" ||
		personalities[0].SourceLibrary == "" ||
		!strings.HasPrefix(personalities[0].Digest, "sha256:") {
		t.Fatalf("personality catalogue is incomplete: %+v", personalities)
	}
	var unusedFound bool
	for _, entry := range personalities {
		if entry.Slug == "unused-spark" {
			unusedFound = true
			if len(entry.Affinities) != 0 {
				t.Fatalf("unused personality acquired affinities: %+v", entry.Affinities)
			}
		}
	}
	if !unusedFound {
		t.Fatal("effective catalogue omitted unused personality")
	}
	roles, err := p.RoleCatalog()
	if err != nil {
		t.Fatal(err)
	}
	if len(roles) != 2 ||
		roles[0].Skill == "" ||
		roles[0].SkillSource == "" ||
		roles[0].FavoriteColor == "" {
		t.Fatalf("role catalogue is incomplete: %+v", roles)
	}
	seats, err := p.SeatCatalog("bulk-captioner")
	if err != nil {
		t.Fatal(err)
	}
	if len(seats) != 1 || seats[0].Role != "bulk-captioner" ||
		seats[0].Seat.Selector() != "chatbot-sonnet-low" {
		t.Fatalf("seat catalogue is incomplete: %+v", seats)
	}
}

func TestExternalLibraryRejectsSymlinks(t *testing.T) {
	link := filepath.Join(t.TempDir(), "linked-library")
	target, err := filepath.Abs(filepath.Join("..", "..", "examples", "shared-personality-library"))
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(target, link); err != nil {
		t.Skipf("symlinks unavailable: %v", err)
	}
	_, err = LoadDirectoryWithLibraries(
		filepath.Join("..", "..", "examples", "person-profile"),
		link,
	)
	if err == nil || !strings.Contains(err.Error(), "symlink") {
		t.Fatalf("symlinked personality library passed: %v", err)
	}
}

func TestExternalLibraryConflictDiagnostics(t *testing.T) {
	profile := filepath.Join("..", "..", "examples", "person-profile")
	library := filepath.Join("..", "..", "examples", "shared-personality-library")
	if _, err := LoadDirectoryWithLibraries(profile, filepath.Join(t.TempDir(), "missing")); err == nil ||
		!strings.Contains(err.Error(), "inspect personality library") {
		t.Fatalf("missing library diagnostic = %v", err)
	}
	if _, err := LoadDirectoryWithLibraries(profile, library, library); err == nil ||
		!strings.Contains(err.Error(), "admitted more than once") {
		t.Fatalf("duplicate library diagnostic = %v", err)
	}

	invalid := filepath.Join(t.TempDir(), "invalid")
	copyTestTree(t, library, invalid)
	if err := os.WriteFile(filepath.Join(invalid, "library.kdl"), []byte(`library "Bad ID"`+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := LoadDirectoryWithLibraries(profile, invalid); err == nil ||
		!strings.Contains(err.Error(), "stable logical id") {
		t.Fatalf("invalid library id diagnostic = %v", err)
	}

	divergent := filepath.Join(t.TempDir(), "divergent")
	copyTestTree(t, library, divergent)
	if err := os.WriteFile(filepath.Join(divergent, "library.kdl"), []byte(`library "shared-example-copy"`+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	skillPath := filepath.Join(divergent, "definitions", "skills", "personality-shared-care", "SKILL.md")
	raw, err := os.ReadFile(skillPath)
	if err != nil {
		t.Fatal(err)
	}
	raw = append(raw, []byte("\nDivergent copy.\n")...)
	if err := os.WriteFile(skillPath, raw, 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := LoadDirectoryWithLibraries(profile, library, divergent); err == nil ||
		!strings.Contains(err.Error(), "definition") ||
		!strings.Contains(err.Error(), "conflicts") {
		t.Fatalf("divergent definition diagnostic = %v", err)
	}
}

func copyTestTree(t *testing.T, source, target string) {
	t.Helper()
	if err := fs.WalkDir(os.DirFS(source), ".", func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		destination := filepath.Join(target, filepath.FromSlash(path))
		if entry.IsDir() {
			return os.MkdirAll(destination, 0o755)
		}
		raw, err := fs.ReadFile(os.DirFS(source), path)
		if err != nil {
			return err
		}
		return os.WriteFile(destination, raw, 0o644)
	}); err != nil {
		t.Fatal(err)
	}
}

func TestAssemblePersonSourceSet(t *testing.T) {
	source := fstest.MapFS{
		"person.kdl": {
			Data: []byte(`person "fixture"`),
		},
		"roles/01-builder.kdl": {
			Data: []byte(`role "builder"`),
		},
		"personalities/01-bright.kdl": {
			Data: []byte(`personality "bright"`),
		},
		"inspirations/01-fixture-builder.kdl": {
			Data: []byte(`inspiration "fixture-builder"`),
		},
	}
	raw, err := assemblePersonSource(source, "fixture person source")
	if err != nil {
		t.Fatal(err)
	}
	want := `person "fixture" {
    role "builder"

    personality "bright"

    inspiration "fixture-builder"
}
`
	if string(raw) != want {
		t.Fatalf("assembled source:\n%s\nwant:\n%s", raw, want)
	}
}

func TestAssemblePersonSourceRejectsMisfiledFragment(t *testing.T) {
	source := fstest.MapFS{
		"person.kdl": {
			Data: []byte(`person "fixture"`),
		},
		"roles/01-builder.kdl": {
			Data: []byte(`role "other"`),
		},
	}
	_, err := assemblePersonSource(source, "fixture person source")
	if err == nil || !strings.Contains(err.Error(), "filename does not match role") {
		t.Fatalf("misfiled fragment error = %v", err)
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

            Keep the work bounded and test the important paths.

            Finish validation and hand back a complete result.
            """
        personality "bright" "steady"
        model-class "frontier"
        inspiration "fixture-builder" {
            fit "The fixture is a useful builder archetype."
        }
    }
    personality "bright" skill="personality-bright" color="#d98e48" motif="sunbeam" {
        emblem { name "lantern"; emoji "🏮"; glyph "✦" }
        form { silhouette "beacon"; geometry "open-rays"; motion "glowing" }
        sound-mark { timbre "bell"; contour "rising"; pulse "triplet" }
        inspiration "fixture-builder" {
            fit "The fixture demonstrates brightness."
        }
    }
    personality "steady" skill="personality-steady" color="#5fa87a" motif="stone" {
        emblem { name "anchor"; emoji "⚓"; glyph "◆" }
        form { silhouette "cairn"; geometry "stacked-rounds"; motion "settling" }
        sound-mark { timbre "wood-block"; contour "returning"; pulse "steady-pair" }
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
		"Keep the work bounded and test the important paths.\n\n" +
		"Finish validation and hand back a complete result."
	if got := p.Roles["builder"].Briefing; got != wantBriefing {
		t.Fatalf("role briefing = %q, want %q", got, wantBriefing)
	}
	modelClasses := p.Roles["builder"].SupportedModelClasses
	if len(modelClasses) != 1 || modelClasses[0] != schema.ModelClassFrontier {
		t.Fatalf("role model classes = %q", modelClasses)
	}
	if got := p.Personalities["bright"]; got.Motif != "sunbeam" ||
		got.Emblem.Name != "lantern" || got.Form.Silhouette != "beacon" ||
		got.SoundMark.Timbre != "bell" {
		t.Fatalf("personality identity = %+v", got)
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
    personality "bright" skill="personality-bright" color="#d98e48" motif="sunbeam" {
        emblem { name "lantern"; emoji "🏮"; glyph "✦" }
        form { silhouette "beacon"; geometry "open-rays"; motion "glowing" }
        sound-mark { timbre "bell"; contour "rising"; pulse "triplet" }
        inspiration "fixture-builder" {
            fit "The fixture demonstrates brightness."
        }
    }`, `
    personality "bright" skill="personality-bright" color="#d98e48" motif="sunbeam" {
        emblem { name "lantern"; emoji "🏮"; glyph "✦" }
        form { silhouette "beacon"; geometry "open-rays"; motion "glowing" }
        sound-mark { timbre "bell"; contour "rising"; pulse "triplet" }
    }`, 1),
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

func TestParseRejectsIncompleteOrAmbiguousIdentity(t *testing.T) {
	valid := inspirationFixture()
	cases := map[string]struct {
		body string
		want string
	}{
		"missing motif": {
			body: strings.Replace(valid, ` motif="sunbeam"`, "", 1),
			want: `personality "bright" needs a semantic motif property`,
		},
		"incomplete emblem": {
			body: strings.Replace(valid, `; glyph "✦"`, "", 1),
			want: `personality bright emblem needs glyph`,
		},
		"invalid form token": {
			body: strings.Replace(valid, `"open-rays"`, `"Open rays"`, 1),
			want: `personality bright form geometry needs a lowercase semantic token`,
		},
		"duplicate emblem": {
			body: strings.Replace(valid, `name "anchor"`, `name "lantern"`, 1),
			want: `share identity value "lantern"`,
		},
		"duplicate emoji": {
			body: strings.Replace(valid, `emoji "⚓"`, `emoji "🏮"`, 1),
			want: `share identity value "🏮"`,
		},
		"duplicate motif": {
			body: strings.Replace(valid, `motif="stone"`, `motif="sunbeam"`, 1),
			want: `share identity value "sunbeam"`,
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			if _, err := parse([]byte(tc.body)); err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("identity parse error = %v, want %q", err, tc.want)
			}
		})
	}
}

func inspirationFixture() string {
	return `person "fixture" {
    role "builder" {
        purpose "Build."
        briefing "Build independently.\n\nVerify the important paths.\n\nFinish the complete handoff."
        personality "bright" "steady"
        inspiration "fixture-builder" {
            fit "The fixture is a useful builder archetype."
        }
    }
    personality "bright" skill="personality-bright" color="#d98e48" motif="sunbeam" {
        emblem { name "lantern"; emoji "🏮"; glyph "✦" }
        form { silhouette "beacon"; geometry "open-rays"; motion "glowing" }
        sound-mark { timbre "bell"; contour "rising"; pulse "triplet" }
        inspiration "fixture-builder" {
            fit "The fixture demonstrates brightness."
        }
    }
    personality "steady" skill="personality-steady" color="#5fa87a" motif="stone" {
        emblem { name "anchor"; emoji "⚓"; glyph "◆" }
        form { silhouette "cairn"; geometry "stacked-rounds"; motion "settling" }
        sound-mark { timbre "wood-block"; contour "returning"; pulse "steady-pair" }
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
			want: "needs a role skill or legacy briefing",
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
		"fewer than three paragraphs": {
			body: `person "fixture" {
    role "builder" {
        purpose "Build."
        briefing "Inspect the work.\n\nFinish the work."
        personality "bright" "steady"
    }
}`,
			want: "briefing needs at least three paragraphs, got 2",
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

func TestParseRejectsInvalidRoleModelClasses(t *testing.T) {
	valid := inspirationFixture()
	cases := map[string]struct {
		body string
		want string
	}{
		"empty": {
			body: strings.Replace(valid, `personality "bright" "steady"`,
				"personality \"bright\" \"steady\"\n        model-class", 1),
			want: "model-class needs at least one argument",
		},
		"unsupported": {
			body: strings.Replace(valid, `personality "bright" "steady"`,
				"personality \"bright\" \"steady\"\n        model-class \"tiny\"", 1),
			want: `unsupported model class "tiny"`,
		},
		"repeated": {
			body: strings.Replace(valid, `personality "bright" "steady"`,
				"personality \"bright\" \"steady\"\n        model-class \"frontier\" \"frontier\"", 1),
			want: `repeats model class "frontier"`,
		},
		"duplicate node": {
			body: strings.Replace(valid, `personality "bright" "steady"`,
				"personality \"bright\" \"steady\"\n        model-class \"frontier\"\n        model-class \"low-context\"", 1),
			want: "duplicate model-class",
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			if _, err := parse([]byte(tc.body)); err == nil ||
				!strings.Contains(err.Error(), tc.want) {
				t.Fatalf("model-class parse error = %v, want %q", err, tc.want)
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

func TestParseCopyContractValidation(t *testing.T) {
	valid := inspirationFixture()
	insert := func(contract string) string {
		return strings.Replace(
			valid,
			`        personality "bright" "steady"`,
			contract+"\n        "+`personality "bright" "steady"`,
			1,
		)
	}
	for name, tc := range map[string]struct {
		body string
		want string
	}{
		"unsupported scope": {
			body: insert(`        copy-contract scope="chat" { forbid "asset" prefer="video" }`),
			want: "supported scope tool-response",
		},
		"missing replacement": {
			body: insert(`        copy-contract scope="tool-response" { forbid "asset" }`),
			want: "needs prefer",
		},
		"normalized duplicate": {
			body: insert(`        copy-contract scope="tool-response" {
            forbid "asset" prefer="video"
            forbid "Asset" prefer="media"
        }`),
			want: "repeats forbid",
		},
		"conflicting declarations": {
			body: insert(`        copy-contract scope="tool-response" { forbid "asset" prefer="video" }
        copy-contract scope="tool-response" { forbid "upload" prefer="add" }`),
			want: "duplicate copy-contract",
		},
	} {
		t.Run(name, func(t *testing.T) {
			if _, err := parse([]byte(tc.body)); err == nil ||
				!strings.Contains(err.Error(), tc.want) {
				t.Fatalf("copy-contract parse error = %v, want %q", err, tc.want)
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
		"unknown personalities": `person "fixture" {
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
