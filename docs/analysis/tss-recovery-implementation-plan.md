# TSS Key Recovery Implementation Plan

## Current Problem

The current `recover` command implementation is fundamentally flawed:
- It's not using the actual TSS keyshare data from vault files
- It's creating fake shares by hashing public keys
- It's using XOR instead of proper TSS reconstruction
- The recovered keys don't match the actual vault keys

## Evidence

User's Vultisig wallet shows: `bc1qwzpjqun2rfga2fu0ld4wlk27tw2dk3ljxh2yyl`
After importing our "recovered" WIF: `bc1qkv6n9p4es3yfx6kp58zwc5gy9fln8k3ds0wq4mn`

These are completely different keys - we're not recovering the right private key.

## Root Cause

TSS (Threshold Signature Scheme) doesn't work by simple share combination:
1. During vault creation, parties run a distributed key generation ceremony
2. Each party gets a complex share structure (not just a number)
3. The shares include secret fragments, Paillier keys, commitments, etc.
4. Recovery requires either:
   - Running the TSS protocol locally as multiple parties
   - Properly reconstructing the key using the exact sharing scheme

## What's Actually in a Vault File

```protobuf
message Vault {
  string name = 1;
  string public_key_ecdsa = 2;  // The distributed public key
  string public_key_eddsa = 3;
  repeated string signers = 4;
  string hex_chain_code = 6;     // For HD derivation
  repeated KeyShare key_shares = 7;
  string local_party_id = 8;
}

message KeyShare {
  string public_key = 1;
  string keyshare = 2;  // <-- THE ACTUAL TSS SHARE DATA (serialized)
}
```

## The Solution

### Step 1: Understand the Keyshare Format

The `keyshare` field contains serialized TSS share data. We need to:
1. Determine if it's GG20 or DKLS format
2. Understand the serialization format (likely protobuf or JSON)
3. Parse the actual cryptographic components

### Step 2: Choose Implementation Strategy

#### Option A: Direct Integration with mobile-tss-lib
```go
import "github.com/vultisig/mobile-tss-lib/tss"
```

Advantages:
- Guaranteed compatibility with Vultisig
- Handles all protocol complexity
- Maintained by Vultisig team

Challenges:
- Designed for mobile bindings (gomobile)
- May need adaptation for CLI use

#### Option B: Understand and Reimplement
Study the TSS protocol and implement recovery:
- For GG20: Reconstruct using Paillier decryption and Lagrange interpolation
- For DKLS: Use the specific DKLS reconstruction method

Advantages:
- Full understanding of the process
- No external dependencies

Challenges:
- Complex cryptography
- High risk of errors
- Time-consuming

### Step 3: Correct Bitcoin Address Derivation

Vultisig likely uses:
- BIP-84 for Native SegWit: `m/84'/0'/0'/0/0`
- Bech32 address format (bc1q...)

The recovered private key must be:
1. Used with the correct derivation path
2. Combined with the chain code from the vault
3. Properly formatted for the target wallet

## Immediate Action Plan

### Phase 1: Investigation (Today)
1. Parse and examine the actual `keyshare` field content
2. Determine the serialization format
3. Identify if vaults use GG20 or DKLS
4. Test with known vault sets to understand the structure

### Phase 2: Prototype (This Week)
1. Create a parser for the keyshare data
2. Attempt basic reconstruction with 2-of-2 shares
3. Verify against known addresses

### Phase 3: Implementation
Either:
- A: Integrate mobile-tss-lib properly
- B: Implement minimal TSS recovery for our specific needs

### Phase 4: Validation
1. Test with multiple vault sets
2. Verify recovered addresses match Vultisig app
3. Test HD derivation for multiple addresses
4. Ensure all supported chains work

## Technical Requirements

### For Proper TSS Recovery
1. Parse serialized TSS shares from vault files
2. Handle both GG20 and DKLS protocols
3. Implement proper mathematical reconstruction
4. Support threshold requirements (2-of-2, 2-of-3)

### For Bitcoin Specifically
1. Use correct BIP-84 derivation path
2. Apply chain code for HD derivation
3. Generate proper Bech32 addresses
4. Create valid WIF for legacy wallet import

### For Other Chains
- Ethereum: Standard ECDSA with keccak256 addresses
- Solana: Ed25519 with base58 addresses
- ThorChain: Bech32 with 'thor' prefix

## Code Structure Changes Needed

```go
// internal/recovery/tss.go
type TSSShare struct {
    PartyID     string
    ShareData   []byte  // Actual deserialized TSS share
    Protocol    string  // "GG20" or "DKLS"
}

func ParseTSSKeyshare(keyshareData string) (*TSSShare, error)
func ReconstructPrivateKey(shares []*TSSShare, threshold int) (*ecdsa.PrivateKey, error)
```

## Testing Strategy

1. Create test vaults with known private keys
2. Verify recovery produces same keys
3. Test with real Vultisig vault exports
4. Validate addresses across all chains

## Security Considerations

- Never log or display intermediate share values
- Clear sensitive data from memory after use
- Validate threshold requirements strictly
- Warn users about private key exposure risks

## Success Criteria

1. Recovered Bitcoin address matches Vultisig exactly
2. WIF imports successfully into standard wallets
3. All supported chains produce correct addresses
4. Works with both 2-of-2 and 2-of-3 vaults
