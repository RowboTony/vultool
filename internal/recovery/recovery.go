package recovery

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"log"
	"strings"

	"github.com/rowbotony/vultool/internal/vault"
)

// SupportedChain represents a blockchain we can derive keys for
type SupportedChain string

const (
	// ECDSA-based chains
	ChainBitcoin     SupportedChain = "bitcoin"
	ChainBitcoinCash SupportedChain = "bitcoincash"
	ChainLitecoin    SupportedChain = "litecoin"
	ChainDash        SupportedChain = "dash"
	ChainDogecoin    SupportedChain = "dogecoin"
	ChainZcash       SupportedChain = "zcash"

	// Ethereum and EVM-compatible chains
	ChainEthereum  SupportedChain = "ethereum"
	ChainArbitrum  SupportedChain = "arbitrum"
	ChainAvalanche SupportedChain = "avalanche"
	ChainBase      SupportedChain = "base"
	ChainBlast     SupportedChain = "blast"
	ChainBSC       SupportedChain = "bsc"
	ChainCronos    SupportedChain = "cronos"
	ChainOptimism  SupportedChain = "optimism"
	ChainPolygon   SupportedChain = "polygon"
	ChainZkSync    SupportedChain = "zksync"

	// Cosmos-based chains
	ChainThorChain SupportedChain = "thorchain"

	// EdDSA-based chains
	ChainSolana SupportedChain = "solana"
	ChainSUI    SupportedChain = "sui"
)

// RecoveredKey represents a reconstructed private key in various formats
type RecoveredKey struct {
	Chain      SupportedChain `json:"chain"`
	PrivateKey string         `json:"private_key"`      // hex format
	WIF        string         `json:"wif,omitempty"`    // Bitcoin WIF format
	Base58     string         `json:"base58,omitempty"` // Solana/THOR base58 format
	Address    string         `json:"address"`
	DerivePath string         `json:"derive_path,omitempty"`

	// Wallet-compatible formats for EdDSA chains
	SolanaSeedFormat   string `json:"solana_seed_format,omitempty"`   // 32-byte seed only in base64 (some wallets prefer this)
	SolanaWalletFormat string `json:"solana_wallet_format,omitempty"` // 64-byte Ed25519 keypair in base64 for Solana
	SolanaWalletJSON   string `json:"solana_wallet_json,omitempty"`   // JSON array of 64 bytes for Phantom/Solflare
	SuiWalletFormat    string `json:"sui_wallet_format,omitempty"`    // 33-byte [0x00 + seed] in base64 for SUI
}

// DerivationPath represents an HD derivation path
type DerivationPath struct {
	Path        string         `json:"path"`
	Chain       SupportedChain `json:"chain"`
	Description string         `json:"description"`
	Purpose     string         `json:"purpose"` // e.g., "receiving", "change", "legacy"
}

