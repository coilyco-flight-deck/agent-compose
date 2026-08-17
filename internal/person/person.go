// Package person embeds the canonical public-safe roles, agent seats,
// personality definitions, and invariant.
package person

import (
	"bytes"
	"crypto/sha256"
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"strings"
	"testing/fstest"
	"unicode"

	kdl "github.com/calico32/kdl-go"
	"golang.org/x/text/unicode/norm"

	"forgejo.coilysiren.me/coilyco-flight-deck/agent-compose/internal/color"
	"forgejo.coilysiren.me/coilyco-flight-deck/agent-compose/internal/schema"
)

//go:embed person.kdl data
var embedded embed.FS

const (
	minRoleSkillBodyWords = 140
	maxRoleSkillBodyWords = 400
)

// Personality and boundary prose carry their own bounds. A floor keeps an entry
// from thinning into a label. See docs/ai-engineer.md.
const (
	minPersonalitySkillBodyWords = 120
	maxPersonalitySkillBodyWords = 320
	minBoundarySideWords         = 80
)

// maxBoundarySkillBodyWords bounds each side of a boundary separately from the
// role charter that declares it. See docs/role-boundaries.md.
const maxBoundarySkillBodyWords = 400

// adjacentsPerRole fixes the out-degree of the role adjacency graph, so the
// roster chooses its sharpest confusions. See docs/role-boundaries.md.
const adjacentsPerRole = 2

// Both sides of a boundary live in one body under conditional headings, so the
// reader self-selects. See docs/ownership.md.
const (
	boundaryOwnHeading   = "## If you own this boundary"
	boundaryDeferHeading = "## If you defer this boundary"
)

var personSections = []struct {
	directory string
	node      string
}{
	{directory: "roles", node: "role"},
	{directory: "boundaries", node: "boundary"},
	{directory: "personalities", node: "personality"},
	{directory: "inspirations", node: "inspiration"},
}

var librarySections = []struct {
	directory string
	node      string
}{
	{directory: "boundaries", node: "boundary"},
	{directory: "personalities", node: "personality"},
	{directory: "inspirations", node: "inspiration"},
}

// Seat is one named agent identity within a role. The harness joins the
// launcher's own catalog, while the name remains opaque here.
type Seat struct {
	// Key is the stable profile-owned seat selector. Harness remains populated
	// for legacy agent entries and v1 result compatibility.
	Key      string `json:"key,omitempty" yaml:"key,omitempty"`
	Harness  string `json:"harness" yaml:"harness"`
	Name     string `json:"name" yaml:"name"`
	Pronouns string `json:"pronouns" yaml:"pronouns"`
	Channel  string `json:"channel,omitempty" yaml:"channel,omitempty"`
	Tier     string `json:"tier,omitempty" yaml:"tier,omitempty"`
}

// AgentIdentity is the role-owned name and pronoun pair shared by every seat.
// Legacy external packages may continue to own identity on individual seats.
type AgentIdentity struct {
	Name     string `json:"name" yaml:"name"`
	Pronouns string `json:"pronouns" yaml:"pronouns"`
}

// InspirationRef records why one role or personality cites a catalog entry.
// Inspirations are acknowledgements and evidence, not identities to imitate.
type InspirationRef struct {
	ID  string `json:"id"`
	Fit string `json:"fit"`
}

type Role struct {
	DisplayName         string         `json:"display_name"`
	Purpose             string         `json:"purpose"`
	Skill               string         `json:"skill"`
	SkillSource         string         `json:"skill_source"`
	SkillDigest         string         `json:"skill_digest"`
	Methods             []string       `json:"methods,omitempty"`
	Boundaries          []string       `json:"boundaries,omitempty"`
	Adjacents           []Adjacent     `json:"adjacents,omitempty"`
	Briefing            string         `json:"briefing"`
	Personalities       []string       `json:"personalities"`
	Identity            *AgentIdentity `json:"identity,omitempty"`
	Seats               []Seat         `json:"seats"`
	Inspiration         InspirationRef `json:"inspiration,omitempty"`
	SupportedModelTiers []string       `json:"supported_model_tiers,omitempty"`
	CopyContract        *CopyContract  `json:"copy_contract,omitempty"`
}

// Adjacent names one role whose work this role most risks absorbing. The reason
// is generator input rather than commentary. See docs/role-boundaries.md.
type Adjacent struct {
	Role   string `json:"role"`
	Reason string `json:"reason"`
}

// SupportsModelTier keeps packages authored before the tier axis compatible.
// First-party roles declare their supported tiers explicitly.
func (r Role) SupportsModelTier(modelTier string) bool {
	if len(r.SupportedModelTiers) == 0 {
		return true
	}
	for _, supported := range r.SupportedModelTiers {
		if modelTier == supported {
			return true
		}
	}
	return false
}

type CopyContract struct {
	Scope  string     `json:"scope"`
	Rules  []CopyRule `json:"rules"`
	Source string     `json:"source"`
	Digest string     `json:"digest"`
}

type CopyRule struct {
	Forbid string `json:"forbid"`
	Prefer string `json:"prefer"`
}

// Emblem gives renderers equivalent plain-text, rich-text, and compact marks.
type Emblem struct {
	Name  string `json:"name"`
	Emoji string `json:"emoji"`
	Glyph string `json:"glyph"`
}

// Form is a renderer-neutral procedural shape language. Renderers decide how
// to draw it while retaining the same silhouette, geometry, and motion cues.
type Form struct {
	Silhouette string `json:"silhouette"`
	Geometry   string `json:"geometry"`
	Motion     string `json:"motion"`
}

// SoundMark describes a short identity cue without prescribing audio files or
// playback behavior.
type SoundMark struct {
	Timbre  string `json:"timbre"`
	Contour string `json:"contour"`
	Pulse   string `json:"pulse"`
}

// Boundary binds one shared doctrine body that any number of roles may activate.
// Its optional owner is described in docs/ownership.md.
type Boundary struct {
	Skill   string `json:"skill"`
	Summary string `json:"summary"`
	Owner   string `json:"owner,omitempty"`
	Source  string `json:"source,omitempty"`
	Digest  string `json:"digest,omitempty"`
}

// Personality binds the definition, visual and sensory identity primitives,
// favorite color, and credited inspiration for one canonical personality.
type Personality struct {
	Skill       string         `json:"skill"`
	Color       string         `json:"color"`
	Motif       string         `json:"motif"`
	Emblem      Emblem         `json:"emblem"`
	Form        Form           `json:"form"`
	SoundMark   SoundMark      `json:"sound_mark"`
	Inspiration InspirationRef `json:"inspiration,omitempty"`
	Aliases     []string       `json:"aliases,omitempty"`
	Verbs       []string       `json:"verbs,omitempty"`
}

// Selector returns the stable key used by all new commands and artifacts.
func (s Seat) Selector() string {
	if s.Key != "" {
		return s.Key
	}
	return s.Harness
}

// NormalizeCue turns a user-facing personality cue into a stable lookup form.
func NormalizeCue(cue string) (string, error) {
	cue = strings.TrimSpace(norm.NFKC.String(cue))
	if cue == "" {
		return "", fmt.Errorf("cue is empty")
	}
	var out strings.Builder
	separator := false
	for _, r := range strings.ToLower(cue) {
		if unicode.IsControl(r) {
			return "", fmt.Errorf("cue contains a control character")
		}
		if unicode.IsSpace(r) || r == '_' || r == '-' {
			separator = out.Len() > 0
			continue
		}
		if separator {
			out.WriteByte('-')
			separator = false
		}
		out.WriteRune(r)
	}
	if separator || out.Len() == 0 {
		return "", fmt.Errorf("cue has a leading or trailing separator")
	}
	return out.String(), nil
}

// LookupCue preserves catalogue order and never hides an ambiguous alias.
// A canonical slug is an exact match and therefore wins over aliases.
func (p *Person) LookupCue(cue string) ([]string, error) {
	normalized, err := NormalizeCue(cue)
	if err != nil {
		return nil, err
	}
	for _, name := range p.personalityOrder() {
		slug, err := NormalizeCue(name)
		if err != nil {
			return nil, fmt.Errorf("normalize personality slug %q: %w", name, err)
		}
		if slug == normalized {
			return []string{name}, nil
		}
	}
	var matches []string
	for _, name := range p.personalityOrder() {
		for _, alias := range p.Personalities[name].Aliases {
			normalizedAlias, err := NormalizeCue(alias)
			if err != nil {
				return nil, fmt.Errorf("normalize alias %q for personality %q: %w", alias, name, err)
			}
			if normalizedAlias == normalized {
				matches = append(matches, name)
				break
			}
		}
	}
	return matches, nil
}

