package person

import (
	"bytes"
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"
)

// YAML fragments are re-emitted as the KDL the assembler concatenates, so the
// semantic parser keeps one grammar. See docs/person-packages.md.

const yamlFragmentExt = ".yaml"

type yamlManifest struct {
	Person  string `yaml:"person"`
	Roster  string `yaml:"roster"`
	Library string `yaml:"library"`
}

type yamlSeat struct {
	Key      string `yaml:"key"`
	Name     string `yaml:"name"`
	Pronouns string `yaml:"pronouns"`
	Channel  string `yaml:"channel"`
	Tier     string `yaml:"tier"`
}

type yamlAgent struct {
	Harness  string `yaml:"harness"`
	Name     string `yaml:"name"`
	Pronouns string `yaml:"pronouns"`
	Channel  string `yaml:"channel"`
	Tier     string `yaml:"tier"`
}

type yamlForbid struct {
	Term   string `yaml:"term"`
	Prefer string `yaml:"prefer"`
}

type yamlCopyContract struct {
	Scope  string       `yaml:"scope"`
	Forbid []yamlForbid `yaml:"forbid"`
}

type yamlRole struct {
	Role         string            `yaml:"role"`
	Purpose      string            `yaml:"purpose"`
	Briefing     string            `yaml:"briefing"`
	Skill        string            `yaml:"skill"`
	ModelTier    []string          `yaml:"model_tier"`
	Personality  []string          `yaml:"personality"`
	CopyContract *yamlCopyContract `yaml:"copy_contract"`
	Seats        []yamlSeat        `yaml:"seats"`
	Agents       []yamlAgent       `yaml:"agents"`
}

type yamlEmblem struct {
	Name  []string `yaml:"name"`
	Emoji string   `yaml:"emoji"`
}

type yamlBody struct {
	Archetype  string `yaml:"archetype"`
	Attachment string `yaml:"attachment"`
}

type yamlSoundMark struct {
	Timbre  string `yaml:"timbre"`
	Contour string `yaml:"contour"`
	Pulse   string `yaml:"pulse"`
}

type yamlPersonality struct {
	Personality string         `yaml:"personality"`
	Skill       string         `yaml:"skill"`
	Color       string         `yaml:"color"`
	Motif       string         `yaml:"motif"`
	Geometry    string         `yaml:"geometry"`
	Alias       string         `yaml:"alias"`
	Emblem      *yamlEmblem    `yaml:"emblem"`
	Body        *yamlBody      `yaml:"body"`
	SoundMark   *yamlSoundMark `yaml:"sound_mark"`
}

// decodeYAML rejects unknown keys, matching the KDL parsers, so a typo fails
// rather than dropping the field it was meant to set.
func decodeYAML(raw []byte, into any) error {
	decoder := yaml.NewDecoder(bytes.NewReader(raw))
	decoder.KnownFields(true)
	if err := decoder.Decode(into); err != nil {
		return err
	}
	var trailing yaml.Node
	if err := decoder.Decode(&trailing); err == nil {
		return fmt.Errorf("needs exactly one document")
	}
	return nil
}

// yamlManifestKDL emits the single manifest node the assembler expects.
func yamlManifestKDL(raw []byte) ([]byte, error) {
	var manifest yamlManifest
	if err := decodeYAML(raw, &manifest); err != nil {
		return nil, err
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
		return nil, fmt.Errorf("manifest needs exactly one of person, roster, or library")
	}
	return []byte(fmt.Sprintf("%s %s", node, kdlString(name))), nil
}

// yamlFragmentKDL emits one entity node for the named section.
func yamlFragmentKDL(raw []byte, node string) ([]byte, error) {
	switch node {
	case "role":
		return yamlRoleKDL(raw)
	case "personality":
		return yamlPersonalityKDL(raw)
	}
	// Boundary fragments have no YAML fixture to prove a shape against, so they
	// stay KDL-authorable rather than shipping an untested emitter. #335.
	return nil, fmt.Errorf("%s fragments are not authorable as YAML yet", node)
}