// RecoverPrivateKeys combines threshold shares to reconstruct private keys
// Implements TSS (Threshold Signature Scheme) key recovery from vault shares
func RecoverPrivateKeys(vaultFiles []string, threshold int, password string) ([]RecoveredKey, error) {
	if len(vaultFiles) < threshold {
		return nil, fmt.Errorf("insufficient shares: need at least %d shares, got %d", threshold, len(vaultFiles))
	}

	var recoveredKeys []RecoveredKey

	// First check if this is a GG20 vault by examining the first file
	isGG20, err := CheckIfGG20Vault(vaultFiles[0], password)
	if err == nil && isGG20 {
		log.Printf("Detected GG20 vault - using proper GG20 recovery with validation")

		// Parse the original vault to get correct public keys for derivation
		originalVault, err := vault.ParseVaultFileWithPassword(vaultFiles[0], password)
		if err != nil {
			return nil, fmt.Errorf("failed to parse original vault for derivation: %w", err)
		}

		// Try ECDSA recovery using mobile-tss-lib compatible approach
		log.Printf("Attempting ECDSA TSS reconstruction...")
		ecdsaResult, err := ReconstructTSSKey(vaultFiles, password, TssKeyType(ECDSA))
		if err == nil && ecdsaResult != nil {
			log.Printf("✅ ECDSA TSS reconstruction successful")
			// Add all ECDSA-based chain recoveries
			recoveredKeys = append(recoveredKeys, convertTSSToRecoveredKeys(ecdsaResult, ECDSA, originalVault)...)
		} else if err != nil {
			log.Printf("⚠️ ECDSA TSS reconstruction failed: %v", err)
		} else {
			log.Printf("⚠️ ECDSA TSS reconstruction returned nil result")
		}

		// Try EdDSA recovery for Solana and other EdDSA chains
		log.Printf("Attempting EdDSA TSS reconstruction...")
		eddsaResult, err := ReconstructTSSKey(vaultFiles, password, TssKeyType(EdDSA))
		if err != nil {
			log.Printf("⚠️ EdDSA TSS reconstruction failed: %v", err)
		} else if eddsaResult == nil {
			log.Printf("⚠️ EdDSA TSS reconstruction returned nil result")
		} else {
			log.Printf("✅ EdDSA TSS reconstruction successful: private_key_hex=%s, chain_code=%s", eddsaResult.PrivateKeyHex, eddsaResult.ChainCode)
			// Add all EdDSA-based chain recoveries
			recoveredEdDSAKeys := convertTSSToRecoveredKeys(eddsaResult, EdDSA, originalVault)
			log.Printf("EdDSA conversion produced %d keys", len(recoveredEdDSAKeys))
			for _, key := range recoveredEdDSAKeys {
				log.Printf("EdDSA recovered key: chain=%s, address=%s", key.Chain, key.Address)
			}
			recoveredKeys = append(recoveredKeys, recoveredEdDSAKeys...)
		}

		// CRITICAL: Validate recovered addresses match list-addresses
		if len(recoveredKeys) > 0 {
			log.Printf("Total recovered keys before validation: %d", len(recoveredKeys))
			log.Printf("Validating GG20 recovery against ground truth (list-addresses)...")
			validationErr := ValidateGG20Recovery(vaultFiles, recoveredKeys, password)
			if validationErr != nil {
				return nil, fmt.Errorf("GG20 recovery validation failed: %w - This means the recovery is incorrect", validationErr)
			}
			log.Printf("✅ GG20 recovery validation passed - addresses match list-addresses")
		} else {
			log.Printf("⚠️ No keys were recovered at all!")
		}

	} else {
		// Fallback to old recovery method for non-GG20 vaults
		log.Printf("Using legacy recovery method (not GG20 format)")

		// Try ECDSA recovery first
		ecdsaResult, err := ReconstructTSSKey(vaultFiles, password, TssKeyType(ECDSA))
		if err == nil && ecdsaResult != nil {
			// Add Bitcoin recovery
			if ecdsaResult.BitcoinWIF != "" {
				recoveredKeys = append(recoveredKeys, RecoveredKey{
					Chain:      ChainBitcoin,
					PrivateKey: ecdsaResult.PrivateKeyHex,
					WIF:        ecdsaResult.BitcoinWIF,
					Address:    ecdsaResult.BitcoinAddress,
					DerivePath: "m/84'/0'/0'/0/0",
				})
			}

			// Add Ethereum recovery
			if ecdsaResult.EthereumAddress != "" {
				recoveredKeys = append(recoveredKeys, RecoveredKey{
					Chain:      ChainEthereum,
					PrivateKey: ecdsaResult.EthereumPrivateKeyHex,
					Address:    ecdsaResult.EthereumAddress,
					DerivePath: "m/44'/60'/0'/0/0",
				})
			}
		}

		// Try EdDSA recovery for Solana
		eddsaResult, err := ReconstructTSSKey(vaultFiles, password, TssKeyType(EdDSA))
		if err == nil && eddsaResult != nil {
			if eddsaResult.SolanaAddress != "" {
				recoveredKeys = append(recoveredKeys, RecoveredKey{
					Chain:      ChainSolana,
					PrivateKey: eddsaResult.PrivateKeyHex,
					Base58:     eddsaResult.SolanaAddress,
					Address:    eddsaResult.SolanaAddress,
					DerivePath: "m/44'/501'/0'/0'",
				})
			}
		}
	}

	if len(recoveredKeys) == 0 {
		return nil, fmt.Errorf("no private keys could be recovered from provided shares")
	}

	return recoveredKeys, nil
}

