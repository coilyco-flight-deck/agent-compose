package bundle

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"forgejo.coilysiren.me/coilyco-flight-deck/agent-compose/internal/resolver"
	"forgejo.coilysiren.me/coilyco-flight-deck/agent-compose/internal/schema"
)

// Verification is the stable, read-only result of checking a bundle.
type Verification struct {
	Manifest       *Manifest
	Files          int
	IdentitySource string
	IdentitySkill  string
}

// Verify checks the consumer contract, rejecting unsafe or incomplete trees
// and identity content that disagrees with the retained selection trace.
func Verify(dir string) (*Verification, error) {
	files, err := verifyTree(dir)
	if err != nil {
		return nil, err
	}
	manifest, err := ReadManifest(dir)
	if err != nil {
		return nil, err
	}
	if err := verifyManifest(dir, manifest); err != nil {
		return nil, err
	}
	trace, err := ReadTrace(dir)
	if err != nil {
		return nil, err
	}
	if err := verifyTraceProfiles(trace, manifest); err != nil {
		return nil, err
	}
	source, skill, err := selectedIdentity(trace)
	if err != nil {
		return nil, err
	}
	if !slices.Contains(manifest.Sources, source) {
		return nil, fmt.Errorf("selected identity source %q is absent from manifest sources", source)
	}
	if err := verifyIdentityTree(dir, source, skill); err != nil {
		return nil, err
	}
	return &Verification{
		Manifest:       manifest,
		Files:          files,
		IdentitySource: source,
		IdentitySkill:  skill,
	}, nil
}

func verifyTree(dir string) (int, error) {
	files := 0
	err := filepath.WalkDir(dir, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return fmt.Errorf("inspect bundle tree: %w", err)
		}
		rel, err := filepath.Rel(dir, path)
		if err != nil {
			return fmt.Errorf("resolve bundle entry %s: %w", path, err)
		}
		if rel != "." {
			if !safePortablePath(filepath.ToSlash(rel)) {
				return fmt.Errorf("bundle entry %s has a non-portable relative path", path)
			}
		}
		info, err := entry.Info()
		if err != nil {
			return fmt.Errorf("inspect bundle entry %s: %w", path, err)
		}
		mode := info.Mode()
		switch {
		case mode&fs.ModeSymlink != 0:
			return fmt.Errorf("bundle entry %s is a symlink", path)
		case mode.IsDir():
			return nil
		case mode.IsRegular():
			files++
			return nil
		default:
			return fmt.Errorf("bundle entry %s is not a regular file or directory", path)
		}
	})
	if err != nil {
		return 0, err
	}
	return files, nil
}

func verifyManifest(dir string, manifest *Manifest) error {
	if manifest.Role == "" || manifest.Personality == "" || manifest.Density == "" {
		return fmt.Errorf("bundle manifest must name role, personality, and density")
	}
	if len(manifest.Sources) == 0 {
		return fmt.Errorf("bundle manifest names no sources")
	}
	for _, source := range manifest.Sources {
		if source == "" {
			return fmt.Errorf("bundle manifest contains an empty source id")
		}
	}
	if err := requireRegularEntry(dir, "instructions", manifest.Delivery.Instructions); err != nil {
		return err
	}
	switch manifest.Delivery.Mode {
	case schema.DeliveryNativeSkills:
		if manifest.Delivery.CompiledContext != "" {
			return fmt.Errorf("native bundle unexpectedly names compiled_context")
		}
		return requireDirectoryEntry(dir, "skills_root", manifest.Delivery.SkillsRoot)
	case schema.DeliveryCompiled:
		if manifest.Delivery.SkillsRoot != "" {
			return fmt.Errorf("compiled bundle unexpectedly names skills_root")
		}
		return requireRegularEntry(dir, "compiled_context", manifest.Delivery.CompiledContext)
	default:
		return fmt.Errorf("bundle manifest has unknown delivery mode %q", manifest.Delivery.Mode)
	}
}

