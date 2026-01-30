package tflint

import (
	"reflect"
	"testing"
)

// testRule is a minimal rule for testing.
type testRule struct {
	DefaultRule
	name    string
	enabled bool
}

func (r *testRule) Name() string     { return r.name }
func (r *testRule) Link() string     { return "" }
func (r *testRule) Check(_ Runner) error { return nil }
func (r *testRule) Enabled() bool    { return r.enabled }

func newTestRule(name string, enabled bool) *testRule {
	return &testRule{name: name, enabled: enabled}
}

func TestBuiltinRuleSet_RuleSetName(t *testing.T) {
	rs := &BuiltinRuleSet{Name: "test-plugin"}
	if got := rs.RuleSetName(); got != "test-plugin" {
		t.Errorf("RuleSetName() = %q, want %q", got, "test-plugin")
	}
}

func TestBuiltinRuleSet_RuleSetVersion(t *testing.T) {
	rs := &BuiltinRuleSet{Version: "1.2.3"}
	if got := rs.RuleSetVersion(); got != "1.2.3" {
		t.Errorf("RuleSetVersion() = %q, want %q", got, "1.2.3")
	}
}

func TestBuiltinRuleSet_RuleNames(t *testing.T) {
	rs := &BuiltinRuleSet{
		Rules: []Rule{
			newTestRule("rule_a", true),
			newTestRule("rule_b", true),
			newTestRule("rule_c", true),
		},
	}

	got := rs.RuleNames()
	want := []string{"rule_a", "rule_b", "rule_c"}

	if !reflect.DeepEqual(got, want) {
		t.Errorf("RuleNames() = %v, want %v", got, want)
	}
}

func TestBuiltinRuleSet_VersionConstraint_Default(t *testing.T) {
	rs := &BuiltinRuleSet{}
	if got := rs.VersionConstraint(); got != ">= 0.1.0" {
		t.Errorf("VersionConstraint() = %q, want %q", got, ">= 0.1.0")
	}
}

func TestBuiltinRuleSet_VersionConstraint_Custom(t *testing.T) {
	rs := &BuiltinRuleSet{Constraint: ">= 1.0.0"}
	if got := rs.VersionConstraint(); got != ">= 1.0.0" {
		t.Errorf("VersionConstraint() = %q, want %q", got, ">= 1.0.0")
	}
}

func TestBuiltinRuleSet_ConfigSchema_Default(t *testing.T) {
	rs := &BuiltinRuleSet{}
	if got := rs.ConfigSchema(); got != nil {
		t.Errorf("ConfigSchema() = %v, want nil", got)
	}
}

func TestBuiltinRuleSet_ApplyGlobalConfig_Nil(t *testing.T) {
	rs := &BuiltinRuleSet{
		Rules: []Rule{
			newTestRule("rule_a", true),
			newTestRule("rule_b", false),
		},
	}

	if err := rs.ApplyGlobalConfig(nil); err != nil {
		t.Fatalf("ApplyGlobalConfig(nil) = %v, want nil", err)
	}

	if !rs.IsRuleEnabled("rule_a") {
		t.Error("rule_a should be enabled (default true)")
	}
	if rs.IsRuleEnabled("rule_b") {
		t.Error("rule_b should be disabled (default false)")
	}
}

func TestBuiltinRuleSet_ApplyGlobalConfig_DisabledByDefault(t *testing.T) {
	rs := &BuiltinRuleSet{
		Rules: []Rule{
			newTestRule("rule_a", true),
			newTestRule("rule_b", true),
			newTestRule("rule_c", true),
		},
	}

	config := &Config{DisabledByDefault: true}
	if err := rs.ApplyGlobalConfig(config); err != nil {
		t.Fatalf("ApplyGlobalConfig() = %v, want nil", err)
	}

	for _, name := range []string{"rule_a", "rule_b", "rule_c"} {
		if rs.IsRuleEnabled(name) {
			t.Errorf("%s should be disabled", name)
		}
	}
}

func TestBuiltinRuleSet_ApplyGlobalConfig_Only(t *testing.T) {
	rs := &BuiltinRuleSet{
		Rules: []Rule{
			newTestRule("rule_a", true),
			newTestRule("rule_b", true),
			newTestRule("rule_c", true),
		},
	}

	config := &Config{Only: []string{"rule_a", "rule_c"}}
	if err := rs.ApplyGlobalConfig(config); err != nil {
		t.Fatalf("ApplyGlobalConfig() = %v, want nil", err)
	}

	if !rs.IsRuleEnabled("rule_a") {
		t.Error("rule_a should be enabled (in Only list)")
	}
	if rs.IsRuleEnabled("rule_b") {
		t.Error("rule_b should be disabled (not in Only list)")
	}
	if !rs.IsRuleEnabled("rule_c") {
		t.Error("rule_c should be enabled (in Only list)")
	}
}

