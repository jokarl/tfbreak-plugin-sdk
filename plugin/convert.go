// Package plugin provides gRPC-based plugin communication for tfbreak.
//
// This file contains conversion functions between protobuf types and
// native Go types used by tflint.RuleSet and tflint.Runner.

package plugin

import (
	"github.com/hashicorp/hcl/v2"
	ctyjson "github.com/zclconf/go-cty/cty/json"

	"github.com/jokarl/tfbreak-plugin-sdk/hclext"
	pb "github.com/jokarl/tfbreak-plugin-sdk/plugin/proto"
	"github.com/jokarl/tfbreak-plugin-sdk/tflint"
)

// =============================================================================
// Config Conversion
// =============================================================================

// toProtoConfig converts tflint.Config to proto.Config.
func toProtoConfig(config *tflint.Config) *pb.Config {
	if config == nil {
		return nil
	}

	protoRules := make(map[string]*pb.RuleConfig)
	for name, rc := range config.Rules {
		protoRules[name] = &pb.RuleConfig{
			Name:    rc.Name,
			Enabled: rc.Enabled,
			// Note: Body is not serialized over gRPC; use DecodeRuleConfig instead
		}
	}

	return &pb.Config{
		Rules:             protoRules,
		DisabledByDefault: config.DisabledByDefault,
		Only:              config.Only,
		PluginDir:         config.PluginDir,
	}
}

// fromProtoConfig converts proto.Config to tflint.Config.
func fromProtoConfig(config *pb.Config) *tflint.Config {
	if config == nil {
		return nil
	}

	rules := make(map[string]*tflint.RuleConfig)
	for name, rc := range config.GetRules() {
		rules[name] = &tflint.RuleConfig{
			Name:    rc.GetName(),
			Enabled: rc.GetEnabled(),
			// Note: Body is not deserialized; use DecodeRuleConfig instead
		}
	}

	return &tflint.Config{
		Rules:             rules,
		DisabledByDefault: config.GetDisabledByDefault(),
		Only:              config.GetOnly(),
		PluginDir:         config.GetPluginDir(),
	}
}

// =============================================================================
// Schema Conversion
// =============================================================================

// toProtoBodySchema converts hclext.BodySchema to proto.BodySchema.
func toProtoBodySchema(schema *hclext.BodySchema) *pb.BodySchema {
	if schema == nil {
		return nil
	}

	protoAttrs := make([]*pb.AttributeSchema, len(schema.Attributes))
	for i, attr := range schema.Attributes {
		protoAttrs[i] = &pb.AttributeSchema{
			Name:     attr.Name,
			Required: attr.Required,
		}
	}

	protoBlocks := make([]*pb.BlockSchema, len(schema.Blocks))
	for i, block := range schema.Blocks {
		protoBlocks[i] = &pb.BlockSchema{
			Type:       block.Type,
			LabelNames: block.LabelNames,
			Body:       toProtoBodySchema(block.Body),
		}
	}

	return &pb.BodySchema{
		Attributes: protoAttrs,
		Blocks:     protoBlocks,
		Mode:       pb.SchemaMode(schema.Mode),
	}
}

// fromProtoBodySchema converts proto.BodySchema to hclext.BodySchema.
func fromProtoBodySchema(schema *pb.BodySchema) *hclext.BodySchema {
	if schema == nil {
		return nil
	}

	attrs := make([]hclext.AttributeSchema, len(schema.GetAttributes()))
	for i, attr := range schema.GetAttributes() {
		attrs[i] = hclext.AttributeSchema{
			Name:     attr.GetName(),
			Required: attr.GetRequired(),
		}
	}

	blocks := make([]hclext.BlockSchema, len(schema.GetBlocks()))
	for i, block := range schema.GetBlocks() {
		blocks[i] = hclext.BlockSchema{
			Type:       block.GetType(),
			LabelNames: block.GetLabelNames(),
			Body:       fromProtoBodySchema(block.GetBody()),
		}
	}

	return &hclext.BodySchema{
		Attributes: attrs,
		Blocks:     blocks,
		Mode:       hclext.SchemaMode(schema.GetMode()),
	}
}