func selectedIdentity(trace *Trace) (string, string, error) {
	var selected []resolver.Decision
	for _, decision := range trace.Decisions {
		if decision.Kind == "skill" && decision.Outcome == resolver.OutcomeSelected {
			selected = append(selected, decision)
		}
	}
	if len(selected) != 1 {
		return "", "", fmt.Errorf("bundle trace must select exactly one identity skill, found %d", len(selected))
	}
	decision := selected[0]
	skill, ok := strings.CutPrefix(decision.Subject, "skill:")
	if !ok || !safeSegment(skill) || !safeSegment(decision.Source) {
		return "", "", fmt.Errorf("bundle trace selects an unsafe identity path %q from %q",
			decision.Subject, decision.Source)
	}
	return decision.Source, skill, nil
}

func verifyTraceProfiles(trace *Trace, manifest *Manifest) error {
	role := "role:" + manifest.Role
	personality := "personality:" + manifest.Personality
	found := map[string]bool{}
	selected := 0
	for _, decision := range trace.Decisions {
		if decision.Kind != "profile" || decision.Outcome != resolver.OutcomeSelected {
			continue
		}
		selected++
		found[decision.Subject] = true
	}
	if !found[role] {
		return fmt.Errorf("bundle trace does not select manifest profile %q", role)
	}
	if !found[personality] {
		return fmt.Errorf("bundle trace does not select manifest profile %q", personality)
	}
	if selected != 2 {
		return fmt.Errorf("bundle trace selects %d profiles, expected role and personality only", selected)
	}
	return nil
}

func verifyIdentityTree(dir, source, skill string) error {
	root := filepath.Join(dir, "content", "skills")
	sources, err := os.ReadDir(root)
	if err != nil {
		return fmt.Errorf("read canonical skills root: %w", err)
	}
	var found []string
	for _, sourceEntry := range sources {
		if !sourceEntry.IsDir() {
			return fmt.Errorf("canonical skills root contains non-directory %s", sourceEntry.Name())
		}
		skills, err := os.ReadDir(filepath.Join(root, sourceEntry.Name()))
		if err != nil {
			return fmt.Errorf("read identity source %s: %w", sourceEntry.Name(), err)
		}
		for _, skillEntry := range skills {
			if !skillEntry.IsDir() {
				return fmt.Errorf("identity source %s contains non-directory %s",
					sourceEntry.Name(), skillEntry.Name())
			}
			found = append(found, sourceEntry.Name()+"/"+skillEntry.Name())
		}
	}
	expected := source + "/" + skill
	if len(found) != 1 || found[0] != expected {
		return fmt.Errorf("bundle must contain exactly selected identity %q, found %q", expected, found)
	}
	skillDoc := filepath.Join(root, source, skill, "SKILL.md")
	info, err := os.Stat(skillDoc)
	if err != nil {
		return fmt.Errorf("selected identity %q has no SKILL.md: %w", expected, err)
	}
	if !info.Mode().IsRegular() {
		return fmt.Errorf("selected identity %q SKILL.md is not a regular file", expected)
	}
	return nil
}

func requireRegularEntry(dir, name, rel string) error {
	path, err := bundleEntry(dir, name, rel)
	if err != nil {
		return err
	}
	info, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("bundle %s entry %q: %w", name, rel, err)
	}
	if !info.Mode().IsRegular() {
		return fmt.Errorf("bundle %s entry %q is not a regular file", name, rel)
	}
	return nil
}

func requireDirectoryEntry(dir, name, rel string) error {
	path, err := bundleEntry(dir, name, rel)
	if err != nil {
		return err
	}
	info, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("bundle %s entry %q: %w", name, rel, err)
	}
	if !info.IsDir() {
		return fmt.Errorf("bundle %s entry %q is not a directory", name, rel)
	}
	return nil
}

func bundleEntry(dir, name, rel string) (string, error) {
	if rel == "" {
		return "", fmt.Errorf("bundle manifest names no %s entry point", name)
	}
	if !safePortablePath(rel) {
		return "", fmt.Errorf("bundle %s entry %q is not a safe relative path", name, rel)
	}
	return filepath.Join(dir, filepath.FromSlash(rel)), nil
}

func safeSegment(value string) bool {
	if value == "" || value == "." || value == ".." || strings.ContainsAny(value, `/\`) {
		return false
	}
	if strings.HasSuffix(value, ".") || strings.HasSuffix(value, " ") {
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

func safePortablePath(value string) bool {
	if value == "" || value == "." || strings.Contains(value, `\`) || !fs.ValidPath(value) {
		return false
	}
	for _, segment := range strings.Split(value, "/") {
		if !safeSegment(segment) {
			return false
		}
	}
	return true
}