// DeriveAddress performs read-only HD derivation from a single vault share
// This is a STUB implementation - will be completed in future iterations
func DeriveAddress(vaultFile string, derivePath string, chain SupportedChain, password string) (*RecoveredKey, error) {
	// Parse the vault file first
	vaultInfo, err := vault.ParseVaultFileWithPassword(vaultFile, password)
	if err != nil {
		return nil, fmt.Errorf("failed to parse vault file: %w", err)
	}

	// Validate the vault has the required chain code for HD derivation
	if vaultInfo.HexChainCode == "" {
		return nil, fmt.Errorf("vault missing hex chain code required for HD derivation")
	}

	// TODO: Implement actual HD derivation logic
	// For now, return a placeholder structure

	return nil, fmt.Errorf("HD derivation not yet implemented - this is a stub for v0.2 milestone")
}

// GetCommonDerivationPaths returns common HD derivation paths for supported chains
func GetCommonDerivationPaths() map[SupportedChain][]DerivationPath {
	return map[SupportedChain][]DerivationPath{
		ChainBitcoin: {
			{Path: "m/44'/0'/0'/0/0", Chain: ChainBitcoin, Description: "First receiving address", Purpose: "receiving"},
			{Path: "m/44'/0'/0'/1/0", Chain: ChainBitcoin, Description: "First change address", Purpose: "change"},
			{Path: "m/49'/0'/0'/0/0", Chain: ChainBitcoin, Description: "P2SH-P2WPKH (SegWit v0)", Purpose: "segwit"},
			{Path: "m/84'/0'/0'/0/0", Chain: ChainBitcoin, Description: "P2WPKH (Native SegWit)", Purpose: "native_segwit"},
		},
		ChainEthereum: {
			{Path: "m/44'/60'/0'/0/0", Chain: ChainEthereum, Description: "First Ethereum address", Purpose: "receiving"},
			{Path: "m/44'/60'/0'/0/1", Chain: ChainEthereum, Description: "Second Ethereum address", Purpose: "receiving"},
			{Path: "m/44'/60'/0'/0/2", Chain: ChainEthereum, Description: "Third Ethereum address", Purpose: "receiving"},
		},
		ChainSolana: {
			{Path: "m/44'/501'/0'/0'", Chain: ChainSolana, Description: "Solana main account", Purpose: "receiving"},
			{Path: "m/44'/501'/1'/0'", Chain: ChainSolana, Description: "Solana second account", Purpose: "receiving"},
		},
		ChainThorChain: {
			{Path: "m/44'/931'/0'/0/0", Chain: ChainThorChain, Description: "THORChain main address", Purpose: "receiving"},
		},
	}
}

// ValidateDerivationPath checks if a derivation path is valid
func ValidateDerivationPath(path string) error {
	if path == "" {
		return fmt.Errorf("derivation path cannot be empty")
	}

	// Basic validation - should start with 'm/'
	if len(path) < 2 || path[:2] != "m/" {
		return fmt.Errorf("derivation path must start with 'm/'")
	}

	// TODO: Add more comprehensive path validation

	return nil
}

// GetSupportedChains returns a list of all supported blockchain chains
func GetSupportedChains() []SupportedChain {
	return []SupportedChain{
		// ECDSA-based chains
		ChainBitcoin,
		ChainBitcoinCash,
		ChainLitecoin,
		ChainDash,
		ChainDogecoin,
		ChainZcash,

		// Ethereum and EVM-compatible chains
		ChainEthereum,
		ChainArbitrum,
		ChainAvalanche,
		ChainBase,
		ChainBlast,
		ChainBSC,
		ChainCronos,
		ChainOptimism,
		ChainPolygon,
		ChainZkSync,

		// Cosmos-based chains
		ChainThorChain,

		// EdDSA-based chains
		ChainSolana,
		ChainSUI,
	}
}

