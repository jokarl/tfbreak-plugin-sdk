// Package plugin provides gRPC-based plugin communication for tfbreak.
//
// This file implements the go-plugin GRPCPlugin interface, which bridges
// the native Go interfaces (tflint.RuleSet, tflint.Runner) with gRPC.

package plugin

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/hashicorp/go-plugin"
	"google.golang.org/grpc"

	"github.com/jokarl/tfbreak-plugin-sdk/hclext"
	pb "github.com/jokarl/tfbreak-plugin-sdk/plugin/proto"
	"github.com/jokarl/tfbreak-plugin-sdk/tflint"
)

// Default timeout for gRPC calls.
// This can be overridden per-call using context.WithTimeout.
const defaultGRPCTimeout = 30 * time.Second

// checkTimeout is the timeout for the Check method, which may take longer.
const checkTimeout = 5 * time.Minute

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
// All rules are executed even if some fail - errors are collected and returned together.
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

	// Execute all enabled rules, collecting errors rather than failing fast.
	// This ensures all rules run even if some fail, giving users a complete picture.
	builtin := s.impl.BuiltinImpl()
	var ruleErrors []error
	for _, rule := range builtin.EnabledRules() {
		// Check for context cancellation between rules
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		if err := rule.Check(wrappedRunner); err != nil {
			ruleErrors = append(ruleErrors, fmt.Errorf("rule %s: %w", rule.Name(), err))
		}
	}

	// If any rules failed, combine errors into a single error
	if len(ruleErrors) > 0 {
		return nil, combineErrors(ruleErrors)
	}

	return &pb.Check_Response{}, nil
}

// combineErrors combines multiple errors into a single error.
func combineErrors(errs []error) error {
	if len(errs) == 0 {
		return nil
	}
	if len(errs) == 1 {
		return errs[0]
	}

	var msg string
	for i, err := range errs {
		if i > 0 {
			msg += "; "
		}
		msg += err.Error()
	}
	return fmt.Errorf("%d rules failed: %s", len(errs), msg)
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
	ctx, cancel := context.WithTimeout(context.Background(), defaultGRPCTimeout)
	defer cancel()

	resp, err := c.client.GetRuleSetName(ctx, &pb.GetRuleSetName_Request{})
	if err != nil {
		// Return empty string on error - caller should handle missing name
		return ""
	}
	return resp.GetName()
}

// RuleSetVersion returns the version of the ruleset.
func (c *GRPCRuleSetClient) RuleSetVersion() string {
	ctx, cancel := context.WithTimeout(context.Background(), defaultGRPCTimeout)
	defer cancel()

	resp, err := c.client.GetRuleSetVersion(ctx, &pb.GetRuleSetVersion_Request{})
	if err != nil {
		// Return empty string on error - caller should handle missing version
		return ""
	}
	return resp.GetVersion()
}

// RuleNames returns the names of all rules in this ruleset.
func (c *GRPCRuleSetClient) RuleNames() []string {
	ctx, cancel := context.WithTimeout(context.Background(), defaultGRPCTimeout)
	defer cancel()

	resp, err := c.client.GetRuleNames(ctx, &pb.GetRuleNames_Request{})
	if err != nil {
		// Return nil on error - caller should handle missing rules
		return nil
	}
	return resp.GetNames()
}

// VersionConstraint returns the tfbreak version constraint.
func (c *GRPCRuleSetClient) VersionConstraint() string {
	ctx, cancel := context.WithTimeout(context.Background(), defaultGRPCTimeout)
	defer cancel()

	resp, err := c.client.GetVersionConstraint(ctx, &pb.GetVersionConstraint_Request{})
	if err != nil {
		// Return empty string on error - no constraint means any version
		return ""
	}
	return resp.GetConstraint()
}

// ConfigSchema returns the schema for plugin-specific configuration.
func (c *GRPCRuleSetClient) ConfigSchema() *hclext.BodySchema {
	ctx, cancel := context.WithTimeout(context.Background(), defaultGRPCTimeout)
	defer cancel()

	resp, err := c.client.GetConfigSchema(ctx, &pb.GetConfigSchema_Request{})
	if err != nil {
		// Return nil on error - no schema means no plugin config
		return nil
	}
	return fromProtoBodySchema(resp.GetSchema())
}

// ApplyGlobalConfig applies global tfbreak configuration.
func (c *GRPCRuleSetClient) ApplyGlobalConfig(config *tflint.Config) error {
	ctx, cancel := context.WithTimeout(context.Background(), defaultGRPCTimeout)
	defer cancel()

	_, err := c.client.ApplyGlobalConfig(ctx, &pb.ApplyGlobalConfig_Request{
		Config: toProtoConfig(config),
	})
	return err
}

// ApplyConfig applies plugin-specific configuration.
func (c *GRPCRuleSetClient) ApplyConfig(content *hclext.BodyContent) error {
	ctx, cancel := context.WithTimeout(context.Background(), defaultGRPCTimeout)
	defer cancel()

	_, err := c.client.ApplyConfig(ctx, &pb.ApplyConfig_Request{
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

	// Use a WaitGroup to ensure the server is ready before calling Check
	var serverReady sync.WaitGroup
	serverReady.Add(1)

	// Track the gRPC server for cleanup
	var grpcServer *grpc.Server
	var serverMu sync.Mutex

	// Use the broker to start a server the plugin can connect to
	serverFunc := func(opts []grpc.ServerOption) *grpc.Server {
		serverMu.Lock()
		grpcServer = grpc.NewServer(opts...)
		serverMu.Unlock()
		pb.RegisterRunnerServer(grpcServer, runnerServer)
		serverReady.Done()
		return grpcServer
	}

	// Start the server in a goroutine since AcceptAndServe blocks
	go c.broker.AcceptAndServe(RunnerBrokerID, serverFunc)

	// Wait for the server to be ready (with a reasonable timeout)
	serverReadyChan := make(chan struct{})
	go func() {
		serverReady.Wait()
		close(serverReadyChan)
	}()

	select {
	case <-serverReadyChan:
		// Server is ready
	case <-time.After(5 * time.Second):
		// Server startup timeout - proceed anyway, plugin may still connect
	}

	// Ensure cleanup of the gRPC server when done
	defer func() {
		serverMu.Lock()
		if grpcServer != nil {
			// Use GracefulStop to allow pending RPCs to complete
			grpcServer.GracefulStop()
		}
		serverMu.Unlock()
	}()

	// Call the plugin's Check method with a timeout
	ctx, cancel := context.WithTimeout(context.Background(), checkTimeout)
	defer cancel()

	_, err := c.client.Check(ctx, &pb.Check_Request{})
	return err
}
