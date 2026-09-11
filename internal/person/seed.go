package person

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
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
	for _, root := range roots {
		info, err := os.Stat(root)
		if err != nil || !info.IsDir() {
			continue
		}
		return os.DirFS(root), seedLabel + " " + root, nil
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
		resolved, err := filepath.EvalSymlinks(candidate)
		if err != nil || resolved == candidate {
			return candidate, "", ""
		}
		return candidate, resolved, volatileKind(resolved)
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
