# Core Interfaces

This document provides a detailed reference for the core interfaces in tfbreak-plugin-sdk.

## Overview

The SDK provides four main interfaces that plugin authors interact with:

| Interface | Purpose |
|-----------|---------|
| `Rule` | Defines a single detection rule |
| `RuleSet` | Groups rules into a plugin |
| `Runner` | Provides config access and issue emission |
| `Severity` | Specifies issue severity levels |

## Rule Interface

The `Rule` interface defines a single breaking change detection rule. Every rule in your plugin must implement this interface.

```go
type Rule interface {
    Name() string
    Enabled() bool
    Severity() Severity
    Link() string
    Check(runner Runner) error
}
```

### Methods

#### `Name() string`

Returns the unique identifier for the rule. Convention: lowercase with underscores.

```go
func (r *MyRule) Name() string {
    return "azurerm_resource_group_location"
}
```

#### `Enabled() bool`

Returns whether the rule is enabled by default. Most rules should return `true`.

```go
func (r *MyRule) Enabled() bool {
    return true
}
```

#### `Severity() Severity`

Returns the default severity level for issues emitted by this rule.

```go
func (r *MyRule) Severity() Severity {
    return tflint.ERROR  // or WARNING, NOTICE
}
```

#### `Link() string`

Returns a URL to documentation explaining what the rule checks and how to resolve issues.

```go
func (r *MyRule) Link() string {
    return "https://example.com/rules/my_rule"
}
```

#### `Check(runner Runner) error`

Executes the rule logic. This method should:
1. Retrieve configuration content via the runner
2. Compare old vs new configurations
3. Emit issues for detected breaking changes
4. Return `nil` on success, or an error for unexpected failures

```go
func (r *MyRule) Check(runner Runner) error {
    // Get old and new configurations
    oldContent, err := runner.GetOldResourceContent("aws_instance", schema, nil)
    if err != nil {
        return err
    }
    newContent, err := runner.GetNewResourceContent("aws_instance", schema, nil)
    if err != nil {
        return err
    }

    // Compare and emit issues
    // ...

    return nil
}
```

### DefaultRule Helper

`DefaultRule` provides default implementations for `Enabled()` and `Severity()`. Embed it in your rule struct to reduce boilerplate:

```go
type MyRule struct {
    tflint.DefaultRule  // Provides Enabled() -> true, Severity() -> ERROR
}

// You only need to implement Name(), Link(), and Check()
func (r *MyRule) Name() string { return "my_rule" }
func (r *MyRule) Link() string { return "https://example.com" }
func (r *MyRule) Check(runner Runner) error { /* ... */ }
```

Override defaults if needed:

```go
func (r *MyRule) Severity() tflint.Severity {
    return tflint.WARNING  // Override default ERROR
}
```

## RuleSet Interface

The `RuleSet` interface groups rules into a plugin and handles configuration.

```go
type RuleSet interface {
    RuleSetName() string
    RuleSetVersion() string
    RuleNames() []string
    VersionConstraint() string
    ConfigSchema() *hclext.BodySchema
    ApplyGlobalConfig(*Config) error
    ApplyConfig(*hclext.BodyContent) error
    NewRunner(Runner) (Runner, error)
    BuiltinImpl() *BuiltinRuleSet
}
```

### Methods

#### `RuleSetName() string`

Returns the name of the ruleset (e.g., "azurerm", "aws").

#### `RuleSetVersion() string`

Returns the semantic version of the ruleset (e.g., "0.1.0").

#### `RuleNames() []string`

Returns the names of all rules in the ruleset.

#### `VersionConstraint() string`

Returns the tfbreak version constraint (e.g., ">= 0.1.0").

#### `ConfigSchema() *hclext.BodySchema`

Returns the schema for plugin-specific configuration. Return `nil` if no configuration is needed.

#### `ApplyGlobalConfig(*Config) error`

Applies global tfbreak configuration (rule enable/disable, etc.).

