package tflint

import "github.com/jokarl/tfbreak-plugin-sdk/hclext"

// BuiltinRuleSet provides default implementations for the RuleSet interface.
// Plugin authors embed this struct and override methods as needed.
//
// Example:
//
//	type AzurermRuleSet struct {
//	    tflint.BuiltinRuleSet
//	}
//
//	rs := &AzurermRuleSet{
//	    BuiltinRuleSet: tflint.BuiltinRuleSet{
//	        Name:       "azurerm",
//	        Version:    "0.1.0",
//	        Constraint: ">= 0.1.0",
//	        Rules:      []tflint.Rule{&ForceNewRule{}},
//	    },
//	}
type BuiltinRuleSet struct {
	// Name is the ruleset name (e.g., "azurerm").
	Name string
	// Version is the ruleset version (e.g., "0.1.0").
	Version string
	// Constraint is the tfbreak version constraint (e.g., ">= 0.1.0").
	Constraint string
	// Rules is the list of rules in this ruleset.
	Rules []Rule
	// enabledRules tracks which rules are enabled after configuration.
	enabledRules map[string]bool
}

// RuleSetName returns the name of the ruleset.
func (rs *BuiltinRuleSet) RuleSetName() string {
	return rs.Name
}

// RuleSetVersion returns the version of the ruleset.
func (rs *BuiltinRuleSet) RuleSetVersion() string {
	return rs.Version
}

// RuleNames returns the names of all rules in this ruleset.
func (rs *BuiltinRuleSet) RuleNames() []string {
	names := make([]string, len(rs.Rules))
	for i, rule := range rs.Rules {
		names[i] = rule.Name()
	}
	return names
}

// VersionConstraint returns the tfbreak version constraint.
func (rs *BuiltinRuleSet) VersionConstraint() string {
	if rs.Constraint == "" {
		return ">= 0.1.0"
	}
	return rs.Constraint
}

// ConfigSchema returns nil (no plugin-specific configuration by default).
// Override this method to define custom plugin configuration.
func (rs *BuiltinRuleSet) ConfigSchema() *hclext.BodySchema {
	return nil
}

// ApplyGlobalConfig applies global tfbreak configuration.
// Handles DisabledByDefault and Only filtering.
func (rs *BuiltinRuleSet) ApplyGlobalConfig(config *Config) error {
	rs.enabledRules = make(map[string]bool)

	// Initialize with rule defaults
	for _, rule := range rs.Rules {
		rs.enabledRules[rule.Name()] = rule.Enabled()
	}

	if config == nil {
		return nil
	}

	// Handle DisabledByDefault
	if config.DisabledByDefault {
		for name := range rs.enabledRules {
			rs.enabledRules[name] = false
		}
	}

	// Handle Only filter
	if len(config.Only) > 0 {
		for name := range rs.enabledRules {
			rs.enabledRules[name] = false
		}
		for _, name := range config.Only {
			if _, ok := rs.enabledRules[name]; ok {
				rs.enabledRules[name] = true
			}
		}
	}

	// Apply per-rule configuration
	for name, ruleConfig := range config.Rules {
		if _, ok := rs.enabledRules[name]; ok {
			rs.enabledRules[name] = ruleConfig.Enabled
		}
	}

	return nil
}

// ApplyConfig applies plugin-specific configuration.
// Default implementation does nothing.
// Override this method to handle custom plugin configuration.
func (rs *BuiltinRuleSet) ApplyConfig(_ *hclext.BodyContent) error {
	return nil
}

// NewRunner returns the runner unchanged by default.
// Override this method to wrap the runner with custom behavior.
func (rs *BuiltinRuleSet) NewRunner(runner Runner) (Runner, error) {
	return runner, nil
}

// BuiltinImpl returns the BuiltinRuleSet itself.
func (rs *BuiltinRuleSet) BuiltinImpl() *BuiltinRuleSet {
	return rs
}

// IsRuleEnabled returns whether a rule is enabled.
// Call this after ApplyGlobalConfig.
func (rs *BuiltinRuleSet) IsRuleEnabled(name string) bool {
	if rs.enabledRules == nil {
		// Not yet configured; use rule default
		for _, rule := range rs.Rules {
			if rule.Name() == name {
				return rule.Enabled()
			}
		}
		return false
	}
	return rs.enabledRules[name]
}

// GetRule returns a rule by name, or nil if not found.
func (rs *BuiltinRuleSet) GetRule(name string) Rule {
	for _, rule := range rs.Rules {
		if rule.Name() == name {
			return rule
		}
	}
	return nil
}

// EnabledRules returns all currently enabled rules.
func (rs *BuiltinRuleSet) EnabledRules() []Rule {
	var enabled []Rule
	for _, rule := range rs.Rules {
		if rs.IsRuleEnabled(rule.Name()) {
			enabled = append(enabled, rule)
		}
	}
	return enabled
}
