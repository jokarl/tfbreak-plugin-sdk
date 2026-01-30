// Package hclext provides extended HCL types for tfbreak plugins.
//
// Types align with tflint-plugin-sdk/hclext for ecosystem familiarity.
// This package provides schema and content types used by the Runner interface
// to specify what configuration content to retrieve.
//
// Key types:
//   - BodySchema: Defines expected attributes and blocks to extract
//   - BodyContent: Contains extracted attributes and blocks
//   - Attribute: An extracted HCL attribute with expression and range
//   - Block: An extracted HCL block with labels and nested content
//
// Conversion functions are provided to translate between these types
// and the github.com/hashicorp/hcl/v2 equivalents.
package hclext

import (
	"github.com/hashicorp/hcl/v2"
)

// SchemaMode specifies how schema matching behaves.
type SchemaMode int

const (
	// SchemaDefaultMode requires explicitly declared attributes and blocks.
	SchemaDefaultMode SchemaMode = iota
	// SchemaJustAttributesMode extracts all attributes without explicit declaration.
	SchemaJustAttributesMode
)

// BodySchema represents the expected structure of an HCL body.
// Use this to specify what attributes and blocks to extract from configuration.
//
// Example:
//
//	schema := &hclext.BodySchema{
//	    Attributes: []hclext.AttributeSchema{
//	        {Name: "location", Required: true},
//	        {Name: "name", Required: true},
//	    },
//	    Blocks: []hclext.BlockSchema{
//	        {Type: "timeouts", LabelNames: nil},
//	    },
//	}
type BodySchema struct {
	// Attributes defines expected attributes.
	Attributes []AttributeSchema
	// Blocks defines expected nested blocks.
	Blocks []BlockSchema
	// Mode specifies schema matching behavior.
	Mode SchemaMode
}

// AttributeSchema represents an expected HCL attribute.
type AttributeSchema struct {
	// Name is the attribute name to match.
	Name string
	// Required indicates if the attribute must be present.
	Required bool
}

// BlockSchema represents an expected HCL block.
type BlockSchema struct {
	// Type is the block type to match (e.g., "resource", "variable").
	Type string
	// LabelNames are the names for block labels (e.g., ["type", "name"] for resources).
	LabelNames []string
	// Body is the schema for the block's body content.
	Body *BodySchema
}

// BodyContent represents extracted content from an HCL body.
type BodyContent struct {
	// Attributes maps attribute names to their content.
	Attributes map[string]*Attribute
	// Blocks contains extracted block content.
	Blocks []*Block
}

// Attribute represents an extracted HCL attribute.
type Attribute struct {
	// Name is the attribute name.
	Name string
	// Expr is the attribute's value expression.
	Expr hcl.Expression
	// Range is the source range of the entire attribute.
	Range hcl.Range
	// NameRange is the source range of just the attribute name.
	NameRange hcl.Range
}

// Block represents an extracted HCL block.
type Block struct {
	// Type is the block type (e.g., "resource").
	Type string
	// Labels are the block's label values.
	Labels []string
	// Body is the block's body content.
	Body *BodyContent
	// DefRange is the source range of the block definition.
	DefRange hcl.Range
	// TypeRange is the source range of the block type.
	TypeRange hcl.Range
	// LabelRanges are the source ranges of each label.
	LabelRanges []hcl.Range
}

// ToHCLBodySchema converts a BodySchema to an hcl.BodySchema.
// This is useful when using hcl.Body.Content() or PartialContent().
func ToHCLBodySchema(schema *BodySchema) *hcl.BodySchema {
	if schema == nil {
		return nil
	}

	hclSchema := &hcl.BodySchema{
		Attributes: make([]hcl.AttributeSchema, len(schema.Attributes)),
		Blocks:     make([]hcl.BlockHeaderSchema, len(schema.Blocks)),
	}

	for i, attr := range schema.Attributes {
		hclSchema.Attributes[i] = hcl.AttributeSchema{
			Name:     attr.Name,
			Required: attr.Required,
		}
	}

	for i, block := range schema.Blocks {
		hclSchema.Blocks[i] = hcl.BlockHeaderSchema{
			Type:       block.Type,
			LabelNames: block.LabelNames,
		}
	}

	return hclSchema
}

// FromHCLAttribute converts an hcl.Attribute to an Attribute.
func FromHCLAttribute(attr *hcl.Attribute) *Attribute {
	if attr == nil {
		return nil
	}
	return &Attribute{
		Name:      attr.Name,
		Expr:      attr.Expr,
		Range:     attr.Range,
		NameRange: attr.NameRange,
	}
}

// FromHCLBlock converts an hcl.Block to a Block.
// Note: Body content must be extracted separately using the block's Body field.
func FromHCLBlock(block *hcl.Block) *Block {
	if block == nil {
		return nil
	}
	return &Block{
		Type:        block.Type,
		Labels:      block.Labels,
		Body:        nil, // Must be populated by caller
		DefRange:    block.DefRange,
		TypeRange:   block.TypeRange,
		LabelRanges: block.LabelRanges,
	}
}

// FromHCLBodyContent converts an hcl.BodyContent to a BodyContent.
// Note: Nested block bodies must be processed separately.
func FromHCLBodyContent(content *hcl.BodyContent) *BodyContent {
	if content == nil {
		return nil
	}

	bc := &BodyContent{
		Attributes: make(map[string]*Attribute, len(content.Attributes)),
		Blocks:     make([]*Block, len(content.Blocks)),
	}

	for name, attr := range content.Attributes {
		bc.Attributes[name] = FromHCLAttribute(attr)
	}

	for i, block := range content.Blocks {
		bc.Blocks[i] = FromHCLBlock(block)
	}

	return bc
}
