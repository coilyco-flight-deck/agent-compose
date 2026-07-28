//go:build !windows

package remoteskills

import (
	"fmt"
	"os"
	"path/filepath"
	"syscall"
)

func lockCache(cacheRoot string) (func(), error) {
	file, err := os.OpenFile(filepath.Join(cacheRoot, "lock"), os.O_CREATE|os.O_RDWR, 0o644)
	if err != nil {
		return nil, err
	}
	if err := syscall.Flock(int(file.Fd()), syscall.LOCK_EX); err != nil {
		file.Close()
		return nil, fmt.Errorf("lock remote skill cache: %w", err)
	}
	return func() {
		_ = syscall.Flock(int(file.Fd()), syscall.LOCK_UN)
		_ = file.Close()
	}, nil
}
