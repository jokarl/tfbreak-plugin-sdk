package tflint

import "github.com/hashicorp/hcl/v2"

// Config represents global tfbreak configuration passed to plugins.
// This configuration is used to enable/disable rules and provide
// global settings.
type Config struct {
	// Rules maps rule names to their configuration.
	Rules map[string]*RuleConfig
	// DisabledByDefault indicates if rules are disabled by default.
	// When true, rules must be explicitly enabled.
	DisabledByDefault bool
	// Only enables only these rules if set.
	// Takes precedence over individual rule configurations.
	Only []string
	// PluginDir is the directory where plugins are installed.
	PluginDir string
}

// RuleConfig represents configuration for a single rule.
type RuleConfig struct {
	// Name is the rule name.
	Name string
	// Enabled indicates if the rule is enabled.
	Enabled bool
	// Body is the raw HCL body for rule-specific configuration.
	// Rules can decode this using runner.DecodeRuleConfig().
	Body hcl.Body
}