func (p *Person) personalityOrder() []string {
	if len(p.PersonalityOrder) == len(p.Personalities) {
		return append([]string(nil), p.PersonalityOrder...)
	}
	names := make([]string, 0, len(p.Personalities))
	for name := range p.Personalities {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

func (p *Person) boundaryOrder() []string {
	if len(p.BoundaryOrder) == len(p.Boundaries) {
		return append([]string(nil), p.BoundaryOrder...)
	}
	names := make([]string, 0, len(p.Boundaries))
	for name := range p.Boundaries {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

func (p *Person) roleOrder() []string {
	if len(p.RoleOrder) == len(p.Roles) {
		return append([]string(nil), p.RoleOrder...)
	}
	names := make([]string, 0, len(p.Roles))
	for name := range p.Roles {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

var expressionVocabulary = [...]string{
	"available",
	"listening",
	"thinking",
	"acting",
	"waiting-for-human",
	"blocked",
	"completed",
	"failed",
	"offline",
}

// ExpressionVocabulary returns the complete stable renderer state vocabulary.
func ExpressionVocabulary() []string {
	return append([]string(nil), expressionVocabulary[:]...)
}

// Appearance is one substantive public speaking record selected as evidence
// for an inspiration's assigned role, personality, or impact mode.
type Appearance struct {
	ID        string   `json:"id"`
	Title     string   `json:"title"`
	Event     string   `json:"event"`
	Year      string   `json:"year"`
	Format    string   `json:"format"`
	Summary   string   `json:"summary"`
	Citations []string `json:"citations"`
}

// Inspiration is one credited human influence. Citation keys refer to public
// evidence recorded on the catalogue's owning issue.
type Inspiration struct {
	Name            string     `json:"name"`
	Achievement     string     `json:"achievement"`
	ImpactMode      string     `json:"impact_mode"`
	ImpactFit       string     `json:"impact_fit"`
	ProfileCitation string     `json:"profile_citation"`
	Appearance      Appearance `json:"appearance"`
}

type Person struct {
	ProviderKind         string                 `json:"provider_kind"`
	Name                 string                 `json:"person"`
	Roles                map[string]Role        `json:"roles"`
	RoleOrder            []string               `json:"role_order"`
	Boundaries           map[string]Boundary    `json:"boundaries,omitempty"`
	BoundaryOrder        []string               `json:"boundary_order,omitempty"`
	Personalities        map[string]Personality `json:"personalities"`
	PersonalityOrder     []string               `json:"personality_order"`
	Inspirations         map[string]Inspiration `json:"inspirations"`
	InspirationOrder     []string               `json:"inspiration_order"`
	Raw                  []byte                 `json:"-"`
	Libraries            map[string]string      `json:"-"`
	PersonalityLibraries map[string]string      `json:"-"`
	roleSkills           map[string][]byte
	roleMethods          map[string]map[string][]byte
	boundarySkills       map[string][]byte
	source               fs.FS
}

// ProviderID returns the package identity. Legacy in-memory fixtures default
// to person so external package callers retain their v1 identity.
func (p *Person) ProviderID() string {
	return p.providerKind() + ":" + p.Name
}

func (p *Person) localSourceID() string {
	return p.ProviderID() + ":local"
}

func (p *Person) providerKind() string {
	if p.ProviderKind == "" {
		return "person"
	}
	return p.ProviderKind
}

// RoleDisplayName returns the authored product label for a role.
func (p *Person) RoleDisplayName(roleName string) string {
	if role, ok := p.Roles[roleName]; ok && role.DisplayName != "" {
		return role.DisplayName
	}
	return displaySlug(roleName)
}

// Load returns the shipped roster:core package.
func Load() (*Person, error) {
	source, _, err := dataLayout(embedded, "embedded core roster")
	if err != nil {
		return nil, err
	}
	p, err := loadSource(source, "embedded core roster")
	if err != nil {
		return nil, err
	}
	if err := validateRosterProseFloors(source, p); err != nil {
		return nil, fmt.Errorf("embedded core roster: %w", err)
	}
	if err := validateResolvedPerson(p); err != nil {
		return nil, err
	}
	if err := validateNoUnusedPersonalities(p); err != nil {
		return nil, err
	}
	if err := validateNoUnusedBoundaries(p); err != nil {
		return nil, err
	}
	if err := validateBoundaryOwners(p); err != nil {
		return nil, err
	}
	if err := validateRoleAdjacents(p); err != nil {
		return nil, err
	}
	if err := validateCorePersonalityMelds(p); err != nil {
		return nil, err
	}
	return p, nil
}

// LoadDirectory reads one complete external person package using the default layout.
// An external package replaces the embedded package rather than extending it.
func LoadDirectory(root string) (*Person, error) {
	return LoadDirectoryWithLibraries(root)
}

// LoadDirectoryWithLibraries loads a compatible profile root, its lexical
// package-local libraries, and explicitly admitted local library roots.
func LoadDirectoryWithLibraries(root string, libraries ...string) (*Person, error) {
	if strings.TrimSpace(root) == "" {
		return nil, fmt.Errorf("external person source path is empty")
	}
	absolute, err := filepath.Abs(root)
	if err != nil {
		return nil, fmt.Errorf("resolve external person source %s: %w", root, err)
	}
	info, err := os.Stat(absolute)
	if err != nil {
		return nil, fmt.Errorf("inspect external person source %s: %w", absolute, err)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("external person source %s is not a directory", absolute)
	}
	if err := filepath.WalkDir(absolute, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.Type()&fs.ModeSymlink != 0 {
			return fmt.Errorf("external person source contains symlink %s", path)
		}
		return nil
	}); err != nil {
		return nil, fmt.Errorf("inspect external person source tree %s: %w", absolute, err)
	}
	p, err := loadSource(os.DirFS(absolute), "external person source "+absolute)
	if err != nil {
		return nil, err
	}
	local, err := discoverLibraries(absolute)
	if err != nil {
		return nil, err
	}
	return mergeLibraries(p, append(local, libraries...))
}

func discoverLibraries(root string) ([]string, error) {
	dir := filepath.Join(root, "libraries")
	entries, err := os.ReadDir(dir)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("read profile libraries %s: %w", dir, err)
	}
	var paths []string
	for _, entry := range entries {
		if !entry.IsDir() {
			return nil, fmt.Errorf("profile libraries has non-directory entry %q", entry.Name())
		}
		paths = append(paths, filepath.Join(dir, entry.Name()))
	}
	sort.Strings(paths)
	return paths, nil
}

func mergeLibraries(p *Person, roots []string) (*Person, error) {
	if len(roots) == 0 {
		return p, validateResolvedPerson(p)
	}
	if p.Libraries == nil {
		p.Libraries = map[string]string{p.localSourceID(): "profile-local"}
	}
	if p.PersonalityLibraries == nil {
		p.PersonalityLibraries = map[string]string{}
		for name := range p.Personalities {
			p.PersonalityLibraries[name] = p.localSourceID()
		}
	}
	overlay, err := definitionOverlay(p.source, p.Personalities, p.Boundaries)
	if err != nil {
		return nil, err
	}
	for _, root := range roots {
		library, id, librarySource, err := loadLibrary(root)
		if err != nil {
			return nil, err
		}
		if err := mergeLoadedLibraryWithOverlay(p, overlay, library, id, librarySource); err != nil {
			return nil, err
		}
	}
	p.source = overlay
	return p, validateResolvedPerson(p)
}

func mergeLoadedLibrary(p *Person, library *Person, id string, source fs.FS) error {
	overlay, err := definitionOverlay(p.source, p.Personalities, p.Boundaries)
	if err != nil {
		return err
	}
	if err := mergeLoadedLibraryWithOverlay(p, overlay, library, id, source); err != nil {
		return err
	}
	p.source = overlay
	return nil
}

func mergeLoadedLibraryWithOverlay(p *Person, overlay fstest.MapFS, library *Person, id string, librarySource fs.FS) error {
	if p.Libraries == nil {
		p.Libraries = map[string]string{p.localSourceID(): "profile-local"}
	}
	if p.PersonalityLibraries == nil {
		p.PersonalityLibraries = map[string]string{}
	}
	if _, exists := p.Libraries[id]; exists {
		return fmt.Errorf("personality library %q is admitted more than once", id)
	}
	for _, name := range library.personalityOrder() {
		binding := library.Personalities[name]
		if existing, exists := p.Personalities[name]; exists {
			if !equalPersonality(existing, binding) {
				return fmt.Errorf("personality %q conflicts between profile libraries", name)
			}
			continue
		}
		for otherName, other := range p.Personalities {
			if other.Skill == binding.Skill && otherName != name {
				return fmt.Errorf("personality skill %q conflicts between %q and %q", binding.Skill, otherName, name)
			}
		}
		p.Personalities[name] = binding
		p.PersonalityLibraries[name] = id
		p.PersonalityOrder = append(p.PersonalityOrder, name)
	}
	for _, name := range library.boundaryOrder() {
		binding := library.Boundaries[name]
		if existing, exists := p.Boundaries[name]; exists {
			if existing.Skill != binding.Skill || existing.Summary != binding.Summary {
				return fmt.Errorf("boundary %q conflicts between profile libraries", name)
			}
			continue
		}
		for otherName, other := range p.Boundaries {
			if other.Skill == binding.Skill && otherName != name {
				return fmt.Errorf("boundary skill %q conflicts between %q and %q", binding.Skill, otherName, name)
			}
		}
		p.Boundaries[name] = binding
		p.BoundaryOrder = append(p.BoundaryOrder, name)
		if raw, ok := library.boundarySkills[name]; ok {
			p.boundarySkills[name] = append([]byte(nil), raw...)
		}
	}
	for _, name := range library.InspirationOrder {
		inspiration := library.Inspirations[name]
		if existing, exists := p.Inspirations[name]; exists && fmt.Sprintf("%#v", existing) != fmt.Sprintf("%#v", inspiration) {
			return fmt.Errorf("inspiration %q conflicts between profile libraries", name)
		}
		p.Inspirations[name] = inspiration
		p.InspirationOrder = append(p.InspirationOrder, name)
	}
	if err := appendDefinitions(overlay, librarySource, library.Personalities, library.Boundaries); err != nil {
		return err
	}
	p.Libraries[id] = "admitted-local"
	return nil
}

func validateResolvedPerson(p *Person) error {
	for _, roleName := range p.roleOrder() {
		for _, personalityName := range p.Roles[roleName].Personalities {
			if _, ok := p.Personalities[personalityName]; !ok {
				return fmt.Errorf("role %q: personality %q has no catalog binding", roleName, personalityName)
			}
		}
		for _, boundaryName := range p.Roles[roleName].Boundaries {
			if _, ok := p.Boundaries[boundaryName]; !ok {
				return fmt.Errorf("role %q: boundary %q has no catalog binding", roleName, boundaryName)
			}
		}
		ref := p.Roles[roleName].Inspiration.ID
		if ref != "" {
			if _, ok := p.Inspirations[ref]; !ok {
				return fmt.Errorf("role %q: inspiration %q has no catalog entry", roleName, ref)
			}
		}
	}
	return nil
}

// validateNoUnusedPersonalities derives garbage collection from the effective
// Core Roster graph without maintaining a second inventory.
func validateNoUnusedPersonalities(p *Person) error {
	used := map[string]bool{}
	for _, roleName := range p.RoleOrder {
		for _, personalityName := range p.Roles[roleName].Personalities {
			used[personalityName] = true
		}
	}
	var unused []string
	for _, personalityName := range p.personalityOrder() {
		if !used[personalityName] {
			unused = append(unused, personalityName)
		}
	}
	if len(unused) != 0 {
		return fmt.Errorf("core roster has unused personalities: %s", strings.Join(unused, ", "))
	}
	return nil
}

// validateNoUnusedBoundaries keeps shared doctrine anchored to at least one role, so
// a boundary cannot linger after the last role drops its reference.
func validateNoUnusedBoundaries(p *Person) error {
	used := map[string]bool{}
	for _, roleName := range p.RoleOrder {
		for _, boundaryName := range p.Roles[roleName].Boundaries {
			used[boundaryName] = true
		}
	}
	var unused []string
	for _, boundaryName := range p.boundaryOrder() {
		if !used[boundaryName] {
			unused = append(unused, boundaryName)
		}
	}
	if len(unused) != 0 {
		return fmt.Errorf("core roster has unused boundaries: %s", strings.Join(unused, ", "))
	}
	return nil
}

// validateBoundaryBodySides requires both sides and bounds each separately, so
// one long half cannot crowd out the other.
func validateBoundaryBodySides(boundaryName, body string) error {
	own := strings.Index(body, boundaryOwnHeading)
	defer_ := strings.Index(body, boundaryDeferHeading)
	if own < 0 || defer_ < 0 {
		return fmt.Errorf(
			"boundary %q skill body needs both %q and %q sections",
			boundaryName, boundaryOwnHeading, boundaryDeferHeading,
		)
	}
	if own > defer_ {
		return fmt.Errorf("boundary %q skill body states the defer side before the own side", boundaryName)
	}
	for label, section := range map[string]string{
		"own":   body[own:defer_],
		"defer": body[defer_:],
	} {
		_, prose, _ := strings.Cut(section, "\n")
		words := roleSkillBodyWordCount(prose)
		if words > maxBoundarySkillBodyWords {
			return fmt.Errorf(
				"boundary %q %s side has %d words, maximum is %d",
				boundaryName, label, words, maxBoundarySkillBodyWords,
			)
		}
	}
	return nil
}

// validateBoundaryOwners keeps a two-sided boundary coherent. An owner receives
// the body by owning it, never by declaring it.
func validateBoundaryOwners(p *Person) error {
	for _, boundaryName := range p.boundaryOrder() {
		owner := p.Boundaries[boundaryName].Owner
		if owner == "" {
			return fmt.Errorf("boundary %q has no owner", boundaryName)
		}
		role, ok := p.Roles[owner]
		if !ok {
			return fmt.Errorf("boundary %q names unknown owner %q", boundaryName, owner)
		}
		for _, declared := range role.Boundaries {
			if declared == boundaryName {
				return fmt.Errorf(
					"boundary %q owner %q also declares it",
					boundaryName,
					owner,
				)
			}
		}
	}
	return nil
}

// validateRoleAdjacents keeps the graph complete. Adjacency is deliberately
// directed, so do not add a symmetry check. See docs/role-boundaries.md.
func validateRoleAdjacents(p *Person) error {
	declared := false
	for _, roleName := range p.roleOrder() {
		if len(p.Roles[roleName].Adjacents) > 0 {
			declared = true
			break
		}
	}
	if !declared {
		return nil
	}
	for _, roleName := range p.roleOrder() {
		adjacents := p.Roles[roleName].Adjacents
		if len(adjacents) != adjacentsPerRole {
			return fmt.Errorf(
				"role %q declares %d adjacent roles, needs exactly %d",
				roleName,
				len(adjacents),
				adjacentsPerRole,
			)
		}
		for _, adjacent := range adjacents {
			if _, ok := p.Roles[adjacent.Role]; !ok {
				return fmt.Errorf(
					"role %q names unknown adjacent role %q", roleName, adjacent.Role,
				)
			}
		}
	}
	return nil
}

func validateCorePersonalityMelds(p *Person) error {
	usage := map[string]int{}
	colors := map[string]string{}
	for _, roleName := range p.RoleOrder {
		role := p.Roles[roleName]
		if len(role.Personalities) != 3 {
			return fmt.Errorf(
				"core role %q has %d personalities, want exactly three",
				roleName,
				len(role.Personalities),
			)
		}
		components := make([]string, 0, len(role.Personalities))
		for _, name := range role.Personalities {
			usage[name]++
			components = append(components, p.Personalities[name].Color)
		}
		favorite, err := color.Favorite(components)
		if err != nil {
			return fmt.Errorf("core role %q favorite color: %w", roleName, err)
		}
		if err := color.Legible(favorite); err != nil {
			return fmt.Errorf("core role %q favorite color: %w", roleName, err)
		}
		if existing, ok := colors[favorite]; ok {
			return fmt.Errorf(
				"core roles %q and %q share melded favorite color %q",
				existing,
				roleName,
				favorite,
			)
		}
		colors[favorite] = roleName
	}
	for name, count := range usage {
		if count > 3 {
			return fmt.Errorf(
				"core personality %q appears in %d roles, want at most three",
				name,
				count,
			)
		}
	}
	return nil
}

func loadLibrary(root string) (*Person, string, fs.FS, error) {
	absolute, err := filepath.Abs(root)
	if err != nil {
		return nil, "", nil, err
	}
	info, err := os.Stat(absolute)
	if err != nil {
		return nil, "", nil, fmt.Errorf("inspect personality library %s: %w", absolute, err)
	}
	if !info.IsDir() {
		return nil, "", nil, fmt.Errorf("personality library %s is not a directory", absolute)
	}
	if err := filepath.WalkDir(absolute, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.Type()&fs.ModeSymlink != 0 {
			return fmt.Errorf("personality library contains symlink %s", path)
		}
		return nil
	}); err != nil {
		return nil, "", nil, fmt.Errorf("inspect personality library tree %s: %w", absolute, err)
	}
	source := os.DirFS(absolute)
	library, id, err := loadLibrarySource(source, "personality library "+absolute)
	if err != nil {
		return nil, "", nil, err
	}
	return library, id, source, nil
}

func loadLibrarySource(source fs.FS, label string) (*Person, string, error) {
	raw, err := fs.ReadFile(source, "library.kdl")
	if err != nil {
		return nil, "", fmt.Errorf("%s: read manifest: %w", label, err)
	}
	doc, err := kdl.ParseString(string(raw))
	if err != nil || len(doc.Nodes) != 1 || doc.Nodes[0].Name() != "library" || len(doc.Nodes[0].Arguments()) != 1 {
		return nil, "", fmt.Errorf("%s: needs one library name", label)
	}
	id := doc.Nodes[0].Arguments()[0].String()
	if !validLogicalID(id) {
		return nil, "", fmt.Errorf("%s: library name %q is not a stable logical id", label, id)
	}
	assembled, err := assembleLibrarySource(source, id)
	if err != nil {
		return nil, "", err
	}
	library, err := parse(assembled)
	if err != nil {
		return nil, "", fmt.Errorf("parse personality library %q: %w", id, err)
	}
	// A library may publish shared doctrine, so its boundary bodies are read and
	// bounded here rather than only in the consuming package.
	if err := loadBoundarySkills(source, library); err != nil {
		return nil, "", fmt.Errorf("personality library %q: %w", id, err)
	}
	return library, id, nil
}

func assembleLibrarySource(source fs.FS, id string) ([]byte, error) {
	var out bytes.Buffer
	fmt.Fprintf(&out, "person %q {\n", id)
	for _, directory := range []string{"boundaries", "personalities", "inspirations"} {
		entries, err := fs.ReadDir(source, directory)
		if os.IsNotExist(err) && (directory == "inspirations" || directory == "boundaries") {
			continue
		}
		if err != nil {
			return nil, fmt.Errorf("read personality library %s: %w", directory, err)
		}
		if len(entries) == 0 && directory == "personalities" {
			return nil, fmt.Errorf("personality library %q has empty %s", id, directory)
		}
		for _, entry := range entries {
			if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".kdl") {
				return nil, fmt.Errorf("personality library %q has unexpected %s entry %q", id, directory, entry.Name())
			}
			raw, err := fs.ReadFile(source, directory+"/"+entry.Name())
			if err != nil {
				return nil, err
			}
			out.WriteString(indentPersonFragment(bytes.TrimSpace(raw)))
			out.WriteByte('\n')
		}
	}
	out.WriteString("}\n")
	return out.Bytes(), nil
}

func equalPersonality(left, right Personality) bool {
	return fmt.Sprintf("%#v", left) == fmt.Sprintf("%#v", right)
}

func definitionOverlay(
	base fs.FS, personalities map[string]Personality, boundaries map[string]Boundary,
) (fstest.MapFS, error) {
	overlay := fstest.MapFS{}
	if err := appendDefinitions(overlay, base, personalities, boundaries); err != nil {
		return nil, err
	}
	return overlay, nil
}

func appendDefinitions(
	overlay fstest.MapFS,
	source fs.FS,
	personalities map[string]Personality,
	boundaries map[string]Boundary,
) error {
	for _, binding := range personalities {
		path := "definitions/skills/" + binding.Skill + "/SKILL.md"
		raw, err := fs.ReadFile(source, path)
		if err != nil {
			return fmt.Errorf("personality skill %q: read definition: %w", binding.Skill, err)
		}
		if existing, ok := overlay[path]; ok && !bytes.Equal(existing.Data, raw) {
			return fmt.Errorf("personality definition %q conflicts between local libraries", binding.Skill)
		}
		overlay[path] = &fstest.MapFile{Data: raw, Mode: 0o644}
	}
	for _, binding := range boundaries {
		path := "definitions/skills/" + binding.Skill + "/SKILL.md"
		raw, err := fs.ReadFile(source, path)
		if err != nil {
			return fmt.Errorf("boundary skill %q: read definition: %w", binding.Skill, err)
		}
		if existing, ok := overlay[path]; ok && !bytes.Equal(existing.Data, raw) {
			return fmt.Errorf("boundary definition %q conflicts between local libraries", binding.Skill)
		}
		overlay[path] = &fstest.MapFile{Data: raw, Mode: 0o644}
	}
	if _, exists := overlay["definitions/INVARIANT.md"]; !exists {
		raw, err := fs.ReadFile(source, "definitions/INVARIANT.md")
		if err != nil {
			return fmt.Errorf("read profile invariant: %w", err)
		}
		overlay["definitions/INVARIANT.md"] = &fstest.MapFile{Data: raw, Mode: 0o644}
	}
	return nil
}

// validatePersonalityBodies bounds each personality body. External libraries
// supply their own definitions, so a missing one is not an error here.
func validatePersonalityBodies(source fs.FS, p *Person) error {
	for name, binding := range p.Personalities {
		raw, err := fs.ReadFile(source, "definitions/skills/"+binding.Skill+"/SKILL.md")
		if err != nil {
			continue
		}
		body, err := skillBody(raw)
		if err != nil {
			return fmt.Errorf("personality %q skill %q: %w", name, binding.Skill, err)
		}
		words := roleSkillBodyWordCount(body)
		if words > maxPersonalitySkillBodyWords {
			return fmt.Errorf(
				"personality %q skill body has %d words, maximum is %d",
				name, words, maxPersonalitySkillBodyWords,
			)
		}
	}
	return nil
}

// validateRosterProseFloors keeps a shipped entry from thinning into a label.
// The ceilings bind every package, the floors bind the roster this repo ships.
func validateRosterProseFloors(source fs.FS, p *Person) error {
	for _, roleName := range p.roleOrder() {
		words := roleSkillBodyWordCount(p.Roles[roleName].Briefing)
		if words < minRoleSkillBodyWords {
			return fmt.Errorf("role %q body has %d words, minimum is %d", roleName, words, minRoleSkillBodyWords)
		}
	}
	for name, binding := range p.Personalities {
		raw, err := fs.ReadFile(source, "definitions/skills/"+binding.Skill+"/SKILL.md")
		if err != nil {
			continue
		}
		body, err := skillBody(raw)
		if err != nil {
			return err
		}
		if words := roleSkillBodyWordCount(body); words < minPersonalitySkillBodyWords {
			return fmt.Errorf("personality %q body has %d words, minimum is %d", name, words, minPersonalitySkillBodyWords)
		}
	}
	for _, boundaryName := range p.boundaryOrder() {
		raw, ok := p.BoundarySkillDefinition(boundaryName)
		if !ok {
			continue
		}
		body, err := skillBody(raw)
		if err != nil {
			return err
		}
		own := strings.Index(body, boundaryOwnHeading)
		defer_ := strings.Index(body, boundaryDeferHeading)
		if own < 0 || defer_ < 0 {
			continue
		}
		for label, section := range map[string]string{"own": body[own:defer_], "defer": body[defer_:]} {
			_, prose, _ := strings.Cut(section, "\n")
			if words := roleSkillBodyWordCount(prose); words < minBoundarySideWords {
				return fmt.Errorf(
					"boundary %q %s side has %d words, minimum is %d",
					boundaryName, label, words, minBoundarySideWords,
				)
			}
		}
	}
	return nil
}

func loadSource(source fs.FS, label string) (*Person, error) {
	raw, err := assemblePersonSource(source, label)
	if err != nil {
		return nil, err
	}
	p, err := parse(raw)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", label, err)
	}
	if err := loadRoleSkills(source, p); err != nil {
		return nil, fmt.Errorf("%s: %w", label, err)
	}
	if err := loadRoleMethods(source, p); err != nil {
		return nil, fmt.Errorf("%s: %w", label, err)
	}
	if err := loadBoundarySkills(source, p); err != nil {
		return nil, fmt.Errorf("%s: %w", label, err)
	}
	if err := validatePersonalityBodies(source, p); err != nil {
		return nil, fmt.Errorf("%s: %w", label, err)
	}
	if err := addCopyContractProvenance(p); err != nil {
		return nil, fmt.Errorf("%s: %w", label, err)
	}
	p.source = source
	return p, nil
}

func addCopyContractProvenance(p *Person) error {
	for _, roleName := range p.RoleOrder {
		role := p.Roles[roleName]
		if role.CopyContract == nil {
			continue
		}
		canonical := struct {
			Scope string     `json:"scope"`
			Rules []CopyRule `json:"rules"`
		}{
			Scope: role.CopyContract.Scope,
			Rules: role.CopyContract.Rules,
		}
		raw, err := json.Marshal(canonical)
		if err != nil {
			return fmt.Errorf("role %q copy contract: %w", roleName, err)
		}
		digest := sha256.Sum256(raw)
		role.CopyContract.Source = p.ProviderID() + ":role:" + roleName + ":copy-contract"
		role.CopyContract.Digest = fmt.Sprintf("sha256:%x", digest)
		p.Roles[roleName] = role
	}
	return nil
}

func loadRoleSkills(source fs.FS, p *Person) error {
	p.roleSkills = map[string][]byte{}
	for _, roleName := range p.RoleOrder {
		role := p.Roles[roleName]
		if role.Briefing != "" {
			raw := []byte(fmt.Sprintf(
				"---\nname: %s\ndescription: Adopt the %s charter. Use when the session activates the %s role.\n---\n\n# %s\n\n%s\n",
				role.Skill,
				roleName,
				roleName,
				roleName,
				role.Briefing,
			))
			p.roleSkills[roleName] = raw
			role.SkillSource = p.ProviderID() + ":legacy-role:" + roleName
			digest := sha256.Sum256(raw)
			role.SkillDigest = fmt.Sprintf("sha256:%x", digest)
			p.Roles[roleName] = role
			continue
		}
		if role.Skill != "role-"+roleName {
			return fmt.Errorf("role %q skill %q must be role-%s", roleName, role.Skill, roleName)
		}
		raw, err := fs.ReadFile(source, "roles/"+roleName+"/SKILL.md")
		if err != nil {
			return fmt.Errorf("role %q skill %q: %w", roleName, role.Skill, err)
		}
		if err := validateSkillDefinition(role.Skill, raw); err != nil {
			return err
		}
		body, err := skillBody(raw)
		if err != nil {
			return fmt.Errorf("role %q skill %q: %w", roleName, role.Skill, err)
		}
		if words := roleSkillBodyWordCount(body); words > maxRoleSkillBodyWords {
			return fmt.Errorf(
				"role %q skill body has %d words, maximum is %d",
				roleName,
				words,
				maxRoleSkillBodyWords,
			)
		}
		if paragraphs := briefingParagraphCount(body); paragraphs < 3 {
			return fmt.Errorf("role %q skill needs at least three paragraphs, got %d", roleName, paragraphs)
		}
		role.Briefing = body
		role.SkillSource = p.ProviderID() + ":role:" + roleName
		digest := sha256.Sum256(raw)
		role.SkillDigest = fmt.Sprintf("sha256:%x", digest)
		p.roleSkills[roleName] = append([]byte(nil), raw...)
		p.Roles[roleName] = role
	}
	return nil
}

// loadBoundarySkills reads each shared doctrine body and bounds it against the boundary
// budget. The body never enters Role.Briefing, so it spends no role budget.
func loadBoundarySkills(source fs.FS, p *Person) error {
	p.boundarySkills = map[string][]byte{}
	for _, boundaryName := range p.BoundaryOrder {
		boundary := p.Boundaries[boundaryName]
		path := "definitions/skills/" + boundary.Skill + "/SKILL.md"
		raw, err := fs.ReadFile(source, path)
		if err != nil {
			return fmt.Errorf("boundary %q skill %q: read definition: %w", boundaryName, boundary.Skill, err)
		}
		if err := validateSkillDefinition(boundary.Skill, raw); err != nil {
			return err
		}
		body, err := skillBody(raw)
		if err != nil {
			return fmt.Errorf("boundary %q skill %q: %w", boundaryName, boundary.Skill, err)
		}
		if err := validateBoundaryBodySides(boundaryName, body); err != nil {
			return err
		}
		digest := sha256.Sum256(raw)
		boundary.Source = p.ProviderID() + ":boundary:" + boundaryName
		boundary.Digest = fmt.Sprintf("sha256:%x", digest)
		p.Boundaries[boundaryName] = boundary
		p.boundarySkills[boundaryName] = append([]byte(nil), raw...)
	}
	return nil
}

// BoundarySkillDefinition returns the raw shared doctrine skill for one boundary.
func (p *Person) BoundarySkillDefinition(boundaryName string) ([]byte, bool) {
	raw, ok := p.boundarySkills[boundaryName]
	if !ok {
		return nil, false
	}
	return append([]byte(nil), raw...), true
}

// RoleOwnedBoundaries returns the boundaries this role owns. An owner receives
// the body without declaring it. See docs/ownership.md.
func (p *Person) RoleOwnedBoundaries(roleName string) []string {
	owned := make([]string, 0, 1)
	for _, name := range p.boundaryOrder() {
		if p.Boundaries[name].Owner == roleName {
			owned = append(owned, name)
		}
	}
	return owned
}

// RoleActiveBoundaries returns every boundary whose body this role receives,
// declared first and then owned.
func (p *Person) RoleActiveBoundaries(roleName string) []string {
	role, ok := p.Roles[roleName]
	if !ok {
		return nil
	}
	active := append([]string(nil), role.Boundaries...)
	return append(active, p.RoleOwnedBoundaries(roleName)...)
}

// RoleBoundarySkillIDs returns the boundary skill ids one role activates, in
// declared then owned order, so callers can name them without loading bodies.
func (p *Person) RoleBoundarySkillIDs(roleName string) []string {
	if _, ok := p.Roles[roleName]; !ok {
		return nil
	}
	active := p.RoleActiveBoundaries(roleName)
	ids := make([]string, 0, len(active))
	for _, boundaryName := range active {
		binding, exists := p.Boundaries[boundaryName]
		if !exists {
			continue
		}
		ids = append(ids, binding.Skill)
	}
	return ids
}

func loadRoleMethods(source fs.FS, p *Person) error {
	p.roleMethods = map[string]map[string][]byte{}
	claimed := map[string]string{}
	for _, roleName := range p.RoleOrder {
		role := p.Roles[roleName]
		methodsRoot := "roles/" + roleName + "/skills"
		if len(role.Methods) == 0 {
			if _, err := fs.Stat(source, methodsRoot); err == nil {
				return fmt.Errorf("role %q has an undeclared skills directory", roleName)
			} else if !os.IsNotExist(err) {
				return fmt.Errorf("role %q skills: %w", roleName, err)
			}
			continue
		}

		expected := make(map[string]bool, len(role.Methods))
		for _, method := range role.Methods {
			if owner, duplicate := claimed[method]; duplicate {
				return fmt.Errorf("roles %q and %q bind the same method skill %q", owner, roleName, method)
			}
			claimed[method] = roleName
			expected[method] = true
		}
		entries, err := fs.ReadDir(source, methodsRoot)
		if err != nil {
			return fmt.Errorf("role %q method skills: %w", roleName, err)
		}
		if len(entries) != len(expected) {
			return fmt.Errorf(
				"role %q binds %d method skills but its directory contains %d entries",
				roleName,
				len(expected),
				len(entries),
			)
		}
		p.roleMethods[roleName] = map[string][]byte{}
		for _, entry := range entries {
			method := entry.Name()
			if !entry.IsDir() || !expected[method] {
				return fmt.Errorf("role %q method skills has unexpected entry %q", roleName, method)
			}
			methodRoot := methodsRoot + "/" + method
			methodEntries, err := fs.ReadDir(source, methodRoot)
			if err != nil {
				return fmt.Errorf("role %q method skill %q: %w", roleName, method, err)
			}
			if len(methodEntries) != 1 || methodEntries[0].IsDir() || methodEntries[0].Name() != "SKILL.md" {
				return fmt.Errorf("role %q method skill %q must contain only SKILL.md", roleName, method)
			}
			raw, err := fs.ReadFile(source, methodRoot+"/SKILL.md")
			if err != nil {
				return fmt.Errorf("role %q method skill %q: %w", roleName, method, err)
			}
			if err := validateSkillDefinition(method, raw); err != nil {
				return err
			}
			p.roleMethods[roleName][method] = append([]byte(nil), raw...)
		}
	}
	return nil
}

func skillBody(raw []byte) (string, error) {
	text := strings.ReplaceAll(string(raw), "\r\n", "\n")
	if !strings.HasPrefix(text, "---\n") {
		return "", fmt.Errorf("missing YAML frontmatter")
	}
	end := strings.Index(text[4:], "\n---\n")
	if end < 0 {
		return "", fmt.Errorf("unterminated YAML frontmatter")
	}
	return strings.TrimSpace(text[end+9:]), nil
}

// RoleSkillDefinition returns the canonical role skill selected by the profile.
func (p *Person) RoleSkillDefinition(roleName string) ([]byte, bool) {
	raw, ok := p.roleSkills[roleName]
	if !ok {
		role, exists := p.Roles[roleName]
		if !exists || strings.TrimSpace(role.Briefing) == "" {
			return nil, false
		}
		skill := role.Skill
		if skill == "" {
			skill = "role-" + roleName
		}
		raw = []byte(fmt.Sprintf(
			"---\nname: %s\ndescription: Adopt the %s charter. Use when the session activates the %s role.\n---\n\n# %s\n\n%s\n",
			skill,
			roleName,
			roleName,
			roleName,
			role.Briefing,
		))
		ok = true
	}
	return append([]byte(nil), raw...), ok
}

// RoleSkillID returns the stable skill binding, including the v1.x legacy
// adapter for callers that construct an inline role in memory.
func (p *Person) RoleSkillID(roleName string) string {
	if role, ok := p.Roles[roleName]; ok && role.Skill != "" {
		return role.Skill
	}
	return "role-" + roleName
}

// RoleMethodDefinition returns one role-bound progressive-disclosure skill.
func (p *Person) RoleMethodDefinition(roleName, method string) ([]byte, bool) {
	methods, ok := p.roleMethods[roleName]
	if !ok {
		return nil, false
	}
	raw, ok := methods[method]
	return append([]byte(nil), raw...), ok
}

func assemblePersonSource(source fs.FS, label string) ([]byte, error) {
	manifest, err := fs.ReadFile(source, "person.kdl")
	if err != nil {
		return nil, fmt.Errorf("%s: read person manifest: %w", label, err)
	}
	manifest = bytes.TrimSpace(manifest)
	doc, err := kdl.ParseString(string(manifest))
	if err != nil {
		return nil, fmt.Errorf("%s: parse person manifest: %w", label, err)
	}
	if len(doc.Nodes) != 1 ||
		(doc.Nodes[0].Name() != "person" && doc.Nodes[0].Name() != "roster") {
		return nil, fmt.Errorf("%s: manifest needs exactly one person or roster node", label)
	}
	root := doc.Nodes[0]
	if len(root.Arguments()) != 1 {
		return nil, fmt.Errorf("%s: manifest node needs one name argument", label)
	}
	if len(root.Children().Nodes) != 0 {
		return nil, fmt.Errorf("%s: person policy nodes belong in section files", label)
	}

	var assembled bytes.Buffer
	assembled.Write(manifest)
	assembled.WriteString(" {\n")
	firstFragment := true
	for _, section := range personSections {
		entries, err := fs.ReadDir(source, section.directory)
		if os.IsNotExist(err) && section.directory != "roles" {
			continue
		}
		if err != nil {
			return nil, fmt.Errorf("%s: read person %s: %w", label, section.directory, err)
		}
		if len(entries) == 0 {
			return nil, fmt.Errorf("%s: person %s section is empty", label, section.directory)
		}
		for _, entry := range entries {
			if entry.IsDir() && section.directory == "roles" {
				if _, ok := pRoleSkillDirectory(entry.Name()); !ok {
					return nil, fmt.Errorf(
						"%s: person roles has unexpected entry %q",
						label,
						entry.Name(),
					)
				}
				continue
			}
			if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".kdl") {
				return nil, fmt.Errorf(
					"%s: person %s has unexpected entry %q",
					label,
					section.directory,
					entry.Name(),
				)
			}
			path := section.directory + "/" + entry.Name()
			fragment, err := fs.ReadFile(source, path)
			if err != nil {
				return nil, fmt.Errorf("%s: read person fragment %q: %w", label, path, err)
			}
			fragment = bytes.TrimSpace(fragment)
			fragmentDoc, err := kdl.ParseString(string(fragment))
			if err != nil {
				return nil, fmt.Errorf("%s: parse person fragment %q: %w", label, path, err)
			}
			if len(fragmentDoc.Nodes) != 1 || fragmentDoc.Nodes[0].Name() != section.node {
				return nil, fmt.Errorf(
					"%s: person fragment %q needs exactly one %s node",
					label,
					path,
					section.node,
				)
			}
			args := fragmentDoc.Nodes[0].Arguments()
			if len(args) != 1 {
				return nil, fmt.Errorf(
					"%s: person fragment %q %s node needs one name argument",
					label,
					path,
					section.node,
				)
			}
			if want, ok := personFragmentSlug(entry.Name()); !ok || args[0].String() != want {
				return nil, fmt.Errorf(
					"%s: person fragment %q filename does not match %s %q",
					label,
					path,
					section.node,
					args[0].String(),
				)
			}
			if !firstFragment {
				assembled.WriteByte('\n')
			}
			assembled.WriteString(indentPersonFragment(fragment))
			assembled.WriteByte('\n')
			firstFragment = false
		}
	}
	assembled.WriteString("}\n")
	return assembled.Bytes(), nil
}

