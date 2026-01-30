// Package plugin provides the entry point for tfbreak plugins.
//
// Plugins use this package to register their RuleSet with tfbreak-core.
// The Serve function is called from main() and handles all communication
// with the tfbreak host process.
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

import "github.com/jokarl/tfbreak-plugin-sdk/tflint"

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
// Current implementation (v0.1.0):
// This is a stub implementation that allows plugins to be built as
// standalone binaries. The plugin will compile and can be executed,
// but actual host communication is not yet implemented.
//
// Future implementation:
// Will use HashiCorp's go-plugin library with gRPC for communication
// between tfbreak-core and plugins. The API (ServeOpts, Serve) will
// remain stable.
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

	// TODO(gRPC): Implement actual plugin serving with go-plugin.
	//
	// Future implementation will:
	// 1. Set up gRPC server with RuleSet service
	// 2. Configure handshake with tfbreak-core
	// 3. Block until host disconnects
	//
	// For now, this stub allows plugins to be built and the main()
	// function to complete. When invoked directly (not via tfbreak),
	// the plugin simply exits.

	// Validate the RuleSet is usable (fail fast on misconfiguration)
	_ = opts.RuleSet.RuleSetName()
	_ = opts.RuleSet.RuleSetVersion()
	_ = opts.RuleSet.RuleNames()
}
