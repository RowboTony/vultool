# `spec.md` – **vultool Command-Line Utility**

## 0 . Executive Summary

`vultool` is a security-first, developer-oriented CLI for creating, inspecting, manipulating, and recovering **`.vult` threshold-share files** used by the Vultisig MPC wallet stack.
It aims to expose *all* safe, native features of the underlying DKLS23 / GG20 libraries (ECDSA **and** EdDSA), while laying the groundwork for a future GUI (Web) that will reuse the exact same Go modules.

---

## 1 . Goals & Non-Goals

|                     | **In scope (v1.x)**                                                 | **Out of scope (for now)**                   |
| ------------------- | ------------------------------------------------------------------- | -------------------------------------------- |
| **File operations** | parse, verify, encrypt/decrypt, rename, diff                        | Editing raw key-share material by hand       |
| **Key ceremonies**  | `keygen`, `reshare`, `migrate GG20→DKLS23`, proactive **refresh**   | BLS / FROST Schnorr (RFC planned for v2)     |
| **Signing**         | Threshold ECDSA & EdDSA, batch, QR / relay session                  | On-chain fee estimation & TX building        |
| **Recovery**        | Multi-file recombination, HD derivation, multi-chain export         | Auto-sweeping funds after export             |
| **Seed import**     | *Optional*, gated behind `--experimental` with strong warnings      | Automatic seed-phrase → TSS splitting in GUI |
| **Transport**       | ASCII QR, local relay API, stdin/stdout pipes                       | BLE / NFC / deep-link transport              |
| **Networking**      | Single-binary **local VultiServer API** for dev/test; prod optional | Public hosted relay infrastructure           |
| **Security**        | Memory-zeroing, Argon2id encryption, reproducible builds            | HSM / SGX attestation                        |

---

## 2 . High-Level Architecture

```
                +--------------------+
                |    vultool CLI     |
                +---------+----------+
                          |
          +---------------+--------------+
          |                              |
 +--------v---------+        +-----------v-----------+
 |  Core Go libs    |        |  Local VultiServer    |  (optional)
 |  (mobile-tss-lib |        |  – HTTP JSON API      |
 |   + dkls23.FFI)  |        |  – WebSocket relay    |
 +--------+---------+        +-----------+-----------+
          |                               \
          |                                \  (QR / stdin)
 +--------v---------+                 +-----v------+
 |   .vult files    |<--base64/protobuf-->+ Devices  |
 +------------------+                 +------------+
```

* **Language stack:** 100 % Go for CLI + embedded HTTP server; CGO / WASM bindings already provided by `mobile-tss-lib`.
* **Cryptography:** DKLS23 (preferred) and GG20 via audited upstream libs; Ed25519 support piggy-backs on DKLS23 EdDSA mode.
* **Transport adapters:** implemented behind a `transport.Interface` (QR, relay, local files); easily extended.

---

## 3 . Command Surface (MVP → v1.0)

> **Implementation Status Table**  
> `[EXISTS]`: Present in main branch (`inspect`)  
> `[ALIAS]`: Alias to existing verb/flag (to be added in v0.1)  
> `[PLANNED]`: Not in codebase yet, on roadmap  
> `[EXPERIMENTAL]`: Gated behind build tag/flag

| Command         | Status       | Implementation/Notes                                                      |
|-----------------|-------------|---------------------------------------------------------------------------|
| `inspect`       | [EXISTS]     | Rich metadata/validation, `--summary`, `--export-file`                    |
| `info`          | [ALIAS]      | Alias to `inspect --summary` (to be added)                                |
| `decode`        | [ALIAS]      | Alias to `inspect --json` / `--yaml` (to be added)                        |
| `verify`        | [ALIAS]      | Alias to `inspect --validate` (to be added)                               |
| `diff`          | [PLANNED]    | Compare two vaults (metadata/share CRCs)                                  |
| `set-password`  | [PLANNED]    | Argon2id + AES-GCM re-encrypt                                             |
| `remove-password`| [PLANNED]   | Strip encryption                                                          |
| `change-password`| [PLANNED]   | Wrapper: decrypt→encrypt                                                  |
| `keygen`        | [PLANNED]    | Distributed DKLS23 DKG (embedded relay)                                   |
| `reshare`       | [PLANNED]    | Add/remove parties, new reshare prefix                                    |
| `refresh`       | [PLANNED]    | Proactive share refresh (same roster)                                     |
| `migrate`       | [PLANNED]    | GG20 → DKLS23 upgrade                                                     |
| `change-threshold`| [PLANNED]  | Custom t-of-n (experimental)                                              |
| `sign`          | [PLANNED]    | Interactive threshold signing                                             |
| `batch-sign`    | [PLANNED]    | CSV/JSON batch signing                                                    |
| `qr-session`    | [PLANNED]    | ASCII QR multi-device session                                             |
| `recover`       | [PLANNED]    | Combine ≥t shares → WIF/hex/base58                                        |
| `derive`        | [PLANNED]    | Read-only HD derivation                                                   |
| `list-addresses`| [PLANNED]    | Enumerate common HD paths                                                 |
| `import-seed`   | [EXPERIMENTAL]| BIP39/private-key → .vult shares (build tag)                             |
| `export`        | [ALIAS]      | `inspect --export-file` (already exists)                                  |

