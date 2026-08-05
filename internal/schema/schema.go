package schema

import (
	"fmt"
	"io/fs"
	"os"
	pathpkg "path"
	"path/filepath"
	"sort"
	"strings"

	kdl "github.com/calico32/kdl-go"

	"forgejo.coilysiren.me/coilyco-flight-deck/agent-compose/internal/personpolicy"
	"forgejo.coilysiren.me/coilyco-flight-deck/agent-compose/internal/skillselector"
)

const (
	DeliveryNativeSkills = "native-skills"
	DeliveryCompiled     = "compiled"
	ModelClassFrontier   = "frontier"
	ModelClassLowContext = "low-context"
	LowContextRequired   = "required"
	LowContextOptional   = "optional"
	legacyDensityFull    = "full"

	providerSkillsPath    = ".agents/skills"
	providerComposedPath  = ".agents/composed"
	providerRolesPath     = ".agents/roles.kdl"
	providerInvariantID   = "personality-invariant"
	providerInvariantPath = ".agents/skills/personality-shared/INVARIANT.md"

	ProviderScopePerson  = "person"
	ProviderScopeRequest = "request"
	ProviderScopeDefault = "default"
	ProviderScopeHarness = "harness"
	ProviderScopeRole    = "role"
)

type Request struct {
	PersonPolicy         string
	PersonSource         string
	PersonalityLibraries []string
	Role                 string
	Delivery             string
	ModelClass           string
	Sources              []SourceLocator
}

type SourceLocator struct {
	ID          string
	Declaration string
	Root        string
	Required    bool
}

type ContentRef struct {
	ID         string
	Path       string
	EntryPoint string
}

type Source struct {
	ID              string
	Root            string
	Files           fs.FS
	Declaration     []byte
	Instructions    []ContentRef
	Skills          []ContentRef
	RoleSkills      map[string][]ContentRef
	AdmissionReason string
	ProviderScope   string
	ExcludedSkills  []ContentRef
	SelectorReason  string
}

// SelectOrdinarySkills applies a role-provider selector after LoadSource has
// validated the provider's complete ordinary and composed catalogues.
func SelectOrdinarySkills(source *Source, patterns []string) error {
	if patterns == nil {
		return nil
	}
	ids := make([]string, 0, len(source.Skills))
	refs := make(map[string]ContentRef, len(source.Skills))
	for _, ref := range source.Skills {
		ids = append(ids, ref.ID)
		refs[ref.ID] = ref
	}
	selected, excluded, err := skillselector.Select(patterns, ids)
	if err != nil {
		return err
	}
	source.Skills = source.Skills[:0]
	for _, id := range selected {
		source.Skills = append(source.Skills, refs[id])
	}
	for _, id := range excluded {
		source.ExcludedSkills = append(source.ExcludedSkills, refs[id])
	}
	source.SelectorReason = fmt.Sprintf(
		"ordinary skill selector %s admitted %d of %d catalogue skills",
		strings.Join(patterns, ", "),
		len(selected),
		len(ids),
	)
	return nil
}

// FileSystem returns embedded content for shipped sources and disk content for
// declared or inferred providers.
func (s *Source) FileSystem() fs.FS {
	if s.Files != nil {
		return s.Files
	}
	return os.DirFS(s.Root)
}

func (s *Source) ReadFile(name string) ([]byte, error) {
	return fs.ReadFile(s.FileSystem(), filepath.ToSlash(name))
}

