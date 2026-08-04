// Package skillselector validates and applies bounded ordinary-skill selectors.
package skillselector

import (
	"fmt"
	"path"
	"strings"
)

// Validate checks selector syntax without a provider catalogue. Nil preserves
// the whole provider, while an explicit empty selector is invalid.
func Validate(patterns []string) error {
	if patterns == nil {
		return nil
	}
	if len(patterns) == 0 {
		return fmt.Errorf("skills selector is empty; omit skills to admit the whole provider")
	}
	for index, pattern := range patterns {
		if strings.TrimSpace(pattern) == "" {
			return fmt.Errorf("skills selector pattern %d is empty", index)
		}
		if _, err := path.Match(pattern, "candidate"); err != nil {
			return fmt.Errorf("skills selector pattern %q is invalid: %w", pattern, err)
		}
	}
	return nil
}

// Select returns selected and excluded catalogue IDs in their original order.
// Every configured pattern must match, and no skill may match two patterns.
func Select(patterns, catalogue []string) ([]string, []string, error) {
	if err := Validate(patterns); err != nil {
		return nil, nil, err
	}
	if patterns == nil {
		return append([]string(nil), catalogue...), nil, nil
	}

	matchedPatterns := make([]bool, len(patterns))
	selected := make([]string, 0, len(catalogue))
	excluded := make([]string, 0, len(catalogue))
	for _, skill := range catalogue {
		var matched []int
		for index, pattern := range patterns {
			ok, err := path.Match(pattern, skill)
			if err != nil {
				return nil, nil, fmt.Errorf("skills selector pattern %q is invalid: %w", pattern, err)
			}
			if ok {
				matched = append(matched, index)
			}
		}
		switch len(matched) {
		case 0:
			excluded = append(excluded, skill)
		case 1:
			matchedPatterns[matched[0]] = true
			selected = append(selected, skill)
		default:
			patternsForSkill := make([]string, 0, len(matched))
			for _, index := range matched {
				patternsForSkill = append(patternsForSkill, patterns[index])
			}
			return nil, nil, fmt.Errorf(
				"skills selector patterns overlap on skill %q: %s",
				skill,
				strings.Join(patternsForSkill, ", "),
			)
		}
	}
	for index, matched := range matchedPatterns {
		if !matched {
			return nil, nil, fmt.Errorf(
				"skills selector pattern %q matches no ordinary provider skill",
				patterns[index],
			)
		}
	}
	return selected, excluded, nil
}
