package tflint

import "github.com/jokarl/tfbreak-plugin-sdk/hclext"

// Rule is the interface that all tfbreak rules must implement.
// This aligns with tflint-plugin-sdk's Rule interface for ecosystem familiarity.
//
// Plugin authors typically embed DefaultRule to get default implementations
// for Enabled() and Severity(), then implement the remaining methods.
//
// Example:
//
//	type MyRule struct {
//	    tflint.DefaultRule
//	}
//
//	func (r *MyRule) Name() string { return "my_rule" }
//	func (r *MyRule) Link() string { return "https://example.com/my_rule" }
//	func (r *MyRule) Check(runner tflint.Runner) error {
//	    // Access old and new configs via runner
//	    oldContent, _ := runner.GetOldResourceContent("azurerm_resource_group", schema, nil)
//	    newContent, _ := runner.GetNewResourceContent("azurerm_resource_group", schema, nil)
//	    // Compare and emit issues
//	    return nil
//	}
type Rule interface {
	// Name returns the unique name of the rule.
	// Convention: lowercase with underscores (e.g., "azurerm_force_new").
	Name() string

	// Enabled returns whether the rule is enabled by default.
	// Most rules return true; embed DefaultRule for this behavior.
	Enabled() bool

	// Severity returns the default severity level for issues.
	// Most rules return ERROR; embed DefaultRule for this behavior.
	Severity() Severity

	// Link returns a URL to documentation about the rule.
	// Should explain what the rule checks and how to resolve issues.
	Link() string

	// Check executes the rule against the configurations accessible via runner.
	// Call runner.EmitIssue() for each finding.
	// Return an error only for unexpected failures, not for findings.
	Check(runner Runner) error
}

// RuleSet is implemented by plugins to provide a collection of rules.
// Plugins typically embed BuiltinRuleSet and override methods as needed.
//
// Example:
//
//	type MyRuleSet struct {
//	    tflint.BuiltinRuleSet
//	}
//
//	func main() {
//	    plugin.Serve(&plugin.ServeOpts{
//	        RuleSet: &MyRuleSet{
//	            BuiltinRuleSet: tflint.BuiltinRuleSet{
//	                Name:    "myplugin",
//	                Version: "0.1.0",
//	                Rules:   []tflint.Rule{&MyRule{}},
//	            },
//	        },
//	    })
//	}
type RuleSet interface {
	// RuleSetName returns the name of the ruleset (e.g., "azurerm").
	RuleSetName() string

	// RuleSetVersion returns the version of the ruleset (e.g., "0.1.0").
	RuleSetVersion() string

	// RuleNames returns the names of all rules in this ruleset.
	RuleNames() []string

	// VersionConstraint returns the tfbreak version constraint (e.g., ">= 0.1.0").
	VersionConstraint() string

	// ConfigSchema returns the schema for plugin-specific configuration.
	// Return nil if no configuration is needed.
	ConfigSchema() *hclext.BodySchema

	// ApplyGlobalConfig applies global tfbreak configuration.
	// Called before ApplyConfig.
	ApplyGlobalConfig(*Config) error

	// ApplyConfig applies plugin-specific configuration.
	// The content matches the schema from ConfigSchema().
	ApplyConfig(*hclext.BodyContent) error

	// NewRunner optionally wraps the runner with custom behavior.
	// Return the runner unchanged if no customization is needed.
	NewRunner(Runner) (Runner, error)

	// BuiltinImpl returns the embedded BuiltinRuleSet.
	// Used internally for rule iteration.
	BuiltinImpl() *BuiltinRuleSet
}