### 3.1 Informational

| Command  | Args / Flags        | Output & Notes                                                             | Status     |
| -------- | ------------------- | -------------------------------------------------------------------------- | ---------- |
| `info`   | `<file.vult>`       | Human summary inc. protocol, ECDSA/EdDSA presence, threshold, signer count | [ALIAS]    |
| `decode` | `--json`, `--yaml`| Full protobuf dump                                                        | [ALIAS]    |
| `verify` | `<file>`            | Structural + cryptographic sanity checks; exit 0/1                         | [ALIAS]    |
| `diff`   | `<a.vult> <b.vult>` | Colored diff of metadata & share checksums                                 | [PLANNED]  |
| `inspect`| `<file.vult>`       | Full detail, validation, keyshares, export, etc                            | [EXISTS]   |

*Note: All new informational commands will be implemented as CLI aliases/wrappers for the existing `inspect` logic.*

### 3.2 Security / Passwords

| Command           | Behaviour                                         | Status      |
| ----------------- | ------------------------------------------------- | ----------- |
| `set-password`    | Re-encrypt with new password (Argon2id + AES-GCM) | [PLANNED]   |
| `remove-password` | Strip encryption                                  | [PLANNED]   |
| `change-password` | Convenience wrapper                               | [PLANNED]   |

### 3.3 Vault Lifecycle

| Command            | Key flags                                                | Description                                   | Status     |
| ------------------ | -------------------------------------------------------- | --------------------------------------------- | ---------- |
| `keygen`           | `--participants`, `--threshold`, `--name`, `--chaincode` | Runs DKG; outputs one `.vult` per participant | [PLANNED]  |
| `reshare`          | `--old`, `--new`, `--threshold?`                         | Generates fresh shares / reshare prefix       | [PLANNED]  |
| `migrate`          | `--in GG20.vult`                                         | GG20 → DKLS23 upgrade (enables EdDSA)         | [PLANNED]  |
| `refresh`          | *DKLS only*                                              | Proactive share refresh w/out roster change   | [PLANNED]  |
| `change-threshold` | **experimental** custom m-of-n                           |                                               | [PLANNED]  |

### 3.4 Signing

| Command      | Mode                        | Notes                              | Status     |
| ------------ | --------------------------- | ---------------------------------- | ---------- |
| `sign`       | auto-detect scheme by vault | One message/tx                     | [PLANNED]  |
| `batch-sign` | `--file csv`, `json`     | Multiple msgs                      | [PLANNED]  |
| `qr-session` | interactive                 | Displays ASCII QR, waits for peers | [PLANNED]  |

### 3.5 Recovery / Derivation

| Command          | Result                                                                                  | Status     |
| ---------------- | --------------------------------------------------------------------------------------- | ---------- |
| `recover`        | Reconstructs private keys **&** derives BTC (WIF), ETH (hex), Solana/THOR (base58) etc. | [PLANNED]  |
| `derive`         | Read-only pub/addr derivation from single share                                         | [PLANNED]  |
| `list-addresses` | Common HD paths per chain                                                               | [PLANNED]  |

### 3.6 Import (controversial)

`import-seed` **--experimental** [EXPERIMENTAL]  
*Splits a BIP-39 seed or raw privkey into t-of-n shares.* Emits strong warnings & suggests sweeping funds into a fresh DKLS vault ASAP.

---

## 4 . Implementation Phases & Milestones

