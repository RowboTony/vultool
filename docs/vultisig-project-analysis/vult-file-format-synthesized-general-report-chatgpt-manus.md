# .VULT MPC/TSS File Format & Ecosystem (July 2025)

---

## 1. .VULT File Format & Capabilities

### 1.1 Origin and Purpose

The `.vult` file format is the backup container used by the Vultisig wallet, a multi-platform threshold-signature (MPC/TSS) wallet originally developed by the founders of THORChain and supported by Vultisig and Silence Laboratories. Vaults are created by a distributed key-generation (DKG) ceremony among user devices using either the GG20 protocol (for legacy vaults) or the newer DKLS23 protocol. Each device receives a vault share—a secure fragment of the private key that never contains the entire secret. A vault share is stored as a `.vult` file, which contains only the data needed to participate in key generation and signing and to restore the vault; it never stores the seed phrase or full private key[^1]. The `.vult` files therefore replace seed phrases; if enough shares are lost, the wallet cannot sign transactions[^2].

### 1.2 Structure

The `.vult` file is basically a base64-encoded Protocol Buffers structure. `vault_container.proto` defines a VaultContainer with fields `version`, `vault` (base64-encoded serialized Vault), and an `is_encrypted` flag[^3]. `vault.proto` defines the Vault message with the vault name, public keys, list of signers, a creation timestamp, a hex chain code, the list of key\_shares (each containing a public\_key and a keyshare string), the local device’s party ID, and the signature scheme type (GG20 or DKLS)[^4]. The key share strings can be deserialized into `tss.LocalState` objects to reconstruct local party state[^5]. VaultContainer.vault is optionally encrypted with a user password.

### 1.3 Security Properties & Advantages Over Seed Phrases

* **No single point of failure:** The DKLS and GG20 schemes split the private key into shares that never exist together. Signing requires at least a threshold number of devices, so a compromised device cannot steal funds[^6]. Devices only connect during deliberate signing sessions and the private key is never stored or reconstructed[^7].
* **Seedless backup:** Vault shares are safely stored as `.vult` files. A seed phrase or full private key never exists, eliminating the need to write down a 12- or 24-word seed[^8]. Never store multiple vault shares together to avoid compromising the vault[^9].
* **Reduced communication rounds:** DKLS23 reduces signing communication from six rounds (GG20) to three and replaces Paillier encryption with efficient oblivious transfer, leading to faster and more reliable signing[^6].
* **No trusted third party:** DKLS23 provides malicious-majority security and ensures no device or server can reconstruct the full key. The DKLS protocol is UC-secure under standard assumptions.
* **Scalability across chains:** Vultisig supports Bitcoin, Ethereum, Solana, THORChain and more. The Vault structure includes chain-agnostic public keys and chain codes, allowing a single vault to hold multiple assets.

### 1.4 Threshold Configurations

Vultisig promotes 2/3 majority safety, recommending that the signing threshold be at least two-thirds of the participants. In the source code, the GetThreshold function calculates the threshold as ceil(n \* 2/3) – 1[^10], meaning a 3-of-4 vault (75%) or 2-of-3 vault (67%) is typical. 2-of-2 vaults have no redundancy and should be avoided[^2]. However, the Signing a Transaction guide mentions that vault types can be 2-of-2, 2-of-3, 3-of-4 or m-of-n; the file format and DKLS protocol support configurable thresholds, although the default threshold is two-thirds. Custom thresholds below two-thirds (e.g., 1-of-2 or 2-of-5) are not recommended due to security concerns and are not exposed in the standard UX.

### 1.5 Role-Based Permissions & Access Control

The `.vult` file itself does not contain role-based permissions. Each share is identical in capability; the DKLS protocol treats all parties as equal participants. Transaction policies (daily limits, address whitelists, delays, etc.) are planned for the future[^12] but not implemented as of July 2025. There are no flags inside the `.vult` structure that mark a share as “view only” or “approve only”; such logic would be enforced at the application layer rather than in the file format.

---

## 2. Key Management & Recovery

### 2.1 Importing Existing Seeds

Vultisig intentionally does not allow importing a BIP39 seed phrase or single-signature private key into a vault. The FAQ explains this improves security: existing seed phrases might have been exposed or generated poorly, and turning a compromised seed into a threshold key would preserve its weaknesses[^9]. The `.vult` standard is for new key generation rather than splitting existing keys.

### 2.2 Exporting & Backing Up

Each device’s vault share is exported as a `.vult` file. The files contain only non-sensitive data and can be backed up to a cloud provider or stored offline[^1]. Users must export shares manually; Vultisig does not auto-upload them to iCloud or other services[^9]. Backups should be stored separately; never keep multiple shares together.

### 2.2 Chain Codes and Key Derivation

