package plugin

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/hcl/v2"

	"github.com/jokarl/tfbreak-plugin-sdk/hclext"
	pb "github.com/jokarl/tfbreak-plugin-sdk/plugin/proto"
	"github.com/jokarl/tfbreak-plugin-sdk/tflint"
)

func TestToProtoConfig(t *testing.T) {
	t.Run("nil config", func(t *testing.T) {
		result := toProtoConfig(nil)
		if result != nil {
			t.Errorf("expected nil, got %v", result)
		}
	})

	t.Run("with values", func(t *testing.T) {
		config := &tflint.Config{
			DisabledByDefault: true,
			Only:              []string{"rule1", "rule2"},
			PluginDir:         "/path/to/plugins",
			Rules: map[string]*tflint.RuleConfig{
				"test_rule": {
					Name:    "test_rule",
					Enabled: true,
				},
			},
		}

		result := toProtoConfig(config)

		if !result.DisabledByDefault {
			t.Error("DisabledByDefault should be true")
		}
		if len(result.Only) != 2 {
			t.Errorf("Only should have 2 items, got %d", len(result.Only))
		}
		if result.PluginDir != "/path/to/plugins" {
			t.Errorf("PluginDir = %q, want %q", result.PluginDir, "/path/to/plugins")
		}
		if rc, ok := result.Rules["test_rule"]; !ok {
			t.Error("Rules should contain test_rule")
		} else if !rc.Enabled {
			t.Error("test_rule should be enabled")
		}
	})
}

func TestFromProtoConfig(t *testing.T) {
	t.Run("nil config", func(t *testing.T) {
		result := fromProtoConfig(nil)
		if result != nil {
			t.Errorf("expected nil, got %v", result)
		}
	})

	t.Run("with values", func(t *testing.T) {
		config := &pb.Config{
			DisabledByDefault: true,
			Only:              []string{"rule1"},
			PluginDir:         "/plugins",
			Rules: map[string]*pb.RuleConfig{
				"my_rule": {
					Name:    "my_rule",
					Enabled: false,
				},
			},
		}

		result := fromProtoConfig(config)

		if !result.DisabledByDefault {
			t.Error("DisabledByDefault should be true")
		}
		if len(result.Only) != 1 {
			t.Errorf("Only should have 1 item, got %d", len(result.Only))
		}
		if rc, ok := result.Rules["my_rule"]; !ok {
			t.Error("Rules should contain my_rule")
		} else if rc.Enabled {
			t.Error("my_rule should be disabled")
		}
	})
}

func TestToProtoBodySchema(t *testing.T) {
	t.Run("nil schema", func(t *testing.T) {
		result := toProtoBodySchema(nil)
		if result != nil {
			t.Errorf("expected nil, got %v", result)
		}
	})

	t.Run("with attributes and blocks", func(t *testing.T) {
		schema := &hclext.BodySchema{
			Mode: hclext.SchemaJustAttributesMode,
			Attributes: []hclext.AttributeSchema{
				{Name: "attr1", Required: true},
				{Name: "attr2", Required: false},
			},
			Blocks: []hclext.BlockSchema{
				{
					Type:       "block1",
					LabelNames: []string{"label1", "label2"},
					Body: &hclext.BodySchema{
						Attributes: []hclext.AttributeSchema{
							{Name: "nested_attr", Required: true},
						},
					},
				},
			},
		}

		result := toProtoBodySchema(schema)

		if result.Mode != pb.SchemaMode_SCHEMA_MODE_JUST_ATTRIBUTES {
			t.Errorf("Mode = %v, want SCHEMA_MODE_JUST_ATTRIBUTES", result.Mode)
		}
		if len(result.Attributes) != 2 {
			t.Errorf("Attributes length = %d, want 2", len(result.Attributes))
		}
		if result.Attributes[0].Name != "attr1" || !result.Attributes[0].Required {
			t.Error("First attribute should be attr1 with Required=true")
		}
		if len(result.Blocks) != 1 {
			t.Errorf("Blocks length = %d, want 1", len(result.Blocks))
		}
		if result.Blocks[0].Type != "block1" {
			t.Errorf("Block type = %q, want %q", result.Blocks[0].Type, "block1")
		}
		if len(result.Blocks[0].LabelNames) != 2 {
			t.Errorf("Block label names length = %d, want 2", len(result.Blocks[0].LabelNames))
		}
		if result.Blocks[0].Body == nil {
			t.Error("Block body should not be nil")
		}
	})
}

