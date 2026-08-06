package main

import (
	"bytes"
	"errors"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"forgejo.coilysiren.me/coilyco-flight-deck/agent-compose/internal/bundle"
	"forgejo.coilysiren.me/coilyco-flight-deck/agent-compose/internal/cascade"
	"forgejo.coilysiren.me/coilyco-flight-deck/agent-compose/internal/compose"
	"forgejo.coilysiren.me/coilyco-flight-deck/agent-compose/internal/describe"
	"forgejo.coilysiren.me/coilyco-flight-deck/agent-compose/internal/evaluation"
	"forgejo.coilysiren.me/coilyco-flight-deck/agent-compose/internal/nativelaunch"
	"forgejo.coilysiren.me/coilyco-flight-deck/agent-compose/internal/overlay"
	"forgejo.coilysiren.me/coilyco-flight-deck/agent-compose/internal/palette"
	"forgejo.coilysiren.me/coilyco-flight-deck/agent-compose/internal/person"
	"forgejo.coilysiren.me/coilyco-flight-deck/agent-compose/internal/personpolicy"
	"forgejo.coilysiren.me/coilyco-flight-deck/agent-compose/internal/resolver"
	"forgejo.coilysiren.me/coilyco-flight-deck/agent-compose/internal/roster"
	"forgejo.coilysiren.me/coilyco-flight-deck/agent-compose/internal/schema"
)

func TestEvaluationOutputUsesYAML(t *testing.T) {
	t.Parallel()
	pack, err := evaluation.Build("engineer", "codex")
	if err != nil {
		t.Fatal(err)
	}
	raw, err := evaluationOutput(pack, "yaml")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(string(raw), "format: agent-compose.evaluation-pack.v2\n") {
		t.Fatalf("evaluation output is not YAML:\n%s", raw)
	}
	if _, err := evaluationOutput(pack, "json"); err == nil {
		t.Fatal("legacy JSON evaluation output remains accepted")
	}
}