func pRoleSkillDirectory(name string) (string, bool) {
	if !validSemanticToken(name) {
		return "", false
	}
	return name, true
}

func personFragmentSlug(name string) (string, bool) {
	if len(name) < len("00-a.kdl") ||
		name[0] < '0' || name[0] > '9' ||
		name[1] < '0' || name[1] > '9' ||
		name[2] != '-' ||
		!strings.HasSuffix(name, ".kdl") {
		return "", false
	}
	return strings.TrimSuffix(name[3:], ".kdl"), true
}

func indentPersonFragment(fragment []byte) string {
	text := strings.ReplaceAll(string(fragment), "\r\n", "\n")
	return "    " + strings.ReplaceAll(text, "\n", "\n    ")
}

// Source returns the selected person's content provider. The role catalog,
// invariant, and every bound personality definition come from the same package.
func Source(p *Person) (*schema.Source, error) {
	if p == nil {
		return nil, fmt.Errorf("person is required")
	}
	source := p.source
	strictDefinitions := true
	if source == nil {
		projected, _, err := dataLayout(embedded, "embedded core roster")
		if err != nil {
			return nil, err
		}
		overlay, err := definitionOverlay(projected, p.Personalities, p.Boundaries)
		if err != nil {
			return nil, err
		}
		source = overlay
		strictDefinitions = false
	}
	definitions, err := fs.Sub(source, "definitions")
	if err != nil {
		return nil, fmt.Errorf("open person %q personality definitions: %w", p.Name, err)
	}
	files := fstest.MapFS{}
	if err := fs.WalkDir(definitions, ".", func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() {
			return nil
		}
		raw, readErr := fs.ReadFile(definitions, path)
		if readErr != nil {
			return readErr
		}
		files[path] = &fstest.MapFile{Data: raw, Mode: 0o644}
		return nil
	}); err != nil {
		return nil, fmt.Errorf("copy person %q definitions: %w", p.Name, err)
	}
	invariant, err := fs.ReadFile(files, "INVARIANT.md")
	if err != nil {
		return nil, fmt.Errorf("read person %q personality invariant: %w", p.Name, err)
	}
	if len(bytes.TrimSpace(invariant)) == 0 {
		return nil, fmt.Errorf("person %q personality invariant is empty", p.Name)
	}

	// The definitions tree carries personality and boundary bodies alike, so the
	// canonical set spans both catalogs.
	expected := make(map[string]bool, len(p.Personalities)+len(p.Boundaries))
	canonicalSkills := make([]string, 0, len(p.Personalities)+len(p.Boundaries))
	for _, binding := range p.Personalities {
		expected[binding.Skill] = true
		canonicalSkills = append(canonicalSkills, binding.Skill)
	}
	boundarySkills := make(map[string]bool, len(p.Boundaries))
	for _, binding := range p.Boundaries {
		expected[binding.Skill] = true
		boundarySkills[binding.Skill] = true
		canonicalSkills = append(canonicalSkills, binding.Skill)
	}
	sort.Strings(canonicalSkills)

	selected := make(map[string]bool, len(p.Personalities))
	for name, binding := range p.Personalities {
		if binding.Skill == "" {
			return nil, fmt.Errorf("personality %q has no skill binding", name)
		}
		if !expected[binding.Skill] {
			return nil, fmt.Errorf("personality %q binds unknown skill %q", name, binding.Skill)
		}
		if selected[binding.Skill] {
			return nil, fmt.Errorf("personality skill %q is bound more than once", binding.Skill)
		}
		selected[binding.Skill] = true
	}

	entries, err := fs.ReadDir(files, "skills")
	if err != nil {
		return nil, fmt.Errorf("read person %q personality skills: %w", p.Name, err)
	}
	if strictDefinitions && len(entries) != len(canonicalSkills) {
		return nil, fmt.Errorf(
			"person %q personality skills: catalog binds %d skills but definitions contain %d entries",
			p.Name,
			len(canonicalSkills),
			len(entries),
		)
	}

	src := &schema.Source{
		ID:    p.ProviderID(),
		Files: files,
		Instructions: []schema.ContentRef{{
			ID:   "personality-invariant",
			Path: "INVARIANT.md",
		}},
	}
	for _, entry := range entries {
		if !entry.IsDir() || strictDefinitions && !expected[entry.Name()] {
			return nil, fmt.Errorf("person %q personality skills: unexpected entry %q", p.Name, entry.Name())
		}
	}
	for _, skill := range canonicalSkills {
		skillPath := "skills/" + skill
		raw, err := fs.ReadFile(files, skillPath+"/SKILL.md")
		if err != nil {
			return nil, fmt.Errorf("read person %q skill %q: %w", p.Name, skill, err)
		}
		if err := validateSkillDefinition(skill, raw); err != nil {
			return nil, err
		}
		if !selected[skill] {
			continue
		}
		src.Skills = append(src.Skills, schema.ContentRef{
			ID:         skill,
			Path:       skillPath,
			EntryPoint: "SKILL.md",
		})
	}
	src.RoleSkills = map[string][]schema.ContentRef{}
	for _, roleName := range p.roleOrder() {
		roleSkill := p.RoleSkillID(roleName)
		raw, ok := p.RoleSkillDefinition(roleName)
		if !ok {
			return nil, fmt.Errorf("person %q role %q has no skill definition", p.Name, roleName)
		}
		path := "skills/" + roleSkill + "/SKILL.md"
		files[path] = &fstest.MapFile{Data: raw, Mode: 0o644}
		src.RoleSkills[roleName] = []schema.ContentRef{{
			ID:         roleSkill,
			Path:       "skills/" + roleSkill,
			EntryPoint: "SKILL.md",
		}}
		for _, method := range p.Roles[roleName].Methods {
			raw, ok := p.RoleMethodDefinition(roleName, method)
			if !ok {
				return nil, fmt.Errorf("person %q role %q has no method skill %q", p.Name, roleName, method)
			}
			path := "skills/" + method + "/SKILL.md"
			if _, collision := files[path]; collision {
				return nil, fmt.Errorf("person %q role %q method skill %q collides with another person skill", p.Name, roleName, method)
			}
			files[path] = &fstest.MapFile{Data: raw, Mode: 0o644}
			src.RoleSkills[roleName] = append(src.RoleSkills[roleName], schema.ContentRef{
				ID:         method,
				Path:       "skills/" + method,
				EntryPoint: "SKILL.md",
			})
		}
		for _, boundaryName := range p.RoleActiveBoundaries(roleName) {
			binding, ok := p.Boundaries[boundaryName]
			if !ok {
				return nil, fmt.Errorf("person %q role %q has no boundary binding %q", p.Name, roleName, boundaryName)
			}
			raw, ok := p.BoundarySkillDefinition(boundaryName)
			if !ok {
				return nil, fmt.Errorf("person %q role %q has no boundary skill %q", p.Name, roleName, boundaryName)
			}
			path := "skills/" + binding.Skill + "/SKILL.md"
			// Several roles share one boundary body, so an identical repeat is the
			// expected case and only a differing body is a real collision.
			if existing, collision := files[path]; collision && !bytes.Equal(existing.Data, raw) {
				return nil, fmt.Errorf(
					"person %q role %q boundary skill %q collides with another person skill",
					p.Name, roleName, binding.Skill,
				)
			}
			files[path] = &fstest.MapFile{Data: raw, Mode: 0o644}
			src.RoleSkills[roleName] = append(src.RoleSkills[roleName], schema.ContentRef{
				ID:         binding.Skill,
				Path:       "skills/" + binding.Skill,
				EntryPoint: "SKILL.md",
			})
		}
	}
	return src, nil
}

