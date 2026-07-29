package bundle

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"forgejo.coilysiren.me/coilyco-flight-deck/agent-compose/internal/color"
	"forgejo.coilysiren.me/coilyco-flight-deck/agent-compose/internal/resolver"
	"forgejo.coilysiren.me/coilyco-flight-deck/agent-compose/internal/schema"
)

type Identity struct {
	Source string
	Skill  string
}

// Verification is the stable, read-only result of checking a bundle.
type Verification struct {
	Manifest   *Manifest
	Files      int
	Identities []Identity
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
	identities, err := selectedIdentities(trace)
	if err != nil {
		return nil, err
	}
	for _, identity := range identities {
		if !slices.Contains(manifest.Sources, identity.Source) {
			return nil, fmt.Errorf("selected identity source %q is absent from manifest sources", identity.Source)
		}
	}
	if err := verifyIdentityTree(dir, identities); err != nil {
		return nil, err
	}
	return &Verification{
		Manifest:   manifest,
		Files:      files,
		Identities: identities,
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
	if manifest.Role == "" || manifest.RoleSkill == "" ||
		manifest.RoleSkillSource == "" || manifest.RoleSkillDigest == "" ||
		len(manifest.Personalities) == 0 {
		return fmt.Errorf("bundle manifest must name role skill provenance and personalities")
	}
	if manifest.RoleSkill != "role-"+manifest.Role {
		return fmt.Errorf("bundle manifest role skill %q does not match role %q", manifest.RoleSkill, manifest.Role)
	}
	if manifest.ModelClass != schema.ModelClassFrontier &&
		manifest.ModelClass != schema.ModelClassLowContext {
		return fmt.Errorf("bundle manifest has unknown model class %q", manifest.ModelClass)
	}
	seenPersonalities := map[string]bool{}
	for _, personality := range manifest.Personalities {
		if !safeSegment(personality) {
			return fmt.Errorf("bundle manifest contains unsafe personality %q", personality)
		}
		if seenPersonalities[personality] {
			return fmt.Errorf("bundle manifest repeats personality %q", personality)
		}
		seenPersonalities[personality] = true
	}
	if err := color.Legible(manifest.Color); err != nil {
		return fmt.Errorf("bundle manifest color: %w", err)
	}
	if len(manifest.Sources) == 0 {
		return fmt.Errorf("bundle manifest names no sources")
	}
	for _, source := range manifest.Sources {
		if source == "" {
			return fmt.Errorf("bundle manifest contains an empty source id")
		}
	}
	if len(manifest.Content) == 0 {
		return fmt.Errorf("bundle manifest names no logical content")
	}
	seenContent := map[string]bool{}
	for _, entry := range manifest.Content {
		if entry.ID == "" || seenContent[entry.ID] {
			return fmt.Errorf("bundle manifest has empty or duplicate logical content id %q", entry.ID)
		}
		if len(entry.Digest) != 71 || !strings.HasPrefix(entry.Digest, "sha256:") {
			return fmt.Errorf("bundle manifest logical content %q has invalid digest", entry.ID)
		}
		seenContent[entry.ID] = true
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

func selectedIdentities(trace *Trace) ([]Identity, error) {
	var identities []Identity
	for _, decision := range trace.Decisions {
		if decision.Kind != "skill" || decision.Outcome != resolver.OutcomeSelected {
			continue
		}
		skill, ok := strings.CutPrefix(decision.Subject, "skill:")
		if !ok || !safeSegment(skill) || !safeSourceID(decision.Source) {
			return nil, fmt.Errorf("bundle trace selects an unsafe identity path %q from %q",
				decision.Subject, decision.Source)
		}
		identities = append(identities, Identity{Source: decision.Source, Skill: skill})
	}
	if len(identities) == 0 {
		return nil, fmt.Errorf("bundle trace selects no skills")
	}
	return identities, nil
}

func verifyTraceProfiles(trace *Trace, manifest *Manifest) error {
	role := "role:" + manifest.Role
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
	for _, personality := range manifest.Personalities {
		subject := "personality:" + personality
		if !found[subject] {
			return fmt.Errorf("bundle trace does not select manifest profile %q", subject)
		}
	}
	expected := 1 + len(manifest.Personalities)
	if selected != expected {
		return fmt.Errorf("bundle trace selects %d profiles, expected role and %d personalities",
			selected, len(manifest.Personalities))
	}
	return nil
}

func verifyIdentityTree(dir string, identities []Identity) error {
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
	expected := make([]string, 0, len(identities))
	for _, identity := range identities {
		expected = append(expected, sourceSegment(identity.Source)+"/"+identity.Skill)
	}
	slices.Sort(found)
	slices.Sort(expected)
	if !slices.Equal(found, expected) {
		return fmt.Errorf("bundle identity tree %q does not match selected identities %q", found, expected)
	}
	for _, identity := range identities {
		name := identity.Source + "/" + identity.Skill
		skillDoc := filepath.Join(root, sourceSegment(identity.Source), identity.Skill, "SKILL.md")
		info, err := os.Stat(skillDoc)
		if err != nil {
			return fmt.Errorf("selected identity %q has no SKILL.md: %w", name, err)
		}
		if !info.Mode().IsRegular() {
			return fmt.Errorf("selected identity %q SKILL.md is not a regular file", name)
		}
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

func safeSourceID(value string) bool {
	if value == "" || value == "." || value == ".." || strings.ContainsAny(value, `/\`) {
		return false
	}
	for _, character := range value {
		if character < 32 {
			return false
		}
	}
	return true
}

// sourceSegment encodes logical source ids into portable filesystem segments.
// The trace and manifest retain the original id, including namespace colons.
func sourceSegment(value string) string {
	var encoded strings.Builder
	const hex = "0123456789ABCDEF"
	for _, octet := range []byte(value) {
		if octet >= 'a' && octet <= 'z' ||
			octet >= 'A' && octet <= 'Z' ||
			octet >= '0' && octet <= '9' ||
			octet == '-' || octet == '_' || octet == '.' {
			encoded.WriteByte(octet)
			continue
		}
		encoded.WriteByte('%')
		encoded.WriteByte(hex[octet>>4])
		encoded.WriteByte(hex[octet&0x0f])
	}
	return encoded.String()
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
