package plugin

import (
	"testing"
)

func TestHandshakeConfig(t *testing.T) {
	if Handshake.ProtocolVersion != ProtocolVersion {
		t.Errorf("Handshake.ProtocolVersion = %d, want %d", Handshake.ProtocolVersion, ProtocolVersion)
	}
	if Handshake.MagicCookieKey != MagicCookieKey {
		t.Errorf("Handshake.MagicCookieKey = %q, want %q", Handshake.MagicCookieKey, MagicCookieKey)
	}
	if Handshake.MagicCookieValue != MagicCookieValue {
		t.Errorf("Handshake.MagicCookieValue = %q, want %q", Handshake.MagicCookieValue, MagicCookieValue)
	}
}

func TestConstants(t *testing.T) {
	if ProtocolVersion < 1 {
		t.Error("ProtocolVersion should be at least 1")
	}
	if MagicCookieKey == "" {
		t.Error("MagicCookieKey should not be empty")
	}
	if MagicCookieValue == "" {
		t.Error("MagicCookieValue should not be empty")
	}
	if PluginName == "" {
		t.Error("PluginName should not be empty")
	}
}

func TestPluginMap(t *testing.T) {
	if len(PluginMap) != 1 {
		t.Errorf("PluginMap should have 1 entry, got %d", len(PluginMap))
	}
	if _, ok := PluginMap[PluginName]; !ok {
		t.Errorf("PluginMap should contain %q", PluginName)
	}
}

func TestMagicCookieFormat(t *testing.T) {
	// Magic cookie key should follow environment variable naming conventions
	if MagicCookieKey != "TFBREAK_PLUGIN_MAGIC_COOKIE" {
		t.Errorf("MagicCookieKey = %q, expected TFBREAK_PLUGIN_MAGIC_COOKIE", MagicCookieKey)
	}

	// Magic cookie value should include version info
	if MagicCookieValue != "tfbreak-plugin-v1" {
		t.Errorf("MagicCookieValue = %q, expected tfbreak-plugin-v1", MagicCookieValue)
	}
}
