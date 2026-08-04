package project

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

func applyOwned(targetDir string, files map[string][]byte, label, origin string) (*Result, error) {
	return applyOwnedWithWriter(targetDir, files, label, origin, os.WriteFile)
}

func applyOwnedWithWriter(
	targetDir string,
	files map[string][]byte,
	label string,
	origin string,
	writeFile func(string, []byte, fs.FileMode) error,
) (*Result, error) {
	previous := ReadProjection(targetDir)
	owned := map[string]bool{}
	for _, rel := range previous.Files {
		if err := validTargetPath(rel); err != nil {
			return nil, fmt.Errorf("previous projection: %w", err)
		}
		if err := rejectSymlinkParents(targetDir, rel); err != nil {
			return nil, fmt.Errorf("previous projection: %w", err)
		}
		owned[rel] = true
	}
	effective := make(map[string][]byte, len(files))
	for rel, content := range files {
		if err := validTargetPath(rel); err != nil {
			return nil, err
		}
		if err := rejectSymlinkParents(targetDir, rel); err != nil {
			return nil, err
		}
		abs := filepath.Join(targetDir, filepath.FromSlash(rel))
		info, err := os.Lstat(abs)
		if err == nil && !owned[rel] {
			if projectionFileMatches(abs, content) {
				continue
			}
			return nil, fmt.Errorf("refusing to overwrite %s: it exists and was not written by a previous projection", abs)
		}
		if err == nil && !info.Mode().IsRegular() {
			return nil, fmt.Errorf("refusing to overwrite %s: an owned path is not a regular file", abs)
		}
		if err != nil && !os.IsNotExist(err) {
			return nil, fmt.Errorf("inspect projection path %s: %w", abs, err)
		}
		effective[rel] = content
	}
	written := sortedPaths(effective)
	next := Projection{Layout: label, Bundle: origin, Files: written}
	raw, err := json.MarshalIndent(next, "", "  ")
	if err != nil {
		return nil, err
	}
	sidecarAbs := filepath.Join(targetDir, filepath.FromSlash(sidecarRel))
	snapshots, err := snapshotProjection(targetDir, previous.Files, written)
	if err != nil {
		return nil, err
	}
	sidecarSnapshot, err := snapshotFile(sidecarAbs)
	if err != nil {
		return nil, err
	}
	if err := writeProjection(targetDir, effective, previous.Files, append(raw, '\n'), writeFile); err != nil {
		rollbackErr := rollbackProjection(targetDir, snapshots, sidecarSnapshot)
		if rollbackErr != nil {
			return nil, fmt.Errorf("apply projection: %w; rollback failed: %v", err, rollbackErr)
		}
		return nil, fmt.Errorf("apply projection: %w; previous projection restored", err)
	}
	return &Result{Layout: label, Files: written}, nil
}

// projectionFileMatches recognizes an identical canonical file already at a
// load point, leaving the sidecar to track only Agent Compose-owned files.
func projectionFileMatches(path string, want []byte) bool {
	info, err := os.Stat(path)
	if err != nil || !info.Mode().IsRegular() {
		return false
	}
	got, err := os.ReadFile(path)
	return err == nil && bytes.Equal(got, want)
}

type fileSnapshot struct {
	Content []byte
	Mode    fs.FileMode
	Exists  bool
}

func snapshotProjection(targetDir string, previous, next []string) (map[string]fileSnapshot, error) {
	paths := map[string]bool{}
	for _, rel := range previous {
		paths[rel] = true
	}
	for _, rel := range next {
		paths[rel] = true
	}
	snapshots := make(map[string]fileSnapshot, len(paths))
	for rel := range paths {
		abs := filepath.Join(targetDir, filepath.FromSlash(rel))
		snapshot, err := snapshotFile(abs)
		if err != nil {
			return nil, fmt.Errorf("snapshot projection path %s: %w", abs, err)
		}
		snapshots[rel] = snapshot
	}
	return snapshots, nil
}

