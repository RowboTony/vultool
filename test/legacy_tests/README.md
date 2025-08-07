# Legacy Tests - Archived

This directory contains test files that are no longer valid due to significant architecture changes in the vultool recovery system.

## Files

- `recovery_test_obsolete.go` - Original recovery tests that tested outdated assumptions about:
  - Expected number of supported chains (expected 4, we now have 19+)
  - Legacy WIF/Base58 formatting expectations
  - DKLS vault support (not yet implemented)
  - Outdated error message formats
  - Old function signatures and behaviors

## Why These Were Removed

The architecture has evolved significantly:

1. **Chain Support**: We now support 19+ blockchains instead of the original 4
2. **TSS Recovery**: Modern TSS recovery system uses proper cryptographic reconstruction
3. **Address Derivation**: Centralized address derivation ensures consistency
4. **Error Handling**: Updated error messages and handling patterns
5. **Test Focus**: Focus shifted to end-to-end integration testing vs unit testing internals

## Current Test Strategy

The new test suite (`internal/recovery/recovery_test.go`) focuses on:
- End-to-end GG20 vault recovery with real test fixtures
- Validation that all 19 chains recover correctly
- SUI address derivation verification (the original issue that prompted this cleanup)
- Basic error handling for edge cases

These legacy tests are kept here for reference but should not be used as they test outdated assumptions about the system.
