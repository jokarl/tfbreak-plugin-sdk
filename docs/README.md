# tfbreak-plugin-sdk Documentation

This directory contains comprehensive documentation for the tfbreak-plugin-sdk.

## Contents

### Getting Started

- [Getting Started](getting-started.md) - Create your first tfbreak plugin, understand project structure, and learn how to build and test plugins.

### Core Concepts

- [Interfaces](interfaces.md) - Detailed reference for core interfaces: `RuleSet`, `Rule`, `Runner`, and helper structs like `DefaultRule` and `BuiltinRuleSet`.

- [HCL Extensions](hclext.md) - Working with HCL types: `BodySchema`, `BodyContent`, `Attribute`, `Block`, and conversion functions.

- [Testing](testing.md) - Testing guide covering `helper.TestRunner`, assertion functions, and writing comprehensive test cases.

### Architecture

- [Deviation from tflint](deviation-from-tflint.md) - Key differences between tfbreak-plugin-sdk and tflint-plugin-sdk, focusing on the dual-config model.

### Architecture Decision Records (ADR)

- [ADR-0001: tflint-Aligned Plugin SDK](adr/ADR-0001-tflint-aligned-plugin-sdk.md) - The foundational architecture decision explaining why and how tfbreak-plugin-sdk aligns with tflint-plugin-sdk.

### Change Records (CR)

- [CR-0001: Minimum Viable SDK](cr/CR-0001-minimum-viable-sdk.md)
- [CR-0002: Core Types](cr/CR-0002-core-types.md)
- [CR-0003: hclext Types](cr/CR-0003-hclext-types.md)
- [CR-0004: Rule Runner Interfaces](cr/CR-0004-rule-runner-interfaces.md)
- [CR-0005: RuleSet Builtin](cr/CR-0005-ruleset-builtin.md)
- [CR-0006: Helper Package](cr/CR-0006-helper-package.md)
- [CR-0007: Plugin Serve](cr/CR-0007-plugin-serve.md)

## Quick Navigation

| I want to... | Read this |
|--------------|-----------|
| Create a new plugin | [Getting Started](getting-started.md) |
| Understand the Rule interface | [Interfaces](interfaces.md#rule-interface) |
| Access old/new configurations | [Interfaces](interfaces.md#runner-interface) |
| Define attribute schemas | [HCL Extensions](hclext.md#bodyschema) |
| Write tests for my rules | [Testing](testing.md) |
| Understand dual-config vs single-config | [Deviation from tflint](deviation-from-tflint.md) |

## External Resources

- [pkg.go.dev/github.com/jokarl/tfbreak-plugin-sdk](https://pkg.go.dev/github.com/jokarl/tfbreak-plugin-sdk) - API reference documentation
- [tflint-plugin-sdk](https://github.com/terraform-linters/tflint-plugin-sdk) - The SDK that inspired this project
- [hashicorp/hcl](https://github.com/hashicorp/hcl) - HCL library documentation