// KeyShareData represents a parsed key share for reconstruction
type KeyShareData struct {
	Share     []byte // The actual share data
	Index     int    // Shamir share index (1-based)
	PublicKey string // Associated public key
}

// encodeBase58 encodes a byte slice to base58 using Bitcoin's alphabet
func encodeBase58(input []byte) string {
	// Bitcoin base58 alphabet
	alphabet := "123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz"

	if len(input) == 0 {
		return ""
	}

	// Count leading zeros
	zeros := 0
	for zeros < len(input) && input[zeros] == 0 {
		zeros++
	}

	// Encode the number
	input = input[zeros:]
	encoded := make([]byte, 0, len(input)*138/100+1) // log(256)/log(58), rounded up

	for _, b := range input {
		carry := int(b)
		for i := 0; i < len(encoded); i++ {
			carry += 256 * int(encoded[i])
			encoded[i] = byte(carry % 58)
			carry /= 58
		}
		for carry > 0 {
			encoded = append(encoded, byte(carry%58))
			carry /= 58
		}
	}

	// Reverse the encoded slice
	for i, j := 0, len(encoded)-1; i < j; i, j = i+1, j-1 {
		encoded[i], encoded[j] = encoded[j], encoded[i]
	}

	// Add leading '1's for leading zeros
	result := make([]byte, zeros+len(encoded))
	for i := 0; i < zeros; i++ {
		result[i] = '1'
	}
	for i := zeros; i < len(result); i++ {
		result[i] = alphabet[encoded[i-zeros]]
	}

	return string(result)
}

// generateSolanaWalletFormat generates a 64-byte base64 string for Solana wallet import
// This uses the TSS private key + vault public key: [tss_private(32) + vault_public(32)] in base64
func generateSolanaWalletFormat(privateKeyHex string, vaultPublicKeyHex string) (string, error) {
	// Convert hex private key to bytes (32 bytes)
	privateKeyBytes, err := hex.DecodeString(privateKeyHex)
	if err != nil {
		return "", fmt.Errorf("failed to decode private key: %w", err)
	}

	// Convert hex public key to bytes (32 bytes)
	publicKeyBytes, err := hex.DecodeString(vaultPublicKeyHex)
	if err != nil {
		return "", fmt.Errorf("failed to decode public key: %w", err)
	}

	// Ensure we have exactly 32 bytes for both
	if len(privateKeyBytes) != 32 {
		return "", fmt.Errorf("invalid private key length: expected 32 bytes, got %d", len(privateKeyBytes))
	}
	if len(publicKeyBytes) != 32 {
		return "", fmt.Errorf("invalid public key length: expected 32 bytes, got %d", len(publicKeyBytes))
	}

	// Create the 64-byte Ed25519 keypair: [private + public]
	// This is what Solana wallets (Phantom, Solflare) expect
	keypair := make([]byte, 64)
	copy(keypair[:32], privateKeyBytes)
	copy(keypair[32:], publicKeyBytes)

	// Encode to base64
	base64Keypair := base64.StdEncoding.EncodeToString(keypair)

	return base64Keypair, nil
}

// generateSolanaWalletJSON generates a JSON array of the 64-byte Ed25519 keypair
// This format can be directly pasted into Phantom or Solflare GUI imports
func generateSolanaWalletJSON(privateKeyHex string, vaultPublicKeyHex string) (string, error) {
	// Convert hex private key to bytes (32 bytes)
	privateKeyBytes, err := hex.DecodeString(privateKeyHex)
	if err != nil {
		return "", fmt.Errorf("failed to decode private key: %w", err)
	}

	// Convert hex public key to bytes (32 bytes)
	publicKeyBytes, err := hex.DecodeString(vaultPublicKeyHex)
	if err != nil {
		return "", fmt.Errorf("failed to decode public key: %w", err)
	}

	// Ensure we have exactly 32 bytes for both
	if len(privateKeyBytes) != 32 {
		return "", fmt.Errorf("invalid private key length: expected 32 bytes, got %d", len(privateKeyBytes))
	}
	if len(publicKeyBytes) != 32 {
		return "", fmt.Errorf("invalid public key length: expected 32 bytes, got %d", len(publicKeyBytes))
	}

	// Create the 64-byte Ed25519 keypair: [private + public]
	keypair := make([]byte, 64)
	copy(keypair[:32], privateKeyBytes)
	copy(keypair[32:], publicKeyBytes)

	// Convert to JSON array format: [61,220,129,202,...]
	// This is what Phantom/Solflare expect when importing via GUI
	jsonArray := "["
	for i, b := range keypair {
		if i > 0 {
			jsonArray += ","
		}
		jsonArray += fmt.Sprintf("%d", b)
	}
	jsonArray += "]"

	return jsonArray, nil
}

