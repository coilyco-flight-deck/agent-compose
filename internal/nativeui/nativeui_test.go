package nativeui

import (
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"forgejo.coilysiren.me/coilyco-flight-deck/agent-compose/internal/person"
)

var hexColor = regexp.MustCompile(`^#[0-9a-f]{6}$`)

// The dark base token set shipped in Claude Code 2.1.221. An upstream rename
// fails here rather than silently blanking a role.
var knownTokens = map[string]bool{
	"autoAccept": true, "autoAcceptShimmer": true, "bashBorder": true,
	"briefLabelClaude": true, "claude": true, "claudeShimmer": true,
	"clawd_body": true, "permission": true, "permissionShimmer": true,
	"promptBorder": true, "promptBorderShimmer": true, "remember": true,
	"skill": true, "suggestion": true,
	"red_FOR_SUBAGENTS_ONLY": true, "blue_FOR_SUBAGENTS_ONLY": true,
	"green_FOR_SUBAGENTS_ONLY": true, "yellow_FOR_SUBAGENTS_ONLY": true,
	"purple_FOR_SUBAGENTS_ONLY": true, "orange_FOR_SUBAGENTS_ONLY": true,
	"pink_FOR_SUBAGENTS_ONLY": true, "cyan_FOR_SUBAGENTS_ONLY": true,
}

func selected(t *testing.T) *person.Person {
	t.Helper()
	p, err := person.Load()
	if err != nil {
		t.Fatalf("load person: %v", err)
	}
	return p
}

func TestBuildCoversEveryRole(t *testing.T) {
	p := selected(t)
	bundles, err := Build(p, Options{})
	if err != nil {
		t.Fatalf("build: %v", err)
	}
	if len(bundles) != len(p.RoleOrder) {
		t.Fatalf("built %d bundles for %d roles", len(bundles), len(p.RoleOrder))
	}
	for i, bundle := range bundles {
		if bundle.Role != p.RoleOrder[i] {
			t.Errorf("bundle %d is role %q, want %q", i, bundle.Role, p.RoleOrder[i])
		}
	}
}

func TestThemeOverridesAreAcceptedByTheHarness(t *testing.T) {
	bundles, err := Build(selected(t), Options{})
	if err != nil {
		t.Fatalf("build: %v", err)
	}
	for _, bundle := range bundles {
		if bundle.Theme.Base != "dark" {
			t.Errorf("role %q uses base %q", bundle.Role, bundle.Theme.Base)
		}
		for token, value := range bundle.Theme.Overrides {
			if !knownTokens[token] {
				t.Errorf("role %q emits unknown token %q, which the harness drops silently", bundle.Role, token)
			}
			if !hexColor.MatchString(value) {
				t.Errorf("role %q token %q is %q, not #rrggbb", bundle.Role, token, value)
			}
		}
	}
}

// A role that recolored text or diffs would trade legibility for identity.
func TestThemeLeavesReadabilityTokensAlone(t *testing.T) {
	reserved := []string{"text", "inverseText", "background", "diffAdded", "diffRemoved", "error", "warning", "success"}
	bundles, err := Build(selected(t), Options{})
	if err != nil {
		t.Fatalf("build: %v", err)
	}
	for _, bundle := range bundles {
		for _, token := range reserved {
			if _, ok := bundle.Theme.Overrides[token]; ok {
				t.Errorf("role %q overrides reserved token %q", bundle.Role, token)
			}
		}
	}
}

func TestEachRoleClaimsExactlyOneSubagentSlot(t *testing.T) {
	bundles, err := Build(selected(t), Options{})
	if err != nil {
		t.Fatalf("build: %v", err)
	}
	for _, bundle := range bundles {
		claimed := 0
		for token := range bundle.Theme.Overrides {
			if strings.HasSuffix(token, "_FOR_SUBAGENTS_ONLY") {
				claimed++
			}
		}
		if claimed != 1 {
			t.Errorf("role %q claims %d subagent slots, want 1", bundle.Role, claimed)
		}
	}
}

func TestSpinnerVerbsCoverTheWholeMeld(t *testing.T) {
	p := selected(t)
	bundles, err := Build(p, Options{})
	if err != nil {
		t.Fatalf("build: %v", err)
	}
	for _, bundle := range bundles {
		role := p.Roles[bundle.Role]
		for _, name := range role.Personalities {
			for _, verb := range p.Personalities[name].Verbs {
				if !contains(bundle.Settings.SpinnerVerbs.Verbs, verb) {
					t.Errorf("role %q drops verb %q from personality %q", bundle.Role, verb, name)
				}
			}
		}
		if bundle.Settings.SpinnerVerbs.Mode != "replace" {
			t.Errorf("role %q defaults to mode %q", bundle.Role, bundle.Settings.SpinnerVerbs.Mode)
		}
	}
}

