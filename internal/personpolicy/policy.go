// Package personpolicy owns the policy vocabulary for choosing a person
// package. It deliberately does not load packages or resolve filesystem paths.
package personpolicy

import (
	"fmt"
	"strings"
)

const ExternalOnly = "external-only"

// Validate rejects unknown policies and requires a source when embedded
// fallback is prohibited.
func Validate(policy, source string) error {
	switch policy {
	case "":
		return nil
	case ExternalOnly:
		if strings.TrimSpace(source) == "" {
			return fmt.Errorf("person policy %q requires person_source", ExternalOnly)
		}
		return nil
	default:
		return fmt.Errorf("person policy must be %q, got %q", ExternalOnly, policy)
	}
}