// =============================================================================
// Content Conversion
// =============================================================================

// toProtoBodyContent converts hclext.BodyContent to proto.BodyContent.
func toProtoBodyContent(content *hclext.BodyContent) *pb.BodyContent {
	if content == nil {
		return nil
	}

	protoAttrs := make(map[string]*pb.Attribute)
	for name, attr := range content.Attributes {
		protoAttrs[name] = toProtoAttribute(attr)
	}

	protoBlocks := make([]*pb.Block, len(content.Blocks))
	for i, block := range content.Blocks {
		protoBlocks[i] = toProtoBlock(block)
	}

	return &pb.BodyContent{
		Attributes: protoAttrs,
		Blocks:     protoBlocks,
	}
}

// fromProtoBodyContent converts proto.BodyContent to hclext.BodyContent.
func fromProtoBodyContent(content *pb.BodyContent) *hclext.BodyContent {
	if content == nil {
		return nil
	}

	attrs := make(map[string]*hclext.Attribute)
	for name, attr := range content.GetAttributes() {
		attrs[name] = fromProtoAttribute(attr)
	}

	blocks := make([]*hclext.Block, len(content.GetBlocks()))
	for i, block := range content.GetBlocks() {
		blocks[i] = fromProtoBlock(block)
	}

	return &hclext.BodyContent{
		Attributes: attrs,
		Blocks:     blocks,
	}
}

// toProtoAttribute converts hclext.Attribute to proto.Attribute.
func toProtoAttribute(attr *hclext.Attribute) *pb.Attribute {
	if attr == nil {
		return nil
	}

	protoAttr := &pb.Attribute{
		Name:      attr.Name,
		Range:     toProtoRange(attr.Range),
		NameRange: toProtoRange(attr.NameRange),
	}

	// Serialize expression value if available
	if attr.Expr != nil {
		// Try to evaluate the expression and serialize the value
		val, diags := attr.Expr.Value(nil)
		if !diags.HasErrors() && val.IsKnown() && !val.IsNull() {
			// Serialize the cty value as JSON
			jsonBytes, err := ctyjson.Marshal(val, val.Type())
			if err == nil {
				protoAttr.ExprValue = jsonBytes
			}
		}
	}

	return protoAttr
}

// fromProtoAttribute converts proto.Attribute to hclext.Attribute.
func fromProtoAttribute(attr *pb.Attribute) *hclext.Attribute {
	if attr == nil {
		return nil
	}

	hclAttr := &hclext.Attribute{
		Name:      attr.GetName(),
		Range:     fromProtoRange(attr.GetRange()),
		NameRange: fromProtoRange(attr.GetNameRange()),
		// Expr cannot be reconstructed from proto; use Value instead
	}

	// Reconstruct the Value from the serialized JSON
	if len(attr.GetExprValue()) > 0 {
		var simpleType ctyjson.SimpleJSONValue
		if err := simpleType.UnmarshalJSON(attr.GetExprValue()); err == nil {
			hclAttr.Value = simpleType.Value
		}
	}

	return hclAttr
}

// toProtoBlock converts hclext.Block to proto.Block.
func toProtoBlock(block *hclext.Block) *pb.Block {
	if block == nil {
		return nil
	}

	labelRanges := make([]*pb.Range, len(block.LabelRanges))
	for i, r := range block.LabelRanges {
		labelRanges[i] = toProtoRange(r)
	}

	return &pb.Block{
		Type:        block.Type,
		Labels:      block.Labels,
		Body:        toProtoBodyContent(block.Body),
		DefRange:    toProtoRange(block.DefRange),
		TypeRange:   toProtoRange(block.TypeRange),
		LabelRanges: labelRanges,
	}
}

// fromProtoBlock converts proto.Block to hclext.Block.
func fromProtoBlock(block *pb.Block) *hclext.Block {
	if block == nil {
		return nil
	}

	labelRanges := make([]hcl.Range, len(block.GetLabelRanges()))
	for i, r := range block.GetLabelRanges() {
		labelRanges[i] = fromProtoRange(r)
	}

	return &hclext.Block{
		Type:        block.GetType(),
		Labels:      block.GetLabels(),
		Body:        fromProtoBodyContent(block.GetBody()),
		DefRange:    fromProtoRange(block.GetDefRange()),
		TypeRange:   fromProtoRange(block.GetTypeRange()),
		LabelRanges: labelRanges,
	}
}

