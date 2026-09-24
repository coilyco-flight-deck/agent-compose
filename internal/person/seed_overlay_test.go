package person

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"testing/fstest"
)

func writeTree(t *testing.T, root string, files map[string]string) {
	t.Helper()
	for name, body := range files {
		path := filepath.Join(root, filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
}

func TestLayerRosterReplacesWholeEntityDirectories(t *testing.T) {
	base := fstest.MapFS{
		"person.yaml":                 {Data: []byte("base person")},
		"data/role-a/role.yaml":       {Data: []byte("base a")},
		"data/role-a/SKILL.md":        {Data: []byte("base a skill")},
		"data/role-b/role.yaml":       {Data: []byte("base b")},
		"data/invariant/INVARIANT.md": {Data: []byte("invariant")},
	}
	top := fstest.MapFS{
		overlayMarker:           {Data: []byte("layers_over: core\n")},
		"data/role-a/role.yaml": {Data: []byte("top a")},
		"data/role-c/role.yaml": {Data: []byte("top c")},
	}
	layered, err := layerRoster(top, base)
	if err != nil {
		t.Fatal(err)
	}
	read := func(path string) string {
		raw, err := fs.ReadFile(layered, path)
		if err != nil {
			return "<missing>"
		}
		return string(raw)
	}
	want := map[string]string{
		"person.yaml":                 "base person",
		"data/role-a/role.yaml":       "top a",
		"data/role-a/SKILL.md":        "<missing>",
		"data/role-b/role.yaml":       "base b",
		"data/role-c/role.yaml":       "top c",
		"data/invariant/INVARIANT.md": "invariant",
		overlayMarker:                 "<missing>",
	}
	for path, body := range want {
		if got := read(path); got != body {
			t.Errorf("%s: got %q, want %q", path, got, body)
		}
	}
}

func TestRosterSeedStacksAnOverlayOnTheNextRoster(t *testing.T) {
	home := t.TempDir()
	overlay := t.TempDir()
	writeTree(t, filepath.Join(home, ".agent-compose", "roster"), map[string]string{
		"data/role-a/role.yaml": "base a",
	})
	writeTree(t, overlay, map[string]string{
		overlayMarker:           "layers_over: core\n",
		"data/role-b/role.yaml": "top b",
	})
	t.Setenv("HOME", home)
	t.Setenv(RosterEnv, overlay)
	source, label, err := rosterSeed()
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(label, " over ") {
		t.Errorf("the label must name both roots, got %q", label)
	}
	for _, path := range []string{"data/role-a/role.yaml", "data/role-b/role.yaml"} {
		if _, err := fs.Stat(source, path); err != nil {
			t.Errorf("%s missing from the layered roster", path)
		}
	}
}

func TestRosterSeedRefusesAMalformedOrStackedOverlay(t *testing.T) {
	home := t.TempDir()
	overlay := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv(RosterEnv, overlay)

	writeTree(t, overlay, map[string]string{overlayMarker: "layers_over: something-else\n"})
	if _, _, err := rosterSeed(); err == nil || !strings.Contains(err.Error(), "layers_over: core") {
		t.Errorf("a marker naming another base must be refused, got %v", err)
	}

	writeTree(t, overlay, map[string]string{overlayMarker: "layers_over: core\n"})
	writeTree(t, filepath.Join(home, ".agent-compose", "roster"), map[string]string{
		overlayMarker: "layers_over: core\n",
	})
	if _, _, err := rosterSeed(); err == nil || !strings.Contains(err.Error(), "do not stack") {
		t.Errorf("an overlay over an overlay must be refused, got %v", err)
	}
}
