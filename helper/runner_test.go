package helper

import (
	"testing"

	"github.com/hashicorp/hcl/v2"
	"github.com/jokarl/tfbreak-plugin-sdk/hclext"
	"github.com/jokarl/tfbreak-plugin-sdk/tflint"
)

// testRule is a minimal rule for testing.
type testRule struct {
	tflint.DefaultRule
	name string
}

func (r *testRule) Name() string        { return r.name }
func (r *testRule) Link() string        { return "" }
func (r *testRule) Check(_ tflint.Runner) error { return nil }

func TestTestRunner_ParsesOldFiles(t *testing.T) {
	runner := TestRunner(t,
		map[string]string{
			"main.tf": `variable "test" { default = "old" }`,
		},
		map[string]string{
			"main.tf": `variable "test" { default = "new" }`,
		},
	)

	if len(runner.oldFiles) != 1 {
		t.Errorf("expected 1 old file, got %d", len(runner.oldFiles))
	}
	if runner.oldFiles["main.tf"] == nil {
		t.Error("expected main.tf in old files")
	}
}

func TestTestRunner_ParsesNewFiles(t *testing.T) {
	runner := TestRunner(t,
		map[string]string{
			"main.tf": `variable "test" { default = "old" }`,
		},
		map[string]string{
			"main.tf": `variable "test" { default = "new" }`,
		},
	)

	if len(runner.newFiles) != 1 {
		t.Errorf("expected 1 new file, got %d", len(runner.newFiles))
	}
	if runner.newFiles["main.tf"] == nil {
		t.Error("expected main.tf in new files")
	}
}

func TestTestRunner_MultipleFiles(t *testing.T) {
	runner := TestRunner(t,
		map[string]string{
			"main.tf":      `variable "a" {}`,
			"variables.tf": `variable "b" {}`,
		},
		map[string]string{
			"main.tf":      `variable "a" {}`,
			"variables.tf": `variable "b" {}`,
		},
	)

	if len(runner.oldFiles) != 2 {
		t.Errorf("expected 2 old files, got %d", len(runner.oldFiles))
	}
	if len(runner.newFiles) != 2 {
		t.Errorf("expected 2 new files, got %d", len(runner.newFiles))
	}
}

func TestTestRunner_EmptyFiles(t *testing.T) {
	runner := TestRunner(t,
		map[string]string{},
		map[string]string{},
	)

	if len(runner.oldFiles) != 0 {
		t.Errorf("expected 0 old files, got %d", len(runner.oldFiles))
	}
	if len(runner.newFiles) != 0 {
		t.Errorf("expected 0 new files, got %d", len(runner.newFiles))
	}
}

func TestRunner_GetOldResourceContent(t *testing.T) {
	runner := TestRunner(t,
		map[string]string{
			"main.tf": `
resource "azurerm_resource_group" "example" {
  name     = "my-rg"
  location = "westeurope"
}`,
		},
		map[string]string{
			"main.tf": `
resource "azurerm_resource_group" "example" {
  name     = "my-rg"
  location = "eastus"
}`,
		},
	)

	schema := &hclext.BodySchema{
		Attributes: []hclext.AttributeSchema{
			{Name: "name"},
			{Name: "location"},
		},
	}

	content, err := runner.GetOldResourceContent("azurerm_resource_group", schema, nil)
	if err != nil {
		t.Fatalf("GetOldResourceContent failed: %v", err)
	}

	if len(content.Blocks) != 1 {
		t.Fatalf("expected 1 block, got %d", len(content.Blocks))
	}

	block := content.Blocks[0]
	if block.Type != "resource" {
		t.Errorf("block type = %q, want %q", block.Type, "resource")
	}
	if len(block.Labels) < 2 || block.Labels[1] != "example" {
		t.Errorf("block labels = %v, want [azurerm_resource_group example]", block.Labels)
	}
}

