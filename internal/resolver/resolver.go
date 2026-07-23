package resolver

import (
	"crypto/sha256"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
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
	OutcomeFallback  = "fallback"
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
	ID     string
	Source string
	Path   string
}

// Resolution is the full composition plan: what was selected, how it is
// delivered, and the trace built while those choices were made.
type Resolution struct {
	Request        *schema.Request
	Person         *person.Person
	Personalities  []string
	Instructions   []Selected
	Skills         []Selected
	CompiledBodies []string
	FavoriteColor  string
	Decisions      []Decision
	SourceIDs      []string
}

func Resolve(req *schema.Request, p *person.Person, sources []*schema.Source, missing []schema.MissingSource) (*Resolution, error) {
	role, ok := p.Roles[req.Role]
	if !ok {
		return nil, fmt.Errorf("role %q is not defined by person %q; defined roles: %s",
			req.Role, p.Name, strings.Join(sortedKeys(p.Roles), ", "))
	}
	if len(role.Personalities) == 0 {
		return nil, fmt.Errorf("role %q defines no personalities", req.Role)
	}

	activeBySkill := make(map[string]string, len(role.Personalities))
	activeSkills := make([]string, 0, len(role.Personalities))
	colors := make([]string, 0, len(role.Personalities))
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
		FavoriteColor: favorite,
		SourceIDs:     []string{"person:" + p.Name},
	}
	for _, src := range sources {
		res.SourceIDs = append(res.SourceIDs, src.ID)
	}
	res.decide(Decision{
		Subject: "role:" + req.Role, Kind: "profile", Source: "person:" + p.Name,
		Outcome: OutcomeSelected,
		Reason:  fmt.Sprintf("person %q defines this role: %s", p.Name, role.Purpose),
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
	for _, src := range sources {
		for _, ref := range src.Instructions {
			abs := filepath.Join(src.Root, ref.Path)
			raw, err := os.ReadFile(abs)
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
			res.Instructions = append(res.Instructions, Selected{ID: ref.ID, Source: src.ID, Path: abs})
			res.decide(Decision{
				Subject: "instruction:" + ref.ID, Kind: "instruction", Source: src.ID,
				Outcome: OutcomeSelected, Reason: "instructions from admitted sources are always selected",
			})
		}
		for _, ref := range src.Skills {
			abs := filepath.Join(src.Root, ref.Path)
			personalityName, active := activeBySkill[ref.ID]
			if !active {
				res.decide(Decision{
					Subject: "skill:" + ref.ID, Kind: "skill", Source: src.ID,
					Outcome: OutcomeExcluded,
					Reason: fmt.Sprintf("role %q activates personalities %s, which bind skills %s",
						req.Role, strings.Join(role.Personalities, ", "), strings.Join(activeSkills, ", ")),
				})
				continue
			}
			digest, err := treeDigest(abs)
			if err != nil {
				return nil, fmt.Errorf("source %q skill %q: %w", src.ID, ref.ID, err)
			}
			if selected, found := selectedBySkill[ref.ID]; found {
				if digest == skillDigests[ref.ID] {
					res.decide(Decision{
						Subject: "skill:" + ref.ID, Kind: "skill", Source: src.ID,
						Outcome: OutcomeShadowed,
						Reason:  fmt.Sprintf("identical to the copy already selected from %s", selected.Source),
					})
					continue
				}
				return nil, fmt.Errorf("skill %q from %s conflicts with a different copy from %s; v0.1 fails non-identical collisions",
					ref.ID, src.ID, selected.Source)
			}
			selectedBySkill[ref.ID] = Selected{ID: ref.ID, Source: src.ID, Path: abs}
			skillDigests[ref.ID] = digest
			res.decide(Decision{
				Subject: "skill:" + ref.ID, Kind: "skill", Source: src.ID,
				Outcome: OutcomeSelected,
				Reason:  fmt.Sprintf("active personality %q binds this skill", personalityName),
			})
		}
	}

	for _, name := range role.Personalities {
		boundSkill := p.Personalities[name].Skill
		selected, found := selectedBySkill[boundSkill]
		if !found {
			return nil, fmt.Errorf("personality %q binds skill %q, but no admitted source provides it", name, boundSkill)
		}
		res.Skills = append(res.Skills, selected)
	}

	if err := res.planDelivery(); err != nil {
		return nil, err
	}
	return res, nil
}

// planDelivery chooses entry points and, for compiled delivery, which document
// represents each active personality at the requested density.
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
		full := filepath.Join(skill.Path, "SKILL.md")
		body := full
		if r.Request.Density == schema.DensityBrief {
			brief := filepath.Join(skill.Path, "BRIEF.md")
			if _, err := os.Stat(brief); err == nil {
				body = brief
				r.decide(Decision{
					Subject: "skill:" + skill.ID + "/BRIEF.md", Kind: "delivery", Source: skill.Source,
					Outcome: OutcomeSelected, Reason: "brief density prefers BRIEF.md when the skill provides one",
				})
			} else {
				r.decide(Decision{
					Subject: "skill:" + skill.ID + "/SKILL.md", Kind: "delivery", Source: skill.Source,
					Outcome: OutcomeFallback, Reason: "brief density requested but the skill provides no BRIEF.md",
				})
			}
		}
		if _, err := os.Stat(body); err != nil {
			return fmt.Errorf("skill %q: compiled delivery needs %s: %w", skill.ID, filepath.Base(body), err)
		}
		r.CompiledBodies = append(r.CompiledBodies, body)
	}
	r.decide(Decision{
		Subject: "delivery/compiled.md", Kind: "delivery",
		Outcome: OutcomeDelivered, Reason: "selected instructions and all active personality prose compiled into one context document",
	})
	return nil
}

func (r *Resolution) decide(d Decision) {
	r.Decisions = append(r.Decisions, d)
}

func treeDigest(root string) (string, error) {
	h := sha256.New()
	err := filepath.WalkDir(root, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.Type()&fs.ModeSymlink != 0 {
			return fmt.Errorf("%s: symlinks are invalid inside a source", p)
		}
		if d.IsDir() {
			return nil
		}
		rel, err := filepath.Rel(root, p)
		if err != nil {
			return err
		}
		raw, err := os.ReadFile(p)
		if err != nil {
			return err
		}
		fmt.Fprintf(h, "%s\x00%d\x00", filepath.ToSlash(rel), len(raw))
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
