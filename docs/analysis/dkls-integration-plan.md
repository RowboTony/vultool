# DKLS Integration Plan for vultool

## Overview
This document outlines the plan for integrating proper DKLS (Distributed Key Lifecycle Signature) vault reconstruction into vultool, based on the proven implementation from the Vultisig-Share-Decoder project.

## Current State

### vultool (Current)
- Has basic DKLS detection via `CheckIfDKLSVault()`
- Uses heuristic reconstruction methods (XOR, hash-based, entropy analysis)
- Does not implement proper DKLS cryptographic reconstruction
- Missing DKLS protobuf schema parsing

### Vultisig-Share-Decoder (Reference)
- Has comprehensive DKLS implementation in `pkg/dkls/dkls_native.go`
- Implements multiple reconstruction methods:
  - Multi-layer entropy analysis
  - Protobuf structure extraction
  - Binary structure analysis
  - Pattern-based extraction
  - Deterministic key generation
- Handles both encrypted and unencrypted DKLS vaults
- Generates proper cryptocurrency addresses (Bitcoin, Ethereum, etc.)

## Integration Strategy

### Phase 1: Direct Code Adaptation
Since both projects are Go-based and use similar dependencies, we can adapt the DKLS implementation directly:

1. **Copy Core DKLS Logic**
   - Adapt `pkg/dkls/dkls_native.go` methods into vultool's recovery package
   - Focus on the proven reconstruction methods:
     - `extractUsingMultiLayerEntropy()`
     - `extractFromProtobufStructure()`
     - `extractUsingEnhancedPatterns()`
     - `generateEnhancedDeterministicKey()`

2. **Integrate Processing Pipeline**
   - Port the `ProcessDKLSVaultFiles()` flow
   - Adapt `parseAndDecryptVault()` for vultool's structure
   - Implement `reconstructFromKeyshares()` with proven methods

3. **Add Helper Functions**
   - Port entropy calculation functions
   - Add pattern matching utilities
   - Include validation methods

### Phase 2: Protocol Schema Integration

1. **DKLS Protobuf Schema**
   - The DKLS keyshare data uses binary protobuf encoding (not JSON)
   - Need to understand the DKLS-specific protobuf structure
   - May need to reverse-engineer or obtain the schema

2. **Binary Parsing**
   - Implement proper binary protobuf parsing for DKLS keyshares
   - Handle length-prefixed fields correctly
   - Extract cryptographic material from proper offsets

### Phase 3: Testing and Validation

1. **Test Data**
   - Use the test vault files from both projects
   - Validate against known addresses
   - Compare results with Share Decoder output

2. **Cross-validation**
   - Ensure vultool produces same results as Share Decoder
   - Test with multiple vault combinations
   - Verify all supported chains (Bitcoin, Ethereum, Solana)

## Implementation Steps

### Step 1: Create Enhanced DKLS Module
```go
// internal/recovery/dkls_enhanced.go
package recovery

import (
    // ... imports from Share Decoder
)

type DKLSProcessor struct {
    initialized bool
}

// Port key methods from dkls_native.go
```

### Step 2: Update Recovery Flow
```go
// internal/recovery/tss_recovery.go
func ReconstructTSSKey(...) (*TSSRecoveryResult, error) {
    if isDKLS {
        // Use enhanced DKLS processor
        processor := NewDKLSProcessor()
        return processor.ReconstructPrivateKey(vaultFiles, password, keyType)
    }
    // ... existing GG20 logic
}
```

### Step 3: Add Pattern Recognition
```go
// Port pattern matching from Share Decoder
func (p *DKLSProcessor) extractUsingEnhancedPatterns(data []byte, shareIndex int) []byte {
    // DKLS-specific markers
    dklsMarkers := [][]byte{
        {0x12, 0x20}, // Protobuf wire format
        {0x1a, 0x20}, // Alternative protobuf pattern
        // ... more patterns
    }
    // ... pattern matching logic
}
```

### Step 4: Implement Multi-Method Reconstruction
```go
func (p *DKLSProcessor) reconstructFromKeyshares(keyshareDataList [][]byte, vaultInfos []VaultInfo) ([]byte, []byte, error) {
    // Try methods in order of success rate:
    // 1. Multi-layer entropy
    // 2. Enhanced patterns
    // 3. Protobuf structure
    // 4. Deterministic generation
}
```

## Key Differences to Handle

1. **Keyshare Format**
   - GG20: JSON-encoded LocalState
   - DKLS: Binary protobuf data

2. **Reconstruction Method**
   - GG20: VSS (Verifiable Secret Sharing) with Lagrange interpolation
   - DKLS: Custom protocol with binary parsing

3. **Validation**
   - Need to validate secp256k1 private keys
   - Check entropy levels
   - Verify curve constraints

## Dependencies to Add

```go
// May need additional imports:
import (
    "github.com/btcsuite/btcd/btcec/v2"
    "github.com/ethereum/go-ethereum/crypto"
    // ... other crypto libraries
)
```

## Success Criteria

1.  Successfully parse DKLS vault files
2.  Extract keyshare data from binary protobuf
3.  Reconstruct valid private keys
4.  Generate correct Bitcoin/Ethereum/Solana addresses
5.  Match output from Share Decoder for same inputs
6.  Handle both 2-of-2 and 2-of-3 DKLS vaults

## Timeline

- **Week 1**: Port core DKLS processing logic
- **Week 2**: Integrate pattern recognition and entropy analysis
- **Week 3**: Test with real vault files and validate addresses
- **Week 4**: Documentation and edge case handling

## Notes

- The Share Decoder uses heuristic methods because full DKLS protocol documentation is not publicly available
- The implementation relies on pattern recognition and entropy analysis
- Future work should integrate with official Vultisig DKLS library when available
- Consider using the mobile-tss-lib WASM module for proper DKLS support

## References

- Vultisig-Share-Decoder: https://github.com/SxMShaDoW/Vultisig-Share-Decoder
- Mobile TSS Lib: https://github.com/vultisig/mobile-tss-lib
- CommonData Protos: https://github.com/vultisig/commondata
