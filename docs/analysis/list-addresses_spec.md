# Multi-Chain Address Derivation Specification

## Overview
This document specifies the requirements and implementation details for deriving addresses across all chains supported by Vultisig.

## Supported Chains Analysis

### Chain Categories by Implementation Complexity

## 1. ECDSA/Secp256k1 Chains (Same elliptic curve as Bitcoin/Ethereum)

### Bitcoin-like Chains (BIP44 derivation + address encoding)
| Chain | Ticker | Derivation Path | Address Format | Implementation |
|-------|--------|-----------------|----------------|----------------|
| Bitcoin | BTC | m/84'/0'/0'/0/0 | Bech32 (bc1...) |  Implemented |
| Bitcoin Cash | BCH | m/44'/145'/0'/0/0 | CashAddr | Phase 1 |
| Litecoin | LTC | m/84'/2'/0'/0/0 | Bech32 (ltc1...) | Phase 1 |
| Dogecoin | DOGE | m/44'/3'/0'/0/0 | Base58 (D...) | Phase 1 |
| Dash | DASH | m/44'/5'/0'/0/0 | Base58 (X...) | Phase 1 |
| Zcash | ZEC | m/44'/133'/0'/0/0 | Base58 (t1...) | Phase 1* |

*Note: Zcash t-addresses only. z-addresses require additional cryptography.

### EVM-Compatible Chains (Same address derivation as Ethereum)
| Chain | Ticker | Derivation Path | Address Format | Implementation |
|-------|--------|-----------------|----------------|----------------|
| Ethereum | ETH | m/44'/60'/0'/0/0 | Hex (0x...) |  Implemented |
| BNB Smart Chain | BSC | m/44'/60'/0'/0/0 | Hex (0x...) | Phase 1 |
| Avalanche C-Chain | AVAX | m/44'/60'/0'/0/0 | Hex (0x...) | Phase 1 |
| Polygon | MATIC | m/44'/60'/0'/0/0 | Hex (0x...) | Phase 1 |
| Cronos | CRO | m/44'/60'/0'/0/0 | Hex (0x...) | Phase 1 |
| Arbitrum | ETH | m/44'/60'/0'/0/0 | Hex (0x...) | Phase 1 |
| Optimism | ETH | m/44'/60'/0'/0/0 | Hex (0x...) | Phase 1 |
| Base | ETH | m/44'/60'/0'/0/0 | Hex (0x...) | Phase 1 |
| Blast | ETH | m/44'/60'/0'/0/0 | Hex (0x...) | Phase 1 |
| zkSync | ETH | m/44'/60'/0'/0/0 | Hex (0x...) | Phase 1 |

## 2. EdDSA/Ed25519 Chains (Different elliptic curve)

| Chain | Ticker | Derivation Path | Address Format | Implementation |
|-------|--------|-----------------|----------------|----------------|
| Solana | SOL | m/44'/501'/0'/0' | Base58 |  Partial |
| TON | TON | m/44'/607'/0'/0/0 | Base64 (bounceable) | Phase 3 |
| SUI | SUI | m/44'/784'/0'/0'/0' | Hex (0x...) | Phase 2 |

## 3. Cosmos Ecosystem (Secp256k1 with Bech32 addresses)

| Chain | Ticker | Derivation Path | Address Format | Implementation |
|-------|--------|-----------------|----------------|----------------|
| Cosmos | ATOM | m/44'/118'/0'/0/0 | Bech32 (cosmos1...) | Phase 2 |
| Osmosis | OSMO | m/44'/118'/0'/0/0 | Bech32 (osmo1...) | Phase 2 |
| Kujira | KUJI | m/44'/118'/0'/0/0 | Bech32 (kujira1...) | Phase 2 |
| Terra | LUNA | m/44'/330'/0'/0/0 | Bech32 (terra1...) | Phase 2 |
| Terra Classic | LUNC | m/44'/330'/0'/0/0 | Bech32 (terra1...) | Phase 2 |
| Akash | AKT | m/44'/118'/0'/0/0 | Bech32 (akash1...) | Phase 2 |
| dYdX | DYDX | m/44'/118'/0'/0/0 | Bech32 (dydx1...) | Phase 2 |
| Noble (USDC) | USDC | m/44'/118'/0'/0/0 | Bech32 (noble1...) | Phase 2 |
| THORChain | RUNE | m/44'/931'/0'/0/0 | Bech32 (thor1...) |  Partial |
| MayaChain | CACAO | m/44'/931'/0'/0/0 | Bech32 (maya1...) | Phase 2 |

## 4. Special Cases