func validateSkillDefinition(skill string, raw []byte) error {
	text := strings.ReplaceAll(string(raw), "\r\n", "\n")
	if !strings.HasPrefix(text, "---\n") {
		return fmt.Errorf("person skill %q: SKILL.md needs YAML frontmatter", skill)
	}
	end := strings.Index(text[4:], "\n---\n")
	if end < 0 {
		return fmt.Errorf("person skill %q: SKILL.md has unterminated frontmatter", skill)
	}
	end += 4
	frontmatter := "\n" + text[4:end] + "\n"
	if !strings.Contains(frontmatter, "\nname: "+skill+"\n") {
		return fmt.Errorf("person skill %q: frontmatter name does not match", skill)
	}
	if !strings.Contains(frontmatter, "\ndescription: ") {
		return fmt.Errorf("person skill %q: frontmatter needs a description", skill)
	}
	if strings.TrimSpace(text[end+5:]) == "" {
		return fmt.Errorf("person skill %q: SKILL.md body is empty", skill)
	}
	return nil
}

func parse(raw []byte) (*Person, error) {
	doc, err := kdl.ParseString(string(raw))
	if err != nil {
		return nil, fmt.Errorf("parse person source: %w", err)
	}
	if len(doc.Nodes) != 1 ||
		(doc.Nodes[0].Name() != "person" && doc.Nodes[0].Name() != "roster") {
		return nil, fmt.Errorf("person source needs exactly one top-level person or roster node")
	}
	root := doc.Nodes[0]
	args := root.Arguments()
	if len(args) != 1 {
		return nil, fmt.Errorf("person source provider node needs one name argument")
	}
	p := &Person{
		ProviderKind:   doc.Nodes[0].Name(),
		Name:           args[0].String(),
		Roles:          map[string]Role{},
		Boundaries:     map[string]Boundary{},
		Personalities:  map[string]Personality{},
		Inspirations:   map[string]Inspiration{},
		roleSkills:     map[string][]byte{},
		roleMethods:    map[string]map[string][]byte{},
		boundarySkills: map[string][]byte{},
		Raw:            raw,
	}
	for _, n := range root.Children().Nodes {
		switch n.Name() {
		case "role":
			rargs := n.Arguments()
			if len(rargs) != 1 {
				return nil, fmt.Errorf("person role node needs one name argument")
			}
			name := rargs[0].String()
			if _, dup := p.Roles[name]; dup {
				return nil, fmt.Errorf("person role %q declared twice", name)
			}
			role := Role{}
			briefingSet := false
			skillSet := false
			for _, c := range n.Children().Nodes {
				switch c.Name() {
				case "identity":
					if role.Identity != nil {
						return nil, fmt.Errorf("role %q: duplicate identity", name)
					}
					if len(c.Arguments()) != 0 {
						return nil, fmt.Errorf("role %q: identity takes name and pronouns properties", name)
					}
					identityName := c.Prop("name")
					pronouns := c.Prop("pronouns")
					if !identityName.IsValid() || strings.TrimSpace(identityName.String()) == "" {
						return nil, fmt.Errorf("role %q: identity needs a name property", name)
					}
					if !pronouns.IsValid() || strings.TrimSpace(pronouns.String()) == "" {
						return nil, fmt.Errorf("role %q: identity needs a pronouns property", name)
					}
					role.Identity = &AgentIdentity{
						Name:     strings.TrimSpace(identityName.String()),
						Pronouns: strings.TrimSpace(pronouns.String()),
					}
				case "display-name":
					if role.DisplayName != "" {
						return nil, fmt.Errorf("role %q: duplicate display-name", name)
					}
					dargs := c.Arguments()
					if len(dargs) != 1 || strings.TrimSpace(dargs[0].String()) == "" {
						return nil, fmt.Errorf("role %q: display-name needs one argument", name)
					}
					role.DisplayName = strings.TrimSpace(dargs[0].String())
				case "purpose":
					if role.Purpose != "" {
						return nil, fmt.Errorf("role %q: duplicate purpose", name)
					}
					pargs := c.Arguments()
					if len(pargs) != 1 {
						return nil, fmt.Errorf("role %q: purpose needs one argument", name)
					}
					role.Purpose = pargs[0].String()
				case "briefing":
					if briefingSet {
						return nil, fmt.Errorf("role %q: duplicate briefing", name)
					}
					briefingSet = true
					bargs := c.Arguments()
					if len(bargs) != 1 {
						return nil, fmt.Errorf("role %q: briefing needs one argument", name)
					}
					role.Briefing = strings.TrimSpace(bargs[0].String())
					if role.Briefing == "" {
						return nil, fmt.Errorf("role %q: briefing must not be empty", name)
					}
				case "skill":
					if skillSet {
						return nil, fmt.Errorf("role %q: duplicate skill", name)
					}
					skillSet = true
					sargs := c.Arguments()
					if len(sargs) != 1 || strings.TrimSpace(sargs[0].String()) == "" {
						return nil, fmt.Errorf("role %q: skill needs one id", name)
					}
					role.Skill = sargs[0].String()
				case "method":
					margs := c.Arguments()
					if len(margs) != 1 || !validSemanticToken(margs[0].String()) {
						return nil, fmt.Errorf("role %q: method needs one stable skill id", name)
					}
					method := margs[0].String()
					for _, existing := range role.Methods {
						if existing == method {
							return nil, fmt.Errorf("role %q repeats method skill %q", name, method)
						}
					}
					role.Methods = append(role.Methods, method)
				case "boundary":
					margs := c.Arguments()
					if len(margs) == 0 {
						return nil, fmt.Errorf("role %q: boundary needs at least one shared doctrine id", name)
					}
					for _, a := range margs {
						boundary := a.String()
						if !validSemanticToken(boundary) {
							return nil, fmt.Errorf("role %q: boundary needs one stable doctrine id", name)
						}
						for _, existing := range role.Boundaries {
							if existing == boundary {
								return nil, fmt.Errorf("role %q repeats boundary %q", name, boundary)
							}
						}
						role.Boundaries = append(role.Boundaries, boundary)
					}
				case "adjacent":
					aargs := c.Arguments()
					if len(aargs) != 1 {
						return nil, fmt.Errorf("role %q: adjacent needs one role id", name)
					}
					target := aargs[0].String()
					if !validSemanticToken(target) {
						return nil, fmt.Errorf("role %q: adjacent needs one stable role id", name)
					}
					if target == name {
						return nil, fmt.Errorf("role %q: adjacent cannot name itself", name)
					}
					for _, existing := range role.Adjacents {
						if existing.Role == target {
							return nil, fmt.Errorf("role %q repeats adjacent %q", name, target)
						}
					}
					reason := c.Prop("reason")
					if !reason.IsValid() || strings.TrimSpace(reason.String()) == "" {
						return nil, fmt.Errorf(
							"role %q: adjacent %q needs a reason property", name, target,
						)
					}
					role.Adjacents = append(role.Adjacents, Adjacent{
						Role:   target,
						Reason: strings.TrimSpace(reason.String()),
					})
				case "personality":
					for _, a := range c.Arguments() {
						role.Personalities = append(role.Personalities, a.String())
					}
				case "model-tier":
					if len(role.SupportedModelTiers) != 0 {
						return nil, fmt.Errorf("role %q: duplicate model-tier", name)
					}
					args := c.Arguments()
					if len(args) == 0 {
						return nil, fmt.Errorf("role %q: model-tier needs at least one argument", name)
					}
					seenModelTiers := map[string]bool{}
					for _, arg := range args {
						modelTier := arg.String()
						if !schema.IsModelTier(modelTier) {
							return nil, fmt.Errorf("role %q: unsupported model tier %q", name, modelTier)
						}
						if seenModelTiers[modelTier] {
							return nil, fmt.Errorf("role %q repeats model tier %q", name, modelTier)
						}
						seenModelTiers[modelTier] = true
						role.SupportedModelTiers = append(role.SupportedModelTiers, modelTier)
					}
				case "copy-contract":
					if role.CopyContract != nil {
						return nil, fmt.Errorf("role %q: duplicate copy-contract", name)
					}
					scope := c.Prop("scope")
					if !scope.IsValid() || scope.String() != "tool-response" {
						return nil, fmt.Errorf("role %q: copy-contract needs supported scope tool-response", name)
					}
					contract := &CopyContract{Scope: scope.String()}
					seen := map[string]bool{}
					for _, ruleNode := range c.Children().Nodes {
						if ruleNode.Name() != "forbid" || len(ruleNode.Arguments()) != 1 {
							return nil, fmt.Errorf("role %q: copy-contract needs forbid rules", name)
						}
						forbid, err := NormalizeCue(ruleNode.Arguments()[0].String())
						if err != nil {
							return nil, fmt.Errorf("role %q copy-contract forbid: %w", name, err)
						}
						prefer := ruleNode.Prop("prefer")
						if !prefer.IsValid() || strings.TrimSpace(prefer.String()) == "" {
							return nil, fmt.Errorf("role %q: copy-contract forbid %q needs prefer", name, forbid)
						}
						if seen[forbid] {
							return nil, fmt.Errorf("role %q: copy-contract repeats forbid %q", name, forbid)
						}
						seen[forbid] = true
						contract.Rules = append(contract.Rules, CopyRule{Forbid: forbid, Prefer: strings.TrimSpace(prefer.String())})
					}
					if len(contract.Rules) == 0 {
						return nil, fmt.Errorf("role %q: copy-contract needs a rule", name)
					}
					role.CopyContract = contract
				case "agent", "seat":
					aargs := c.Arguments()
					if len(aargs) != 1 {
						return nil, fmt.Errorf("role %q: %s node needs one seat key argument", name, c.Name())
					}
					seat := Seat{Key: aargs[0].String()}
					if c.Name() == "agent" {
						seat.Harness = seat.Key
					}
					if n := c.Prop("name"); n.IsValid() {
						seat.Name = n.String()
					}
					if p := c.Prop("pronouns"); p.IsValid() {
						seat.Pronouns = p.String()
					}
					if value := c.Prop("channel"); value.IsValid() {
						seat.Channel = value.String()
					}
					if value := c.Prop("tier"); value.IsValid() {
						seat.Tier = value.String()
						if !schema.IsModelTier(seat.Tier) {
							return nil, fmt.Errorf(
								"role %q: seat %q has unsupported model tier %q",
								name,
								seat.Key,
								seat.Tier,
							)
						}
					}
					for _, existing := range role.Seats {
						if existing.Key == seat.Key {
							return nil, fmt.Errorf("role %q: duplicate seat %q", name, seat.Key)
						}
					}
					role.Seats = append(role.Seats, seat)
				case "inspiration":
					if role.Inspiration.ID != "" {
						return nil, fmt.Errorf("role %q: duplicate inspiration", name)
					}
					ref, err := parseInspirationRef(c, "role "+name)
					if err != nil {
						return nil, err
					}
					role.Inspiration = ref
				default:
					return nil, fmt.Errorf("role %q: unknown node %q", name, c.Name())
				}
			}
			if briefingSet && skillSet {
				return nil, fmt.Errorf("role %q cannot define both briefing and skill", name)
			}
			if !briefingSet && !skillSet {
				return nil, fmt.Errorf("role %q needs a role skill or legacy briefing", name)
			}
			if briefingSet {
				role.Skill = "role-" + name
			}
			for _, method := range role.Methods {
				if method == role.Skill {
					return nil, fmt.Errorf("role %q method skill %q duplicates its role skill", name, method)
				}
			}
			if briefingSet && briefingParagraphCount(role.Briefing) < 3 {
				paragraphs := briefingParagraphCount(role.Briefing)
				return nil, fmt.Errorf("role %q: briefing needs at least three paragraphs, got %d", name, paragraphs)
			}
			if len(role.Personalities) == 0 {
				return nil, fmt.Errorf("role %q needs at least one personality", name)
			}
			for index := range role.Seats {
				seat := &role.Seats[index]
				if role.Identity != nil {
					if seat.Name != "" || seat.Pronouns != "" {
						return nil, fmt.Errorf(
							"role %q: seat %q cannot redefine role identity",
							name,
							seat.Key,
						)
					}
					seat.Name = role.Identity.Name
					seat.Pronouns = role.Identity.Pronouns
					continue
				}
				if seat.Name == "" {
					return nil, fmt.Errorf("role %q: seat %q needs a name property", name, seat.Key)
				}
			}
			for _, seat := range role.Seats {
				if seat.Tier != "" && !role.SupportsModelTier(seat.Tier) {
					return nil, fmt.Errorf(
						"role %q: seat %q uses model tier %q outside the role compatibility set",
						name,
						seat.Key,
						seat.Tier,
					)
				}
			}
			seenPersonalities := map[string]bool{}
			for _, personalityName := range role.Personalities {
				if seenPersonalities[personalityName] {
					return nil, fmt.Errorf("role %q repeats personality %q", name, personalityName)
				}
				seenPersonalities[personalityName] = true
			}
			p.Roles[name] = role
			p.RoleOrder = append(p.RoleOrder, name)
		case "boundary":
			margs := n.Arguments()
			if len(margs) != 1 {
				return nil, fmt.Errorf("person boundary node needs one name argument")
			}
			boundaryName := margs[0].String()
			if !validSemanticToken(boundaryName) {
				return nil, fmt.Errorf("person boundary %q needs a stable doctrine id", boundaryName)
			}
			if _, dup := p.Boundaries[boundaryName]; dup {
				return nil, fmt.Errorf("person boundary %q declared twice", boundaryName)
			}
			boundarySkill := n.Prop("skill")
			if !boundarySkill.IsValid() {
				return nil, fmt.Errorf("boundary %q needs a skill property", boundaryName)
			}
			if boundarySkill.String() != "boundary-"+boundaryName {
				return nil, fmt.Errorf(
					"boundary %q skill %q must be boundary-%s", boundaryName, boundarySkill.String(), boundaryName,
				)
			}
			boundarySummary := n.Prop("summary")
			if !boundarySummary.IsValid() || strings.TrimSpace(boundarySummary.String()) == "" {
				return nil, fmt.Errorf("boundary %q needs a summary property", boundaryName)
			}
			if children := n.Children(); children != nil && len(children.Nodes) > 0 {
				return nil, fmt.Errorf("boundary %q accepts no child nodes", boundaryName)
			}
			ownerProp := n.Prop("owner")
			if !ownerProp.IsValid() {
				return nil, fmt.Errorf("boundary %q needs an owner property", boundaryName)
			}
			owner := strings.TrimSpace(ownerProp.String())
			if !validSemanticToken(owner) {
				return nil, fmt.Errorf("boundary %q owner needs a stable role id", boundaryName)
			}
			p.Boundaries[boundaryName] = Boundary{
				Skill:   boundarySkill.String(),
				Summary: strings.TrimSpace(boundarySummary.String()),
				Owner:   owner,
			}
			p.BoundaryOrder = append(p.BoundaryOrder, boundaryName)
		case "personality":
			pargs := n.Arguments()
			if len(pargs) != 1 {
				return nil, fmt.Errorf("person personality node needs one name argument")
			}
			name := pargs[0].String()
			if _, dup := p.Personalities[name]; dup {
				return nil, fmt.Errorf("person personality %q declared twice", name)
			}
			skill := n.Prop("skill")
			if !skill.IsValid() {
				return nil, fmt.Errorf("personality %q needs a skill property", name)
			}
			favorite := n.Prop("color")
			if !favorite.IsValid() {
				return nil, fmt.Errorf("personality %q needs a color property", name)
			}
			personality := Personality{Skill: skill.String(), Color: favorite.String()}
			if err := color.Legible(personality.Color); err != nil {
				return nil, fmt.Errorf("personality %q: %w", name, err)
			}
			motif := n.Prop("motif")
			if !motif.IsValid() || !validSemanticToken(motif.String()) {
				return nil, fmt.Errorf("personality %q needs a semantic motif property", name)
			}
			personality.Motif = motif.String()
			for _, c := range n.Children().Nodes {
				switch c.Name() {
				case "emblem":
					if personality.Emblem.Name != "" {
						return nil, fmt.Errorf("personality %q: duplicate emblem", name)
					}
					parts, err := parseSemanticParts(c, "personality "+name+" emblem", "name", "emoji", "glyph")
					if err != nil {
						return nil, err
					}
					personality.Emblem = Emblem{
						Name: parts["name"], Emoji: parts["emoji"], Glyph: parts["glyph"],
					}
				case "form":
					if personality.Form.Silhouette != "" {
						return nil, fmt.Errorf("personality %q: duplicate form", name)
					}
					parts, err := parseSemanticParts(
						c, "personality "+name+" form", "silhouette", "geometry", "motion",
					)
					if err != nil {
						return nil, err
					}
					personality.Form = Form{
						Silhouette: parts["silhouette"],
						Geometry:   parts["geometry"],
						Motion:     parts["motion"],
					}
				case "sound-mark":
					if personality.SoundMark.Timbre != "" {
						return nil, fmt.Errorf("personality %q: duplicate sound-mark", name)
					}
					parts, err := parseSemanticParts(
						c, "personality "+name+" sound-mark", "timbre", "contour", "pulse",
					)
					if err != nil {
						return nil, err
					}
					personality.SoundMark = SoundMark{
						Timbre: parts["timbre"], Contour: parts["contour"], Pulse: parts["pulse"],
					}
				case "inspiration":
					if personality.Inspiration.ID != "" {
						return nil, fmt.Errorf("personality %q: duplicate inspiration", name)
					}
					ref, err := parseInspirationRef(c, "personality "+name)
					if err != nil {
						return nil, err
					}
					personality.Inspiration = ref
				case "alias":
					args := c.Arguments()
					if len(args) != 1 {
						return nil, fmt.Errorf("personality %q alias needs one cue", name)
					}
					alias := args[0].String()
					normalized, err := NormalizeCue(alias)
					if err != nil {
						return nil, fmt.Errorf("personality %q alias %q: %w", name, alias, err)
					}
					for _, prior := range personality.Aliases {
						priorNormalized, _ := NormalizeCue(prior)
						if priorNormalized == normalized {
							return nil, fmt.Errorf("personality %q repeats alias %q", name, alias)
						}
					}
					personality.Aliases = append(personality.Aliases, alias)
				case "verb":
					args := c.Arguments()
					if len(args) == 0 {
						return nil, fmt.Errorf("personality %q verb needs at least one word", name)
					}
					for _, arg := range args {
						verb := strings.TrimSpace(arg.String())
						if verb == "" {
							return nil, fmt.Errorf("personality %q has an empty verb", name)
						}
						if slices.Contains(personality.Verbs, verb) {
							return nil, fmt.Errorf("personality %q repeats verb %q", name, verb)
						}
						personality.Verbs = append(personality.Verbs, verb)
					}
				default:
					return nil, fmt.Errorf("personality %q: unknown node %q", name, c.Name())
				}
			}
			if personality.Emblem.Name == "" {
				return nil, fmt.Errorf("personality %q needs an emblem", name)
			}
			if personality.Form.Silhouette == "" {
				return nil, fmt.Errorf("personality %q needs a form", name)
			}
			if personality.SoundMark.Timbre == "" {
				return nil, fmt.Errorf("personality %q needs a sound-mark", name)
			}
			p.Personalities[name] = personality
			p.PersonalityOrder = append(p.PersonalityOrder, name)
		case "inspiration":
			id, inspiration, err := parseInspiration(n)
			if err != nil {
				return nil, err
			}
			if _, dup := p.Inspirations[id]; dup {
				return nil, fmt.Errorf("inspiration %q declared twice", id)
			}
			p.Inspirations[id] = inspiration
			p.InspirationOrder = append(p.InspirationOrder, id)
		default:
			return nil, fmt.Errorf("person source: unknown node %q", n.Name())
		}
	}
	inspirationNames := map[string]string{}
	for _, id := range p.InspirationOrder {
		nameKey := strings.ToLower(strings.TrimSpace(p.Inspirations[id].Name))
		if existing, ok := inspirationNames[nameKey]; ok {
			return nil, fmt.Errorf("inspirations %q and %q name the same person", existing, id)
		}
		inspirationNames[nameKey] = id
	}
	referencedInspirations := map[string]bool{}
	for _, roleName := range p.RoleOrder {
		ref := p.Roles[roleName].Inspiration.ID
		if ref != "" {
			if _, ok := p.Inspirations[ref]; !ok {
				return nil, fmt.Errorf("role %q: inspiration %q has no catalog entry", roleName, ref)
			}
			referencedInspirations[ref] = true
		}
	}
	for name, personality := range p.Personalities {
		ref := personality.Inspiration.ID
		if ref == "" {
			continue
		}
		if _, ok := p.Inspirations[ref]; !ok {
			return nil, fmt.Errorf("personality %q: inspiration %q has no catalog entry", name, ref)
		}
		referencedInspirations[ref] = true
	}
	if err := validateIdentityCatalog(p.Personalities); err != nil {
		return nil, err
	}
	for _, id := range p.InspirationOrder {
		if !referencedInspirations[id] {
			return nil, fmt.Errorf("inspiration %q is not used by a role or personality", id)
		}
	}
	return p, nil
}

