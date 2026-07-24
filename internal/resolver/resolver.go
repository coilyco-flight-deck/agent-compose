package resolver

import (
	"crypto/sha256"
	"fmt"
	"io/fs"
	"path"
	"sort"
	"strings"

	"forgejo.coilysiren.me/coilyco-flight-deck/agent-compose/internal/color"
	"forgejo.coilysiren.me/coilyco-flight-deck/agent-compose/internal/person"
	"forgejo.coilysiren.me/coilyco-flight-deck/agent-compose/internal/schema"
)

const (
	OutcomeSelected  = "selected"
	OutcomeExcluded  = "excluded"
	OutcomeShadowed  = "shadowed"
	OutcomeDelivered = "delivered"
)

type Decision struct {
	Subject string `json:"subject"`
	Kind    string `json:"kind"`
	Source  string `json:"source,omitempty"`
	Outcome string `json:"outcome"`
	Reason  string `json:"reason"`
}

type Selected struct {
	ID         string
	Source     string
	Files      fs.FS
	Path       string
	EntryPoint string
}

// Resolution is the full composition plan: what was selected, how it is
// delivered, and the trace built while those choices were made.
type Resolution struct {
	Request        *schema.Request
	Person         *person.Person
	Personalities  []string
	RolePurpose    string
	RoleBriefing   string
	Instructions   []Selected
	Skills         []Selected
	CompiledBodies []Selected
	FavoriteColor  string
	Decisions      []Decision
	SourceIDs      []string
}

