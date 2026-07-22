// Package person embeds the canonical public-safe person source: the role
// catalog and named agent seats adapted from ward's roles.kdl sketch. The
// personality catalog beyond engineer lands with issue #10.
package person

import (
	_ "embed"
	"fmt"

	kdl "github.com/calico32/kdl-go"
)

//go:embed person.kdl
var embedded []byte

// Seat is one named agent identity within a role. The harness is the join
// key against the launcher's own catalog; the name is an opaque string here.
type Seat struct {
	Harness  string
	Name     string
	Pronouns string
}

type Role struct {
	Purpose       string
	Personalities []string
	Seats         []Seat
}

type Person struct {
	Name          string
	Roles         map[string]Role
	Personalities map[string]string
	Raw           []byte
}

func Load() (*Person, error) {
	return parse(embedded)
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
		Personalities: map[string]string{},
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
			p.Roles[name] = role
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
			p.Personalities[name] = skill.String()
		default:
			return nil, fmt.Errorf("embedded person source: unknown node %q", n.Name())
		}
	}
	return p, nil
}
