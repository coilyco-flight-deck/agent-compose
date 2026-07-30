package cascade

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

type env struct {
	dir      string
	paths    Paths
	claude   string
	codex    string
	projects string
}

func newEnv(t *testing.T) env {
	t.Helper()
	dir := t.TempDir()
	return env{
		dir: dir,
		paths: Paths{
			Config:       filepath.Join(dir, "agent-compose.yaml"),
			Composed:     filepath.Join(dir, "out", "COMPOSED.md"),
			ProjectsRoot: filepath.Join(dir, "projects"),
		},
		claude:   filepath.Join(dir, "links", "CLAUDE.md"),
		codex:    filepath.Join(dir, "links", "AGENTS.md"),
		projects: filepath.Join(dir, "projects"),
	}
}

func (e env) write(t *testing.T, rel, content string) string {
	t.Helper()
	path := filepath.Join(e.dir, rel)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

func (e env) config(t *testing.T, body string) {
	t.Helper()
	base := "load_points:\n  claude: " + e.claude + "\n  codex: " + e.codex + "\n"
	if err := os.MkdirAll(filepath.Dir(e.paths.Config), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(e.paths.Config, []byte(body+base), 0o644); err != nil {
		t.Fatal(err)
	}
}

func (e env) run(t *testing.T, dryRun bool) (int, string, string) {
	t.Helper()
	return e.runOptions(t, RunOptions{DryRun: dryRun})
}

func (e env) runOptions(t *testing.T, opts RunOptions) (int, string, string) {
	t.Helper()
	var out, errOut bytes.Buffer
	code := Run(e.paths, opts, &out, &errOut)
	return code, out.String(), errOut.String()
}

func (e env) check(t *testing.T) (int, string, string) {
	t.Helper()
	var out, errOut bytes.Buffer
	code := Check(e.paths, &out, &errOut)
	return code, out.String(), errOut.String()
}

func TestNoConfigIsNoOp(t *testing.T) {
	e := newEnv(t)
	code, out, _ := e.run(t, false)
	if code != 0 || !strings.Contains(out, "nothing to do (opt-in)") {
		t.Fatalf("absent config must no-op: %d %q", code, out)
	}
}

func TestLoadConfigValidatesPersonPolicy(t *testing.T) {
	e := newEnv(t)
	valid := e.write(
		t,
		"valid.yaml",
		"person_policy: external-only\nperson_source: ./person\n",
	)
	if _, err := LoadConfig(valid); err != nil {
		t.Fatalf("valid person policy failed: %v", err)
	}
	for name, body := range map[string]string{
		"missing source": "person_policy: external-only\n",
		"unknown policy": "person_policy: prefer-external\nperson_source: ./person\n",
	} {
		t.Run(name, func(t *testing.T) {
			path := e.write(t, name+".yaml", body)
			if _, err := LoadConfig(path); err == nil {
				t.Fatal("invalid person policy passed")
			}
		})
	}
}

func TestLoadConfigRejectsRemovedV1IntegrationKeys(t *testing.T) {
	e := newEnv(t)
	for name, body := range map[string]string{
		"remote sources": "remote_skill_sources:\n  - owner/catalog/skills@v1\n",
		"remote ttl":     "remote_skill_cache_ttl: 45m\n",
		"mcp inventory":  "mcp_inventory: /tmp/mcporter.json\n",
	} {
		t.Run(name, func(t *testing.T) {
			if _, err := LoadConfig(e.write(t, name+".yaml", body)); err == nil {
				t.Fatal("removed v1 key passed strict config loading")
			}
		})
	}
}

func TestComposeBasicsAndSilentRecompose(t *testing.T) {
	e := newEnv(t)
	a := e.write(t, "src/a/AGENTS.COMPOSE.md", "# Alpha\n\nalpha doctrine\n")
	b := e.write(t, "src/b/AGENTS.COMPOSE.md", "# Beta\n\nbeta doctrine\n")
	e.config(t, "sources:\n  - "+a+"\n  - "+b+"\n")

	code, out, errOut := e.run(t, false)
	if code != 0 {
		t.Fatalf("run failed: %s %s", out, errOut)
	}
	body, err := os.ReadFile(e.paths.Composed)
	if err != nil {
		t.Fatal(err)
	}
	text := string(body)
	if !strings.HasPrefix(text, Banner+"\n") || !strings.HasSuffix(text, "\n") {
		t.Fatalf("banner or trailing newline missing:\n%s", text)
	}
	if !strings.Contains(text, "<!-- source: "+a+" -->") || strings.Index(text, "Alpha") > strings.Index(text, "Beta") {
		t.Fatalf("fences or order wrong:\n%s", text)
	}
	for _, link := range []string{e.claude, e.codex} {
		target, err := os.Readlink(link)
		if err != nil || target != e.paths.Composed {
			t.Fatalf("load point %s not linked: %v %s", link, err, target)
		}
	}

	code, out, _ = e.run(t, false)
	if code != 0 || strings.Contains(out, "wrote") || strings.Contains(out, "linked") {
		t.Fatalf("converged recompose must be silent, got %q", out)
	}
}

func TestReapplyRewritesCurrentLayout(t *testing.T) {
	e := newEnv(t)
	src := e.write(t, "src/AGENTS.COMPOSE.md", "# Doc\n")
	e.config(t, "sources:\n  - "+src+"\n")

	if code, _, errOut := e.run(t, false); code != 0 {
		t.Fatalf("initial run failed: %s", errOut)
	}
	code, out, errOut := e.runOptions(t, RunOptions{Reapply: true})
	if code != 0 {
		t.Fatalf("reapply failed: %s %s", out, errOut)
	}
	manifestPath := filepath.Join(filepath.Dir(e.paths.Composed), "mount-eligibility.json")
	for _, want := range []string{
		"wrote   " + e.paths.Composed,
		"wrote   " + manifestPath,
		"linked  " + e.claude,
		"linked  " + e.codex,
		"changed=4",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("reapply output missing %q:\n%s", want, out)
		}
	}
}

func TestVerbosePrintsEveryLayoutMapping(t *testing.T) {
	e := newEnv(t)
	a := e.write(t, "src/a/AGENTS.COMPOSE.md", "# Alpha\n")
	b := e.write(t, "src/b/AGENTS.COMPOSE.md", "# Beta\n")
	override := e.write(t, "src/b/AGENTS.claude.md", "# Beta\n\nClaude\n")
	e.config(t, "sources:\n  - "+a+"\n  - "+b+"\n")

	code, out, errOut := e.runOptions(t, RunOptions{Verbose: true})
	if code != 0 {
		t.Fatalf("verbose run failed: %s %s", out, errOut)
	}
	manifestPath := filepath.Join(filepath.Dir(e.paths.Composed), "mount-eligibility.json")
	claudeOut := harnessOutputPath(e.paths.Composed, "claude")
	codexOut := harnessOutputPath(e.paths.Composed, "codex")
	for _, want := range []string{
		"layout  " + a + " => " + claudeOut,
		"layout  " + b + " => " + claudeOut,
		"layout  " + override + " => " + claudeOut,
		"layout  " + a + " => " + codexOut,
		"layout  " + b + " => " + codexOut,
		"layout  " + e.paths.Config + " => " + manifestPath,
		"layout  " + claudeOut + " => " + e.claude,
		"layout  " + codexOut + " => " + e.codex,
	} {
		if !strings.Contains(out, want) {
			t.Errorf("verbose output missing %q:\n%s", want, out)
		}
	}
}

func TestScopeFiltering(t *testing.T) {
	e := newEnv(t)
	e.write(t, "src/work/AGENTS.COMPOSE.md", "---\nscopes: [work]\n---\n# Work\n")
	e.write(t, "src/play/AGENTS.COMPOSE.md", "---\nscopes: [play]\n---\n# Play\n")
	e.write(t, "src/untagged/AGENTS.COMPOSE.md", "# Untagged\n")
	e.config(t, "scopes: [work]\nroots:\n  - "+filepath.Join(e.dir, "src")+"\n")

	if code, out, errOut := e.run(t, false); code != 0 {
		t.Fatalf("run failed: %s %s", out, errOut)
	}
	text := readFile(t, e.paths.Composed)
	if !strings.Contains(text, "# Work") || strings.Contains(text, "# Play") || strings.Contains(text, "# Untagged") {
		t.Fatalf("scope filtering wrong:\n%s", text)
	}

	e.config(t, "roots:\n  - "+filepath.Join(e.dir, "src")+"\n")
	if code, _, _ := e.run(t, false); code != 0 {
		t.Fatal("unfiltered run failed")
	}
	text = readFile(t, e.paths.Composed)
	if !strings.Contains(text, "# Play") || !strings.Contains(text, "# Untagged") {
		t.Fatalf("absent scopes must disable filtering:\n%s", text)
	}

	e.config(t, "scopes: []\nroots:\n  - "+filepath.Join(e.dir, "src")+"\n")
	if code, _, errOut := e.run(t, false); code != 1 || !strings.Contains(errOut, "refusing to write an empty COMPOSED.md") {
		t.Fatalf("empty scope match must refuse: %d %s", code, errOut)
	}
}

func TestOverridesDivergeAndStaleCleanup(t *testing.T) {
	e := newEnv(t)
	e.write(t, "src/doc/AGENTS.COMPOSE.md", "# Doc\n\n## Shared\n\nshared body\n\n## Tail\n\ntail body\n")
	override := e.write(t, "src/doc/AGENTS.claude.md", "## Shared\n\nclaude body\n\n## Extra\n\nappended\n")
	e.config(t, "roots:\n  - "+filepath.Join(e.dir, "src")+"\n")

	if code, out, errOut := e.run(t, false); code != 0 {
		t.Fatalf("run failed: %s %s", out, errOut)
	}
	claudeOut := harnessOutputPath(e.paths.Composed, "claude")
	codexOut := harnessOutputPath(e.paths.Composed, "codex")
	claudeText := readFile(t, claudeOut)
	codexText := readFile(t, codexOut)
	if !strings.Contains(claudeText, "claude body") || strings.Contains(claudeText, "shared body") {
		t.Fatalf("override must replace the matched section:\n%s", claudeText)
	}
	if !strings.Contains(claudeText, "## Extra") || !strings.Contains(claudeText, "## Tail") {
		t.Fatalf("override must append new sections and keep others:\n%s", claudeText)
	}
	if !strings.Contains(codexText, "shared body") || strings.Contains(codexText, "claude body") {
		t.Fatalf("codex slice must stay unpatched:\n%s", codexText)
	}
	if target, _ := os.Readlink(e.claude); target != claudeOut {
		t.Fatalf("claude link must point at divergent output, got %s", target)
	}

	if err := os.Remove(override); err != nil {
		t.Fatal(err)
	}
	if code, out, _ := e.run(t, false); code != 0 || !strings.Contains(out, "removed "+claudeOut) {
		t.Fatalf("converging must remove stale divergent outputs: %s", out)
	}
	if _, err := os.Stat(codexOut); !os.IsNotExist(err) {
		t.Fatal("stale codex output must be gone")
	}
	if target, _ := os.Readlink(e.claude); target != e.paths.Composed {
		t.Fatal("links must repoint at the shared output")
	}
}

func TestRewrites(t *testing.T) {
	e := newEnv(t)
	src := e.write(t, "src/doc/AGENTS.COMPOSE.md",
		"# Doc\n\nsee [readme](README.md#usage) and [site](https://example.com) and [abs](/etc/hosts)\n\n## See also\n\n* [sibling](sibling.md)\n")
	e.config(t, "sources:\n  - "+src+"\n")
	if code, _, errOut := e.run(t, false); code != 0 {
		t.Fatal(errOut)
	}
	text := readFile(t, e.paths.Composed)
	resolvedDir, err := filepath.EvalSymlinks(filepath.Dir(src))
	if err != nil {
		t.Fatal(err)
	}
	wantAbs := filepath.Join(resolvedDir, "README.md") + "#usage"
	if !strings.Contains(text, "]("+wantAbs+")") {
		t.Fatalf("relative link must absolutize with fragment:\n%s", text)
	}
	if !strings.Contains(text, "https://example.com") || !strings.Contains(text, "(/etc/hosts)") {
		t.Fatal("global targets must stay untouched")
	}
	if strings.Contains(text, "See also") || strings.Contains(text, "sibling.md") {
		t.Fatalf("navigation section must be stripped:\n%s", text)
	}
}

func TestSymlinkBackupAndOptOut(t *testing.T) {
	e := newEnv(t)
	src := e.write(t, "src/AGENTS.COMPOSE.md", "# Doc\n")
	e.write(t, "links/CLAUDE.md", "hand-authored\n")
	e.config(t, "sources:\n  - "+src+"\n")
	if code, out, _ := e.run(t, false); code != 0 || !strings.Contains(out, "backed up prior file") {
		t.Fatalf("regular file at load point must be backed up: %s", out)
	}
	if readFile(t, e.claude+".bak") != "hand-authored\n" {
		t.Fatal("backup must preserve the prior file")
	}

	fresh := newEnv(t)
	src2 := fresh.write(t, "src/AGENTS.COMPOSE.md", "# Doc\n")
	if err := os.MkdirAll(filepath.Dir(fresh.paths.Config), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(fresh.paths.Config, []byte("sources:\n  - "+src2+"\nload_points:\n  claude: "+fresh.claude+"\n  codex: null\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if code, _, _ := fresh.run(t, false); code != 0 {
		t.Fatal("opt-out run failed")
	}
	if _, err := os.Lstat(fresh.codex); !os.IsNotExist(err) {
		t.Fatal("opted-out harness must never gain a link")
	}
	if _, err := os.Lstat(fresh.claude); err != nil {
		t.Fatal("remaining harness must still be linked")
	}
}

func TestMountEligibilityManifest(t *testing.T) {
	e := newEnv(t)
	inRepo := e.write(t, "projects/org/repo/deep/AGENTS.COMPOSE.md", "# InRepo\n")
	outOfRepo := e.write(t, "elsewhere/AGENTS.COMPOSE.md", "# Outside\n")
	e.config(t, "sources:\n  - "+inRepo+"\n  - "+outOfRepo+"\n")
	if code, _, errOut := e.run(t, false); code != 0 {
		t.Fatal(errOut)
	}
	manifest := readFile(t, filepath.Join(filepath.Dir(e.paths.Composed), "mount-eligibility.json"))
	projects, err := filepath.EvalSymlinks(e.projects)
	if err != nil {
		t.Fatal(err)
	}
	inRepoPath := strings.ReplaceAll(filepath.Join(projects, "org", "repo"), `\`, `\\`)
	if !strings.Contains(manifest, inRepoPath) {
		t.Fatalf("in-repo source must make its repo mountable:\n%s", manifest)
	}
	defaultPath := strings.ReplaceAll(filepath.Join(projects, "coilyco-bridge", "lore"), `\`, `\\`)
	if !strings.Contains(manifest, defaultPath) {
		t.Fatalf("defaults must be unioned in:\n%s", manifest)
	}
	if strings.Contains(manifest, "elsewhere") {
		t.Fatalf("out-of-repo source must back no repo:\n%s", manifest)
	}
}

func TestDryRunTouchesNothing(t *testing.T) {
	e := newEnv(t)
	src := e.write(t, "src/AGENTS.COMPOSE.md", "# Doc\n")
	e.config(t, "sources:\n  - "+src+"\n")
	code, out, _ := e.run(t, true)
	if code != 0 || !strings.Contains(out, "would write") || !strings.Contains(out, "would link") {
		t.Fatalf("dry run must preview: %s", out)
	}
	if _, err := os.Stat(e.paths.Composed); !os.IsNotExist(err) {
		t.Fatal("dry run must write nothing")
	}

	if code, _, _ := e.run(t, false); code != 0 {
		t.Fatal("real run failed")
	}
	if code, out, _ := e.run(t, true); code != 0 || strings.Contains(out, "would") {
		t.Fatalf("converged dry run must preview as no-op: %s", out)
	}
}

func TestCheckDetectsDrift(t *testing.T) {
	e := newEnv(t)
	src := e.write(t, "src/AGENTS.COMPOSE.md", "# Doc\n")
	e.config(t, "sources:\n  - "+src+"\n")
	if code, _, _ := e.run(t, false); code != 0 {
		t.Fatal("run failed")
	}
	if code, out, _ := e.check(t); code != 0 || !strings.Contains(out, "in sync") {
		t.Fatalf("converged check must pass: %d %s", code, out)
	}

	if err := os.WriteFile(e.paths.Composed, []byte(Banner+"\nhand edit\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	code, _, errOut := e.check(t)
	if code != 1 || !strings.Contains(errOut, "drift") || !strings.Contains(errOut, "+") {
		t.Fatalf("drift must fail with a diff: %d %s", code, errOut)
	}
}

func TestAmbiguousOverrideHeadingFails(t *testing.T) {
	e := newEnv(t)
	e.write(t, "src/doc/AGENTS.COMPOSE.md", "## Dup\n\none\n\n## Dup\n\ntwo\n")
	e.write(t, "src/doc/AGENTS.claude.md", "## Dup\n\npatched\n")
	e.config(t, "roots:\n  - "+filepath.Join(e.dir, "src")+"\n")
	code, _, errOut := e.run(t, false)
	if code != 1 || !strings.Contains(errOut, "matches 2 sections") {
		t.Fatalf("ambiguous override must fail loudly: %d %s", code, errOut)
	}
}

func readFile(t *testing.T, path string) string {
	t.Helper()
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(raw)
}
