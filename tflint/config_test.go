package tflint

import "testing"

func TestConfig_Fields(t *testing.T) {
	config := &Config{
		Rules: map[string]*RuleConfig{
			"rule_a": {Name: "rule_a", Enabled: true},
		},
		DisabledByDefault: true,
		Only:              []string{"rule_a"},
		PluginDir:         "/custom/plugins",
	}

	if len(config.Rules) != 1 {
		t.Errorf("Rules has %d entries, want 1", len(config.Rules))
	}
	if !config.DisabledByDefault {
		t.Error("DisabledByDefault = false, want true")
	}
	if len(config.Only) != 1 || config.Only[0] != "rule_a" {
		t.Errorf("Only = %v, want [\"rule_a\"]", config.Only)
	}
	if config.PluginDir != "/custom/plugins" {
		t.Errorf("PluginDir = %q, want %q", config.PluginDir, "/custom/plugins")
	}
}

func TestRuleConfig_Fields(t *testing.T) {
	rc := &RuleConfig{
		Name:    "my_rule",
		Enabled: true,
		Body:    nil,
	}

	if rc.Name != "my_rule" {
		t.Errorf("Name = %q, want %q", rc.Name, "my_rule")
	}
	if !rc.Enabled {
		t.Error("Enabled = false, want true")
	}
}
