package vault

import (
	"encoding/hex"
	"sort"
	"strings"

	"github.com/btcsuite/btcd/btcutil/base58"
	"github.com/btcsuite/btcd/btcutil/hdkeychain"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/decred/dcrd/dcrec/secp256k1/v4"
	"github.com/rowbotony/vultool/internal/types"
)

// PathAddress represents a derivation path and the resulting address
type PathAddress struct {
	Chain       string `json:"chain"`
	Ticker      string `json:"ticker"`
	Address     string `json:"address"`
	DerivePath  string `json:"derive_path"`
	Description string `json:"description"`
	Purpose     string `json:"purpose"`
}

// DerivePathAddresses derives addresses for all common derivation paths from a vault
func DerivePathAddresses(vaultInfo *VaultInfo, pathsByChain map[types.SupportedChain][]types.DerivationPath, maxCount int) []PathAddress {
	var results []PathAddress

	// Process ECDSA-based chains (Bitcoin, Ethereum, etc.)
	if vaultInfo.PublicKeyECDSA != "" {
		results = append(results, deriveECDSAPathAddresses(vaultInfo.PublicKeyECDSA, vaultInfo.HexChainCode, pathsByChain)...)
	}

	// Process EdDSA-based chains (Solana, etc.)
	if vaultInfo.PublicKeyEDDSA != "" {
		results = append(results, deriveEdDSAPathAddresses(vaultInfo.PublicKeyEDDSA, pathsByChain)...)
	}

	// Sort by chain and path for consistent output
	sort.Slice(results, func(i, j int) bool {
		if results[i].Chain != results[j].Chain {
			return results[i].Chain < results[j].Chain
		}
		return results[i].DerivePath < results[j].DerivePath
	})

	// Apply count limit if specified
	if maxCount > 0 && len(results) > maxCount {
		results = results[:maxCount]
	}

	return results
}

// deriveECDSAPathAddresses derives addresses for ECDSA-based chains using their paths
func deriveECDSAPathAddresses(pubKeyHex string, chainCodeHex string, pathsByChain map[types.SupportedChain][]types.DerivationPath) []PathAddress {
	var results []PathAddress

	// Define chain mappings for ECDSA chains
	echainMappings := map[types.SupportedChain]struct{
		chain  string
		ticker string
	}{
		types.ChainBitcoin:     {"Bitcoin", "BTC"},
		types.ChainBitcoinCash: {"Bitcoin-Cash", "BCH"},
		types.ChainLitecoin:    {"Litecoin", "LTC"},
		types.ChainDogecoin:    {"Dogecoin", "DOGE"},
		types.ChainDash:        {"Dash", "DASH"},
		types.ChainZcash:       {"Zcash", "ZEC"},
		types.ChainEthereum:    {"Ethereum", "ETH"},
		types.ChainBSC:         {"BSC", "BSC"},
		types.ChainAvalanche:   {"Avalanche", "AVAX"},
		types.ChainPolygon:     {"Polygon", "MATIC"},
		types.ChainCronosChain: {"CronosChain", "CRO"},
		types.ChainArbitrum:    {"Arbitrum", "ETH"},
		types.ChainOptimism:    {"Optimism", "ETH"},
		types.ChainBase:        {"Base", "ETH"},
		types.ChainBlast:       {"Blast", "ETH"},
		types.ChainZksync:      {"Zksync", "ETH"},
		types.ChainThorChain:   {"THORChain", "RUNE"},
	}

	// Process all ECDSA chains
	for chainType, mapping := range echainMappings {
		if paths, ok := pathsByChain[chainType]; ok {
			for _, path := range paths {
				addr := deriveECDSAAddressForPath(pubKeyHex, chainCodeHex, path.Path, mapping.chain, mapping.ticker)
				if addr != "" && !strings.HasPrefix(addr, "error:") {
					results = append(results, PathAddress{
						Chain:       mapping.chain,
						Ticker:      mapping.ticker,
						Address:     addr,
						DerivePath:  path.Path,
						Description: path.Description,
						Purpose:     path.Purpose,
					})
				}
			}
		}
	}

	return results
}

// deriveEdDSAPathAddresses derives addresses for EdDSA-based chains using their paths
func deriveEdDSAPathAddresses(pubKeyHex string, pathsByChain map[types.SupportedChain][]types.DerivationPath) []PathAddress {
	var results []PathAddress

	// Process Solana paths
	if paths, ok := pathsByChain[types.ChainSolana]; ok {
		for _, path := range paths {
			// For Solana, we currently just use the master key directly (no HD derivation)
			// In a future version, we could implement proper ed25519 HD derivation
			addr := deriveSolanaAddress(pubKeyHex)
			if addr != "" {
				results = append(results, PathAddress{
					Chain:       "Solana",
					Ticker:      "SOL",
					Address:     addr,
					DerivePath:  path.Path,
					Description: path.Description,
					Purpose:     path.Purpose,
				})
			}
		}
	}

	// Process SUI paths
	if paths, ok := pathsByChain[types.ChainSUI]; ok {
		for _, path := range paths {
			// For SUI, we currently use hex encoding as a placeholder
			// In a future version, we could implement proper SUI address derivation
			addr := deriveSUIAddress(pubKeyHex)
			if addr != "" {
				results = append(results, PathAddress{
					Chain:       "SUI",
					Ticker:      "SUI",
					Address:     addr,
					DerivePath:  path.Path,
					Description: path.Description,
					Purpose:     path.Purpose,
				})
			}
		}
	}

	return results
}