func Resolve(req *schema.Request, p *person.Person, sources []*schema.Source, missing []schema.MissingSource) (*Resolution, error) {
	personSource, err := person.Source(p)
	if err != nil {
		return nil, err
	}
	sources = append([]*schema.Source{personSource}, sources...)

	role, ok := p.Roles[req.Role]
	if !ok {
		return nil, fmt.Errorf("role %q is not defined by person %q; defined roles: %s",
			req.Role, p.Name, strings.Join(sortedKeys(p.Roles), ", "))
	}
	if len(role.Personalities) == 0 {
		return nil, fmt.Errorf("role %q defines no personalities", req.Role)
	}
	for _, src := range sources {
		for _, roleName := range sortedKeys(src.RoleSkills) {
			if _, ok := p.Roles[roleName]; !ok {
				return nil, fmt.Errorf("source %q binds composed skills to undefined role %q", src.ID, roleName)
			}
		}
	}

	activeBySkill := make(map[string]string, len(role.Personalities))
	personalityBySkill := make(map[string]string, len(p.Personalities))
	activeSkills := make([]string, 0, len(role.Personalities))
	colors := make([]string, 0, len(role.Personalities))
	for name, binding := range p.Personalities {
		personalityBySkill[binding.Skill] = name
	}
	for _, name := range role.Personalities {
		binding, ok := p.Personalities[name]
		if !ok {
			return nil, fmt.Errorf("role %q names personality %q without a catalog binding", req.Role, name)
		}
		if prior, duplicate := activeBySkill[binding.Skill]; duplicate {
			return nil, fmt.Errorf("role %q personalities %q and %q bind the same skill %q",
				req.Role, prior, name, binding.Skill)
		}
		activeBySkill[binding.Skill] = name
		activeSkills = append(activeSkills, binding.Skill)
		colors = append(colors, binding.Color)
	}
	favorite, err := color.Favorite(colors)
	if err != nil {
		return nil, fmt.Errorf("derive favorite color for role %q: %w", req.Role, err)
	}

	res := &Resolution{
		Request:       req,
		Person:        p,
		Personalities: append([]string(nil), role.Personalities...),
		RolePurpose:   role.Purpose,
		RoleBriefing:  role.Briefing,
		FavoriteColor: favorite,
	}
	for _, src := range sources {
		res.SourceIDs = append(res.SourceIDs, src.ID)
	}
	res.decide(Decision{
		Subject: "role:" + req.Role, Kind: "profile", Source: "person:" + p.Name,
		Outcome: OutcomeSelected,
		Reason:  fmt.Sprintf("person %q defines this role: %s", p.Name, role.Purpose),
	})
	res.decide(Decision{
		Subject: "instruction:role-briefing", Kind: "instruction", Source: "person:" + p.Name,
		Outcome: OutcomeSelected,
		Reason:  fmt.Sprintf("role %q activates its canonical operating charter", req.Role),
	})
	for _, name := range role.Personalities {
		res.decide(Decision{
			Subject: "personality:" + name, Kind: "profile", Source: "person:" + p.Name,
			Outcome: OutcomeSelected,
			Reason:  fmt.Sprintf("role %q activates its full personality set: %s", req.Role, strings.Join(role.Personalities, ", ")),
		})
	}
	for _, m := range missing {
		res.decide(Decision{
			Subject: "source:" + m.ID, Kind: "source", Source: m.ID,
			Outcome: OutcomeExcluded, Reason: m.Reason,
		})
	}

	instructionBytes := map[string][]byte{}
	instructionOwner := map[string]string{}
	selectedBySkill := map[string]Selected{}
	skillDigests := map[string]string{}
	considerSkill := func(src *schema.Source, ref schema.ContentRef, reason string) error {
		digest, err := treeDigest(src.FileSystem(), ref.Path)
		if err != nil {
			return fmt.Errorf("source %q skill %q: %w", src.ID, ref.ID, err)
		}
		if selected, found := selectedBySkill[ref.ID]; found {
			if digest == skillDigests[ref.ID] && entryPoint(ref) == selected.EntryPoint {
				res.decide(Decision{
					Subject: "skill:" + ref.ID, Kind: "skill", Source: src.ID,
					Outcome: OutcomeShadowed,
					Reason:  fmt.Sprintf("identical to the copy already selected from %s", selected.Source),
				})
				return nil
			}
			return fmt.Errorf("skill %q from %s conflicts with a different copy from %s; v0.1 fails non-identical collisions",
				ref.ID, src.ID, selected.Source)
		}
		selectedBySkill[ref.ID] = Selected{
			ID: ref.ID, Source: src.ID, Files: src.FileSystem(), Path: ref.Path, EntryPoint: entryPoint(ref),
		}
		skillDigests[ref.ID] = digest
		res.decide(Decision{
			Subject: "skill:" + ref.ID, Kind: "skill", Source: src.ID,
			Outcome: OutcomeSelected, Reason: reason,
		})
		return nil
	}
	for _, src := range sources {
		for _, ref := range src.Instructions {
			raw, err := src.ReadFile(ref.Path)
			if err != nil {
				return nil, fmt.Errorf("source %q instruction %q: %w", src.ID, ref.ID, err)
			}
			if prior, dup := instructionBytes[ref.ID]; dup {
				if string(prior) == string(raw) {
					res.decide(Decision{
						Subject: "instruction:" + ref.ID, Kind: "instruction", Source: src.ID,
						Outcome: OutcomeShadowed,
						Reason:  fmt.Sprintf("identical to the copy already selected from %s", instructionOwner[ref.ID]),
					})
					continue
				}
				return nil, fmt.Errorf("instruction %q from %s conflicts with a different copy from %s; v0.1 fails non-identical collisions",
					ref.ID, src.ID, instructionOwner[ref.ID])
			}
			instructionBytes[ref.ID] = raw
			instructionOwner[ref.ID] = src.ID
			res.Instructions = append(res.Instructions, Selected{
				ID: ref.ID, Source: src.ID, Files: src.FileSystem(), Path: ref.Path,
			})
			res.decide(Decision{
				Subject: "instruction:" + ref.ID, Kind: "instruction", Source: src.ID,
				Outcome: OutcomeSelected, Reason: "instructions from admitted sources are always selected",
			})
		}
		for _, ref := range src.Skills {
			_, isPersonality := personalityBySkill[ref.ID]
			activePersonality, active := activeBySkill[ref.ID]
			if isPersonality && !active {
				res.decide(Decision{
					Subject: "skill:" + ref.ID, Kind: "skill", Source: src.ID,
					Outcome: OutcomeExcluded,
					Reason: fmt.Sprintf("role %q activates personalities %s, which bind skills %s",
						req.Role, strings.Join(role.Personalities, ", "), strings.Join(activeSkills, ", ")),
				})
				continue
			}
			reason := "ordinary provider skills are discoverable for every role"
			if isPersonality {
				reason = fmt.Sprintf("active personality %q binds this skill", activePersonality)
			}
			if err := considerSkill(src, ref, reason); err != nil {
				return nil, err
			}
		}
		for _, ref := range src.RoleSkills[req.Role] {
			if err := considerSkill(
				src,
				ref,
				fmt.Sprintf("role %q composes this skill", req.Role),
			); err != nil {
				return nil, err
			}
		}
	}

	added := map[string]bool{}
	for _, name := range role.Personalities {
		boundSkill := p.Personalities[name].Skill
		selected, found := selectedBySkill[boundSkill]
		if !found {
			return nil, fmt.Errorf("personality %q binds skill %q, but no admitted source provides it", name, boundSkill)
		}
		res.Skills = append(res.Skills, selected)
		added[boundSkill] = true
	}
	for _, src := range sources {
		refs := append([]schema.ContentRef{}, src.Skills...)
		refs = append(refs, src.RoleSkills[req.Role]...)
		for _, ref := range refs {
			if added[ref.ID] {
				continue
			}
			selected, found := selectedBySkill[ref.ID]
			if !found {
				continue
			}
			res.Skills = append(res.Skills, selected)
			added[ref.ID] = true
		}
	}

	if err := res.planDelivery(); err != nil {
		return nil, err
	}
	return res, nil
}

