package person

import (
	"fmt"
	"strings"
)

// RenderRoleMetadata returns the compact public-safe person facts that belong
// in every bundle. Long-form catalogue prose and citation records stay in the
// documentation surface rather than entering each agent's runtime context.
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
	fmt.Fprintf(&out, "* Role: `%s`\n", roleName)
	fmt.Fprintf(&out, "* Purpose: %s\n", role.Purpose)
	out.WriteString("* Component personalities:\n")
	for _, name := range role.Personalities {
		binding, exists := p.Personalities[name]
		if !exists {
			return "", fmt.Errorf("render role metadata: personality %q is not defined", name)
		}
		fmt.Fprintf(&out, "  * `%s`: skill `%s`, favorite color `%s`\n",
			name, binding.Skill, binding.Color)
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

	out.WriteString("* Inspiration credits:\n")
	if err := writeCredit(&out, "Role `"+roleName+"`", role.Inspiration, p.Inspirations); err != nil {
		return "", err
	}
	for _, name := range role.Personalities {
		binding := p.Personalities[name]
		if err := writeCredit(&out, "Personality `"+name+"`", binding.Inspiration, p.Inspirations); err != nil {
			return "", err
		}
	}
	return out.String(), nil
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
	fmt.Fprintf(out,
		"  * %s: `%s` (`%s`), impact mode `%s`, appearance `%s` (`%s`) at %s (%s, %s)\n",
		subject,
		inspiration.Name,
		ref.ID,
		inspiration.ImpactMode,
		inspiration.Appearance.Title,
		inspiration.Appearance.ID,
		inspiration.Appearance.Event,
		inspiration.Appearance.Year,
		inspiration.Appearance.Format,
	)
	return nil
}
