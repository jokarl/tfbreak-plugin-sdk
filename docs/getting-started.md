# Getting Started with tfbreak-plugin-sdk

This guide walks you through creating your first tfbreak plugin, from project setup to testing.

## Prerequisites

- Go 1.23 or later
- Basic familiarity with Go and Terraform
- Understanding of HCL (HashiCorp Configuration Language)

## Creating a New Plugin

### Step 1: Initialize the Project

Create a new directory for your plugin and initialize a Go module:

```bash
mkdir tfbreak-ruleset-myprovider
cd tfbreak-ruleset-myprovider
go mod init github.com/yourname/tfbreak-ruleset-myprovider
```

Add the SDK dependency:

```bash
go get github.com/jokarl/tfbreak-plugin-sdk
```

### Step 2: Project Structure

A typical tfbreak plugin follows this structure:

```
tfbreak-ruleset-myprovider/
├── main.go              # Plugin entry point
├── go.mod
├── go.sum
├── rules/               # Rule implementations
│   ├── rules.go         # Rule registry
│   ├── force_new.go     # Individual rule
│   └── force_new_test.go
└── README.md
```

### Step 3: Create the Entry Point

The `main.go` file is the plugin entry point. It registers your ruleset with tfbreak:

```go
// main.go
package main

import (
    "github.com/jokarl/tfbreak-plugin-sdk/plugin"
    "github.com/jokarl/tfbreak-plugin-sdk/tflint"
    "github.com/yourname/tfbreak-ruleset-myprovider/rules"
)

func main() {
    plugin.Serve(&plugin.ServeOpts{
        RuleSet: &MyProviderRuleSet{
            BuiltinRuleSet: tflint.BuiltinRuleSet{
                Name:       "myprovider",
                Version:    "0.1.0",
                Constraint: ">= 0.1.0",
                Rules:      rules.Rules,
            },
        },
    })
}

// MyProviderRuleSet wraps BuiltinRuleSet for provider-specific customization.
type MyProviderRuleSet struct {
    tflint.BuiltinRuleSet
}
```

### Step 4: Create the Rule Registry

Create a file to register all your rules:

```go
// rules/rules.go
package rules

import "github.com/jokarl/tfbreak-plugin-sdk/tflint"

// Rules contains all rules in this ruleset.
var Rules = []tflint.Rule{
    &ForceNewRule{},
    // Add more rules here
}
```

### Step 5: Implement Your First Rule

Create a rule that detects breaking changes:

```go
// rules/force_new.go
package rules

import (
    "fmt"

    "github.com/hashicorp/hcl/v2"
    "github.com/jokarl/tfbreak-plugin-sdk/hclext"
    "github.com/jokarl/tfbreak-plugin-sdk/tflint"
    "github.com/zclconf/go-cty/cty"
)

// ForceNewRule detects changes to ForceNew attributes.
type ForceNewRule struct {
    tflint.DefaultRule
}

// Name returns the rule name.
func (r *ForceNewRule) Name() string {
    return "myprovider_force_new_location"
}

// Link returns documentation URL.
func (r *ForceNewRule) Link() string {
    return "https://example.com/rules/force_new_location"
}

// Check executes the rule.
func (r *ForceNewRule) Check(runner tflint.Runner) error {
    // Define the schema for attributes we want to extract
    schema := &hclext.BodySchema{
        Attributes: []hclext.AttributeSchema{
            {Name: "location", Required: false},
        },
    }

    // Get resources from OLD configuration
    oldContent, err := runner.GetOldResourceContent("myprovider_resource", schema, nil)
    if err != nil {
        return err
    }

    // Get resources from NEW configuration
    newContent, err := runner.GetNewResourceContent("myprovider_resource", schema, nil)
    if err != nil {
        return err
    }

    // Build a map of old resources by name
    oldResources := make(map[string]*hclext.Block)
    for _, block := range oldContent.Blocks {
        if len(block.Labels) >= 2 {
            oldResources[block.Labels[1]] = block
        }
    }

    // Compare each new resource with its old version
    for _, newBlock := range newContent.Blocks {
        if len(newBlock.Labels) < 2 {
            continue
        }
        resourceName := newBlock.Labels[1]

        oldBlock, exists := oldResources[resourceName]
        if !exists {
            continue // New resource, not a breaking change
        }

        // Compare the "location" attribute
        if err := r.compareAttribute(runner, "location", oldBlock, newBlock); err != nil {
            return err
        }
    }

    return nil
}

func (r *ForceNewRule) compareAttribute(runner tflint.Runner, attrName string, oldBlock, newBlock *hclext.Block) error {
    oldAttr := oldBlock.Body.Attributes[attrName]
    newAttr := newBlock.Body.Attributes[attrName]

    if oldAttr == nil || newAttr == nil {
        return nil // Attribute added or removed
    }

    // Evaluate expressions to get values
    oldVal, _ := oldAttr.Expr.Value(nil)
    newVal, _ := newAttr.Expr.Value(nil)

    // Compare values
    if !oldVal.RawEquals(newVal) {
        return runner.EmitIssue(
            r,
            fmt.Sprintf("%s: ForceNew attribute changed from %s to %s",
                attrName, formatValue(oldVal), formatValue(newVal)),
            newAttr.Range,
        )
    }

    return nil
}

func formatValue(v cty.Value) string {
    if v.Type() == cty.String {
        return v.AsString()
    }
    return v.GoString()
}
```