// planDelivery chooses the compiled body for each skill.
func (r *Resolution) planDelivery() error {
	r.decide(Decision{
		Subject: "content/instructions.md", Kind: "delivery",
		Outcome: OutcomeDelivered, Reason: "canonical selected instructions",
	})
	r.decide(Decision{
		Subject: "content/skills", Kind: "delivery",
		Outcome: OutcomeDelivered, Reason: "canonical selected skill trees",
	})
	if r.Request.Delivery != schema.DeliveryCompiled {
		return nil
	}
	for _, skill := range r.Skills {
		body := path.Join(skill.Path, skill.EntryPoint)
		if _, err := fs.Stat(skill.Files, body); err != nil {
			return fmt.Errorf("skill %q: compiled delivery needs %s: %w", skill.ID, path.Base(body), err)
		}
		r.CompiledBodies = append(r.CompiledBodies, skill)
	}
	r.decide(Decision{
		Subject: "delivery/compiled.md", Kind: "delivery",
		Outcome: OutcomeDelivered, Reason: "selected instructions and skill prose compiled into one context document",
	})
	return nil
}

func entryPoint(ref schema.ContentRef) string {
	if ref.EntryPoint != "" {
		return ref.EntryPoint
	}
	return "SKILL.md"
}

func (r *Resolution) decide(d Decision) {
	r.Decisions = append(r.Decisions, d)
}

func treeDigest(files fs.FS, root string) (string, error) {
	h := sha256.New()
	err := fs.WalkDir(files, root, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.Type()&fs.ModeSymlink != 0 {
			return fmt.Errorf("%s: symlinks are invalid inside a source", p)
		}
		if d.IsDir() {
			return nil
		}
		rel := strings.TrimPrefix(p, strings.TrimSuffix(root, "/")+"/")
		if rel == p {
			return fmt.Errorf("%s is not beneath %s", p, root)
		}
		raw, err := fs.ReadFile(files, p)
		if err != nil {
			return err
		}
		fmt.Fprintf(h, "%s\x00%d\x00", rel, len(raw))
		h.Write(raw)
		return nil
	})
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", h.Sum(nil)), nil
}

func sortedKeys[V any](m map[string]V) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
