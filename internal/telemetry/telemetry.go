// Package telemetry decides the OpenTelemetry environment a Claude seat
// launches with: token metrics tagged by seat, and never logs or events.
// See docs/claude-launch-identity.md.
package telemetry

import (
	"fmt"
	"net/url"
	"strings"
)

const (
	// EnvSwitch is the launch-time kill switch. "off" disables export.
	EnvSwitch = "AGENT_COMPOSE_TELEMETRY"
	// EnvShadow is the native session id AOS sets for a shadowed session.
	EnvShadow = "AOS_NATIVE_SESSION"

	// DefaultProtocol matches the collector's OTLP/HTTP listener.
	DefaultProtocol = "http/protobuf"
)

// Config is the `telemetry` block of the host configuration.
type Config struct {
	OTLPMetricsEndpoint string `yaml:"otlp_metrics_endpoint"`
	Protocol            string `yaml:"protocol"`
}

// Validate rejects an endpoint the exporter could not use.
func (c *Config) Validate() error {
	if c == nil || c.OTLPMetricsEndpoint == "" {
		return nil
	}
	parsed, err := url.Parse(c.OTLPMetricsEndpoint)
	if err != nil || (parsed.Scheme != "http" && parsed.Scheme != "https") || parsed.Host == "" {
		return fmt.Errorf("telemetry.otlp_metrics_endpoint %q must be an absolute http or https URL", c.OTLPMetricsEndpoint)
	}
	switch c.Protocol {
	case "", "http/protobuf", "http/json", "grpc":
	default:
		return fmt.Errorf("telemetry.protocol %q: want http/protobuf, http/json, or grpc", c.Protocol)
	}
	return nil
}

// Every variable a launch owns. An inherited value from a parent seat would
// otherwise leak its seat label or its enablement into this one.
var managed = []string{
	"CLAUDE_CODE_ENABLE_TELEMETRY",
	"CLAUDE_CODE_ENHANCED_TELEMETRY_BETA",
	"ENABLE_ENHANCED_TELEMETRY_BETA",
	"OTEL_METRICS_EXPORTER",
	"OTEL_LOGS_EXPORTER",
	"OTEL_TRACES_EXPORTER",
	"OTEL_EXPORTER_OTLP_ENDPOINT",
	"OTEL_EXPORTER_OTLP_PROTOCOL",
	"OTEL_EXPORTER_OTLP_METRICS_ENDPOINT",
	"OTEL_EXPORTER_OTLP_METRICS_PROTOCOL",
	"OTEL_EXPORTER_OTLP_LOGS_ENDPOINT",
	"OTEL_EXPORTER_OTLP_TRACES_ENDPOINT",
	"OTEL_LOG_USER_PROMPTS",
	"OTEL_LOG_ASSISTANT_RESPONSES",
	"OTEL_LOG_TOOL_DETAILS",
	"OTEL_LOG_TOOL_CONTENT",
	"OTEL_LOG_RAW_API_BODIES",
	"OTEL_RESOURCE_ATTRIBUTES",
}

// Plan is the environment change a launch applies: Set wins, Unset clears.
type Plan struct {
	Set   map[string]string
	Unset []string
}

// Environment returns the plan for one Claude seat. Disabled or endpointless,
// it clears every managed variable, so an exporting parent leaks nothing.
func Environment(cfg *Config, harness, seat, shadow, killSwitch string) Plan {
	if harness != "claude" {
		return Plan{}
	}
	plan := Plan{Set: map[string]string{}}
	enabled := cfg != nil && cfg.OTLPMetricsEndpoint != "" &&
		!strings.EqualFold(strings.TrimSpace(killSwitch), "off")
	if !enabled {
		plan.Unset = append(plan.Unset, managed...)
		return plan
	}
	protocol := cfg.Protocol
	if protocol == "" {
		protocol = DefaultProtocol
	}
	plan.Set["CLAUDE_CODE_ENABLE_TELEMETRY"] = "1"
	plan.Set["OTEL_METRICS_EXPORTER"] = "otlp"
	// Logs and events can carry prompt and tool text. Metrics cannot.
	plan.Set["OTEL_LOGS_EXPORTER"] = "none"
	plan.Set["OTEL_TRACES_EXPORTER"] = "none"
	plan.Set["OTEL_EXPORTER_OTLP_METRICS_ENDPOINT"] = cfg.OTLPMetricsEndpoint
	plan.Set["OTEL_EXPORTER_OTLP_METRICS_PROTOCOL"] = protocol
	attributes := []string{"seat=" + attributeValue(seat)}
	if strings.TrimSpace(shadow) != "" {
		attributes = append(attributes, "shadow="+attributeValue(shadow))
	}
	plan.Set["OTEL_RESOURCE_ATTRIBUTES"] = strings.Join(attributes, ",")
	for _, name := range managed {
		if _, set := plan.Set[name]; !set {
			plan.Unset = append(plan.Unset, name)
		}
	}
	return plan
}

// attributeValue percent-encodes what OTEL_RESOURCE_ATTRIBUTES forbids raw.
func attributeValue(value string) string {
	var out strings.Builder
	for _, b := range []byte(strings.TrimSpace(value)) {
		if b > 0x20 && b < 0x7f && !strings.ContainsRune(`",;\%=`, rune(b)) {
			out.WriteByte(b)
			continue
		}
		fmt.Fprintf(&out, "%%%02X", b)
	}
	return out.String()
}