// LowContextPolicy reads the optional top-level skill frontmatter key. Skills
// stay required by default so older providers fail open toward capability.
func (s *Source) LowContextPolicy(ref ContentRef) (string, error) {
	entryPoint := ref.EntryPoint
	if entryPoint == "" {
		entryPoint = "SKILL.md"
	}
	path := filepath.ToSlash(filepath.Join(ref.Path, entryPoint))
	raw, err := s.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("read low-context policy from %s: %w", path, err)
	}
	lines := strings.Split(strings.ReplaceAll(string(raw), "\r\n", "\n"), "\n")
	if len(lines) == 0 || lines[0] != "---" {
		return LowContextRequired, nil
	}
	policy := LowContextRequired
	seen := false
	closed := false
	for _, line := range lines[1:] {
		if line == "---" {
			closed = true
			break
		}
		if !strings.HasPrefix(line, "low-context:") {
			continue
		}
		if seen {
			return "", fmt.Errorf("skill %q repeats low-context frontmatter", ref.ID)
		}
		seen = true
		policy = strings.TrimSpace(strings.TrimPrefix(line, "low-context:"))
		if policy != LowContextRequired && policy != LowContextOptional {
			return "", fmt.Errorf(
				"skill %q low-context must be %q or %q, got %q",
				ref.ID, LowContextRequired, LowContextOptional, policy,
			)
		}
	}
	if !closed {
		return "", fmt.Errorf("skill %q has unterminated YAML frontmatter", ref.ID)
	}
	return policy, nil
}

