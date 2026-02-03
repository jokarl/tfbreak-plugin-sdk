// Package plugin provides the entry point for tfbreak plugins.
//
// Plugins use this package to register their RuleSet with tfbreak-core.
// The Serve function is called from main() and handles all communication
// with the tfbreak host process using gRPC via HashiCorp's go-plugin library.
//
// Example plugin main.go:
//
//	package main
//
//	import (
//	    "github.com/jokarl/tfbreak-plugin-sdk/plugin"
//	    "github.com/jokarl/tfbreak-plugin-sdk/tflint"
//	)
//
//	func main() {
//	    plugin.Serve(&plugin.ServeOpts{
//	        RuleSet: &AzurermRuleSet{
//	            BuiltinRuleSet: tflint.BuiltinRuleSet{
//	                Name:    "azurerm",
//	                Version: "0.1.0",
//	                Rules:   rules.Rules,
//	            },
//	        },
//	    })
//	}
package plugin

import (
	"os"

	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/go-plugin"

	"github.com/jokarl/tfbreak-plugin-sdk/tflint"
)

// ServeOpts contains options for serving the plugin.
type ServeOpts struct {
	// RuleSet is the plugin's rule set implementation.
	RuleSet tflint.RuleSet
}

// Serve starts the plugin server.
//
// This function registers the plugin's RuleSet and handles communication
// with the tfbreak host process. It should be called from the plugin's
// main() function.
//
// The function blocks until the host disconnects. When invoked directly
// (outside of tfbreak), the plugin will print a message and exit.
//
// Communication uses gRPC with HashiCorp's go-plugin library, which provides:
// - Magic cookie handshake to prevent direct execution
// - Protocol versioning for compatibility
// - Bidirectional gRPC for Runner callbacks
//
// Example:
//
//	func main() {
//	    plugin.Serve(&plugin.ServeOpts{
//	        RuleSet: &MyRuleSet{...},
//	    })
//	}
func Serve(opts *ServeOpts) {
	if opts == nil || opts.RuleSet == nil {
		// Nothing to serve
		return
	}

	// Validate the RuleSet is usable (fail fast on misconfiguration)
	_ = opts.RuleSet.RuleSetName()
	_ = opts.RuleSet.RuleSetVersion()
	_ = opts.RuleSet.RuleNames()

	// Check if we're being invoked by tfbreak (via magic cookie)
	// If not, print a helpful message and exit
	if os.Getenv(MagicCookieKey) != MagicCookieValue {
		printDirectInvocationMessage(opts.RuleSet)
		return
	}

	// Create a logger for the plugin
	logger := hclog.New(&hclog.LoggerOptions{
		Name:   "plugin",
		Level:  hclog.Warn,
		Output: os.Stderr,
	})

	// Create the plugin map with our implementation
	pluginMap := map[string]plugin.Plugin{
		PluginName: &RuleSetPlugin{Impl: opts.RuleSet},
	}

	// Serve the plugin
	plugin.Serve(&plugin.ServeConfig{
		HandshakeConfig: Handshake,
		Plugins:         pluginMap,
		GRPCServer:      plugin.DefaultGRPCServer,
		Logger:          logger,
	})
}

// printDirectInvocationMessage prints a helpful message when the plugin
// is invoked directly instead of via tfbreak.
func printDirectInvocationMessage(rs tflint.RuleSet) {
	// Use simple printf since we don't want to pull in extra dependencies
	os.Stderr.WriteString("This is a tfbreak plugin.\n\n")
	os.Stderr.WriteString("Plugin: " + rs.RuleSetName() + "\n")
	os.Stderr.WriteString("Version: " + rs.RuleSetVersion() + "\n")
	os.Stderr.WriteString("Rules:\n")
	for _, name := range rs.RuleNames() {
		os.Stderr.WriteString("  - " + name + "\n")
	}
	os.Stderr.WriteString("\nTo use this plugin, run it via tfbreak:\n")
	os.Stderr.WriteString("  tfbreak [options]\n\n")
	os.Stderr.WriteString("For more information, see: https://github.com/jokarl/tfbreak\n")
}
