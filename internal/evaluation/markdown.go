package evaluation

import (
	"fmt"
	"strings"
)

func Markdown(pack *Pack) []byte {
	var out strings.Builder
	fmt.Fprintf(&out, "# Agent-compose behavior evaluation\n\n")
	fmt.Fprintf(&out, "* Format: `%s`\n", pack.Format)
	fmt.Fprintf(&out, "* Role: `%s` - %s\n", pack.Role, pack.Purpose)
	fmt.Fprintf(&out, "* Seat: `%s` - `%s` (pronouns: `%s`)\n", pack.Seat.Harness, pack.Seat.Name, pack.Seat.Pronouns)
	fmt.Fprintf(&out, "* Personalities: `%s`\n", personalityNames(pack.Personalities))
	fmt.Fprintf(&out, "* Melded favorite color: `%s`\n\n", pack.MeldedFavoriteColor)

	out.WriteString("## Role briefing\n\n")
	out.WriteString(pack.Briefing)
	out.WriteString("\n\n## Personality invariant\n\n")
	out.WriteString(pack.Invariant)
	out.WriteString("\n\n## Personality definitions\n\n")
	for _, personality := range pack.Personalities {
		fmt.Fprintf(&out, "### %s\n\n%s\n\n", personality.Name, personality.Definition)
	}

	out.WriteString("## Run protocol\n\n")
	for i, step := range pack.RunProtocol {
		fmt.Fprintf(&out, "%d. %s\n", i+1, step)
	}
	fmt.Fprintf(&out, "\nPass each case at %d/8 or higher. %s\n\n",
		pack.ReviewRule.PassingTotal, pack.ReviewRule.HardFailRule)

	out.WriteString("## Four-case matrix\n\n")
	for _, evalCase := range pack.Cases {
		fmt.Fprintf(&out, "### %s\n\n", evalCase.ID)
		fmt.Fprintf(&out, "* Model tier: `%s`\n", evalCase.ModelTier)
		fmt.Fprintf(&out, "* Bundle model class: `%s`\n", evalCase.BundleModelClass)
		fmt.Fprintf(&out, "* Dimension: `%s`\n\n", evalCase.Dimension)
		fmt.Fprintf(&out, "**Prompt**\n\n%s\n\n", evalCase.Prompt)
		fmt.Fprintf(&out, "**Reviewer question**\n\n%s\n\n", evalCase.ReviewerQuestion)
		out.WriteString("**Rubric**\n\n")
		for _, criterion := range evalCase.Rubric {
			hardFail := ""
			if criterion.HardFail {
				hardFail = " Hard fail at 0."
			}
			fmt.Fprintf(&out, "* `%s` - %s%s\n", criterion.ID, criterion.Question, hardFail)
			fmt.Fprintf(&out, "  * 2 - %s\n", criterion.Scale.Strong)
			fmt.Fprintf(&out, "  * 1 - %s\n", criterion.Scale.Partial)
			fmt.Fprintf(&out, "  * 0 - %s\n", criterion.Scale.Missing)
		}
		out.WriteString("\n")
	}
	return []byte(out.String())
}

func personalityNames(personalities []PersonalityContext) string {
	names := make([]string, 0, len(personalities))
	for _, personality := range personalities {
		names = append(names, personality.Name)
	}
	return strings.Join(names, "`, `")
}