func TestEveryPersonalityCarriesVerbs(t *testing.T) {
	p := selected(t)
	for _, name := range p.PersonalityOrder {
		if len(p.Personalities[name].Verbs) == 0 {
			t.Errorf("personality %q has no spinner verbs", name)
		}
	}
}

func TestSpinnerModeIsSelectable(t *testing.T) {
	bundle, err := BuildRole(selected(t), "design", Options{SpinnerMode: "append"})
	if err != nil {
		t.Fatalf("build role: %v", err)
	}
	if bundle.Settings.SpinnerVerbs.Mode != "append" {
		t.Errorf("mode is %q, want append", bundle.Settings.SpinnerVerbs.Mode)
	}
}

func TestBuildRoleRejectsUnknownRole(t *testing.T) {
	if _, err := BuildRole(selected(t), "nonexistent", Options{}); err == nil {
		t.Fatal("expected an error for an unknown role")
	}
}

func TestWriteAllProducesLoadableFiles(t *testing.T) {
	bundles, err := Build(selected(t), Options{})
	if err != nil {
		t.Fatalf("build: %v", err)
	}
	dir := t.TempDir()
	written, err := WriteAll(dir, bundles)
	if err != nil {
		t.Fatalf("write: %v", err)
	}
	if written != len(bundles)*2 {
		t.Fatalf("wrote %d files for %d bundles", written, len(bundles))
	}
	for _, bundle := range bundles {
		var theme Theme
		readJSON(t, filepath.Join(dir, "themes", bundle.Slug+".json"), &theme)
		if theme.Name != bundle.Theme.Name {
			t.Errorf("role %q theme name round-tripped as %q", bundle.Role, theme.Name)
		}
		var settings Settings
		readJSON(t, filepath.Join(dir, "settings."+bundle.Role+".json"), &settings)
		if settings.Theme != "custom:"+bundle.Slug {
			t.Errorf("role %q settings select %q", bundle.Role, settings.Theme)
		}
	}
}

func readJSON(t *testing.T, path string, into any) {
	t.Helper()
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	if err := json.Unmarshal(raw, into); err != nil {
		t.Fatalf("parse %s: %v", path, err)
	}
}

func contains(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}

// A verb the harness already ships does no identity work under replace mode.
// testdata holds the 184 defaults from Claude Code 2.1.221.
func TestVerbsDoNotRepeatTheHarnessDefaults(t *testing.T) {
	raw, err := os.ReadFile(filepath.Join("testdata", "harness-default-verbs.txt"))
	if err != nil {
		t.Fatalf("read defaults: %v", err)
	}
	defaults := map[string]bool{}
	for _, line := range strings.Split(string(raw), "\n") {
		if verb := strings.TrimSpace(line); verb != "" {
			defaults[verb] = true
		}
	}

	p := selected(t)
	for _, name := range p.PersonalityOrder {
		for _, verb := range p.Personalities[name].Verbs {
			if defaults[verb] {
				t.Errorf("personality %q reuses harness default verb %q", name, verb)
			}
		}
	}
}

func TestTipsCarryCharterAndKeepHarnessDefaults(t *testing.T) {
	p := selected(t)
	bundles, err := Build(p, Options{})
	if err != nil {
		t.Fatalf("build: %v", err)
	}
	for _, bundle := range bundles {
		tips := bundle.Settings.SpinnerTips
		if tips.ExcludeDefault {
			t.Errorf("role %q excludes the harness tips", bundle.Role)
		}
		if !bundle.Settings.SpinnerTipsEnabled {
			t.Errorf("role %q emits tips but leaves them disabled", bundle.Role)
		}
		if len(tips.Tips) != 3 {
			t.Errorf("role %q has %d tips, want purpose, charter, and meld", bundle.Role, len(tips.Tips))
		}
		role := p.Roles[bundle.Role]
		if tips.Tips[0] != role.Purpose {
			t.Errorf("role %q leads with %q, not its purpose", bundle.Role, tips.Tips[0])
		}
		if !strings.Contains(tips.Tips[1], role.Identity.Name) {
			t.Errorf("role %q charter tip omits the seat name", bundle.Role)
		}
		for _, personality := range role.Personalities {
			if !strings.Contains(tips.Tips[2], personality) {
				t.Errorf("role %q meld tip omits %q", bundle.Role, personality)
			}
		}
	}
}
