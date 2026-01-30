# Testing Guide

The `helper` package provides testing utilities for tfbreak plugins. Use these to test rules in isolation without running tfbreak-core.

## Overview

| Function/Type | Purpose |
|---------------|---------|
| `TestRunner` | Creates a mock Runner with test configurations |
| `AssertIssues` | Compares expected and actual issues |
| `AssertIssuesWithoutRange` | Compares issues ignoring source ranges |
| `AssertNoIssues` | Verifies no issues were emitted |
| `Issue` | Represents a finding for test assertions |
| `Issues` | Slice of Issue for convenience |

## TestRunner

`TestRunner` creates a mock `tflint.Runner` implementation for testing. It accepts two configuration maps: old (baseline) and new (changed).

### Signature

```go
func TestRunner(t *testing.T, oldFiles, newFiles map[string]string) *Runner
```

### Basic Usage

```go
import (
    "testing"
    "github.com/jokarl/tfbreak-plugin-sdk/helper"
)

func TestMyRule(t *testing.T) {
    runner := helper.TestRunner(t,
        // Old configuration (baseline)
        map[string]string{
            "main.tf": `
resource "azurerm_resource_group" "test" {
    name     = "rg-test"
    location = "westus"
}
`,
        },
        // New configuration (changed)
        map[string]string{
            "main.tf": `
resource "azurerm_resource_group" "test" {
    name     = "rg-test"
    location = "eastus"
}
`,
        },
    )

    rule := &MyRule{}
    if err := rule.Check(runner); err != nil {
        t.Fatal(err)
    }

    // Access issues via runner.Issues
    if len(runner.Issues) != 1 {
        t.Errorf("expected 1 issue, got %d", len(runner.Issues))
    }
}
```

### Multiple Files

```go
runner := helper.TestRunner(t,
    map[string]string{
        "main.tf": `
module "rg" {
    source   = "./modules/rg"
    location = "westus"
}
`,
        "modules/rg/main.tf": `
resource "azurerm_resource_group" "this" {
    name     = var.name
    location = var.location
}
`,
    },
    map[string]string{
        "main.tf": `
module "rg" {
    source   = "./modules/rg"
    location = "eastus"
}
`,
        "modules/rg/main.tf": `
resource "azurerm_resource_group" "this" {
    name     = var.name
    location = var.location
}
`,
    },
)
```

### Empty Configurations

Test scenarios with new or deleted resources:

```go
// New resource (didn't exist before)
runner := helper.TestRunner(t,
    map[string]string{},  // Empty old config
    map[string]string{
        "main.tf": `resource "azurerm_resource_group" "new" { location = "westus" }`,
    },
)

// Deleted resource
runner := helper.TestRunner(t,
    map[string]string{
        "main.tf": `resource "azurerm_resource_group" "old" { location = "westus" }`,
    },
    map[string]string{},  // Empty new config
)
```

## Issue Type

`Issue` represents a finding from a rule for test assertions.

```go
type Issue struct {
    Rule    tflint.Rule  // The rule that emitted the issue
    Message string       // Issue message
    Range   hcl.Range    // Source location
}

type Issues []Issue
```

## AssertIssues

Compares expected and actual issues. It ignores:
- Issue order (sorted before comparison)
- Byte positions in ranges (only compares line/column)

### Signature

```go
func AssertIssues(t *testing.T, want, got Issues)
```

### Usage

```go
func TestMyRule(t *testing.T) {
    runner := helper.TestRunner(t, oldFiles, newFiles)

    rule := &MyRule{}
    rule.Check(runner)

    helper.AssertIssues(t, helper.Issues{
        {
            Rule:    rule,
            Message: "location: ForceNew attribute changed",
            Range: hcl.Range{
                Filename: "main.tf",
                Start:    hcl.Pos{Line: 4, Column: 5},
                End:      hcl.Pos{Line: 4, Column: 23},
            },
        },
    }, runner.Issues)
}
```

### Multiple Issues

