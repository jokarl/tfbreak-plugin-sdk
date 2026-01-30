---
status: approved
date: 2026-01-30
decision-makers: [jokarl]
consulted: [tflint-plugin-sdk maintainers (via documentation)]
informed: [tfbreak-ruleset-azurerm consumers]
---

# tflint-Aligned Plugin SDK Architecture for tfbreak

## Context and Problem Statement

tfbreak is a Terraform breaking change detection tool that analyzes differences between two Terraform configurations (old and new) to identify changes that would cause resource recreation or other breaking changes. To support provider-specific detection logic (e.g., Azure-specific ForceNew attributes), tfbreak needs a plugin system.

The question is: How should the tfbreak plugin SDK be designed to enable plugin authors to write provider-specific rules while maintaining consistency with the Terraform ecosystem?

## Decision Drivers

* **Ecosystem familiarity**: Plugin authors should find the SDK familiar if they've written tflint plugins
* **Dual-config model**: Unlike tflint (single config), tfbreak compares old vs new configurations
* **Testing in isolation**: Plugins must be testable without running tfbreak-core
* **Minimal surface area**: SDK should provide only what's needed, not expose unnecessary complexity
* **Type safety**: Interface types should leverage Go's type system for correctness
* **gRPC compatibility**: Types must be serializable for future gRPC plugin communication

## Considered Options

* **Option 1**: Fork tflint-plugin-sdk and modify for dual-config
* **Option 2**: Create a new SDK loosely inspired by tflint-plugin-sdk
* **Option 3**: Create a tflint-aligned SDK with minimal deviations for dual-config support

## Decision Outcome

Chosen option: **Option 3 - tflint-aligned SDK with minimal deviations**, because it provides maximum familiarity for Terraform ecosystem developers while accommodating tfbreak's unique dual-config requirement.

### Consequences

* Good, because plugin authors familiar with tflint will find the API intuitive
* Good, because the SDK can evolve independently while maintaining conceptual alignment
* Good, because testing utilities mirror tflint's helper package patterns
* Bad, because some tflint features (expression evaluation, autofix) are not needed and won't be implemented
* Neutral, because the dual-config Runner interface is a deliberate deviation that must be documented

### Confirmation

The implementation will be confirmed by:
1. tfbreak-ruleset-azurerm successfully importing and using the SDK
2. All plugin tests passing using the helper.TestRunner
3. The SDK building and passing vet/lint checks independently

## Pros and Cons of the Options

### Option 1: Fork tflint-plugin-sdk

Fork the tflint-plugin-sdk repository and modify it for tfbreak's needs.

* Good, because it starts with a proven, production-tested codebase
* Good, because it includes all gRPC infrastructure already implemented
* Bad, because it brings significant complexity that tfbreak doesn't need (autofix, expression evaluation, module expansion)
* Bad, because maintaining a fork creates long-term burden
* Bad, because the single-config assumption is deeply embedded

### Option 2: New SDK loosely inspired by tflint

Create a completely new SDK with different interface names and patterns.

* Good, because it can be designed specifically for tfbreak's needs
* Good, because there's no legacy baggage
* Bad, because plugin authors must learn a completely new API
* Bad, because it loses the benefits of ecosystem familiarity
* Bad, because design decisions must be made from scratch

### Option 3: tflint-aligned SDK with minimal deviations

Create a new SDK that mirrors tflint-plugin-sdk's interface names and patterns, but with minimal implementations and deliberate deviations for dual-config support.

* Good, because interface names (RuleSet, Rule, Runner, Severity) are familiar
* Good, because the helper package follows tflint's TestRunner pattern
* Good, because only necessary features are implemented
* Good, because dual-config methods (GetOldX/GetNewX) are explicit deviations
* Neutral, because some tflint methods are not implemented (marked as N/A)

## More Information