func yamlRoleKDL(raw []byte) ([]byte, error) {
	var role yamlRole
	if err := decodeYAML(raw, &role); err != nil {
		return nil, err
	}
	if role.Role == "" {
		return nil, fmt.Errorf("role fragment needs a role name")
	}
	var out bytes.Buffer
	fmt.Fprintf(&out, "role %s {\n", kdlString(role.Role))
	writeChildString(&out, "purpose", role.Purpose)
	writeChildString(&out, "briefing", role.Briefing)
	writeChildList(&out, "model-tier", role.ModelTier)
	writeChildString(&out, "skill", role.Skill)
	writeChildList(&out, "personality", role.Personality)
	if contract := role.CopyContract; contract != nil {
		fmt.Fprintf(&out, "    copy-contract scope=%s {\n", kdlString(contract.Scope))
		for _, forbid := range contract.Forbid {
			fmt.Fprintf(
				&out,
				"        forbid %s prefer=%s\n",
				kdlString(forbid.Term),
				kdlString(forbid.Prefer),
			)
		}
		out.WriteString("    }\n")
	}
	for _, seat := range role.Seats {
		fmt.Fprintf(&out, "    seat %s", kdlString(seat.Key))
		writeProp(&out, "name", seat.Name)
		writeProp(&out, "pronouns", seat.Pronouns)
		writeProp(&out, "channel", seat.Channel)
		writeProp(&out, "tier", seat.Tier)
		out.WriteByte('\n')
	}
	for _, agent := range role.Agents {
		fmt.Fprintf(&out, "    agent %s", kdlString(agent.Harness))
		writeProp(&out, "name", agent.Name)
		writeProp(&out, "pronouns", agent.Pronouns)
		writeProp(&out, "channel", agent.Channel)
		writeProp(&out, "tier", agent.Tier)
		out.WriteByte('\n')
	}
	out.WriteString("}")
	return out.Bytes(), nil
}

func yamlPersonalityKDL(raw []byte) ([]byte, error) {
	var personality yamlPersonality
	if err := decodeYAML(raw, &personality); err != nil {
		return nil, err
	}
	if personality.Personality == "" {
		return nil, fmt.Errorf("personality fragment needs a personality name")
	}
	var out bytes.Buffer
	fmt.Fprintf(&out, "personality %s", kdlString(personality.Personality))
	writeProp(&out, "skill", personality.Skill)
	writeProp(&out, "color", personality.Color)
	writeProp(&out, "motif", personality.Motif)
	writeProp(&out, "geometry", personality.Geometry)
	out.WriteString(" {\n")
	if emblem := personality.Emblem; emblem != nil {
		out.WriteString("    emblem {\n")
		writeGrandchildList(&out, "name", emblem.Name)
		writeGrandchildString(&out, "emoji", emblem.Emoji)
		out.WriteString("    }\n")
	}
	if body := personality.Body; body != nil {
		out.WriteString("    body {\n")
		writeGrandchildString(&out, "archetype", body.Archetype)
		writeGrandchildString(&out, "attachment", body.Attachment)
		out.WriteString("    }\n")
	}
	if mark := personality.SoundMark; mark != nil {
		out.WriteString("    sound-mark {\n")
		writeGrandchildString(&out, "timbre", mark.Timbre)
		writeGrandchildString(&out, "contour", mark.Contour)
		writeGrandchildString(&out, "pulse", mark.Pulse)
		out.WriteString("    }\n")
	}
	writeChildString(&out, "alias", personality.Alias)
	out.WriteString("}")
	return out.Bytes(), nil
}

func writeChildString(out *bytes.Buffer, name, value string) {
	if value == "" {
		return
	}
	fmt.Fprintf(out, "    %s %s\n", name, kdlString(value))
}

func writeChildList(out *bytes.Buffer, name string, values []string) {
	if len(values) == 0 {
		return
	}
	quoted := make([]string, 0, len(values))
	for _, value := range values {
		quoted = append(quoted, kdlString(value))
	}
	fmt.Fprintf(out, "    %s %s\n", name, strings.Join(quoted, " "))
}

func writeGrandchildString(out *bytes.Buffer, name, value string) {
	if value == "" {
		return
	}
	fmt.Fprintf(out, "        %s %s\n", name, kdlString(value))
}

func writeGrandchildList(out *bytes.Buffer, name string, values []string) {
	if len(values) == 0 {
		return
	}
	quoted := make([]string, 0, len(values))
	for _, value := range values {
		quoted = append(quoted, kdlString(value))
	}
	fmt.Fprintf(out, "        %s %s\n", name, strings.Join(quoted, " "))
}

func writeProp(out *bytes.Buffer, name, value string) {
	if value == "" {
		return
	}
	fmt.Fprintf(out, " %s=%s", name, kdlString(value))
}

// kdlString escapes to KDL's quoted form. Newlines survive as \n rather than a
// multi-line literal, so an authored briefing needs no dedent contract.
func kdlString(value string) string {
	var out strings.Builder
	out.WriteByte('"')
	for _, r := range value {
		switch r {
		case '"':
			out.WriteString(`\"`)
		case '\\':
			out.WriteString(`\\`)
		case '\n':
			out.WriteString(`\n`)
		case '\r':
			out.WriteString(`\r`)
		case '\t':
			out.WriteString(`\t`)
		default:
			out.WriteRune(r)
		}
	}
	out.WriteByte('"')
	return out.String()
}
