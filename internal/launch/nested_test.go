package launch

import (
	"strconv"
	"testing"
)

func TestNestedDepthAllowsATopLevelLaunch(t *testing.T) {
	t.Parallel()
	depth, err := NestedDepth("", "")
	if err != nil {
		t.Fatalf("top-level launch refused: %v", err)
	}
	if depth != 0 {
		t.Fatalf("child depth = %d, want 0", depth)
	}
}

// The opt-in is gone, so a launch inside another one is permitted on its own
// and the bound below is what stops a runaway. agent-compose#403
func TestNestedDepthPermitsOneHopWithoutBeingAsked(t *testing.T) {
	t.Parallel()
	depth, err := NestedDepth("1", "0")
	if err != nil {
		t.Fatalf("nested launch refused: %v", err)
	}
	if depth != 1 {
		t.Fatalf("child depth = %d, want 1", depth)
	}
}

func TestNestedDepthStopsAtTheBound(t *testing.T) {
	t.Parallel()
	_, err := NestedDepth("1", strconv.Itoa(MaxNestedDepth))
	if err == nil {
		t.Fatalf("a launch at depth %d was permitted to launch again", MaxNestedDepth)
	}
}

func TestNestedDepthFailsClosedOnAMalformedCount(t *testing.T) {
	t.Parallel()
	for _, value := range []string{"deep", "-1"} {
		t.Run(value, func(t *testing.T) {
			t.Parallel()
			if _, err := NestedDepth("1", value); err == nil {
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
