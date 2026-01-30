# HCL Extension Types

The `hclext` package provides extended HCL types for tfbreak plugins. These types align with tflint-plugin-sdk for ecosystem familiarity and are used by the Runner interface to specify what configuration content to retrieve.

## Overview

| Type | Purpose |
|------|---------|
| `BodySchema` | Defines expected attributes and blocks to extract |
| `BodyContent` | Contains extracted attributes and blocks |
| `Attribute` | An extracted HCL attribute with expression and range |
| `Block` | An extracted HCL block with labels and nested content |

## BodySchema

`BodySchema` specifies what attributes and blocks to extract from HCL configuration. Pass this to Runner methods to request specific content.

```go
type BodySchema struct {
    Attributes []AttributeSchema
    Blocks     []BlockSchema
    Mode       SchemaMode
}
```

### Basic Usage

```go
// Request specific attributes
schema := &hclext.BodySchema{
    Attributes: []hclext.AttributeSchema{
        {Name: "location", Required: true},
        {Name: "name", Required: true},
        {Name: "tags", Required: false},
    },
}

content, err := runner.GetOldResourceContent("azurerm_resource_group", schema, nil)
```

### With Nested Blocks

```go
// Request attributes and nested blocks
schema := &hclext.BodySchema{
    Attributes: []hclext.AttributeSchema{
        {Name: "name"},
        {Name: "location"},
    },
    Blocks: []hclext.BlockSchema{
        {
            Type:       "timeouts",
            LabelNames: nil,
            Body: &hclext.BodySchema{
                Attributes: []hclext.AttributeSchema{
                    {Name: "create"},
                    {Name: "delete"},
                },
            },
        },
        {
            Type:       "identity",
            LabelNames: nil,
            Body: &hclext.BodySchema{
                Attributes: []hclext.AttributeSchema{
                    {Name: "type"},
                    {Name: "identity_ids"},
                },
            },
        },
    },
}
```

### SchemaMode

Controls how schema matching behaves:

```go
const (
    SchemaDefaultMode        SchemaMode = iota  // Require explicit declarations
    SchemaJustAttributesMode                     // Extract all attributes
)
```

```go
// Extract all attributes without explicit declaration
schema := &hclext.BodySchema{
    Mode: hclext.SchemaJustAttributesMode,
}
```

## AttributeSchema

Defines an expected HCL attribute in a schema.

```go
type AttributeSchema struct {
    Name     string  // Attribute name to match
    Required bool    // Whether the attribute must be present
}
```

```go
schema := &hclext.BodySchema{
    Attributes: []hclext.AttributeSchema{
        {Name: "location", Required: true},   // Must be present
        {Name: "tags", Required: false},       // Optional
    },
}
```

## BlockSchema

Defines an expected HCL block in a schema.

```go
type BlockSchema struct {
    Type       string       // Block type (e.g., "resource", "variable")
    LabelNames []string     // Names for block labels
    Body       *BodySchema  // Schema for the block's body
}
```

### Block Types

```go
// Resource block: resource "type" "name" { ... }
{
    Type:       "resource",
    LabelNames: []string{"type", "name"},
}

// Variable block: variable "name" { ... }
{
    Type:       "variable",
    LabelNames: []string{"name"},
}

// Nested block without labels: timeouts { ... }
{
    Type:       "timeouts",
    LabelNames: nil,
}
```

## BodyContent

`BodyContent` contains extracted attributes and blocks from HCL configuration. This is returned by Runner methods.

```go
type BodyContent struct {
    Attributes map[string]*Attribute  // Extracted attributes by name
    Blocks     []*Block               // Extracted blocks
}
```

### Accessing Attributes

```go
content, _ := runner.GetOldResourceContent("azurerm_resource_group", schema, nil)

// Access attribute by name
if attr, ok := content.Attributes["location"]; ok {
    // attr is *hclext.Attribute
    val, _ := attr.Expr.Value(nil)
    fmt.Println(val.AsString())
}

// Check if attribute exists
if content.Attributes["tags"] != nil {
    // tags attribute is present
}
```

### Iterating Blocks

```go
content, _ := runner.GetNewResourceContent("azurerm_resource_group", schema, nil)

for _, block := range content.Blocks {
    // block.Type is "resource"
    // block.Labels[0] is resource type ("azurerm_resource_group")
    // block.Labels[1] is resource name
    resourceName := block.Labels[1]

    // Access block body
    if location, ok := block.Body.Attributes["location"]; ok {
        // Process location attribute
    }
}
```

## Attribute

An extracted HCL attribute with its expression and source range.

```go
type Attribute struct {
    Name      string          // Attribute name
    Expr      hcl.Expression  // Value expression
    Range     hcl.Range       // Source range of entire attribute
    NameRange hcl.Range       // Source range of attribute name
}
```

### Getting Attribute Values

```go
attr := content.Attributes["location"]

// Evaluate the expression to get the value
val, diags := attr.Expr.Value(nil)
if diags.HasErrors() {
    // Expression could not be evaluated (e.g., contains variables)
}

// Check type and get value
if val.Type() == cty.String {
    location := val.AsString()
}

// Use Range for issue reporting
runner.EmitIssue(rule, "message", attr.Range)
```

### Handling Different Value Types

