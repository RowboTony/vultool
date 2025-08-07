Absolutely! This is a **major clarity milestone**—and you’re right:

* **mobile-tss-lib** is for GG20 (legacy) and is probably “feature frozen.”
* **dkls23** (from Silence Labs) is the *standard* for all modern DKLS-based MPC wallets, including Vultisig iOS/Android.

---

# **Technical Workflow & Integration Plan: DKLS23 in a CLI (.vult Compatible)**

---

## **1. Project Architecture Overview**

**Goal:**
Build a cross-platform CLI (vultool) that is fully compatible with Vultisig DKLS `.vult` files.

* **Must:** Recombine DKLS keyshares for recovery and signing.
* **Must:** Produce valid addresses and WIF/private keys matching Vultisig apps.
* **Should:** Auto-detect GG20 (legacy) vs DKLS (modern) and branch accordingly.

---

## **2. Core Components**

### **A. dkls23 Library (Rust/Go)**

* **Source:** [https://github.com/silence-laboratories/dkls23](https://github.com/silence-laboratories/dkls23)
* **Implements:**

  * DKLS ECDSA/EdDSA keygen
  * Keyshare serialization/deserialization
  * Share combination & threshold signing
* **FFI/Bindings:**

  * Rust is the reference implementation.
  * Bindings exist for Go, Node.js, and C.
  * Used by Vultisig apps via FFI.

### **B. Vultool CLI Wrapper**

* **Written in:** Go (or your preferred language)
* **Responsibilities:**

  * Read and parse `.vult` files (protobuf, base64 decoding)
  * Identify GG20 vs DKLS scheme (`ResharePrefix` or JSON vs binary)
  * For DKLS:

    * Use dkls23 bindings to parse and combine shares
    * Output recovered private key/address
  * For GG20:

    * Fall back to mobile-tss-lib
  * Match Vultisig wallet behaviors exactly.

---

## **3. Detailed Workflow**

### **A. Parsing .vult Files**

* **Detect format:**

  * If keyshare decodes as JSON → GG20
  * If keyshare is base64’d binary protobuf → DKLS
* **For DKLS:**

  * Use Vultisig’s protobuf schema (`VaultContainer`, `LocalState`)
  * Extract all `Xi` shares, IDs, and related metadata.

### **B. Calling dkls23 for Key Reconstruction**

* **Language:**

  * Prefer Go for CLI (since Go bindings exist), but Rust/C/Node.js all possible.
* **Import/share-combine function:**

  * Use dkls23’s “combine shares” or “recover key” routine (sometimes called `reconstruct` or `threshold_sign`).
  * Pass in the required threshold of shares (k-of-n), matching the vault config.

### **C. Address/Key Formatting**

* **Once you have the recovered private key:**

  * For Bitcoin: Output as WIF (Base58Check) and derive bc1 (P2WPKH) and legacy addresses.
  * For Ethereum: Output as hex and derive 0x address.
  * For Solana: Output as base58.
* **Verify:**

  * Ensure the derived public address matches the one shown in the original Vultisig vault/app.

### **D. (Optional) Signing Support**

* You can extend vultool to sign transactions using the reconstructed DKLS private key, via dkls23, for advanced workflows (cold storage, multisig, etc.).

---

## **4. Integration Steps**

### **Step-by-Step Plan**

**1. Get familiar with dkls23:**

* Read docs and run their test vectors.
* Build/test their sample apps (Rust, Go).

**2. Choose FFI/Bindings for your CLI:**

* Use the Go binding: [https://github.com/silence-laboratories/dkls23/tree/main/wrapper/go-dkls/go-wrapper](https://github.com/silence-laboratories/dkls23/tree/main/wrapper/go-dkls/go-wrapper)
* Or call the Rust lib via cgo or a thin CLI shell if you prefer.

**3. Implement .vult file decoding:**

* Use the protobuf definitions from [commondata](https://github.com/vultisig/commondata/tree/main/go/vultisig/vault/v1)
* Parse vault files, extract keyshare data.

**4. DKLS Keyshare Deserialization:**

* Decode each keyshare with the correct protobuf schema.
* Gather `Xi` values, IDs, threshold, etc.

**5. Invoke dkls23 Share Combination:**

* Call `combine_shares()` or equivalent, passing all decoded shares.

**6. Format & Validate Output:**

* Derive all supported address formats and verify against the vault metadata.

**7. Package and Document:**

* Add clear error handling (wrong share count, incompatible shares, etc.)
* Document DKLS version and Go/Rust dependency management.

---

## **5. Technical & Developer Notes**

* **Build Challenges:**

  * You may need to build and statically link the dkls23 Rust code for Go.
  * Pay attention to Rust/go module versions, and Protobuf compatibility.

* **Testing:**

  * Use real test vectors (from the iOS app, share-decoder, or known-good vaults).
  * Validate output matches Vultisig’s address and public key exactly.

* **Security:**

  * DKLS is threshold-secure; never allow key reconstruction from fewer than k shares.
  * Zero out sensitive memory when done.

---

## **6. Visual Workflow Diagram**

```mermaid
flowchart TD
    A[.vult Files] --> B{GG20 or DKLS?}
    B -- GG20 --> C[mobile-tss-lib]
    B -- DKLS --> D[dkls23 (Rust/Go)]
    D --> E[Combine Shares]
    E --> F[Recovered Private Key]
    F --> G[Derive Address/Export Key]
```

---

## **7. Summary Table**

| Step                | Tool/Library          | Details                   |
| ------------------- | --------------------- | ------------------------- |
| Parse .vult         | commondata/vault/v1   | Protobuf, base64          |
| Detect scheme       | vultool logic         | ResharePrefix/JSON/binary |
| DKLS combine shares | dkls23 (Rust/Go)      | Threshold/ID/Xi, etc      |
| GG20 combine shares | mobile-tss-lib        | JSON LocalState           |
| Address derivation  | bitcoinlib/ethers/sol | Standard formats          |
| Validate            | vultool, user         | Compare to vault metadata |

---

**If you want a sample code snippet, build script, or Go module instructions for dkls23 FFI, just ask!**