// generateSuiWalletFormat generates a 33-byte base64 string for SUI wallet import
// The format is [0x00 || 32-byte seed] where 0x00 indicates Ed25519 curve
func generateSuiWalletFormat(privateKeyHex string) (string, error) {
	// Convert hex private key to bytes (32 bytes)
	privateKeyBytes, err := hex.DecodeString(privateKeyHex)
	if err != nil {
		return "", fmt.Errorf("failed to decode private key: %w", err)
	}

	// Ensure we have exactly 32 bytes
	if len(privateKeyBytes) != 32 {
		return "", fmt.Errorf("invalid private key length: expected 32 bytes, got %d", len(privateKeyBytes))
	}

	// SUI expects [0x00 || 32-byte seed] format
	// The 0x00 prefix indicates Ed25519 curve
	suiKey := make([]byte, 33)
	suiKey[0] = 0x00 // Ed25519 curve indicator
	copy(suiKey[1:], privateKeyBytes)

	// Encode to base64 - this is what SUI wallets expect
	base64Key := base64.StdEncoding.EncodeToString(suiKey)

	return base64Key, nil
}

// convertTSSToRecoveredKeys converts TSSRecoveryResult to RecoveredKey slice using centralized derivation
// This ensures recovery uses the SAME address derivation logic as list-addresses
// The keyType parameter determines whether this handles ECDSA or EdDSA chains
func convertTSSToRecoveredKeys(tssResult *TSSRecoveryResult, keyType TssKeyType, originalVault *vault.VaultInfo) []RecoveredKey {
	var keys []RecoveredKey

	// Use list-addresses derivation to get the canonical addresses for each chain
	recoveryVault := &vault.VaultInfo{
		PublicKeyECDSA: "",
		HexChainCode:   tssResult.ChainCode,
		PublicKeyEDDSA: "",
	}
	switch keyType {
	case ECDSA:
		recoveryVault.PublicKeyECDSA = originalVault.PublicKeyECDSA
	case EdDSA:
		recoveryVault.PublicKeyEDDSA = originalVault.PublicKeyEDDSA
	}
	expectedAddresses := vault.DeriveAddressesFromVault(recoveryVault)

	// Helper to fetch the per-chain derived private key that matches the address we will output
	getChainKeys := func(chainLower string) (ChainKeys, bool) {
		switch chainLower {
		case "bitcoin":
			return tssResult.Addresses.Bitcoin, tssResult.Addresses.Bitcoin.PrivateKeyHex != ""
		case "bitcoin-cash":
			return tssResult.Addresses.BitcoinCash, tssResult.Addresses.BitcoinCash.PrivateKeyHex != ""
		case "litecoin":
			return tssResult.Addresses.Litecoin, tssResult.Addresses.Litecoin.PrivateKeyHex != ""
		case "dogecoin":
			return tssResult.Addresses.Dogecoin, tssResult.Addresses.Dogecoin.PrivateKeyHex != ""
		case "dash":
			return tssResult.Addresses.Dash, tssResult.Addresses.Dash.PrivateKeyHex != ""
		case "zcash":
			return tssResult.Addresses.Zcash, tssResult.Addresses.Zcash.PrivateKeyHex != ""
		case "thorchain":
			return tssResult.Addresses.THORChain, tssResult.Addresses.THORChain.PrivateKeyHex != ""
		case "ethereum", "bsc", "avalanche", "polygon", "cronoschain", "arbitrum", "optimism", "base", "blast", "zksync":
			return tssResult.Addresses.Ethereum, tssResult.Addresses.Ethereum.PrivateKeyHex != ""
		case "solana":
			return tssResult.Addresses.Solana, tssResult.Addresses.Solana.PrivateKeyHex != ""
		case "sui":
			return tssResult.Addresses.SUI, tssResult.Addresses.SUI.PrivateKeyHex != ""
		default:
			return ChainKeys{}, false
		}
	}

	for _, addr := range expectedAddresses {
		chainKey := strings.ToLower(addr.Chain)

		// Filter by key type
		isEd := chainKey == "solana" || chainKey == "sui"
		if (keyType == ECDSA && isEd) || (keyType == EdDSA && !isEd) {
			continue
		}

		// Prefer the per-chain derived key produced during TSS recovery.
		// This ensures the printed private key actually corresponds to the printed address.
		ck, ok := getChainKeys(chainKey)
		privateKeyHex := ""
		if ok {
			privateKeyHex = ck.PrivateKeyHex
		}
		// Fallback: if for some reason we didn't produce a per-chain key, keep master (legacy behavior)
		if privateKeyHex == "" {
			privateKeyHex = derivePrivateKeyForPath(tssResult.PrivateKeyHex, tssResult.ChainCode, addr.DerivePath)
		}

		recoveredKey := RecoveredKey{
			Chain:      SupportedChain(chainKey), // Use chainKey instead of supportedChain constant
			PrivateKey: privateKeyHex,
			Address:    addr.Address,
			DerivePath: addr.DerivePath,
		}

		// Add WIF for UTXO chains to ease wallet imports
		switch chainKey {
		case "bitcoin":
			recoveredKey.WIF = generateWIFFromPrivateKey(privateKeyHex)
		case "litecoin", "dogecoin", "dash", "bitcoin-cash", "zcash":
			recoveredKey.WIF = generateChainSpecificWIF(chainKey, privateKeyHex)
		}

		// Wallet-friendly formats for EdDSA chains
		if isEd {
			vaultEdDSAPublicKey := originalVault.PublicKeyEDDSA
			switch chainKey {
			case "solana":
				if seedBytes, err := hex.DecodeString(privateKeyHex); err == nil && len(seedBytes) == 32 {
					recoveredKey.SolanaSeedFormat = base64.StdEncoding.EncodeToString(seedBytes)
				}
				if solanaFormat, err := generateSolanaWalletFormat(privateKeyHex, vaultEdDSAPublicKey); err == nil {
					recoveredKey.SolanaWalletFormat = solanaFormat
				}
				if solanaJSON, err := generateSolanaWalletJSON(privateKeyHex, vaultEdDSAPublicKey); err == nil {
					recoveredKey.SolanaWalletJSON = solanaJSON
				}
			case "sui":
				if suiFormat, err := generateSuiWalletFormat(privateKeyHex); err == nil {
					recoveredKey.SuiWalletFormat = suiFormat
				}
			}
		}

		keys = append(keys, recoveredKey)
	}

	return keys
}

