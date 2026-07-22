// Package person embeds the canonical public-safe person source. Issue #10
// replaces this fixture-grade roster with the complete one; the contract
// stays as docs/person-contract.md defines it.
package person

import (
	_ "embed"
	"fmt"

	kdl "github.com/calico32/kdl-go"
)

//go:embed person.kdl
var embedded []byte

type Role struct {
	Purpose       string
	Personalities []string
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