func briefingParagraphCount(briefing string) int {
	normalized := strings.ReplaceAll(briefing, "\r\n", "\n")
	count := 0
	for _, paragraph := range strings.Split(normalized, "\n\n") {
		if strings.TrimSpace(paragraph) != "" {
			count++
		}
	}
	return count
}

func roleSkillBodyWordCount(body string) int {
	content := strings.TrimSpace(strings.ReplaceAll(body, "\r\n", "\n"))
	firstLine, remainder, found := strings.Cut(content, "\n")
	if found && strings.HasPrefix(strings.TrimSpace(firstLine), "# ") {
		content = remainder
	}
	return len(strings.Fields(content))
}

func validLogicalID(value string) bool {
	parts := strings.Split(value, ":")
	if len(parts) == 0 {
		return false
	}
	for _, part := range parts {
		if !validSemanticToken(part) {
			return false
		}
	}
	return true
}

func parseSemanticParts(n *kdl.Node, owner string, expected ...string) (map[string]string, error) {
	allowed := make(map[string]bool, len(expected))
	for _, name := range expected {
		allowed[name] = true
	}
	parts := make(map[string]string, len(expected))
	for _, child := range n.Children().Nodes {
		name := child.Name()
		if !allowed[name] {
			return nil, fmt.Errorf("%s: unknown node %q", owner, name)
		}
		if _, duplicate := parts[name]; duplicate {
			return nil, fmt.Errorf("%s: duplicate %s", owner, name)
		}
		value, err := oneTextArgument(child, owner+" "+name)
		if err != nil {
			return nil, err
		}
		if name != "emoji" && name != "glyph" && !validSemanticToken(value) {
			return nil, fmt.Errorf("%s %s needs a lowercase semantic token", owner, name)
		}
		parts[name] = value
	}
	for _, name := range expected {
		if parts[name] == "" {
			return nil, fmt.Errorf("%s needs %s", owner, name)
		}
	}
	return parts, nil
}