func TestFromProtoBodySchema(t *testing.T) {
	t.Run("nil schema", func(t *testing.T) {
		result := fromProtoBodySchema(nil)
		if result != nil {
			t.Errorf("expected nil, got %v", result)
		}
	})

	t.Run("with attributes and blocks", func(t *testing.T) {
		schema := &pb.BodySchema{
			Mode: pb.SchemaMode_SCHEMA_MODE_DEFAULT,
			Attributes: []*pb.AttributeSchema{
				{Name: "test", Required: true},
			},
			Blocks: []*pb.BlockSchema{
				{
					Type:       "resource",
					LabelNames: []string{"type", "name"},
				},
			},
		}

		result := fromProtoBodySchema(schema)

		if result.Mode != hclext.SchemaDefaultMode {
			t.Errorf("Mode = %v, want SchemaDefaultMode", result.Mode)
		}
		if len(result.Attributes) != 1 {
			t.Errorf("Attributes length = %d, want 1", len(result.Attributes))
		}
		if len(result.Blocks) != 1 {
			t.Errorf("Blocks length = %d, want 1", len(result.Blocks))
		}
	})
}

func TestToProtoBodyContent(t *testing.T) {
	t.Run("nil content", func(t *testing.T) {
		result := toProtoBodyContent(nil)
		if result != nil {
			t.Errorf("expected nil, got %v", result)
		}
	})

	t.Run("with attributes and blocks", func(t *testing.T) {
		content := &hclext.BodyContent{
			Attributes: map[string]*hclext.Attribute{
				"test_attr": {
					Name: "test_attr",
					Range: hcl.Range{
						Filename: "test.tf",
						Start:    hcl.Pos{Line: 1, Column: 1, Byte: 0},
						End:      hcl.Pos{Line: 1, Column: 10, Byte: 9},
					},
				},
			},
			Blocks: []*hclext.Block{
				{
					Type:   "resource",
					Labels: []string{"aws_instance", "test"},
				},
			},
		}

		result := toProtoBodyContent(content)

		if len(result.Attributes) != 1 {
			t.Errorf("Attributes length = %d, want 1", len(result.Attributes))
		}
		if attr, ok := result.Attributes["test_attr"]; !ok {
			t.Error("Attributes should contain test_attr")
		} else if attr.Name != "test_attr" {
			t.Errorf("Attribute name = %q, want %q", attr.Name, "test_attr")
		}
		if len(result.Blocks) != 1 {
			t.Errorf("Blocks length = %d, want 1", len(result.Blocks))
		}
	})
}

func TestFromProtoBodyContent(t *testing.T) {
	t.Run("nil content", func(t *testing.T) {
		result := fromProtoBodyContent(nil)
		if result != nil {
			t.Errorf("expected nil, got %v", result)
		}
	})

	t.Run("with values", func(t *testing.T) {
		content := &pb.BodyContent{
			Attributes: map[string]*pb.Attribute{
				"name": {
					Name: "name",
					Range: &pb.Range{
						Filename: "main.tf",
						Start:    &pb.Position{Line: 2, Column: 3, Byte: 10},
						End:      &pb.Position{Line: 2, Column: 15, Byte: 22},
					},
				},
			},
			Blocks: []*pb.Block{
				{
					Type:   "variable",
					Labels: []string{"input"},
				},
			},
		}

		result := fromProtoBodyContent(content)

		if len(result.Attributes) != 1 {
			t.Errorf("Attributes length = %d, want 1", len(result.Attributes))
		}
		if len(result.Blocks) != 1 {
			t.Errorf("Blocks length = %d, want 1", len(result.Blocks))
		}
	})
}

func TestRangeConversion(t *testing.T) {
	original := hcl.Range{
		Filename: "test.tf",
		Start:    hcl.Pos{Line: 10, Column: 5, Byte: 100},
		End:      hcl.Pos{Line: 10, Column: 20, Byte: 115},
	}

	proto := toProtoRange(original)
	result := fromProtoRange(proto)

	if diff := cmp.Diff(original, result); diff != "" {
		t.Errorf("Range roundtrip mismatch (-want +got):\n%s", diff)
	}
}