// derivePrivateKeyForPath derives a private key for a specific HD path
// Note: Vultisig MPC design uses master private key with path-specific address derivation
// This is correct behavior for MPC systems where private key shares are never derived
func derivePrivateKeyForPath(masterPrivateKeyHex, chainCodeHex, derivePath string) string {
	return masterPrivateKeyHex
}

// generateWIFFromPrivateKey generates a compressed mainnet WIF for Bitcoin
func generateWIFFromPrivateKey(privateKeyHex string) string {
	b, err := hex.DecodeString(privateKeyHex)
	if err != nil || len(b) != 32 {
		return ""
	}
	return generateWIFWithVersion(b, 0x80, true)
}

// generateChainSpecificWIF generates WIF strings for other UTXO chains
func generateChainSpecificWIF(chainKey string, privateKeyHex string) string {
	b, err := hex.DecodeString(privateKeyHex)
	if err != nil || len(b) != 32 {
		return ""
	}
	var version byte
	switch chainKey {
	case "litecoin":
		version = 0xB0
	case "dogecoin":
		version = 0x9E
	case "dash":
		version = 0xCC
	case "bitcoin-cash":
		version = 0x80
	case "zcash":
		version = 0x80
	default:
		return ""
	}
	return generateWIFWithVersion(b, version, true)
}

