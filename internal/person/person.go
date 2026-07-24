// Package person embeds the canonical public-safe roles, agent seats,
// personality definitions, invariant, and credited inspirations.
package person

import (
	"bytes"
	"embed"
	"fmt"
	"io/fs"
	"sort"
	"strings"

	kdl "github.com/calico32/kdl-go"

	"forgejo.coilysiren.me/coilyco-flight-deck/agent-compose/internal/color"
	"forgejo.coilysiren.me/coilyco-flight-deck/agent-compose/internal/schema"
)

//go:embed person.kdl definitions
var embedded embed.FS

// Seat is one named agent identity within a role. The harness joins the
// launcher's own catalog, while the name remains opaque here.
type Seat struct {
	Harness  string `json:"harness"`
	Name     string `json:"name"`
	Pronouns string `json:"pronouns"`
}

// InspirationRef records why one role or personality cites a catalog entry.
// Inspirations are acknowledgements and evidence, not identities to imitate.
type InspirationRef struct {
	ID  string `json:"id"`
	Fit string `json:"fit"`
}

type Role struct {
	Purpose       string         `json:"purpose"`
	Briefing      string         `json:"briefing"`
	Personalities []string       `json:"personalities"`
	Seats         []Seat         `json:"seats"`
	Inspiration   InspirationRef `json:"inspiration"`
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

// Personality binds the definition, visual and sensory identity primitives,
// favorite color, and credited inspiration for one canonical personality.
type Personality struct {
	Skill       string         `json:"skill"`
	Color       string         `json:"color"`
	Motif       string         `json:"motif"`
	Emblem      Emblem         `json:"emblem"`
	Form        Form           `json:"form"`
	SoundMark   SoundMark      `json:"sound_mark"`
	Inspiration InspirationRef `json:"inspiration"`
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
	Name             string                 `json:"person"`
	Roles            map[string]Role        `json:"roles"`
	RoleOrder        []string               `json:"role_order"`
	Personalities    map[string]Personality `json:"personalities"`
	Inspirations     map[string]Inspiration `json:"inspirations"`
	InspirationOrder []string               `json:"inspiration_order"`
	Raw              []byte                 `json:"-"`
}

func Load() (*Person, error) {
	raw, err := fs.ReadFile(embedded, "person.kdl")
	if err != nil {
		return nil, fmt.Errorf("read embedded person source: %w", err)
	}
	return parse(raw)
}

// Source returns the canonical person:kai content provider. The role catalog,
// invariant, and every bound personality definition ship in one binary.
func Source(p *Person) (*schema.Source, error) {
	files, err := fs.Sub(embedded, "definitions")
	if err != nil {
		return nil, fmt.Errorf("open embedded personality definitions: %w", err)
	}
	invariant, err := fs.ReadFile(files, "INVARIANT.md")
	if err != nil {
		return nil, fmt.Errorf("read embedded personality invariant: %w", err)
	}
	if len(bytes.TrimSpace(invariant)) == 0 {
		return nil, fmt.Errorf("embedded personality invariant is empty")
	}

	canonical, err := Load()
	if err != nil {
		return nil, err
	}
	expected := make(map[string]bool, len(canonical.Personalities))
	canonicalSkills := make([]string, 0, len(canonical.Personalities))
	for _, binding := range canonical.Personalities {
		expected[binding.Skill] = true
		canonicalSkills = append(canonicalSkills, binding.Skill)
	}
	sort.Strings(canonicalSkills)

	selected := make(map[string]bool, len(p.Personalities))
	for name, binding := range p.Personalities {
		if binding.Skill == "" {
			return nil, fmt.Errorf("personality %q has no skill binding", name)
		}
		if !expected[binding.Skill] {
			return nil, fmt.Errorf("personality %q binds non-canonical skill %q", name, binding.Skill)
		}
		if selected[binding.Skill] {
			return nil, fmt.Errorf("personality skill %q is bound more than once", binding.Skill)
		}
		selected[binding.Skill] = true
	}

	entries, err := fs.ReadDir(files, "skills")
	if err != nil {
		return nil, fmt.Errorf("read embedded personality skills: %w", err)
	}
	if len(entries) != len(canonicalSkills) {
		return nil, fmt.Errorf(
			"embedded personality skills: catalog binds %d skills but definitions contain %d entries",
			len(canonicalSkills),
			len(entries),
		)
	}

	src := &schema.Source{
		ID:    "person:" + p.Name,
		Files: files,
		Instructions: []schema.ContentRef{{
			ID:   "personality-invariant",
			Path: "INVARIANT.md",
		}},
	}
	for _, entry := range entries {
		if !entry.IsDir() || !expected[entry.Name()] {
			return nil, fmt.Errorf("embedded personality skills: unexpected entry %q", entry.Name())
		}
	}
	for _, skill := range canonicalSkills {
		skillPath := "skills/" + skill
		raw, err := fs.ReadFile(files, skillPath+"/SKILL.md")
		if err != nil {
			return nil, fmt.Errorf("read embedded skill %q: %w", skill, err)
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
	return src, nil
}

func validateSkillDefinition(skill string, raw []byte) error {
	text := string(raw)
	if !strings.HasPrefix(text, "---\n") {
		return fmt.Errorf("embedded skill %q: SKILL.md needs YAML frontmatter", skill)
	}
	end := strings.Index(text[4:], "\n---\n")
	if end < 0 {
		return fmt.Errorf("embedded skill %q: SKILL.md has unterminated frontmatter", skill)
	}
	end += 4
	frontmatter := "\n" + text[4:end] + "\n"
	if !strings.Contains(frontmatter, "\nname: "+skill+"\n") {
		return fmt.Errorf("embedded skill %q: frontmatter name does not match", skill)
	}
	if !strings.Contains(frontmatter, "\ndescription: ") {
		return fmt.Errorf("embedded skill %q: frontmatter needs a description", skill)
	}
	if strings.TrimSpace(text[end+5:]) == "" {
		return fmt.Errorf("embedded skill %q: SKILL.md body is empty", skill)
	}
	return nil
}

func parse(raw []byte) (*Person, error) {
	doc, err := kdl.ParseString(string(raw))
	if err != nil {
		return nil, fmt.Errorf("parse embedded person source: %w", err)
	}
	if len(doc.Nodes) != 1 || doc.Nodes[0].Name() != "person" {
		return nil, fmt.Errorf("embedded person source: expected exactly one top-level person node")
	}
	root := doc.Nodes[0]
	args := root.Arguments()
	if len(args) != 1 {
		return nil, fmt.Errorf("embedded person source: person node needs one name argument")
	}
	p := &Person{
		Name:          args[0].String(),
		Roles:         map[string]Role{},
		Personalities: map[string]Personality{},
		Inspirations:  map[string]Inspiration{},
		Raw:           raw,
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
			for _, c := range n.Children().Nodes {
				switch c.Name() {
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
				case "personality":
					for _, a := range c.Arguments() {
						role.Personalities = append(role.Personalities, a.String())
					}
				case "agent":
					aargs := c.Arguments()
					if len(aargs) != 1 {
						return nil, fmt.Errorf("role %q: agent node needs one harness argument", name)
					}
					seat := Seat{Harness: aargs[0].String()}
					if n := c.Prop("name"); n.IsValid() {
						seat.Name = n.String()
					}
					if seat.Name == "" {
						return nil, fmt.Errorf("role %q: agent %q needs a name property", name, seat.Harness)
					}
					if p := c.Prop("pronouns"); p.IsValid() {
						seat.Pronouns = p.String()
					}
					for _, existing := range role.Seats {
						if existing.Harness == seat.Harness {
							return nil, fmt.Errorf("role %q: duplicate agent seat for %q", name, seat.Harness)
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
			if !briefingSet {
				return nil, fmt.Errorf("role %q needs a briefing", name)
			}
			if paragraphs := briefingParagraphCount(role.Briefing); paragraphs < 3 {
				return nil, fmt.Errorf("role %q: briefing needs at least three paragraphs, got %d", name, paragraphs)
			}
			if len(role.Personalities) < 2 || len(role.Personalities) > 3 {
				return nil, fmt.Errorf("role %q needs two or three personalities, got %d", name, len(role.Personalities))
			}
			if role.Inspiration.ID == "" {
				return nil, fmt.Errorf("role %q needs an inspiration", name)
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
			if personality.Inspiration.ID == "" {
				return nil, fmt.Errorf("personality %q needs an inspiration", name)
			}
			p.Personalities[name] = personality
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
			return nil, fmt.Errorf("embedded person source: unknown node %q", n.Name())
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
		if _, ok := p.Inspirations[ref]; !ok {
			return nil, fmt.Errorf("role %q: inspiration %q has no catalog entry", roleName, ref)
		}
		referencedInspirations[ref] = true
		for _, personalityName := range p.Roles[roleName].Personalities {
			if _, ok := p.Personalities[personalityName]; !ok {
				return nil, fmt.Errorf("role %q: personality %q has no catalog binding", roleName, personalityName)
			}
		}
	}
	for name, personality := range p.Personalities {
		ref := personality.Inspiration.ID
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
