package tflint

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/jokarl/tfbreak-plugin-sdk/hclext"
)

// Runner provides access to Terraform configurations during rule execution.
//
// DEVIATION FROM TFLINT (see ADR-0001):
// Unlike tflint's Runner which accesses a single configuration, tfbreak's
// Runner provides dual-config access via GetOld* and GetNew* methods.
// This is because tfbreak compares two configurations to detect breaking changes.
//
//	tflint:  runner.GetResourceContent(...)     // Single config
//	tfbreak: runner.GetOldResourceContent(...)  // Old (baseline) config
//	         runner.GetNewResourceContent(...)  // New (changed) config
//
// The old configuration represents the baseline state before the change.
// The new configuration represents the state after the change.
// Rules compare these to detect breaking changes.
type Runner interface {
	// GetOldModuleContent retrieves module content from the OLD (baseline) configuration.
	// Use this to access the state before the change.
	GetOldModuleContent(schema *hclext.BodySchema, opts *GetModuleContentOption) (*hclext.BodyContent, error)

	// GetNewModuleContent retrieves module content from the NEW configuration.
	// Use this to access the state after the change.
	GetNewModuleContent(schema *hclext.BodySchema, opts *GetModuleContentOption) (*hclext.BodyContent, error)

	// GetOldResourceContent retrieves resources of a specific type from the OLD configuration.
	// This is a convenience method that filters module content for a specific resource type.
	//
	// Example:
	//
	//	content, err := runner.GetOldResourceContent("azurerm_resource_group", &hclext.BodySchema{
	//	    Attributes: []hclext.AttributeSchema{
	//	        {Name: "location", Required: true},
	//	    },
	//	}, nil)
	GetOldResourceContent(resourceType string, schema *hclext.BodySchema, opts *GetModuleContentOption) (*hclext.BodyContent, error)

	// GetNewResourceContent retrieves resources of a specific type from the NEW configuration.
	// This is a convenience method that filters module content for a specific resource type.
	//
	// Example:
	//
	//	content, err := runner.GetNewResourceContent("azurerm_resource_group", &hclext.BodySchema{
	//	    Attributes: []hclext.AttributeSchema{
	//	        {Name: "location", Required: true},
	//	    },
	//	}, nil)
	GetNewResourceContent(resourceType string, schema *hclext.BodySchema, opts *GetModuleContentOption) (*hclext.BodyContent, error)

	// EmitIssue reports a finding from the rule.
	// The issueRange should point to the relevant location in the NEW configuration.
	// For breaking changes, this is typically where the problematic change was made.
	//
	// Example:
	//
	//	if oldLocation != newLocation {
	//	    runner.EmitIssue(rule, "location: ForceNew attribute changed", newAttr.Range)
	//	}
	EmitIssue(rule Rule, message string, issueRange hcl.Range) error

	// DecodeRuleConfig retrieves and decodes the rule's configuration.
	// The target should be a pointer to a struct with hcl tags.
	// Returns nil if no configuration is provided for the rule.
	//
	// Example:
	//
	//	type MyRuleConfig struct {
	//	    IgnorePatterns []string `hcl:"ignore_patterns,optional"`
	//	}
	//	var config MyRuleConfig
	//	if err := runner.DecodeRuleConfig("my_rule", &config); err != nil {
	//	    return err
	//	}
	DecodeRuleConfig(ruleName string, target any) error
}

// GetModuleContentOption configures how content is retrieved.
// Aligns with tflint-plugin-sdk's option struct.
type GetModuleContentOption struct {
	// ModuleCtx specifies which module context to use.
	ModuleCtx ModuleCtxType
	// ExpandMode specifies how to handle dynamic blocks.
	ExpandMode ExpandMode
	// Hint provides hints for optimization.
	Hint GetModuleContentHint
}

// ModuleCtxType specifies the module context for content retrieval.
type ModuleCtxType int

const (
	// ModuleCtxSelf retrieves content from the current module only.
	ModuleCtxSelf ModuleCtxType = iota
	// ModuleCtxRoot retrieves content from the root module.
	ModuleCtxRoot
	// ModuleCtxAll retrieves content from all modules.
	ModuleCtxAll
)

// ExpandMode specifies how dynamic blocks are handled.
type ExpandMode int

const (
	// ExpandModeNone does not expand dynamic blocks.
	ExpandModeNone ExpandMode = iota
	// ExpandModeExpand expands dynamic blocks (not currently implemented).
	ExpandModeExpand
)

// GetModuleContentHint provides optimization hints for content retrieval.
type GetModuleContentHint struct {
	// ResourceType hints at the expected resource type for optimization.
	ResourceType string
}