## Building the Plugin

Build your plugin as a standalone binary:

```bash
go build -o tfbreak-ruleset-myprovider
```

The binary name should follow the pattern `tfbreak-ruleset-<name>` where `<name>` matches the ruleset name.

## Testing Your Plugin

### Unit Testing with TestRunner

The `helper.TestRunner` allows testing rules in isolation without running tfbreak-core:

```go
// rules/force_new_test.go
package rules

import (
    "testing"

    "github.com/jokarl/tfbreak-plugin-sdk/helper"
)

func TestForceNewRule_LocationChanged(t *testing.T) {
    // Create a test runner with old and new configurations
    runner := helper.TestRunner(t,
        // Old configuration (baseline)
        map[string]string{
            "main.tf": `
resource "myprovider_resource" "test" {
    location = "westus"
}
`,
        },
        // New configuration (changed)
        map[string]string{
            "main.tf": `
resource "myprovider_resource" "test" {
    location = "eastus"
}
`,
        },
    )

    // Run the rule
    rule := &ForceNewRule{}
    if err := rule.Check(runner); err != nil {
        t.Fatal(err)
    }

    // Assert expected issues
    helper.AssertIssues(t, helper.Issues{
        {
            Rule:    rule,
            Message: "location: ForceNew attribute changed from westus to eastus",
        },
    }, runner.Issues)
}

func TestForceNewRule_NoChange(t *testing.T) {
    runner := helper.TestRunner(t,
        map[string]string{
            "main.tf": `resource "myprovider_resource" "test" { location = "westus" }`,
        },
        map[string]string{
            "main.tf": `resource "myprovider_resource" "test" { location = "westus" }`,
        },
    )

    rule := &ForceNewRule{}
    if err := rule.Check(runner); err != nil {
        t.Fatal(err)
    }

    helper.AssertNoIssues(t, runner.Issues)
}
```

Run tests:

```bash
go test ./...
```

### Table-Driven Tests

For comprehensive coverage, use table-driven tests:

```go
func TestForceNewRule(t *testing.T) {
    tests := []struct {
        name   string
        old    string
        new    string
        issues helper.Issues
    }{
        {
            name: "location changed",
            old:  `resource "myprovider_resource" "test" { location = "westus" }`,
            new:  `resource "myprovider_resource" "test" { location = "eastus" }`,
            issues: helper.Issues{
                {Rule: &ForceNewRule{}, Message: "location: ForceNew attribute changed from westus to eastus"},
            },
        },
        {
            name:   "no change",
            old:    `resource "myprovider_resource" "test" { location = "westus" }`,
            new:    `resource "myprovider_resource" "test" { location = "westus" }`,
            issues: helper.Issues{},
        },
        {
            name:   "new resource",
            old:    ``,
            new:    `resource "myprovider_resource" "test" { location = "westus" }`,
            issues: helper.Issues{},
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            oldFiles := map[string]string{}
            if tt.old != "" {
                oldFiles["main.tf"] = tt.old
            }
            newFiles := map[string]string{}
            if tt.new != "" {
                newFiles["main.tf"] = tt.new
            }

            runner := helper.TestRunner(t, oldFiles, newFiles)
            rule := &ForceNewRule{}

            if err := rule.Check(runner); err != nil {
                t.Fatal(err)
            }

            helper.AssertIssuesWithoutRange(t, tt.issues, runner.Issues)
        })
    }
}
```

## Next Steps

- Read [Interfaces](interfaces.md) for detailed interface documentation
- See [HCL Extensions](hclext.md) for working with complex schemas
- Check [Testing](testing.md) for advanced testing patterns
- Review [Deviation from tflint](deviation-from-tflint.md) if coming from tflint plugin development
