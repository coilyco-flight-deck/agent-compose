package schema

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	kdl "github.com/calico32/kdl-go"
)

const (
	DeliveryNativeSkills = "native-skills"
	DeliveryCompiled     = "compiled"
	DensityBrief         = "brief"
	DensityFull          = "full"
)

type Request struct {
	Role        string
	Personality string
	Delivery    string
	Density     string
	Sources     []SourceLocator
}

type SourceLocator struct {
	ID          string
	Declaration string
	Required    bool
}

type ContentRef struct {
	ID   string
	Path string
}

type Source struct {
	ID           string
	Root         string
	Declaration  []byte
	Instructions []ContentRef
	Skills       []ContentRef
}

// MissingSource records an optional source whose declaration was absent, so
// the resolver can note the exclusion in the trace.
type MissingSource struct {
	ID     string
	Reason string
}

func ParseRequest(path string) (*Request, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read request %s: %w", path, err)
	}
	doc, err := kdl.ParseString(string(raw))
	if err != nil {
		return nil, fmt.Errorf("parse request %s: %w", path, err)
	}
	if len(doc.Nodes) != 1 || doc.Nodes[0].Name() != "compose" {
		return nil, fmt.Errorf("request %s: expected exactly one top-level compose node", path)
	}

	req := &Request{}
	seen := map[string]bool{}
	for _, n := range doc.Nodes[0].Children().Nodes {
		switch n.Name() {
		case "role", "personality", "delivery", "density":
			if seen[n.Name()] {
				return nil, fmt.Errorf("request %s: duplicate %s node", path, n.Name())
			}
			seen[n.Name()] = true
			v, err := oneStringArg(n)
			if err != nil {
				return nil, fmt.Errorf("request %s: %w", path, err)
			}
			switch n.Name() {
			case "role":
				req.Role = v
			case "personality":
				req.Personality = v
			case "delivery":
				req.Delivery = v
			case "density":
				req.Density = v
			}
		case "source":
			id, err := oneStringArg(n)
			if err != nil {
				return nil, fmt.Errorf("request %s: %w", path, err)
			}
			decl := n.Prop("declaration")
			if !decl.IsValid() {
				return nil, fmt.Errorf("request %s: source %q needs a declaration property", path, id)
			}
			loc := SourceLocator{ID: id, Declaration: decl.String()}
			if p := n.Prop("required"); p.IsValid() {
				loc.Required = p.Bool()
			}
			for _, other := range req.Sources {
				if other.ID == loc.ID {
					return nil, fmt.Errorf("request %s: duplicate source id %q", path, loc.ID)
				}
			}
			req.Sources = append(req.Sources, loc)
		default:
			return nil, fmt.Errorf("request %s: unknown node %q", path, n.Name())
		}
	}

	for _, field := range []struct{ name, value string }{
		{"role", req.Role}, {"personality", req.Personality},
		{"delivery", req.Delivery}, {"density", req.Density},
	} {
		if field.value == "" {
			return nil, fmt.Errorf("request %s: missing %s", path, field.name)
		}
	}
	if req.Delivery != DeliveryNativeSkills && req.Delivery != DeliveryCompiled {
		return nil, fmt.Errorf("request %s: delivery must be %q or %q, got %q",
			path, DeliveryNativeSkills, DeliveryCompiled, req.Delivery)
	}
	if req.Density != DensityBrief && req.Density != DensityFull {
		return nil, fmt.Errorf("request %s: density must be %q or %q, got %q",
			path, DensityBrief, DensityFull, req.Density)
	}
	return req, nil
}

// LoadSources reads each declared source relative to the request file. A
// missing required source fails; a missing optional one returns for the trace.
func LoadSources(req *Request, requestPath string) ([]*Source, []MissingSource, error) {
	baseDir := filepath.Dir(requestPath)
	var sources []*Source
	var missing []MissingSource
	for _, loc := range req.Sources {
		declPath := filepath.Join(baseDir, loc.Declaration)
		if strings.Contains(loc.Declaration, "..") || filepath.IsAbs(loc.Declaration) {
			return nil, nil, fmt.Errorf("source %q: declaration path %q must be relative and clean", loc.ID, loc.Declaration)
		}
		src, err := parseSource(declPath)
		if err != nil {
			if os.IsNotExist(err) && !loc.Required {
				missing = append(missing, MissingSource{ID: loc.ID, Reason: fmt.Sprintf("optional source declaration %s is absent", loc.Declaration)})
				continue
			}
			return nil, nil, fmt.Errorf("source %q: %w", loc.ID, err)
		}
		if src.ID != loc.ID {
			return nil, nil, fmt.Errorf("source %q: declaration %s declares id %q", loc.ID, loc.Declaration, src.ID)
		}
		sources = append(sources, src)
	}
	return sources, missing, nil
}

func parseSource(declPath string) (*Source, error) {
	raw, err := os.ReadFile(declPath)
	if err != nil {
		return nil, err
	}
	doc, err := kdl.ParseString(string(raw))
	if err != nil {
		return nil, fmt.Errorf("parse %s: %w", declPath, err)
	}
	if len(doc.Nodes) != 1 || doc.Nodes[0].Name() != "source" {
		return nil, fmt.Errorf("%s: expected exactly one top-level source node", declPath)
	}
	root := doc.Nodes[0]
	id, err := oneStringArg(root)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", declPath, err)
	}
	src := &Source{ID: id, Root: filepath.Dir(declPath), Declaration: raw}
	for _, n := range root.Children().Nodes {
		switch n.Name() {
		case "instruction", "skill":
			cid, err := oneStringArg(n)
			if err != nil {
				return nil, fmt.Errorf("%s: %w", declPath, err)
			}
			p := n.Prop("path")
			if !p.IsValid() {
				return nil, fmt.Errorf("%s: %s %q needs a path property", declPath, n.Name(), cid)
			}
			rel := p.String()
			if strings.Contains(rel, "..") || filepath.IsAbs(rel) {
				return nil, fmt.Errorf("%s: %s %q path %q must stay beneath the source root", declPath, n.Name(), cid, rel)
			}
			ref := ContentRef{ID: cid, Path: rel}
			if n.Name() == "instruction" {
				src.Instructions = append(src.Instructions, ref)
			} else {
				src.Skills = append(src.Skills, ref)
			}
		default:
			return nil, fmt.Errorf("%s: unknown node %q", declPath, n.Name())
		}
	}
	return src, nil
}

func oneStringArg(n *kdl.Node) (string, error) {
	args := n.Arguments()
	if len(args) != 1 {
		return "", fmt.Errorf("node %q: expected exactly one argument", n.Name())
	}
	v := args[0].String()
	if v == "" {
		return "", fmt.Errorf("node %q: argument must be a non-empty string", n.Name())
	}
	return v, nil
}
