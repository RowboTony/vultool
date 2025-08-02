# Contributing to vultool

Thank you for your interest in contributing to vultool! This document provides guidelines and information for contributors.

## Table of Contents

- [Code of Conduct](#code-of-conduct)
- [Getting Started](#getting-started)
- [Development Workflow](#development-workflow)
- [Testing](#testing)
- [Code Style](#code-style)
- [Submitting Changes](#submitting-changes)
- [Security](#security)
- [Community](#community)

## Code of Conduct

This project adheres to a code of conduct that we expect all contributors to follow. Please be respectful and professional in all interactions.

## Getting Started

### Prerequisites

- **Go 1.21+** (tested with Go 1.21-1.23)
- **Git** with submodule support
- **Make** (optional, for development commands)

### Setup Development Environment

```bash
# Fork and clone the repository
git clone https://github.com/YOUR_USERNAME/vultool.git
cd vultool

# Initialize submodules for test fixtures
git submodule init
git submodule update

# Install dependencies
go mod tidy

# Build and test
go build ./cmd/vultool
./vultool --version

# Run tests
go test ./...
```

### Development Philosophy

Vultool follows a **CLI-first development approach**:

1. **Single Source of Truth**: All vault operations should be consistent between CLI and library usage
2. **Security by Default**: Always validate inputs, sanitize outputs, and handle secrets securely
3. **Cross-Platform**: Support Linux, macOS, and Windows equally
4. **Minimal Dependencies**: Keep the dependency tree lean and well-maintained

## Development Workflow

### Project Structure

```
vultool/
├── cmd/vultool/          # CLI entry point and commands
├── internal/vault/       # Core vault parsing logic (private)
├── pkg/client/           # Public API for library usage
├── test/fixtures/        # Test vault files (git submodule)
├── docs/                 # Documentation
└── .github/workflows/    # CI/CD automation
```

### Local Development Commands

```bash
# Run full CI suite locally
make ci-local

# Individual development tasks
make build                # Build for current platform
make test                 # Run all tests
make security-scan        # Run security analysis
make lint                 # Code formatting and linting
make coverage            # Generate test coverage report

# Cross-platform builds
make build-all-platforms  # Build for all supported platforms
```

### Branch Strategy

- `main`: Stable, production-ready code
- `feature/your-feature`: New features and improvements
- `fix/issue-description`: Bug fixes
- `security/cve-fix`: Security patches

## Testing

### Test Categories

1. **Unit Tests**: Test individual functions and components
   ```bash
   go test ./internal/vault -v
   go test ./pkg/client -v
   ```

2. **Integration Tests**: Test CLI commands with real vault files
   ```bash
   go test ./cmd/vultool -v
   ```

3. **Security Tests**: Automated security scanning
   ```bash
   make security-scan
   ```

### Writing Tests

- **Test with real vault files**: Use fixtures from `test/fixtures/`
- **Test error conditions**: Ensure proper error handling
- **Test security boundaries**: Validate input sanitization
- **Cross-platform compatibility**: Consider path separators and file permissions

Example test:
```go
func TestParseVaultFile(t *testing.T) {
    vaultPath := "../../test/fixtures/testGG20-part1of2.vult"
    
    vault, err := ParseVaultFile(vaultPath)
    require.NoError(t, err)
    
    assert.Equal(t, "Test private key vault", vault.Name)
    assert.False(t, vault.IsEncrypted)
    assert.Equal(t, 2, len(vault.KeyShares))
}
```

## Code Style

### Go Code Standards

- **gofmt**: All code must be formatted with `gofmt`
- **golangci-lint**: Follow linting rules in `.golangci.yml`
- **Error handling**: Always handle errors explicitly
- **Documentation**: Document exported functions and types

```go
// ParseVaultFile parses a .vult file and returns vault information.
// It supports both encrypted and unencrypted vault files.
//
// Parameters:
//   - filePath: Path to the .vult file
//
// Returns:
//   - *VaultInfo: Parsed vault information
//   - error: Any parsing or validation errors
func ParseVaultFile(filePath string) (*VaultInfo, error) {
    // Implementation...
}
```

### CLI Design Principles

- **Consistent flags**: Use standard flag patterns (`-f`, `--file`)
- **Clear output**: Human-readable by default, machine-readable with flags
- **Error messages**: Provide actionable error messages
- **Help text**: Comprehensive help for all commands

## Submitting Changes

### Pull Request Process

1. **Create an issue** (for significant changes)
2. **Fork the repository**
3. **Create a feature branch**
4. **Make your changes**
5. **Write/update tests**
6. **Run the full CI suite locally**
7. **Submit a pull request**

### Pull Request Requirements

- [ ] **All tests pass** (`make ci-local`)
- [ ] **Security scans pass** (gosec, govulncheck)
- [ ] **Code is formatted** (gofmt, golangci-lint)
- [ ] **Documentation updated** (if applicable)
- [ ] **CHANGELOG.md updated** (for user-facing changes)

### Commit Message Format

```
type(scope): brief description

Longer explanation if needed

Fixes #123
```

**Types:**
- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation changes
- `style`: Code formatting
- `refactor`: Code restructuring
- `test`: Adding/updating tests
- `chore`: Maintenance tasks
- `security`: Security-related changes

## Security

### Reporting Security Issues

**DO NOT** file public GitHub issues for security vulnerabilities.

Instead:
1. Email security concerns to the maintainers
2. Provide detailed reproduction steps
3. Allow reasonable time for response and patching

### Security Best Practices

- **Input validation**: Always validate file paths and user inputs
- **Memory safety**: Handle sensitive data (passwords, keys) securely
- **Dependency scanning**: Keep dependencies updated and secure
- **Static analysis**: Use gosec and CodeQL for vulnerability detection

## Community

### Communication Channels

- **GitHub Issues**: Bug reports, feature requests, questions
- **GitHub Discussions**: General discussions, ideas, support
- **Pull Requests**: Code contributions and reviews

### Recognition

Contributors will be recognized in:
- `CHANGELOG.md` for their contributions
- GitHub contributor graphs
- Release notes for significant contributions

### Maintainer Guidelines

Maintainers are expected to:
- Review PRs within 72 hours
- Provide constructive feedback
- Maintain code quality standards
- Keep the project secure and up-to-date

---

## Quick Reference

```bash
# Essential commands for contributors
git clone https://github.com/YOUR_USERNAME/vultool.git
cd vultool && git submodule update --init
make ci-local                    # Test everything
make build && ./vultool --help   # Try your changes
```

**Questions?** Open an issue or start a discussion. We're here to help!
