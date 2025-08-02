# CHANGELOG

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- Enhanced test fixtures integration and validation
- Additional CLI export formats (YAML, Markdown)

### Changed
- Performance optimizations for large vault files

### Security
- Enhanced input validation for file path sanitization

## [v0.0.6-dev] - 2025-08-02
### BREAKING CHANGES
- **Standalone vultool**: Extracted from vulticore as independent CLI tool with own repository
- **Import changes**: vultool now has independent Go module `github.com/rowbotony/vultool`
- **Repository structure**: Complete standalone project with own CI/CD, documentation, and release process

### Added
- **Comprehensive CI/CD pipeline**: GitHub Actions workflows for testing, security scanning, and releases
- **Multi-platform support**: Automated builds for Linux (amd64/arm64), macOS (amd64/arm64), Windows (amd64)
- **Security integration**: gosec, govulncheck, CodeQL, and automated dependency updates
- **Enhanced Makefile**: Local CI simulation, security scanning, cross-platform builds, and pre-commit hooks
- **Professional documentation**: Concise CI/CD guide and updated README with development workflows
- **Independent project structure**: Complete separation from vulticore with own versioning and releases

### Enhanced
- **Public API**: Clean client interface in `pkg/client/` for library usage by other Go projects
- **Quality gates**: Comprehensive linting, formatting, and validation checks
- **Release automation**: Tag-triggered releases with checksums and professional release notes
- **Documentation**: Focused on CLI-first development and standalone usage

### Technical
- Established independent Go module `github.com/rowbotony/vultool`
- Implemented automated cross-platform binary compilation for releases
- Added security scanning with SARIF integration for GitHub Security tab
- Established branch protection rules and quality gate enforcement
- Complete CI/CD pipeline with multi-platform testing matrix

## [v0.0.5-dev] - 2025-08-02 (Final vulticore-embedded version)
### Added
- **Comprehensive CI/CD documentation**: Complete workflow guide with local testing instructions
- **Flutter test suite**: Working widget tests for VulticoreApp with proper provider setup
- **Real-time status badges**: GitHub Actions CI status badges with platform and security indicators
- **Local testing commands**: Documented Go and Flutter test procedures excluding problematic packages
- **Developer onboarding**: Step-by-step contribution guidelines with CI/CD integration workflows

### Fixed
- **Flutter test compilation errors**: Resolved MyApp constructor issue (updated to VulticoreApp)
- **Missing validation method**: Added comprehensive `_validateVultFile` method with JSON validation
- **Go test command errors**: Fixed `go test ./...` wasm package conflicts with targeted testing
- **Import dependencies**: Added missing provider imports and corrected widget test assertions

### Enhanced
- **Testing infrastructure**: Established foundation for Go unit tests with proper package exclusion
- **CI/CD reliability**: Verified all test commands work locally and integrate with GitHub Actions
- **Documentation structure**: Added CI/CD workflow documentation with cross-references
- **Status visibility**: Project health indicators via badges for CI, security, and platform support

## [v0.0.4-dev] - 2025-08-02 (Inherited from vulticore)
### Added
- **Content-based duplicate detection**: SHA-256 hashing for vault file imports
- **Multi-vault support**: Enhanced handling of multiple vault files simultaneously  
- **Cross-platform crypto compatibility**: Migrated crypto libraries for broader platform support

### Fixed
- **Duplicate detection logic**: Robust content hashing instead of filename-based checking
- **Cross-platform compilation**: Resolved platform-specific crypto library issues

## [v0.0.3-dev] - 2025-08-01 (Inherited from vulticore)
### Added
- **WebAssembly integration**: Complete WASM compilation pipeline for browser usage
- **JavaScript bridge**: Production-ready vault_parser_interface.js for web interop
- **Version management**: VERSION file integration with --version flag
- **Makefile build system**: Streamlined build process with version embedding

### Enhanced
- **Cross-platform parsing**: Browser-native .vult file analysis without CLI dependencies
- **Data flow optimization**: JSON serialization between WASM and runtime environments

### Technical
- Compiled Go vault parser to WebAssembly with exported JavaScript functions
- Cross-platform compatibility verified across major browsers
- Established foundation for encrypted vault and transaction signing support

## [v0.0.2-dev] - 2025-07-31 (Inherited from vulticore)
### Added
- **Security scanning**: gosec and govulncheck integration
- **Dependency management**: Automated upstream sync capabilities
- **CI/CD foundation**: Basic workflow structure and security gates

### Enhanced
- **Documentation**: Comprehensive development process documentation
- **Security practices**: Industry-standard security scanning and validation

### Technical
- Updated Go requirement to 1.21+
- Implemented change tracking and validation systems
- Clean commit history with milestone tracking

## [v0.0.1-dev] - 2025-07-31 (Inherited from vulticore)
### Added
- **Initial CLI implementation**: vault inspect command with summary, keyshares, validate, and export
- **Vault parsing**: Support for both encrypted and unencrypted .vult files  
- **Protobuf integration**: Official Vultisig commondata schema support
- **Security features**: Path validation and safe cryptographic operations

### Core Features
- Parse and display vault metadata
- Comprehensive vault file validation  
- Export vault data to JSON format
- Handle both GG20 and DKLS vault types
- Interactive and parameter-based password support for encrypted vaults

### Technical
- Project structure with cmd/, internal/, pkg/ organization
- Integration with Vultisig commondata protobuf schemas
- CLI framework using Cobra (industry standard)
- AES-GCM encryption support for encrypted vaults
