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

type eligibility struct {
	Defaults  []string            `json:"defaults"`
	Harnesses map[string][]string `json:"harnesses"`
}

// Catalog is a verified skill root projected to every configured load point.
// Catalogs apply after local repositories in declaration order.
type Catalog struct {
	Path string
}

// Result summarizes one convergence. Warnings retain the affected path so a
// host operator can repair unavailable catalog entries.
type Result struct {
	Linked     int
	Removed    int
	Skipped    int
	Verified   int
	Managed    int
	LoadPoints int
	Warnings   []string
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

func orderedRepos(manifest eligibility, harness string) []string {
	seen := map[string]bool{}
	repos := make([]string, 0, len(manifest.Defaults)+len(manifest.Harnesses[harness]))
	for _, repo := range manifest.Defaults {
		repos = append(repos, repo)
		seen[repo] = true
	}
	extra := append([]string(nil), manifest.Harnesses[harness]...)
	sort.Strings(extra)
	for _, repo := range extra {
		if !seen[repo] {
			repos = append(repos, repo)
			seen[repo] = true
		}
	}
	return repos
}

func discover(manifestPath string, loadPoints map[string]string, catalogs []Catalog) (map[string]link, []string, error) {
	raw, err := os.ReadFile(manifestPath)
	if err != nil {
		return nil, nil, fmt.Errorf("read mount eligibility %s: %w", manifestPath, err)
	}
	var manifest eligibility
	if err := json.Unmarshal(raw, &manifest); err != nil {
		return nil, nil, fmt.Errorf("parse mount eligibility %s: %w", manifestPath, err)
	}
	desired := map[string]link{}
	destinations := map[string]string{}
	warnings := map[string]bool{}
	harnesses := make([]string, 0, len(loadPoints))
	for harness := range loadPoints {
		harnesses = append(harnesses, harness)
	}
	sort.Strings(harnesses)
	for _, harness := range harnesses {
		destination, err := expand(loadPoints[harness])
		if err != nil {
			return nil, nil, fmt.Errorf("resolve %s skill load point: %w", harness, err)
		}
		if owner, exists := destinations[destination]; exists {
			return nil, nil, fmt.Errorf("skill load point %s is shared by %s and %s", destination, owner, harness)
		}
		destinations[destination] = harness
		skills := map[string]string{}
		addRoot := func(root string) error {
			info, err := os.Stat(root)
			if errors.Is(err, os.ErrNotExist) {
				return nil
			}
			if err != nil {
				return fmt.Errorf("inspect eligible skill root %s: %w", root, err)
			}
			if !info.IsDir() {
				return nil
			}
			entries, err := os.ReadDir(root)
			if err != nil {
				return fmt.Errorf("read eligible skill root %s: %w", root, err)
			}
			for _, entry := range entries {
				target := filepath.Join(root, entry.Name())
				info, err := os.Stat(target)
				if errors.Is(err, os.ErrNotExist) {
					warnings[fmt.Sprintf("inspect skill %s: %v", target, err)] = true
					continue
				}
				if err != nil {
					return fmt.Errorf("inspect skill %s: %w", target, err)
				}
				if info.IsDir() {
					skills[entry.Name()] = target
				}
			}
			return nil
		}
		for _, repo := range orderedRepos(manifest, harness) {
			if err := addRoot(filepath.Join(repo, ".agents", "skills")); err != nil {
				return nil, nil, err
			}
		}
		for _, catalog := range catalogs {
			if err := addRoot(catalog.Path); err != nil {
				return nil, nil, err
			}
		}
		for name, target := range skills {
			item := link{Destination: destination, Name: name, Target: target}
			desired[linkKey(destination, name)] = item
		}
	}
	reported := make([]string, 0, len(warnings))
	for warning := range warnings {
		reported = append(reported, warning)
	}
	sort.Strings(reported)
	return desired, reported, nil
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

// Apply converges eligible repository skill roots into configured load points.
// Later eligible repositories win duplicate names. Unowned entries always win.
func Apply(manifestPath string, loadPoints map[string]string, stateDir string) (Result, error) {
	return ApplyWithCatalogs(manifestPath, loadPoints, stateDir, nil)
}

// ApplyWithCatalogs converges local eligible repositories plus verified
// additional catalogs into the configured native load points.
func ApplyWithCatalogs(
	manifestPath string,
	loadPoints map[string]string,
	stateDir string,
	catalogs []Catalog,
) (Result, error) {
	if len(loadPoints) == 0 {
		return Result{}, nil
	}
	desired, warnings, err := discover(manifestPath, loadPoints, catalogs)
	if err != nil {
		return Result{Warnings: warnings}, err
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

	result := Result{Managed: len(desired), LoadPoints: len(loadPoints), Warnings: warnings}
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
	keys := make([]string, 0, len(desired))
	for key := range desired {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, path := range keys {
		item := desired[path]
		if err := os.MkdirAll(item.Destination, 0o755); err != nil {
			return result, fmt.Errorf("create skill load point %s: %w", item.Destination, err)
		}
		if ownedLinkMatches(item) {
			current.Links = append(current.Links, item)
			result.Verified++
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
	if err := writeSidecar(sidecarPath, current); err != nil {
		return result, fmt.Errorf("write skill mount ownership: %w", err)
	}
	return result, nil
}
