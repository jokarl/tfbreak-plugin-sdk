package helper

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/hashicorp/hcl/v2"
	"github.com/jokarl/tfbreak-plugin-sdk/tflint"
)

// Issue represents a finding from a rule for test assertions.
type Issue struct {
	// Rule is the rule that emitted the issue.
	Rule tflint.Rule
	// Message is the issue message.
	Message string
	// Range is the source location of the issue.
	Range hcl.Range
}

// Issues is a slice of Issue for convenience.
type Issues []Issue

// AssertIssues compares expected and actual issues.
// It ignores issue order and byte positions in ranges.
//
// Example:
//
//	helper.AssertIssues(t, helper.Issues{
//	    {Rule: rule, Message: "location changed"},
//	}, runner.Issues)
func AssertIssues(t *testing.T, want, got Issues) {
	t.Helper()

	opts := []cmp.Option{
		// Ignore byte positions (only compare line/column)
		cmpopts.IgnoreFields(hcl.Pos{}, "Byte"),
		// Ignore issue order
		cmpopts.SortSlices(func(a, b Issue) bool {
			if a.Message != b.Message {
				return a.Message < b.Message
			}
			if a.Range.Filename != b.Range.Filename {
				return a.Range.Filename < b.Range.Filename
			}
			return a.Range.Start.Line < b.Range.Start.Line
		}),
		// Compare rules by name only
		cmp.Comparer(func(a, b tflint.Rule) bool {
			if a == nil && b == nil {
				return true
			}
			if a == nil || b == nil {
				return false
			}
			return a.Name() == b.Name()
		}),
	}

	if diff := cmp.Diff(want, got, opts...); diff != "" {
		t.Errorf("issues mismatch (-want +got):\n%s", diff)
	}
}

// AssertIssuesWithoutRange compares issues ignoring the Range field entirely.
// Use this when exact source locations are not important for the test.
//
// Example:
//
//	helper.AssertIssuesWithoutRange(t, helper.Issues{
//	    {Rule: rule, Message: "location changed"},
//	}, runner.Issues)
func AssertIssuesWithoutRange(t *testing.T, want, got Issues) {
	t.Helper()

	opts := []cmp.Option{
		// Ignore Range field entirely
		cmpopts.IgnoreFields(Issue{}, "Range"),
		// Ignore issue order
		cmpopts.SortSlices(func(a, b Issue) bool {
			return a.Message < b.Message
		}),
		// Compare rules by name only
		cmp.Comparer(func(a, b tflint.Rule) bool {
			if a == nil && b == nil {
				return true
			}
			if a == nil || b == nil {
				return false
			}
			return a.Name() == b.Name()
		}),
	}

	if diff := cmp.Diff(want, got, opts...); diff != "" {
		t.Errorf("issues mismatch (-want +got):\n%s", diff)
	}
}

// AssertNoIssues verifies that no issues were emitted.
func AssertNoIssues(t *testing.T, got Issues) {
	t.Helper()
	if len(got) > 0 {
		t.Errorf("expected no issues, got %d:", len(got))
		for i, issue := range got {
			t.Errorf("  [%d] %s: %s", i, issue.Rule.Name(), issue.Message)
		}
	}
}
