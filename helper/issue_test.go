package helper

import (
	"testing"

	"github.com/hashicorp/hcl/v2"
	"github.com/jokarl/tfbreak-plugin-sdk/tflint"
)

// testRuleForIssue is a minimal rule for testing assertions.
type testRuleForIssue struct {
	tflint.DefaultRule
	name string
}

func (r *testRuleForIssue) Name() string        { return r.name }
func (r *testRuleForIssue) Link() string        { return "" }
func (r *testRuleForIssue) Check(_ tflint.Runner) error { return nil }

func TestAssertIssues_Match(t *testing.T) {
	rule := &testRuleForIssue{name: "test_rule"}
	issueRange := hcl.Range{
		Filename: "main.tf",
		Start:    hcl.Pos{Line: 3, Column: 3},
		End:      hcl.Pos{Line: 3, Column: 20},
	}

	want := Issues{{Rule: rule, Message: "test message", Range: issueRange}}
	got := Issues{{Rule: rule, Message: "test message", Range: issueRange}}

	// This should pass without error
	AssertIssues(t, want, got)
}

func TestAssertIssues_IgnoresOrder(t *testing.T) {
	rule1 := &testRuleForIssue{name: "rule1"}
	rule2 := &testRuleForIssue{name: "rule2"}

	want := Issues{
		{Rule: rule1, Message: "message 1"},
		{Rule: rule2, Message: "message 2"},
	}
	got := Issues{
		{Rule: rule2, Message: "message 2"},
		{Rule: rule1, Message: "message 1"},
	}

	// Should pass even though order is different
	AssertIssues(t, want, got)
}

func TestAssertIssues_IgnoresBytePos(t *testing.T) {
	rule := &testRuleForIssue{name: "test_rule"}

	want := Issues{{
		Rule:    rule,
		Message: "test message",
		Range: hcl.Range{
			Filename: "main.tf",
			Start:    hcl.Pos{Line: 3, Column: 3, Byte: 100},
			End:      hcl.Pos{Line: 3, Column: 20, Byte: 117},
		},
	}}
	got := Issues{{
		Rule:    rule,
		Message: "test message",
		Range: hcl.Range{
			Filename: "main.tf",
			Start:    hcl.Pos{Line: 3, Column: 3, Byte: 50}, // Different byte
			End:      hcl.Pos{Line: 3, Column: 20, Byte: 67}, // Different byte
		},
	}}

	// Should pass because byte positions are ignored
	AssertIssues(t, want, got)
}

func TestAssertIssues_Empty(t *testing.T) {
	want := Issues{}
	got := Issues{}

	AssertIssues(t, want, got)
}

func TestAssertIssuesWithoutRange_Match(t *testing.T) {
	rule := &testRuleForIssue{name: "test_rule"}

	want := Issues{{
		Rule:    rule,
		Message: "test message",
		Range: hcl.Range{
			Filename: "main.tf",
			Start:    hcl.Pos{Line: 1, Column: 1},
			End:      hcl.Pos{Line: 1, Column: 10},
		},
	}}
	got := Issues{{
		Rule:    rule,
		Message: "test message",
		Range: hcl.Range{
			Filename: "other.tf", // Different file
			Start:    hcl.Pos{Line: 99, Column: 99}, // Different position
			End:      hcl.Pos{Line: 99, Column: 199},
		},
	}}

	// Should pass because Range is completely ignored
	AssertIssuesWithoutRange(t, want, got)
}

func TestAssertIssuesWithoutRange_IgnoresOrder(t *testing.T) {
	rule1 := &testRuleForIssue{name: "rule1"}
	rule2 := &testRuleForIssue{name: "rule2"}

	want := Issues{
		{Rule: rule1, Message: "aaa"},
		{Rule: rule2, Message: "bbb"},
	}
	got := Issues{
		{Rule: rule2, Message: "bbb"},
		{Rule: rule1, Message: "aaa"},
	}

	AssertIssuesWithoutRange(t, want, got)
}

func TestAssertNoIssues_Empty(t *testing.T) {
	got := Issues{}

	// This should pass
	AssertNoIssues(t, got)
}

func TestAssertIssues_NilRules(t *testing.T) {
	want := Issues{{Rule: nil, Message: "no rule"}}
	got := Issues{{Rule: nil, Message: "no rule"}}

	AssertIssues(t, want, got)
}

// Note: Testing assertion failures would require interfaces instead of *testing.T.
// For now, we only test successful comparisons. The assertion functions are
// simple wrappers around go-cmp, so extensive failure testing is not critical.
