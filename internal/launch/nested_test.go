package launch

import (
	"strconv"
	"strings"
	"testing"
)

func TestNestedDepthAllowsATopLevelLaunch(t *testing.T) {
	t.Parallel()
	for name, nested := range map[string]bool{"plain": false, "opted in": true} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			depth, err := NestedDepth("", "", nested)
			if err != nil {
				t.Fatalf("top-level launch refused: %v", err)
			}
			if depth != 0 {
				t.Fatalf("child depth = %d, want 0", depth)
			}
		})
	}
}

func TestNestedDepthRefusesAnAccidentalNestedLaunch(t *testing.T) {
	t.Parallel()
	_, err := NestedDepth("1", "0", false)
	if err == nil {
		t.Fatal("a launch inside another launch was permitted without the opt-in")
	}
	if !strings.Contains(err.Error(), "--nested") {
		t.Fatalf("refusal does not name the opt-in: %v", err)
	}
}

func TestNestedDepthPermitsOneDeliberateHop(t *testing.T) {
	t.Parallel()
	depth, err := NestedDepth("1", "0", true)
	if err != nil {
		t.Fatalf("deliberate nested launch refused: %v", err)
	}
	if depth != 1 {
		t.Fatalf("child depth = %d, want 1", depth)
	}
}

func TestNestedDepthStopsAtTheBound(t *testing.T) {
	t.Parallel()
	if _, err := NestedDepth("1", strconv.Itoa(MaxNestedDepth), true); err == nil {
		t.Fatalf("a launch at depth %d was permitted to launch again", MaxNestedDepth)
	}
}

func TestNestedDepthFailsClosedOnAMalformedCount(t *testing.T) {
	t.Parallel()
	for _, value := range []string{"deep", "-1"} {
		t.Run(value, func(t *testing.T) {
			t.Parallel()
			if _, err := NestedDepth("1", value, true); err == nil {
				t.Fatalf("depth %q was accepted", value)
			}
		})
	}
}

func TestDepthEnvStampsTheChildDepth(t *testing.T) {
	t.Parallel()
	if got, want := DepthEnv(1), EnvDepth+"=1"; got != want {
		t.Fatalf("DepthEnv(1) = %q, want %q", got, want)
	}
}
