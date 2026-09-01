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