```go
helper.AssertIssues(t, helper.Issues{
    {Rule: rule, Message: "location changed"},
    {Rule: rule, Message: "sku changed"},
    {Rule: rule, Message: "tags modified"},
}, runner.Issues)
```

## AssertIssuesWithoutRange

Compares issues ignoring the `Range` field entirely. Use this when exact source locations are not important.

### Signature

```go
func AssertIssuesWithoutRange(t *testing.T, want, got Issues)
```

### Usage

```go
func TestMyRule(t *testing.T) {
    runner := helper.TestRunner(t, oldFiles, newFiles)

    rule := &MyRule{}
    rule.Check(runner)

    // Only check rule and message, ignore Range
    helper.AssertIssuesWithoutRange(t, helper.Issues{
        {Rule: rule, Message: "location: ForceNew attribute changed"},
    }, runner.Issues)
}
```

## AssertNoIssues

Verifies that no issues were emitted. Fails the test if any issues exist.

### Signature

```go
func AssertNoIssues(t *testing.T, got Issues)
```

### Usage

```go
func TestMyRule_NoChange(t *testing.T) {
    runner := helper.TestRunner(t,
        map[string]string{
            "main.tf": `resource "azurerm_resource_group" "test" { location = "westus" }`,
        },
        map[string]string{
            "main.tf": `resource "azurerm_resource_group" "test" { location = "westus" }`,
        },
    )

    rule := &MyRule{}
    rule.Check(runner)

    helper.AssertNoIssues(t, runner.Issues)
}
```

## Table-Driven Tests

Use table-driven tests for comprehensive coverage:

```go
func TestForceNewRule(t *testing.T) {
    rule := &ForceNewRule{}

    tests := []struct {
        name   string
        old    map[string]string
        new    map[string]string
        issues helper.Issues
    }{
        {
            name: "location changed - should detect",
            old: map[string]string{
                "main.tf": `resource "azurerm_resource_group" "test" {
                    name     = "rg-test"
                    location = "westus"
                }`,
            },
            new: map[string]string{
                "main.tf": `resource "azurerm_resource_group" "test" {
                    name     = "rg-test"
                    location = "eastus"
                }`,
            },
            issues: helper.Issues{
                {Rule: rule, Message: "location: ForceNew attribute changed"},
            },
        },
        {
            name: "no change - no issues",
            old: map[string]string{
                "main.tf": `resource "azurerm_resource_group" "test" { location = "westus" }`,
            },
            new: map[string]string{
                "main.tf": `resource "azurerm_resource_group" "test" { location = "westus" }`,
            },
            issues: helper.Issues{},
        },
        {
            name: "new resource - no issues",
            old:  map[string]string{},
            new: map[string]string{
                "main.tf": `resource "azurerm_resource_group" "test" { location = "westus" }`,
            },
            issues: helper.Issues{},
        },
        {
            name: "deleted resource - no issues",
            old: map[string]string{
                "main.tf": `resource "azurerm_resource_group" "test" { location = "westus" }`,
            },
            new:    map[string]string{},
            issues: helper.Issues{},
        },
        {
            name: "name changed - no issues (not ForceNew)",
            old: map[string]string{
                "main.tf": `resource "azurerm_resource_group" "test" {
                    name     = "rg-old"
                    location = "westus"
                }`,
            },
            new: map[string]string{
                "main.tf": `resource "azurerm_resource_group" "test" {
                    name     = "rg-new"
                    location = "westus"
                }`,
            },
            issues: helper.Issues{},
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            runner := helper.TestRunner(t, tt.old, tt.new)

            if err := rule.Check(runner); err != nil {
                t.Fatal(err)
            }

            helper.AssertIssuesWithoutRange(t, tt.issues, runner.Issues)
        })
    }
}
```

## Testing Multiple Resources

```go
func TestMultipleResources(t *testing.T) {
    runner := helper.TestRunner(t,
        map[string]string{
            "main.tf": `
resource "azurerm_resource_group" "rg1" {
    name     = "rg-1"
    location = "westus"
}

