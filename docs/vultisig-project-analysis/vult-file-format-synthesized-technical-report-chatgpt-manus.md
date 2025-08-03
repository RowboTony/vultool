# `.vult` File Format and CLI Operations – **Technical Specification** (Aug 2025)

---

## 1. Cryptographic Context: ECDSA & EdDSA in Vultisig

* **Threshold Cryptography:**
  Vultisig implements multi-party computation (MPC) to enable distributed threshold signature schemes:

  * **ECDSA:** Elliptic Curve Digital Signature Algorithm—used for Bitcoin, Ethereum, and most blockchains. Supported via both legacy GG20 and DKLS23 threshold protocols.
  * **EdDSA:** Edwards-curve Digital Signature Algorithm—used for chains like Solana, THORChain, and MayaChain. EdDSA threshold signing is supported via DKLS23 and (optionally) via FROST/Schnorr libraries.

* **Signature Ceremony:**
  For both ECDSA and EdDSA, the private key never exists in one place. Each party holds a share; only by collaborating can they compute a valid signature (`(r,s)` for ECDSA, scalar for EdDSA).

* **Key Generation:**

  * Distributed Key Generation (DKG) produces the keypair and securely splits the private scalar `d` (ECDSA) or `a` (EdDSA) into `n` shares, written to `.vult` files.
  * Each share is a TSS `LocalState` object containing cryptographic material required for MPC signing.

* **Recovery:**
  If `t` of `n` `.vult` shares are combined, the full private scalar can be recomputed for emergency export in standard formats:

  * **ECDSA:** `privkey` (big int, serialized for Bitcoin WIF, Ethereum hex, etc)
  * **EdDSA:** `privkey` (scalar, for Solana/THORChain/Litecoin, etc)

---

## 2. `.vult` File Format – **Protocol Buffer Structure**

**High-level structure:**
A `.vult` file is a base64-encoded serialization of:

```protobuf
message VaultContainer {
  uint32 version = 1;
  bytes vault = 2;           // serialized Vault
  bool is_encrypted = 3;     // encryption flag
}
```

**`Vault` message** contains:

```protobuf
message Vault {
  string name = 1;
  bytes public_key_ecdsa = 2;
  bytes public_key_eddsa = 3;           // optional (if EdDSA supported)
  repeated string signers = 4;          // participant IDs
  int64 timestamp = 5;
  string hex_chain_code = 6;
  repeated KeyShare key_shares = 7;     // party shares
  string local_party_id = 8;
  string reshare_prefix = 9;
  string lib_type = 10;                 // GG20 or DKLS23
}
message KeyShare {
  bytes public_key = 1;
  string keyshare = 2;  // serialized TSS LocalState object (for ECDSA/EdDSA)
}
```

* **ECDSA context:** `public_key_ecdsa` and corresponding `KeyShare.keyshare` objects encode the ECDSA scalar shares and protocol transcript.
* **EdDSA context:** `public_key_eddsa` and optional EdDSA keyshares (if vault supports it).

**Key technical notes:**

* Each `.vult` file contains **one party's share** (not the full key).
* The cryptographic secret is always split; threshold recombination uses Shamir-like techniques but wrapped in DKLS23 or GG20 logic.
* **Chain code:** Used for HD derivation of keys/addresses (compatible with BIP32/44).

---

## 3. CLI Tooling: ECDSA/EdDSA-Aware Operations

### 3.1 Vault Lifecycle & Share Management

* **Keygen:** Initiates DKG; produces `.vult` for each party with ECDSA/EdDSA key shares.
* **Reshare:** Refreshes or changes the committee (`n`, threshold, or party IDs) using TSS reshare protocols; outputs new `.vult` files.
* **Upgrade/Migrate:** Migrate GG20-based (ECDSA only) vaults to DKLS23 (ECDSA and EdDSA), enabling EdDSA cross-chain support.

### 3.2 Inspection & Metadata (Protocol/Algo-aware)

| Command         | Output                                                                                                                                     |
| --------------- | ------------------------------------------------------------------------------------------------------------------------------------------ |
| `info`          | Vault name, protocol (GG20/DKLS23), scheme(s) supported (ECDSA, EdDSA), chain code, threshold, timestamp, signers, reshare prefix.         |
| `decode`        | JSON/YAML dump, with explicit parsing of ECDSA vs EdDSA public keys, party IDs, and keyshare types.                                        |
| `list-networks` | Enumerates all supported chains and their type (e.g., Bitcoin/Ethereum = ECDSA; Solana/THORChain = EdDSA). Uses chain code for derivation. |