The hex chain code, alongside the vault's public keys, is crucial for Vultisig's ability to derive keys across different blockchain networks without needing separate keys for each. This enables the generation of wallet addresses for multiple cryptocurrencies from a single `.vult` file, providing a seamless and secure experience.

Upcoming features will utilize this capability to support key derivation recovery fully by incorporating private keys for traditional wallet reconstruction. Furthermore, a new feature will enable adding and unlocking an encrypted `.vault` with a password through the CLI and a `[password]` input on the web app, enhancing security for encrypted vaults.

### 2.3 Recovery and Recombination

Because the secret key is never reconstructed during normal operation, recovery requires gathering at least the threshold number of vault shares. Vultisig offers an emergency recovery tool (CLI) in the mobile-tss-lib repo. The tool reads the `.vult` file, base64-decodes the VaultContainer, extracts each keyshare and deserializes it into a LocalState object[^5]. When enough shares are combined, the tool computes a standard private key for Bitcoin (WIF), Ethereum (hex), etc. The recovery process is one-way; once the private key is reconstructed and used, it is no longer part of an MPC environment and should be swept immediately[^13]. This enables interoperability with non-MPC wallets in emergency situations.

### 2.4 Consequences of Lost Shares

If fewer than the threshold number of vault shares remain, funds cannot be recovered. The FAQ advises users to enable remote wipe so that compromised devices can be erased and replaced. Without remote wipe, the recommendation is to import backups on new devices and move funds out of the compromised vault[^14]. Vultisig’s terms of service acknowledge that a critical bug or zero-day in the TSS implementation could corrupt shares or expose key material; users are encouraged to export shares immediately and keep off-platform backups[^15].

---

## 3. User Experience & Interoperability

### 3.1 Signing Workflow

Vultisig distinguishes between Fast and Secure vaults. In a secure vault (multi-device), one device initiates the transaction and displays a QR code. Other devices scan the QR code to join the keysign session. Users must confirm transaction details on their devices, after which the transaction is signed jointly and broadcast by one of the devices[^16][^17]. Vault types include 2-of-2, 2-of-3, 3-of-4 or m-of-n, and the signing “automatically starts when the threshold of devices joined”[^11][^17]. Local mode lets users perform the session over the local network without using the Vultisig relay server[^16].

### 3.2 Transport Mechanisms

The default transport is QR codes. Vultisig also offers remote signing via the Vultisig relay server, which relays encrypted signing messages between devices (useful when devices are not physically together). The local mode toggle indicates that network transport is pluggable; while the docs do not mention NFC or deep links, the architecture could support them. Devices can also sign via direct local network using the vultisig-relay-server component.

### 3.3 Device Support & UX Maturity

Vultisig runs on iOS, Android, macOS and Windows (via desktop). The DKLS scheme is default for new vaults; Windows requires enabling DKLS via an advanced toggle[^18]. Advanced features such as transaction policies, role-based approvals, and cross-vault swap routes are listed as “coming soon”[^12]. Core TSS functionality is production-ready, but advanced controls are still evolving.

---

## 4. Competing Consumer MPC/TSS Standards

### 4.1 ZenGo (KZen)

ZenGo is a consumer MPC wallet. It splits the private key into two shares: one on the user’s device and one on ZenGo’s servers. The shares sign transactions jointly; the key is never reconstructed[^19]. This eliminates seed phrases but introduces a semi-custodial element (ZenGo holds one share and can enforce policies). ZenGo does not publish a .vult-like file format.

### 4.2 Safeheron

Safeheron provides an enterprise MPC platform, allowing organizations to create millions of MPC wallets via APIs, with smart contract workflow integration[^20]. It combines MPC with trusted-execution environments (TEE) and supports policy engines for customizable approvals and address whitelists[^21][^22]. Safeheron open-sourced their MPC-TSS algorithm in C++[^23]. No widely adopted consumer file format comparable to .vult, but demonstrates advanced policy control and enterprise integrations.

### 4.3 Web3Auth (by MetaMask)

Web3Auth is a developer-focused platform that uses MPC and account abstraction to create non-custodial, seed-phrase-free wallets. Wallet creation is fast, seedless, and uses social logins[^24][^25]. Unlike Vultisig, Web3Auth offers a hosted key-share service and is oriented around dApp integration rather than stand-alone wallet UX. Based on DKLs19.

### 4.4 Lit Protocol and Others

Lit Protocol is a decentralized network for programmable signing and encryption, providing threshold signing (DKLsT23) for bridges, agents, or wallets[^26]. Acts more like a distributed signing service; does not define a .vult-style file format. Other DKLs implementers include Copper, BlockDaemon, Utila, Sodot, Visa, and Web3Auth[^27], mainly targeting institutional custody.

### 4.5 Summary Comparison

