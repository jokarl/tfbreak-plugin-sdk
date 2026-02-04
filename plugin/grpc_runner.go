// Package plugin provides gRPC-based plugin communication for tfbreak.
//
// This file implements the gRPC Runner service for bidirectional communication.
// The host (tfbreak-core) implements the Runner server, and plugins use the
// Runner client to call back to the host for configuration access.

package plugin

import (
	"context"
	"encoding/json"
	"time"

	"github.com/hashicorp/hcl/v2"

	"github.com/jokarl/tfbreak-plugin-sdk/hclext"
	pb "github.com/jokarl/tfbreak-plugin-sdk/plugin/proto"
	"github.com/jokarl/tfbreak-plugin-sdk/tflint"
)

// runnerCallTimeout is the timeout for individual runner callback calls.
// These should be fast since they're just data retrieval from the host.
const runnerCallTimeout = 30 * time.Second

// =============================================================================
// GRPCRunnerClient - Plugin side (calls back to host)
// =============================================================================

// GRPCRunnerClient implements tflint.Runner by calling back to the host.
// This runs in the plugin process and makes gRPC calls to the host's Runner server.
type GRPCRunnerClient struct {
	client pb.RunnerClient
}

// Ensure GRPCRunnerClient implements tflint.Runner.
var _ tflint.Runner = (*GRPCRunnerClient)(nil)

// GetOldModuleContent retrieves module content from the OLD (baseline) configuration.
func (r *GRPCRunnerClient) GetOldModuleContent(schema *hclext.BodySchema, opts *tflint.GetModuleContentOption) (*hclext.BodyContent, error) {
	ctx, cancel := context.WithTimeout(context.Background(), runnerCallTimeout)
	defer cancel()

	resp, err := r.client.GetOldModuleContent(ctx, &pb.GetModuleContent_Request{
		Schema: toProtoBodySchema(schema),
		Option: toProtoGetModuleContentOption(opts),
	})
	if err != nil {
		return nil, err
	}
	return fromProtoBodyContent(resp.GetContent()), nil
}

// GetNewModuleContent retrieves module content from the NEW configuration.
func (r *GRPCRunnerClient) GetNewModuleContent(schema *hclext.BodySchema, opts *tflint.GetModuleContentOption) (*hclext.BodyContent, error) {
	ctx, cancel := context.WithTimeout(context.Background(), runnerCallTimeout)
	defer cancel()

	resp, err := r.client.GetNewModuleContent(ctx, &pb.GetModuleContent_Request{
		Schema: toProtoBodySchema(schema),
		Option: toProtoGetModuleContentOption(opts),
	})
	if err != nil {
		return nil, err
	}
	return fromProtoBodyContent(resp.GetContent()), nil
}

// GetOldResourceContent retrieves resources of a specific type from the OLD configuration.
func (r *GRPCRunnerClient) GetOldResourceContent(resourceType string, schema *hclext.BodySchema, opts *tflint.GetModuleContentOption) (*hclext.BodyContent, error) {
	ctx, cancel := context.WithTimeout(context.Background(), runnerCallTimeout)
	defer cancel()

	resp, err := r.client.GetOldResourceContent(ctx, &pb.GetResourceContent_Request{
		ResourceType: resourceType,
		Schema:       toProtoBodySchema(schema),
		Option:       toProtoGetModuleContentOption(opts),
	})
	if err != nil {
		return nil, err
	}
	return fromProtoBodyContent(resp.GetContent()), nil
}

// GetNewResourceContent retrieves resources of a specific type from the NEW configuration.
func (r *GRPCRunnerClient) GetNewResourceContent(resourceType string, schema *hclext.BodySchema, opts *tflint.GetModuleContentOption) (*hclext.BodyContent, error) {
	ctx, cancel := context.WithTimeout(context.Background(), runnerCallTimeout)
	defer cancel()

	resp, err := r.client.GetNewResourceContent(ctx, &pb.GetResourceContent_Request{
		ResourceType: resourceType,
		Schema:       toProtoBodySchema(schema),
		Option:       toProtoGetModuleContentOption(opts),
	})
	if err != nil {
		return nil, err
	}
	return fromProtoBodyContent(resp.GetContent()), nil
}

// EmitIssue reports a finding from the rule.
func (r *GRPCRunnerClient) EmitIssue(rule tflint.Rule, message string, issueRange hcl.Range) error {
	ctx, cancel := context.WithTimeout(context.Background(), runnerCallTimeout)
	defer cancel()

	_, err := r.client.EmitIssue(ctx, &pb.EmitIssue_Request{
		Rule:    toProtoRule(rule),
		Message: message,
		Range:   toProtoRange(issueRange),
	})
	return err
}

// DecodeRuleConfig retrieves and decodes the rule's configuration.
func (r *GRPCRunnerClient) DecodeRuleConfig(ruleName string, target any) error {
	ctx, cancel := context.WithTimeout(context.Background(), runnerCallTimeout)
	defer cancel()

	resp, err := r.client.DecodeRuleConfig(ctx, &pb.DecodeRuleConfig_Request{
		RuleName: ruleName,
	})
	if err != nil {
		return err
	}

	// If no config was provided, return nil
	if !resp.GetHasConfig() || len(resp.GetConfigBytes()) == 0 {
		return nil
	}

	// Decode the JSON-encoded config into the target
	return json.Unmarshal(resp.GetConfigBytes(), target)
}

