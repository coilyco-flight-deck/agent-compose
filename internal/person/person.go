// Package person embeds the canonical public-safe person source: the role
// catalog, named agent seats, personality invariant, and complete personality
// definitions.
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

// Seat is one named agent identity within a role. The harness is the join
// key against the launcher's own catalog; the name is an opaque string here.
type Seat struct {
	Harness  string
	Name     string
	Pronouns string
}

type Role struct {
	Purpose       string
	Briefing      string
	Personalities []string
	Seats         []Seat
}

// Personality binds a name to the skill defining it and a terminal-legible
// favorite color.
type Personality struct {
	Skill string
	Color string
}

type Person struct {
	Name          string
	Roles         map[string]Role
	RoleOrder     []string
	Personalities map[string]Personality
	Raw           []byte
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
				default:
					return nil, fmt.Errorf("role %q: unknown node %q", name, c.Name())
				}
			}
			if !briefingSet {
				return nil, fmt.Errorf("role %q needs a briefing", name)
			}
			if len(role.Personalities) < 2 || len(role.Personalities) > 3 {
				return nil, fmt.Errorf("role %q needs two or three personalities, got %d", name, len(role.Personalities))
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
			p.Personalities[name] = personality
		default:
			return nil, fmt.Errorf("embedded person source: unknown node %q", n.Name())
		}
	}
	for _, roleName := range p.RoleOrder {
		for _, personalityName := range p.Roles[roleName].Personalities {
			if _, ok := p.Personalities[personalityName]; !ok {
				return nil, fmt.Errorf("role %q: personality %q has no catalog binding", roleName, personalityName)
			}
		}
	}
	return p, nil
}
