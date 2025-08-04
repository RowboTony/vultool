# vultool Documentation

Welcome to the vultool documentation! This directory contains comprehensive guides for using, developing, and contributing to vultool.

## Documentation Index

### For Users
- **[Main README](../README.md)** - Project overview, installation, and basic usage
- **[Usage Examples](../README.md#usage)** - CLI commands and examples
- **[Library Usage](../README.md#library-usage)** - Using vultool as a Go library

### For Contributors
- **[Contributing Guide](../CONTRIBUTING.md)** - How to contribute to the project
- **[CI/CD Pipeline](CI-CD.md)** - Build automation, testing, and releases
- **[Development Workflow](../CONTRIBUTING.md#development-workflow)** - Local development setup

### For Maintainers
- **[Security Scanning](CI-CD.md#security-scanning)** - Security analysis and vulnerability management
- **[Release Process](CI-CD.md#creating-releases)** - How to create and manage releases
- **[Quality Gates](CI-CD.md#quality-gates)** - Requirements for merging changes

## Quick Navigation

| I want to... | Go to |
|--------------|-------|
| **Install and use vultool** | [Installation Guide](../README.md#installation) |
| **Contribute code** | [Contributing Guide](../CONTRIBUTING.md) |
| **Report a bug** | [GitHub Issues](https://github.com/rowbotony/vultool/issues) |
| **Understand the CI/CD** | [CI/CD Pipeline](CI-CD.md) |
| **See recent changes** | [CHANGELOG](../CHANGELOG.md) |

## Project Architecture

```
vultool/
├── cmd/vultool/          # CLI entry point and commands
├── internal/vault/       # Core vault parsing logic (private)
├── pkg/client/           # Public API for library usage
├── test/fixtures/        # Test vault files (git submodule)
├── docs/                 # Documentation (you are here!)
└── .github/workflows/    # CI/CD automation
```

## Security

For security-related questions or to report vulnerabilities:
- **Public questions**: Use [GitHub Discussions](https://github.com/rowbotony/vultool/discussions)
- **Security issues**: Follow our [Security Policy](../CONTRIBUTING.md#security)

## External Resources

- **[Vultisig Project](https://github.com/vultisig)** - The broader Vultisig ecosystem
- **[Go Documentation](https://golang.org/doc/)** - Official Go language documentation
- **[Cobra CLI Framework](https://cobra.dev/)** - CLI framework used by vultool

---

**Need help?** Don't hesitate to [open an issue](https://github.com/rowbotony/vultool/issues) or [start a discussion](https://github.com/rowbotony/vultool/discussions)!