```go
import "github.com/zclconf/go-cty/cty"

val, _ := attr.Expr.Value(nil)

switch {
case val.Type() == cty.String:
    str := val.AsString()

case val.Type() == cty.Number:
    num, _ := val.AsBigFloat().Float64()

case val.Type() == cty.Bool:
    b := val.True()

case val.Type().IsListType():
    for it := val.ElementIterator(); it.Next(); {
        _, elem := it.Element()
        // process elem
    }

case val.Type().IsMapType():
    for it := val.ElementIterator(); it.Next(); {
        key, elem := it.Element()
        // process key, elem
    }
}
```

## Block

An extracted HCL block with labels and nested content.

```go
type Block struct {
    Type        string      // Block type (e.g., "resource")
    Labels      []string    // Block label values
    Body        *BodyContent // Block body content
    DefRange    hcl.Range   // Source range of block definition
    TypeRange   hcl.Range   // Source range of block type
    LabelRanges []hcl.Range // Source ranges of each label
}
```

### Working with Resource Blocks

```go
content, _ := runner.GetOldResourceContent("azurerm_storage_account", schema, nil)

for _, block := range content.Blocks {
    // For resource blocks: resource "type" "name" { ... }
    resourceType := block.Labels[0]  // "azurerm_storage_account"
    resourceName := block.Labels[1]  // e.g., "example"

    // Access attributes in the block body
    if nameAttr := block.Body.Attributes["name"]; nameAttr != nil {
        // ...
    }

    // Access nested blocks
    for _, nestedBlock := range block.Body.Blocks {
        if nestedBlock.Type == "identity" {
            // Process identity block
        }
    }
}
```

## Conversion Functions

The package provides functions to convert between `hclext` types and `github.com/hashicorp/hcl/v2` types.

### ToHCLBodySchema

Converts `BodySchema` to `hcl.BodySchema`:

```go
hclSchema := hclext.ToHCLBodySchema(schema)
// Use with hcl.Body.Content() or PartialContent()
```

### FromHCLAttribute

Converts `hcl.Attribute` to `Attribute`:

```go
hclAttr := &hcl.Attribute{...}
attr := hclext.FromHCLAttribute(hclAttr)
```

### FromHCLBlock

Converts `hcl.Block` to `Block`:

```go
hclBlock := &hcl.Block{...}
block := hclext.FromHCLBlock(hclBlock)
// Note: block.Body must be populated separately
```

### FromHCLBodyContent

Converts `hcl.BodyContent` to `BodyContent`:

```go
hclContent := &hcl.BodyContent{...}
content := hclext.FromHCLBodyContent(hclContent)
```

## Complete Example

Here's a complete example showing schema definition and content processing:

```go
package rules

import (
    "fmt"

    "github.com/jokarl/tfbreak-plugin-sdk/hclext"
    "github.com/jokarl/tfbreak-plugin-sdk/tflint"
    "github.com/zclconf/go-cty/cty"
)

func (r *MyRule) Check(runner tflint.Runner) error {
    // Define schema for storage account resources
    schema := &hclext.BodySchema{
        Attributes: []hclext.AttributeSchema{
            {Name: "name", Required: true},
            {Name: "resource_group_name", Required: true},
            {Name: "location", Required: true},
            {Name: "account_tier", Required: true},
            {Name: "account_replication_type", Required: true},
        },
        Blocks: []hclext.BlockSchema{
            {
                Type: "blob_properties",
                Body: &hclext.BodySchema{
                    Attributes: []hclext.AttributeSchema{
                        {Name: "versioning_enabled"},
                    },
                },
            },
        },
    }

    // Get old and new resources
    oldContent, err := runner.GetOldResourceContent("azurerm_storage_account", schema, nil)
    if err != nil {
        return err
    }
    newContent, err := runner.GetNewResourceContent("azurerm_storage_account", schema, nil)
    if err != nil {
        return err
    }

    // Index old resources by name
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
        resourceName := newBlock.Labels[1]

        oldBlock, exists := oldByName[resourceName]
        if !exists {
            continue
        }

        // Compare account_tier (ForceNew attribute)
        if err := compareStringAttr(runner, r, "account_tier", oldBlock, newBlock); err != nil {
            return err
        }

        // Check nested blob_properties
        for _, newNested := range newBlock.Body.Blocks {
            if newNested.Type == "blob_properties" {
                // Find matching old block
                for _, oldNested := range oldBlock.Body.Blocks {
                    if oldNested.Type == "blob_properties" {
                        // Compare versioning_enabled
                        if err := compareAttr(runner, r, "versioning_enabled", oldNested, newNested); err != nil {
                            return err
                        }
                    }
                }
            }
        }
    }

    return nil
}

func compareStringAttr(runner tflint.Runner, rule tflint.Rule, attrName string, oldBlock, newBlock *hclext.Block) error {
    oldAttr := oldBlock.Body.Attributes[attrName]
    newAttr := newBlock.Body.Attributes[attrName]

    if oldAttr == nil || newAttr == nil {
        return nil
    }

    oldVal, _ := oldAttr.Expr.Value(nil)
    newVal, _ := newAttr.Expr.Value(nil)

    if oldVal.Type() == cty.String && newVal.Type() == cty.String {
        if oldVal.AsString() != newVal.AsString() {
            return runner.EmitIssue(
                rule,
                fmt.Sprintf("%s: ForceNew attribute changed from %q to %q",
                    attrName, oldVal.AsString(), newVal.AsString()),
                newAttr.Range,
            )
        }
    }

    return nil
}
```
