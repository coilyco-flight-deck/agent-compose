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
	ai := p.Roles["ai"]
	for _, method := range ai.Methods {
		if raw, ok := p.RoleMethodDefinition("ai", method); !ok ||
			!strings.Contains(string(raw), "\nname: "+method+"\n") {
			t.Errorf("AI role method %q is missing or mismatched", method)
		}
	}
	identityNames := map[string]string{}
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
		if role.Identity == nil || role.Identity.Name == "" || role.Identity.Pronouns == "" {
			t.Errorf("role %q identity is incomplete: %+v", roleName, role.Identity)
			continue
		}
		if prior, exists := identityNames[role.Identity.Name]; exists {
			t.Errorf("roles %q and %q share identity name %q", prior, roleName, role.Identity.Name)
		}
		identityNames[role.Identity.Name] = roleName
		seats := map[string]Seat{}
		for _, seat := range role.Seats {
			seats[seat.Harness] = seat
			if seat.Name != role.Identity.Name || seat.Pronouns != role.Identity.Pronouns {
				t.Errorf("role %q seat %q redefines identity: %+v", roleName, seat.Selector(), seat)
			}
		}
		for _, harness := range []string{"claude", "codex"} {
			seat, exists := seats[harness]
			if !exists || seat.Name != role.Identity.Name || seat.Pronouns != role.Identity.Pronouns {
				t.Errorf("role %q %s seat is incomplete: %+v", roleName, harness, seat)
			}
		}
		if role.Inspiration.ID != "" {
			t.Errorf("role %q retains Core inspiration %q", roleName, role.Inspiration.ID)
		}
		for _, name := range role.Personalities {
			binding, ok := p.Personalities[name]
			if !ok {
				t.Fatalf("role %q personality %q has no catalog binding", roleName, name)
			}
			if want := "personality-" + name; binding.Skill != want {
				t.Errorf("personality %q skill = %q, want %q", name, binding.Skill, want)
			}
			if binding.Inspiration.ID != "" {
				t.Errorf("personality %q retains Core inspiration %q", name, binding.Inspiration.ID)
			}
		}
	}
	if len(p.Inspirations) != 0 || len(p.InspirationOrder) != 0 {
		t.Fatalf("Core retains inspirations: order=%v catalog=%v", p.InspirationOrder, p.Inspirations)
	}
}

