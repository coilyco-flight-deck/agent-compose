package person

import (
	"fmt"
	"strings"

	"forgejo.coilysiren.me/coilyco-flight-deck/agent-compose/internal/color"
)

type RoleTranscriptOptions struct {
	Color     bool
	TrueColor bool
}

// RenderRoleMetadata returns complete selected identity facts for each bundle.
// Appearance summaries remain in the person snapshot and documentation.
func (p *Person) RenderRoleMetadata(roleName, meldedColor string) (string, error) {
	role, ok := p.Roles[roleName]
	if !ok {
		return "", fmt.Errorf("render role metadata: role %q is not defined", roleName)
	}
	if meldedColor == "" {
		return "", fmt.Errorf("render role metadata: role %q has no melded color", roleName)
	}

	var out strings.Builder
	out.WriteString("## Active role metadata\n\n")
	out.WriteString("Agent-compose selected these public-safe facts for the agent. ")
	out.WriteString("Credits acknowledge influences and do not assign another identity.\n\n")
	fmt.Fprintf(&out, "* Person: `%s`\n", p.Name)
	fmt.Fprintf(&out, "* Provider: `person:%s`\n", p.Name)
	fmt.Fprintf(&out, "* Role: `%s`\n", roleName)
	fmt.Fprintf(&out, "* Purpose: %s\n", role.Purpose)
	out.WriteString("* Role inspiration:\n")
	if err := writeCredit(&out, "Role `"+roleName+"`", role.Inspiration, p.Inspirations); err != nil {
		return "", err
	}
	out.WriteString("* Component personalities:\n")
	for _, name := range role.Personalities {
		binding, exists := p.Personalities[name]
		if !exists {
			return "", fmt.Errorf("render role metadata: personality %q is not defined", name)
		}
		fmt.Fprintf(
			&out,
			"  * `%s`: skill `%s`, favorite color `%s`, emblem `%s` / `%s` / `%s`, motif `%s`, form `%s` / `%s` / `%s`, sound mark `%s` / `%s` / `%s`\n",
			name,
			binding.Skill,
			binding.Color,
			binding.Emblem.Name,
			binding.Emblem.Emoji,
			binding.Emblem.Glyph,
			binding.Motif,
			binding.Form.Silhouette,
			binding.Form.Geometry,
			binding.Form.Motion,
			binding.SoundMark.Timbre,
			binding.SoundMark.Contour,
			binding.SoundMark.Pulse,
		)
		if err := writeCredit(&out, "Personality `"+name+"`", binding.Inspiration, p.Inspirations); err != nil {
			return "", err
		}
	}
	fmt.Fprintf(&out, "* Melded favorite color: `%s`\n", meldedColor)

	if len(role.Seats) == 0 {
		out.WriteString("* Known agent seats: none\n")
	} else {
		out.WriteString("* Known agent seats:\n")
		for _, seat := range role.Seats {
			fmt.Fprintf(&out, "  * `%s`: `%s` (pronouns: `%s`)\n",
				seat.Harness, seat.Name, seat.Pronouns)
		}
	}
	fmt.Fprintf(&out, "* Renderer expressions: `%s`\n", strings.Join(ExpressionVocabulary(), "`, `"))
	return out.String(), nil
}

// RenderRoleTranscript returns the complete selected slice of the person
// snapshot as deterministic, flush-left terminal text.
func (p *Person) RenderRoleTranscript(
	roleName, meldedColor string,
	opts RoleTranscriptOptions,
) (string, error) {
	role, ok := p.Roles[roleName]
	if !ok {
		return "", fmt.Errorf("render role transcript: role %q is not defined", roleName)
	}
	if meldedColor == "" {
		return "", fmt.Errorf("render role transcript: role %q has no melded color", roleName)
	}
	roleCredit, ok := p.Inspirations[role.Inspiration.ID]
	if !ok {
		return "", fmt.Errorf("render role transcript: inspiration %q is not defined", role.Inspiration.ID)
	}

	var out strings.Builder
	var roleBlock strings.Builder
	roleBlock.WriteString("role metadata\n")
	fmt.Fprintf(&roleBlock, "person: %s // provided by: person:%s\n", p.Name, p.Name)
	fmt.Fprintf(&roleBlock, "role: %s\n", roleName)
	fmt.Fprintf(&roleBlock, "purpose: %s\n", role.Purpose)
	fmt.Fprintf(&roleBlock, "personalities: %s\n", strings.Join(role.Personalities, " // "))
	fmt.Fprintf(&roleBlock, "melded color: %s\n", meldedColor)
	writeTranscriptInspiration(&roleBlock, "role inspiration", role.Inspiration, roleCredit)
	writeTranscriptParagraphs(&roleBlock, "briefing", role.Briefing)
	roleBlock.WriteString("seats:\n")
	for _, seat := range role.Seats {
		fmt.Fprintf(&roleBlock, "seat %s: %s // pronouns: %s\n", seat.Harness, seat.Name, seat.Pronouns)
	}
	writeTranscriptSection(&out, meldedColor, roleBlock.String(), opts)

	out.WriteByte('\n')
	writeTranscriptSection(&out, meldedColor, "personality metadata\n", opts)
	for index, name := range role.Personalities {
		binding, exists := p.Personalities[name]
		if !exists {
			return "", fmt.Errorf("render role transcript: personality %q is not defined", name)
		}
		credit, exists := p.Inspirations[binding.Inspiration.ID]
		if !exists {
			return "", fmt.Errorf("render role transcript: inspiration %q is not defined", binding.Inspiration.ID)
		}
		if index > 0 {
			out.WriteByte('\n')
		}
		var personalityBlock strings.Builder
		fmt.Fprintf(&personalityBlock, "personality: %s\n", name)
		fmt.Fprintf(&personalityBlock, "skill: %s\n", binding.Skill)
		fmt.Fprintf(&personalityBlock, "color: %s\n", binding.Color)
		fmt.Fprintf(&personalityBlock, "motif: %s\n", binding.Motif)
		fmt.Fprintf(&personalityBlock, "emblem: %s // emoji: %s // glyph: %s\n",
			binding.Emblem.Name, binding.Emblem.Emoji, binding.Emblem.Glyph)
		fmt.Fprintf(&personalityBlock, "form: silhouette %s // geometry %s // motion %s\n",
			binding.Form.Silhouette, binding.Form.Geometry, binding.Form.Motion)
		fmt.Fprintf(&personalityBlock, "sound mark: timbre %s // contour %s // pulse %s\n",
			binding.SoundMark.Timbre, binding.SoundMark.Contour, binding.SoundMark.Pulse)
		writeTranscriptInspiration(&personalityBlock, "inspiration", binding.Inspiration, credit)
		writeTranscriptSection(&out, binding.Color, personalityBlock.String(), opts)
	}

	expressions := fmt.Sprintf(
		"renderer expressions: %s\n",
		strings.Join(ExpressionVocabulary(), " // "),
	)
	out.WriteByte('\n')
	writeTranscriptSection(&out, meldedColor, expressions, opts)
	return out.String(), nil
}

