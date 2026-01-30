// Package helper provides testing utilities for tfbreak plugins.
// Use TestRunner to test rules without running tfbreak-core.
//
// Example:
//
//	func TestMyRule(t *testing.T) {
//	    runner := helper.TestRunner(t,
//	        map[string]string{"main.tf": `resource "azurerm_rg" "test" { location = "westus" }`},
//	        map[string]string{"main.tf": `resource "azurerm_rg" "test" { location = "eastus" }`},
//	    )
//
//	    rule := &MyRule{}
//	    if err := rule.Check(runner); err != nil {
//	        t.Fatal(err)
//	    }
//
//	    helper.AssertIssues(t, helper.Issues{
//	        {Rule: rule, Message: "location changed"},
//	    }, runner.Issues)
//	}
package helper

import (
	"testing"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclparse"
	"github.com/jokarl/tfbreak-plugin-sdk/hclext"
	"github.com/jokarl/tfbreak-plugin-sdk/tflint"
)

// Runner is a mock tflint.Runner for testing.
// Use TestRunner to create an instance.
type Runner struct {
	t        *testing.T
	oldFiles map[string]*hcl.File
	newFiles map[string]*hcl.File
	// Issues contains all issues emitted during rule execution.
	Issues Issues
}

// Ensure Runner implements tflint.Runner.
var _ tflint.Runner = (*Runner)(nil)

// TestRunner creates a new Runner for testing.
//
// DEVIATION FROM TFLINT (see ADR-0001):
// Unlike tflint's helper.TestRunner which takes a single file map,
// tfbreak's TestRunner takes two maps: old (baseline) and new (changed).
//
// Example:
//
//	runner := helper.TestRunner(t,
//	    map[string]string{
//	        "main.tf": `resource "azurerm_resource_group" "rg" { location = "westus" }`,
//	    },
//	    map[string]string{
//	        "main.tf": `resource "azurerm_resource_group" "rg" { location = "eastus" }`,
//	    },
//	)
//
//	rule := &MyRule{}
//	rule.Check(runner)
//	helper.AssertIssues(t, expected, runner.Issues)
func TestRunner(t *testing.T, oldFiles, newFiles map[string]string) *Runner {
	t.Helper()

	runner := &Runner{
		t:        t,
		oldFiles: make(map[string]*hcl.File),
		newFiles: make(map[string]*hcl.File),
		Issues:   make(Issues, 0),
	}

	parser := hclparse.NewParser()

	// Parse old files
	for name, content := range oldFiles {
		file, diags := parser.ParseHCL([]byte(content), name)
		if diags.HasErrors() {
			t.Fatalf("failed to parse old file %s: %s", name, diags.Error())
		}
		runner.oldFiles[name] = file
	}

	// Parse new files
	for name, content := range newFiles {
		file, diags := parser.ParseHCL([]byte(content), name)
		if diags.HasErrors() {
			t.Fatalf("failed to parse new file %s: %s", name, diags.Error())
		}
		runner.newFiles[name] = file
	}

	return runner
}

// GetOldModuleContent retrieves content from old files.
func (r *Runner) GetOldModuleContent(schema *hclext.BodySchema, _ *tflint.GetModuleContentOption) (*hclext.BodyContent, error) {
	return r.getModuleContent(r.oldFiles, schema)
}

// GetNewModuleContent retrieves content from new files.
func (r *Runner) GetNewModuleContent(schema *hclext.BodySchema, _ *tflint.GetModuleContentOption) (*hclext.BodyContent, error) {
	return r.getModuleContent(r.newFiles, schema)
}

// GetOldResourceContent retrieves resources of a specific type from old files.
func (r *Runner) GetOldResourceContent(resourceType string, schema *hclext.BodySchema, _ *tflint.GetModuleContentOption) (*hclext.BodyContent, error) {
	return r.getResourceContent(r.oldFiles, resourceType, schema)
}

// GetNewResourceContent retrieves resources of a specific type from new files.
func (r *Runner) GetNewResourceContent(resourceType string, schema *hclext.BodySchema, _ *tflint.GetModuleContentOption) (*hclext.BodyContent, error) {
	return r.getResourceContent(r.newFiles, resourceType, schema)
}

// EmitIssue records an issue.
func (r *Runner) EmitIssue(rule tflint.Rule, message string, issueRange hcl.Range) error {
	r.Issues = append(r.Issues, Issue{
		Rule:    rule,
		Message: message,
		Range:   issueRange,
	})
	return nil
}

// DecodeRuleConfig decodes rule configuration.
// This is a stub implementation that always returns nil (no config).
func (r *Runner) DecodeRuleConfig(_ string, _ any) error {
	return nil
}

// getModuleContent extracts content from files using the schema.
func (r *Runner) getModuleContent(files map[string]*hcl.File, schema *hclext.BodySchema) (*hclext.BodyContent, error) {
	content := &hclext.BodyContent{
		Attributes: make(map[string]*hclext.Attribute),
		Blocks:     make([]*hclext.Block, 0),
	}

	hclSchema := hclext.ToHCLBodySchema(schema)

	for _, file := range files {
		bodyContent, _, diags := file.Body.PartialContent(hclSchema)
		if diags.HasErrors() {
			return nil, diags
		}

		// Merge attributes
		for name, attr := range bodyContent.Attributes {
			content.Attributes[name] = hclext.FromHCLAttribute(attr)
		}

		// Append blocks
		for _, block := range bodyContent.Blocks {
			b := hclext.FromHCLBlock(block)
			// Process nested body if schema specifies it
			if schema != nil {
				for _, bs := range schema.Blocks {
					if bs.Type == block.Type && bs.Body != nil {
						nestedContent, _ := r.extractBlockContent(block.Body, bs.Body)
						b.Body = nestedContent
					}
				}
			}
			content.Blocks = append(content.Blocks, b)
		}
	}

	return content, nil
}

// getResourceContent extracts resources of a specific type.
func (r *Runner) getResourceContent(files map[string]*hcl.File, resourceType string, bodySchema *hclext.BodySchema) (*hclext.BodyContent, error) {
	// Create a schema that looks for resource blocks
	resourceSchema := &hclext.BodySchema{
		Blocks: []hclext.BlockSchema{
			{
				Type:       "resource",
				LabelNames: []string{"type", "name"},
				Body:       bodySchema,
			},
		},
	}

	allContent, err := r.getModuleContent(files, resourceSchema)
	if err != nil {
		return nil, err
	}

	// Filter to only the requested resource type
	result := &hclext.BodyContent{
		Attributes: make(map[string]*hclext.Attribute),
		Blocks:     make([]*hclext.Block, 0),
	}

	for _, block := range allContent.Blocks {
		if block.Type == "resource" && len(block.Labels) >= 1 && block.Labels[0] == resourceType {
			result.Blocks = append(result.Blocks, block)
		}
	}

	return result, nil
}

// extractBlockContent extracts nested block content.
func (r *Runner) extractBlockContent(body hcl.Body, schema *hclext.BodySchema) (*hclext.BodyContent, error) {
	if body == nil || schema == nil {
		return nil, nil
	}

	hclSchema := hclext.ToHCLBodySchema(schema)
	bodyContent, _, diags := body.PartialContent(hclSchema)
	if diags.HasErrors() {
		return nil, diags
	}

	return hclext.FromHCLBodyContent(bodyContent), nil
}
