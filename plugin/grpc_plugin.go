// Package plugin provides gRPC-based plugin communication for tfbreak.
//
// This file implements the go-plugin GRPCPlugin interface, which bridges
// the native Go interfaces (tflint.RuleSet, tflint.Runner) with gRPC.

package plugin

import (
	"context"

	"github.com/hashicorp/go-plugin"
	"google.golang.org/grpc"

	"github.com/jokarl/tfbreak-plugin-sdk/hclext"
	pb "github.com/jokarl/tfbreak-plugin-sdk/plugin/proto"
	"github.com/jokarl/tfbreak-plugin-sdk/tflint"
)

// Ensure RuleSetPlugin implements plugin.GRPCPlugin.
var _ plugin.GRPCPlugin = (*RuleSetPlugin)(nil)

// RuleSetPlugin is the implementation of plugin.GRPCPlugin for the RuleSet service.
// This is used by both the host (to create a client) and the plugin (to create a server).
type RuleSetPlugin struct {
	plugin.Plugin
	// Impl is the concrete implementation of the RuleSet interface.
	// Only used when serving (plugin side).
	Impl tflint.RuleSet
}

// GRPCServer is called by the plugin to register the gRPC server.
// This is called on the plugin side.
func (p *RuleSetPlugin) GRPCServer(broker *plugin.GRPCBroker, s *grpc.Server) error {
	pb.RegisterRuleSetServer(s, &GRPCRuleSetServer{
		impl:   p.Impl,
		broker: broker,
	})
	return nil
}

// GRPCClient is called by the host to create a gRPC client.
// This is called on the host side (tfbreak-core).
func (p *RuleSetPlugin) GRPCClient(ctx context.Context, broker *plugin.GRPCBroker, c *grpc.ClientConn) (interface{}, error) {
	return &GRPCRuleSetClient{
		client: pb.NewRuleSetClient(c),
		broker: broker,
	}, nil
}

// =============================================================================
// GRPCRuleSetServer - Plugin side (implements proto.RuleSetServer)
// =============================================================================

// GRPCRuleSetServer wraps a tflint.RuleSet to implement the gRPC server.
// This runs in the plugin process and handles requests from the host.
type GRPCRuleSetServer struct {
	pb.UnimplementedRuleSetServer
	impl   tflint.RuleSet
	broker *plugin.GRPCBroker
}

// GetRuleSetName returns the name of the ruleset.
func (s *GRPCRuleSetServer) GetRuleSetName(ctx context.Context, req *pb.GetRuleSetName_Request) (*pb.GetRuleSetName_Response, error) {
	return &pb.GetRuleSetName_Response{
		Name: s.impl.RuleSetName(),
	}, nil
}

// GetRuleSetVersion returns the version of the ruleset.
func (s *GRPCRuleSetServer) GetRuleSetVersion(ctx context.Context, req *pb.GetRuleSetVersion_Request) (*pb.GetRuleSetVersion_Response, error) {
	return &pb.GetRuleSetVersion_Response{
		Version: s.impl.RuleSetVersion(),
	}, nil
}

// GetRuleNames returns the names of all rules in this ruleset.
func (s *GRPCRuleSetServer) GetRuleNames(ctx context.Context, req *pb.GetRuleNames_Request) (*pb.GetRuleNames_Response, error) {
	return &pb.GetRuleNames_Response{
		Names: s.impl.RuleNames(),
	}, nil
}

// GetVersionConstraint returns the tfbreak version constraint.
func (s *GRPCRuleSetServer) GetVersionConstraint(ctx context.Context, req *pb.GetVersionConstraint_Request) (*pb.GetVersionConstraint_Response, error) {
	return &pb.GetVersionConstraint_Response{
		Constraint: s.impl.VersionConstraint(),
	}, nil
}

// GetConfigSchema returns the schema for plugin-specific configuration.
func (s *GRPCRuleSetServer) GetConfigSchema(ctx context.Context, req *pb.GetConfigSchema_Request) (*pb.GetConfigSchema_Response, error) {
	schema := s.impl.ConfigSchema()
	return &pb.GetConfigSchema_Response{
		Schema: toProtoBodySchema(schema),
	}, nil
}

// ApplyGlobalConfig applies global tfbreak configuration.
func (s *GRPCRuleSetServer) ApplyGlobalConfig(ctx context.Context, req *pb.ApplyGlobalConfig_Request) (*pb.ApplyGlobalConfig_Response, error) {
	config := fromProtoConfig(req.GetConfig())
	if err := s.impl.ApplyGlobalConfig(config); err != nil {
		return nil, err
	}
	return &pb.ApplyGlobalConfig_Response{}, nil
}

// ApplyConfig applies plugin-specific configuration.
func (s *GRPCRuleSetServer) ApplyConfig(ctx context.Context, req *pb.ApplyConfig_Request) (*pb.ApplyConfig_Response, error) {
	content := fromProtoBodyContent(req.GetContent())
	if err := s.impl.ApplyConfig(content); err != nil {
		return nil, err
	}
	return &pb.ApplyConfig_Response{}, nil
}

