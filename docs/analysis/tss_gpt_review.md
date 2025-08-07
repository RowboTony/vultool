Based on your code and the accompanying documentation, the current **DKLS recovery logic in `vultool` does not implement a true TSS key‐combination ceremony**.  Instead, it applies heuristic “key hunting” over the raw keyshare bytes.  Here’s what is happening and why it leads to the mismatches you’re seeing:

1. **GG20 vs DKLS formats** – The docs highlight that Vultisig has two schemes: GG20 (legacy) and DKLS (new).  GG20 stores `LocalState` as JSON; DKLS stores binary protobuf data.  The DKLS recovery code must decode the binary format and then run the correct reconstruction algorithm.

2. **Heuristic “DKLS” recovery** – Because the DKLS protobuf schema is unknown, the current vultool code adopts heuristic methods inspired by the community “Share Decoder” tool.  These methods search each keyshare individually for high-entropy 32‑byte chunks or certain protobuf patterns (e.g., `0x12 0x20`, `0x1a 0x20`) and treat those as potential private keys.  They do **not** combine the shares via Lagrange interpolation; instead they look at each share separately and try XOR or hash tricks.  This is why each share on its own appears to produce its “own” key and address and why combining shares in vultool sometimes yields a key that does not match the vault’s public address.

3. **False positives by design** – The docs explicitly warn that the heuristic methods recover “valid keys” but that the resulting addresses **do not necessarily match the official Vultisig address**, implying that proper DKLS cryptographic reconstruction is still missing.  The enhanced `dkls_enhanced.go` file still loops through each share individually, scanning for patterns, rather than combining the shares.

4. **How real DKLS/TSS should work** – In a true threshold scheme, no single share contains the private key.  Each share holds a piece of secret `Xi`; only by interpolating at least `k` shares (using the protocol’s Lagrange coefficients) can the private key be reconstructed.  That’s why Vultisig’s own CLI or Share Decoder recombines multiple shares and yields the correct address (e.g., `bc1q0pap5…`) while a single share imported into a wallet produces a different “fallback” key.

### Suggestions / next steps

* **Use a proper DKLS implementation**.  Either integrate Vultisig’s own `mobile-tss-lib` for DKLS or port the DKLS reconstruction logic from the Vultisig‑Share‑Decoder (which includes the actual combine routine).  Relying on pattern-matching in raw keyshare bytes will always risk false positives.

* **Parse the DKLS keyshare structure instead of heuristics**.  The docs note that the keyshare field may contain base64‑encoded binary protobuf rather than JSON.  Once you have the schema, you can decode the `LocalState` and extract the `Xi` shares.  Those can then be recombined via verifiable secret sharing (VSS) and used with the mobile‑tss‑lib ECDSA/EdDSA routines.

* **Validate against vault metadata**.  After reconstructing a key, derive the expected addresses (`1…`/`bc1…`/`0x…`) and compare them to the vault’s stored public keys.  If the addresses don’t match, discard that candidate and try a different offset or method.

* **Short-term workaround** – If you just want to get correct DKLS recovery now, you can call into the Share Decoder or the Vultisig CLI from within vultool (e.g., via `exec.Command`) until a proper Go implementation is ready.  That library already produces the `bc1q0pap5…` address you expect.

In summary, the current vultool DKLS “recovery” doesn’t perform the true MPC/TSS reconstruction.  It finds plausible private keys by scanning each share’s data, leading to addresses that don’t match the vault.  To fix this, you’ll need to replace the heuristic scanning with real DKLS share decoding and TSS recombination.
