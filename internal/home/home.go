// Package home resolves the single state directory, ~/.agent-compose, and
// migrates a pre-consolidation ~/.config/agent-compose into it once.
package home

import (
	"fmt"
	"os"
	"path/filepath"
)

// Dir returns the canonical state directory, performing the one-time
// migration when only the legacy directory exists.
func Dir() (string, error) {
	root, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return dirFrom(root, os.Stderr.WriteString)
}

func dirFrom(root string, warn func(string) (int, error)) (string, error) {
	canonical := filepath.Join(root, ".agent-compose")
	legacy := filepath.Join(root, ".config", "agent-compose")

	if info, err := os.Lstat(canonical); err == nil && info.IsDir() {
		if legacyInfo, err := os.Lstat(legacy); err == nil && legacyInfo.IsDir() && legacyInfo.Mode()&os.ModeSymlink == 0 {
			warn(fmt.Sprintf("agent-compose: both %s and %s exist; using %s (reconcile and symlink the legacy dir)\n", canonical, legacy, canonical))
		}
		return canonical, nil
	}

	legacyInfo, err := os.Lstat(legacy)
	if err != nil || legacyInfo.Mode()&os.ModeSymlink != 0 || !legacyInfo.IsDir() {
		return canonical, nil
	}

	if err := os.Rename(legacy, canonical); err != nil {
		return "", fmt.Errorf("migrate %s to %s: %w", legacy, canonical, err)
	}
	if err := os.Symlink(canonical, legacy); err != nil {
		return "", fmt.Errorf("leave compatibility symlink at %s: %w", legacy, err)
	}
	warn(fmt.Sprintf("agent-compose: migrated %s to %s (compatibility symlink left behind)\n", legacy, canonical))
	return canonical, nil
}