func validSemanticToken(value string) bool {
	if value == "" {
		return false
	}
	for index, char := range value {
		if char >= 'a' && char <= 'z' || char >= '0' && char <= '9' {
			continue
		}
		if char == '-' && index > 0 && index < len(value)-1 {
			continue
		}
		return false
	}
	return true
}

func validateIdentityCatalog(catalog map[string]Personality) error {
	emblems := map[string]string{}
	emojis := map[string]string{}
	glyphs := map[string]string{}
	motifs := map[string]string{}
	for name, personality := range catalog {
		for value, owners := range map[string]map[string]string{
			personality.Emblem.Name:  emblems,
			personality.Emblem.Emoji: emojis,
			personality.Emblem.Glyph: glyphs,
			personality.Motif:        motifs,
		} {
			if owner, duplicate := owners[value]; duplicate {
				return fmt.Errorf("personalities %q and %q share identity value %q", owner, name, value)
			}
			owners[value] = name
		}
	}
	return nil
}

func parseInspirationRef(n *kdl.Node, owner string) (InspirationRef, error) {
	args := n.Arguments()
	if len(args) != 1 || strings.TrimSpace(args[0].String()) == "" {
		return InspirationRef{}, fmt.Errorf("%s inspiration needs one catalog id", owner)
	}
	ref := InspirationRef{ID: strings.TrimSpace(args[0].String())}
	for _, c := range n.Children().Nodes {
		if c.Name() != "fit" {
			return InspirationRef{}, fmt.Errorf("%s inspiration: unknown node %q", owner, c.Name())
		}
		if ref.Fit != "" {
			return InspirationRef{}, fmt.Errorf("%s inspiration: duplicate fit", owner)
		}
		text, err := oneTextArgument(c, owner+" inspiration fit")
		if err != nil {
			return InspirationRef{}, err
		}
		ref.Fit = text
	}
	if ref.Fit == "" {
		return InspirationRef{}, fmt.Errorf("%s inspiration needs a fit", owner)
	}
	return ref, nil
}

