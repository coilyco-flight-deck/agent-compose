package person

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing/fstest"

	"gopkg.in/yaml.v3"
)

// The binary carries no roster and mounts one instead.
// Search order and install locations: docs/FEATURES.md.

// RosterEnv names the explicit override, which wins over every search location.
const RosterEnv = "AGENT_COMPOSE_ROSTER"

const seedLabel = "core roster"

// rosterSearch returns the candidate roots in precedence order, so a diagnostic
// and the resolver cannot disagree about where it looked.
func rosterSearch() []string {
	var roots []string
	if override := strings.TrimSpace(os.Getenv(RosterEnv)); override != "" {
		roots = append(roots, override)
	}
	if hostHome, err := os.UserHomeDir(); err == nil {
		roots = append(roots, filepath.Join(hostHome, ".agent-compose", "roster"))
	}
	if executable, err := os.Executable(); err == nil {
		if resolved, err := filepath.EvalSymlinks(executable); err == nil {
			executable = resolved
		}
		binDir := filepath.Dir(executable)
		// Homebrew installs into its own prefix rather than the state
		// directory, and scoop keeps the seed next to the executable.
		roots = append(roots,
			filepath.Join(filepath.Dir(binDir), "share", "agent-compose", "roster"),
			filepath.Join(binDir, "roster"),
		)
	}
	return roots
}

// rosterSeed resolves the mounted roster, or reports every path it tried.
func rosterSeed() (fs.FS, string, error) {
	roots := rosterSearch()
	for index, root := range roots {
		if !isDirectory(root) {
			continue
		}
		overlay, err := isRosterOverlay(root)
		if err != nil {
			return nil, "", err
		}
		if !overlay {
			return os.DirFS(root), seedLabel + " " + root, nil
		}
		return layeredSeed(root, roots[index+1:])
	}
	if len(roots) == 0 {
		return nil, "", fmt.Errorf(
			"no roster is mounted and no search location resolved; set %s",
			RosterEnv,
		)
	}
	return nil, "", fmt.Errorf(
		"no roster is mounted; searched %s. Install the seed or set %s",
		strings.Join(roots, ", "),
		RosterEnv,
	)
}

// RosterProvenance reports the resolved roster root and whether reaching it
// crosses a symlink into a tree another session owns (agent-compose#7363).
func RosterProvenance() (root string, target string, volatile bool) {
	root, target, kind := rosterProvenanceKind()
	return root, target, kind != ""
}

// SessionShadow is the provenance worth interrupting for: another live session
// owns the roster and may edit it mid-compose (agent-compose#7363).
const SessionShadow = "session"

// RosterProvenanceKind is RosterProvenance plus which kind, so a caller can say
// "another session owns this" without saying it about an ordinary temp dir.
func RosterProvenanceKind() (root string, target string, kind string) {
	return rosterProvenanceKind()
}

func rosterProvenanceKind() (root string, target string, kind string) {
	for _, candidate := range rosterSearch() {
		info, err := os.Stat(candidate)
		if err != nil || !info.IsDir() {
			continue
		}
		// Volatility is where the roster lands, not whether a symlink was
		// crossed: a roster sitting directly in a temp dir was invisible.
		resolved, err := filepath.EvalSymlinks(candidate)
		if err != nil {
			resolved = candidate
		}
		target := resolved
		if resolved == candidate {
			target = ""
		}
		return candidate, target, volatileKind(resolved)
	}
	return "", "", ""
}

// isVolatileRoot reports a path the operating system may purge, or one inside
// a per-session tree. Both mean the roster has an owner other than this seat.
func isVolatileRoot(path string) bool { return volatileKind(path) != "" }

// volatileKind names why a path is not stable, or is empty when it is.
func volatileKind(path string) string {
	sep := string(os.PathSeparator)
	// A session shadow is the case that actually bites, and it is named
	// rather than inferred, because TMPDIR does not cover it on every host.
	if strings.Contains(path, sep+"aos"+sep+"native"+sep) {
		return SessionShadow
	}
	roots := []string{os.TempDir(), "/tmp", "/private/tmp", "/var/folders"}
	for _, root := range roots {
		root = strings.TrimSuffix(strings.TrimSpace(root), sep)
		if root == "" {
			continue
		}
		if strings.HasPrefix(path, root+sep) {
			return "temporary"
		}
		if resolved, err := filepath.EvalSymlinks(root); err == nil && resolved != root {
			if strings.HasPrefix(path, strings.TrimSuffix(resolved, sep)+sep) {
				return "temporary"
			}
		}
	}
	return ""
}

// overlayMarker makes a roster root add entities on top of the next roster in
// the search order instead of replacing it. See docs/roster-composition.md.
const overlayMarker = "overlay" + yamlFragmentExt

func isDirectory(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}

func isRosterOverlay(root string) (bool, error) {
	raw, err := os.ReadFile(filepath.Join(root, overlayMarker))
	if os.IsNotExist(err) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	var marker struct {
		LayersOver string `yaml:"layers_over"`
	}
	if err := yaml.Unmarshal(raw, &marker); err != nil || marker.LayersOver != "core" {
		return false, fmt.Errorf("%s: %s must say layers_over: core", root, overlayMarker)
	}
	return true, nil
}

// layeredSeed stacks the overlay root on the next plain roster it finds.
func layeredSeed(overlay string, rest []string) (fs.FS, string, error) {
	for _, base := range rest {
		if !isDirectory(base) {
			continue
		}
		if stacked, err := isRosterOverlay(base); err != nil || stacked {
			return nil, "", fmt.Errorf("roster overlay %s sits on overlay %s; overlays do not stack", overlay, base)
		}
		layered, err := layerRoster(os.DirFS(overlay), os.DirFS(base))
		if err != nil {
			return nil, "", err
		}
		return layered, seedLabel + " " + overlay + " over " + base, nil
	}
	return nil, "", fmt.Errorf("roster overlay %s has no roster under it; searched %s",
		overlay, strings.Join(rest, ", "))
}

// layerRoster replaces whole entity directories, never single files, so an
// overlay role cannot end up with its base namesake's SKILL.md.
func layerRoster(top, base fs.FS) (fs.FS, error) {
	layered := fstest.MapFS{}
	shadowed := func(path string) bool {
		if path == overlayMarker {
			return true
		}
		if _, err := fs.Stat(top, path); err == nil && !strings.Contains(path, "/") {
			return true
		}
		parts := strings.SplitN(path, "/", 3)
		if len(parts) == 3 && parts[0] == dataRoot {
			_, err := fs.Stat(top, dataRoot+"/"+parts[1])
			return err == nil
		}
		return false
	}
	copyTree := func(source fs.FS, skip func(string) bool) error {
		return fs.WalkDir(source, ".", func(path string, entry fs.DirEntry, err error) error {
			if err != nil || entry.IsDir() || skip(path) {
				return err
			}
			raw, err := fs.ReadFile(source, path)
			if err != nil {
				return err
			}
			layered[path] = &fstest.MapFile{Data: raw, Mode: 0o644}
			return nil
		})
	}
	if err := copyTree(base, shadowed); err != nil {
		return nil, err
	}
	if err := copyTree(top, func(path string) bool { return path == overlayMarker }); err != nil {
		return nil, err
	}
	return layered, nil
}