func TestValidateCoreBoundariesRejectsUnbalancedRoster(t *testing.T) {
	t.Run("wrong slot count", func(t *testing.T) {
		p, err := Load()
		if err != nil {
			t.Fatal(err)
		}
		role := p.Roles["engineer"]
		role.Personalities = append(role.Personalities, "grounded")
		p.Roles["engineer"] = role
		if err := validateCorePersonalityMelds(p); err == nil ||
			!strings.Contains(err.Error(), "want exactly three") {
			t.Fatalf("slot-count validation error = %v", err)
		}
	})

	t.Run("overused personality", func(t *testing.T) {
		p, err := Load()
		if err != nil {
			t.Fatal(err)
		}
		// meticulous already sits at the cap of three, so a fourth role tips it.
		role := p.Roles["ops"]
		role.Personalities[2] = "meticulous"
		p.Roles["ops"] = role
		if err := validateCorePersonalityMelds(p); err == nil ||
			!strings.Contains(err.Error(), "want at most three") {
			t.Fatalf("usage validation error = %v", err)
		}
	})

	t.Run("favorite colors too close", func(t *testing.T) {
		p, err := Load()
		if err != nil {
			t.Fatal(err)
		}
		// Two roles melding the same personalities land on the same anchor, so
		// the spread step can only separate them by twice the drift cap.
		role := p.Roles["director"]
		role.Personalities = append([]string(nil), p.Roles["engineer"].Personalities...)
		p.Roles["director"] = role
		if err := p.ResolveFavoriteColors(); err != nil {
			t.Fatal(err)
		}
		if err := validateCorePersonalityMelds(p); err == nil ||
			!strings.Contains(err.Error(), "apart at their closest") {
			t.Fatalf("favorite-color separation error = %v", err)
		}
	})
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

func TestLoadRoleSkillsEnforcesAuthoredBodyWordLimit(t *testing.T) {
	for name, words := range map[string]int{
		"at limit":   maxRoleSkillBodyWords,
		"over limit": maxRoleSkillBodyWords + 1,
	} {
		t.Run(name, func(t *testing.T) {
			files := fstest.MapFS{
				"roles/builder/SKILL.md": {
					Data: roleSkillWordLimitFixture(words),
				},
			}
			p := &Person{
				Name:      "fixture",
				RoleOrder: []string{"builder"},
				Roles: map[string]Role{
					"builder": {Skill: "role-builder"},
				},
			}
			err := loadRoleSkills(files, p)
			if words == maxRoleSkillBodyWords && err != nil {
				t.Fatalf("role skill at word limit failed: %v", err)
			}
			if words > maxRoleSkillBodyWords &&
				(err == nil || !strings.Contains(err.Error(), "maximum is 400")) {
				t.Fatalf("role skill over word limit error = %v", err)
			}
		})
	}
}

// TestBoundaryBodiesDoNotConsumeTheRoleWordBudget is the load-bearing guarantee:
// shared doctrine must not compete with the charter for the same 400 words.
func TestBoundaryBodiesDoNotConsumeTheRoleWordBudget(t *testing.T) {
	files := fstest.MapFS{
		"roles/builder/SKILL.md": {
			Data: roleSkillWordLimitFixture(maxRoleSkillBodyWords),
		},
		"definitions/skills/boundary-shared/SKILL.md": {
			Data: boundaryWordLimitFixture(maxBoundarySkillBodyWords),
		},
	}
	p := &Person{
		Name:      "fixture",
		RoleOrder: []string{"builder"},
		Roles: map[string]Role{
			"builder": {Skill: "role-builder", Boundaries: []string{"shared"}},
		},
		BoundaryOrder: []string{"shared"},
		Boundaries: map[string]Boundary{
			"shared": {Skill: "boundary-shared", Summary: "shared fixture doctrine"},
		},
	}
	if err := loadRoleSkills(files, p); err != nil {
		t.Fatalf("role skill at its word limit failed alongside a boundary: %v", err)
	}
	if err := loadBoundarySkills(files, p); err != nil {
		t.Fatalf("boundary at its own word limit failed: %v", err)
	}
	// Role at its cap plus boundary at its own proves the boundary is additive.
	if got := roleSkillBodyWordCount(p.Roles["builder"].Briefing); got != maxRoleSkillBodyWords {
		t.Fatalf("role briefing word count = %d, want %d", got, maxRoleSkillBodyWords)
	}
	if _, ok := p.BoundarySkillDefinition("shared"); !ok {
		t.Fatal("boundary body did not load")
	}

	over := fstest.MapFS{
		"definitions/skills/boundary-shared/SKILL.md": {
			Data: boundaryWordLimitFixture(maxBoundarySkillBodyWords + 1),
		},
	}
	err := loadBoundarySkills(over, p)
	if err == nil || !strings.Contains(err.Error(), "maximum is 400") {
		t.Fatalf("boundary over its own word limit error = %v", err)
	}
}

func TestValidateBoundaryOwnersRejectsIncoherentPairs(t *testing.T) {
	for name, p := range map[string]*Person{
		"missing owner": {
			RoleOrder: []string{"builder"},
			Roles: map[string]Role{
				"builder": {Skill: "role-builder", Boundaries: []string{"shared"}},
			},
			BoundaryOrder: []string{"shared"},
			Boundaries: map[string]Boundary{
				"shared": {Skill: "boundary-shared"},
			},
		},
		"unknown owner": {
			RoleOrder: []string{"builder"},
			Roles: map[string]Role{
				"builder": {Skill: "role-builder", Boundaries: []string{"shared"}},
			},
			BoundaryOrder: []string{"shared"},
			Boundaries: map[string]Boundary{
				"shared": {Skill: "boundary-shared", Owner: "absent"},
			},
		},
		"owner declares the boundary": {
			RoleOrder: []string{"builder", "mirror"},
			Roles: map[string]Role{
				"builder": {Skill: "role-builder", Boundaries: []string{"shared"}},
				"mirror":  {Skill: "role-mirror", Boundaries: []string{"shared"}},
			},
			BoundaryOrder: []string{"shared"},
			Boundaries: map[string]Boundary{
				"shared": {Skill: "boundary-shared", Owner: "mirror"},
			},
		},
	} {
		t.Run(name, func(t *testing.T) {
			if err := validateBoundaryOwners(p); err == nil {
				t.Fatal("incoherent boundary owner passed validation")
			}
		})
	}

	coherent := &Person{
		RoleOrder: []string{"builder", "mirror"},
		Roles: map[string]Role{
			"builder": {Skill: "role-builder", Boundaries: []string{"shared"}},
			"mirror":  {Skill: "role-mirror"},
		},
		BoundaryOrder: []string{"shared"},
		Boundaries: map[string]Boundary{
			"shared": {Skill: "boundary-shared", Owner: "mirror"},
		},
	}
	if err := validateBoundaryOwners(coherent); err != nil {
		t.Fatalf("coherent boundary owner rejected: %v", err)
	}
}

func TestValidateRoleAdjacentsRejectsIncompleteGraphs(t *testing.T) {
	for name, p := range map[string]*Person{
		"out-degree below the fixed width": {
			RoleOrder: []string{"builder", "mirror"},
			Roles: map[string]Role{
				"builder": {
					Skill:     "role-builder",
					Adjacents: []Adjacent{{Role: "mirror", Reason: "absorbs review"}},
				},
				"mirror": {
					Skill: "role-mirror",
					Adjacents: []Adjacent{
						{Role: "builder", Reason: "absorbs build"},
						{Role: "builder", Reason: "duplicate is caught at parse"},
					},
				},
			},
		},
		"partial roster leaves a role undeclared": {
			RoleOrder: []string{"builder", "mirror", "scribe"},
			Roles: map[string]Role{
				"builder": {
					Skill: "role-builder",
					Adjacents: []Adjacent{
						{Role: "mirror", Reason: "absorbs review"},
						{Role: "scribe", Reason: "absorbs the record"},
					},
				},
				"mirror": {
					Skill: "role-mirror",
					Adjacents: []Adjacent{
						{Role: "builder", Reason: "absorbs build"},
						{Role: "scribe", Reason: "absorbs the record"},
					},
				},
				"scribe": {Skill: "role-scribe"},
			},
		},
		"unknown adjacent target": {
			RoleOrder: []string{"builder", "mirror"},
			Roles: map[string]Role{
				"builder": {
					Skill: "role-builder",
					Adjacents: []Adjacent{
						{Role: "mirror", Reason: "absorbs review"},
						{Role: "absent", Reason: "names nothing"},
					},
				},
				"mirror": {
					Skill: "role-mirror",
					Adjacents: []Adjacent{
						{Role: "builder", Reason: "absorbs build"},
						{Role: "builder", Reason: "duplicate is caught at parse"},
					},
				},
			},
		},
	} {
		t.Run(name, func(t *testing.T) {
			if err := validateRoleAdjacents(p); err == nil {
				t.Fatal("incomplete adjacency graph passed validation")
			}
		})
	}

	complete := &Person{
		RoleOrder: []string{"builder", "mirror", "builder2"},
		Roles: map[string]Role{
			"builder": {
				Skill: "role-builder",
				Adjacents: []Adjacent{
					{Role: "mirror", Reason: "absorbs review"},
					{Role: "builder2", Reason: "absorbs the second thing"},
				},
			},
			"mirror": {
				Skill: "role-mirror",
				Adjacents: []Adjacent{
					{Role: "builder", Reason: "absorbs build"},
					{Role: "builder2", Reason: "absorbs the second thing"},
				},
			},
			"builder2": {
				Skill: "role-builder2",
				Adjacents: []Adjacent{
					{Role: "builder", Reason: "absorbs build"},
					{Role: "mirror", Reason: "absorbs review"},
				},
			},
		},
	}
	if err := validateRoleAdjacents(complete); err != nil {
		t.Fatalf("complete adjacency graph rejected: %v", err)
	}
}

// An external person package authored before the adjacency axis keeps loading.
// Adjacency is all-or-nothing across a roster rather than required of every one.
func TestValidateRoleAdjacentsSkipsRostersWithoutAdjacency(t *testing.T) {
	p := &Person{
		RoleOrder: []string{"builder", "mirror"},
		Roles: map[string]Role{
			"builder": {Skill: "role-builder"},
			"mirror":  {Skill: "role-mirror"},
		},
	}
	if err := validateRoleAdjacents(p); err != nil {
		t.Fatalf("roster without adjacency rejected: %v", err)
	}
}

// The core roster is the thing the evaluation board reads adjacency from, so a
// missing edge there is a silently thinner board rather than a load failure.
func TestCoreRosterDeclaresACompleteAdjacencyGraph(t *testing.T) {
	p, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	for _, roleName := range p.roleOrder() {
		role := p.Roles[roleName]
		if len(role.Adjacents) != adjacentsPerRole {
			t.Fatalf(
				"core role %q declares %d adjacent roles, want %d",
				roleName,
				len(role.Adjacents),
				adjacentsPerRole,
			)
		}
		for _, adjacent := range role.Adjacents {
			if adjacent.Role == roleName {
				t.Fatalf("core role %q names itself adjacent", roleName)
			}
			if _, ok := p.Roles[adjacent.Role]; !ok {
				t.Fatalf(
					"core role %q names unknown adjacent %q", roleName, adjacent.Role,
				)
			}
			if strings.TrimSpace(adjacent.Reason) == "" {
				t.Fatalf(
					"core role %q adjacent %q carries no reason",
					roleName,
					adjacent.Role,
				)
			}
		}
	}
}

func boundaryWordLimitFixture(words int) []byte {
	body := strings.TrimSpace(strings.Repeat("word ", words))
	return []byte(
		"---\nname: boundary-shared\ndescription: Shared fixture doctrine.\n---\n\n" +
			boundaryOwnHeading + "\n\nowner side.\n\n" +
			boundaryDeferHeading + "\n\n" + body + "\n",
	)
}

func roleSkillWordLimitFixture(words int) []byte {
	body := strings.TrimSpace(strings.Repeat("word ", words-2))
	return []byte(
		"---\nname: role-builder\ndescription: Build fixtures.\n---\n\n" +
			"# Builder\n\n" + body + "\n\nSecond.\n\nThird.\n",
	)
}

func TestLoadRoleMethodsRejectsIncompleteOrExtraDefinitions(t *testing.T) {
	valid := []byte("---\nname: eval-fixture\ndescription: Evaluate the fixture.\n---\n\n# Fixture\n")
	for name, files := range map[string]fstest.MapFS{
		"missing directory": {},
		"mismatched frontmatter": {
			"roles/builder/skills/eval-fixture/SKILL.md": {
				Data: []byte("---\nname: eval-other\ndescription: Wrong.\n---\n\n# Wrong\n"),
			},
		},
		"extra file": {
			"roles/builder/skills/eval-fixture/SKILL.md":  {Data: valid},
			"roles/builder/skills/eval-fixture/README.md": {Data: []byte("extra")},
		},
		"undeclared skill": {
			"roles/builder/skills/eval-fixture/SKILL.md": {Data: valid},
			"roles/builder/skills/eval-other/SKILL.md":   {Data: valid},
		},
	} {
		t.Run(name, func(t *testing.T) {
			p := &Person{
				Name:      "fixture",
				RoleOrder: []string{"builder"},
				Roles: map[string]Role{
					"builder": {Skill: "role-builder", Methods: []string{"eval-fixture"}},
				},
			}
			if err := loadRoleMethods(files, p); err == nil {
				t.Fatal("invalid role method layout must fail")
			}
		})
	}
}

func TestPersonSourceBindsAIOnlyRoleMethods(t *testing.T) {
	p, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	src, err := Source(p)
	if err != nil {
		t.Fatal(err)
	}
	// Methods belong to one role and boundaries are shared, so both counts derive
	// from the loaded model rather than a second copy of the roster policy.
	for _, roleName := range p.RoleOrder {
		role := p.Roles[roleName]
		want := 1 + len(role.Methods) + len(p.RoleActiveBoundaries(roleName))
		if got := len(src.RoleSkills[roleName]); got != want {
			t.Errorf("role %q selected %d person skills, want %d", roleName, got, want)
		}
	}
	for roleName, role := range p.Roles {
		if roleName != "ai" && len(role.Methods) != 0 {
			t.Errorf("role %q owns methods, which only the AI role may declare", roleName)
		}
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
		{cue: " encouraging ", want: []string{"nurturing"}},
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
		role.Seats[0].Tier != schema.ModelTierCommodity ||
		role.Seats[0].Pronouns != "they" {
		t.Fatalf("generalized seat was not preserved: %+v", role.Seats)
	}
	if strings.Join(role.SupportedModelTiers, ",") != "frontier,commodity" {
		t.Fatalf("role model tiers = %q", role.SupportedModelTiers)
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
	if src.ID != "roster:core" {
		t.Fatalf("source id = %q, want roster:core", src.ID)
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
		"unknown role inspiration": {
			body: strings.Replace(valid, `inspiration "fixture-builder" {`, `inspiration "missing-builder" {`, 1),
			want: `role "builder": inspiration "missing-builder" has no catalog entry`,
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

func TestParseRejectsRemovedRoleModelClass(t *testing.T) {
	valid := inspirationFixture()
	body := strings.Replace(valid, `personality "bright" "steady"`,
		"personality \"bright\" \"steady\"\n        model-class \"frontier\"", 1)
	if _, err := parse([]byte(body)); err == nil ||
		!strings.Contains(err.Error(), `unknown node "model-class"`) {
		t.Fatalf("removed model-class parse error = %v", err)
	}
}

func TestParseRejectsInvalidRoleModelTiers(t *testing.T) {
	valid := inspirationFixture()
	cases := map[string]struct {
		body string
		want string
	}{
		"empty": {
			body: strings.Replace(valid, `personality "bright" "steady"`,
				"personality \"bright\" \"steady\"\n        model-tier", 1),
			want: "model-tier needs at least one argument",
		},
		"unsupported": {
			body: strings.Replace(valid, `personality "bright" "steady"`,
				"personality \"bright\" \"steady\"\n        model-tier \"premium\"", 1),
			want: `unsupported model tier "premium"`,
		},
		"repeated": {
			body: strings.Replace(valid, `personality "bright" "steady"`,
				"personality \"bright\" \"steady\"\n        model-tier \"frontier\" \"frontier\"", 1),
			want: `repeats model tier "frontier"`,
		},
		"seat outside role tiers": {
			body: strings.Replace(valid, `personality "bright" "steady"`,
				"personality \"bright\" \"steady\"\n        model-tier \"frontier\"\n        seat \"local\" name=\"local builder\" pronouns=\"they\" tier=\"oss\"", 1),
			want: `seat "local" uses model tier "oss" outside the role compatibility set`,
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			if _, err := parse([]byte(tc.body)); err == nil ||
				!strings.Contains(err.Error(), tc.want) {
				t.Fatalf("model-tier parse error = %v, want %q", err, tc.want)
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

func TestParseRoleIdentityAppliesAcrossSeatsAndRejectsMixedDeclarations(t *testing.T) {
	legacy := strings.Replace(
		inspirationFixture(),
		`        personality "bright" "steady"`,
		`        agent "claude" name="legacy guide"
        personality "bright" "steady"`,
		1,
	)
	legacyPerson, err := parse([]byte(legacy))
	if err != nil {
		t.Fatalf("legacy per-seat identity: %v", err)
	}
	if got := legacyPerson.Roles["builder"].Seats[0]; got.Name != "legacy guide" || got.Pronouns != "" {
		t.Fatalf("legacy seat identity = %+v", got)
	}

	valid := strings.Replace(
		inspirationFixture(),
		`        personality "bright" "steady"`,
		`        identity name="fixture guide" pronouns="she"
        agent "claude"
        seat "local" channel="chatbot" tier="commodity"
        personality "bright" "steady"`,
		1,
	)
	p, err := parse([]byte(valid))
	if err != nil {
		t.Fatal(err)
	}
	role := p.Roles["builder"]
	if role.Identity == nil || role.Identity.Name != "fixture guide" || role.Identity.Pronouns == "" {
		t.Fatalf("role identity = %+v", role.Identity)
	}
	for _, seat := range role.Seats {
		if seat.Name != role.Identity.Name || seat.Pronouns != role.Identity.Pronouns {
			t.Fatalf("seat %q did not inherit role identity: %+v", seat.Selector(), seat)
		}
	}

	mixed := strings.Replace(valid, `agent "claude"`, `agent "claude" name="other guide"`, 1)
	if _, err := parse([]byte(mixed)); err == nil ||
		!strings.Contains(err.Error(), `seat "claude" cannot redefine role identity`) {
		t.Fatalf("mixed role and seat identity error = %v", err)
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

func TestRosterProseFloorsRejectAThinnedEntry(t *testing.T) {
	p, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	source := p.source
	role := p.Roles["engineer"]
	role.Briefing = "Too short."
	p.Roles["engineer"] = role
	if err := validateRosterProseFloors(source, p); err == nil ||
		!strings.Contains(err.Error(), "minimum is") {
		t.Fatalf("thinned role body error = %v", err)
	}
}

func TestRosterProseCeilingsAndFloorsBoundEveryShippedEntry(t *testing.T) {
	p, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	// The shipped roster must sit inside both bounds, which is what makes the
	// floors a live gate rather than a constant nobody reaches.
	for _, roleName := range p.RoleOrder {
		words := roleSkillBodyWordCount(p.Roles[roleName].Briefing)
		if words < minRoleSkillBodyWords || words > maxRoleSkillBodyWords {
			t.Errorf("role %q body has %d words, bounds are %d..%d",
				roleName, words, minRoleSkillBodyWords, maxRoleSkillBodyWords)
		}
	}
}