#### `ApplyConfig(*hclext.BodyContent) error`

Applies plugin-specific configuration matching `ConfigSchema()`.

#### `NewRunner(Runner) (Runner, error)`

Optionally wraps the runner with custom behavior. Return unchanged if not needed.

### BuiltinRuleSet Helper

`BuiltinRuleSet` provides default implementations for all `RuleSet` methods. Embed it in your ruleset struct:

```go
type AzurermRuleSet struct {
    tflint.BuiltinRuleSet
}

func main() {
    plugin.Serve(&plugin.ServeOpts{
        RuleSet: &AzurermRuleSet{
            BuiltinRuleSet: tflint.BuiltinRuleSet{
                Name:       "azurerm",
                Version:    "0.1.0",
                Constraint: ">= 0.1.0",
                Rules: []tflint.Rule{
                    &ForceNewRule{},
                    &RenameRule{},
                },
            },
        },
    })
}
```

### BuiltinRuleSet Additional Methods

`BuiltinRuleSet` provides helper methods beyond the interface:

```go
// Check if a rule is enabled after configuration
if rs.IsRuleEnabled("my_rule") {
    // ...
}

// Get a rule by name
rule := rs.GetRule("my_rule")

// Get all enabled rules
enabledRules := rs.EnabledRules()
```

## Runner Interface

The `Runner` interface provides access to Terraform configurations during rule execution. This is the primary way rules interact with configuration data.

**IMPORTANT: Dual-Config Model**

Unlike tflint's Runner (single configuration), tfbreak's Runner provides access to two configurations: old (baseline) and new (changed). This is the fundamental deviation from tflint-plugin-sdk. See [deviation-from-tflint.md](deviation-from-tflint.md) for details.

```go
type Runner interface {
    GetOldModuleContent(schema *hclext.BodySchema, opts *GetModuleContentOption) (*hclext.BodyContent, error)
    GetNewModuleContent(schema *hclext.BodySchema, opts *GetModuleContentOption) (*hclext.BodyContent, error)
    GetOldResourceContent(resourceType string, schema *hclext.BodySchema, opts *GetModuleContentOption) (*hclext.BodyContent, error)
    GetNewResourceContent(resourceType string, schema *hclext.BodySchema, opts *GetModuleContentOption) (*hclext.BodyContent, error)
    EmitIssue(rule Rule, message string, issueRange hcl.Range) error
    DecodeRuleConfig(ruleName string, target any) error
}
```

### Methods

#### `GetOldModuleContent` / `GetNewModuleContent`

Retrieves module-level content from the old or new configuration.

```go
schema := &hclext.BodySchema{
    Blocks: []hclext.BlockSchema{
        {Type: "variable", LabelNames: []string{"name"}},
    },
}

oldVars, _ := runner.GetOldModuleContent(schema, nil)
newVars, _ := runner.GetNewModuleContent(schema, nil)
```

#### `GetOldResourceContent` / `GetNewResourceContent`

Retrieves resources of a specific type from the old or new configuration. This is a convenience method that filters for resource blocks.

```go
schema := &hclext.BodySchema{
    Attributes: []hclext.AttributeSchema{
        {Name: "location", Required: false},
        {Name: "name", Required: true},
    },
}

// Get all azurerm_resource_group resources from old config
oldRGs, err := runner.GetOldResourceContent("azurerm_resource_group", schema, nil)

// Get all azurerm_resource_group resources from new config
newRGs, err := runner.GetNewResourceContent("azurerm_resource_group", schema, nil)
```

#### `EmitIssue`

Reports a finding from the rule. The `issueRange` should typically point to the location in the NEW configuration where the breaking change was detected.

```go
if oldLocation != newLocation {
    runner.EmitIssue(
        rule,
        fmt.Sprintf("location: ForceNew attribute changed from %s to %s", oldLocation, newLocation),
        newLocationAttr.Range,
    )
}
```

#### `DecodeRuleConfig`

