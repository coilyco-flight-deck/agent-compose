package converge

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"forgejo.coilysiren.me/coilyco-flight-deck/agent-compose/internal/cascade"
)

func run(t *testing.T, paths cascade.Paths) (int, string, string) {
	t.Helper()
	var out, errOut bytes.Buffer
	code := Run(paths, &out, &errOut)
	return code, out.String(), errOut.String()
}

func TestConvergeComposesRosterIntoCascade(t *testing.T) {
	dir := t.TempDir()
	paths := cascade.Paths{
		Config:       filepath.Join(dir, "agent-compose.yaml"),
		Composed:     filepath.Join(dir, "COMPOSED.md"),
		ProjectsRoot: filepath.Join(dir, "projects"),
	}
	doctrine := filepath.Join(dir, "doctrine", "AGENTS.COMPOSE.md")
	if err := os.MkdirAll(filepath.Dir(doctrine), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(doctrine, []byte("# Doctrine\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	skillRoot := filepath.Join(paths.ProjectsRoot, "coilyco-flight-deck", "agentic-os", ".agents", "skills")
	if err := os.MkdirAll(filepath.Join(skillRoot, "coding-go"), 0o755); err != nil {
		t.Fatal(err)
	}
	fixture, err := filepath.Abs(filepath.Join("..", "..", "testdata", "contracts", "source-public.kdl"))
	if err != nil {
		t.Fatal(err)
	}
	config := "sources:\n  - " + doctrine + "\nroots:\n  - " + filepath.Join(dir, "sources") + "\n" +
		"roster_sources:\n  - " + fixture + "\n" +
		"skill_load_points:\n  codex: " + filepath.Join(dir, "links", "skills") + "\n" +
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
	if !strings.Contains(composed, "# Doctrine") || !strings.Contains(composed, "# Agent seats") {
		t.Fatalf("composed output must carry doctrine and the dispatch table:\n%s", composed)
	}
	if !strings.Contains(composed, "opal engineer") {
		t.Fatalf("dispatch table incomplete:\n%s", composed)
	}
	if _, err := os.Stat(filepath.Join(dir, "sources", "personality", "personalities", "curious.md")); err != nil {
		t.Fatal("personality bodies must land under sources/personality")
	}
	if target, err := os.Readlink(filepath.Join(dir, "links", "skills", "coding-go")); err != nil || target != filepath.Join(skillRoot, "coding-go") {
		t.Fatalf("skill root must mount before cascade: target=%q err=%v", target, err)
	}

	code, out, _ = run(t, paths)
	if code != 0 || strings.Contains(out, "wrote") {
		t.Fatalf("second converge must be silent on the cascade side: %s", out)
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

func TestConvergeDegradesOnBadRosterSource(t *testing.T) {
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
	if !strings.Contains(readFile(t, paths.Composed), "definition pending") {
		t.Fatalf("table must degrade to pending bodies: %s", out)
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