| Chain | Ticker | Derivation Path | Address Format | Implementation |
|-------|--------|-----------------|----------------|----------------|
| Polkadot | DOT | m/44'/354'/0'/0/0 | SS58 | Phase 3 (Sr25519) |

## Implementation Phases

### Phase 1: ECDSA Chains (Easy - 1 day)
**Goal:** Add all EVM-compatible and Bitcoin-like chains that use ECDSA/Secp256k1

**Requirements:**
- Use existing ECDSA key from TSS recovery
- Implement address encoding for each format:
  - Base58 for legacy Bitcoin-like chains
  - Bech32 for SegWit chains
  - CashAddr for Bitcoin Cash
  - Keccak256 for EVM chains

**Chains to implement:**
- All EVM chains (BSC, AVAX, MATIC, CRO, Arbitrum, Optimism, Base, Blast, zkSync)
- Bitcoin-like chains (BCH, LTC, DOGE, DASH, ZEC)

### Phase 2: Cosmos Ecosystem & EdDSA Improvements (Moderate - 2-3 days)
**Goal:** Add Cosmos SDK chains and improve EdDSA chain support

**Requirements:**
- Bech32 encoding with custom HRP (Human Readable Part)
- Cosmos address derivation (different from Bitcoin's Bech32)
- Proper Ed25519 address encoding for SUI

**Chains to implement:**
- All Cosmos chains (ATOM, OSMO, KUJI, LUNA, LUNC, AKT, DYDX, USDC/Noble, CACAO)
- SUI with proper address format

### Phase 3: Complex Chains (Complex - 1 week)
**Goal:** Add chains with non-standard cryptography or complex address formats

**Requirements:**
- Sr25519 curve implementation for Polkadot
- SS58 address encoding
- TON's complex workchain/address structure
- Additional dependencies for non-standard crypto

**Chains to implement:**
- Polkadot (DOT)
- TON

## Address Derivation Functions

### Bitcoin-like Address Derivation
```go
func deriveBitcoinLikeAddress(publicKey []byte, chainType string) string {
    switch chainType {
    case "BTC":
        return encodeBech32("bc", hash160(publicKey))
    case "LTC":
        return encodeBech32("ltc", hash160(publicKey))
    case "BCH":
        return encodeCashAddr(hash160(publicKey))
    case "DOGE":
        return encodeBase58Check(0x1E, hash160(publicKey)) // D prefix
    case "DASH":
        return encodeBase58Check(0x4C, hash160(publicKey)) // X prefix
    case "ZEC":
        return encodeBase58Check(0x1CB8, hash160(publicKey)) // t1 prefix
    }
}
```

### EVM Address Derivation
```go
func deriveEVMAddress(publicKey []byte) string {
    // Remove the 0x04 prefix if present (uncompressed key)
    if len(publicKey) == 65 && publicKey[0] == 0x04 {
        publicKey = publicKey[1:]
    }
    
    // Keccak256 hash of the public key
    hash := keccak256(publicKey)
    
    // Take last 20 bytes
    return "0x" + hex.EncodeToString(hash[12:])
}
```

### Cosmos Address Derivation
```go
func deriveCosmosAddress(publicKey []byte, hrp string) string {
    // Cosmos uses compressed secp256k1 public keys
    compressedPubKey := compressPublicKey(publicKey)
    
    // SHA256 then RIPEMD160
    sha := sha256.Sum256(compressedPubKey)
    ripemd := ripemd160.New()
    ripemd.Write(sha[:])
    hash := ripemd.Sum(nil)
    
    // Encode with custom HRP
    return encodeBech32(hrp, hash)
}
```

## Testing Requirements

Each chain implementation should be tested with:
1. Known test vectors from official chain documentation
2. Comparison with addresses generated by official wallets
3. Validation against the expected addresses in our test vaults

## Security Considerations

1. **Key Derivation:** Always use proper HD derivation paths as specified by each chain
2. **Address Validation:** Implement checksum validation for each address format
3. **Test Coverage:** Ensure 100% test coverage for address generation functions
4. **No Private Keys:** Never log or expose private keys during address derivation

## References

- [BIP44 - Multi-Account Hierarchy](https://github.com/bitcoin/bips/blob/master/bip-0044.mediawiki)
- [SLIP44 - Registered coin types](https://github.com/satoshilabs/slips/blob/master/slip-0044.md)
- [EIP55 - Ethereum Address Checksum](https://eips.ethereum.org/EIPS/eip-55)
- [Cosmos Address Format](https://docs.cosmos.network/main/spec/addresses/bech32)
- [Bitcoin Cash Address Format](https://github.com/bitcoincashorg/bitcoincash.org/blob/master/spec/cashaddr.md)
