//go:build windows

package project

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// lockTarget serializes projections via an exclusive lock file; Windows has
// no flock, so contenders retry briefly instead of blocking indefinitely.
func lockTarget(targetDir string) (func(), error) {
	dir := filepath.Join(targetDir, ".agent-compose")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, err
	}
	lockPath := filepath.Join(dir, "lock.excl")
	deadline := time.Now().Add(10 * time.Second)
	for {
		f, err := os.OpenFile(lockPath, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o644)
		if err == nil {
			f.Close()
			return func() { os.Remove(lockPath) }, nil
		}
		if !os.IsExist(err) {
			return nil, fmt.Errorf("lock projection target: %w", err)
		}
		if time.Now().After(deadline) {
			return nil, fmt.Errorf("projection target is locked by another process (remove %s if stale)", lockPath)
		}
		time.Sleep(50 * time.Millisecond)
	}
}
