# Development Principles

This document outlines the core development principles that guide vultool's implementation and all contributions to the project.

## Core Philosophy: No Silent Placeholders

**The most important rule:** Never implement silent placeholders, stubs, or hardcoded values that could be mistaken for working functionality.

### The Problem with Silent Stubs

Silent placeholders cause significant developer frustration and hidden bugs by:
- Creating false confidence that features are implemented
- Wasting debugging time when "working" features produce incorrect results
- Making it difficult to distinguish between incomplete and broken functionality
- Hiding technical debt until it becomes a critical issue

### Anti-Pattern Examples

```go
// ❌ BAD: Silent placeholder
func deriveBitcoinCashAddress(pubkey []byte) string {
    return "bitcoincash:qr6r525sj29cayjqqpzgkeqtzeeyj49hxuqqqqqqqp" // TODO: implement properly
}

// ❌ BAD: Hardcoded test value without indication
func deriveSUIAddress(pubkey []byte) string {
    return "0xe36ca893894810713425724d15aedc3bf928013852cb1cd2d3676b1579f7501a"
}

// ❌ BAD: Silent failure
func validateVaultFile(path string) bool {
    // TODO: implement validation
    return true
}
```

## The Four Pillars of Reliable Development

### 1. Fail Fast with Explicit Errors

Unimplemented functionality should fail immediately with clear error messages.

```go
// ✅ GOOD: Explicit error for unimplemented functionality
func deriveBitcoinCashAddress(pubkey []byte) (string, error) {
    return "", fmt.Errorf("Bitcoin Cash address derivation not yet implemented (issue #123)")
}

// ✅ GOOD: Panic for critical missing functionality
func deriveSUIAddress(pubkey []byte) string {
    panic("SUI address derivation not implemented - see docs/IMPLEMENTATION_STATUS.md")
}
```

### 2. Feature Flags with Clear Status

Use feature flags and build tags to clearly indicate incomplete functionality.

```go
//go:build experimental

package experimental

// ExperimentalFeatures tracks incomplete implementations
var ExperimentalFeatures = map[string]string{
    "bitcoin_cash_cashaddr": "Partial implementation - checksum validation pending",
    "sui_blake2b_derivation": "Not implemented - placeholder returns hardcoded value",
    "thorchain_addresses":   "Complete but untested with mainnet",
}

// ✅ GOOD: Feature flag with status
func DeriveBitcoinCashAddress(pubkey []byte) (string, error) {
    if !FeatureEnabled("bitcoin_cash_cashaddr") {
        return "", errors.New("Bitcoin Cash address derivation is experimental - enable with --experimental-features")
    }
    // ... implementation
}
```

### 3. Progressive Implementation with Validation

Implement features incrementally with validation at each step.

```go
// ✅ GOOD: Progressive implementation with validation
func deriveAddressWithValidation(chain string, pubkey []byte) (string, error) {
    switch chain {
    case "bitcoin":
        return deriveBitcoinAddress(pubkey) // ✓ Fully implemented
    case "ethereum":
        addr, err := deriveEthereumAddress(pubkey) // ✓ Implemented but needs testing
        if err != nil {
            return "", fmt.Errorf("ethereum derivation failed: %w", err)
        }
        return addr, validateEthereumAddress(addr) // ✓ Validation included
    case "bitcoin_cash":
        return "", fmt.Errorf("Bitcoin Cash not yet supported - track progress in issue #456")
    default:
        return "", fmt.Errorf("unsupported chain: %s - see docs/SUPPORTED_CHAINS.md", chain)
    }
}
```

### 4. Build Tags for Incomplete Features

Use build tags to exclude incomplete functionality from production builds.

```go
//go:build !production

// TestOnlyFunctions should never be included in production builds
package testonly

func DeriveTestAddress(chain string) string {
    // This function only exists for testing and is excluded from production
    return fmt.Sprintf("test-%s-address", chain)
}
```

## Implementation Guidelines

### Error Messages Should Be Actionable

```go
// ❌ BAD: Vague error
return "", errors.New("address derivation failed")

// ✅ GOOD: Actionable error
return "", fmt.Errorf("Bitcoin Cash CashAddr derivation failed: invalid public key length %d, expected 33 bytes", len(pubkey))
```

### Document Implementation Status

Maintain clear documentation of what's implemented, what's not, and what's experimental.

