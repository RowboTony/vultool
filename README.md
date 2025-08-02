# vultool

[![CI](https://github.com/rowbotony/vultool/workflows/CI/badge.svg)](https://github.com/rowbotony/vultool/actions/workflows/ci.yml)
[![Security](https://github.com/rowbotony/vultool/workflows/Security/badge.svg)](https://github.com/rowbotony/vultool/actions/workflows/security.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/rowbotony/vultool)](https://goreportcard.com/report/github.com/rowbotony/vultool)
[![Release](https://img.shields.io/github/v/release/rowbotony/vultool)](https://github.com/rowbotony/vultool/releases/latest)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](LICENSE)

**A standalone cross-platform Go CLI tool for managing `.vult` vault file operations, compatible with Vultisig security models.**

**[Download Latest Release](https://github.com/rowbotony/vultool/releases/latest)** | **[Documentation](docs/)** | **[Contributing](CONTRIBUTING.md)**

## Quickstart

```bash
# Install vultool
go install github.com/rowbotony/vultool/cmd/vultool@latest

# Inspect a vault file (try with our test fixtures)
vultool inspect -f test/fixtures/testGG20-part1of2.vult --summary

# Export vault metadata to JSON
vultool inspect -f your-vault.vult --export vault-metadata.json
```

**Sample Output:**
```
Vault: Test private key vault
File: test/fixtures/testGG20-part1of2.vult
Encrypted: false
Version: 0
Local Party: Pixel 5a-a9b
ECDSA Public Key: 0267db81657a956f364167c3986a426b448a74ac0db2092f6665c4c202b37f6f1d
Key Shares: 2
```

## Overview

Vultool is a command-line interface extracted from vulticore that focuses specifically on `.vult` file operations. It provides a clean, standalone tool for inspecting, validating, and working with Vultisig vault files.

## Features

- **Vault Inspection**: Parse and display vault metadata
- **Validation**: Comprehensive vault file validation
- **Export**: Export vault data to JSON format
- **Encryption Support**: Handle both encrypted and unencrypted vaults
- **Security**: Built-in path validation and safety checks

## Installation

**Requirements:** Go 1.21+ (tested with Go 1.21-1.23)

```bash
# Install from source
go install github.com/rowbotony/vultool/cmd/vultool@latest

# Or build locally
git clone https://github.com/rowbotony/vultool.git
cd vultool
go build -o vultool ./cmd/vultool

# Verify installation
vultool --version
```

## Usage

### Basic Inspection

```bash
# Inspect a vault file (shows summary by default)
vultool inspect -f path/to/vault.vult

# Show detailed summary
vultool inspect -f path/to/vault.vult --summary

# Show key share information
vultool inspect -f path/to/vault.vult --show-keyshares
```

### Validation

```bash
# Validate vault structure
vultool inspect -f path/to/vault.vult --validate
```

### Export

```bash
# Export vault metadata to JSON
vultool inspect -f path/to/vault.vult --export output.json
```

### Encrypted Vaults

```bash
# Interactive password prompt
vultool inspect -f encrypted-vault.vult

# Provide password as parameter
vultool inspect -f encrypted-vault.vult --password mypassword
```

## Library Usage

Vultool can also be used as a Go library:

```go
package main

import (
    "fmt"
    "github.com/rowbotony/vultool/pkg/client"
)

func main() {
    vaultInfo, err := client.ParseVaultFile("path/to/vault.vult")
    if err != nil {
        panic(err)
    }
    
    fmt.Printf("Vault: %s\n", vaultInfo.Name)
    fmt.Printf("Encrypted: %t\n", vaultInfo.IsEncrypted)
}
```

## Architecture

- `cmd/vultool/`: CLI entry point
- `internal/vault/`: Core vault parsing and validation logic
- `pkg/client/`: Public API for library usage

## Dependencies

- **Vultisig commondata**: Official protobuf schemas
- **Cobra**: CLI framework (industry standard)
- **Go crypto libraries**: AES-GCM encryption support

## Development

```bash
# Clone and build
git clone https://github.com/rowbotony/vultool.git
cd vultool

# Initialize submodules (required for test fixtures)
git submodule init
git submodule update

# Install dependencies and build
go mod tidy
go build ./cmd/vultool

# Test with shared fixtures
./vultool inspect -f test/fixtures/testGG20-part1of2.vult --summary

# Run tests
go test ./...
```

### CI/CD

Automated pipeline with multi-platform builds and security scanning:

- **Platforms**: Linux (amd64/arm64), macOS (amd64/arm64), Windows (amd64)
- **Security**: gosec, govulncheck, CodeQL, automated dependency updates
- **Quality**: Tests across Go 1.21-1.23, comprehensive linting

```bash
# Run full CI locally
make ci-local

# Create release (automated binary builds)
echo "1.0.0" > VERSION && git tag v1.0.0 && git push --tags
```

See [`docs/CI-CD.md`](docs/CI-CD.md) for details.

## Relationship to vulticore

Vultool was extracted from the vulticore project to provide a focused, standalone CLI for `.vult` file operations. Vulticore now imports vultool as a dependency:

```go
// In vulticore/go.mod
require github.com/rowbotony/vultool v0.0.6-dev

// For local development:
replace github.com/rowbotony/vultool => ../vultool
```

This separation provides:
- **Clean separation of concerns**
- **Independent versioning**
- **Reusable vault operations**
- **Focused CLI tool**

## Security

Vultool includes several security features:
- Path validation to prevent directory traversal
- Safe output path validation
- Secure handling of encrypted vaults
- Memory-safe cryptographic operations

## License

This project is licensed under the Apache License 2.0 - see the [LICENSE](LICENSE) file for details.

## Contributing

We welcome contributions!

To get started:

1. **Read our [Contributing Guidelines](CONTRIBUTING.md)** for detailed setup and workflow information
2. **Check existing [Issues](https://github.com/rowbotony/vultool/issues)** for tasks that need help

**Quick Start for Contributors:**
```bash
git clone https://github.com/YOUR_USERNAME/vultool.git
cd vultool && git submodule update --init
make ci-local  # Run full test suite
```

For major changes, please open an issue first to discuss what you would like to change.

**New to OSS?** We're happy to help! Look for issues labeled `good first issue`.
