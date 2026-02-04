package plugin

import (
	"fmt"
	"testing"

	"github.com/hashicorp/hcl/v2"

	"github.com/jokarl/tfbreak-plugin-sdk/hclext"
	pb "github.com/jokarl/tfbreak-plugin-sdk/plugin/proto"
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

// =============================================================================
// GRPCRunnerServer method tests
// =============================================================================

func TestGRPCRunnerServer_GetOldModuleContent(t *testing.T) {
	expectedContent := &hclext.BodyContent{
		Attributes: map[string]*hclext.Attribute{
			"name": {Name: "name"},
		},
		Blocks: []*hclext.Block{},
	}

	runner := &recordingRunner{
		onGetOldModuleContent: func(schema *hclext.BodySchema, opts *tflint.GetModuleContentOption) (*hclext.BodyContent, error) {
			return expectedContent, nil
		},
	}
	server := &GRPCRunnerServer{impl: runner}

	resp, err := server.GetOldModuleContent(nil, &pb.GetModuleContent_Request{
		Schema: &pb.BodySchema{},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Content == nil {
		t.Fatal("expected content, got nil")
	}
	if _, ok := resp.Content.Attributes["name"]; !ok {
		t.Error("expected 'name' attribute in response")
	}
}

func TestGRPCRunnerServer_GetNewModuleContent(t *testing.T) {
	expectedContent := &hclext.BodyContent{
		Attributes: map[string]*hclext.Attribute{
			"version": {Name: "version"},
		},
		Blocks: []*hclext.Block{},
	}

	runner := &recordingRunner{
		onGetNewModuleContent: func(schema *hclext.BodySchema, opts *tflint.GetModuleContentOption) (*hclext.BodyContent, error) {
			return expectedContent, nil
		},
	}
	server := &GRPCRunnerServer{impl: runner}

	resp, err := server.GetNewModuleContent(nil, &pb.GetModuleContent_Request{
		Schema: &pb.BodySchema{},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Content == nil {
		t.Fatal("expected content, got nil")
	}
	if _, ok := resp.Content.Attributes["version"]; !ok {
		t.Error("expected 'version' attribute in response")
	}
}

func TestGRPCRunnerServer_GetOldResourceContent(t *testing.T) {
	var receivedResourceType string
	expectedContent := &hclext.BodyContent{
		Blocks: []*hclext.Block{
			{Type: "resource", Labels: []string{"aws_instance", "example"}},
		},
	}

	runner := &recordingRunner{
		onGetOldResourceContent: func(resourceType string, schema *hclext.BodySchema, opts *tflint.GetModuleContentOption) (*hclext.BodyContent, error) {
			receivedResourceType = resourceType
			return expectedContent, nil
		},
	}
	server := &GRPCRunnerServer{impl: runner}

	resp, err := server.GetOldResourceContent(nil, &pb.GetResourceContent_Request{
		ResourceType: "aws_instance",
		Schema:       &pb.BodySchema{},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if receivedResourceType != "aws_instance" {
		t.Errorf("resource type = %q, want %q", receivedResourceType, "aws_instance")
	}
	if len(resp.Content.Blocks) != 1 {
		t.Errorf("expected 1 block, got %d", len(resp.Content.Blocks))
	}
}

func TestGRPCRunnerServer_GetNewResourceContent(t *testing.T) {
	var receivedResourceType string
	expectedContent := &hclext.BodyContent{
		Blocks: []*hclext.Block{
			{Type: "resource", Labels: []string{"azurerm_storage_account", "main"}},
		},
	}

	runner := &recordingRunner{
		onGetNewResourceContent: func(resourceType string, schema *hclext.BodySchema, opts *tflint.GetModuleContentOption) (*hclext.BodyContent, error) {
			receivedResourceType = resourceType
			return expectedContent, nil
		},
	}
	server := &GRPCRunnerServer{impl: runner}

	resp, err := server.GetNewResourceContent(nil, &pb.GetResourceContent_Request{
		ResourceType: "azurerm_storage_account",
		Schema:       &pb.BodySchema{},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if receivedResourceType != "azurerm_storage_account" {
		t.Errorf("resource type = %q, want %q", receivedResourceType, "azurerm_storage_account")
	}
	if len(resp.Content.Blocks) != 1 {
		t.Errorf("expected 1 block, got %d", len(resp.Content.Blocks))
	}
}

func TestGRPCRunnerServer_EmitIssue(t *testing.T) {
	var capturedRule tflint.Rule
	var capturedMessage string
	var capturedRange hcl.Range

	runner := &recordingRunner{
		onEmitIssue: func(rule tflint.Rule, message string, issueRange hcl.Range) error {
			capturedRule = rule
			capturedMessage = message
			capturedRange = issueRange
			return nil
		},
	}
	server := &GRPCRunnerServer{impl: runner}

	resp, err := server.EmitIssue(nil, &pb.EmitIssue_Request{
		Rule: &pb.Rule{
			Name:     "test_rule",
			Enabled:  true,
			Severity: pb.Severity_SEVERITY_WARNING,
			Link:     "https://example.com/docs",
		},
		Message: "Something went wrong",
		Range: &pb.Range{
			Filename: "main.tf",
			Start:    &pb.Position{Line: 10, Column: 5, Byte: 100},
			End:      &pb.Position{Line: 10, Column: 20, Byte: 115},
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp == nil {
		t.Fatal("expected response, got nil")
	}
	if capturedRule.Name() != "test_rule" {
		t.Errorf("rule name = %q, want %q", capturedRule.Name(), "test_rule")
	}
	if capturedRule.Severity() != tflint.WARNING {
		t.Errorf("rule severity = %v, want WARNING", capturedRule.Severity())
	}
	if capturedMessage != "Something went wrong" {
		t.Errorf("message = %q, want %q", capturedMessage, "Something went wrong")
	}
	if capturedRange.Filename != "main.tf" {
		t.Errorf("range filename = %q, want %q", capturedRange.Filename, "main.tf")
	}
	if capturedRange.Start.Line != 10 {
		t.Errorf("range start line = %d, want 10", capturedRange.Start.Line)
	}
}

func TestGRPCRunnerServer_DecodeRuleConfig_NoConfig(t *testing.T) {
	runner := &recordingRunner{
		onDecodeRuleConfig: func(ruleName string, target any) error {
			// Don't modify target - simulates no config found
			return nil
		},
	}
	server := &GRPCRunnerServer{impl: runner}

	resp, err := server.DecodeRuleConfig(nil, &pb.DecodeRuleConfig_Request{
		RuleName: "test_rule",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.HasConfig {
		t.Error("expected HasConfig=false for no config")
	}
	if len(resp.ConfigBytes) != 0 {
		t.Errorf("expected empty ConfigBytes, got %d bytes", len(resp.ConfigBytes))
	}
}

func TestGRPCRunnerServer_DecodeRuleConfig_WithConfig(t *testing.T) {
	runner := &recordingRunner{
		onDecodeRuleConfig: func(ruleName string, target any) error {
			// Set target to a config map
			if m, ok := target.(*map[string]interface{}); ok {
				*m = map[string]interface{}{
					"enabled": true,
					"level":   "strict",
				}
			}
			return nil
		},
	}
	server := &GRPCRunnerServer{impl: runner}

	resp, err := server.DecodeRuleConfig(nil, &pb.DecodeRuleConfig_Request{
		RuleName: "test_rule",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !resp.HasConfig {
		t.Error("expected HasConfig=true")
	}
	if len(resp.ConfigBytes) == 0 {
		t.Error("expected non-empty ConfigBytes")
	}
}

func TestGRPCRunnerServer_MethodsReturnError(t *testing.T) {
	expectedErr := fmt.Errorf("test error")

	t.Run("GetOldModuleContent error", func(t *testing.T) {
		runner := &recordingRunner{
			onGetOldModuleContent: func(schema *hclext.BodySchema, opts *tflint.GetModuleContentOption) (*hclext.BodyContent, error) {
				return nil, expectedErr
			},
		}
		server := &GRPCRunnerServer{impl: runner}
		_, err := server.GetOldModuleContent(nil, &pb.GetModuleContent_Request{})
		if err == nil {
			t.Error("expected error, got nil")
		}
	})

	t.Run("GetNewModuleContent error", func(t *testing.T) {
		runner := &recordingRunner{
			onGetNewModuleContent: func(schema *hclext.BodySchema, opts *tflint.GetModuleContentOption) (*hclext.BodyContent, error) {
				return nil, expectedErr
			},
		}
		server := &GRPCRunnerServer{impl: runner}
		_, err := server.GetNewModuleContent(nil, &pb.GetModuleContent_Request{})
		if err == nil {
			t.Error("expected error, got nil")
		}
	})

	t.Run("GetOldResourceContent error", func(t *testing.T) {
		runner := &recordingRunner{
			onGetOldResourceContent: func(resourceType string, schema *hclext.BodySchema, opts *tflint.GetModuleContentOption) (*hclext.BodyContent, error) {
				return nil, expectedErr
			},
		}
		server := &GRPCRunnerServer{impl: runner}
		_, err := server.GetOldResourceContent(nil, &pb.GetResourceContent_Request{})
		if err == nil {
			t.Error("expected error, got nil")
		}
	})

	t.Run("GetNewResourceContent error", func(t *testing.T) {
		runner := &recordingRunner{
			onGetNewResourceContent: func(resourceType string, schema *hclext.BodySchema, opts *tflint.GetModuleContentOption) (*hclext.BodyContent, error) {
				return nil, expectedErr
			},
		}
		server := &GRPCRunnerServer{impl: runner}
		_, err := server.GetNewResourceContent(nil, &pb.GetResourceContent_Request{})
		if err == nil {
			t.Error("expected error, got nil")
		}
	})

	t.Run("EmitIssue error", func(t *testing.T) {
		runner := &recordingRunner{
			onEmitIssue: func(rule tflint.Rule, message string, issueRange hcl.Range) error {
				return expectedErr
			},
		}
		server := &GRPCRunnerServer{impl: runner}
		_, err := server.EmitIssue(nil, &pb.EmitIssue_Request{
			Rule: &pb.Rule{Name: "test"},
		})
		if err == nil {
			t.Error("expected error, got nil")
		}
	})

	t.Run("DecodeRuleConfig error", func(t *testing.T) {
		runner := &recordingRunner{
			onDecodeRuleConfig: func(ruleName string, target any) error {
				return expectedErr
			},
		}
		server := &GRPCRunnerServer{impl: runner}
		_, err := server.DecodeRuleConfig(nil, &pb.DecodeRuleConfig_Request{RuleName: "test"})
		if err == nil {
			t.Error("expected error, got nil")
		}
	})
}