func TestPositionConversion(t *testing.T) {
	original := hcl.Pos{Line: 42, Column: 13, Byte: 500}

	proto := toProtoPosition(original)
	result := fromProtoPosition(proto)

	if diff := cmp.Diff(original, result); diff != "" {
		t.Errorf("Position roundtrip mismatch (-want +got):\n%s", diff)
	}
}

func TestSeverityConversion(t *testing.T) {
	tests := []struct {
		name     string
		input    tflint.Severity
		expected pb.Severity
	}{
		{"ERROR", tflint.ERROR, pb.Severity_SEVERITY_ERROR},
		{"WARNING", tflint.WARNING, pb.Severity_SEVERITY_WARNING},
		{"NOTICE", tflint.NOTICE, pb.Severity_SEVERITY_NOTICE},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			proto := toProtoSeverity(tt.input)
			if proto != tt.expected {
				t.Errorf("toProtoSeverity(%v) = %v, want %v", tt.input, proto, tt.expected)
			}

			back := fromProtoSeverity(proto)
			if back != tt.input {
				t.Errorf("fromProtoSeverity(%v) = %v, want %v", proto, back, tt.input)
			}
		})
	}
}

func TestToProtoRule(t *testing.T) {
	t.Run("nil rule", func(t *testing.T) {
		result := toProtoRule(nil)
		if result != nil {
			t.Errorf("expected nil, got %v", result)
		}
	})

	t.Run("with rule", func(t *testing.T) {
		rule := &testRule{
			DefaultRule: tflint.DefaultRule{},
			name:        "test_rule",
		}

		result := toProtoRule(rule)

		if result.Name != "test_rule" {
			t.Errorf("Name = %q, want %q", result.Name, "test_rule")
		}
		if !result.Enabled {
			t.Error("Enabled should be true (from DefaultRule)")
		}
		if result.Severity != pb.Severity_SEVERITY_ERROR {
			t.Errorf("Severity = %v, want SEVERITY_ERROR", result.Severity)
		}
	})
}

func TestGetModuleContentOptionConversion(t *testing.T) {
	t.Run("nil option", func(t *testing.T) {
		proto := toProtoGetModuleContentOption(nil)
		if proto != nil {
			t.Errorf("expected nil, got %v", proto)
		}

		result := fromProtoGetModuleContentOption(nil)
		if result != nil {
			t.Errorf("expected nil, got %v", result)
		}
	})

	t.Run("with values", func(t *testing.T) {
		original := &tflint.GetModuleContentOption{
			ModuleCtx:  tflint.ModuleCtxRoot,
			ExpandMode: tflint.ExpandModeExpand,
			Hint: tflint.GetModuleContentHint{
				ResourceType: "aws_instance",
			},
		}

		proto := toProtoGetModuleContentOption(original)
		if proto.ModuleCtx != pb.ModuleCtxType_MODULE_CTX_ROOT {
			t.Errorf("ModuleCtx = %v, want MODULE_CTX_ROOT", proto.ModuleCtx)
		}
		if proto.ExpandMode != pb.ExpandMode_EXPAND_MODE_EXPAND {
			t.Errorf("ExpandMode = %v, want EXPAND_MODE_EXPAND", proto.ExpandMode)
		}
		if proto.ResourceTypeHint != "aws_instance" {
			t.Errorf("ResourceTypeHint = %q, want %q", proto.ResourceTypeHint, "aws_instance")
		}

		result := fromProtoGetModuleContentOption(proto)
		if result.ModuleCtx != tflint.ModuleCtxRoot {
			t.Errorf("ModuleCtx = %v, want ModuleCtxRoot", result.ModuleCtx)
		}
		if result.ExpandMode != tflint.ExpandModeExpand {
			t.Errorf("ExpandMode = %v, want ExpandModeExpand", result.ExpandMode)
		}
		if result.Hint.ResourceType != "aws_instance" {
			t.Errorf("Hint.ResourceType = %q, want %q", result.Hint.ResourceType, "aws_instance")
		}
	})
}
