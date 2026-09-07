package bundle_test

import (
	"os"
	"testing"

	"github.com/coilyco-flight-deck/agent-compose/v2/internal/rostertest"
)

func TestMain(m *testing.M) {
	rostertest.Use()
	os.Exit(m.Run())
}
