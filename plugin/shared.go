// Package plugin provides the entry point for tfbreak plugins.
//
// This file contains shared configuration used by both the host (tfbreak-core)
// and plugins for establishing gRPC communication via hashicorp/go-plugin.

package plugin

import (
	"github.com/hashicorp/go-plugin"
)

// ProtocolVersion is the plugin protocol version.
// Increment this when making breaking changes to the plugin interface.
const ProtocolVersion = 1

// MagicCookieKey is the environment variable name for the magic cookie.
const MagicCookieKey = "TFBREAK_PLUGIN_MAGIC_COOKIE"

// MagicCookieValue is the expected value of the magic cookie.
// This prevents plugins from being executed directly (outside of tfbreak).
const MagicCookieValue = "tfbreak-plugin-v1"

// Handshake is the HandshakeConfig used to configure go-plugin.
// The host and plugin must agree on these values to communicate.
var Handshake = plugin.HandshakeConfig{
	ProtocolVersion:  ProtocolVersion,
	MagicCookieKey:   MagicCookieKey,
	MagicCookieValue: MagicCookieValue,
}

// PluginName is the name used to identify the RuleSet plugin.
const PluginName = "ruleset"

// PluginMap is the map of plugins we can dispense.
// Used by both the host and plugin.
var PluginMap = map[string]plugin.Plugin{
	PluginName: &RuleSetPlugin{},
}
