package tflint

// DefaultRule provides default implementations for optional Rule interface methods.
// Plugin authors can embed this struct in their rule implementations to get
// sensible defaults for Enabled() and Severity().
//
// Example:
//
//	type MyRule struct {
//	    tflint.DefaultRule
//	}
//
//	func (r *MyRule) Name() string { return "my_rule" }
//	func (r *MyRule) Link() string { return "https://example.com/my_rule" }
//	func (r *MyRule) Check(runner Runner) error { ... }
//
// With DefaultRule embedded, MyRule automatically gets:
//   - Enabled() returning true (rules are enabled by default)
//   - Severity() returning ERROR (the default severity)
//
// Override these methods if your rule needs different defaults:
//
//	func (r *MyRule) Severity() tflint.Severity {
//	    return tflint.WARNING
//	}
type DefaultRule struct{}

// Enabled returns true, indicating rules are enabled by default.
// Override this method to disable a rule by default.
func (r DefaultRule) Enabled() bool {
	return true
}

// Severity returns ERROR, the default severity for rules.
// Override this method to specify a different default severity.
func (r DefaultRule) Severity() Severity {
	return ERROR
}