func writeTranscriptSection(
	out *strings.Builder,
	hex, text string,
	opts RoleTranscriptOptions,
) {
	if opts.Color {
		out.WriteString(color.ANSI(hex, text, opts.TrueColor))
		return
	}
	out.WriteString(text)
}

func writeTranscriptInspiration(
	out *strings.Builder,
	label string,
	ref InspirationRef,
	inspiration Inspiration,
) {
	fmt.Fprintf(out, "%s: %s (%s)\n", label, inspiration.Name, ref.ID)
	fmt.Fprintf(out, "%s fit: %s\n", label, ref.Fit)
	fmt.Fprintf(out, "%s achievement: %s\n", label, inspiration.Achievement)
	fmt.Fprintf(out, "%s impact mode: %s\n", label, inspiration.ImpactMode)
	fmt.Fprintf(out, "%s impact fit: %s\n", label, inspiration.ImpactFit)
	fmt.Fprintf(out, "%s profile citation: %s\n", label, inspiration.ProfileCitation)
	fmt.Fprintf(out, "%s appearance: %s (%s)\n",
		label, inspiration.Appearance.Title, inspiration.Appearance.ID)
	fmt.Fprintf(out, "%s appearance event: %s // year: %s // format: %s\n",
		label,
		inspiration.Appearance.Event, inspiration.Appearance.Year, inspiration.Appearance.Format)
	fmt.Fprintf(out, "%s appearance citations: %s\n",
		label, strings.Join(inspiration.Appearance.Citations, " // "))
}

func writeTranscriptParagraphs(out *strings.Builder, label, value string) {
	fmt.Fprintf(out, "%s:\n", label)
	wrote := false
	for _, raw := range strings.Split(strings.TrimSpace(value), "\n\n") {
		paragraph := strings.Join(strings.Fields(raw), " ")
		if paragraph == "" {
			continue
		}
		if wrote {
			out.WriteByte('\n')
		}
		out.WriteString(paragraph)
		out.WriteByte('\n')
		wrote = true
	}
	if !wrote {
		out.WriteString("(none)\n")
	}
}

func writeCredit(
	out *strings.Builder,
	subject string,
	ref InspirationRef,
	catalog map[string]Inspiration,
) error {
	inspiration, ok := catalog[ref.ID]
	if !ok {
		return fmt.Errorf("render role metadata: inspiration %q is not defined", ref.ID)
	}
	fmt.Fprintf(out, "  * %s: `%s` (`%s`)\n", subject, inspiration.Name, ref.ID)
	fmt.Fprintf(out, "    * Fit: %s\n", ref.Fit)
	fmt.Fprintf(out, "    * Achievement: %s\n", inspiration.Achievement)
	fmt.Fprintf(out, "    * Impact mode: `%s`\n", inspiration.ImpactMode)
	fmt.Fprintf(out, "    * Impact fit: %s\n", inspiration.ImpactFit)
	fmt.Fprintf(out, "    * Profile citation: `%s`\n", inspiration.ProfileCitation)
	fmt.Fprintf(out, "    * Appearance: `%s` (`%s`) at %s (%s, %s)\n",
		inspiration.Appearance.Title,
		inspiration.Appearance.ID,
		inspiration.Appearance.Event,
		inspiration.Appearance.Year,
		inspiration.Appearance.Format,
	)
	fmt.Fprintf(out, "    * Appearance citations: `%s`\n",
		strings.Join(inspiration.Appearance.Citations, "`, `"))
	return nil
}
