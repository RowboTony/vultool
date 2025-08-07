# Understanding Vultisig's TSS Architecture and Recovery Process

## Key Discoveries

After deep investigation into the mobile-tss-lib codebase, I've uncovered how Vultisig's TSS recovery actually works:

### 1. Vault File Structure

The `.vult` files contain:
- **Encrypted/Unencrypted protobuf data** (VaultContainer)
- Inside the vault: **KeyShares** with JSON-encoded `tss.LocalState` structures
- The `keyshare` field is NOT raw binary TSS data, but JSON containing the full TSS party state

### 2. LocalState Structure

```go
type LocalState struct {
    PubKey              string                         `json:"pub_key"`
    ECDSALocalData      keygen.LocalPartySaveData      `json:"ecdsa_local_data"`
    EDDSALocalData      eddsaKeygen.LocalPartySaveData `json:"eddsa_local_data"`
    KeygenCommitteeKeys []string                       `json:"keygen_committee_keys"`
    LocalPartyKey       string                         `json:"local_party_key"`
    ChainCodeHex        string                         `json:"chain_code_hex"`
    ResharePrefix       string                         `json:"reshare_prefix"`
}
```

The ECDSA/EDDSALocalData contains:
- `Xi`: The actual secret share
- `ShareID`: The party's share identifier
- Other TSS protocol data

### 3. Multi-Party Simulation Architecture

Vultisig's TSS is designed for true multi-party computation across devices. Recovery requires simulating multiple parties on one device:

#### Recovery-Web Approach
- Uses WASM to run TSS recovery in browser
- Compiles Go code to WebAssembly
- Allows client-side key recovery without server trust

#### Recovery-CLI Approach
- Local command-line recovery
- Reads multiple vault files (threshold shares)
- Reconstructs private key using Shamir's secret sharing

#### WASM-Coordinator
- Coordinates multi-party protocols in browser
- Supports key generation, resharing, and signing
- Used by Vultisig web extension

### 4. Recovery Process

1. **Parse vault files** - Extract encrypted protobuf data
2. **Decrypt if needed** - Using Argon2 or SHA256 of password
3. **Extract LocalState** - JSON unmarshal from keyshare field
4. **Collect threshold shares** - Need at least t-of-n shares
5. **Reconstruct secret** - Using VSS (Verifiable Secret Sharing)
6. **Derive keys** - HD derivation for different chains

### 5. Why Simple XOR Doesn't Work

- TSS uses proper Shamir's secret sharing with Lagrange interpolation
- Shares are points on a polynomial, not simple XOR masks
- The mobile-tss-lib uses binance's tss-lib for cryptographic operations
- Recovery requires proper threshold cryptography, not just combining bytes

## Implementation Strategy

### Current Approach (Incorrect)
- Trying to parse keyshare as binary data
- Attempting simple XOR reconstruction
- Missing the JSON LocalState structure

### Correct Approach
1. Parse vault protobuf properly
2. Extract and JSON unmarshal the keyshare field
3. Use the tss.LocalState structure
4. Implement proper VSS reconstruction using mobile-tss-lib
5. Support both ECDSA and EdDSA key types

## Next Steps

1. Fix vault parser to properly export JSON keyshare data
2. Update recovery implementation to use LocalState
3. Implement proper VSS reconstruction
4. Test with real vault shares
5. Verify recovered addresses match expected wallets

## Architecture Implications

This discovery reveals that Vultisig's architecture is more sophisticated than initially understood:

- **True MPC**: Designed for genuine multi-party computation
- **No single point of failure**: Keys never exist in one place
- **Client-side recovery**: Can be done without trusting servers
- **Multiple recovery paths**: CLI, web, and potentially native apps

The recovery tools (recovery-cli, recovery-web, wasm-coordinator) are essentially creating a local simulation of the multi-party protocol to reconstruct keys when needed.
