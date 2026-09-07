// Package rostertest points a test binary at the repository's seed roster.
//
// The binary mounts a roster rather than embedding one, and only the justfile
// exported the path, so `go test ./...` failed in every package that loads a
// person. Search order and install locations: docs/FEATURES.md.
package rostertest

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// Env names the override the person loader consults first.
const Env = "AGENT_COMPOSE_ROSTER"

// Use points Env at seed/roster unless the caller already set it, so a
// deliberate override from a contract test or CI still wins.
func Use() {
	if strings.TrimSpace(os.Getenv(Env)) != "" {
		return
	}
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		return
	}
	root := filepath.Dir(filepath.Dir(filepath.Dir(file)))
	seed := filepath.Join(root, "seed", "roster")
	if info, err := os.Stat(seed); err == nil && info.IsDir() {
		os.Setenv(Env, seed)
	}
}
