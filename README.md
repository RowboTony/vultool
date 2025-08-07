# vultool

[![CI](https://github.com/rowbotony/vultool/workflows/CI/badge.svg)](https://github.com/rowbotony/vultool/actions/workflows/ci.yml) [![Security](https://github.com/rowbotony/vultool/workflows/Security/badge.svg)](https://github.com/rowbotony/vultool/actions/workflows/security.yml) [![Go Version](https://img.shields.io/github/go-mod/go-version/rowbotony/vultool)](https://github.com/rowbotony/vultool/blob/main/go.mod) [![GitHub tag (latest by date)](https://img.shields.io/github/v/tag/rowbotony/vultool)](https://github.com/rowbotony/vultool/releases) [![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](LICENSE)

--- 

**A standalone cross-platform Go CLI tool for managing `.vult` vault file operations, compatible with Vultisig security models.**

**[Download Latest Release](https://github.com/rowbotony/vultool/releases/latest)** | **[Documentation](docs/)** | **[Contributing](CONTRIBUTING.md)** | **[Implementation Status](docs/IMPLEMENTATION_STATUS.md)**

## Latest Updates: GG20 Recovery with Centralized Address Validation COMPLETE

**Major Milestone in v0.2.0-dev - August 2025:**
- **Centralized Address Derivation**: Recovery now uses same logic as `list-addresses` for 100% consistency
- **17-Chain Support**: Full recovery for Bitcoin, Ethereum, and all 17+ supported blockchain addresses
- **Automatic Validation**: Every recovery is validated against expected vault addresses
- **Perfect Address Match**: 17/17 chains pass validation (100% success rate)
- **Chain Name Consistency**: Standardized naming across all commands and functions
- **GG20 TSS Recovery**: Full implementation with mathematically correct Lagrange interpolation
- **Multi-Algorithm Support**: Both ECDSA and EdDSA key reconstruction working
- **Production Ready**: All recovered addresses match exactly what users see in Vultisig UI

**[See GG20 Recovery Status »](docs/analysis/GG20_RECOVERY_STATUS.md)** | **[Technical Documentation »](docs/medic-milestone-implementation.md)**

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

Vultool is a standalone command-line interface that focuses specifically on `.vult` file operations. It provides a clean, focused tool for inspecting, validating, and working with Vultisig vault files.

## Features

- **Vault Inspection**: Parse and display vault metadata
- **Validation**: Comprehensive vault file validation
- **Export**: Export vault data to JSON and YAML formats
- **Vault Comparison**: Compare two vault files with detailed diff output
- **TSS Key Recovery**: Reconstruct private keys from threshold shares (NEW!)
- **Multi-Chain Support**: Bitcoin, Ethereum, Solana, THORChain key formats
- **Command Aliases**: Quick shortcuts for common operations
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

# Linux/macOS - using make
make build

# Linux/macOS - using build script
./build.sh

# Windows - using PowerShell
.\build.ps1

# Or build manually with version injection
# Linux/macOS:
go build -ldflags "-X main.version=$(cat VERSION)" -o vultool ./cmd/vultool

# Windows PowerShell:
go build -ldflags "-X main.version=$(Get-Content VERSION)" -o vultool.exe ./cmd/vultool

# Verify installation
vultool --version
```

## Usage

### Quick Commands (New in v0.0.7+)

```bash
# Show concise vault information (alias for inspect --summary)
vultool info -f path/to/vault.vult

# Decode vault to JSON or YAML format
vultool decode -f path/to/vault.vult
vultool decode -f path/to/vault.vult --yaml

# Verify vault integrity (alias for inspect --validate)
vultool verify -f path/to/vault.vult

# Compare two vault files (New in v0.0.8)
vultool diff vault1.vult vault2.vult
```

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

# Or use the quick alias
vultool verify -f path/to/vault.vult
```

### Export

```bash
# Export vault metadata to JSON
vultool inspect -f path/to/vault.vult --export output.json

# Output JSON directly to stdout
vultool decode -f path/to/vault.vult

# Output YAML directly to stdout
vultool decode -f path/to/vault.vult --yaml
```

### Comparison

```bash
# Compare two vault files
vultool diff vault1.vult vault2.vult

# Compare with password support for encrypted vaults
vultool diff --password mypass vault1.vult vault2.vult
```

### Address Discovery

```bash
# List all blockchain addresses derived from vault
vultool list-addresses --vault path/to/vault.vult

# Show addresses in JSON format
vultool list-addresses --vault path/to/vault.vult --json

# Filter specific chains only
vultool list-addresses --vault path/to/vault.vult --chains Bitcoin,Ethereum

# Export to CSV format
vultool list-addresses --vault path/to/vault.vult --csv
```

### Derivation Path Analysis

```bash
# Show common derivation paths for supported chains
vultool list-paths

# Show paths in JSON format for integration
vultool list-paths --json

# Filter paths for specific chains
vultool list-paths --chains bitcoin,ethereum,solana
```

### TSS Key Recovery (Complete in v0.2.0!)

```bash
# Recover private keys from threshold shares (with automatic validation)
vultool recover share1.vult share2.vult --threshold 2

# Recovery automatically validates all 17+ chains against list-addresses:
# bitcoin address validation passed: bc1qvn203p8pp30fk945eywrjey937qpaanha8hc4r
# ethereum address validation passed: 0x55a7ea16a40f8c908cbc935d229ebe4c6658e90d
# GG20 recovery validation passed - all 17 addresses match list-addresses

# Recover only specific blockchain keys
vultool recover share*.vult --threshold 2 --chain bitcoin

# Export recovered keys to JSON file with full validation
vultool recover share*.vult --threshold 2 --output keys.json

# Recover with password-protected vaults
vultool recover encrypted*.vult --threshold 2 --password mypassword
```

**Recovery Features:**
- **100% Address Accuracy**: All recovered addresses match exactly what `list-addresses` shows
- **Automatic Validation**: Every recovery is validated against expected vault addresses
- **17+ Chain Support**: Bitcoin, Ethereum, BSC, Avalanche, Polygon, Arbitrum, Optimism, Base, Blast, ZkSync, THORChain, Litecoin, Dogecoin, Dash, Zcash, Bitcoin Cash, Solana, SUI
- **Wallet Import Formats**: Solana JSON array format, SUI base64 format, comprehensive wallet compatibility documentation
- **TSS Wallet Integration**: See [TSS Wallet Limitations Guide](docs/TSS_WALLET_LIMITATIONS.md) for wallet import workflows and solutions
- **Centralized Logic**: Uses same address derivation as `list-addresses` for perfect consistency

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

# Install dependencies
go mod tidy

# Install development tools (required for linting and CI)
curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin v1.61.0

# Ensure Go bin directory is in your PATH (add to ~/.bashrc, ~/.zshrc, etc.)
export PATH=$PATH:$(go env GOPATH)/bin

# Build using make (Linux/macOS)
make build

# Build using scripts
./build.sh              # Linux/macOS
.\build.ps1            # Windows PowerShell

# Build manually
# Linux/macOS:
go build -ldflags "-X main.version=$(cat VERSION)" -o vultool ./cmd/vultool

# Windows PowerShell:
go build -ldflags "-X main.version=$(Get-Content VERSION)" -o vultool.exe ./cmd/vultool

# Test with shared fixtures
./vultool inspect -f test/fixtures/testGG20-part1of2.vult --summary

# Run tests
go test ./...
```

### CI/CD

Robust automated pipeline with multi-platform builds and comprehensive quality gates:

- **Platforms**: Linux (amd64/arm64), macOS (amd64/arm64), Windows (amd64)
- **Security**: gosec, govulncheck, CodeQL, automated dependency updates
- **Quality**: Tests across Go 1.21-1.23, comprehensive linting, automated fixture validation
- **Reliability**: Zero-intervention CI with encrypted test file handling

```bash
# Run full CI locally (includes all checks)
make ci-local

# Individual components
make lint          # Code quality checks
make security-scan # Security vulnerability scanning  
make validate      # Test fixture validation (handles encrypted files automatically)

# Create release (automated binary builds)
echo "1.0.0" > VERSION && git tag v1.0.0 && git push --tags
```

**New in v0.1.0:** CI pipeline now handles encrypted test fixtures automatically without hanging, ensuring reliable automated builds and deployments.

See [`docs/CI-CD.md`](docs/CI-CD.md) for details.


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
