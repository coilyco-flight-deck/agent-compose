package person

import (
	"fmt"
	"io/fs"
	"regexp"
	"sort"
	"strings"
)

// bannedIdentityTokens catches a person-name or owner-org token in
// roster:core, which ships to anyone who installs agent-compose. agent-compose#8039.
var bannedIdentityTokens = regexp.MustCompile(`(?i)\b(kai|coilyco|coilysiren)\b`)

// scanForBannedIdentityTokens walks every regular file under source and
// reports one violation per matching line, sorted for a stable message.
func scanForBannedIdentityTokens(source fs.FS) ([]string, error) {
	var violations []string
	err := fs.WalkDir(source, ".", func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() {
			return nil
		}
		raw, err := fs.ReadFile(source, path)
		if err != nil {
			return fmt.Errorf("read %s: %w", path, err)
		}
		for i, line := range strings.Split(string(raw), "\n") {
			if match := bannedIdentityTokens.FindString(line); match != "" {
				violations = append(violations, fmt.Sprintf("%s:%d: %s", path, i+1, match))
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	sort.Strings(violations)
	return violations, nil
}

// validateNoPersonOrOwnerLeak fails closed on a person-name or owner token
// anywhere in the loaded source tree. See agent-compose#8039.
func validateNoPersonOrOwnerLeak(p *Person) error {
	violations, err := scanForBannedIdentityTokens(p.source)
	if err != nil {
		return err
	}
	if len(violations) > 0 {
		return fmt.Errorf("roster:core names a person or owner:\n%s", strings.Join(violations, "\n"))
	}
	return nil
}
