# Address Derivation Discrepancies Analysis

## Date: 2025-08-06

## Executive Summary
Analysis of address derivation discrepancies between `vultool list-addresses` output and Vultisig UI ground truth addresses reveals significant issues with Bitcoin Cash, THORChain, and SUI address derivation implementations.

## Ground Truth Addresses (from Vultisig UI)

### Verified Matches 
These addresses match perfectly between vultool and Vultisig UI:

| Chain | Address | Status |
|-------|---------|--------|
| Bitcoin | `bc1qvn203p8pp30fk945eywrjey937qpaanha8hc4r` |  Perfect Match |
| Litecoin | `ltc1qkgguledp08hpmcqsccxvwgr7xvhj7422qyz0l7` |  Perfect Match |
| Dogecoin | `DBMQ8aectXEd264wa7UoHT8YsghnXoxyrC` |  Perfect Match |
| Dash | `XkoQBncrZgAmHSYYhkjZqMF7NhPTBhbWbC` |  Perfect Match |
| Zcash | `t1ZiDZcAQMkRPQMEZTkJFAi7oZSJjn73Shb` |  Perfect Match |
| Ethereum | `0x55a7ea16a40f8c908cbc935d229ebe4c6658e90d` |  Perfect Match |
| BSC | `0x55a7ea16a40f8c908cbc935d229ebe4c6658e90d` |  Perfect Match |
| Solana | `EPEg1C2pEwiEbPaBTuyydnvGpZoa6y3iXVVNzv7JYT8H` |  Perfect Match |

### Critical Discrepancies 

#### 1. Bitcoin Cash (BCH)
**Issue**: Address format mismatch

| Source | Address | Format |
|--------|---------|--------|
| list-addresses | `1BgHJLWCqiby29HLks582e8j2DndyLAYdy` | Legacy P2PKH |
| Vultisig UI | `qp6379srrchrk2mfs32d2czxkx9wz2gx4qekc0x4xx` | CashAddr (no prefix) |

**Root Cause**: 
- Current implementation uses simplified legacy address format with "bitcoincash:" prefix
- Vultisig UI uses proper CashAddr format without the prefix
- The underlying public key derivation may be correct, but address encoding is wrong

**Fix Required**:
```go
// Current (incorrect)
legacyAddr := base58.CheckEncode(hash160, 0x00)
return "bitcoincash:" + legacyAddr

// Should be (correct CashAddr encoding)
// Implement proper CashAddr encoding according to spec
// https://github.com/bitcoincashorg/bitcoincash.org/blob/master/spec/cashaddr.md
```

#### 2. THORChain (RUNE)
**Issue**: Completely different addresses

| Source | Address |
|--------|---------|
| list-addresses | `thor1suABg8BSrQqUvgUyiU4zj9FMayb9PNwKZ92bnEii8N3` |
| Vultisig UI | `thor1d2y7x9tdqutkrwqcq9du9wfcgxch8zpcyff5ha` |

**Root Cause**:
- Simplified bech32 encoding implementation
- Possible issues with derivation path handling
- The `encode()` function is using a placeholder implementation

**Fix Required**:
```go
// Current (simplified/incorrect)
func encode(hrp string, data []byte) (string, error) {
    // Simplified bech32 encoding - this would need proper implementation
    return hrp + "1" + base58.Encode(data), nil
}

// Should use proper bech32 encoding library
// Import github.com/cosmos/cosmos-sdk/types/bech32
```

#### 3. SUI
**Issue**: Completely different addresses

| Source | Address |
|--------|---------|
| list-addresses | `0xc6da2ad7b18728f6481d747a7335fd52a5eed82f3c3d95a51deed03399c5c0b6` |
| Vultisig UI | `0xe36ca893894810713425724d15aedc3bf928013852cb1cd2d3676b1579f7501a` |

**Root Cause**:
- Current implementation just prefixes the EdDSA public key with "0x"
- SUI requires special address derivation using Blake2b hashing
- Missing proper SUI address format implementation

**Fix Required**:
```go
// Current (incorrect - just hex encoding)
suiAddr := "0x" + pubKeyHex

// Should be (proper SUI address derivation)
// 1. Create SUI address scheme flag byte (0x00 for Ed25519)
// 2. Concatenate flag + pubkey
// 3. Blake2b hash the result
// 4. Take first 32 bytes as address
```

## Key Insights

### 1. HD Derivation Paths
The implementation correctly treats all derivation paths as **non-hardened**, which is critical for Vultisig's ability to derive addresses from public keys only:

```go
// CRITICAL: Strip the hardened marker - Vultisig treats all paths as non-hardened!
component = strings.TrimSuffix(component, "'")
```

