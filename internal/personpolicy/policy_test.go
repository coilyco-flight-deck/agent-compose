package personpolicy

import (
	"strings"
	"testing"
)

func TestValidate(t *testing.T) {
	t.Parallel()
	if err := Validate("", ""); err != nil {
		t.Fatalf("default policy failed: %v", err)
	}
	if err := Validate(ExternalOnly, "/person"); err != nil {
		t.Fatalf("external-only policy failed: %v", err)
	}
	if err := Validate(ExternalOnly, ""); err == nil ||
		!strings.Contains(err.Error(), "requires person_source") {
		t.Fatalf("missing external source error = %v", err)
	}
	if err := Validate("prefer-external", "/person"); err == nil ||
		!strings.Contains(err.Error(), ExternalOnly) {
		t.Fatalf("unknown policy error = %v", err)
	}
}
