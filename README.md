# tfbreak-plugin-sdk

[![Go Reference](https://pkg.go.dev/badge/github.com/jokarl/tfbreak-plugin-sdk.svg)](https://pkg.go.dev/github.com/jokarl/tfbreak-plugin-sdk)

The official SDK for building [tfbreak](https://github.com/jokarl/tfbreak) plugins. This SDK provides the interfaces, types, and testing utilities needed to create provider-specific breaking change detection rules.

## Overview

tfbreak is a Terraform breaking change detection tool that analyzes differences between two Terraform configurations (old and new) to identify changes that would cause resource recreation or other breaking changes. This SDK enables plugin authors to write provider-specific detection rules (e.g., Azure-specific ForceNew attributes).

### Key Feature: Dual-Config Model

Unlike tflint (which analyzes a single configuration), tfbreak compares two configurations to detect breaking changes. This fundamental difference is reflected in the SDK's Runner interface:

```go
// tflint: single configuration
runner.GetResourceContent("aws_instance", schema, nil)

// tfbreak: dual configuration (old vs new)
runner.GetOldResourceContent("azurerm_resource_group", schema, nil)  // Baseline
runner.GetNewResourceContent("azurerm_resource_group", schema, nil)  // Changed
```

See [docs/deviation-from-tflint.md](docs/deviation-from-tflint.md) for a complete comparison.

## Requirements

- Go 1.23 or later
- tfbreak >= 0.1.0 (when using plugins with tfbreak-core)

## Installation

```bash
go get github.com/jokarl/tfbreak-plugin-sdk
```

## Quick Start

### 1. Create a new plugin

```go
// main.go
package main

import (
    "github.com/jokarl/tfbreak-plugin-sdk/plugin"
    "github.com/jokarl/tfbreak-plugin-sdk/tflint"
)

func main() {
    plugin.Serve(&plugin.ServeOpts{
        RuleSet: &MyRuleSet{
            BuiltinRuleSet: tflint.BuiltinRuleSet{
                Name:    "myprovider",
                Version: "0.1.0",
                Rules:   []tflint.Rule{&MyRule{}},
            },
        },
    })
}

type MyRuleSet struct {
    tflint.BuiltinRuleSet
}
```

### 2. Implement a rule

```go
// rules/my_rule.go
package rules

import (
    "github.com/jokarl/tfbreak-plugin-sdk/hclext"
    "github.com/jokarl/tfbreak-plugin-sdk/tflint"
)

type MyRule struct {
    tflint.DefaultRule
}

func (r *MyRule) Name() string { return "my_provider_force_new" }
func (r *MyRule) Link() string { return "https://example.com/rules/my_rule" }

func (r *MyRule) Check(runner tflint.Runner) error {
    schema := &hclext.BodySchema{
        Attributes: []hclext.AttributeSchema{
            {Name: "location", Required: false},
        },
    }

    oldContent, err := runner.GetOldResourceContent("my_resource", schema, nil)
    if err != nil {
        return err
    }
    newContent, err := runner.GetNewResourceContent("my_resource", schema, nil)
    if err != nil {
        return err
    }

    // Compare old and new, emit issues for breaking changes
    // ...

    return nil
}
```

### 3. Test your rule

```go
// rules/my_rule_test.go
package rules

import (
    "testing"
    "github.com/jokarl/tfbreak-plugin-sdk/helper"
)

func TestMyRule(t *testing.T) {
    runner := helper.TestRunner(t,
        map[string]string{
            "main.tf": `resource "my_resource" "test" { location = "westus" }`,
        },
        map[string]string{
            "main.tf": `resource "my_resource" "test" { location = "eastus" }`,
        },
    )

    rule := &MyRule{}
    if err := rule.Check(runner); err != nil {
        t.Fatal(err)
    }

    helper.AssertIssues(t, helper.Issues{
        {Rule: rule, Message: "location: ForceNew attribute changed"},
    }, runner.Issues)
}
```

## Architecture

The SDK follows tflint-plugin-sdk's architecture for ecosystem familiarity:

```
tfbreak-plugin-sdk/
├── tflint/           # Core interfaces (tflint-aligned naming)
│   ├── interface.go  # RuleSet, Rule interfaces
│   ├── runner.go     # Runner with GetOld*/GetNew* (deviation)
│   ├── ruleset.go    # BuiltinRuleSet, DefaultRule
│   ├── severity.go   # ERROR, WARNING, NOTICE
│   └── config.go     # Config, RuleConfig
├── hclext/           # HCL extension types
│   └── bodycontent.go
├── helper/           # Testing utilities
│   ├── runner.go     # TestRunner (dual-config)
│   └── issue.go      # Issue, AssertIssues
└── plugin/           # Plugin entry point
    └── serve.go      # Serve()
```

## Documentation

- [Getting Started](docs/getting-started.md) - Create your first plugin
- [Interfaces](docs/interfaces.md) - Core interfaces reference
- [HCL Extensions](docs/hclext.md) - Working with HCL types
- [Testing](docs/testing.md) - Testing your plugins
- [Deviation from tflint](docs/deviation-from-tflint.md) - Key differences from tflint-plugin-sdk
- [ADR-0001](docs/adr/ADR-0001-tflint-aligned-plugin-sdk.md) - Architecture decision record

## API Reference

Full API documentation is available at [pkg.go.dev/github.com/jokarl/tfbreak-plugin-sdk](https://pkg.go.dev/github.com/jokarl/tfbreak-plugin-sdk).

## Development

### Prerequisites

- Go 1.23+
- Make (optional)

### Building

```bash
go build ./...
```

### Testing

```bash
go test ./...
```

### Linting

```bash
go vet ./...
```

## Related Projects

- [tfbreak](https://github.com/jokarl/tfbreak) - The tfbreak CLI tool
- [tfbreak-ruleset-azurerm](https://github.com/jokarl/tfbreak-ruleset-azurerm) - Azure provider rules
- [tflint-plugin-sdk](https://github.com/terraform-linters/tflint-plugin-sdk) - Inspiration for this SDK

## License

MIT License - see [LICENSE](LICENSE) for details.