func snapshotFile(path string) (fileSnapshot, error) {
	info, err := os.Lstat(path)
	if os.IsNotExist(err) {
		return fileSnapshot{}, nil
	}
	if err != nil {
		return fileSnapshot{}, err
	}
	if !info.Mode().IsRegular() {
		return fileSnapshot{}, fmt.Errorf("path is not a regular file")
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		return fileSnapshot{}, err
	}
	return fileSnapshot{Content: raw, Mode: info.Mode().Perm(), Exists: true}, nil
}

func writeProjection(
	targetDir string,
	files map[string][]byte,
	previous []string,
	sidecarRaw []byte,
	writeFile func(string, []byte, fs.FileMode) error,
) error {
	for _, rel := range sortedPaths(files) {
		abs := filepath.Join(targetDir, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(abs), 0o755); err != nil {
			return err
		}
		if err := writeFile(abs, files[rel], 0o644); err != nil {
			return err
		}
		if err := os.Chmod(abs, 0o644); err != nil {
			return err
		}
	}
	for _, rel := range previous {
		if _, still := files[rel]; still {
			continue
		}
		abs := filepath.Join(targetDir, filepath.FromSlash(rel))
		if err := os.Remove(abs); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("remove stale projection %s: %w", abs, err)
		}
		pruneEmptyDirs(filepath.Dir(abs), targetDir)
	}
	sidecarAbs := filepath.Join(targetDir, filepath.FromSlash(sidecarRel))
	if err := os.MkdirAll(filepath.Dir(sidecarAbs), 0o755); err != nil {
		return err
	}
	if err := writeFile(sidecarAbs, sidecarRaw, 0o644); err != nil {
		return err
	}
	return os.Chmod(sidecarAbs, 0o644)
}

func rollbackProjection(targetDir string, snapshots map[string]fileSnapshot, sidecarSnapshot fileSnapshot) error {
	var rollbackErrors []error
	var absent []string
	for rel, snapshot := range snapshots {
		if !snapshot.Exists {
			absent = append(absent, rel)
		}
	}
	sort.Slice(absent, func(i, j int) bool { return len(absent[i]) > len(absent[j]) })
	for _, rel := range absent {
		abs := filepath.Join(targetDir, filepath.FromSlash(rel))
		if err := os.Remove(abs); err != nil && !os.IsNotExist(err) {
			rollbackErrors = append(rollbackErrors, fmt.Errorf("remove new path %s: %w", abs, err))
		}
		pruneEmptyDirs(filepath.Dir(abs), targetDir)
	}
	for rel, snapshot := range snapshots {
		if !snapshot.Exists {
			continue
		}
		abs := filepath.Join(targetDir, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(abs), 0o755); err != nil {
			rollbackErrors = append(rollbackErrors, fmt.Errorf("restore parent for %s: %w", abs, err))
			continue
		}
		if err := os.WriteFile(abs, snapshot.Content, snapshot.Mode); err != nil {
			rollbackErrors = append(rollbackErrors, fmt.Errorf("restore %s: %w", abs, err))
			continue
		}
		if err := os.Chmod(abs, snapshot.Mode); err != nil {
			rollbackErrors = append(rollbackErrors, fmt.Errorf("restore mode for %s: %w", abs, err))
		}
	}
	sidecarAbs := filepath.Join(targetDir, filepath.FromSlash(sidecarRel))
	if sidecarSnapshot.Exists {
		if err := os.MkdirAll(filepath.Dir(sidecarAbs), 0o755); err != nil {
			rollbackErrors = append(rollbackErrors, fmt.Errorf("restore sidecar parent: %w", err))
		} else if err := os.WriteFile(sidecarAbs, sidecarSnapshot.Content, sidecarSnapshot.Mode); err != nil {
			rollbackErrors = append(rollbackErrors, fmt.Errorf("restore sidecar: %w", err))
		} else if err := os.Chmod(sidecarAbs, sidecarSnapshot.Mode); err != nil {
			rollbackErrors = append(rollbackErrors, fmt.Errorf("restore sidecar mode: %w", err))
		}
	} else if err := os.Remove(sidecarAbs); err != nil && !os.IsNotExist(err) {
		rollbackErrors = append(rollbackErrors, fmt.Errorf("remove new sidecar: %w", err))
	}
	return errors.Join(rollbackErrors...)
}