func TestBuiltinRuleSet_ApplyGlobalConfig_RuleConfig(t *testing.T) {
	rs := &BuiltinRuleSet{
		Rules: []Rule{
			newTestRule("rule_a", true),
			newTestRule("rule_b", true),
		},
	}

	config := &Config{
		Rules: map[string]*RuleConfig{
			"rule_a": {Name: "rule_a", Enabled: false},
		},
	}
	if err := rs.ApplyGlobalConfig(config); err != nil {
		t.Fatalf("ApplyGlobalConfig() = %v, want nil", err)
	}

	if rs.IsRuleEnabled("rule_a") {
		t.Error("rule_a should be disabled (config set enabled=false)")
	}
	if !rs.IsRuleEnabled("rule_b") {
		t.Error("rule_b should be enabled (no config, default true)")
	}
}

func TestBuiltinRuleSet_IsRuleEnabled_BeforeConfig(t *testing.T) {
	rs := &BuiltinRuleSet{
		Rules: []Rule{
			newTestRule("rule_a", true),
			newTestRule("rule_b", false),
		},
	}

	// Before ApplyGlobalConfig, should use rule defaults
	if !rs.IsRuleEnabled("rule_a") {
		t.Error("rule_a should be enabled (default)")
	}
	if rs.IsRuleEnabled("rule_b") {
		t.Error("rule_b should be disabled (default)")
	}
	if rs.IsRuleEnabled("unknown") {
		t.Error("unknown rule should return false")
	}
}

func TestBuiltinRuleSet_GetRule(t *testing.T) {
	rule := newTestRule("my_rule", true)
	rs := &BuiltinRuleSet{
		Rules: []Rule{rule},
	}

	got := rs.GetRule("my_rule")
	if got != rule {
		t.Errorf("GetRule(\"my_rule\") = %v, want %v", got, rule)
	}
}

func TestBuiltinRuleSet_GetRule_NotFound(t *testing.T) {
	rs := &BuiltinRuleSet{
		Rules: []Rule{newTestRule("rule_a", true)},
	}

	if got := rs.GetRule("unknown"); got != nil {
		t.Errorf("GetRule(\"unknown\") = %v, want nil", got)
	}
}

func TestBuiltinRuleSet_EnabledRules(t *testing.T) {
	rs := &BuiltinRuleSet{
		Rules: []Rule{
			newTestRule("rule_a", true),
			newTestRule("rule_b", false),
			newTestRule("rule_c", true),
		},
	}

	_ = rs.ApplyGlobalConfig(nil)
	enabled := rs.EnabledRules()

	if len(enabled) != 2 {
		t.Fatalf("EnabledRules() returned %d rules, want 2", len(enabled))
	}

	names := make([]string, len(enabled))
	for i, r := range enabled {
		names[i] = r.Name()
	}

	if !contains(names, "rule_a") {
		t.Error("EnabledRules() should contain rule_a")
	}
	if contains(names, "rule_b") {
		t.Error("EnabledRules() should not contain rule_b")
	}
	if !contains(names, "rule_c") {
		t.Error("EnabledRules() should contain rule_c")
	}
}

func TestBuiltinRuleSet_NewRunner(t *testing.T) {
	rs := &BuiltinRuleSet{}

	// NewRunner should return the input runner unchanged
	got, err := rs.NewRunner(nil)
	if err != nil {
		t.Fatalf("NewRunner() error = %v", err)
	}
	if got != nil {
		t.Errorf("NewRunner(nil) = %v, want nil", got)
	}
}

func TestBuiltinRuleSet_BuiltinImpl(t *testing.T) {
	rs := &BuiltinRuleSet{Name: "test"}
	if got := rs.BuiltinImpl(); got != rs {
		t.Errorf("BuiltinImpl() = %v, want %v", got, rs)
	}
}

// TestBuiltinRuleSet_ImplementsRuleSet verifies BuiltinRuleSet satisfies RuleSet.
func TestBuiltinRuleSet_ImplementsRuleSet(t *testing.T) {
	var _ RuleSet = &BuiltinRuleSet{}
}

func contains(slice []string, s string) bool {
	for _, item := range slice {
		if item == s {
			return true
		}
	}
	return false
}