func parseInspiration(n *kdl.Node) (string, Inspiration, error) {
	args := n.Arguments()
	if len(args) != 1 || strings.TrimSpace(args[0].String()) == "" {
		return "", Inspiration{}, fmt.Errorf("inspiration node needs one catalog id")
	}
	id := strings.TrimSpace(args[0].String())
	name, err := requiredStringProp(n, "name", "inspiration "+id)
	if err != nil {
		return "", Inspiration{}, err
	}
	profile, err := requiredStringProp(n, "profile-citation", "inspiration "+id)
	if err != nil {
		return "", Inspiration{}, err
	}
	impactMode, err := requiredStringProp(n, "impact-mode", "inspiration "+id)
	if err != nil {
		return "", Inspiration{}, err
	}
	inspiration := Inspiration{
		Name:            name,
		ProfileCitation: profile,
		ImpactMode:      impactMode,
	}
	for _, c := range n.Children().Nodes {
		switch c.Name() {
		case "achievement":
			if inspiration.Achievement != "" {
				return "", Inspiration{}, fmt.Errorf("inspiration %q: duplicate achievement", id)
			}
			inspiration.Achievement, err = oneTextArgument(c, "inspiration "+id+" achievement")
		case "impact-fit":
			if inspiration.ImpactFit != "" {
				return "", Inspiration{}, fmt.Errorf("inspiration %q: duplicate impact-fit", id)
			}
			inspiration.ImpactFit, err = oneTextArgument(c, "inspiration "+id+" impact-fit")
		case "appearance":
			if inspiration.Appearance.ID != "" {
				return "", Inspiration{}, fmt.Errorf("inspiration %q: duplicate appearance", id)
			}
			inspiration.Appearance, err = parseAppearance(c, id)
		default:
			return "", Inspiration{}, fmt.Errorf("inspiration %q: unknown node %q", id, c.Name())
		}
		if err != nil {
			return "", Inspiration{}, err
		}
	}
	if inspiration.Achievement == "" {
		return "", Inspiration{}, fmt.Errorf("inspiration %q needs an achievement", id)
	}
	if inspiration.ImpactFit == "" {
		return "", Inspiration{}, fmt.Errorf("inspiration %q needs an impact-fit", id)
	}
	if inspiration.Appearance.ID == "" {
		return "", Inspiration{}, fmt.Errorf("inspiration %q needs an appearance", id)
	}
	return id, inspiration, nil
}

