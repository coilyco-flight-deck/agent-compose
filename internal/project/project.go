// Package project places bundle content at harness load points. This is the
// one layer allowed to know harness vocabulary; composition stays blind to it.
package project

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"forgejo.coilysiren.me/coilyco-flight-deck/agent-compose/internal/bundle"
	"forgejo.coilysiren.me/coilyco-flight-deck/agent-compose/internal/schema"
)

type LoadPoints struct {
	Instructions string
	SkillsDir    string
}

// Layout declares load points per delivery mode; a nil mode is unsupported
// by that harness and fails with a diagnostic.
type Layout struct {
	Native   *LoadPoints
	Compiled *LoadPoints
}

// Registry is the fixed v0.1 layout set. Layout names and load-point paths
// live here and nowhere else; conventions are recorded in docs/projection.md.
var Registry = map[string]Layout{
	"claude": {
		Native:   &LoadPoints{Instructions: "CLAUDE.md", SkillsDir: ".claude/skills"},
		Compiled: &LoadPoints{Instructions: "CLAUDE.md"},
	},
	"codex": {
		Native:   &LoadPoints{Instructions: "AGENTS.md", SkillsDir: ".agents/skills"},
		Compiled: &LoadPoints{Instructions: "AGENTS.md"},
	},
	"goose": {
		Native:   &LoadPoints{Instructions: ".goosehints", SkillsDir: ".agents/skills"},
		Compiled: &LoadPoints{Instructions: ".goosehints"},
	},
	"opencode": {
		Native:   &LoadPoints{Instructions: "AGENTS.md", SkillsDir: ".agents/skills"},
		Compiled: &LoadPoints{Instructions: "AGENTS.md"},
	},
}

const sidecarRel = ".agent-compose/projection.json"

type sidecar struct {
	Layout string   `json:"layout"`
	Bundle string   `json:"bundle"`
	Files  []string `json:"files"`
}

type Result struct {
	Layout string
	Files  []string
}

func Project(bundleDir, layoutName, targetDir string) (*Result, error) {
	layout, ok := Registry[layoutName]
	if !ok {
		return nil, fmt.Errorf("unknown layout %q; v0.1 layouts: %s", layoutName, strings.Join(registryNames(), ", "))
	}
	manifest, err := bundle.ReadManifest(bundleDir)
	if err != nil {
		return nil, err
	}
	var points *LoadPoints
	switch manifest.Delivery.Mode {
	case schema.DeliveryNativeSkills:
		points = layout.Native
	case schema.DeliveryCompiled:
		points = layout.Compiled
	}
	if points == nil {
		return nil, fmt.Errorf("layout %q does not support bundles that deliver %s",
			layoutName, manifest.Delivery.Mode)
	}

	plan, err := buildPlan(bundleDir, manifest, *points)
	if err != nil {
		return nil, err
	}
	previous := readSidecar(targetDir)
	owned := map[string]bool{}
	for _, rel := range previous.Files {
		owned[rel] = true
	}
	for rel := range plan {
		abs := filepath.Join(targetDir, filepath.FromSlash(rel))
		if _, err := os.Stat(abs); err == nil && !owned[rel] {
			return nil, fmt.Errorf("refusing to overwrite %s: it exists and was not written by a previous projection", abs)
		}
	}

	var written []string
	for rel, srcAbs := range plan {
		raw, err := os.ReadFile(srcAbs)
		if err != nil {
			return nil, err
		}
		abs := filepath.Join(targetDir, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(abs), 0o755); err != nil {
			return nil, err
		}
		if err := os.WriteFile(abs, raw, 0o644); err != nil {
			return nil, err
		}
		written = append(written, rel)
	}
	sort.Strings(written)

	for _, rel := range previous.Files {
		if _, still := plan[rel]; still {
			continue
		}
		abs := filepath.Join(targetDir, filepath.FromSlash(rel))
		if err := os.Remove(abs); err != nil && !os.IsNotExist(err) {
			return nil, fmt.Errorf("remove stale projection %s: %w", abs, err)
		}
		pruneEmptyDirs(filepath.Dir(abs), targetDir)
	}

	next := sidecar{Layout: layoutName, Bundle: bundleDir, Files: written}
	raw, err := json.MarshalIndent(next, "", "  ")
	if err != nil {
		return nil, err
	}
	sidecarAbs := filepath.Join(targetDir, filepath.FromSlash(sidecarRel))
	if err := os.MkdirAll(filepath.Dir(sidecarAbs), 0o755); err != nil {
		return nil, err
	}
	if err := os.WriteFile(sidecarAbs, append(raw, '\n'), 0o644); err != nil {
		return nil, err
	}
	return &Result{Layout: layoutName, Files: written}, nil
}

// buildPlan maps target-relative slash paths to absolute bundle sources.
func buildPlan(bundleDir string, manifest *bundle.Manifest, points LoadPoints) (map[string]string, error) {
	plan := map[string]string{}
	switch manifest.Delivery.Mode {
	case schema.DeliveryCompiled:
		if manifest.Delivery.CompiledContext == "" {
			return nil, fmt.Errorf("bundle manifest names no compiled_context entry point")
		}
		plan[points.Instructions] = filepath.Join(bundleDir, filepath.FromSlash(manifest.Delivery.CompiledContext))
	case schema.DeliveryNativeSkills:
		if manifest.Delivery.Instructions == "" || manifest.Delivery.SkillsRoot == "" {
			return nil, fmt.Errorf("bundle manifest names no native entry points")
		}
		plan[points.Instructions] = filepath.Join(bundleDir, filepath.FromSlash(manifest.Delivery.Instructions))
		skillsRoot := filepath.Join(bundleDir, filepath.FromSlash(manifest.Delivery.SkillsRoot))
		sourceDirs, err := os.ReadDir(skillsRoot)
		if err != nil {
			return nil, err
		}
		for _, sourceDir := range sourceDirs {
			if !sourceDir.IsDir() {
				continue
			}
			base := filepath.Join(skillsRoot, sourceDir.Name())
			err := filepath.WalkDir(base, func(p string, d fs.DirEntry, err error) error {
				if err != nil || d.IsDir() {
					return err
				}
				rel, err := filepath.Rel(base, p)
				if err != nil {
					return err
				}
				plan[points.SkillsDir+"/"+filepath.ToSlash(rel)] = p
				return nil
			})
			if err != nil {
				return nil, err
			}
		}
	}
	return plan, nil
}

func readSidecar(targetDir string) sidecar {
	var s sidecar
	raw, err := os.ReadFile(filepath.Join(targetDir, filepath.FromSlash(sidecarRel)))
	if err != nil {
		return s
	}
	if json.Unmarshal(raw, &s) != nil {
		return sidecar{}
	}
	return s
}

func pruneEmptyDirs(dir, stop string) {
	for {
		if dir == stop || len(dir) <= len(stop) {
			return
		}
		if os.Remove(dir) != nil {
			return
		}
		dir = filepath.Dir(dir)
	}
}

func registryNames() []string {
	names := make([]string, 0, len(Registry))
	for name := range Registry {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}
