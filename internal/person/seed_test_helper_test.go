package person

import (
	"io/fs"
	"testing"
)

// seedForTest opens the roster the test run mounted, so a test comparing
// against the shipped bodies reads the same tree the loader did.
func seedForTest(t *testing.T) fs.FS {
	t.Helper()
	source, _, err := rosterSeed()
	if err != nil {
		t.Fatalf("resolve roster seed: %v", err)
	}
	return source
}
