package person

import (
	"fmt"
	"io/fs"
	"path"
)

type PersonalityCatalogEntry struct {
	Slug          string            `json:"slug"`
	Skill         string            `json:"skill"`
	Description   string            `json:"description"`
	Aliases       []string          `json:"aliases"`
	Color         string            `json:"color"`
	Motif         string            `json:"motif"`
	Emblem        Emblem            `json:"emblem"`
	Form          Form              `json:"form"`
	SoundMark     SoundMark         `json:"sound_mark"`
	SourceLibrary string            `json:"source_library"`
	Digest        string            `json:"digest"`
	Affinities    []PersonalityMeld `json:"affinities"`
}

type RoleCatalogEntry struct {
	Slug          string   `json:"slug"`
	DisplayName   string   `json:"display_name"`
	Purpose       string   `json:"purpose"`
	Skill         string   `json:"skill"`
	SkillSource   string   `json:"skill_source"`
	SkillDigest   string   `json:"skill_digest"`
	Seats         []Seat   `json:"seats"`
	Personalities []string `json:"personalities"`
	FavoriteColor string   `json:"favorite_color"`
}

type SeatCatalogEntry struct {
	Role string `json:"role"`
	Seat Seat   `json:"seat"`
}

func (p *Person) PersonalityCatalog(names []string) ([]PersonalityCatalogEntry, error) {
	snapshot, err := BuildSnapshotV4(p)
	if err != nil {
		return nil, err
	}
	if err := ValidateSnapshotV4(snapshot); err != nil {
		return nil, err
	}
	source, err := Source(p)
	if err != nil {
		return nil, err
	}
	if names == nil {
		names = p.personalityOrder()
	}
	out := make([]PersonalityCatalogEntry, 0, len(names))
	for _, name := range names {
		binding, ok := p.Personalities[name]
		if !ok {
			return nil, fmt.Errorf("catalogue personality %q is not defined", name)
		}
		raw, err := fs.ReadFile(source.FileSystem(), path.Join(
			"skills", binding.Skill, "SKILL.md",
		))
		if err != nil {
			return nil, fmt.Errorf("read personality %q definition: %w", name, err)
		}
		description, err := skillDescription(raw)
		if err != nil {
			return nil, fmt.Errorf("personality %q: %w", name, err)
		}
		projected := snapshot.Personalities[name]
		out = append(out, PersonalityCatalogEntry{
			Slug:          name,
			Skill:         binding.Skill,
			Description:   description,
			Aliases:       append([]string(nil), binding.Aliases...),
			Color:         binding.Color,
			Motif:         binding.Motif,
			Emblem:        binding.Emblem,
			Form:          binding.Form,
			SoundMark:     binding.SoundMark,
			SourceLibrary: projected.SourceLibrary,
			Digest:        projected.Digest,
			Affinities:    append([]PersonalityMeld(nil), projected.Affinities...),
		})
	}
	return out, nil
}

func (p *Person) RoleCatalog() ([]RoleCatalogEntry, error) {
	snapshot, err := BuildSnapshot(p)
	if err != nil {
		return nil, err
	}
	out := make([]RoleCatalogEntry, 0, len(p.RoleOrder))
	for _, name := range p.RoleOrder {
		role := p.Roles[name]
		out = append(out, RoleCatalogEntry{
			Slug:          name,
			DisplayName:   p.RoleDisplayName(name),
			Purpose:       role.Purpose,
			Skill:         p.RoleSkillID(name),
			SkillSource:   role.SkillSource,
			SkillDigest:   role.SkillDigest,
			Seats:         append([]Seat(nil), role.Seats...),
			Personalities: append([]string(nil), role.Personalities...),
			FavoriteColor: snapshot.Roles[name].FavoriteColor,
		})
	}
	return out, nil
}

func (p *Person) SeatCatalog(roleFilter string) ([]SeatCatalogEntry, error) {
	roleNames := p.RoleOrder
	if roleFilter != "" {
		if _, ok := p.Roles[roleFilter]; !ok {
			return nil, fmt.Errorf("role %q is not defined", roleFilter)
		}
		roleNames = []string{roleFilter}
	}
	var out []SeatCatalogEntry
	for _, roleName := range roleNames {
		for _, seat := range p.Roles[roleName].Seats {
			out = append(out, SeatCatalogEntry{Role: roleName, Seat: seat})
		}
	}
	return out, nil
}
