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
