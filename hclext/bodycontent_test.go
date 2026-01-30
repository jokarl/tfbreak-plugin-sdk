package hclext

import (
	"reflect"
	"testing"

	"github.com/hashicorp/hcl/v2"
)

func TestSchemaMode_Values(t *testing.T) {
	tests := []struct {
		name string
		mode SchemaMode
		want int
	}{
		{"SchemaDefaultMode is 0", SchemaDefaultMode, 0},
		{"SchemaJustAttributesMode is 1", SchemaJustAttributesMode, 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := int(tt.mode); got != tt.want {
				t.Errorf("SchemaMode = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestToHCLBodySchema_Nil(t *testing.T) {
	result := ToHCLBodySchema(nil)
	if result != nil {
		t.Errorf("ToHCLBodySchema(nil) = %v, want nil", result)
	}
}

func TestToHCLBodySchema_Attributes(t *testing.T) {
	schema := &BodySchema{
		Attributes: []AttributeSchema{
			{Name: "location", Required: true},
			{Name: "name", Required: false},
		},
	}

	result := ToHCLBodySchema(schema)

	if len(result.Attributes) != 2 {
		t.Fatalf("got %d attributes, want 2", len(result.Attributes))
	}

	if result.Attributes[0].Name != "location" {
		t.Errorf("Attributes[0].Name = %q, want %q", result.Attributes[0].Name, "location")
	}
	if !result.Attributes[0].Required {
		t.Error("Attributes[0].Required = false, want true")
	}

	if result.Attributes[1].Name != "name" {
		t.Errorf("Attributes[1].Name = %q, want %q", result.Attributes[1].Name, "name")
	}
	if result.Attributes[1].Required {
		t.Error("Attributes[1].Required = true, want false")
	}
}

func TestToHCLBodySchema_Blocks(t *testing.T) {
	schema := &BodySchema{
		Blocks: []BlockSchema{
			{Type: "resource", LabelNames: []string{"type", "name"}},
			{Type: "timeouts", LabelNames: nil},
		},
	}

	result := ToHCLBodySchema(schema)

	if len(result.Blocks) != 2 {
		t.Fatalf("got %d blocks, want 2", len(result.Blocks))
	}

	if result.Blocks[0].Type != "resource" {
		t.Errorf("Blocks[0].Type = %q, want %q", result.Blocks[0].Type, "resource")
	}
	if !reflect.DeepEqual(result.Blocks[0].LabelNames, []string{"type", "name"}) {
		t.Errorf("Blocks[0].LabelNames = %v, want %v", result.Blocks[0].LabelNames, []string{"type", "name"})
	}

	if result.Blocks[1].Type != "timeouts" {
		t.Errorf("Blocks[1].Type = %q, want %q", result.Blocks[1].Type, "timeouts")
	}
}

func TestToHCLBodySchema_Empty(t *testing.T) {
	schema := &BodySchema{}

	result := ToHCLBodySchema(schema)

	if result == nil {
		t.Fatal("ToHCLBodySchema(&BodySchema{}) = nil, want non-nil")
	}
	if len(result.Attributes) != 0 {
		t.Errorf("got %d attributes, want 0", len(result.Attributes))
	}
	if len(result.Blocks) != 0 {
		t.Errorf("got %d blocks, want 0", len(result.Blocks))
	}
}

func TestFromHCLAttribute_Nil(t *testing.T) {
	result := FromHCLAttribute(nil)
	if result != nil {
		t.Errorf("FromHCLAttribute(nil) = %v, want nil", result)
	}
}

func TestFromHCLAttribute_Valid(t *testing.T) {
	hclAttr := &hcl.Attribute{
		Name: "location",
		Range: hcl.Range{
			Filename: "main.tf",
			Start:    hcl.Pos{Line: 1, Column: 1, Byte: 0},
			End:      hcl.Pos{Line: 1, Column: 20, Byte: 19},
		},
		NameRange: hcl.Range{
			Filename: "main.tf",
			Start:    hcl.Pos{Line: 1, Column: 1, Byte: 0},
			End:      hcl.Pos{Line: 1, Column: 9, Byte: 8},
		},
	}

	result := FromHCLAttribute(hclAttr)

	if result.Name != "location" {
		t.Errorf("Name = %q, want %q", result.Name, "location")
	}
	if result.Range.Filename != "main.tf" {
		t.Errorf("Range.Filename = %q, want %q", result.Range.Filename, "main.tf")
	}
	if result.NameRange.Start.Line != 1 {
		t.Errorf("NameRange.Start.Line = %d, want 1", result.NameRange.Start.Line)
	}
}

func TestFromHCLBlock_Nil(t *testing.T) {
	result := FromHCLBlock(nil)
	if result != nil {
		t.Errorf("FromHCLBlock(nil) = %v, want nil", result)
	}
}

func TestFromHCLBlock_Valid(t *testing.T) {
	hclBlock := &hcl.Block{
		Type:   "resource",
		Labels: []string{"azurerm_resource_group", "example"},
		DefRange: hcl.Range{
			Filename: "main.tf",
			Start:    hcl.Pos{Line: 1, Column: 1, Byte: 0},
			End:      hcl.Pos{Line: 1, Column: 50, Byte: 49},
		},
		TypeRange: hcl.Range{
			Filename: "main.tf",
			Start:    hcl.Pos{Line: 1, Column: 1, Byte: 0},
			End:      hcl.Pos{Line: 1, Column: 9, Byte: 8},
		},
		LabelRanges: []hcl.Range{
			{Filename: "main.tf", Start: hcl.Pos{Line: 1, Column: 10}, End: hcl.Pos{Line: 1, Column: 35}},
			{Filename: "main.tf", Start: hcl.Pos{Line: 1, Column: 36}, End: hcl.Pos{Line: 1, Column: 45}},
		},
	}

	result := FromHCLBlock(hclBlock)

	if result.Type != "resource" {
		t.Errorf("Type = %q, want %q", result.Type, "resource")
	}
	if !reflect.DeepEqual(result.Labels, []string{"azurerm_resource_group", "example"}) {
		t.Errorf("Labels = %v, want %v", result.Labels, []string{"azurerm_resource_group", "example"})
	}
	if result.Body != nil {
		t.Error("Body should be nil (must be populated by caller)")
	}
	if result.DefRange.Filename != "main.tf" {
		t.Errorf("DefRange.Filename = %q, want %q", result.DefRange.Filename, "main.tf")
	}
	if len(result.LabelRanges) != 2 {
		t.Errorf("got %d LabelRanges, want 2", len(result.LabelRanges))
	}
}

func TestFromHCLBodyContent_Nil(t *testing.T) {
	result := FromHCLBodyContent(nil)
	if result != nil {
		t.Errorf("FromHCLBodyContent(nil) = %v, want nil", result)
	}
}

func TestFromHCLBodyContent_Valid(t *testing.T) {
	hclContent := &hcl.BodyContent{
		Attributes: hcl.Attributes{
			"name": &hcl.Attribute{
				Name: "name",
				Range: hcl.Range{
					Filename: "main.tf",
					Start:    hcl.Pos{Line: 2, Column: 3},
					End:      hcl.Pos{Line: 2, Column: 20},
				},
			},
			"location": &hcl.Attribute{
				Name: "location",
				Range: hcl.Range{
					Filename: "main.tf",
					Start:    hcl.Pos{Line: 3, Column: 3},
					End:      hcl.Pos{Line: 3, Column: 25},
				},
			},
		},
		Blocks: hcl.Blocks{
			&hcl.Block{
				Type: "timeouts",
				DefRange: hcl.Range{
					Filename: "main.tf",
					Start:    hcl.Pos{Line: 5, Column: 3},
					End:      hcl.Pos{Line: 5, Column: 15},
				},
			},
		},
	}

	result := FromHCLBodyContent(hclContent)

	if len(result.Attributes) != 2 {
		t.Fatalf("got %d attributes, want 2", len(result.Attributes))
	}
	if result.Attributes["name"] == nil {
		t.Error("Attributes[\"name\"] is nil")
	}
	if result.Attributes["location"] == nil {
		t.Error("Attributes[\"location\"] is nil")
	}

	if len(result.Blocks) != 1 {
		t.Fatalf("got %d blocks, want 1", len(result.Blocks))
	}
	if result.Blocks[0].Type != "timeouts" {
		t.Errorf("Blocks[0].Type = %q, want %q", result.Blocks[0].Type, "timeouts")
	}
}

func TestFromHCLBodyContent_Empty(t *testing.T) {
	hclContent := &hcl.BodyContent{
		Attributes: hcl.Attributes{},
		Blocks:     hcl.Blocks{},
	}

	result := FromHCLBodyContent(hclContent)

	if result == nil {
		t.Fatal("FromHCLBodyContent returned nil for empty content")
	}
	if len(result.Attributes) != 0 {
		t.Errorf("got %d attributes, want 0", len(result.Attributes))
	}
	if len(result.Blocks) != 0 {
		t.Errorf("got %d blocks, want 0", len(result.Blocks))
	}
}
