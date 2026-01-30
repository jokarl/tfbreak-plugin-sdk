package tflint

import (
	"testing"

	"github.com/hashicorp/hcl/v2"
)

// mockRule is a test implementation of the Rule interface.
type mockRule struct {
	DefaultRule
	name string
	link string
}

func (r *mockRule) Name() string { return r.name }
func (r *mockRule) Link() string { return r.link }
func (r *mockRule) Check(_ Runner) error {
	return nil
}

// TestRule_Interface verifies that a struct with DefaultRule embedded
// can satisfy the Rule interface.
func TestRule_Interface(t *testing.T) {
	rule := &mockRule{
		name: "test_rule",
		link: "https://example.com/test_rule",
	}

	// This is a compile-time check that mockRule satisfies Rule
	var _ Rule = rule

	// Verify the methods work as expected
	t.Run("Name", func(t *testing.T) {
		if got := rule.Name(); got != "test_rule" {
			t.Errorf("Name() = %q, want %q", got, "test_rule")
		}
	})

	t.Run("Enabled from DefaultRule", func(t *testing.T) {
		if got := rule.Enabled(); got != true {
			t.Errorf("Enabled() = %v, want true", got)
		}
	})

	t.Run("Severity from DefaultRule", func(t *testing.T) {
		if got := rule.Severity(); got != ERROR {
			t.Errorf("Severity() = %v, want ERROR", got)
		}
	})

	t.Run("Link", func(t *testing.T) {
		if got := rule.Link(); got != "https://example.com/test_rule" {
			t.Errorf("Link() = %q, want %q", got, "https://example.com/test_rule")
		}
	})

	t.Run("Check returns nil", func(t *testing.T) {
		if err := rule.Check(nil); err != nil {
			t.Errorf("Check() = %v, want nil", err)
		}
	})
}

// mockRuleOverrideSeverity demonstrates overriding DefaultRule's Severity.
type mockRuleOverrideSeverity struct {
	DefaultRule
}

func (r *mockRuleOverrideSeverity) Name() string     { return "warning_rule" }
func (r *mockRuleOverrideSeverity) Link() string     { return "" }
func (r *mockRuleOverrideSeverity) Check(_ Runner) error { return nil }
func (r *mockRuleOverrideSeverity) Severity() Severity   { return WARNING }

func TestRule_OverrideSeverity(t *testing.T) {
	rule := &mockRuleOverrideSeverity{}

	// Verify DefaultRule's Enabled is still used
	if got := rule.Enabled(); got != true {
		t.Errorf("Enabled() = %v, want true", got)
	}

	// Verify overridden Severity
	if got := rule.Severity(); got != WARNING {
		t.Errorf("Severity() = %v, want WARNING", got)
	}
}

// mockRunner is a minimal Runner implementation for testing interface compilation.
type mockRunner struct {
	issues []struct {
		Rule    Rule
		Message string
		Range   hcl.Range
	}
}

func (r *mockRunner) GetOldModuleContent(_ *struct{ Attributes []struct{ Name string; Required bool }; Blocks []struct{ Type string; LabelNames []string; Body interface{} }; Mode int }, _ *GetModuleContentOption) (*struct{ Attributes map[string]*struct{ Name string; Expr interface{}; Range struct{ Filename string; Start, End struct{ Line, Column, Byte int } }; NameRange struct{ Filename string; Start, End struct{ Line, Column, Byte int } } }; Blocks []*struct{ Type string; Labels []string; Body interface{}; DefRange, TypeRange struct{ Filename string; Start, End struct{ Line, Column, Byte int } }; LabelRanges []struct{ Filename string; Start, End struct{ Line, Column, Byte int } } } }, error) {
	return nil, nil
}

// Verify mockRunner satisfies Runner at compile time
var _ interface {
	EmitIssue(rule Rule, message string, issueRange hcl.Range) error
} = (*mockRunner)(nil)

func (r *mockRunner) EmitIssue(rule Rule, message string, issueRange hcl.Range) error {
	r.issues = append(r.issues, struct {
		Rule    Rule
		Message string
		Range   hcl.Range
	}{rule, message, issueRange})
	return nil
}
