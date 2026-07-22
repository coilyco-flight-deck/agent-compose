//go:build !windows

package project

import (
	"fmt"
	"os"
	"path/filepath"
	"syscall"
)

// lockTarget serializes projections into one target across processes.
func lockTarget(targetDir string) (func(), error) {
	dir := filepath.Join(targetDir, ".agent-compose")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, err
	}
	f, err := os.OpenFile(filepath.Join(dir, "lock"), os.O_CREATE|os.O_RDWR, 0o644)
	if err != nil {
		return nil, err
	}
	if err := syscall.Flock(int(f.Fd()), syscall.LOCK_EX); err != nil {
		f.Close()
		return nil, fmt.Errorf("lock projection target: %w", err)
	}
	return func() {
		syscall.Flock(int(f.Fd()), syscall.LOCK_UN)
		f.Close()
	}, nil
}
