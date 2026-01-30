package tflint

import "testing"

func TestDefaultRule_Enabled(t *testing.T) {
	rule := DefaultRule{}

	if got := rule.Enabled(); got != true {
		t.Errorf("DefaultRule.Enabled() = %v, want true", got)
	}
}

func TestDefaultRule_Severity(t *testing.T) {
	rule := DefaultRule{}

	if got := rule.Severity(); got != ERROR {
		t.Errorf("DefaultRule.Severity() = %v, want ERROR", got)
	}
}

// TestDefaultRule_Embedding verifies that DefaultRule can be embedded
// in a custom rule struct and provides the expected defaults.
func TestDefaultRule_Embedding(t *testing.T) {
	type CustomRule struct {
		DefaultRule
	}

	rule := CustomRule{}

	t.Run("embedded Enabled", func(t *testing.T) {
		if got := rule.Enabled(); got != true {
			t.Errorf("CustomRule.Enabled() = %v, want true", got)
		}
	})

	t.Run("embedded Severity", func(t *testing.T) {
		if got := rule.Severity(); got != ERROR {
			t.Errorf("CustomRule.Severity() = %v, want ERROR", got)
		}
	})
}