// =============================================================================
// Range Conversion
// =============================================================================

// toProtoRange converts hcl.Range to proto.Range.
func toProtoRange(r hcl.Range) *pb.Range {
	return &pb.Range{
		Filename: r.Filename,
		Start:    toProtoPosition(r.Start),
		End:      toProtoPosition(r.End),
	}
}

// fromProtoRange converts proto.Range to hcl.Range.
func fromProtoRange(r *pb.Range) hcl.Range {
	if r == nil {
		return hcl.Range{}
	}
	return hcl.Range{
		Filename: r.GetFilename(),
		Start:    fromProtoPosition(r.GetStart()),
		End:      fromProtoPosition(r.GetEnd()),
	}
}

// toProtoPosition converts hcl.Pos to proto.Position.
func toProtoPosition(p hcl.Pos) *pb.Position {
	return &pb.Position{
		Line:   int64(p.Line),
		Column: int64(p.Column),
		Byte:   int64(p.Byte),
	}
}

// fromProtoPosition converts proto.Position to hcl.Pos.
func fromProtoPosition(p *pb.Position) hcl.Pos {
	if p == nil {
		return hcl.Pos{}
	}
	return hcl.Pos{
		Line:   int(p.GetLine()),
		Column: int(p.GetColumn()),
		Byte:   int(p.GetByte()),
	}
}

// =============================================================================
// Rule Conversion
// =============================================================================

// toProtoRule converts a tflint.Rule to proto.Rule.
func toProtoRule(rule tflint.Rule) *pb.Rule {
	if rule == nil {
		return nil
	}
	return &pb.Rule{
		Name:     rule.Name(),
		Enabled:  rule.Enabled(),
		Severity: toProtoSeverity(rule.Severity()),
		Link:     rule.Link(),
	}
}

// toProtoSeverity converts tflint.Severity to proto.Severity.
func toProtoSeverity(s tflint.Severity) pb.Severity {
	switch s {
	case tflint.ERROR:
		return pb.Severity_SEVERITY_ERROR
	case tflint.WARNING:
		return pb.Severity_SEVERITY_WARNING
	case tflint.NOTICE:
		return pb.Severity_SEVERITY_NOTICE
	default:
		return pb.Severity_SEVERITY_UNSPECIFIED
	}
}

// fromProtoSeverity converts proto.Severity to tflint.Severity.
func fromProtoSeverity(s pb.Severity) tflint.Severity {
	switch s {
	case pb.Severity_SEVERITY_ERROR:
		return tflint.ERROR
	case pb.Severity_SEVERITY_WARNING:
		return tflint.WARNING
	case pb.Severity_SEVERITY_NOTICE:
		return tflint.NOTICE
	default:
		return tflint.ERROR
	}
}

// =============================================================================
// Option Conversion
// =============================================================================

// toProtoGetModuleContentOption converts tflint.GetModuleContentOption to proto.
func toProtoGetModuleContentOption(opt *tflint.GetModuleContentOption) *pb.GetModuleContentOption {
	if opt == nil {
		return nil
	}
	return &pb.GetModuleContentOption{
		ModuleCtx:        pb.ModuleCtxType(opt.ModuleCtx),
		ExpandMode:       pb.ExpandMode(opt.ExpandMode),
		ResourceTypeHint: opt.Hint.ResourceType,
	}
}

// fromProtoGetModuleContentOption converts proto.GetModuleContentOption to tflint.
func fromProtoGetModuleContentOption(opt *pb.GetModuleContentOption) *tflint.GetModuleContentOption {
	if opt == nil {
		return nil
	}
	return &tflint.GetModuleContentOption{
		ModuleCtx:  tflint.ModuleCtxType(opt.GetModuleCtx()),
		ExpandMode: tflint.ExpandMode(opt.GetExpandMode()),
		Hint: tflint.GetModuleContentHint{
			ResourceType: opt.GetResourceTypeHint(),
		},
	}
}
