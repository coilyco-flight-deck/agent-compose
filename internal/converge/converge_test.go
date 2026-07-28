package converge

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"forgejo.coilysiren.me/coilyco-flight-deck/agent-compose/internal/cascade"
	"forgejo.coilysiren.me/coilyco-flight-deck/agent-compose/internal/person"
	"forgejo.coilysiren.me/coilyco-flight-deck/agent-compose/internal/remoteskills"
)

func run(t *testing.T, paths cascade.Paths) (int, string, string) {
	t.Helper()
	var out, errOut bytes.Buffer
	code := Run(paths, Options{}, &out, &errOut)
	return code, out.String(), errOut.String()
}

func TestConvergeComposesRosterIntoCascade(t *testing.T) {
	dir := t.TempDir()
	paths := cascade.Paths{
		Config:       filepath.Join(dir, "agent-compose.yaml"),
		Composed:     filepath.Join(dir, "COMPOSED.md"),
		ProjectsRoot: filepath.Join(dir, "projects"),
		Home:         filepath.Join(dir, "home"),
	}
	doctrine := filepath.Join(dir, "doctrine", "AGENTS.COMPOSE.md")
	if err := os.MkdirAll(filepath.Dir(doctrine), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(doctrine, []byte("# Doctrine\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	providerRoot := filepath.Join(paths.ProjectsRoot, "coilyco-flight-deck", "agentic-os")
	skillRoot := filepath.Join(providerRoot, ".agents", "skills")
	if err := os.MkdirAll(filepath.Join(skillRoot, "coding-go"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(skillRoot, "coding-go", "SKILL.md"), []byte("# Go\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	mcpInventory := filepath.Join(dir, "mcporter.json")
	if err := os.WriteFile(mcpInventory, []byte(`{"imports":[],"mcpServers":{"reader":{"baseUrl":"https://mcp.example.test/mcp"}}}`+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	config := "sources:\n  - " + doctrine + "\nroots:\n  - " + filepath.Join(dir, "sources") + "\n" +
		"roster_sources:\n  - " + providerRoot + "\n" +
		"skill_load_points:\n  codex: " + filepath.Join(dir, "links", "skills") + "\n" +
		"mcp_inventory: " + mcpInventory + "\n" +
		"load_points:\n  claude: " + filepath.Join(dir, "links", "CLAUDE.md") + "\n  codex: null\n"
	if err := os.WriteFile(paths.Config, []byte(config), 0o644); err != nil {
		t.Fatal(err)
	}

	code, out, errOut := run(t, paths)
	if code != 0 {
		t.Fatalf("converge failed: %s %s", out, errOut)
	}
	if !strings.Contains(out, "roster  ") || !strings.Contains(out, "wrote") {
		t.Fatalf("converge must refresh roster then cascade: %s", out)
	}
	composed := readFile(t, paths.Composed)
	if !strings.Contains(composed, "# Doctrine") ||
		!strings.Contains(composed, "# Personality invariant") ||
		!strings.Contains(composed, "# Agent seats") {
		t.Fatalf("composed output must carry doctrine and the dispatch table:\n%s", composed)
	}
	if !strings.Contains(composed, "opal engineer") {
		t.Fatalf("dispatch table incomplete:\n%s", composed)
	}
	if _, err := os.Stat(filepath.Join(dir, "sources", "personality", "personalities", "curious.md")); err != nil {
		t.Fatal("personality bodies must land under sources/personality")
	}
	personSnapshot := readFile(t, filepath.Join(dir, "sources", "personality", "person.json"))
	if !strings.Contains(personSnapshot, fmt.Sprintf(`"format": %q`, person.SnapshotFormat)) ||
		!strings.Contains(personSnapshot, `"briefing":`) {
		t.Fatal("normal convergence must emit the complete versioned person snapshot")
	}
	expectedSkillTarget, err := filepath.EvalSymlinks(filepath.Join(skillRoot, "coding-go"))
	if err != nil {
		t.Fatal(err)
	}
	if target, err := os.Readlink(filepath.Join(dir, "links", "skills", "coding-go")); err != nil || target != expectedSkillTarget {
		t.Fatalf("skill root must mount before cascade: target=%q err=%v", target, err)
	}
	for _, path := range []string{
		filepath.Join(paths.Home, ".mcporter", "mcporter.json"),
		filepath.Join(paths.Home, ".claude.json"),
		filepath.Join(paths.Home, ".codex", "config.toml"),
	} {
		if _, err := os.Stat(path); err != nil {
			t.Fatalf("native MCP projection missing %s: %v", path, err)
		}
	}

	code, out, _ = run(t, paths)
	if code != 0 || strings.Contains(out, "wrote") {
		t.Fatalf("second converge must not rewrite current cascade state: %s", out)
	}
	if !strings.Contains(out, "cascade outputs=1 load-points=1 manifest=1 changed=0") {
		t.Fatalf("second converge must summarize the full cascade check: %s", out)
	}
	if !strings.Contains(out, "mcp     servers=1 state=unchanged") {
		t.Fatalf("second converge must report stable native MCP projection: %s", out)
	}
	skillEntries, err := os.ReadDir(skillRoot)
	if err != nil {
		t.Fatal(err)
	}
	wantSkillSummary := fmt.Sprintf(
		"skills  managed=%d load-points=1 verified=%d linked=0 removed=0 preserved=0",
		len(skillEntries),
		len(skillEntries),
	)
	if !strings.Contains(out, wantSkillSummary) {
		t.Fatalf("second converge must summarize the full skill check: %s", out)
	}
}

func TestConvergeWithoutConfigIsNoOp(t *testing.T) {
	dir := t.TempDir()
	paths := cascade.Paths{
		Config:       filepath.Join(dir, "agent-compose.yaml"),
		Composed:     filepath.Join(dir, "COMPOSED.md"),
		ProjectsRoot: dir,
	}
	code, out, _ := run(t, paths)
	if code != 0 || !strings.Contains(out, "nothing to do (opt-in)") {
		t.Fatalf("absent config must stay the documented no-op: %d %s", code, out)
	}
	if _, err := os.Stat(filepath.Join(dir, "sources")); !os.IsNotExist(err) {
		t.Fatal("no-op must write nothing")
	}
}

func TestConvergeHydratesRemoteSkillsThroughLocalProjection(t *testing.T) {
	dir := t.TempDir()
	paths := cascade.Paths{
		Config:       filepath.Join(dir, "agent-compose.yaml"),
		Composed:     filepath.Join(dir, "COMPOSED.md"),
		ProjectsRoot: filepath.Join(dir, "projects"),
		Home:         filepath.Join(dir, "home"),
	}
	doctrine := filepath.Join(dir, "doctrine", "AGENTS.COMPOSE.md")
	if err := os.MkdirAll(filepath.Dir(doctrine), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(doctrine, []byte("# Doctrine\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	local := filepath.Join(paths.ProjectsRoot, "coilyco-flight-deck", "agentic-os")
	writeTestSkill(t, filepath.Join(local, ".agents", "skills"), "shared", "local")

	origin := filepath.Join(dir, "remote")
	writeTestSkill(t, filepath.Join(origin, ".agents", "skills"), "shared", "remote")
	writeTestSkill(t, filepath.Join(origin, ".agents", "skills"), "remote-only", "remote")
	runTestGit(t, origin, "init", "-b", "main", ".")
	runTestGit(t, origin, "add", ".")
	runTestGit(t, origin, "commit", "-m", "skills")

	skillLoadPoint := filepath.Join(dir, "links", "skills")
	config := "sources:\n  - " + doctrine + "\n" +
		"skill_load_points:\n  codex: " + skillLoadPoint + "\n" +
		"remote_skill_cache_ttl: 1h\n" +
		"remote_skill_sources:\n" +
		"  - url: " + origin + "\n" +
		"    harnesses: [codex]\n" +
		"load_points:\n  claude: null\n  codex: " + filepath.Join(dir, "links", "AGENTS.md") + "\n"
	if err := os.WriteFile(paths.Config, []byte(config), 0o644); err != nil {
		t.Fatal(err)
	}

	code, out, errOut := run(t, paths)
	if code != 0 {
		t.Fatalf("remote converge failed: %s %s", out, errOut)
	}
	if !strings.Contains(out, "remote  sources=1 cached=0 hydrated=1 refreshed=0 fallback=0") {
		t.Fatalf("remote hydration summary missing: %s", out)
	}
	shared, err := os.Readlink(filepath.Join(skillLoadPoint, "shared"))
	if err != nil {
		t.Fatal(err)
	}
	if body := readFile(t, filepath.Join(shared, "SKILL.md")); body != "remote" {
		t.Fatalf("remote catalog did not overlay local catalog: %q", body)
	}
	if _, err := os.Readlink(filepath.Join(skillLoadPoint, "remote-only")); err != nil {
		t.Fatal("remote-only skill was not projected")
	}

	code, out, errOut = run(t, paths)
	if code != 0 || !strings.Contains(out, "remote  sources=1 cached=1 hydrated=0 refreshed=0 fallback=0") {
		t.Fatalf("fresh remote cache was not reused: %d %s %s", code, out, errOut)
	}
}

func TestActiveRemoteSourcesSkipsUnusedHarnesses(t *testing.T) {
	sources := []remoteskills.Source{
		{URL: "all"},
		{URL: "codex", Harnesses: []string{" codex "}},
		{URL: "claude", Harnesses: []string{"claude"}},
	}
	if got := activeRemoteSources(sources, nil); len(got) != 0 {
		t.Fatalf("sources without load points = %v", got)
	}
	got := activeRemoteSources(sources, map[string]string{"codex": "/skills"})
	if len(got) != 2 || got[0].URL != "all" || got[1].URL != "codex" {
		t.Fatalf("active sources = %+v", got)
	}
}

func TestConvergeUsesConfiguredExternalPersonExclusively(t *testing.T) {
	dir := t.TempDir()
	paths := cascade.Paths{
		Config:       filepath.Join(dir, "agent-compose.yaml"),
		Composed:     filepath.Join(dir, "COMPOSED.md"),
		ProjectsRoot: dir,
		Home:         filepath.Join(dir, "home"),
	}
	personRoot, err := filepath.Abs(filepath.Join(
		"..", "..", "testdata", "contracts", "person-independent",
	))
	if err != nil {
		t.Fatal(err)
	}
	config := "person_policy: external-only\n" +
		"person_source: " + personRoot + "\n" +
		"roots:\n  - " + filepath.Join(dir, "sources") + "\n" +
		"load_points:\n  claude: " + filepath.Join(dir, "links", "CLAUDE.md") + "\n" +
		"  codex: null\n"
	if err := os.WriteFile(paths.Config, []byte(config), 0o644); err != nil {
		t.Fatal(err)
	}

	code, out, errOut := run(t, paths)
	if code != 0 {
		t.Fatalf("external converge failed: %s %s", out, errOut)
	}
	composed := readFile(t, paths.Composed)
	if !strings.Contains(composed, "workbench builder") ||
		!strings.Contains(composed, "# Workbench invariant") {
		t.Fatalf("external person did not reach host context:\n%s", composed)
	}
	if strings.Contains(composed, "opal engineer") ||
		strings.Contains(composed, "# Personality invariant") {
		t.Fatalf("external person inherited embedded context:\n%s", composed)
	}
	snapshot := readFile(t, filepath.Join(dir, "sources", "personality", "person.json"))
	if !strings.Contains(snapshot, `"source": "person:workbench"`) ||
		strings.Contains(snapshot, `"person:kai"`) {
		t.Fatalf("external person snapshot crossed source boundaries:\n%s", snapshot)
	}
}

func TestConvergeExternalOnlyFailsWithoutSource(t *testing.T) {
	dir := t.TempDir()
	paths := cascade.Paths{
		Config:       filepath.Join(dir, "agent-compose.yaml"),
		Composed:     filepath.Join(dir, "COMPOSED.md"),
		ProjectsRoot: dir,
		Home:         filepath.Join(dir, "home"),
	}
	config := "person_policy: external-only\n" +
		"roots:\n  - " + filepath.Join(dir, "sources") + "\n"
	if err := os.WriteFile(paths.Config, []byte(config), 0o644); err != nil {
		t.Fatal(err)
	}
	code, _, errOut := run(t, paths)
	if code != 1 || !strings.Contains(errOut, "requires person_source") {
		t.Fatalf("unguarded convergence did not fail closed: %d %s", code, errOut)
	}
	if _, err := os.Stat(filepath.Join(dir, "sources", "personality")); !os.IsNotExist(err) {
		t.Fatalf("failed convergence wrote a roster: %v", err)
	}
}

func TestConvergeRetainsEmbeddedPersonalitiesOnBadRosterSource(t *testing.T) {
	dir := t.TempDir()
	paths := cascade.Paths{
		Config:       filepath.Join(dir, "agent-compose.yaml"),
		Composed:     filepath.Join(dir, "COMPOSED.md"),
		ProjectsRoot: dir,
	}
	config := "roots:\n  - " + filepath.Join(dir, "sources") + "\n" +
		"roster_sources:\n  - " + filepath.Join(dir, "missing.kdl") + "\n" +
		"load_points:\n  claude: " + filepath.Join(dir, "links", "CLAUDE.md") + "\n  codex: null\n"
	if err := os.WriteFile(paths.Config, []byte(config), 0o644); err != nil {
		t.Fatal(err)
	}
	code, out, errOut := run(t, paths)
	if code != 0 || !strings.Contains(errOut, "warning: roster source") {
		t.Fatalf("bad roster source must warn and continue: %d %s", code, errOut)
	}
	composed := readFile(t, paths.Composed)
	if strings.Contains(composed, "definition pending") ||
		!strings.Contains(composed, "# Personality invariant") {
		t.Fatalf("embedded personality definitions must survive a bad overlay: %s", out)
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

func writeTestSkill(t *testing.T, root, name, body string) {
	t.Helper()
	path := filepath.Join(root, name)
	if err := os.MkdirAll(path, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(path, "SKILL.md"), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

func runTestGit(t *testing.T, dir string, args ...string) {
	t.Helper()
	argv := append([]string{
		"-C", dir,
		"-c", "user.email=test@example.test",
		"-c", "user.name=Test",
		"-c", "commit.gpgSign=false",
	}, args...)
	if raw, err := exec.Command("git", argv...).CombinedOutput(); err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, raw)
	}
}