| Solution     | Target audience         | Key management model                        | File/standard          | Notable features                                                         |
| ------------ | ----------------------- | ------------------------------------------- | ---------------------- | ------------------------------------------------------------------------ |
| Vultisig     | Retail & small treasury | Self-custodial; shares in .vult files; DKLS | .vult (open, protobuf) | Seedless multi-device vaults, configurable threshold, emergency recovery |
| ZenGo        | Retail                  | 2-share MPC (client + server)               | Proprietary            | Seedless, easy recovery, server share                                    |
| Safeheron    | Enterprise              | 3-of-3 MPC-TEE; server-hosted               | None                   | Policy engine, APIs, open-sourced C++ TSS                                |
| Web3Auth     | Devs/dApps              | MPC + AA; hosted share                      | None                   | Fast creation, social logins, dApp focus                                 |
| Lit Protocol | Developers              | Decentralized DKLsT23 network               | None (service)         | Programmable signing/encryption, bridges, agents                         |

---

## 5. Security & Auditing

### 5.1 Audits

* **Kudelski audit of mobile-tss-lib (2024):** Vultisig commissioned Kudelski Security; all findings addressed in follow-up PRs. Full report not public; issues minor[^28].
* **Code4rena & Zenith audits:** Vultisig’s \$VULT token and staking contracts audited in June 2024, focusing on tokenomics, not TSS[^29].
* **Trail of Bits audit of DKLS23 (Feb 2024):** Silence Laboratories’ DKLS23 reference audited by ToB. 15 issues found: high-severity included nonce reuse (risk of key destruction), mishandling selective-abort, replay attacks\[^30]. All patched; highlights the complexity of TSS implementations.

### 5.2 Known Vulnerabilities or Incidents

No major exploits of Vultisig’s production wallets as of July 2025. Early DKLS23 implementations were vulnerable to key-destruction and selective-abort; all patched. Vultisig terms warn of unknown zero-days; regular backup and private ceremonies urged[^15].

### 5.3 Best Practices for Production

* **Stay up-to-date:** Use the latest libraries and update apps when notified.
* **Secure your devices:** No jailbreaking; use hardware enclaves when possible.
* **Separate vault shares:** Never keep enough shares together to satisfy the threshold[^9].
* **Enable remote wipe:** Protects against stolen device attacks[^14].
* **Private key ceremonies:** Use private settings for keygen/signing[^15].
* **Emergency recovery:** Keep offline copies of the recovery tool/instructions[^13].

---

## 6. Library Relationships

### 6.1 Silence Labs’ DKLS23 Library

Silence Labs’ `dkls23` Rust repo implements DKLS23 protocol. Vultisig uses this for keygen/signing, via a cross-platform wrapper and WebAssembly integration.

### 6.2 Vultisig mobile-tss-lib

Open-source library containing Go/Rust code that wraps DKLS and exposes keygen, signing, and vault management. CLI tools for recovery and threshold logic implemented in `common.go`[^10]. Used by all mobile/desktop Vultisig apps.

### 6.3 Other Dependencies

* `vultisig-relay-server` for device message relaying
* Flutter front-end (UI)
* `commondata` repo for protobuf definitions
* Chain-specific RPCs and cryptographic crates (curve25519-dalek, threshold\_crypto, etc.)

---

## 7. Open Questions & Future Directions

* **Custom thresholds:** Docs and code enforce 67% threshold; m-of-n hinted but not exposed. Modifying this requires library/code changes.
* **Role-based permissions:** Not yet in .vult; transaction policies “coming soon.”
* **Standardization:** No independent standard body; .vult is open but limited to Vultisig.
* **Proactive share refresh:** DKLS23 supports this in theory; not yet automated in Vultisig.
* **Interoperability:** .vult not compatible with other MPC wallets; a universal format or converter could help.
* **Regulatory:** Legal clarity needed for multi-party custody models.

---

## Conclusion

The `.vult` file format is one of the first open, consumer-oriented threshold signature containers. It removes seed phrases and eliminates single points of failure. The DKLS23 protocol behind `.vult` improves both security and usability, but robust audits and secure usage practices remain essential. Competing models (ZenGo, Safeheron, Web3Auth) each have trade-offs around custody, developer focus, and policy. `.vult`’s future as a universal standard depends on adoption, formalization, richer permissions, and ongoing security review.

---

### References