### 3.3 Recovery and Signing

* **recover:** Accepts threshold `.vult` files, recombines ECDSA and EdDSA key shares:

  * Outputs:

    * **ECDSA:** WIF, hex, or raw (for Bitcoin, Ethereum, etc.)
    * **EdDSA:** Raw private scalar, or keypairs for Solana/THORChain, etc.
  * Warns that after export, security properties are lost (single key now exists).

* **signECDSA / signEDDSA:** Participates in threshold signing session (hex-encoded message or chain-specific transaction); outputs (r,s) signature or EdDSA scalar.

* **derive:** Given a derivation path, outputs derived ECDSA/EdDSA public keys (uses chain code and master keys).

### 3.4 Password/Encryption Management

* **set-password / change-password / remove-password:** All cryptographic, use symmetric encryption to secure the Vault bytes.

### 3.5 Advanced/Creative TSS Features

* **custom-threshold:** Advanced; allows non-standard t-of-n schemes, for research/test.
* **batch-sign:** Signs multiple transactions/messages (with both ECDSA and EdDSA if supported).
* **policy, role (future):** Reserved for when vault format supports differentiated signing roles (not present as of Aug 2025).

---

## 4. Security, Interoperability, and Implementation Notes

* **All cryptographic primitives and ceremony transcripts must preserve MPC security (no party ever learns full private scalar).**
* **DKLS23 and GG20 protocols are audited and open-source; implementation should reuse audited code from Silence Labs and Vultisig.**
* **No single point of failure; signing/derivation protocols never reconstruct key on device unless explicitly running a `recover` operation.**
* **No EdDSA support in legacy GG20 vaults—migration to DKLS23 required.**
* **Chain code/derivation must strictly follow BIP32/BIP44 for ECDSA (BTC/ETH) and Ed25519 derivation for EdDSA (SOL, THORChain, etc).**

---

## 5. Example CLI Usage

```sh
# Keygen (ECDSA + EdDSA, DKLS23)
vultool keygen --protocol DKLS23 --parties deviceA,deviceB,deviceC --threshold 2 --out-dir ./vaults

# Reshare (add/remove devices)
vultool reshare --old-vaults old1.vult,old2.vult --new-parties deviceA,deviceD,deviceE --protocol DKLS23

# Recover ECDSA and EdDSA keys
vultool recover --shares share1.vult,share2.vult,share3.vult --out-format json

# Derive Solana address (EdDSA)
vultool derive --file myvault.vult --chain solana --path "m/44'/501'/0'/0'"

# Sign message with ECDSA (BTC)
vultool signECDSA --file myvault.vult --message "..." --participants deviceA,deviceB
```

---

## 6. Reference Implementations

* [DKLS23 Rust Reference](https://github.com/silence-laboratories/dkls23)
* [Vultisig mobile-tss-lib](https://github.com/vultisig/mobile-tss-lib)
* [Threshold signature protocols: GG20, DKLS23, FROST/Schnorr](https://dkls.info/)
* [Protobuf schemas](https://github.com/vultisig/commondata/tree/main/proto/vultisig/vault/v1)

---

## 7. Appendix: Cryptographic Structures

### ECDSA TSS (GG20/DKLS23)

* Private key split: \$d = \sum\_{i=1}^{n} d\_i\$ (each party gets \$d\_i\$)
* Signing ceremony uses MPC to produce signature \$(r, s)\$, never reconstructs \$d\$.

### EdDSA TSS (DKLS23, FROST)

* Private scalar split: \$a = \sum\_{i=1}^{n} a\_i\$
* Aggregated signature is a single EdDSA scalar; derivation and verification use Ed25519/BLS APIs.

---

**This spec is designed for direct ingestion by LLMs or project generators.
All cryptographic terminology and protocol mechanics are explicit for downstream codegen or security review.**

---

Let me know if you need this split into user, dev, or cryptographer-focused sections—or need additional detail on ceremony steps, protobuf fields, or chain-specific implementation notes!