func TestConfigValidateRejectsRemovedRoleProviderKeys(t *testing.T) {
	valid := filepath.Join(t.TempDir(), "agent-compose.yaml")
	if err := os.WriteFile(valid, []byte(
		"person_policy: external-only\nperson_source: ./person\n",
	), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := cascade.LoadConfig(valid); err != nil {
		t.Fatalf("valid strict config: %v", err)
	}

	for name, body := range map[string]string{
		"inline": "role_providers: {}\n",
		"file":   "role_providers_file: role-providers.yaml\n",
	} {
		t.Run(name, func(t *testing.T) {
			invalid := filepath.Join(t.TempDir(), "agent-compose.yaml")
			if err := os.WriteFile(invalid, []byte(body), 0o600); err != nil {
				t.Fatal(err)
			}
			if _, err := cascade.LoadConfig(invalid); err == nil || !strings.Contains(err.Error(), "field role_providers") {
				t.Fatalf("removed role-provider config passed: %v", err)
			}
		})
	}
}

func TestWriteEvaluationPacksEmitsCompleteDigestIndex(t *testing.T) {
	t.Parallel()
	packs, err := evaluation.BuildCorePacks("codex")
	if err != nil {
		t.Fatal(err)
	}
	output := t.TempDir()
	if err := writeEvaluationPacks(packs, "yaml", output); err != nil {
		t.Fatal(err)
	}
	entries, err := os.ReadDir(output)
	if err != nil {
		t.Fatal(err)
	}
	profile, err := person.Load()
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != len(profile.RoleOrder)+1 {
		t.Fatalf("evaluation output entries = %d, want loader role count plus index %d", len(entries), len(profile.RoleOrder)+1)
	}
	index, err := os.ReadFile(filepath.Join(output, "index.json"))
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{
		`"format": "agent-compose.evaluation-index.v1"`,
		`"role": "engineer"`,
		`"role": "content"`,
		`"pack_digest": "sha256:`,
	} {
		if !strings.Contains(string(index), want) {
			t.Errorf("evaluation index omitted %q:\n%s", want, index)
		}
	}
	if err := writeEvaluationPacks(packs, "yaml", output); err == nil ||
		!strings.Contains(err.Error(), "must be empty") {
		t.Fatalf("non-empty output directory error = %v", err)
	}
}

func TestWriteScorecardSupportsOutputAndFreshnessCheck(t *testing.T) {
	t.Parallel()
	raw := []byte("# scorecard\n")
	var stdout bytes.Buffer
	if err := writeScorecard(&stdout, raw, "", false); err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(stdout.Bytes(), raw) {
		t.Fatalf("scorecard stdout = %q", stdout.Bytes())
	}

	output := filepath.Join(t.TempDir(), "SCORECARD.md")
	if err := writeScorecard(&stdout, raw, output, false); err != nil {
		t.Fatal(err)
	}
	if err := writeScorecard(&stdout, raw, output, true); err != nil {
		t.Fatalf("fresh scorecard failed check: %v", err)
	}
	if err := os.WriteFile(output, []byte("stale\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := writeScorecard(&stdout, raw, output, true); err == nil ||
		!strings.Contains(err.Error(), "stale") {
		t.Fatalf("stale scorecard error = %v", err)
	}
	if err := writeScorecard(&stdout, raw, "", true); err == nil ||
		!strings.Contains(err.Error(), "requires --out") {
		t.Fatalf("missing output check error = %v", err)
	}
}

func TestExternalPersonProfileExampleExercisesEveryPersonSurface(t *testing.T) {
	profile := filepath.Join("..", "..", "examples", "person-profile")
	library := filepath.Join("..", "..", "examples", "shared-personality-library")
	request := filepath.Join(profile, "request.kdl")
	p, err := person.LoadDirectoryWithLibraries(profile, library)
	if err != nil {
		t.Fatal(err)
	}
	result, err := compose.Run(request, t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	if _, err := bundle.Verify(result.Bundle.Dir); err != nil {
		t.Fatal(err)
	}
	archive := filepath.Join(t.TempDir(), "example.tar.gz")
	if err := bundle.Export(result.Bundle.Dir, archive); err != nil {
		t.Fatal(err)
	}
	if info, err := os.Stat(archive); err != nil || info.Size() == 0 {
		t.Fatalf("example export is missing: info=%v err=%v", info, err)
	}
	if _, err := describe.Bundle(result.Bundle.Dir, describe.Options{}); err != nil {
		t.Fatal(err)
	}
	pack, err := evaluation.BuildFor(p, "bulk-captioner", "chatbot-sonnet-low")
	if err != nil || len(pack.Cases) != 1 {
		t.Fatalf("example evaluation failed: cases=%v err=%v", pack, err)
	}
	projected, err := overlay.Build(p, "bulk-captioner", "chatbot-sonnet-low", "available")
	if err != nil || projected.Seat.Pronouns != "they" {
		t.Fatalf("example overlay failed: doc=%v err=%v", projected, err)
	}
	if _, err := palette.Build(p); err != nil {
		t.Fatal(err)
	}
	files, err := roster.Render(p, nil, t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	for _, path := range []string{
		"person.json",
		"person.v4.json",
		"personality-index.md",
		".agents/skills/role-bulk-captioner/SKILL.md",
	} {
		if len(files[path]) == 0 {
			t.Errorf("example roster omitted %q", path)
		}
	}
	if entries, err := p.PersonalityCatalog(nil); err != nil || len(entries) != 3 {
		t.Fatalf("example personality catalogue failed: entries=%v err=%v", entries, err)
	}
	if entries, err := p.RoleCatalog(); err != nil || len(entries) != 2 {
		t.Fatalf("example role catalogue failed: entries=%v err=%v", entries, err)
	}
	if entries, err := p.SeatCatalog("bulk-captioner"); err != nil || len(entries) != 1 {
		t.Fatalf("example seat catalogue failed: entries=%v err=%v", entries, err)
	}
}

func TestDirectPersonSelectionInheritsExternalOnlyHost(t *testing.T) {
	dir := t.TempDir()
	personRoot, err := filepath.Abs(filepath.Join(
		"..", "..", "testdata", "contracts", "person-independent",
	))
	if err != nil {
		t.Fatal(err)
	}
	paths := cascade.Paths{
		Config: filepath.Join(dir, "agent-compose.yaml"),
		Home:   filepath.Join(dir, "home"),
	}
	config := "person_policy: " + personpolicy.ExternalOnly + "\n" +
		"person_source: " + personRoot + "\n"
	if err := os.WriteFile(paths.Config, []byte(config), 0o644); err != nil {
		t.Fatal(err)
	}
	p, external, err := loadSelectedPersonAt("", paths)
	if err != nil {
		t.Fatal(err)
	}
	if !external || p.Name != "workbench" {
		t.Fatalf("direct selection = external:%t person:%q", external, p.Name)
	}
}

func TestDirectPersonSelectionFailsClosedWithBrokenHostGuard(t *testing.T) {
	dir := t.TempDir()
	paths := cascade.Paths{
		Config: filepath.Join(dir, "agent-compose.yaml"),
		Home:   filepath.Join(dir, "home"),
	}
	if err := os.WriteFile(
		paths.Config,
		[]byte("person_policy: "+personpolicy.ExternalOnly+"\n"),
		0o644,
	); err != nil {
		t.Fatal(err)
	}
	if _, _, err := loadSelectedPersonAt("", paths); err == nil ||
		!strings.Contains(err.Error(), "requires person_source") {
		t.Fatalf("broken host guard error = %v", err)
	}
}

func TestProjectPersonPolicyRejectsEmbeddedBundle(t *testing.T) {
	result, err := compose.Run(
		filepath.Join("..", "..", "testdata", "contracts", "native.kdl"),
		t.TempDir(),
	)
	if err != nil {
		t.Fatal(err)
	}
	err = validateProjectPersonPolicy(result.Bundle.Dir, compose.Options{
		PersonPolicy: personpolicy.ExternalOnly,
		PersonSource: "/person",
	})
	if err == nil || !strings.Contains(err.Error(), "roster:core") {
		t.Fatalf("embedded bundle policy error = %v", err)
	}
}

func TestProjectPersonPolicyAcceptsExternalBundle(t *testing.T) {
	result, err := compose.Run(
		filepath.Join("..", "..", "testdata", "contracts", "custom-person.kdl"),
		t.TempDir(),
	)
	if err != nil {
		t.Fatal(err)
	}
	if err := validateProjectPersonPolicy(result.Bundle.Dir, compose.Options{
		PersonPolicy: personpolicy.ExternalOnly,
		PersonSource: "/person",
	}); err != nil {
		t.Fatalf("external bundle rejected: %v", err)
	}
}

func TestDispatchArgs(t *testing.T) {
	cases := map[string]struct{ in, want []string }{
		"canonical name untouched": {
			[]string{"/usr/local/bin/agent-compose", "describe", "x"},
			[]string{"/usr/local/bin/agent-compose", "describe", "x"},
		},
		"acompose injects compose": {
			[]string{"/opt/homebrew/bin/acompose"},
			[]string{"/opt/homebrew/bin/acompose", "compose"},
		},
		"acompose keeps trailing args": {
			[]string{"acompose", "--", "claude"},
			[]string{"acompose", "compose", "--", "claude"},
		},
		"acompose role and harness inject launch": {
			[]string{"acompose", "design", "codex", "--model", "gpt"},
			[]string{"acompose", "launch", "design", "codex", "--model", "gpt"},
		},
		"acompose request and layout remain compose": {
			[]string{"acompose", "--layout", "codex", "request.kdl", "--", "codex"},
			[]string{"acompose", "compose", "--layout", "codex", "request.kdl", "--", "codex"},
		},
		"acompose exposes statusline directly": {
			[]string{"acompose", "statusline", "--target", "/workspace"},
			[]string{"acompose", "statusline", "--target", "/workspace"},
		},
		"windows exe suffix": {
			[]string{`C:\shims\acompose.exe`, "req.kdl"},
			[]string{`C:\shims\acompose.exe`, "compose", "req.kdl"},
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			if got := dispatchArgs(tc.in); !reflect.DeepEqual(got, tc.want) {
				t.Fatalf("got %v want %v", got, tc.want)
			}
		})
	}
}

func TestPrintSummaryUsesSlashSeparators(t *testing.T) {
	t.Parallel()
	p, err := person.Load()
	if err != nil {
		t.Fatal(err)
	}
	result := summaryFixture(t, p)

	var output strings.Builder
	if err := printSummary(&output, result, person.RoleTranscriptOptions{}); err != nil {
		t.Fatal(err)
	}
	got := output.String()
	for _, want := range []string{
		"request: model tier frontier // delivery native-skills",
		"roster: core // provided by: roster:core",
		"role: engineer",
		"personalities: curious // grounded // meticulous",
		"melded color: #90a66a",
		"personality: curious",
		"inspiration achievement:",
		"renderer expressions: available // listening // thinking",
		"decisions: 1 selected // 1 excluded // 1 shadowed // 1 delivered",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("summary missing %q:\n%s", want, got)
		}
	}
	if strings.Contains(got, "·") {
		t.Fatalf("summary retained middle-dot separator:\n%s", got)
	}
	for _, line := range strings.Split(strings.TrimSuffix(got, "\n"), "\n") {
		if len(line) > 0 && line[0] == ' ' {
			t.Fatalf("summary line is not flush-left: %q", line)
		}
	}

	var colored strings.Builder
	if err := printSummary(&colored, result, person.RoleTranscriptOptions{
		Color: true, TrueColor: true,
	}); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(colored.String(), "\x1b[38;2;144;166;106mbundle") {
		t.Fatalf("summary intro did not use melded truecolor:\n%q", colored.String())
	}
}

func TestPrintCompositionWarningsUsesExplicitWarningPrefix(t *testing.T) {
	var output strings.Builder
	printCompositionWarnings(&output, []string{
		`source "aos" provider role "content" matched composed skill "writing-kai-voice" through selectors "*writing*", "*voice*", selected once`,
	})
	want := `agent-compose: warning: source "aos" provider role "content" matched composed skill "writing-kai-voice" through selectors "*writing*", "*voice*", selected once` + "\n"
	if output.String() != want {
		t.Fatalf("warning output = %q, want %q", output.String(), want)
	}
}

func summaryFixture(t *testing.T, p *person.Person) *compose.Result {
	t.Helper()
	return &compose.Result{
		Bundle: &bundle.Result{Key: "abc123", Dir: "/tmp/bundle", Reused: true},
		Resolution: &resolver.Resolution{
			Request: &schema.Request{
				Role:      "engineer",
				ModelTier: schema.ModelTierFrontier,
				Delivery:  "native-skills",
			},
			Personalities: []string{"curious", "grounded", "meticulous"},
			FavoriteColor: "#90a66a",
			SourceIDs:     []string{"roster:core", "aos-public"},
			Person:        p,
			Decisions: []resolver.Decision{
				{Outcome: resolver.OutcomeSelected},
				{Outcome: resolver.OutcomeExcluded},
				{Outcome: resolver.OutcomeShadowed},
				{Outcome: resolver.OutcomeDelivered},
			},
		},
	}
}

func TestNativeHarnessCommandPromptsFreshCodexSession(t *testing.T) {
	t.Parallel()
	for name, args := range map[string][]string{
		"bare": nil,
		"AOS trust override": {
			"--config",
			`projects={"/tmp/aos-native/session/projects"={trust_level="trusted"}}`,
		},
		"model selection": {"--model", "gpt-5.6"},
	} {
		t.Run(name, func(t *testing.T) {
			got := nativeHarnessCommand("codex", args)
			if got[len(got)-1] != nativeCodexIntroductionPrompt {
				t.Fatalf("command does not end with the introduction prompt: %#v", got)
			}
		})
	}
}

func TestNativeHarnessCommandPreservesExplicitCodexWork(t *testing.T) {
	t.Parallel()
	for name, args := range map[string][]string{
		"prompt":         {"help me debug this"},
		"subcommand":     {"exec", "run the tests"},
		"unknown option": {"--future-option"},
	} {
		t.Run(name, func(t *testing.T) {
			want := append([]string{"codex"}, args...)
			if got := nativeHarnessCommand("codex", args); !reflect.DeepEqual(got, want) {
				t.Fatalf("command = %#v, want unchanged %#v", got, want)
			}
		})
	}
}

func TestNativeHarnessCommandDoesNotPromptOtherHarnesses(t *testing.T) {
	t.Parallel()
	if got, want := nativeHarnessCommand("claude", nil), []string{"claude"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("command = %#v, want %#v", got, want)
	}
}

func TestNativeLaunchSummaryPutsRoleTranscriptLast(t *testing.T) {
	t.Parallel()
	p, err := person.Load()
	if err != nil {
		t.Fatal(err)
	}
	result := summaryFixture(t, p)
	var output strings.Builder
	if err := printNativeLaunchSummary(
		&output,
		result,
		person.RoleTranscriptOptions{},
		true,
	); err != nil {
		t.Fatal(err)
	}
	got := output.String()
	audit := strings.Index(got, "sources:")
	role := strings.Index(got, "role metadata")
	personality := strings.Index(got, "personality metadata")
	if audit < 0 || personality < audit || role < personality {
		t.Fatalf("native launch summary order is wrong:\n%s", got)
	}
	if strings.Contains(got[role:], "sources:") {
		t.Fatalf("routine audit followed the role metadata:\n%s", got)
	}
}

func TestAcknowledgeNativeLaunchWaitsForEnter(t *testing.T) {
	t.Parallel()
	input := &enterGateReader{
		started: make(chan struct{}),
		release: make(chan struct{}),
	}
	var output strings.Builder
	done := make(chan error, 1)
	go func() {
		done <- acknowledgeNativeLaunch(input, &output, true)
	}()
	<-input.started
	if got, want := output.String(), "Press Enter to continue"; got != want {
		t.Fatalf("visible acknowledgement = %q, want %q", got, want)
	}
	select {
	case err := <-done:
		t.Fatalf("acknowledgement returned before Enter: %v", err)
	default:
	}
	close(input.release)
	if err := <-done; err != nil {
		t.Fatal(err)
	}
	if got, want := output.String(), "Press Enter to continue\n"; got != want {
		t.Fatalf("acknowledgement output = %q, want %q", got, want)
	}
}

func TestAcknowledgeNativeLaunchSkipsNonInteractiveInput(t *testing.T) {
	t.Parallel()
	input := &failReader{err: errors.New("stdin must not be read")}
	var output strings.Builder
	if err := acknowledgeNativeLaunch(input, &output, false); err != nil {
		t.Fatal(err)
	}
	if input.reads != 0 || output.Len() != 0 {
		t.Fatalf("non-interactive acknowledgement touched I/O: reads=%d output=%q",
			input.reads, output.String())
	}
}

func TestAcknowledgeNativeLaunchStopsBeforeExecOnInputFailure(t *testing.T) {
	t.Parallel()
	want := errors.New("interrupted")
	err := acknowledgeNativeLaunch(&failReader{err: want}, io.Discard, true)
	if !errors.Is(err, want) {
		t.Fatalf("acknowledgement error = %v, want wrapped interruption", err)
	}
}

func TestNativeLaunchInteractiveRejectsPipes(t *testing.T) {
	t.Parallel()
	input, inputWriter, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	defer input.Close()
	defer inputWriter.Close()
	output, outputWriter, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	defer output.Close()
	defer outputWriter.Close()
	if nativeLaunchInteractive(input, outputWriter) {
		t.Fatal("piped input and output were treated as an interactive terminal")
	}
}

type failReader struct {
	err   error
	reads int
}

func (r *failReader) Read(_ []byte) (int, error) {
	r.reads++
	return 0, r.err
}

type enterGateReader struct {
	started chan struct{}
	release chan struct{}
}

func (r *enterGateReader) Read(p []byte) (int, error) {
	close(r.started)
	<-r.release
	p[0] = '\n'
	return 1, nil
}

func TestPrintVerificationUsesBoundedCounts(t *testing.T) {
	t.Parallel()
	verification := &bundle.Verification{
		Files: 128,
		Identities: []bundle.Identity{
			{Source: "roster:core", Skill: "personality-curious"},
			{Source: "aos-public", Skill: "coding-go"},
		},
	}

	var output strings.Builder
	printVerification(&output, verification)
	if got, want := output.String(), "bundle verified: 2 skills // 128 files\n"; got != want {
		t.Fatalf("verification output = %q, want %q", got, want)
	}
}

func TestActivateNativeRuntimeHome(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", "old-home")
	t.Setenv("USERPROFILE", "old-profile")
	t.Setenv("CODEX_HOME", "old-codex")
	t.Setenv("XDG_CONFIG_HOME", "old-config")
	t.Setenv("CLAUDE_CONFIG_DIR", "old-claude")

	if err := activateNativeRuntimeHome(home, "claude"); err != nil {
		t.Fatal(err)
	}
	for name, want := range map[string]string{
		"HOME":              home,
		"USERPROFILE":       home,
		"CODEX_HOME":        filepath.Join(home, ".codex"),
		"XDG_CONFIG_HOME":   filepath.Join(home, ".config"),
		"CLAUDE_CONFIG_DIR": filepath.Join(home, ".claude"),
	} {
		if got := os.Getenv(name); got != want {
			t.Errorf("%s = %q, want %q", name, got, want)
		}
	}
}

func TestClearNativeLaunchEnvironmentRemovesSelectionFacts(t *testing.T) {
	names := []string{
		nativelaunch.EnvModelTier,
		"AGENT_COMPOSE_MODEL_CLASS",
		nativelaunch.EnvRuntimeHome,
	}
	for _, name := range names {
		t.Setenv(name, "fixture")
	}
	if err := clearNativeLaunchEnvironment(); err != nil {
		t.Fatal(err)
	}
	for _, name := range names {
		if _, ok := os.LookupEnv(name); ok {
			t.Errorf("%s remains in the harness environment", name)
		}
	}
}

func TestActivateNativeRuntimeHomePreservesCanonicalCodexState(t *testing.T) {
	hostCodexHome := filepath.Join(t.TempDir(), ".codex")
	if err := os.MkdirAll(hostCodexHome, 0o700); err != nil {
		t.Fatal(err)
	}
	hostCodexHome, err := filepath.EvalSymlinks(hostCodexHome)
	if err != nil {
		t.Fatal(err)
	}
	home := t.TempDir()
	if err := os.Symlink(hostCodexHome, filepath.Join(home, ".codex")); err != nil {
		t.Fatal(err)
	}

	if err := activateNativeRuntimeHome(home, "codex"); err != nil {
		t.Fatal(err)
	}

	if got := os.Getenv("CODEX_HOME"); got != hostCodexHome {
		t.Fatalf("CODEX_HOME = %q, want canonical host state %q", got, hostCodexHome)
	}
	if got := os.Getenv("HOME"); got != home {
		t.Fatalf("HOME = %q, want isolated runtime home %q", got, home)
	}
}