// MissingSource records an optional source whose declaration was absent, so
// the resolver can note the exclusion in the trace.
type MissingSource struct {
	ID            string
	Reason        string
	Skills        []string
	ProviderScope string
	Warning       bool
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
		case "personality-library":
			v, err := oneStringArg(n)
			if err != nil {
				return nil, fmt.Errorf("request %s: %w", path, err)
			}
			if filepath.IsAbs(v) || strings.HasPrefix(v, "/") || strings.HasPrefix(v, `\`) {
				return nil, fmt.Errorf("request %s: personality-library path %q must be relative and clean", path, v)
			}
			req.PersonalityLibraries = append(req.PersonalityLibraries, v)
		case "person-policy", "person-source", "role", "delivery", "model-class":
			if seen[n.Name()] {
				return nil, fmt.Errorf("request %s: duplicate %s node", path, n.Name())
			}
			seen[n.Name()] = true
			v, err := oneStringArg(n)
			if err != nil {
				return nil, fmt.Errorf("request %s: %w", path, err)
			}
			switch n.Name() {
			case "person-policy":
				req.PersonPolicy = v
			case "person-source":
				if strings.Contains(v, "..") || filepath.IsAbs(v) ||
					strings.HasPrefix(v, "/") || strings.HasPrefix(v, `\`) {
					return nil, fmt.Errorf(
						"request %s: person-source path %q must be relative and clean",
						path,
						v,
					)
				}
				req.PersonSource = v
			case "role":
				req.Role = v
			case "delivery":
				req.Delivery = v
			case "model-class":
				req.ModelClass = v
			}
		case "density":
			if seen[n.Name()] {
				return nil, fmt.Errorf("request %s: duplicate density node", path)
			}
			seen[n.Name()] = true
			v, err := oneStringArg(n)
			if err != nil {
				return nil, fmt.Errorf("request %s: %w", path, err)
			}
			if v != legacyDensityFull {
				return nil, fmt.Errorf(
					"request %s: density is removed; delete the node (only legacy %q is accepted)",
					path,
					legacyDensityFull,
				)
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
		{"role", req.Role}, {"delivery", req.Delivery},
	} {
		if field.value == "" {
			return nil, fmt.Errorf("request %s: missing %s", path, field.name)
		}
	}
	if req.Delivery != DeliveryNativeSkills && req.Delivery != DeliveryCompiled {
		return nil, fmt.Errorf("request %s: delivery must be %q or %q, got %q",
			path, DeliveryNativeSkills, DeliveryCompiled, req.Delivery)
	}
	if req.ModelClass == "" {
		req.ModelClass = ModelClassFrontier
	}
	if req.ModelClass != ModelClassFrontier && req.ModelClass != ModelClassLowContext {
		return nil, fmt.Errorf("request %s: model-class must be %q or %q, got %q",
			path, ModelClassFrontier, ModelClassLowContext, req.ModelClass)
	}
	if err := personpolicy.Validate(req.PersonPolicy, req.PersonSource); err != nil {
		return nil, fmt.Errorf("request %s: %w", path, err)
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
					ID:      loc.ID,
					Reason:  fmt.Sprintf("optional source %s %s is absent", kind, sourcePath),
					Warning: true,
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

// LoadSource reads one explicit declaration or infers an AOS knowledge
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

// inferProvider applies the AOS knowledge-provider filesystem convention.
// ReadDir returns names in lexical order, keeping source decisions stable.
func inferProvider(id, root string) (*Source, error) {
	info, err := os.Stat(root)
	if err != nil {
		return nil, err
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("provider root %s is not a directory", root)
	}

	skillsRoot := filepath.Join(root, filepath.FromSlash(providerSkillsPath))
	entries, err := os.ReadDir(skillsRoot)
	if err != nil {
		return nil, fmt.Errorf("provider skills %s: %w", providerSkillsPath, err)
	}
	src := &Source{
		ID:   id,
		Root: root,
	}
	invariant := filepath.Join(root, filepath.FromSlash(providerInvariantPath))
	if info, err := os.Stat(invariant); err == nil {
		if !info.Mode().IsRegular() {
			return nil, fmt.Errorf("provider invariant %s is not a regular file", providerInvariantPath)
		}
		src.Instructions = append(src.Instructions, ContentRef{
			ID:   providerInvariantID,
			Path: providerInvariantPath,
		})
	} else if !os.IsNotExist(err) {
		return nil, fmt.Errorf("provider invariant %s: %w", providerInvariantPath, err)
	}
	ordinary := map[string]bool{}
	for _, entry := range entries {
		name := entry.Name()
		if !entry.IsDir() || name == "personality-shared" {
			continue
		}
		skillPath := filepath.Join(skillsRoot, name, "SKILL.md")
		if info, err := os.Stat(skillPath); err != nil {
			return nil, fmt.Errorf("provider skill %s: %w", name, err)
		} else if !info.Mode().IsRegular() {
			return nil, fmt.Errorf("provider skill %s SKILL.md is not a regular file", name)
		}
		src.Skills = append(src.Skills, ContentRef{
			ID:         name,
			Path:       filepath.ToSlash(filepath.Join(providerSkillsPath, name)),
			EntryPoint: "SKILL.md",
		})
		ordinary[name] = true
	}
	if len(src.Skills) == 0 {
		return nil, fmt.Errorf("provider root %s has no skills under %s", root, providerSkillsPath)
	}

	composedRoot := filepath.Join(root, filepath.FromSlash(providerComposedPath))
	if _, err := os.Stat(composedRoot); os.IsNotExist(err) {
		rolesPath := filepath.Join(root, filepath.FromSlash(providerRolesPath))
		if _, rolesErr := os.Stat(rolesPath); rolesErr == nil {
			return nil, fmt.Errorf(
				"provider role bindings %s exist without composed skills %s",
				providerRolesPath,
				providerComposedPath,
			)
		} else if !os.IsNotExist(rolesErr) {
			return nil, fmt.Errorf("provider role bindings %s: %w", providerRolesPath, rolesErr)
		}
		return src, nil
	} else if err != nil {
		return nil, fmt.Errorf("provider composed skills %s: %w", providerComposedPath, err)
	}
	composed, err := inspectComposedSkills(composedRoot, ordinary)
	if err != nil {
		return nil, err
	}
	roleSkills, err := parseRoleBindings(
		filepath.Join(root, filepath.FromSlash(providerRolesPath)),
		composed,
	)
	if err != nil {
		return nil, err
	}
	src.RoleSkills = roleSkills
	return src, nil
}

func inspectComposedSkills(root string, ordinary map[string]bool) (map[string]string, error) {
	entries, err := os.ReadDir(root)
	if err != nil {
		return nil, fmt.Errorf("provider composed skills %s: %w", providerComposedPath, err)
	}
	composed := map[string]string{}
	for _, entry := range entries {
		name := entry.Name()
		if strings.HasPrefix(name, ".") {
			continue
		}
		if !entry.IsDir() {
			return nil, fmt.Errorf("provider composed entry %s must be a directory", name)
		}
		if ordinary[name] {
			return nil, fmt.Errorf("provider skill %q exists in both %s and %s",
				name, providerSkillsPath, providerComposedPath)
		}
		skillRoot := filepath.Join(root, name)
		if err := filepath.WalkDir(skillRoot, func(path string, item os.DirEntry, walkErr error) error {
			if walkErr != nil {
				return walkErr
			}
			if item.Type()&os.ModeSymlink != 0 {
				return fmt.Errorf("%s: symlinks are invalid inside a composed skill", path)
			}
			if !item.IsDir() && item.Name() == "SKILL.md" {
				return fmt.Errorf("provider composed skill %q contains SKILL.md; source entry points must be COMPOSED.md", name)
			}
			return nil
		}); err != nil {
			return nil, err
		}
		entryPoint := filepath.Join(skillRoot, "COMPOSED.md")
		if info, err := os.Stat(entryPoint); err != nil {
			return nil, fmt.Errorf("provider composed skill %q: %w", name, err)
		} else if !info.Mode().IsRegular() {
			return nil, fmt.Errorf("provider composed skill %q COMPOSED.md is not a regular file", name)
		}
		composed[name] = filepath.ToSlash(filepath.Join(providerComposedPath, name))
	}
	return composed, nil
}

func parseRoleBindings(path string, composed map[string]string) (map[string][]ContentRef, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("provider role bindings %s: %w", providerRolesPath, err)
	}
	doc, err := kdl.ParseString(string(raw))
	if err != nil {
		return nil, fmt.Errorf("parse provider role bindings %s: %w", providerRolesPath, err)
	}
	if len(doc.Nodes) != 1 || doc.Nodes[0].Name() != "roles" {
		return nil, fmt.Errorf("provider role bindings %s: expected exactly one top-level roles node", providerRolesPath)
	}

	refs := map[string][]ContentRef{}
	seenRoles := map[string]bool{}
	for _, roleNode := range doc.Nodes[0].Children().Nodes {
		if roleNode.Name() != "role" {
			return nil, fmt.Errorf("provider role bindings %s: unknown node %q", providerRolesPath, roleNode.Name())
		}
		role, err := oneStringArg(roleNode)
		if err != nil {
			return nil, fmt.Errorf("provider role bindings %s: %w", providerRolesPath, err)
		}
		if seenRoles[role] {
			return nil, fmt.Errorf("provider role bindings %s: duplicate role %q", providerRolesPath, role)
		}
		seenRoles[role] = true
		seenSkills := map[string]bool{}
		for _, child := range roleNode.Children().Nodes {
			switch child.Name() {
			case "composed-skill":
				pattern, err := oneStringArg(child)
				if err != nil {
					return nil, fmt.Errorf("provider role %q: %w", role, err)
				}
				skills, err := expandComposedSkillPattern(pattern, composed)
				if err != nil {
					return nil, fmt.Errorf("provider role %q: %w", role, err)
				}
				for _, skill := range skills {
					if seenSkills[skill] {
						return nil, fmt.Errorf("provider role %q repeats composed skill %q", role, skill)
					}
					seenSkills[skill] = true
					refs[role] = append(refs[role], ContentRef{
						ID:         skill,
						Path:       composed[skill],
						EntryPoint: "COMPOSED.md",
					})
				}
			default:
				return nil, fmt.Errorf("provider role %q: unknown node %q", role, child.Name())
			}
		}
	}
	return refs, nil
}

func expandComposedSkillPattern(pattern string, composed map[string]string) ([]string, error) {
	if _, err := pathpkg.Match(pattern, ""); err != nil {
		return nil, fmt.Errorf("composed skill pattern %q is invalid: %w", pattern, err)
	}
	names := make([]string, 0, len(composed))
	for name := range composed {
		names = append(names, name)
	}
	sort.Strings(names)

	matches := make([]string, 0, len(names))
	for _, name := range names {
		matched, _ := pathpkg.Match(pattern, name)
		if matched {
			matches = append(matches, name)
		}
	}
	if len(matches) == 0 {
		if strings.ContainsAny(pattern, "*?[") {
			return nil, fmt.Errorf("composed skill pattern %q matches no skills", pattern)
		}
		return nil, fmt.Errorf("names missing composed skill %q", pattern)
	}
	return matches, nil
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
				ref.EntryPoint = "SKILL.md"
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
