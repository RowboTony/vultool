package recovery

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/btcutil/base58"
	"github.com/btcsuite/btcd/btcutil/bech32"
	btcchaincfg "github.com/btcsuite/btcd/chaincfg"
	"github.com/decred/dcrd/dcrec/secp256k1/v4"
	bchchaincfg "github.com/gcash/bchd/chaincfg"
	"github.com/gcash/bchutil"
	"golang.org/x/crypto/ripemd160"
	"golang.org/x/crypto/sha3"

	"github.com/rowbotony/vultool/internal/vault"
)

// VerifyResult summarizes verification for one chain
type VerifyResult struct {
	Chain    string
	Expected string
	Computed string
	Passed   bool
	Skipped  bool
	Error    string
}

// VerifyRecoveredKeys validates that each recovered private key reproduces the expected address
// Only ECDSA chains are verified in v1. EdDSA chains are reported as SKIPPED.
func VerifyRecoveredKeys(vaultFiles []string, keys []RecoveredKey, password string) ([]VerifyResult, error) {
	results := make([]VerifyResult, 0, len(keys))
	if len(vaultFiles) == 0 {
		return results, fmt.Errorf("no vault files provided for verification")
	}

	// Parse vault for expected addresses
	v, err := vault.ParseVaultFileWithPassword(vaultFiles[0], password)
	if err != nil {
		return results, fmt.Errorf("failed to parse vault for verification: %w", err)
	}
	expected := vault.DeriveAddressesFromVault(v)
	expectedByChain := map[string]string{}
	for _, ea := range expected {
		expectedByChain[strings.ToLower(ea.Chain)] = ea.Address
	}

	var fails int
	for _, k := range keys {
		chain := strings.ToLower(string(k.Chain))

		// Skip EdDSA in v1
		if chain == "solana" || chain == "sui" {
			results = append(results, VerifyResult{Chain: chain, Skipped: true, Error: "EdDSA verification planned for v2"})
			continue
		}

		exp := expectedByChain[chain]
		if exp == "" {
			results = append(results, VerifyResult{Chain: chain, Passed: false, Error: "no expected address for chain"})
			fails++
			continue
		}

		computed, cerr := computeAddressFromPrivate(chain, k.PrivateKey)
		if cerr != nil {
			results = append(results, VerifyResult{Chain: chain, Expected: exp, Computed: "", Passed: false, Error: cerr.Error()})
			fails++
			continue
		}

		passed := strings.EqualFold(exp, computed)
		if !passed {
			fails++
		}
		results = append(results, VerifyResult{Chain: chain, Expected: exp, Computed: computed, Passed: passed})
	}

	if fails > 0 {
		return results, fmt.Errorf("verification failed for %d chain(s)", fails)
	}
	return results, nil
}

// computeAddressFromPrivate recomputes the address for a chain from a hex private key
func computeAddressFromPrivate(chain string, privHex string) (string, error) {
	b, err := hex.DecodeString(strings.TrimPrefix(privHex, "0x"))
	if err != nil {
		return "", fmt.Errorf("invalid private key hex: %w", err)
	}
	if len(b) != 32 {
		return "", fmt.Errorf("invalid private key length: %d", len(b))
	}
	priv := secp256k1.PrivKeyFromBytes(b)
	pub := priv.PubKey()

	switch chain {
	case "ethereum", "arbitrum", "avalanche", "bsc", "base", "blast", "cronoschain", "optimism", "polygon", "zksync":
		return ethAddressFromPub(pub), nil
	case "bitcoin":
		return btcSegwitAddress(pub), nil
	case "litecoin":
		return ltcSegwitAddress(pub), nil
	case "dogecoin":
		return dogeP2PKH(pub), nil
	case "dash":
		return dashP2PKH(pub), nil
	case "zcash":
		return zecTransparent(pub), nil
	case "bitcoin-cash":
		return bchCashAddr(pub)
	case "thorchain":
		return thorBech32(pub)
	default:
		return "", fmt.Errorf("unsupported chain for verification: %s", chain)
	}
}

// Helpers
func hash160(data []byte) []byte {
	sha := sha256.Sum256(data)
	ripemd := ripemd160.New()
	ripemd.Write(sha[:])
	return ripemd.Sum(nil)
}

func ethAddressFromPub(pub *secp256k1.PublicKey) string {
	uncompressed := pub.SerializeUncompressed()
	if len(uncompressed) == 65 && uncompressed[0] == 0x04 {
		uncompressed = uncompressed[1:]
	}
	h := sha3.NewLegacyKeccak256()
	h.Write(uncompressed)
	sum := h.Sum(nil)
	return "0x" + hex.EncodeToString(sum[12:])
}

func btcSegwitAddress(pub *secp256k1.PublicKey) string {
	compressed := pub.SerializeCompressed()
	h160 := hash160(compressed)
	addr, err := btcutil.NewAddressWitnessPubKeyHash(h160, &btcchaincfg.MainNetParams)
	if err != nil {
		return "error: " + err.Error()
	}
	return addr.EncodeAddress()
}

func ltcSegwitAddress(pub *secp256k1.PublicKey) string {
	compressed := pub.SerializeCompressed()
	h160 := hash160(compressed)
	params := btcchaincfg.MainNetParams
	params.Bech32HRPSegwit = "ltc"
	addr, err := btcutil.NewAddressWitnessPubKeyHash(h160, &params)
	if err != nil {
		return "error: " + err.Error()
	}
	addrStr := addr.EncodeAddress()
	if strings.HasPrefix(addrStr, "bc1") {
		return "ltc1" + addrStr[3:]
	}
	return addrStr
}

func dogeP2PKH(pub *secp256k1.PublicKey) string {
	compressed := pub.SerializeCompressed()
	h160 := hash160(compressed)
	return base58.CheckEncode(h160, 0x1E)
}

func dashP2PKH(pub *secp256k1.PublicKey) string {
	compressed := pub.SerializeCompressed()
	h160 := hash160(compressed)
	return base58.CheckEncode(h160, 0x4C)
}

func zecTransparent(pub *secp256k1.PublicKey) string {
	compressed := pub.SerializeCompressed()
	h160 := hash160(compressed)
	payload := append([]byte{0x1C, 0xB8}, h160...)
	chk := sha256.Sum256(payload)
	chk = sha256.Sum256(chk[:])
	full := append(payload, chk[:4]...)
	return base58.Encode(full)
}

func bchCashAddr(pub *secp256k1.PublicKey) (string, error) {
	compressed := pub.SerializeCompressed()
	h160 := hash160(compressed)
	addr, err := bchutil.NewAddressPubKeyHash(h160, &bchchaincfg.MainNetParams)
	if err != nil {
		return "", err
	}
	return addr.EncodeAddress(), nil
}

func thorBech32(pub *secp256k1.PublicKey) (string, error) {
	compressed := pub.SerializeCompressed()
	h160 := hash160(compressed)
	conv, err := bech32.ConvertBits(h160, 8, 5, true)
	if err != nil {
		return "", err
	}
	addr, err := bech32.Encode("thor", conv)
	if err != nil {
		return "", err
	}
	return addr, nil
}