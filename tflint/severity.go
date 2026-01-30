// Package tflint provides tflint-aligned interfaces for tfbreak plugins.
//
// This package contains the core types, interfaces, and utilities needed
// to build tfbreak plugins. The naming and structure align with tflint-plugin-sdk
// for ecosystem familiarity, but with adaptations for tfbreak's dual-config
// comparison model.
//
// Key types:
//   - Severity: Issue severity levels (ERROR, WARNING, NOTICE)
//   - DefaultRule: Embeddable struct providing default Rule method implementations
//   - Rule: Interface that plugins implement for each detection rule
//   - Runner: Interface providing config access and issue emission (dual-config model)
//   - RuleSet: Interface for plugin registration and rule enumeration
//   - BuiltinRuleSet: Embeddable struct providing default RuleSet implementations
package tflint

// Severity represents the severity level of an issue.
// Values align with tflint-plugin-sdk for ecosystem familiarity.
type Severity int

const (
	// ERROR indicates a critical issue (e.g., breaking change causing recreation).
	ERROR Severity = iota + 1
	// WARNING indicates a potential issue that may need attention.
	WARNING
	// NOTICE indicates an informational finding.
	NOTICE
)

// String returns the string representation of the severity.
func (s Severity) String() string {
	switch s {
	case ERROR:
		return "ERROR"
	case WARNING:
		return "WARNING"
	case NOTICE:
		return "NOTICE"
	default:
		return "UNKNOWN"
	}
}
