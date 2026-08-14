package agentid

import "testing"

// The dropped characters, so a regression that widened the alphabet fails.
func TestValidRejectsConfusableCharacters(t *testing.T) {
	for _, id := range []string{
		"il45", // i and l are dropped letters
		"no45", // n and o are dropped letters
		"ab01", // 0 and 1 are dropped digits
		"ab23", // 2 and 3 are dropped digits
	} {
		if Valid(id) {
			t.Errorf("Valid(%q) = true, want false: that character is dropped for a reason", id)
		}
	}
}

func TestValidShape(t *testing.T) {
	for _, id := range []string{"uz86", "ab45", "zz99"} {
		if !Valid(id) {
			t.Errorf("Valid(%q) = false, want true", id)
		}
	}
	for _, id := range []string{
		"",      // empty
		"uz8",   // too short
		"uz866", // too long
		"8uz6",  // digits and letters transposed
		"45ab",  // shape reversed
		"UZ86",  // uppercase is not canonical
		"uz-6",  // punctuation
	} {
		if Valid(id) {
			t.Errorf("Valid(%q) = true, want false", id)
		}
	}
}

// A dictated id arrives however the human typed it; the stored form is lower.
func TestNormalize(t *testing.T) {
	for raw, want := range map[string]string{
		"uz86":     "uz86",
		"  uz86  ": "uz86",
		"UZ86":     "uz86",
		"Uz86":     "uz86",
		"nope":     "",
		"":         "",
		"ab01":     "",
	} {
		if got := Normalize(raw); got != want {
			t.Errorf("Normalize(%q) = %q, want %q", raw, got, want)
		}
	}
}

func TestFromEnv(t *testing.T) {
	t.Setenv(SessionEnv, "uz86")
	if got := FromEnv(); got != "uz86" {
		t.Errorf("FromEnv = %q, want uz86", got)
	}
}

// Absent is the ordinary case outside a native session, not an error.
func TestFromEnvUnsetYieldsNoID(t *testing.T) {
	t.Setenv(SessionEnv, "")
	if got := FromEnv(); got != "" {
		t.Errorf("FromEnv = %q, want empty", got)
	}
}

// Dropped rather than displayed: the value comes from the environment, so it
// cannot be assumed well-formed.
func TestFromEnvDropsAMalformedValue(t *testing.T) {
	for _, raw := range []string{"not-an-id", "ab01", "toolongvalue", "12"} {
		t.Setenv(SessionEnv, raw)
		if got := FromEnv(); got != "" {
			t.Errorf("FromEnv with %q = %q, want empty", raw, got)
		}
	}
}
