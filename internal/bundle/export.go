package bundle

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"time"
)

// Export writes a verified bundle as a deterministic gzip-compressed tarball.
func Export(dir, output string) error {
	if _, err := Verify(dir); err != nil {
		return fmt.Errorf("verify bundle before export: %w", err)
	}
	paths := []string{}
	err := filepath.WalkDir(dir, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() {
			return nil
		}
		rel, err := filepath.Rel(dir, path)
		if err != nil {
			return err
		}
		if !safePortablePath(filepath.ToSlash(rel)) {
			return fmt.Errorf("unsafe bundle export path %q", rel)
		}
		paths = append(paths, rel)
		return nil
	})
	if err != nil {
		return fmt.Errorf("enumerate bundle export: %w", err)
	}
	sort.Strings(paths)
	file, err := os.OpenFile(output, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o644)
	if err != nil {
		return fmt.Errorf("open bundle export %s: %w", output, err)
	}
	defer file.Close()
	gzipWriter := gzip.NewWriter(file)
	gzipWriter.Header.ModTime = time.Unix(0, 0)
	gzipWriter.Header.OS = 255
	tarWriter := tar.NewWriter(gzipWriter)
	for _, rel := range paths {
		path := filepath.Join(dir, rel)
		input, err := os.Open(path)
		if err != nil {
			return err
		}
		info, err := input.Stat()
		if err != nil {
			input.Close()
			return err
		}
		header := &tar.Header{Name: filepath.ToSlash(rel), Mode: 0o644, Size: info.Size(), ModTime: time.Unix(0, 0), Format: tar.FormatUSTAR}
		if err := tarWriter.WriteHeader(header); err != nil {
			input.Close()
			return err
		}
		if _, err := io.Copy(tarWriter, input); err != nil {
			input.Close()
			return err
		}
		if err := input.Close(); err != nil {
			return err
		}
	}
	if err := tarWriter.Close(); err != nil {
		return err
	}
	if err := gzipWriter.Close(); err != nil {
		return err
	}
	return file.Close()
}
