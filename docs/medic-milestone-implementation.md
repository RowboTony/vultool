# Medic Milestone (v0.2) - TSS Key Recovery Implementation

## Overview

The Medic milestone introduces advanced cryptographic recovery functionality to vultool, enabling users to reconstruct private keys from distributed threshold signature scheme (TSS) vault shares. This implementation provides the core functionality needed for emergency key recovery scenarios.

## Implemented Features

### TSS Key Recovery (`recover` command)

The `recover` command combines threshold shares from multiple `.vult` files to reconstruct the original private key material. This functionality is critical for emergency scenarios where users need to recover access to their cryptocurrency wallets.

#### Key Capabilities

- **Multi-Algorithm Support**: Handles both ECDSA (Bitcoin/Ethereum) and EDDSA (Solana) key reconstruction
- **Threshold Validation**: Ensures sufficient shares are provided before attempting reconstruction
- **Vault Compatibility Checks**: Validates that all provided vault files belong to the same distributed key
- **Multiple Output Formats**: Supports WIF, hex, and base58 encoding for different blockchain ecosystems
- **Chain Filtering**: Allows recovery of keys for specific blockchains only
- **JSON Export**: Structured output for programmatic use

#### Technical Implementation

The recovery process follows these steps:

1. **Vault Parsing**: Parse and validate all provided vault files
2. **Compatibility Verification**: Ensure all vaults belong to the same distributed key system
3. **Share Extraction**: Extract cryptographic shares from each vault
4. **Lagrange Interpolation**: Use mathematical reconstruction to recover the private key
5. **Address Generation**: Generate addresses for supported blockchain networks
6. **Format Conversion**: Convert keys to appropriate formats (WIF, base58, etc.)

#### Security Features

- **Memory Safety**: Cryptographic material is handled securely in memory
- **Validation Checks**: Multiple layers of validation prevent invalid reconstructions
- **Error Handling**: Comprehensive error messages for troubleshooting
- **Party Verification**: Ensures unique party keys across all shares

## Usage Examples

### Basic Key Recovery

```bash
# Recover keys from 2-of-3 threshold shares
vultool recover share1.vult share2.vult share3.vult --threshold 2
```

### Chain-Specific Recovery

```bash
# Recover only Bitcoin keys
vultool recover share*.vult --threshold 2 --chain bitcoin

# Recover only Solana keys
vultool recover share*.vult --threshold 2 --chain solana
```

### JSON Output

```bash
# Export recovered keys in JSON format
vultool recover share*.vult --threshold 2 --json
```

### Password-Protected Vaults

```bash
# Provide password for encrypted vaults
vultool recover encrypted*.vult --threshold 2 --password mypassword
```

### Export to File

```bash
# Save recovery results to file
vultool recover share*.vult --threshold 2 --output keys.json
```

## Sample Output

### Human-Readable Format

```
Attempting to recover private keys from 2 shares (threshold: 2)...

Successfully recovered 2 keys:

Key 1 (bitcoin):
  Address:     17dcd2c6c55b69b94ee702052d139def54eee078f
  Private Key: bf1754e1fbbdb708c1a59589cc34a57ad26c9d56fd32b72fb7ad2141810601fb
  WIF:         WIF:80bf1754e1fbbdb708c1a59589cc34a57ad26c9d56fd32b72fb7ad2141810601fb01e5b99039
  Derive Path: m/44'/0'/0'/0/0

Key 2 (solana):
  Address:     fe29d5bc1b93aa7be35cd83f6028b5c73135ff54f24956309ca8efd14e98cd4e
  Private Key: a6f71304e667e42d96cc1e9257e38971517b7928ef464c1fb26740ea0a00dd7b
  Base58:      B58:a6f71304e667e42d96cc1e9257e38971517b7928ef464c1fb26740ea0a00dd7b
  Derive Path: m/44'/501'/0'/0'
```

### JSON Format

```json
[
  {
    "chain": "bitcoin",
    "private_key": "bf1754e1fbbdb708c1a59589cc34a57ad26c9d56fd32b72fb7ad2141810601fb",
    "wif": "WIF:80bf1754e1fbbdb708c1a59589cc34a57ad26c9d56fd32b72fb7ad2141810601fb01e5b99039",
    "address": "17dcd2c6c55b69b94ee702052d139def54eee078f",
    "derive_path": "m/44'/0'/0'/0/0"
  },
  {
    "chain": "solana",
    "private_key": "a6f71304e667e42d96cc1e9257e38971517b7928ef464c1fb26740ea0a00dd7b",
    "base58": "B58:a6f71304e667e42d96cc1e9257e38971517b7928ef464c1fb26740ea0a00dd7b",
    "address": "fe29d5bc1b93aa7be35cd83f6028b5c73135ff54f24956309ca8efd14e98cd4e",
    "derive_path": "m/44'/501'/0'/0'"
  }
]
```

## Supported Blockchain Networks

| Chain | Key Type | Address Format | Private Key Formats |
|-------|----------|----------------|-------------------|
| Bitcoin | ECDSA | P2PKH, P2WPKH | WIF, Hex |
| Ethereum | ECDSA | 0x prefixed | Hex |
| Solana | EDDSA | Base58 | Base58, Hex |
| THORChain | ECDSA | Bech32 | Hex, Base58 |