**File: `docs/IMPLEMENTATION_STATUS.md`**
```markdown
# Implementation Status

## Address Derivation Support

| Chain          | Status      | Notes                           |
|----------------|-------------|--------------------------------|
| Bitcoin        | ✅ Complete | Fully tested with mainnet      |
| Ethereum       | ✅ Complete | Fully tested with mainnet      |
| Bitcoin Cash   | ❌ Missing  | Issue #123 - CashAddr encoding |
| SUI            | ⚠️ Partial  | Hardcoded test value only      |
| Thorchain      | 🧪 Experimental | Works but needs more testing |
```

### Use Runtime Checks for Critical Paths

```go
// ✅ GOOD: Runtime validation for critical functionality
func init() {
    // Validate that all required address derivation functions are implemented
    requiredChains := []string{"bitcoin", "ethereum"}
    for _, chain := range requiredChains {
        if !IsChainSupported(chain) {
            panic(fmt.Sprintf("Critical chain %s is not properly implemented", chain))
        }
    }
}
```

### Test Stubs Should Fail by Default

```go
// ✅ GOOD: Test stub that fails explicitly
func TestBitcoinCashAddressDerivation(t *testing.T) {
    t.Skip("Bitcoin Cash address derivation not implemented - see issue #123")
    
    // When implemented, this test should pass:
    // addr, err := deriveBitcoinCashAddress(testPubkey)
    // require.NoError(t, err)
    // assert.Equal(t, "expected-address", addr)
}
```

## Development Workflow

### Before Implementing New Features

1. **Check for existing stubs**: Search for TODO comments and placeholder implementations
2. **Document the plan**: Create or update implementation status documentation  
3. **Write failing tests**: Tests should fail explicitly until implementation is complete
4. **Use feature flags**: Mark experimental or incomplete features clearly

### Code Review Checklist

- [ ] No silent placeholders or hardcoded test values
- [ ] Unimplemented functions return explicit errors
- [ ] Feature flags are used for experimental functionality
- [ ] Implementation status is documented
- [ ] Error messages are actionable and specific
- [ ] Tests fail explicitly for unimplemented features

### Continuous Integration

The CI pipeline should catch silent placeholders:

```yaml
# Check for problematic patterns
- name: Check for silent placeholders
  run: |
    # Look for TODO comments without associated errors
    if grep -r "TODO.*implement" --include="*.go" . | grep -v "return.*error\|panic\|Skip"; then
      echo "Found silent TODO placeholders that should return errors"
      exit 1
    fi
    
    # Look for hardcoded addresses without feature flags
    if grep -r "bitcoincash:\|0x[0-9a-f]\{40\}" --include="*.go" . | grep -v "test\|example\|FeatureFlag\|experimental"; then
      echo "Found hardcoded addresses that should use proper derivation"
      exit 1
    fi
```

## Migration Strategy for Existing Code

When fixing existing silent placeholders:

1. **Identify all placeholders**: Search for hardcoded values, TODO comments, and stub implementations
2. **Categorize by impact**: Critical path functions should return errors; test functions can panic
3. **Add explicit status**: Use feature flags, build tags, or clear error messages
4. **Update documentation**: Reflect the true implementation status
5. **Fix incrementally**: Address critical paths first, then less critical functionality

## Benefits of This Approach

- **Developer confidence**: No surprises about what's implemented vs. what's a placeholder
- **Easier debugging**: Failures are explicit and immediate rather than silent and confusing  
- **Better project management**: Clear visibility into what needs implementation
- **Reduced technical debt**: Incomplete features are tracked and addressed systematically
- **Improved reliability**: Production builds don't include incomplete functionality by accident

## Examples from Recent Fixes

The Bitcoin Cash and SUI address derivation issues that prompted these principles demonstrate the problem:

- **Silent placeholders** returned hardcoded addresses that looked correct
- **Debugging took hours** to discover the addresses were fake
- **No clear indication** that the functionality was incomplete
- **Technical debt** accumulated until it became a blocking issue

The fixes applied these principles:
- **Explicit errors** for unimplemented functionality
- **Proper library usage** replacing hardcoded stubs
- **Clear status tracking** in documentation
- **Validation** to catch future regressions

---

**Remember**: It's better to have obviously broken functionality than silently incorrect functionality. Fail fast, fail clearly, and document what's not implemented.
