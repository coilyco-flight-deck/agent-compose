package nativeui

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// WriteAll lays out themes/ plus one settings fragment per role. Convergence
// installs the tree, so nothing here touches ~/.claude.
func WriteAll(dir string, bundles []Bundle) (int, error) {
	themeDir := filepath.Join(dir, "themes")
	if err := os.MkdirAll(themeDir, 0o755); err != nil {
		return 0, err
	}
	written := 0
	for _, bundle := range bundles {
		theme := filepath.Join(themeDir, bundle.Slug+".json")
		if err := writeJSON(theme, bundle.Theme); err != nil {
			return written, fmt.Errorf("write theme for role %q: %w", bundle.Role, err)
		}
		written++

		settings := filepath.Join(dir, "settings."+bundle.Role+".json")
		if err := writeJSON(settings, bundle.Settings); err != nil {
			return written, fmt.Errorf("write settings for role %q: %w", bundle.Role, err)
		}
		written++
	}
	return written, nil
}

func writeJSON(path string, value any) error {
	raw, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, append(raw, '\n'), 0o644)
}
