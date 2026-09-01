package person

import (
	"fmt"
)

// The manifest names the package and nothing else. Entity shapes live in
// yaml_entity.go. See docs/person-packages.md.

// RosterEnvUnused is intentionally absent; see seed.go for the mount override.

const yamlFragmentExt = ".yaml"

type yamlManifest struct {
	Person  string `yaml:"person"`
	Roster  string `yaml:"roster"`
	Library string `yaml:"library"`
}

// yamlManifestName returns the declared node and name for a YAML manifest.
func yamlManifestName(raw []byte) (string, string, error) {
	var manifest yamlManifest
	if err := decodeEntity(raw, &manifest); err != nil {
		return "", "", err
	}
	declared := 0
	var node, name string
	for _, candidate := range []struct{ node, value string }{
		{"person", manifest.Person},
		{"roster", manifest.Roster},
		{"library", manifest.Library},
	} {
		if candidate.value != "" {
			declared++
			node, name = candidate.node, candidate.value
		}
	}
	if declared != 1 {
		return "", "", fmt.Errorf("manifest needs exactly one of person, roster, or library")
	}
	return node, name, nil
}