| Milestone           | Features                                                                  | Success Criteria                               |
| ------------------- | ------------------------------------------------------------------------- | ---------------------------------------------- |
| **0.1 “Inspector”** | `inspect` (EXISTS), plus aliases: `info`, `decode`, `verify`, implement `diff` | Aliases work, CI passes, help/README clear     |
| **0.2 “Medic”**     | `recover`, `derive`, `list-addresses`                                     | End-to-end recovery of BTC & ETH keys          |
| **0.3 “Creator”**   | `keygen`, `reshare`, QR transport (offline)                               | Two-laptop demo produces valid vault & signs   |
| **0.4 “Networker”** | Embedded local VultiServer, relay WS transport                            | Keys generated & signed across LAN             |
| **1.0 “Pro”**       | `migrate`, `refresh`, `batch-sign`, seed import (exp)                     | SemVer-stable, reproducible release builds     |

---

## 5 . Security Requirements

1. **Memory hygiene** – zero out decrypted share buffers; avoid heap copies.
2. **Reproducible builds** – `go build -trimpath` verified via CI.
3. **FIPS-level crypto** – use upstream audited libs only.
4. **No silent overwrite** – all destructive ops require `--force`.
5. **Entropy** – rely on OS CSPRNG only; fail fast otherwise.
6. **Seed import gating** – compile-time `build tag experimental` + runtime `--i-understand-the-risk`.

---

## 6 . CI / CD & Tooling

* **GitHub Actions**: lint → unit → integration (dockerised relay) → regression (sign & verify on BTC, ETH, SOL test vectors) → build matrix (linux/mac/win, amd64+arm64) → release binaries + SHA256.
* **Static analysis**: `go vet`, `staticcheck`, `gosec`, `golangci-lint`.
* **Fuzzing**: protobuf unmarshal, password decrypt, share merge.
* **Docs**: auto-publish `godoc` + man-pages (`cobra` + `gendoc`).
* **Homebrew / Scoop / APK** formulas once v1.0 tagged.

---

## 7 . Dependencies

| Purpose                | Repo / Module                                  |
| ---------------------- | ---------------------------------------------- |
| Threshold cryptography | `github.com/vultisig/mobile-tss-lib` (Go)      |
| DKLS23 core            | `github.com/silence-laboratories/dkls23` (cgo) |
| QR encode              | `github.com/skip2/go-qrcode`                   |
| Argon2id               | `golang.org/x/crypto/argon2`                   |
| CLI framework          | `github.com/spf13/cobra`                       |
| Config / logging       | `spf13/viper`, `rs/zerolog`                    |

---

## 8 . Extensibility Hooks

* **`transport.Interface`** – add BLE/NFC later.
* **`vault.Policy`** – placeholder struct for future per-transaction rules.
* **Plugin system** – Go interfaces à la `Hashicorp/go-plugin` (v2).

---

## 9 . Open Questions

1. Should `vultool` auto-detect & refuse >1 share stored on the same disk?
2. Which chains beyond BTC/ETH/SOL should ship in core vs plugin?
3. Strategy for deterministic QR size limits (chunking large `.vult` files).
4. How to version future protobuf changes without breaking old shares?

---

## 10 . Licensing

Apache-2.0, matching upstream Vultisig repositories.
Third-party libs retain their respective OSS licenses; all compatible.

---

## 11 . Contributing

* Sign the **DCO**.
* All new features **must** include unit tests + docs.
* Security issues → `security@TBD` (public key in repo).

---

## 12 . Appendix – Example Quickstart

```bash
# Inspect a vault (detailed, power-user; exists now)
vultool inspect --vault alice.vult

# Show concise info (alias; to be added)
vultool info alice.vult

# JSON/YAML dump (alias; to be added)
vultool decode --json alice.vult

# Validate vault integrity (alias; to be added)
vultool verify alice.vult

# Compare two backups (planned)
vultool diff alice.vult backup.vult

# All other advanced features per milestone table.
```

---

### `ROADMAP.md` (concise)

```
0.1 Inspector  – inspect/aliases/diff/passwords
0.2 Medic      – recovery & derivation (BTC/ETH/SOL)
0.3 Creator    – keygen/reshare + QR transport
0.4 Networker  – local VultiServer API, LAN relay
0.5 Polisher   – UX, shell completions, man pages
1.0 Stable     – migrate, refresh, Windows MSI, Homebrew
```

*Pure grind mode - milestones ship when they're ready.*

---