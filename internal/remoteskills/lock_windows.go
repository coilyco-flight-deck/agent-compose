//go:build windows

package remoteskills

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

func lockCache(cacheRoot string) (func(), error) {
	path := filepath.Join(cacheRoot, "lock.excl")
	deadline := time.Now().Add(10 * time.Second)
	for {
		file, err := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o644)
		if err == nil {
			_ = file.Close()
			return func() { _ = os.Remove(path) }, nil
		}
		if !os.IsExist(err) {
			return nil, fmt.Errorf("lock remote skill cache: %w", err)
		}
		if time.Now().After(deadline) {
			return nil, fmt.Errorf("remote skill cache is locked by another process")
		}
		time.Sleep(50 * time.Millisecond)
	}
}