### 2. Chain Code Handling
The implementation attempts to use chain codes for proper HD derivation, but may fall back to zero chain codes if not available:

```go
if err != nil {
    // Use zero chain code if not provided
    chainCodeBytes = make([]byte, 32)
}
```

### 3. Address Encoding Issues
Most discrepancies are in the final address encoding step, not in key derivation:
- Bitcoin Cash: Needs proper CashAddr encoding
- THORChain: Needs proper bech32 implementation
- SUI: Needs Blake2b hashing and proper format

## Recommendations

### Immediate Actions

1. **Fix Bitcoin Cash CashAddr Encoding**
   - Implement proper CashAddr encoding per specification
   - Remove "bitcoincash:" prefix to match UI display
   - Test against known test vectors

2. **Fix THORChain Bech32 Encoding**
   - Use proper bech32 library (cosmos-sdk or btcutil)
   - Verify against Cosmos address derivation standards
   - Test with known THORChain addresses

3. **Implement SUI Address Derivation**
   - Add Blake2b hashing dependency
   - Implement SUI's specific address format
   - Follow SUI's official address derivation spec

### Testing Strategy

1. **Unit Tests**: Add tests for each address encoding function with known test vectors
2. **Integration Tests**: Verify derived addresses match Vultisig UI exactly
3. **Regression Tests**: Ensure changes don't break existing working addresses

### Code Quality Improvements

1. Replace simplified/placeholder implementations with proper libraries
2. Add comprehensive error handling and logging
3. Document each chain's specific requirements
4. Add validation for derived addresses

## Implementation Status

###  COMPLETED FIXES

1. **Bitcoin Cash (BCH)** -  **FIXED**
   - **Issue**: Using legacy P2PKH format instead of CashAddr
   - **Solution**: Implemented proper CashAddr encoding with bech32-like algorithm
   - **Result**: Now properly generates addresses like `qp6379srrchrk2mfs32d2czxkx9wz2gx4qekc0x4xx`
   - **Status**: Working correctly 

2. **THORChain (RUNE)** -  **FIXED** 
   - **Issue**: Simplified placeholder bech32 encoding
   - **Solution**: Integrated proper bech32 library from btcsuite
   - **Result**: Now properly generates addresses like `thor1d2y7x9tdqutkrwqcq9du9wfcgxch8zpcyff5ha`
   - **Status**: Working correctly 

###  REMAINING TODO

3. **SUI** -  **TODO**
   - **Issue**: Missing Blake2b hashing for proper SUI address format
   - **Current**: Simple hex prefix (placeholder)
   - **Required**: Implement Blake2b hashing with address scheme flag
   - **Priority**: Lower (can be addressed in future iterations)

## Final Results

**Before fixes**: 8/11 chains working (72.7%)
**After fixes**: 10/11 chains working (90.9%)

### Summary of Improvements
-  **Bitcoin Cash**: Now uses proper CashAddr encoding
-  **THORChain**: Now uses proper bech32 encoding with correct libraries
-  **All other chains**: Continue to work perfectly
-  **SUI**: Remains as TODO for future implementation

## Technical Implementation Details

### Bitcoin Cash CashAddr Fix
```go
func deriveBitcoinCashAddress(pubKey *secp256k1.PublicKey) string {
    // Convert to 5-bit groups for CashAddr encoding
    conv, err := convertBits(hash160, 8, 5, true)
    data := append([]byte{0}, conv...) // Add version byte
    
    // Encode using CashAddr format
    addr, err := encodeCashAddr("bitcoincash", data)
    
    // Remove prefix to match Vultisig UI display
    return addr[12:] // Remove "bitcoincash:" prefix
}
```

### THORChain Bech32 Fix
```go
// Use proper bech32 library from btcsuite
import "github.com/btcsuite/btcd/btcutil/bech32"

func deriveThorchainAddress(pubKey *secp256k1.PublicKey) string {
    conv, err := bech32.ConvertBits(hash160, 8, 5, true)
    addr, err := bech32.Encode("thor", conv)
    return addr
}
```

## Conclusion

 **SUCCESS!** The address derivation fixes have been implemented successfully:

- **Bitcoin Cash** and **THORChain** addresses now match Vultisig UI exactly
- **10 out of 11 supported chains** are working correctly (90.9% success rate)
- **Core HD derivation logic** was confirmed to be working properly
- **Only SUI remains** as a future enhancement (lower priority)

This represents a significant improvement in address derivation accuracy and brings the implementation very close to perfect compatibility with Vultisig UI ground truth addresses.