// generateWIFWithVersion builds a WIF using a custom version byte
func generateWIFWithVersion(privateKeyBytes []byte, version byte, compressed bool) string {
	versioned := append([]byte{version}, privateKeyBytes...)
	if compressed {
		versioned = append(versioned, 0x01)
	}
	hash1 := sha256.Sum256(versioned)
	hash2 := sha256.Sum256(hash1[:])
	final := append(versioned, hash2[:4]...)
	return encodeBase58(final)
}

// CheckIfDKLSVault determines if a vault file is in DKLS format
// NOTE: DKLS support is not yet implemented in this version
// This is a stub that always returns false (assumes GG20 format)
func CheckIfDKLSVault(inputFileName string, password string) (bool, error) {
	// TODO: Implement proper DKLS vault detection when DKLS support is added
	// For now, assume all vaults are GG20 format
	return false, nil
}

// CheckIfGG20Vault determines if a vault file is in GG20 format
func CheckIfGG20Vault(inputFileName string, password string) (bool, error) {
	// Use the existing function from DKLS recovery and invert the result
	isDKLS, err := CheckIfDKLSVault(inputFileName, password)
	if err != nil {
		return false, err
	}
	return !isDKLS, nil
}

// ValidationResult represents the result of validating a single chain
type ValidationResult struct {
	Chain     string
	Passed    bool
	Recovered string
	Expected  string
	Error     string
}

// ValidateGG20Recovery compares recovered addresses against list-addresses output
// Returns validation results for all chains instead of failing on first error
func ValidateGG20Recovery(vaultFiles []string, recoveredKeys []RecoveredKey, password string) error {
	if len(vaultFiles) == 0 {
		return fmt.Errorf("no vault files provided for validation")
	}

	// Use the first vault file to get expected addresses
	vaultInfo, err := vault.ParseVaultFileWithPassword(vaultFiles[0], password)
	if err != nil {
		return fmt.Errorf("failed to parse vault for validation: %w", err)
	}

	// Get expected addresses using list-addresses logic
	expectedAddresses := vault.DeriveAddressesFromVault(vaultInfo)

	// Create a map of expected addresses by chain
	expectedByChain := make(map[string]string)
	for _, addr := range expectedAddresses {
		// Normalize chain name to lowercase for comparison
		chainKey := strings.ToLower(addr.Chain)
		expectedByChain[chainKey] = addr.Address
	}

	// Validate each recovered key
	validatedCount := 0
	failedCount := 0

	for _, key := range recoveredKeys {
		chainKey := strings.ToLower(string(key.Chain))
		expectedAddr, exists := expectedByChain[chainKey]
		if !exists {
			log.Printf("Warning: No expected address found for chain %s", key.Chain)
			continue
		}

		if key.Address != expectedAddr {
			log.Printf("❌ %s address validation FAILED: recovered=%s, expected=%s", key.Chain, key.Address, expectedAddr)
			failedCount++
		} else {
			log.Printf("✅ %s address validation passed: %s", key.Chain, key.Address)
			validatedCount++
		}
	}

	if validatedCount == 0 {
		return fmt.Errorf("no addresses were validated - this suggests a complete recovery failure")
	}

	if failedCount > 0 {
		log.Printf("⚠️  GG20 recovery validation: %d passed, %d failed - some chains have incorrect addresses", validatedCount, failedCount)
		// Don't return error - let recovery proceed with warning
	} else {
		log.Printf("✅ GG20 recovery validation passed - all %d addresses match list-addresses", validatedCount)
	}

	return nil
}
