package tflint

import "testing"

func TestModuleCtxType_Values(t *testing.T) {
	tests := []struct {
		name string
		ctx  ModuleCtxType
		want int
	}{
		{"ModuleCtxSelf is 0", ModuleCtxSelf, 0},
		{"ModuleCtxRoot is 1", ModuleCtxRoot, 1},
		{"ModuleCtxAll is 2", ModuleCtxAll, 2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := int(tt.ctx); got != tt.want {
				t.Errorf("ModuleCtxType = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestExpandMode_Values(t *testing.T) {
	tests := []struct {
		name string
		mode ExpandMode
		want int
	}{
		{"ExpandModeNone is 0", ExpandModeNone, 0},
		{"ExpandModeExpand is 1", ExpandModeExpand, 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := int(tt.mode); got != tt.want {
				t.Errorf("ExpandMode = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestGetModuleContentOption_Fields(t *testing.T) {
	// Test that GetModuleContentOption can be instantiated with all fields
	opt := GetModuleContentOption{
		ModuleCtx:  ModuleCtxRoot,
		ExpandMode: ExpandModeNone,
		Hint: GetModuleContentHint{
			ResourceType: "azurerm_resource_group",
		},
	}

	if opt.ModuleCtx != ModuleCtxRoot {
		t.Errorf("ModuleCtx = %v, want ModuleCtxRoot", opt.ModuleCtx)
	}
	if opt.ExpandMode != ExpandModeNone {
		t.Errorf("ExpandMode = %v, want ExpandModeNone", opt.ExpandMode)
	}
	if opt.Hint.ResourceType != "azurerm_resource_group" {
		t.Errorf("Hint.ResourceType = %q, want %q", opt.Hint.ResourceType, "azurerm_resource_group")
	}
}

func TestGetModuleContentHint_Fields(t *testing.T) {
	hint := GetModuleContentHint{
		ResourceType: "aws_instance",
	}

	if hint.ResourceType != "aws_instance" {
		t.Errorf("ResourceType = %q, want %q", hint.ResourceType, "aws_instance")
	}
}
