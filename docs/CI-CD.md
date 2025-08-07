# CI/CD Pipeline

**Navigation**: [README](../README.md) | [Contributing Guide](../CONTRIBUTING.md) | [Security](#security-scanning)

## Overview

Vultool uses a comprehensive CI/CD pipeline that ensures code quality, security, and reliable releases across all platforms.

**Platforms**: Linux (amd64/arm64), macOS (amd64/arm64), Windows (amd64)  
**Security**: gosec, govulncheck, CodeQL, dependabot  
**Quality**: golangci-lint, formatting, tests across Go 1.21-1.23  
**Automation**: Automated builds, releases, and dependency updates

## Workflows

- **`.github/workflows/ci.yml`**: Main CI (runs on PRs and main branch)
- **`.github/workflows/release.yml`**: Release automation (runs on tags)
- **`.github/workflows/security.yml`**: Daily security scans

## Local Development

**Prerequisites for local CI**: Ensure `golangci-lint` is installed:
```bash
# Install golangci-lint (required for make ci-local)
curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin v1.61.0
export PATH=$PATH:$(go env GOPATH)/bin

# Run full CI locally
make ci-local

# Individual components
make security-scan           # gosec + govulncheck
make build-all-platforms     # Cross-platform builds
make coverage               # Test coverage report
make setup-hooks           # Install pre-commit hooks
```

## Creating Releases

```bash
# Update version and create release
echo "1.0.0" > VERSION
git add VERSION
git commit -m "chore: bump version to 1.0.0"
git tag v1.0.0
git push origin main --tags
```

The release workflow automatically:
- Builds binaries for all platforms
- Generates SHA256 checksums
- Creates GitHub release with assets
- Generates release notes from commits

## Configuration Files

- **`.gosec.json`**: Security scanner settings
- **`.golangci.yml`**: Linter configuration
- **`.github/dependabot.yml`**: Dependency auto-updates
- **`.github/codeql/`**: Advanced security analysis

## Security Scanning

Vultool implements multi-layered security scanning:

### Static Analysis
- **gosec**: Scans Go source code for security flaws
- **govulncheck**: Checks for known vulnerabilities in dependencies
- **CodeQL**: Advanced semantic code analysis (GitHub Security tab)

### Dependency Management
- **Dependabot**: Automated dependency updates with security patches
- **License scanning**: Ensures license compatibility
- **Supply chain security**: Verified build provenance

### Running Security Scans Locally
```bash
# Full security suite
make security-scan

# Individual scans
gosec ./...
govulncheck ./...
```

## Quality Gates

All PRs must pass:
- Tests on all platforms and Go versions
- Security scans (gosec, govulncheck)
- Linting and formatting checks
- Cross-platform build validation

## Repository Setup

1. **Enable GitHub security features**:
   - Settings > Security & analysis > Enable all

2. **Set branch protection** (Settings > Branches):
   - Require status checks: `test`, `lint`, `security`, `build`, `validate`
   - Require up-to-date before merge
   - Include administrators

3. **First release**:
   ```bash
   make ci-local  # Test locally first
   git tag v0.1.0 && git push --tags
   ```

That's it. The pipeline handles the rest automatically.
