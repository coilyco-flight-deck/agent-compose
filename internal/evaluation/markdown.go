package evaluation

import (
	"fmt"
	"strings"
)

func Markdown(pack *Pack) []byte {
	var out strings.Builder
	fmt.Fprintf(&out, "# Agent-compose behavior evaluation\n\n")
	fmt.Fprintf(&out, "* Format: `%s`\n", pack.Format)
	fmt.Fprintf(&out, "* Person: `%s`\n", pack.Person)
	fmt.Fprintf(&out, "* Role: `%s` - %s\n", pack.Role, pack.Purpose)
	fmt.Fprintf(&out, "* Seat: `%s` - `%s` (pronouns: `%s`)\n", pack.Seat.Selector(), pack.Seat.Name, pack.Seat.Pronouns)
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
	fmt.Fprintf(&out, "\nPass each case at %d/8 or higher. %s\n",
		pack.ReviewRule.PassingTotal, pack.ReviewRule.HardFailRule)
	fmt.Fprintf(&out, "Role minimums: mission-fit %d/2, authority-and-escalation %d/2. ",
		pack.ReviewRule.RoleMinimumScores["mission-fit"],
		pack.ReviewRule.RoleMinimumScores["authority-and-escalation"])
	fmt.Fprintf(&out, "Personality minimums: behavioral-expression %d/2, invariant-and-role %d/2.\n\n",
		pack.ReviewRule.PersonalityMinimumScores["behavioral-expression"],
		pack.ReviewRule.PersonalityMinimumScores["invariant-and-role"])

	fmt.Fprintf(&out, "## Scenario matrix (%d cases)\n\n", len(pack.Cases))
	for _, evalCase := range pack.Cases {
		fmt.Fprintf(&out, "### %s\n\n", evalCase.ID)
		fmt.Fprintf(&out, "* Model tier: `%s`\n", evalCase.ModelTier)
		fmt.Fprintf(&out, "* Bundle model class: `%s`\n", evalCase.BundleModelClass)
		fmt.Fprintf(&out, "* Dimension: `%s`\n", evalCase.Dimension)
		if evalCase.Scenario != "" {
			fmt.Fprintf(&out, "* Scenario: `%s` (`%s`)\n", evalCase.Scenario, evalCase.ScenarioKind)
		}
		if evalCase.AdjacentRole != "" {
			fmt.Fprintf(&out, "* Adjacent role: `%s`\n", evalCase.AdjacentRole)
		}
		out.WriteString("\n")
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