// deriveECDSAAddressForPath derives an address for a specific chain and path
func deriveECDSAAddressForPath(pubKeyHex string, chainCodeHex string, derivePath string, chain string, ticker string) string {
	// Get the extended key
	extendedKey, err := createExtendedPublicKey(pubKeyHex, chainCodeHex)
	if err != nil {
		return "error: " + err.Error()
	}

	// Derive child key at the specified path
	childPubKey := deriveChildPublicKey(extendedKey, derivePath)
	if childPubKey == nil {
		return "error: failed to derive child key"
	}

	// Encode address for the specific chain
	switch chain {
	// Bitcoin and Bitcoin-like chains
	case "Bitcoin":
		// Check the derivation path to determine address format
		if strings.Contains(derivePath, "m/84'") {
			return deriveBitcoinSegwitAddress(childPubKey) // Native SegWit
		} else if strings.Contains(derivePath, "m/49'") {
			return deriveBitcoinP2SHSegwitAddress(childPubKey) // P2SH-wrapped SegWit
		} else {
			return deriveBitcoinLegacyAddress(childPubKey) // Legacy P2PKH
		}
	case "Bitcoin-Cash":
		return deriveBitcoinCashAddress(childPubKey)
	case "Litecoin":
		if strings.Contains(derivePath, "m/84'") {
			return deriveLitecoinSegwitAddress(childPubKey)
		} else {
			return deriveLitecoinLegacyAddress(childPubKey)
		}
	case "Dogecoin":
		return deriveDogecoinAddress(childPubKey)
	case "Dash":
		return deriveDashAddress(childPubKey)
	case "Zcash":
		return deriveZcashAddress(childPubKey)
	// Ethereum and all EVM chains use the same address format
	case "Ethereum", "BSC", "Avalanche", "Polygon", "CronosChain", "Arbitrum", "Optimism", "Base", "Blast", "Zksync":
		return deriveEthereumAddress(childPubKey)
	case "THORChain":
		return deriveThorchainAddress(childPubKey)
	default:
		return "error: unsupported chain " + chain
	}
}

// Helper function to create an extended public key from hex
func createExtendedPublicKey(pubKeyHex string, chainCodeHex string) (*hdkeychain.ExtendedKey, error) {
	// Decode the hex public key
	pubKeyBytes, err := hex.DecodeString(pubKeyHex)
	if err != nil {
		return nil, err
	}

	// Parse the public key
	pubKey, err := secp256k1.ParsePubKey(pubKeyBytes)
	if err != nil {
		return nil, err
	}

	// Decode chain code
	chainCodeBytes, err := hex.DecodeString(chainCodeHex)
	if err != nil {
		// Use zero chain code if not provided
		chainCodeBytes = make([]byte, 32)
	}

	// Create extended public key for HD derivation
	net := &chaincfg.MainNetParams
	extendedKey := hdkeychain.NewExtendedKey(
		net.HDPublicKeyID[:],
		pubKey.SerializeCompressed(),
		chainCodeBytes,
		[]byte{0x00, 0x00, 0x00, 0x00},
		0,
		0,
		false, // isPrivate = false for public key
	)

	return extendedKey, nil
}

// Helper function for Solana address derivation
func deriveSolanaAddress(pubKeyHex string) string {
	pubKeyBytes, err := hex.DecodeString(pubKeyHex)
	if err != nil {
		return "error: " + err.Error()
	}

	// Solana uses base58 encoding of the public key directly
	return base58.Encode(pubKeyBytes)
}

// Additional Bitcoin address derivation functions
func deriveBitcoinLegacyAddress(pubKey *secp256k1.PublicKey) string {
	// P2PKH address using version byte 0x00
	pubKeyCompressed := pubKey.SerializeCompressed()
	hash160 := hash160(pubKeyCompressed)
	return base58.CheckEncode(hash160, 0x00)
}

func deriveBitcoinP2SHSegwitAddress(pubKey *secp256k1.PublicKey) string {
	// P2SH-wrapped SegWit (P2WPKH-in-P2SH)
	pubKeyCompressed := pubKey.SerializeCompressed()
	pubkeyHash := hash160(pubKeyCompressed)
	
	// Create witness program: version + hash160
	witnessProgram := append([]byte{0x00, 0x14}, pubkeyHash...)
	// Hash the witness program for P2SH
	scriptHash := hash160(witnessProgram)
	// Return P2SH address with version byte 0x05
	return base58.CheckEncode(scriptHash, 0x05)
}

func deriveLitecoinLegacyAddress(pubKey *secp256k1.PublicKey) string {
	// Litecoin P2PKH uses version byte 0x30 (48)
	pubKeyCompressed := pubKey.SerializeCompressed()
	hash160 := hash160(pubKeyCompressed)
	return base58.CheckEncode(hash160, 0x30)
}

// Helper function for SUI address derivation
func deriveSUIAddress(pubKeyHex string) string {
	// SUI uses a special format (0x + blake2b hash)
	// For now, using hex encoding as placeholder until proper SUI derivation is implemented
	return "0x" + pubKeyHex
}