// Check executes all enabled rules.
func (s *GRPCRuleSetServer) Check(ctx context.Context, req *pb.Check_Request) (*pb.Check_Response, error) {
	// The broker provides a unique ID for this call.
	// The host starts a Runner server and tells us the ID.
	// We connect back to the host's Runner server.
	//
	// For simplicity in this implementation, we use go-plugin's built-in
	// bidirectional communication. The Runner client will be passed via
	// the context or a separate broker connection.

	// Get the runner connection from the broker.
	// The host should have started a Runner server for us.
	conn, err := s.broker.Dial(RunnerBrokerID)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	runnerClient := pb.NewRunnerClient(conn)
	runner := &GRPCRunnerClient{client: runnerClient}

	// Let the ruleset optionally wrap the runner
	wrappedRunner, err := s.impl.NewRunner(runner)
	if err != nil {
		return nil, err
	}

	// Execute all enabled rules
	builtin := s.impl.BuiltinImpl()
	for _, rule := range builtin.EnabledRules() {
		if err := rule.Check(wrappedRunner); err != nil {
			return nil, err
		}
	}

	return &pb.Check_Response{}, nil
}

// RunnerBrokerID is the broker ID used for the Runner callback service.
// The host starts a server with this ID, and the plugin connects to it.
const RunnerBrokerID uint32 = 1

// =============================================================================
// GRPCRuleSetClient - Host side (implements tflint.RuleSet)
// =============================================================================

// GRPCRuleSetClient wraps the gRPC client to implement tflint.RuleSet.
// This runs in the host process (tfbreak-core) and calls the plugin.
type GRPCRuleSetClient struct {
	client pb.RuleSetClient
	broker *plugin.GRPCBroker
}

// RuleSetName returns the name of the ruleset.
func (c *GRPCRuleSetClient) RuleSetName() string {
	resp, err := c.client.GetRuleSetName(context.Background(), &pb.GetRuleSetName_Request{})
	if err != nil {
		return ""
	}
	return resp.GetName()
}

// RuleSetVersion returns the version of the ruleset.
func (c *GRPCRuleSetClient) RuleSetVersion() string {
	resp, err := c.client.GetRuleSetVersion(context.Background(), &pb.GetRuleSetVersion_Request{})
	if err != nil {
		return ""
	}
	return resp.GetVersion()
}

// RuleNames returns the names of all rules in this ruleset.
func (c *GRPCRuleSetClient) RuleNames() []string {
	resp, err := c.client.GetRuleNames(context.Background(), &pb.GetRuleNames_Request{})
	if err != nil {
		return nil
	}
	return resp.GetNames()
}

// VersionConstraint returns the tfbreak version constraint.
func (c *GRPCRuleSetClient) VersionConstraint() string {
	resp, err := c.client.GetVersionConstraint(context.Background(), &pb.GetVersionConstraint_Request{})
	if err != nil {
		return ""
	}
	return resp.GetConstraint()
}

// ConfigSchema returns the schema for plugin-specific configuration.
func (c *GRPCRuleSetClient) ConfigSchema() *hclext.BodySchema {
	resp, err := c.client.GetConfigSchema(context.Background(), &pb.GetConfigSchema_Request{})
	if err != nil {
		return nil
	}
	return fromProtoBodySchema(resp.GetSchema())
}

// ApplyGlobalConfig applies global tfbreak configuration.
func (c *GRPCRuleSetClient) ApplyGlobalConfig(config *tflint.Config) error {
	_, err := c.client.ApplyGlobalConfig(context.Background(), &pb.ApplyGlobalConfig_Request{
		Config: toProtoConfig(config),
	})
	return err
}

// ApplyConfig applies plugin-specific configuration.
func (c *GRPCRuleSetClient) ApplyConfig(content *hclext.BodyContent) error {
	_, err := c.client.ApplyConfig(context.Background(), &pb.ApplyConfig_Request{
		Content: toProtoBodyContent(content),
	})
	return err
}

// NewRunner optionally wraps the runner with custom behavior.
// On the client side, this is a no-op since wrapping happens on the plugin side.
func (c *GRPCRuleSetClient) NewRunner(runner tflint.Runner) (tflint.Runner, error) {
	return runner, nil
}

// BuiltinImpl returns nil on the client side.
// The actual implementation is on the plugin side.
func (c *GRPCRuleSetClient) BuiltinImpl() *tflint.BuiltinRuleSet {
	return nil
}

// Check executes all enabled rules via the plugin.
// The host must provide a Runner implementation that the plugin can call back to.
func (c *GRPCRuleSetClient) Check(runner tflint.Runner) error {
	// Start a Runner server that the plugin can call back to
	runnerServer := &GRPCRunnerServer{impl: runner}

	// Use the broker to start a server the plugin can connect to
	serverFunc := func(opts []grpc.ServerOption) *grpc.Server {
		s := grpc.NewServer(opts...)
		pb.RegisterRunnerServer(s, runnerServer)
		return s
	}

	// Start the server in a goroutine since AcceptAndServe blocks
	go c.broker.AcceptAndServe(RunnerBrokerID, serverFunc)

	// Call the plugin's Check method
	_, err := c.client.Check(context.Background(), &pb.Check_Request{})
	return err
}
