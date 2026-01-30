# Deviation from tflint-plugin-sdk

This document describes the key differences between tfbreak-plugin-sdk and tflint-plugin-sdk. Understanding these differences is crucial for developers familiar with tflint who are building tfbreak plugins.

## Design Philosophy

tfbreak-plugin-sdk is intentionally aligned with tflint-plugin-sdk for ecosystem familiarity. Interface names, type names, and package structure mirror tflint where possible. However, there are deliberate deviations to support tfbreak's unique requirement: **comparing two Terraform configurations** instead of analyzing a single configuration.

For the full architectural rationale, see [ADR-0001](adr/ADR-0001-tflint-aligned-plugin-sdk.md).

## The Fundamental Difference: Dual-Config Model

### tflint: Single Configuration

tflint analyzes a single Terraform configuration to find issues like:
- Invalid resource configurations
- Best practice violations
- Provider-specific warnings

```go
// tflint: Access the single configuration
content, err := runner.GetResourceContent("aws_instance", schema, nil)
```

### tfbreak: Dual Configuration

tfbreak compares two configurations (old vs new) to detect breaking changes like:
- ForceNew attribute modifications that will recreate resources
- Renamed resources
- Configuration changes that will cause downtime

```go
// tfbreak: Access both configurations
oldContent, err := runner.GetOldResourceContent("azurerm_resource_group", schema, nil)
newContent, err := runner.GetNewResourceContent("azurerm_resource_group", schema, nil)
```

## Runner Interface Deviation

This is the most significant deviation in the SDK.

### tflint Runner

```go
type Runner interface {
    GetModuleContent(schema *hclext.BodySchema, opts *GetModuleContentOption) (*hclext.BodyContent, error)
    GetResourceContent(resourceType string, schema *hclext.BodySchema, opts *GetModuleContentOption) (*hclext.BodyContent, error)
    EmitIssue(rule Rule, message string, issueRange hcl.Range) error
    // ... other methods
}
```

### tfbreak Runner

```go
type Runner interface {
    // Old configuration access (DEVIATION)
    GetOldModuleContent(schema *hclext.BodySchema, opts *GetModuleContentOption) (*hclext.BodyContent, error)
    GetOldResourceContent(resourceType string, schema *hclext.BodySchema, opts *GetModuleContentOption) (*hclext.BodyContent, error)

    // New configuration access (DEVIATION)
    GetNewModuleContent(schema *hclext.BodySchema, opts *GetModuleContentOption) (*hclext.BodyContent, error)
    GetNewResourceContent(resourceType string, schema *hclext.BodySchema, opts *GetModuleContentOption) (*hclext.BodyContent, error)

    // Issue reporting (same as tflint)
    EmitIssue(rule Rule, message string, issueRange hcl.Range) error

    // Rule configuration (same concept as tflint)
    DecodeRuleConfig(ruleName string, target any) error
}
```

### Method Mapping

| tflint Method | tfbreak Equivalent |
|---------------|-------------------|
| `GetModuleContent()` | `GetOldModuleContent()` + `GetNewModuleContent()` |
| `GetResourceContent()` | `GetOldResourceContent()` + `GetNewResourceContent()` |
| `EmitIssue()` | `EmitIssue()` (identical) |

## TestRunner Deviation

The helper package's TestRunner also reflects the dual-config model.

### tflint TestRunner

```go
// Single configuration map
runner := helper.TestRunner(t, map[string]string{
    "main.tf": `resource "aws_instance" "foo" { ... }`,
})
```

### tfbreak TestRunner

```go
// Two configuration maps: old and new
runner := helper.TestRunner(t,
    // Old (baseline) configuration
    map[string]string{
        "main.tf": `resource "azurerm_resource_group" "rg" { location = "westeurope" }`,
    },
    // New (changed) configuration
    map[string]string{
        "main.tf": `resource "azurerm_resource_group" "rg" { location = "eastus" }`,
    },
)
```

## Features NOT Implemented

The following tflint-plugin-sdk features are intentionally not implemented in tfbreak-plugin-sdk:

| Feature | tflint Use Case | Reason for Omission |
|---------|----------------|---------------------|
| `EvaluateExpr` | Evaluate Terraform expressions | tfbreak compares static configs, doesn't need evaluation |
| `EmitIssueWithFix` | Auto-fix issues | Autofix doesn't apply to breaking change detection |
| `Fixer` interface | Provide fixes for issues | Same as above |
| `WalkExpressions` | Traverse all expressions | Not needed for attribute comparison |
| Module expansion | Analyze expanded module calls | tfbreak works with flat configs |
| Provider content | Access provider configurations | Not needed for breaking change detection |
| Dynamic block expansion | Expand dynamic blocks | tfbreak compares configs as-written |
| `GetOriginalwd` | Get original working directory | Not applicable |
| `GetModulePath` | Get module path | Not needed |

## What Remains Identical

These aspects are identical between tflint-plugin-sdk and tfbreak-plugin-sdk:

### Interfaces and Types

