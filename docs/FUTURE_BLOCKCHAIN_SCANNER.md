# Future Enhancement: Phase 2 Blockchain Scanner

## Overview

This document outlines a future enhancement to vultool's recovery capabilities: automated blockchain scanning to detect funded addresses during gap limit recovery scenarios.

## Current State (Phase 1: Complete )

**Enhanced `list-paths` command** with sequential address generation:

```bash
# Generate 20 sequential Ethereum addresses for gap limit scanning
vultool list-paths -f vault.vult --chain ethereum --sequential --count 20

# Generate sequential paths for all supported chains
vultool list-paths -f vault.vult --sequential --count 10
```

**Features:**
-  Hybrid approach: Common paths (default) + sequential scanning (`--sequential`)
-  Default gap limit of 20 addresses for sequential mode  
-  Support for all major chains (Bitcoin, Ethereum, Solana, THORChain, etc.)
-  Proper derivation path generation for each chain type

## Future Enhancement: Phase 2 Blockchain Scanner

### Proposed Commands

```bash
# Ultimate recovery command - scan addresses and check blockchain for activity
vultool recover-scan -f vault.vult --chain ethereum --count 20 --check-balance --api etherscan

# Automated scanning for all chains with funded address detection
vultool scan-addresses -f vault.vult --sequential --count 20 --detect-funds --api-key YOUR_KEY

# Check specific addresses for activity (from generated list)
vultool list-paths -f vault.vult --chain ethereum --sequential --count 20 --json > addresses.json
vultool check-addresses addresses.json --api etherscan
```

### Implementation Strategy

#### Blockchain APIs Integration
- **Ethereum**: Etherscan API, Alchemy, Infura
- **Bitcoin**: Blockstream API, BlockCypher, Mempool.space
- **Solana**: Solana RPC endpoints
- **Multi-chain**: Moralis, Covalent APIs

#### Enhanced Recovery Flow
1. **Generate Addresses**: Use existing sequential path generation
2. **Batch API Calls**: Query multiple addresses per API call where supported
3. **Activity Detection**: Check transaction count, current balance, historical activity
4. **Gap Limit Logic**: Stop scanning when N consecutive unused addresses found
5. **Results Summary**: Report found addresses with balance/activity summary

### Real-World Recovery Scenarios

#### Scenario 1: "Lost MetaMask Recovery"
```bash
# User: "I lost my MetaMask, I know I used multiple accounts"
vultool recover-scan -f vault.vult --chain ethereum --count 20 --api etherscan

# Output:
#  Address #3: 2 transactions, 0.5 ETH balance
#  Address #7: 15 transactions, 0.0 ETH balance (used but empty)  
#  Addresses #8-20: No transactions (gap limit reached)
# 
# Summary: Found funds in 1 address, detected 2 used addresses
```

#### Scenario 2: "Multi-Chain Recovery"
```bash
# User: "I used this vault across multiple chains, not sure where my funds are"
vultool recover-scan -f vault.vult --all-chains --api-key YOUR_KEY

# Output:
# Bitcoin: Found funds in address #1 (0.05 BTC)
# Ethereum: Found funds in address #7 (1.2 ETH)  
# Solana: No funds detected (scanned 20 addresses)
# THORChain: Found funds in address #1 (500 RUNE)
```

### Technical Implementation Notes

#### API Rate Limiting
- Implement exponential backoff for API rate limits
- Support multiple API providers with failover
- Cache results to avoid duplicate queries

#### Security Considerations
- Never send private keys to APIs (address-only queries)
- Support local node connections where possible
- Optional API key encryption for storage

#### Performance Optimization
- Batch address queries where API supports it
- Parallel scanning across different chains
- Progress indicators for long scans

### Configuration

```yaml
# ~/.config/vultool/blockchain_apis.yaml
ethereum:
  primary: "etherscan"
  apis:
    etherscan:
      base_url: "https://api.etherscan.io/api"
      rate_limit: 5 # requests per second
    alchemy:
      base_url: "https://eth-mainnet.g.alchemy.com/v2"
      
bitcoin:
  primary: "blockstream"
  apis:
    blockstream:
      base_url: "https://blockstream.info/api"
      rate_limit: 2
```

### Future Command Extensions

```bash
# Advanced recovery with custom gap limits per chain
vultool recover-scan -f vault.vult --config recovery_config.yaml

# Historical balance checking (find peak balances)
vultool scan-addresses -f vault.vult --historical --from 2020-01-01

# Export funded addresses for external tools
vultool recover-scan -f vault.vult --export funded_addresses.csv --format csv
```

## Priority and Timeline

**Priority**: Medium (after core recovery features are solid)
**Estimated Implementation**: 2-4 weeks for basic blockchain scanning
**Dependencies**: 
- Stable Phase 1 implementation 
- API provider selection and testing
- Rate limiting and error handling framework

## Related Issues

- Gap limit standards vary by wallet (20 is common, but some use 10 or 50)
- Different chains have different optimal scanning strategies
- Balance thresholds (ignore dust amounts vs. report everything)
- Historical vs. current balance reporting preferences

---

*This document serves as a technical specification for future development. The Phase 1 sequential address generation is complete and ready for production use.*