func TestRunner_GetNewResourceContent(t *testing.T) {
	runner := TestRunner(t,
		map[string]string{
			"main.tf": `
resource "azurerm_resource_group" "example" {
  location = "westeurope"
}`,
		},
		map[string]string{
			"main.tf": `
resource "azurerm_resource_group" "example" {
  location = "eastus"
}`,
		},
	)

	schema := &hclext.BodySchema{
		Attributes: []hclext.AttributeSchema{
			{Name: "location"},
		},
	}

	content, err := runner.GetNewResourceContent("azurerm_resource_group", schema, nil)
	if err != nil {
		t.Fatalf("GetNewResourceContent failed: %v", err)
	}

	if len(content.Blocks) != 1 {
		t.Fatalf("expected 1 block, got %d", len(content.Blocks))
	}
}

func TestRunner_GetResourceContent_FiltersByType(t *testing.T) {
	runner := TestRunner(t,
		map[string]string{
			"main.tf": `
resource "azurerm_resource_group" "rg" {
  name = "my-rg"
}
resource "azurerm_virtual_network" "vnet" {
  name = "my-vnet"
}
resource "azurerm_resource_group" "rg2" {
  name = "my-rg-2"
}`,
		},
		map[string]string{},
	)

	schema := &hclext.BodySchema{
		Attributes: []hclext.AttributeSchema{
			{Name: "name"},
		},
	}

	content, err := runner.GetOldResourceContent("azurerm_resource_group", schema, nil)
	if err != nil {
		t.Fatalf("GetOldResourceContent failed: %v", err)
	}

	if len(content.Blocks) != 2 {
		t.Errorf("expected 2 resource_group blocks, got %d", len(content.Blocks))
	}

	for _, block := range content.Blocks {
		if block.Labels[0] != "azurerm_resource_group" {
			t.Errorf("unexpected resource type: %s", block.Labels[0])
		}
	}
}

func TestRunner_GetOldModuleContent(t *testing.T) {
	runner := TestRunner(t,
		map[string]string{
			"main.tf": `
variable "location" {
  default = "westeurope"
}`,
		},
		map[string]string{},
	)

	schema := &hclext.BodySchema{
		Blocks: []hclext.BlockSchema{
			{
				Type:       "variable",
				LabelNames: []string{"name"},
			},
		},
	}

	content, err := runner.GetOldModuleContent(schema, nil)
	if err != nil {
		t.Fatalf("GetOldModuleContent failed: %v", err)
	}

	if len(content.Blocks) != 1 {
		t.Fatalf("expected 1 variable block, got %d", len(content.Blocks))
	}

	if content.Blocks[0].Labels[0] != "location" {
		t.Errorf("variable name = %q, want %q", content.Blocks[0].Labels[0], "location")
	}
}

func TestRunner_EmitIssue(t *testing.T) {
	runner := TestRunner(t, map[string]string{}, map[string]string{})

	rule := &testRule{name: "test_rule"}
	issueRange := hcl.Range{
		Filename: "main.tf",
		Start:    hcl.Pos{Line: 3, Column: 3},
		End:      hcl.Pos{Line: 3, Column: 20},
	}

	err := runner.EmitIssue(rule, "test message", issueRange)
	if err != nil {
		t.Fatalf("EmitIssue failed: %v", err)
	}

	if len(runner.Issues) != 1 {
		t.Fatalf("expected 1 issue, got %d", len(runner.Issues))
	}

	issue := runner.Issues[0]
	if issue.Rule.Name() != "test_rule" {
		t.Errorf("issue rule name = %q, want %q", issue.Rule.Name(), "test_rule")
	}
	if issue.Message != "test message" {
		t.Errorf("issue message = %q, want %q", issue.Message, "test message")
	}
	if issue.Range.Filename != "main.tf" {
		t.Errorf("issue range filename = %q, want %q", issue.Range.Filename, "main.tf")
	}
}

func TestRunner_EmitIssue_Multiple(t *testing.T) {
	runner := TestRunner(t, map[string]string{}, map[string]string{})

	rule := &testRule{name: "test_rule"}

	_ = runner.EmitIssue(rule, "issue 1", hcl.Range{})
	_ = runner.EmitIssue(rule, "issue 2", hcl.Range{})
	_ = runner.EmitIssue(rule, "issue 3", hcl.Range{})

	if len(runner.Issues) != 3 {
		t.Errorf("expected 3 issues, got %d", len(runner.Issues))
	}
}