### Architecture Comparison: tflint-plugin-sdk vs tfbreak-plugin-sdk

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                           tflint-plugin-sdk                                 │
├─────────────────────────────────────────────────────────────────────────────┤
│  tflint/                                                                    │
│  ├── interface.go      # RuleSet, Rule, Runner interfaces                   │
│  ├── ruleset.go        # BuiltinRuleSet implementation                      │
│  └── severity.go       # ERROR, WARNING, NOTICE                             │
│                                                                             │
│  hclext/               # Extended HCL utilities                             │
│  ├── bodycontent.go    # BodySchema, BodyContent                            │
│  └── expression.go     # Expression helpers                                 │
│                                                                             │
│  helper/               # Testing utilities                                  │
│  ├── runner.go         # TestRunner mock implementation                     │
│  └── issue.go          # Issue struct and assertions                        │
│                                                                             │
│  plugin/               # gRPC infrastructure                                │
│  ├── internal/proto/   # Protocol buffer definitions                        │
│  ├── internal/host2plugin/  # Host->Plugin gRPC                             │
│  └── internal/plugin2host/  # Plugin->Host gRPC                             │
└─────────────────────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────────────────────┐
│                          tfbreak-plugin-sdk                                 │
├─────────────────────────────────────────────────────────────────────────────┤
│  tflint/               # tflint-aligned interfaces (SAME NAMES)             │
│  ├── interface.go      # RuleSet, Rule interfaces                           │
│  ├── runner.go         # Runner with GetOld*/GetNew* (DEVIATION)            │
│  ├── ruleset.go        # BuiltinRuleSet, DefaultRule                        │
│  └── severity.go       # ERROR, WARNING, NOTICE                             │
│                                                                             │
│  hclext/               # Minimal HCL extensions                             │
│  └── bodycontent.go    # BodySchema, BodyContent                            │
│                                                                             │
│  helper/               # Testing utilities (CRITICAL)                       │
│  ├── runner.go         # TestRunner(t, oldFiles, newFiles)                  │
│  └── issue.go          # Issue, AssertIssues                                │
│                                                                             │
│  plugin/               # gRPC infrastructure (FUTURE)                       │
│  └── serve.go          # Serve() entry point                                │
└─────────────────────────────────────────────────────────────────────────────┘
```

### Key Interface Alignment

| tflint-plugin-sdk | tfbreak-plugin-sdk | Notes |
|-------------------|-------------------|-------|
| `tflint.RuleSet` | `tflint.RuleSet` | Same interface name, similar methods |
| `tflint.Rule` | `tflint.Rule` | Identical interface |
| `tflint.Runner` | `tflint.Runner` | **DEVIATION**: GetOld*/GetNew* instead of single Get* |
| `tflint.Severity` | `tflint.Severity` | Identical (ERROR, WARNING, NOTICE) |
| `tflint.BuiltinRuleSet` | `tflint.BuiltinRuleSet` | Similar embedding pattern |
| `tflint.DefaultRule` | `tflint.DefaultRule` | Identical |
| `hclext.BodySchema` | `hclext.BodySchema` | Identical structure |
| `hclext.BodyContent` | `hclext.BodyContent` | Identical structure |
| `helper.TestRunner` | `helper.TestRunner` | **DEVIATION**: Takes old+new files |
| `helper.AssertIssues` | `helper.AssertIssues` | Identical |
| `plugin.Serve` | `plugin.Serve` | Same entry point pattern |

### Runner Interface Deviation (Critical)

The fundamental difference between tflint and tfbreak is:
- **tflint**: Analyzes a single Terraform configuration
- **tfbreak**: Compares two configurations (old vs new)

**tflint Runner**:
```go
type Runner interface {
    GetModuleContent(schema *hclext.BodySchema, opts *GetModuleContentOption) (*hclext.BodyContent, error)
    GetResourceContent(resourceType string, schema *hclext.BodySchema, opts *GetModuleContentOption) (*hclext.BodyContent, error)
    EmitIssue(rule Rule, message string, issueRange hcl.Range) error
    // ... other methods
}
```

**tfbreak Runner** (deviation):
```go
type Runner interface {
    // Old configuration access
    GetOldModuleContent(schema *hclext.BodySchema, opts *GetModuleContentOption) (*hclext.BodyContent, error)
    GetOldResourceContent(resourceType string, schema *hclext.BodySchema, opts *GetModuleContentOption) (*hclext.BodyContent, error)

    // New configuration access
    GetNewModuleContent(schema *hclext.BodySchema, opts *GetModuleContentOption) (*hclext.BodyContent, error)
    GetNewResourceContent(resourceType string, schema *hclext.BodySchema, opts *GetModuleContentOption) (*hclext.BodyContent, error)

    // Issue reporting (same as tflint)
    EmitIssue(rule Rule, message string, issueRange hcl.Range) error
}
```

### Features NOT Implemented (Intentional Omissions)

The following tflint-plugin-sdk features are intentionally not implemented in tfbreak-plugin-sdk:

| Feature | Reason for Omission |
|---------|---------------------|
| `EvaluateExpr` | tfbreak compares static configs, doesn't evaluate expressions |
| `EmitIssueWithFix` / `Fixer` | Autofix not needed for breaking change detection |
| `WalkExpressions` | Not needed for attribute comparison |
| Module expansion | tfbreak works with flat configs |
| Provider content | Not needed for breaking change detection |
| Dynamic block expansion | tfbreak compares configs as-written |

### Testing Pattern (helper.TestRunner)

**tflint TestRunner** (single config):
```go
runner := helper.TestRunner(t, map[string]string{
    "main.tf": `resource "aws_instance" "foo" { ... }`,
})
```

**tfbreak TestRunner** (dual config):
```go
runner := helper.TestRunner(t,
    // Old configuration
    map[string]string{
        "main.tf": `resource "azurerm_resource_group" "rg" { location = "westeurope" }`,
    },
    // New configuration
    map[string]string{
        "main.tf": `resource "azurerm_resource_group" "rg" { location = "eastus" }`,
    },
)
```

### Package Import Paths

Plugins will import the SDK as:
```go
import (
    "github.com/jokarl/tfbreak-plugin-sdk/tflint"
    "github.com/jokarl/tfbreak-plugin-sdk/hclext"
    "github.com/jokarl/tfbreak-plugin-sdk/helper"
    "github.com/jokarl/tfbreak-plugin-sdk/plugin"
)
```

### Dependencies

The SDK has minimal dependencies:
- `github.com/hashicorp/hcl/v2` - HCL parsing and types
- `github.com/google/go-cmp` - Testing comparisons (helper package only)

Future gRPC support will add:
- `github.com/hashicorp/go-plugin` - Plugin framework
- `google.golang.org/grpc` - gRPC communication
- `google.golang.org/protobuf` - Protocol buffers