func parseAppearance(n *kdl.Node, inspirationID string) (Appearance, error) {
	args := n.Arguments()
	if len(args) != 1 || strings.TrimSpace(args[0].String()) == "" {
		return Appearance{}, fmt.Errorf("inspiration %q appearance needs one id", inspirationID)
	}
	appearance := Appearance{ID: strings.TrimSpace(args[0].String())}
	owner := "appearance " + appearance.ID
	var err error
	appearance.Title, err = requiredStringProp(n, "title", owner)
	if err != nil {
		return Appearance{}, err
	}
	appearance.Event, err = requiredStringProp(n, "event", owner)
	if err != nil {
		return Appearance{}, err
	}
	appearance.Year, err = requiredStringProp(n, "year", owner)
	if err != nil {
		return Appearance{}, err
	}
	appearance.Format, err = requiredStringProp(n, "format", owner)
	if err != nil {
		return Appearance{}, err
	}
	for _, c := range n.Children().Nodes {
		switch c.Name() {
		case "summary":
			if appearance.Summary != "" {
				return Appearance{}, fmt.Errorf("appearance %q: duplicate summary", appearance.ID)
			}
			appearance.Summary, err = oneTextArgument(c, "appearance "+appearance.ID+" summary")
		case "citation":
			citation, citationErr := oneTextArgument(c, "appearance "+appearance.ID+" citation")
			if citationErr != nil {
				return Appearance{}, citationErr
			}
			for _, existing := range appearance.Citations {
				if existing == citation {
					return Appearance{}, fmt.Errorf("appearance %q repeats citation %q", appearance.ID, citation)
				}
			}
			appearance.Citations = append(appearance.Citations, citation)
		default:
			return Appearance{}, fmt.Errorf("appearance %q: unknown node %q", appearance.ID, c.Name())
		}
		if err != nil {
			return Appearance{}, err
		}
	}
	if appearance.Summary == "" {
		return Appearance{}, fmt.Errorf("appearance %q needs a summary", appearance.ID)
	}
	if len(appearance.Citations) == 0 {
		return Appearance{}, fmt.Errorf("appearance %q needs at least one citation", appearance.ID)
	}
	return appearance, nil
}

func requiredStringProp(n *kdl.Node, prop, owner string) (string, error) {
	value := n.Prop(prop)
	if !value.IsValid() || strings.TrimSpace(value.String()) == "" {
		return "", fmt.Errorf("%s needs a %s property", owner, prop)
	}
	return strings.TrimSpace(value.String()), nil
}

func oneTextArgument(n *kdl.Node, owner string) (string, error) {
	args := n.Arguments()
	if len(args) != 1 || strings.TrimSpace(args[0].String()) == "" {
		return "", fmt.Errorf("%s needs one non-empty argument", owner)
	}
	return strings.TrimSpace(args[0].String()), nil
}
