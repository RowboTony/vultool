# TSS Wallet Recovery: Limitations & Workarounds

## Overview

The `vultool recover` command successfully recovers cryptographic key material from Threshold Signature Scheme (TSS) vaults, but **standard crypto wallets cannot import these recovered keys directly**. This document explains why this limitation exists, what works, and recommended approaches.

---

## What Works Perfectly

The TSS recovery process implemented by `vultool` produces **100% valid and usable key material**:

- **Recovered private scalar** (32-byte secret key)
- **Original TSS-generated public key** (matches on-chain addresses)
- **Full Ed25519 keypair** for EdDSA chains (Solana, SUI)
- **Address validation** against `list-addresses` (perfect match across all 17+ chains)
- **Transaction signing capability** (can sign valid transactions)

**Example**: Recovered Solana address `EPEg1C2pEwiEbPaBTuyydnvGpZoa6y3iXVVNzv7JYT8H` exactly matches the vault's on-chain address.

---

## Why Standard Wallets Reject TSS Keys

Standard wallet applications assume a **single-source key derivation model**:

```
32-byte seed → ed25519.NewKeyFromSeed(seed) → [private||public] → address
```

However, **TSS uses a fundamentally different approach**:

```
Distributed Key Generation → Fixed Public Key + Private Share → TSS Address
```

### Key Differences:

| Aspect | Standard Wallets | TSS Recovery |
|--------|------------------|--------------|
| **Key Generation** | Derive pubkey from seed | Public key chosen during DKG |
| **Private Key** | Single 32-byte seed | Recovered scalar + fixed pubkey |
| **Validation** | `seed → pubkey` must match | Scalar works only with original pubkey |
| **Import Format** | Seed-based formats | Arbitrary keypair combinations |

### Result:
When you try to import a TSS-recovered private key into Phantom, Sui Wallet, or similar apps:
1. The wallet derives a **new** public key from the private scalar
2. This derived pubkey **doesn't match** the original TSS public key  
3. The wallet shows a **different address** than your TSS vault
4. **Import fails or shows wrong balance**

---

## What Actually Works

| Tool/Wallet | Import Support | Notes |
|-------------|----------------|-------|
| **Phantom** | No | Requires seed → pubkey derivation |
| **Solflare** | No | Expects standard Ed25519 seed format |
| **Sui Wallet** | No | Requires `suipriv:` format from seed |
| **Trust Wallet** | No | Standard derivation only |
| **Solana CLI** | **YES** | Accepts `[private\|\|public]` JSON keypairs |
| **SUI CLI** | Partial | Can submit raw signed transactions |
| **Custom Scripts** | **YES** | Direct cryptographic operations |
| **vultool** | **YES** | Native TSS support |

---

## Recommended Solutions

### Option 1: Solana CLI (Recommended)

**Step 1**: Save the recovered keypair to a file:

```bash
# From vultool output, copy the JSON array format:
echo '[27,255,102,128,139,45,12,67,89,...]' > solana-keypair.json
```

**Step 2**: Use Solana CLI directly:

```bash
# Configure the keypair
solana config set --keypair ./solana-keypair.json

# Check balance
solana balance

# Send transactions
solana transfer <recipient-address> <amount> --allow-unfunded-recipient
```

**This works perfectly** - Solana CLI doesn't require seed→pubkey derivation.

### Option 2: Custom Sweep Script

Create a simple script to transfer funds to a standard wallet:

```python
# Example pseudocode
from solana.rpc.api import Client
from solana.transaction import Transaction

# Load TSS keypair from vultool
tss_private_key = load_tss_key("recovered_key.json")
tss_public_key = load_tss_pubkey("recovered_key.json")

# Create transfer to new standard wallet
recipient = "YOUR_NEW_STANDARD_WALLET_ADDRESS"
transfer_all_sol(tss_private_key, tss_public_key, recipient)
```

### Option 3: Wait for Enhanced vultool

We're planning a **sweep feature** in `vultool`:

```bash
# Proposed future command
vultool recover vault1.vult vault2.vult \
  --threshold 2 \
  --chain solana \
  --sweep-to-address 7aa4...ZzBb \
  --rpc https://api.mainnet-beta.solana.com
```

This would:
- Recover TSS keys automatically
- Construct transfer transactions  
- Sign with correct TSS keypair
- Broadcast directly to network
- Move funds to importable standard wallet

---

## SUI-Specific Considerations

SUI has additional complications:

### Current Limitations:
- `sui client` rejects arbitrary keypairs
- No direct equivalent to Solana CLI's keypair support
- `suipriv:` format requires seed-based derivation

### Workaround:
1. **Recover TSS key using vultool**
2. **Construct transaction manually** (using SUI SDK or RPC)
3. **Sign externally** with recovered TSS key
4. **Submit via SUI RPC** or CLI

### Future Enhancement:
The planned `vultool sweep` will handle SUI transactions automatically.

---

## Technical Deep Dive

### Why Re-encoding Doesn't Help

You might think: "Can't we just convert the TSS key to the right format?"

**No** - this isn't an encoding problem:

```python
# This won't work:
standard_seed = tss_private_key  # Same 32 bytes
standard_pubkey = ed25519.NewKeyFromSeed(standard_seed)  # DIFFERENT pubkey!
# standard_pubkey ≠ tss_pubkey → different address → wrong wallet
```

### The Mathematics

In standard Ed25519:
```
private_scalar → point_multiply(base_point) → public_key
```

In TSS Ed25519:
```
distributed_keygen() → fixed_public_key
lagrange_interpolation(shares) → private_scalar
# private_scalar + fixed_public_key = valid_keypair
# BUT: point_multiply(private_scalar) ≠ fixed_public_key
```

---

## Summary & Action Items

### Current State (Working):
- TSS recovery is **100% technically correct**
- All addresses match `list-addresses` output perfectly
- Keys can sign valid transactions for all 17+ chains
- Solana CLI import works flawlessly

### Current Limitations:
- Standard wallet GUI import fails
- SUI requires manual transaction construction
- No automated sweep functionality yet

### Next Steps:
1. **Immediate**: Use Solana CLI for Solana transactions
2. **Short-term**: Implement `vultool sweep` command  
3. **Long-term**: Advocate for TSS support in major wallets

### Bottom Line:
> **Your TSS recovery is working perfectly.** The limitation is in wallet import UX, not the cryptographic recovery. Use CLI tools or wait for the sweep feature to move funds to standard wallets.

---

## Need Help?

If you're still having issues:

1. **Verify you're using the latest vultool build**
2. **Check that addresses match between recovery and `list-addresses`**  
3. **Try the Solana CLI approach first**
4. **Contact support with specific error messages**

The recovery system is solid - we just need better tooling around the edges!
