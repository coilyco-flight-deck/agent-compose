package person

import (
	"fmt"
	"io/fs"
	"strings"
	"testing/fstest"

	"forgejo.coilysiren.me/coilyco-flight-deck/agent-compose/internal/schema"
)

// CoreLibraryRoot admits the embedded core personalities wherever a
// personality-library root is accepted. See docs/personality.md.
const CoreLibraryRoot = schema.CoreLibraryRoot

const coreLibraryLabel = "embedded core personality library"

// IsCoreLibrary reports whether an admitted personality-library value names the
// embedded library rather than a local directory, so callers skip path resolution.
func IsCoreLibrary(root string) bool {
	return strings.TrimSpace(root) == CoreLibraryRoot
}

// coreLibrarySource projects the embedded personalities onto the library layout.
// Only shared disposition crosses. Roles, seats, identity, and the invariant do not.
func coreLibrarySource() (fs.FS, error) {
	projected, _, err := dataLayout(embedded, coreLibraryLabel)
	if err != nil {
		return nil, err
	}
	manifest := fmt.Sprintf("library %q\n", CoreLibraryRoot)
	library := fstest.MapFS{
		"library.kdl": &fstest.MapFile{Data: []byte(manifest), Mode: 0o644},
	}
	walkErr := fs.WalkDir(projected, ".", func(path string, entry fs.DirEntry, err error) error {
		if err != nil || entry.IsDir() || !corePersonalityPath(path) {
			return err
		}
		raw, err := fs.ReadFile(projected, path)
		if err != nil {
			return err
		}
		library[path] = &fstest.MapFile{Data: raw, Mode: 0o644}
		return nil
	})
	if walkErr != nil {
		return nil, fmt.Errorf("%s: %w", coreLibraryLabel, walkErr)
	}
	if len(library) == 1 {
		return nil, fmt.Errorf("%s: the embedded roster declares no personalities", coreLibraryLabel)
	}
	return library, nil
}

func corePersonalityPath(path string) bool {
	return strings.HasPrefix(path, "personalities/") ||
		strings.HasPrefix(path, "definitions/skills/personality-")
}
