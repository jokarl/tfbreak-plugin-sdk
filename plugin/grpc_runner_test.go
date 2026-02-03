package plugin

import (
	"testing"

	"github.com/hashicorp/hcl/v2"

	"github.com/jokarl/tfbreak-plugin-sdk/hclext"
	"github.com/jokarl/tfbreak-plugin-sdk/tflint"
)

func TestGRPCRunnerClientImplementsRunner(t *testing.T) {
	// Compile-time check that GRPCRunnerClient implements tflint.Runner
	var _ tflint.Runner = (*GRPCRunnerClient)(nil)
}

func TestProtoRuleImplementsRule(t *testing.T) {
	// Test the protoRule implementation
	rule := &protoRule{
		name:     "test_rule",
		enabled:  true,
		severity: tflint.WARNING,
		link:     "https://example.com",
	}

	if rule.Name() != "test_rule" {
		t.Errorf("Name() = %q, want %q", rule.Name(), "test_rule")
	}
	if !rule.Enabled() {
		t.Error("Enabled() should be true")
	}
	if rule.Severity() != tflint.WARNING {
		t.Errorf("Severity() = %v, want WARNING", rule.Severity())
	}
	if rule.Link() != "https://example.com" {
		t.Errorf("Link() = %q, want %q", rule.Link(), "https://example.com")
	}
	if err := rule.Check(nil); err != nil {
		t.Errorf("Check() returned error: %v", err)
	}
}

func TestProtoRuleDisabled(t *testing.T) {
	rule := &protoRule{
		name:     "disabled_rule",
		enabled:  false,
		severity: tflint.NOTICE,
		link:     "",
	}

	if rule.Enabled() {
		t.Error("Enabled() should be false")
	}
	if rule.Link() != "" {
		t.Errorf("Link() = %q, want empty string", rule.Link())
	}
}

func TestGRPCRunnerServerEmitIssue(t *testing.T) {
	// Create a mock runner that records emitted issues
	issues := make([]struct {
		rule    tflint.Rule
		message string
		r       hcl.Range
	}, 0)

	runner := &recordingRunner{
		onEmitIssue: func(rule tflint.Rule, message string, issueRange hcl.Range) error {
			issues = append(issues, struct {
				rule    tflint.Rule
				message string
				r       hcl.Range
			}{rule, message, issueRange})
			return nil
		},
	}

	server := &GRPCRunnerServer{impl: runner}

	// Test would need a full gRPC setup to call EmitIssue
	// For now, just verify the server can be created
	if server.impl == nil {
		t.Error("Server impl should not be nil")
	}
}

// recordingRunner records calls for testing
type recordingRunner struct {
	onGetOldModuleContent   func(*hclext.BodySchema, *tflint.GetModuleContentOption) (*hclext.BodyContent, error)
	onGetNewModuleContent   func(*hclext.BodySchema, *tflint.GetModuleContentOption) (*hclext.BodyContent, error)
	onGetOldResourceContent func(string, *hclext.BodySchema, *tflint.GetModuleContentOption) (*hclext.BodyContent, error)
	onGetNewResourceContent func(string, *hclext.BodySchema, *tflint.GetModuleContentOption) (*hclext.BodyContent, error)
	onEmitIssue             func(tflint.Rule, string, hcl.Range) error
	onDecodeRuleConfig      func(string, any) error
}

func (r *recordingRunner) GetOldModuleContent(schema *hclext.BodySchema, opts *tflint.GetModuleContentOption) (*hclext.BodyContent, error) {
	if r.onGetOldModuleContent != nil {
		return r.onGetOldModuleContent(schema, opts)
	}
	return &hclext.BodyContent{Attributes: map[string]*hclext.Attribute{}, Blocks: []*hclext.Block{}}, nil
}

func (r *recordingRunner) GetNewModuleContent(schema *hclext.BodySchema, opts *tflint.GetModuleContentOption) (*hclext.BodyContent, error) {
	if r.onGetNewModuleContent != nil {
		return r.onGetNewModuleContent(schema, opts)
	}
	return &hclext.BodyContent{Attributes: map[string]*hclext.Attribute{}, Blocks: []*hclext.Block{}}, nil
}

func (r *recordingRunner) GetOldResourceContent(resourceType string, schema *hclext.BodySchema, opts *tflint.GetModuleContentOption) (*hclext.BodyContent, error) {
	if r.onGetOldResourceContent != nil {
		return r.onGetOldResourceContent(resourceType, schema, opts)
	}
	return &hclext.BodyContent{Attributes: map[string]*hclext.Attribute{}, Blocks: []*hclext.Block{}}, nil
}

func (r *recordingRunner) GetNewResourceContent(resourceType string, schema *hclext.BodySchema, opts *tflint.GetModuleContentOption) (*hclext.BodyContent, error) {
	if r.onGetNewResourceContent != nil {
		return r.onGetNewResourceContent(resourceType, schema, opts)
	}
	return &hclext.BodyContent{Attributes: map[string]*hclext.Attribute{}, Blocks: []*hclext.Block{}}, nil
}

func (r *recordingRunner) EmitIssue(rule tflint.Rule, message string, issueRange hcl.Range) error {
	if r.onEmitIssue != nil {
		return r.onEmitIssue(rule, message, issueRange)
	}
	return nil
}

func (r *recordingRunner) DecodeRuleConfig(ruleName string, target any) error {
	if r.onDecodeRuleConfig != nil {
		return r.onDecodeRuleConfig(ruleName, target)
	}
	return nil
}