// =============================================================================
// GRPCRunnerServer - Host side (implements proto.RunnerServer)
// =============================================================================

// GRPCRunnerServer wraps a tflint.Runner to implement the gRPC server.
// This runs in the host process and handles requests from the plugin.
type GRPCRunnerServer struct {
	pb.UnimplementedRunnerServer
	impl tflint.Runner
}

// GetOldModuleContent handles the gRPC call for old module content.
func (s *GRPCRunnerServer) GetOldModuleContent(ctx context.Context, req *pb.GetModuleContent_Request) (*pb.GetModuleContent_Response, error) {
	content, err := s.impl.GetOldModuleContent(
		fromProtoBodySchema(req.GetSchema()),
		fromProtoGetModuleContentOption(req.GetOption()),
	)
	if err != nil {
		return nil, err
	}
	return &pb.GetModuleContent_Response{
		Content: toProtoBodyContent(content),
	}, nil
}

// GetNewModuleContent handles the gRPC call for new module content.
func (s *GRPCRunnerServer) GetNewModuleContent(ctx context.Context, req *pb.GetModuleContent_Request) (*pb.GetModuleContent_Response, error) {
	content, err := s.impl.GetNewModuleContent(
		fromProtoBodySchema(req.GetSchema()),
		fromProtoGetModuleContentOption(req.GetOption()),
	)
	if err != nil {
		return nil, err
	}
	return &pb.GetModuleContent_Response{
		Content: toProtoBodyContent(content),
	}, nil
}

// GetOldResourceContent handles the gRPC call for old resource content.
func (s *GRPCRunnerServer) GetOldResourceContent(ctx context.Context, req *pb.GetResourceContent_Request) (*pb.GetResourceContent_Response, error) {
	content, err := s.impl.GetOldResourceContent(
		req.GetResourceType(),
		fromProtoBodySchema(req.GetSchema()),
		fromProtoGetModuleContentOption(req.GetOption()),
	)
	if err != nil {
		return nil, err
	}
	return &pb.GetResourceContent_Response{
		Content: toProtoBodyContent(content),
	}, nil
}

// GetNewResourceContent handles the gRPC call for new resource content.
func (s *GRPCRunnerServer) GetNewResourceContent(ctx context.Context, req *pb.GetResourceContent_Request) (*pb.GetResourceContent_Response, error) {
	content, err := s.impl.GetNewResourceContent(
		req.GetResourceType(),
		fromProtoBodySchema(req.GetSchema()),
		fromProtoGetModuleContentOption(req.GetOption()),
	)
	if err != nil {
		return nil, err
	}
	return &pb.GetResourceContent_Response{
		Content: toProtoBodyContent(content),
	}, nil
}

// EmitIssue handles the gRPC call to emit an issue.
func (s *GRPCRunnerServer) EmitIssue(ctx context.Context, req *pb.EmitIssue_Request) (*pb.EmitIssue_Response, error) {
	// Create a minimal rule implementation for the callback
	rule := &protoRule{
		name:     req.GetRule().GetName(),
		enabled:  req.GetRule().GetEnabled(),
		severity: fromProtoSeverity(req.GetRule().GetSeverity()),
		link:     req.GetRule().GetLink(),
	}

	err := s.impl.EmitIssue(rule, req.GetMessage(), fromProtoRange(req.GetRange()))
	if err != nil {
		return nil, err
	}
	return &pb.EmitIssue_Response{}, nil
}

// DecodeRuleConfig handles the gRPC call to decode rule configuration.
func (s *GRPCRunnerServer) DecodeRuleConfig(ctx context.Context, req *pb.DecodeRuleConfig_Request) (*pb.DecodeRuleConfig_Response, error) {
	// Create a temporary target to capture the config
	var configMap map[string]interface{}
	err := s.impl.DecodeRuleConfig(req.GetRuleName(), &configMap)
	if err != nil {
		return nil, err
	}

	// If no config was found, return empty response
	if configMap == nil {
		return &pb.DecodeRuleConfig_Response{
			HasConfig:   false,
			ConfigBytes: nil,
		}, nil
	}

	// Encode the config as JSON
	configBytes, err := json.Marshal(configMap)
	if err != nil {
		return nil, err
	}

	return &pb.DecodeRuleConfig_Response{
		HasConfig:   true,
		ConfigBytes: configBytes,
	}, nil
}

// protoRule is a minimal Rule implementation used for EmitIssue callbacks.
type protoRule struct {
	name     string
	enabled  bool
	severity tflint.Severity
	link     string
}

func (r *protoRule) Name() string          { return r.name }
func (r *protoRule) Enabled() bool         { return r.enabled }
func (r *protoRule) Severity() tflint.Severity { return r.severity }
func (r *protoRule) Link() string          { return r.link }
func (r *protoRule) Check(tflint.Runner) error { return nil }
