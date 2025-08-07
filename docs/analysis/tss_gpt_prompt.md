## **AI Coder Prompt: Vultool – GG20 and DKLS .vult Recovery CLI**

---

### **Prompt for Marj: GG20 & DKLS .vult Recovery Implementation in vultool**

#### **Context**

You are implementing robust, Vultisig-compatible `.vult` file recovery in the vultool CLI. The .vult file format is used by Vultisig and supports two TSS protocols:

* **GG20 (legacy)**: ECDSA only. Use the official [mobile-tss-lib](https://github.com/vultisig/mobile-tss-lib) for parsing and recovery.
* **DKLS (modern, default in 2025)**: ECDSA & EdDSA. Use [dkls23](https://github.com/silence-laboratories/dkls23/) for parsing and share recombination.

**NO homebrew, custom, or "pattern-based" recovery methods are allowed.** Only audited, Vultisig-official libraries are to be used for both parsing and recombination.

---

#### **Project Requirements**

1. **Implement GG20 .vult Recovery First**

   * Use only `mobile-tss-lib` for parsing and recovery of GG20 .vult files.
   * Do not attempt custom share recombination.
   * **SUCCESS CRITERION:** The recovered private key/address **must match** the output of `vultool list-addresses` for the same GG20 .vult. This validates true compatibility.

2. **DKLS Support—Implement After GG20 Is Proven**

   * Once GG20 recovery is robust, add DKLS23 support **using only the official Silence Labs [dkls23](https://github.com/silence-laboratories/dkls23/) library** for key recombination.
   * Do not attempt to "extract" or "pattern-match" key material from DKLS .vult shares. Always use the audited library’s recombination flow.

3. **List-Addresses as Ground Truth**

   * Use `vultool list-addresses` as the *ground truth* validator. The recovered key from your implementation **must** produce the same Bitcoin/Ethereum address as `list-addresses` for the same .vult file (or recombined vault).
   * If the addresses do not match, **recovery is NOT valid**—do not proceed to DKLS until GG20 is 100% correct.

4. **False Positives to Avoid**

   * **Do not** accept any recovery result unless the address matches exactly. False positives occur when “extracting” a private key from a share, rather than doing proper recombination.
   * **Never** implement custom recovery methods outside the official libraries, even if they appear to "work" (e.g., pattern-based methods, byte-offset hacks).
   * If a share can be decoded alone (i.e., outputs a valid key on its own), this is a false positive for Vultisig's threshold scheme—the private key must require threshold recombination.

5. **Only Proceed to DKLS Once GG20 Recovery is Verified**

   * Do not attempt DKLS implementation until GG20 recovery passes all tests.
   * DKLS recombination will require calling the WASM/Rust server or using the Rust bindings—integrate the library, do not re-implement the cryptography.

---

#### **Concrete Implementation Steps**

1. **GG20 .vult Recovery**

   * Use the official mobile-tss-lib for all share decoding, recombination, and private key derivation.
   * Add a CLI command (e.g., `vultool recover`) that takes ≥threshold GG20 .vult files and outputs the recovered private key (WIF, hex, etc.).
   * After recovery, **verify** that the output address matches `vultool list-addresses`.

2. **Validation**

   * Add tests to confirm that for each supported chain (BTC, ETH, etc.), the address produced by your recovered private key matches `list-addresses`.
   * If validation fails, halt and fix before continuing.

3. **Documentation & Warnings**

   * Document clearly that this tool is for emergency/private key recovery, which destroys MPC security—users must sweep funds immediately after use.

4. **DKLS Implementation (After GG20 Proven)**

   * Integrate the Silence Labs dkls23 library, use the official recombination method only.
   * Repeat the recovery/address-matching validation for DKLS.

---

**References:**

* `.vult` format: [commondata](https://github.com/vultisig/commondata/tree/main/go/vultisig/vault/v1)
* GG20: [mobile-tss-lib](https://github.com/vultisig/mobile-tss-lib)
* DKLS23: [dkls23 Rust](https://github.com/silence-laboratories/dkls23/)
* Address derivation spec: [wallet-core](https://github.com/rowbotony/wallet-core) (for key to address)

---

**Remember: The address matching test is the ultimate proof—if the recovered key doesn’t match the address in list-addresses, recovery is not valid.**
No shortcuts, no pattern-matching. Use the official, audited libraries only.

---

#### **Success = GG20 recover matches list-addresses, then repeat for DKLS.**

---

Let me know when GG20 is fully validated and ready for DKLS, or if you need any clarification during implementation!
