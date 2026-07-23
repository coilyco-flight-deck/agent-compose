// Package skillmount links configured skill roots into harness-native skill
// directories while tracking exactly which links agent-compose owns.
package skillmount

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

const sidecarName = "skill-mounts.json"

type link struct {
	Destination string `json:"destination"`
	Name        string `json:"name"`
	Target      string `json:"target"`
}

type sidecar struct {
	Links []link `json:"links"`
}

// Result summarizes one convergence without exposing host-specific paths.
type Result struct {
	Linked  int
	Removed int
	Skipped int
}

func expand(path string) (string, error) {
	if path != "~" && !strings.HasPrefix(path, "~/") {
		return filepath.Abs(path)
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("resolve home for %q: %w", path, err)
	}
	return filepath.Abs(filepath.Join(home, strings.TrimPrefix(strings.TrimPrefix(path, "~"), "/")))
}

func linkKey(destination, name string) string {
	return filepath.Join(destination, name)
}

func discover(roots []string, loadPoints map[string]string) (map[string]string, []string, error) {
	skills := map[string]string{}
	for _, configured := range roots {
		root, err := expand(configured)
		if err != nil {
			return nil, nil, err
		}
		entries, err := os.ReadDir(root)
		if err != nil {
			return nil, nil, fmt.Errorf("read skill root %s: %w", root, err)
		}
		for _, entry := range entries {
			target := filepath.Join(root, entry.Name())
			info, err := os.Stat(target)
			if err != nil {
				return nil, nil, fmt.Errorf("inspect skill %s: %w", target, err)
			}
			if info.IsDir() {
				skills[entry.Name()] = target
			}
		}
	}

	destinations := make([]string, 0, len(loadPoints))
	for harness, configured := range loadPoints {
		destination, err := expand(configured)
		if err != nil {
			return nil, nil, fmt.Errorf("resolve %s skill load point: %w", harness, err)
		}
		destinations = append(destinations, destination)
	}
	sort.Strings(destinations)
	return skills, destinations, nil
}

func readSidecar(path string) (sidecar, error) {
	raw, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return sidecar{}, nil
	}
	if err != nil {
		return sidecar{}, err
	}
	var state sidecar
	if err := json.Unmarshal(raw, &state); err != nil {
		return sidecar{}, fmt.Errorf("parse %s: %w", path, err)
	}
	return state, nil
}

func ownedLinkMatches(item link) bool {
	target, err := os.Readlink(linkKey(item.Destination, item.Name))
	return err == nil && target == item.Target
}

func writeSidecar(path string, state sidecar) error {
	raw, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return err
	}
	raw = append(raw, '\n')
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	temp, err := os.CreateTemp(filepath.Dir(path), ".skill-mounts-*.tmp")
	if err != nil {
		return err
	}
	tempPath := temp.Name()
	defer os.Remove(tempPath)
	if _, err := temp.Write(raw); err != nil {
		temp.Close()
		return err
	}
	if err := temp.Close(); err != nil {
		return err
	}
	return os.Rename(tempPath, path)
}

// Apply converges every configured skill root into every configured load point.
// Later roots win duplicate names. Unowned destination entries always win.
func Apply(roots []string, loadPoints map[string]string, stateDir string) (Result, error) {
	if len(roots) == 0 && len(loadPoints) == 0 {
		return Result{}, nil
	}
	if len(roots) == 0 || len(loadPoints) == 0 {
		return Result{}, fmt.Errorf("skill_roots and skill_load_points must be configured together")
	}
	skills, destinations, err := discover(roots, loadPoints)
	if err != nil {
		return Result{}, err
	}
	sidecarPath := filepath.Join(stateDir, sidecarName)
	previous, err := readSidecar(sidecarPath)
	if err != nil {
		return Result{}, err
	}
	owned := map[string]link{}
	for _, item := range previous.Links {
		owned[linkKey(item.Destination, item.Name)] = item
	}

	desired := map[string]link{}
	names := make([]string, 0, len(skills))
	for name := range skills {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, destination := range destinations {
		for _, name := range names {
			item := link{Destination: destination, Name: name, Target: skills[name]}
			desired[linkKey(destination, name)] = item
		}
	}

	result := Result{}
	for key, item := range owned {
		next, wanted := desired[key]
		if wanted && next.Target == item.Target {
			continue
		}
		if ownedLinkMatches(item) {
			if err := os.Remove(key); err != nil {
				return result, fmt.Errorf("remove stale skill link %s: %w", key, err)
			}
			result.Removed++
		}
	}

	current := sidecar{}
	for _, destination := range destinations {
		if err := os.MkdirAll(destination, 0o755); err != nil {
			return result, fmt.Errorf("create skill load point %s: %w", destination, err)
		}
		for _, name := range names {
			item := desired[linkKey(destination, name)]
			path := linkKey(destination, name)
			if ownedLinkMatches(item) {
				current.Links = append(current.Links, item)
				continue
			}
			if _, err := os.Lstat(path); err == nil {
				result.Skipped++
				continue
			} else if !errors.Is(err, os.ErrNotExist) {
				return result, fmt.Errorf("inspect skill load point %s: %w", path, err)
			}
			if err := os.Symlink(item.Target, path); err != nil {
				return result, fmt.Errorf("link skill %s: %w", path, err)
			}
			current.Links = append(current.Links, item)
			result.Linked++
		}
	}
	if err := writeSidecar(sidecarPath, current); err != nil {
		return result, fmt.Errorf("write skill mount ownership: %w", err)
	}
	return result, nil
}