Retrieves and decodes rule-specific configuration. The target should be a pointer to a struct with `hcl` tags.

```go
type MyRuleConfig struct {
    IgnorePatterns []string `hcl:"ignore_patterns,optional"`
    Threshold      int      `hcl:"threshold,optional"`
}

var config MyRuleConfig
if err := runner.DecodeRuleConfig("my_rule", &config); err != nil {
    return err
}
// config is now populated if configuration was provided
```

### GetModuleContentOption

Options for controlling content retrieval:

```go
type GetModuleContentOption struct {
    ModuleCtx  ModuleCtxType
    ExpandMode ExpandMode
    Hint       GetModuleContentHint
}
```

- `ModuleCtx`: Which module context to use (`ModuleCtxSelf`, `ModuleCtxRoot`, `ModuleCtxAll`)
- `ExpandMode`: How to handle dynamic blocks (`ExpandModeNone`, `ExpandModeExpand`)
- `Hint`: Optimization hints

Most rules can pass `nil` for options to use defaults.

## Severity Type

`Severity` represents the severity level of an issue.

```go
const (
    ERROR   Severity = iota + 1  // Critical issue (e.g., resource recreation)
    WARNING                       // Potential issue needing attention
    NOTICE                        // Informational finding
)
```

Use severity to communicate the impact of findings:

| Severity | Use Case |
|----------|----------|
| `ERROR` | Breaking changes that will cause resource recreation |
| `WARNING` | Potential issues that may cause problems |
| `NOTICE` | Informational changes worth noting |

```go
// In rule implementation
func (r *MyRule) Severity() tflint.Severity {
    return tflint.ERROR  // Default for breaking changes
}
```

## Config Types

### Config

Global tfbreak configuration passed to plugins:

```go
type Config struct {
    Rules             map[string]*RuleConfig
    DisabledByDefault bool
    Only              []string
    PluginDir         string
}
```

### RuleConfig

Per-rule configuration:

```go
type RuleConfig struct {
    Name    string
    Enabled bool
    Body    hcl.Body  // Rule-specific configuration
}
```

## Complete Example

Here's a complete rule implementation using all the concepts:

```go
package rules

import (
    "fmt"

    "github.com/hashicorp/hcl/v2"
    "github.com/jokarl/tfbreak-plugin-sdk/hclext"
    "github.com/jokarl/tfbreak-plugin-sdk/tflint"
)

type ResourceGroupLocationRule struct {
    tflint.DefaultRule
}

func (r *ResourceGroupLocationRule) Name() string {
    return "azurerm_resource_group_location"
}

func (r *ResourceGroupLocationRule) Link() string {
    return "https://registry.terraform.io/providers/hashicorp/azurerm/latest/docs/resources/resource_group"
}

func (r *ResourceGroupLocationRule) Check(runner tflint.Runner) error {
    schema := &hclext.BodySchema{
        Attributes: []hclext.AttributeSchema{
            {Name: "location"},
        },
    }

    oldContent, err := runner.GetOldResourceContent("azurerm_resource_group", schema, nil)
    if err != nil {
        return err
    }

    newContent, err := runner.GetNewResourceContent("azurerm_resource_group", schema, nil)
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

    // Check each new resource
    for _, newBlock := range newContent.Blocks {
        if len(newBlock.Labels) < 2 {
            continue
        }

        oldBlock, exists := oldByName[newBlock.Labels[1]]
        if !exists {
            continue // New resource
        }

        oldLoc := oldBlock.Body.Attributes["location"]
        newLoc := newBlock.Body.Attributes["location"]

        if oldLoc == nil || newLoc == nil {
            continue
        }

        oldVal, _ := oldLoc.Expr.Value(nil)
        newVal, _ := newLoc.Expr.Value(nil)

        if !oldVal.RawEquals(newVal) {
            runner.EmitIssue(
                r,
                fmt.Sprintf("location: ForceNew attribute changed (will recreate resource)"),
                newLoc.Range,
            )
        }
    }

    return nil
}
```