func sortedPaths(files map[string][]byte) []string {
	paths := make([]string, 0, len(files))
	for rel := range files {
		paths = append(paths, rel)
	}
	sort.Strings(paths)
	return paths
}

func validTargetPath(rel string) error {
	if rel == "" || rel == "." || strings.Contains(rel, `\`) || !fs.ValidPath(rel) {
		return fmt.Errorf("projection path %q is not a safe relative path", rel)
	}
	if rel == ".agent-compose" || strings.HasPrefix(rel, ".agent-compose/") {
		return fmt.Errorf("projection path %q collides with projection metadata", rel)
	}
	for _, segment := range strings.Split(rel, "/") {
		if !portableTargetSegment(segment) {
			return fmt.Errorf("projection path %q is not portable", rel)
		}
	}
	return nil
}

func portableTargetSegment(value string) bool {
	if value == "" || value == "." || value == ".." ||
		strings.HasSuffix(value, ".") || strings.HasSuffix(value, " ") {
		return false
	}
	for _, character := range value {
		if character < 32 || strings.ContainsRune(`<>:"|?*`, character) {
			return false
		}
	}
	base := strings.ToUpper(strings.SplitN(value, ".", 2)[0])
	switch base {
	case "CON", "PRN", "AUX", "NUL",
		"COM1", "COM2", "COM3", "COM4", "COM5", "COM6", "COM7", "COM8", "COM9",
		"LPT1", "LPT2", "LPT3", "LPT4", "LPT5", "LPT6", "LPT7", "LPT8", "LPT9":
		return false
	}
	return true
}

func rejectSymlinkParents(targetDir, rel string) error {
	current := targetDir
	parent := filepath.Dir(filepath.FromSlash(rel))
	if parent == "." {
		return nil
	}
	for _, segment := range strings.Split(filepath.ToSlash(parent), "/") {
		current = filepath.Join(current, segment)
		info, err := os.Lstat(current)
		if os.IsNotExist(err) {
			return nil
		}
		if err != nil {
			return fmt.Errorf("inspect projection parent %s: %w", current, err)
		}
		if info.Mode()&fs.ModeSymlink != 0 {
			return fmt.Errorf("projection parent %s is a symlink", current)
		}
		if !info.IsDir() {
			return fmt.Errorf("projection parent %s is not a directory", current)
		}
	}
	return nil
}

func ensureSeparateTarget(bundleDir, targetDir string) error {
	bundleAbs, err := filepath.Abs(bundleDir)
	if err != nil {
		return err
	}
	targetAbs, err := filepath.Abs(targetDir)
	if err != nil {
		return err
	}
	inside, err := pathWithin(bundleAbs, targetAbs)
	if err != nil {
		return err
	}
	if inside {
		return fmt.Errorf("projection target %s must not be the bundle or a directory inside it", targetDir)
	}
	bundleResolved, bundleErr := filepath.EvalSymlinks(bundleAbs)
	targetResolved, targetErr := filepath.EvalSymlinks(targetAbs)
	if bundleErr == nil && targetErr == nil {
		inside, err := pathWithin(bundleResolved, targetResolved)
		if err != nil {
			return err
		}
		if inside {
			return fmt.Errorf("projection target %s resolves inside bundle %s", targetDir, bundleDir)
		}
	}
	return nil
}

func pathWithin(parent, child string) (bool, error) {
	rel, err := filepath.Rel(parent, child)
	if err != nil {
		if !strings.EqualFold(filepath.VolumeName(parent), filepath.VolumeName(child)) {
			return false, nil
		}
		return false, err
	}
	return rel == "." || (rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator))), nil
}