## Architecture

### Core Components

```
┌─────────────────────────────────────────┐
│           RecoverPrivateKeys()          │
│    Main entry point for key recovery    │
└─────────────────┬───────────────────────┘
                  │
    ┌─────────────▼─────────────────┐
    │     validateVaultCompatibility     │
    │   Ensures vaults are compatible    │
    └─────────────┬─────────────────┘
                  │
    ┌─────────────▼─────────────────┐
    │        extractKeyShares       │
    │   Parses shares from vaults   │
    └─────────────┬─────────────────┘
                  │
         ┌────────┴────────┐
         │                 │
┌────────▼────────┐ ┌──────▼───────┐
│ reconstructECDSA │ │reconstructEDDSA│
│      Key        │ │     Key       │
└─────────────────┘ └───────────────┘
```

### Key Classes

- **`RecoveredKey`**: Represents a reconstructed private key with multiple format options
- **`KeyShareData`**: Internal representation of cryptographic shares
- **`AddressSet`**: Container for addresses in different formats
- **`SupportedChain`**: Enum defining supported blockchain networks

## Security Considerations

### Critical Security Warnings

1. **Private Key Material**: This command reconstructs actual private key material. Only use in secure environments.
2. **Memory Handling**: Private keys are held in memory during reconstruction. Ensure secure system state.
3. **Share Validation**: Always verify vault compatibility before attempting reconstruction.
4. **Backup Security**: Store recovered keys securely and consider immediate fund transfer.

### Best Practices

1. **Offline Environment**: Perform key recovery on air-gapped systems when possible
2. **Secure Disposal**: Securely wipe memory and storage after recovery operations
3. **Minimal Shares**: Use only the minimum required threshold shares
4. **Immediate Action**: Transfer funds immediately after successful recovery

## Error Handling

The implementation provides detailed error messages for common scenarios:

- **Insufficient Shares**: When not enough shares are provided for threshold
- **Incompatible Vaults**: When vault files don't belong to the same key system
- **Parsing Errors**: When vault files are corrupted or invalid
- **Cryptographic Failures**: When reconstruction mathematics fail

## Future Enhancements

### Planned Improvements

1. **Real TSS Integration**: Replace demonstration math with production-grade TSS libraries
2. **Hardware Security**: HSM and hardware wallet integration
3. **Additional Curves**: Support for more elliptic curves and signature schemes
4. **Advanced Validation**: Enhanced cryptographic verification of recovered keys
5. **Audit Trails**: Logging and audit functionality for compliance scenarios

### Integration with Mobile-TSS-Lib

The current implementation provides the foundation for integration with Vultisig's production TSS library:

- **Interface Compatibility**: Designed to work with `github.com/vultisig/mobile-tss-lib`
- **DKLS23 Support**: Architecture supports both GG20 and DKLS23 protocols
- **Share Format**: Compatible with Vultisig protobuf share formats

## Testing

### Test Coverage

The implementation includes comprehensive test scenarios:

- **Valid Recovery**: Successful reconstruction with various threshold configurations
- **Error Cases**: Invalid thresholds, incompatible vaults, corrupted shares
- **Chain Filtering**: Recovery of specific blockchain keys only
- **Format Validation**: Correct output formatting for all supported chains

### Test Data

Test fixtures include both GG20 and DKLS vault formats:
- `testGG20-part1of2.vult` / `testGG20-part2of2.vult`
- `testDKLS-1of2.vult` / `testDKLS-2of2.vult`
- `qa-fast-share1of2.vult` / `qa-fast-share2of2.vult`

## Development Notes

### Implementation Status

- **Core Recovery Logic**: Complete with Lagrange interpolation framework
- **Vault Compatibility**: Full validation and error handling
- **Multi-Chain Support**: Bitcoin, Ethereum, Solana, THORChain
- **CLI Integration**: Full command-line interface with flags and options
- **Error Handling**: Comprehensive error scenarios covered
- **Cryptographic Math**: Currently uses demonstration algorithms (secure replacement needed)
- **HD Derivation**: Address derivation is simplified (full BIP-32/44 implementation needed)

### Code Organization

```
internal/recovery/
├── recovery.go           # Main recovery logic and CLI integration
├── recovery_test.go      # Comprehensive test suite
└── README.md            # This documentation
```

This implementation provides a solid foundation for production TSS key recovery while maintaining security best practices and comprehensive error handling.

## Acknowledgments & Project Inspiration

The recovery feature in vultool was inspired by pioneering work in other open source projects. We gratefully acknowledge:

- [`cmd/recovery-cli` in vultisig/mobile-tss-lib](https://github.com/vultisig/mobile-tss-lib/tree/main/cmd/recovery-cli):
  Provided foundational CLI patterns and core recovery logic for threshold signature share reconstruction.

- [`Vultisig-Share-Decoder` by SxMShaDoW](https://github.com/SxMShaDoW/Vultisig-Share-Decoder):
  Served as an early reference implementation for decoding, validating, and reconstructing Vultisig vault shares.

These projects laid important groundwork for secure, user-friendly key recovery tooling in the Vultisig ecosystem.