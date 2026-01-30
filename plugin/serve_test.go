package plugin

import (
	"testing"

	"github.com/jokarl/tfbreak-plugin-sdk/tflint"
)

// testRule is a minimal rule for testing.
type testRule struct {
	tflint.DefaultRule
	name string
}

func (r *testRule) Name() string                   { return r.name }
func (r *testRule) Link() string                   { return "" }
func (r *testRule) Check(_ tflint.Runner) error { return nil }

func TestServe_NilOpts(t *testing.T) {
	// Should not panic with nil opts
	Serve(nil)
}

func TestServe_NilRuleSet(t *testing.T) {
	// Should not panic with nil RuleSet
	Serve(&ServeOpts{RuleSet: nil})
}

func TestServe_ValidRuleSet(t *testing.T) {
	rs := &tflint.BuiltinRuleSet{
		Name:    "test",
		Version: "1.0.0",
		Rules:   []tflint.Rule{&testRule{name: "test_rule"}},
	}

	// Should not panic with valid RuleSet
	Serve(&ServeOpts{RuleSet: rs})
}

func TestServe_RuleSetValidation(t *testing.T) {
	// Serve validates that RuleSet methods work
	rs := &tflint.BuiltinRuleSet{
		Name:    "validation-test",
		Version: "0.1.0",
		Rules: []tflint.Rule{
			&testRule{name: "rule1"},
			&testRule{name: "rule2"},
		},
	}

	// This should exercise the validation code path
	Serve(&ServeOpts{RuleSet: rs})

	// Verify the RuleSet is valid by checking its methods
	if rs.RuleSetName() != "validation-test" {
		t.Errorf("RuleSetName() = %q, want %q", rs.RuleSetName(), "validation-test")
	}
	if len(rs.RuleNames()) != 2 {
		t.Errorf("RuleNames() length = %d, want 2", len(rs.RuleNames()))
	}
}

func TestServeOpts_RuleSetField(t *testing.T) {
	rs := &tflint.BuiltinRuleSet{Name: "test"}
	opts := &ServeOpts{RuleSet: rs}

	if opts.RuleSet != rs {
		t.Error("ServeOpts.RuleSet should hold the provided RuleSet")
	}
}
