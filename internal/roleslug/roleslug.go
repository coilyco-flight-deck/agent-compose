// Package roleslug maps a retired role slug to the slug that replaced it, so a
// launch or a provider roles.kdl still naming the old slug keeps resolving while
// the fleet migrates. Temporary by design: teable:coilyco-flight-deck/agent-compose#8086.
// The table is retired.json, which evalkit reads too, so the two cannot drift.
package roleslug

import (
	_ "embed"
	"encoding/json"
	"strings"
)

//go:embed retired.json
var retiredJSON []byte

var retired = func() map[string]string {
	table := map[string]string{}
	if err := json.Unmarshal(retiredJSON, &table); err != nil {
		panic("roleslug: retired.json: " + err.Error())
	}
	return table
}()

// Canonical trims role and returns its current slug, or role unchanged when it
// is not a retired one.
func Canonical(role string) string {
	role = strings.TrimSpace(role)
	if current, ok := retired[role]; ok {
		return current
	}
	return role
}