resource "azurerm_resource_group" "rg2" {
    name     = "rg-2"
    location = "eastus"
}

resource "azurerm_resource_group" "rg3" {
    name     = "rg-3"
    location = "northeurope"
}
`,
        },
        map[string]string{
            "main.tf": `
resource "azurerm_resource_group" "rg1" {
    name     = "rg-1"
    location = "eastus"
}

resource "azurerm_resource_group" "rg2" {
    name     = "rg-2"
    location = "westus"
}

resource "azurerm_resource_group" "rg3" {
    name     = "rg-3"
    location = "northeurope"
}
`,
        },
    )

    rule := &ForceNewRule{}
    rule.Check(runner)

    // Expect issues for rg1 and rg2 (both changed), but not rg3
    helper.AssertIssuesWithoutRange(t, helper.Issues{
        {Rule: rule, Message: "location: ForceNew attribute changed"},
        {Rule: rule, Message: "location: ForceNew attribute changed"},
    }, runner.Issues)
}
```

## Testing with Nested Blocks

```go
func TestNestedBlocks(t *testing.T) {
    runner := helper.TestRunner(t,
        map[string]string{
            "main.tf": `
resource "azurerm_storage_account" "test" {
    name                     = "storageaccount"
    resource_group_name      = "rg-test"
    location                 = "westus"
    account_tier             = "Standard"
    account_replication_type = "LRS"

    blob_properties {
        versioning_enabled = false
    }
}
`,
        },
        map[string]string{
            "main.tf": `
resource "azurerm_storage_account" "test" {
    name                     = "storageaccount"
    resource_group_name      = "rg-test"
    location                 = "eastus"
    account_tier             = "Premium"
    account_replication_type = "LRS"

    blob_properties {
        versioning_enabled = true
    }
}
`,
        },
    )

    rule := &StorageAccountRule{}
    rule.Check(runner)

    helper.AssertIssuesWithoutRange(t, helper.Issues{
        {Rule: rule, Message: "location: ForceNew attribute changed"},
        {Rule: rule, Message: "account_tier: ForceNew attribute changed"},
    }, runner.Issues)
}
```

## Testing Error Conditions

```go
func TestRuleError(t *testing.T) {
    runner := helper.TestRunner(t,
        map[string]string{
            "main.tf": `resource "azurerm_resource_group" "test" { location = "westus" }`,
        },
        map[string]string{
            "main.tf": `resource "azurerm_resource_group" "test" { location = "eastus" }`,
        },
    )

    rule := &MyRule{}
    err := rule.Check(runner)

    // Test that the rule returns an error in specific conditions
    if err != nil {
        t.Errorf("expected no error, got: %v", err)
    }
}
```

## Best Practices

### 1. Test Both Positive and Negative Cases

```go
// Positive: Issue should be detected
func TestRule_Detects(t *testing.T) { ... }

// Negative: No issue should be emitted
func TestRule_NoIssue(t *testing.T) { ... }
```

### 2. Use Meaningful Test Names

```go
t.Run("location_changed_emits_issue", ...)
t.Run("no_change_no_issue", ...)
t.Run("new_resource_no_issue", ...)
t.Run("deleted_resource_no_issue", ...)
```

### 3. Test Edge Cases

- Empty configurations
- Missing attributes
- New resources
- Deleted resources
- Multiple resources
- Nested blocks

### 4. Use AssertIssuesWithoutRange for Simple Tests

When exact source locations don't matter, use `AssertIssuesWithoutRange` for cleaner tests.

### 5. Keep Test Configurations Minimal

Only include the configuration needed to test the specific scenario.

### 6. Document Complex Test Cases

```go
{
    name: "renamed resource should not trigger ForceNew",
    // When a resource is renamed (different Labels[1]), it's a new resource
    // from tfbreak's perspective. The old resource will be deleted and a new
    // one created, but that's expected behavior for a rename.
    old: map[string]string{...},
    new: map[string]string{...},
    issues: helper.Issues{},
}
```
