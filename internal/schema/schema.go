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

	providerSkillsPath    = ".agents/skills"
	providerInvariantID   = "personality-invariant"
	providerInvariantPath = ".agents/skills/personality-shared/INVARIANT.md"
)

type Request struct {
	Role     string
	Delivery string
	Density  string
	Sources  []SourceLocator
}

type SourceLocator struct {
	ID          string
	Declaration string
	Root        string
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
		case "role", "delivery", "density":
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
			root := n.Prop("root")
			if decl.IsValid() == root.IsValid() {
				return nil, fmt.Errorf("request %s: source %q needs exactly one of declaration or root", path, id)
			}
			loc := SourceLocator{ID: id}
			if decl.IsValid() {
				loc.Declaration = decl.String()
				if loc.Declaration == "" {
					return nil, fmt.Errorf("request %s: source %q declaration cannot be empty", path, id)
				}
			} else {
				loc.Root = root.String()
				if loc.Root == "" {
					return nil, fmt.Errorf("request %s: source %q root cannot be empty", path, id)
				}
			}
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
		{"role", req.Role}, {"delivery", req.Delivery}, {"density", req.Density},
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
		sourcePath := loc.Declaration
		kind := "declaration"
		if loc.Root != "" {
			sourcePath = loc.Root
			kind = "root"
		}
		if strings.Contains(sourcePath, "..") || filepath.IsAbs(sourcePath) {
			return nil, nil, fmt.Errorf("source %q: %s path %q must be relative and clean", loc.ID, kind, sourcePath)
		}
		resolvedPath := filepath.Join(baseDir, sourcePath)
		var src *Source
		var err error
		if loc.Root != "" {
			src, err = inferProvider(loc.ID, resolvedPath)
		} else {
			src, err = parseSource(resolvedPath)
		}
		if err != nil {
			_, rootErr := os.Stat(resolvedPath)
			if os.IsNotExist(rootErr) && !loc.Required {
				missing = append(missing, MissingSource{
					ID:     loc.ID,
					Reason: fmt.Sprintf("optional source %s %s is absent", kind, sourcePath),
				})
				continue
			}
			return nil, nil, fmt.Errorf("source %q: %w", loc.ID, err)
		}
		if loc.Declaration != "" && src.ID != loc.ID {
			return nil, nil, fmt.Errorf("source %q: declaration %s declares id %q", loc.ID, loc.Declaration, src.ID)
		}
		sources = append(sources, src)
	}
	return sources, missing, nil
}

// LoadSource reads one explicit declaration or infers an AOS personality
// provider root for direct consumers such as roster.
func LoadSource(sourcePath string) (*Source, error) {
	info, err := os.Stat(sourcePath)
	if err != nil {
		return nil, err
	}
	if info.IsDir() {
		root, err := filepath.Abs(sourcePath)
		if err != nil {
			return nil, err
		}
		return inferProvider(filepath.Base(root), root)
	}
	return parseSource(sourcePath)
}

// inferProvider applies the AOS personality-provider filesystem convention.
// ReadDir returns names in lexical order, keeping source decisions stable.
func inferProvider(id, root string) (*Source, error) {
	info, err := os.Stat(root)
	if err != nil {
		return nil, err
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("provider root %s is not a directory", root)
	}

	invariant := filepath.Join(root, filepath.FromSlash(providerInvariantPath))
	if info, err := os.Stat(invariant); err != nil {
		return nil, fmt.Errorf("provider invariant %s: %w", providerInvariantPath, err)
	} else if !info.Mode().IsRegular() {
		return nil, fmt.Errorf("provider invariant %s is not a regular file", providerInvariantPath)
	}

	skillsRoot := filepath.Join(root, filepath.FromSlash(providerSkillsPath))
	entries, err := os.ReadDir(skillsRoot)
	if err != nil {
		return nil, fmt.Errorf("provider skills %s: %w", providerSkillsPath, err)
	}
	src := &Source{
		ID:   id,
		Root: root,
		Instructions: []ContentRef{{
			ID:   providerInvariantID,
			Path: providerInvariantPath,
		}},
	}
	for _, entry := range entries {
		name := entry.Name()
		if !entry.IsDir() || !strings.HasPrefix(name, "personality-") || name == "personality-shared" {
			continue
		}
		skillPath := filepath.Join(skillsRoot, name, "SKILL.md")
		if info, err := os.Stat(skillPath); err != nil {
			return nil, fmt.Errorf("provider skill %s: %w", name, err)
		} else if !info.Mode().IsRegular() {
			return nil, fmt.Errorf("provider skill %s SKILL.md is not a regular file", name)
		}
		src.Skills = append(src.Skills, ContentRef{
			ID:   name,
			Path: filepath.ToSlash(filepath.Join(providerSkillsPath, name)),
		})
	}
	if len(src.Skills) == 0 {
		return nil, fmt.Errorf("provider root %s has no personality-* skills under %s", root, providerSkillsPath)
	}
	return src, nil
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
