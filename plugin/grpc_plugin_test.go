package plugin

import (
	"testing"

	"github.com/hashicorp/hcl/v2"

	"github.com/jokarl/tfbreak-plugin-sdk/hclext"
	"github.com/jokarl/tfbreak-plugin-sdk/tflint"
)

// mockRuleSet is a mock implementation for testing.
type mockRuleSet struct {
	tflint.BuiltinRuleSet
}

func TestGRPCRuleSetClientImplementsRuleSet(t *testing.T) {
	// This is a compile-time check that GRPCRuleSetClient implements tflint.RuleSet.
	// If this doesn't compile, the interface isn't properly implemented.
	var _ tflint.RuleSet = (*GRPCRuleSetClient)(nil)
}

func TestRuleSetPluginImplementsGRPCPlugin(t *testing.T) {
	// This is a compile-time check that RuleSetPlugin implements plugin.GRPCPlugin.
	// The actual check is in grpc_plugin.go via the var _ declaration.
	p := &RuleSetPlugin{
		Impl: &mockRuleSet{
			BuiltinRuleSet: tflint.BuiltinRuleSet{
				Name:    "test",
				Version: "1.0.0",
			},
		},
	}
	if p.Impl == nil {
		t.Error("Impl should not be nil")
	}
}

func TestGRPCRuleSetClientMethods(t *testing.T) {
	// Test that the client methods handle nil broker gracefully
	client := &GRPCRuleSetClient{
		client: nil, // nil client - methods will panic or fail
		broker: nil,
	}

	// NewRunner should work even without a real client
	runner := &mockRunner{}
	result, err := client.NewRunner(runner)
	if err != nil {
		t.Errorf("NewRunner returned error: %v", err)
	}
	if result != runner {
		t.Error("NewRunner should return the same runner on client side")
	}

	// BuiltinImpl should return nil on client side
	if client.BuiltinImpl() != nil {
		t.Error("BuiltinImpl should return nil on client side")
	}
}

// Note: TestGRPCRuleSetClientConfigSchema and TestGRPCRuleSetClientApplyConfig
// are not included because they would require a full gRPC server setup.
// The actual gRPC communication is tested via integration tests.

func TestRunnerBrokerID(t *testing.T) {
	// Verify the broker ID is a reasonable value
	if RunnerBrokerID == 0 {
		t.Error("RunnerBrokerID should not be 0")
	}
}

// mockRunner is a minimal tflint.Runner implementation for testing.
type mockRunner struct{}

func (r *mockRunner) GetOldModuleContent(schema *hclext.BodySchema, opts *tflint.GetModuleContentOption) (*hclext.BodyContent, error) {
	return &hclext.BodyContent{}, nil
}

func (r *mockRunner) GetNewModuleContent(schema *hclext.BodySchema, opts *tflint.GetModuleContentOption) (*hclext.BodyContent, error) {
	return &hclext.BodyContent{}, nil
}

func (r *mockRunner) GetOldResourceContent(resourceType string, schema *hclext.BodySchema, opts *tflint.GetModuleContentOption) (*hclext.BodyContent, error) {
	return &hclext.BodyContent{}, nil
}

func (r *mockRunner) GetNewResourceContent(resourceType string, schema *hclext.BodySchema, opts *tflint.GetModuleContentOption) (*hclext.BodyContent, error) {
	return &hclext.BodyContent{}, nil
}

func (r *mockRunner) EmitIssue(rule tflint.Rule, message string, issueRange hcl.Range) error {
	return nil
}

func (r *mockRunner) DecodeRuleConfig(ruleName string, target any) error {
	return nil
}