func TestRunner_DecodeRuleConfig(t *testing.T) {
	runner := TestRunner(t, map[string]string{}, map[string]string{})

	var target struct{ Value string }
	err := runner.DecodeRuleConfig("test_rule", &target)
	if err != nil {
		t.Errorf("DecodeRuleConfig failed: %v", err)
	}
}

func TestRunner_ImplementsInterface(t *testing.T) {
	runner := TestRunner(t, map[string]string{}, map[string]string{})

	// Verify Runner satisfies tflint.Runner
	var _ tflint.Runner = runner
}

func TestRunner_GetResourceContent_DeeplyNested(t *testing.T) {
	// Test three levels of nesting: resource > blob_properties > cors_rule
	runner := TestRunner(t,
		map[string]string{
			"main.tf": `
resource "azurerm_storage_account" "example" {
  name = "storageacct"

  blob_properties {
    versioning_enabled = true

    cors_rule {
      allowed_methods = ["GET", "POST"]
      allowed_origins = ["*"]
    }
  }
}`,
		},
		map[string]string{},
	)

	// Three-level nested schema
	schema := &hclext.BodySchema{
		Attributes: []hclext.AttributeSchema{
			{Name: "name"},
		},
		Blocks: []hclext.BlockSchema{
			{
				Type: "blob_properties",
				Body: &hclext.BodySchema{
					Attributes: []hclext.AttributeSchema{
						{Name: "versioning_enabled"},
					},
					Blocks: []hclext.BlockSchema{
						{
							Type: "cors_rule",
							Body: &hclext.BodySchema{
								Attributes: []hclext.AttributeSchema{
									{Name: "allowed_methods"},
									{Name: "allowed_origins"},
								},
							},
						},
					},
				},
			},
		},
	}

	content, err := runner.GetOldResourceContent("azurerm_storage_account", schema, nil)
	if err != nil {
		t.Fatalf("GetOldResourceContent() error = %v", err)
	}

	if len(content.Blocks) != 1 {
		t.Fatalf("got %d resource blocks, want 1", len(content.Blocks))
	}

	resourceBlock := content.Blocks[0]
	if resourceBlock.Body == nil {
		t.Fatal("resource block body is nil")
	}

	// Find blob_properties
	var blobProps *hclext.Block
	for _, b := range resourceBlock.Body.Blocks {
		if b.Type == "blob_properties" {
			blobProps = b
			break
		}
	}
	if blobProps == nil {
		t.Fatal("blob_properties block not found")
	}
	if blobProps.Body == nil {
		t.Fatal("blob_properties body is nil")
	}

	// Verify versioning_enabled attribute
	if blobProps.Body.Attributes["versioning_enabled"] == nil {
		t.Error("versioning_enabled attribute not found in blob_properties")
	}

	// Find cors_rule (third level)
	var corsRule *hclext.Block
	for _, b := range blobProps.Body.Blocks {
		if b.Type == "cors_rule" {
			corsRule = b
			break
		}
	}
	if corsRule == nil {
		t.Fatal("cors_rule block not found (third level)")
	}
	if corsRule.Body == nil {
		t.Fatal("cors_rule body is nil")
	}

	// Verify deeply nested attributes
	if corsRule.Body.Attributes["allowed_methods"] == nil {
		t.Error("allowed_methods attribute not found in cors_rule")
	}
	if corsRule.Body.Attributes["allowed_origins"] == nil {
		t.Error("allowed_origins attribute not found in cors_rule")
	}
}

func TestLabelsMatch(t *testing.T) {
	tests := []struct {
		name     string
		a        []string
		b        []string
		expected bool
	}{
		{"both empty", []string{}, []string{}, true},
		{"both nil", nil, nil, true},
		{"same single", []string{"foo"}, []string{"foo"}, true},
		{"same multiple", []string{"foo", "bar"}, []string{"foo", "bar"}, true},
		{"different length", []string{"foo"}, []string{"foo", "bar"}, false},
		{"different values", []string{"foo"}, []string{"bar"}, false},
		{"one nil one empty", nil, []string{}, true},
		{"different order", []string{"foo", "bar"}, []string{"bar", "foo"}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := labelsMatch(tt.a, tt.b)
			if result != tt.expected {
				t.Errorf("labelsMatch(%v, %v) = %v, want %v", tt.a, tt.b, result, tt.expected)
			}
		})
	}
}
