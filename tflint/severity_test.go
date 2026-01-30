package tflint

import "testing"

func TestSeverity_Values(t *testing.T) {
	tests := []struct {
		name     string
		severity Severity
		want     int
	}{
		{"ERROR is 1", ERROR, 1},
		{"WARNING is 2", WARNING, 2},
		{"NOTICE is 3", NOTICE, 3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := int(tt.severity); got != tt.want {
				t.Errorf("Severity = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestSeverity_String(t *testing.T) {
	tests := []struct {
		name     string
		severity Severity
		want     string
	}{
		{"ERROR string", ERROR, "ERROR"},
		{"WARNING string", WARNING, "WARNING"},
		{"NOTICE string", NOTICE, "NOTICE"},
		{"Unknown string", Severity(99), "UNKNOWN"},
		{"Zero value string", Severity(0), "UNKNOWN"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.severity.String(); got != tt.want {
				t.Errorf("Severity.String() = %q, want %q", got, tt.want)
			}
		})
	}
}