| Component | Notes |
|-----------|-------|
| `Rule` interface | Identical: `Name()`, `Enabled()`, `Severity()`, `Link()`, `Check()` |
| `RuleSet` interface | Identical structure and methods |
| `Severity` type | Identical: `ERROR`, `WARNING`, `NOTICE` |
| `DefaultRule` | Identical embedding pattern |
| `BuiltinRuleSet` | Identical embedding pattern |

### hclext Package

| Component | Notes |
|-----------|-------|
| `BodySchema` | Identical structure |
| `BodyContent` | Identical structure |
| `Attribute` | Identical structure |
| `Block` | Identical structure |
| `AttributeSchema` | Identical structure |
| `BlockSchema` | Identical structure |

### helper Package

| Component | Notes |
|-----------|-------|
| `Issue` | Identical structure |
| `Issues` | Identical type |
| `AssertIssues` | Identical function (compare issues) |
| `AssertIssuesWithoutRange` | Identical function |
| `AssertNoIssues` | Identical function |

### plugin Package

| Component | Notes |
|-----------|-------|
| `Serve()` | Same entry point pattern |
| `ServeOpts` | Same options structure |

## Migration Guide: tflint to tfbreak

If you're porting a tflint rule to tfbreak:

### 1. Update Runner Method Calls

```go
// Before (tflint)
content, err := runner.GetResourceContent("aws_instance", schema, nil)

// After (tfbreak)
oldContent, err := runner.GetOldResourceContent("aws_instance", schema, nil)
newContent, err := runner.GetNewResourceContent("aws_instance", schema, nil)
```

### 2. Add Comparison Logic

tflint rules typically check a single value against expectations. tfbreak rules compare old vs new values:

```go
// Before (tflint) - check if value is valid
if value != expectedValue {
    runner.EmitIssue(rule, "invalid value", attr.Range)
}

// After (tfbreak) - check if value changed
if oldValue != newValue {
    runner.EmitIssue(rule, "value changed (breaking)", newAttr.Range)
}
```

### 3. Update Tests

```go
// Before (tflint)
runner := helper.TestRunner(t, files)

// After (tfbreak)
runner := helper.TestRunner(t, oldFiles, newFiles)
```

### 4. Remove Unsupported Features

If your tflint rule uses `EvaluateExpr`, `EmitIssueWithFix`, or other unsupported features, you'll need to adapt your approach or remove those features.

## Example: Same Rule in Both SDKs

### tflint Version (hypothetical)

```go
func (r *LocationRule) Check(runner tflint.Runner) error {
    content, err := runner.GetResourceContent("aws_instance", schema, nil)
    if err != nil {
        return err
    }

    for _, block := range content.Blocks {
        // Check if location is in allowed list
        if loc := block.Body.Attributes["availability_zone"]; loc != nil {
            val, _ := loc.Expr.Value(nil)
            if !isAllowedZone(val.AsString()) {
                runner.EmitIssue(r, "availability zone not allowed", loc.Range)
            }
        }
    }
    return nil
}
```

### tfbreak Version

```go
func (r *LocationRule) Check(runner tflint.Runner) error {
    oldContent, err := runner.GetOldResourceContent("azurerm_resource_group", schema, nil)
    if err != nil {
        return err
    }
    newContent, err := runner.GetNewResourceContent("azurerm_resource_group", schema, nil)
    if err != nil {
        return err
    }

    // Index old resources
    oldByName := make(map[string]*hclext.Block)
    for _, block := range oldContent.Blocks {
        if len(block.Labels) >= 2 {
            oldByName[block.Labels[1]] = block
        }
    }

    // Compare each resource
    for _, newBlock := range newContent.Blocks {
        if len(newBlock.Labels) < 2 {
            continue
        }

        oldBlock, exists := oldByName[newBlock.Labels[1]]
        if !exists {
            continue // New resource, not a breaking change
        }

        // Check if location changed (ForceNew attribute)
        oldLoc := oldBlock.Body.Attributes["location"]
        newLoc := newBlock.Body.Attributes["location"]

        if oldLoc != nil && newLoc != nil {
            oldVal, _ := oldLoc.Expr.Value(nil)
            newVal, _ := newLoc.Expr.Value(nil)

            if !oldVal.RawEquals(newVal) {
                runner.EmitIssue(r, "location changed (will recreate resource)", newLoc.Range)
            }
        }
    }
    return nil
}
```

## Summary

| Aspect | tflint-plugin-sdk | tfbreak-plugin-sdk |
|--------|-------------------|-------------------|
| Purpose | Static analysis of single config | Breaking change detection between two configs |
| Runner model | Single config access | Dual config access (old/new) |
| TestRunner | Single file map | Two file maps (old/new) |
| Expression evaluation | Supported | Not supported |
| Autofix | Supported | Not supported |
| Interface names | Original | Aligned (same names) |
| Type names | Original | Aligned (same names) |
| Package structure | Original | Aligned (similar) |

The goal is maximum familiarity for developers in the Terraform ecosystem while supporting tfbreak's unique comparison-based analysis model.