[^1]: Vault Backups | Vultisig
    [https://docs.vultisig.com/vultisig-vault-user-actions/managing-your-vault/vault-backup](https://docs.vultisig.com/vultisig-vault-user-actions/managing-your-vault/vault-backup)

[^2]: Creating a Vault | Vultisig
    [https://docs.vultisig.com/vultisig-vault-user-actions/creating-a-vault](https://docs.vultisig.com/vultisig-vault-user-actions/creating-a-vault)

[^3]: vault\_container.proto
    [https://github.com/vultisig/commondata/blob/dd4705e1345732eab864e6bf9e8b1454601cf063/proto/vultisig/vault/v1/vault\_container.proto](https://github.com/vultisig/commondata/blob/dd4705e1345732eab864e6bf9e8b1454601cf063/proto/vultisig/vault/v1/vault_container.proto)

[^4]: vault.proto
    [https://github.com/vultisig/commondata/blob/dd4705e1345732eab864e6bf9e8b1454601cf063/proto/vultisig/vault/v1/vault.proto](https://github.com/vultisig/commondata/blob/dd4705e1345732eab864e6bf9e8b1454601cf063/proto/vultisig/vault/v1/vault.proto)

[^5]: main.go
    [https://github.com/vultisig/mobile-tss-lib/blob/2e7e570a4a74c9d961f04593f228355e6b5f6adf/cmd/recovery-cli/main.go](https://github.com/vultisig/mobile-tss-lib/blob/2e7e570a4a74c9d961f04593f228355e6b5f6adf/cmd/recovery-cli/main.go)

[^6]: DKLS23 Protocol
    [https://raw.githubusercontent.com/vultisig/docs/c0df2bfde1da010693ec09a221ad6e50efd83761/threshold-signature-scheme/threshold-signature-schemes-used-by-vultisig/how-dkls23-works.md](https://raw.githubusercontent.com/vultisig/docs/c0df2bfde1da010693ec09a221ad6e50efd83761/threshold-signature-scheme/threshold-signature-schemes-used-by-vultisig/how-dkls23-works.md)

[^7]: Security notes
    [https://raw.githubusercontent.com/vultisig/docs/main/other/security.md](https://raw.githubusercontent.com/vultisig/docs/main/other/security.md)

[^8]: Ibid

[^9]: FAQ | Vultisig
    [https://docs.vultisig.com/other/faq](https://docs.vultisig.com/other/faq)

[^10]: common.go
    [https://github.com/vultisig/mobile-tss-lib/blob/2e7e570a4a74c9d961f04593f228355e6b5f6adf/tss/common.go](https://github.com/vultisig/mobile-tss-lib/blob/2e7e570a4a74c9d961f04593f228355e6b5f6adf/tss/common.go)

[^11]: Signing a Transaction | Vultisig
    [https://docs.vultisig.com/vultisig-vault-user-actions/signing-a-transaction/signing-a-transaction](https://docs.vultisig.com/vultisig-vault-user-actions/signing-a-transaction/signing-a-transaction)

[^12]: Transaction Policies | Vultisig
    [https://docs.vultisig.com/vultisig-infrastructure/what-is-vultisigner/what-can-be-configured](https://docs.vultisig.com/vultisig-infrastructure/what-is-vultisigner/what-can-be-configured)

[^13]: Emergency Recovery | Vultisig
    [https://docs.vultisig.com/threshold-signature-scheme/emergency-recovery](https://docs.vultisig.com/threshold-signature-scheme/emergency-recovery)

[^14]: FAQ | Vultisig
    [https://docs.vultisig.com/other/faq](https://docs.vultisig.com/other/faq)

[^15]: Terms | Vultisig
    [https://docs.vultisig.com/other/terms](https://docs.vultisig.com/other/terms)

[^16]: Signing guide
    [https://docs.vultisig.com/vultisig-vault-user-actions/signing-a-transaction/signing-a-transaction](https://docs.vultisig.com/vultisig-vault-user-actions/signing-a-transaction/signing-a-transaction)

[^17]: Ibid

[^18]: Security notes
    [https://raw.githubusercontent.com/vultisig/docs/main/other/security.md](https://raw.githubusercontent.com/vultisig/docs/main/other/security.md)

[^19]: MPC Wallet - What is MPC? | ZenGo
    [https://zengo.com/mpc-wallet/](https://zengo.com/mpc-wallet/)

[^20]: Safeheron Product
    [https://safeheron.com/product/mpc-self-custody/](https://safeheron.com/product/mpc-self-custody/)

[^21]: Ibid

[^22]: Ibid

[^23]: Open-source C++ TSS
    [https://safeheron.com/product/mpc-self-custody/](https://safeheron.com/product/mpc-self-custody/)

[^24]: Web3Auth - Key Management SDKs
    [https://web3auth.io/](https://web3auth.io/)

[^25]: Ibid

[^26]: Lit Protocol
    [https://www.litprotocol.com/](https://www.litprotocol.com/)

[^27]: DKLs.info
    [https://dkls.info/](https://dkls.info/)

[^28]: Security audit summary
    [https://raw.githubusercontent.com/vultisig/docs/main/other/security.md](https://raw.githubusercontent.com/vultisig/docs/main/other/security.md)

[^29]: Code4rena Audit
    [https://code4rena.com/audits/2024-06-vult](https://code4rena.com/audits/2024-06-vult)

