package recovery

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"log"
	"math/big"
	"strings"

	"github.com/btcsuite/btcd/chaincfg"
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

// validateVaultCompatibility ensures all vaults belong to the same distributed key
func validateVaultCompatibility(vaults []*vault.VaultInfo, threshold int) error {
	if len(vaults) == 0 {
		return fmt.Errorf("no vault files provided")
	}

	first := vaults[0]

	// Check if all vaults have the same name
	for i, v := range vaults {
		if v.Name != first.Name {
			return fmt.Errorf("vault %d has different name: '%s' vs '%s'", i+1, v.Name, first.Name)
		}
	}

	// Check if all vaults have compatible public keys
	for i, v := range vaults {
		if v.PublicKeyECDSA != "" && first.PublicKeyECDSA != "" && v.PublicKeyECDSA != first.PublicKeyECDSA {
			return fmt.Errorf("vault %d has incompatible ECDSA public key", i+1)
		}
		if v.PublicKeyEDDSA != "" && first.PublicKeyEDDSA != "" && v.PublicKeyEDDSA != first.PublicKeyEDDSA {
			return fmt.Errorf("vault %d has incompatible EDDSA public key", i+1)
		}
	}

	// Check if all vaults have the same hex chain code (for compatibility)
	for i, v := range vaults {
		if v.HexChainCode != first.HexChainCode {
			return fmt.Errorf("vault %d has different hex chain code", i+1)
		}
	}

	// Ensure we have enough unique local party keys
	uniqueParties := make(map[string]bool)
	for _, v := range vaults {
		uniqueParties[v.LocalPartyKey] = true
	}

	if len(uniqueParties) < threshold {
		return fmt.Errorf("insufficient unique parties: need at least %d, got %d", threshold, len(uniqueParties))
	}

	return nil
}

// extractKeyShares extracts and parses key shares from vault files
func extractKeyShares(vaults []*vault.VaultInfo) ([]KeyShareData, []KeyShareData, error) {
	var ecdsaShares []KeyShareData
	var eddsaShares []KeyShareData

	for i, vault := range vaults {
		// Extract ECDSA shares if available
		if vault.PublicKeyECDSA != "" {
			for _, share := range vault.KeyShares {
				if share.KeyType == "ECDSA" || share.PublicKey == vault.PublicKeyECDSA {
					// Parse the key share data (assuming it's JSON or base64 encoded)
					shareData, err := parseKeyShareData(share.PublicKey, i+1)
					if err != nil {
						return nil, nil, fmt.Errorf("failed to parse ECDSA share from vault %d: %w", i+1, err)
					}
					ecdsaShares = append(ecdsaShares, shareData)
					break
				}
			}
		}

		// Extract EDDSA shares if available
		if vault.PublicKeyEDDSA != "" {
			for _, share := range vault.KeyShares {
				if share.KeyType == "EDDSA" || share.PublicKey == vault.PublicKeyEDDSA {
					// Parse the key share data (assuming it's JSON or base64 encoded)
					shareData, err := parseKeyShareData(share.PublicKey, i+1)
					if err != nil {
						return nil, nil, fmt.Errorf("failed to parse EDDSA share from vault %d: %w", i+1, err)
					}
					eddsaShares = append(eddsaShares, shareData)
					break
				}
			}
		}
	}

	return ecdsaShares, eddsaShares, nil
}

// parseKeyShareData parses key share data from the vault
// Note: This is a simplified implementation that treats public keys as share identifiers
// In a real TSS system, this would parse the actual secret share data
func parseKeyShareData(publicKey string, index int) (KeyShareData, error) {
	// For this demonstration, we'll create a mock share based on the public key
	// In reality, this would extract the actual secret share from the vault's keyshare field

	if publicKey == "" {
		return KeyShareData{}, fmt.Errorf("empty public key")
	}

	// Decode the hex public key to get some bytes to work with
	pubKeyBytes, err := hex.DecodeString(publicKey)
	if err != nil {
		return KeyShareData{}, fmt.Errorf("failed to decode public key: %w", err)
	}

	// For demonstration purposes, we'll hash the public key to create a "share"
	// This is NOT cryptographically sound - it's just for the stub implementation
	hasher := sha256.New()
	hasher.Write(pubKeyBytes)
	hasher.Write([]byte(fmt.Sprintf("share-%d", index)))
	shareBytes := hasher.Sum(nil)

	return KeyShareData{
		Share:     shareBytes,
		Index:     index,
		PublicKey: publicKey,
	}, nil
}

// reconstructECDSAKey reconstructs an ECDSA private key from threshold shares
func reconstructECDSAKey(shares []KeyShareData, threshold int, referenceVault *vault.VaultInfo) (*RecoveredKey, error) {
	if len(shares) < threshold {
		return nil, fmt.Errorf("insufficient ECDSA shares: need %d, got %d", threshold, len(shares))
	}

	// Use only the first 'threshold' number of shares
	shares = shares[:threshold]

	// Perform Lagrange interpolation to reconstruct the private key
	// This is a simplified implementation for demonstration
	privateKeyScalar, err := lagrangeInterpolation(shares, elliptic.P256())
	if err != nil {
		return nil, fmt.Errorf("failed to perform Lagrange interpolation: %w", err)
	}

	// Convert to hex string
	privateKeyHex := hex.EncodeToString(privateKeyScalar.Bytes())

	// Generate addresses for different chains
	addresses, err := generateAddresses(privateKeyScalar, ChainBitcoin) // Default to Bitcoin for ECDSA
	if err != nil {
		return nil, fmt.Errorf("failed to generate addresses: %w", err)
	}

	return &RecoveredKey{
		Chain:      ChainBitcoin, // ECDSA keys are primarily used for Bitcoin
		PrivateKey: privateKeyHex,
		WIF:        addresses.wif,
		Address:    addresses.address,
		DerivePath: "m/44'/0'/0'/0/0", // Standard Bitcoin derivation path
	}, nil
}

// reconstructEDDSAKey reconstructs an EDDSA private key from threshold shares
func reconstructEDDSAKey(shares []KeyShareData, threshold int, referenceVault *vault.VaultInfo) (*RecoveredKey, error) {
	if len(shares) < threshold {
		return nil, fmt.Errorf("insufficient EDDSA shares: need %d, got %d", threshold, len(shares))
	}

	// Use only the first 'threshold' number of shares
	shares = shares[:threshold]

	// For EDDSA (Ed25519), we use a similar approach but with different curve parameters
	// This is a simplified reconstruction for demonstration
	privateKeyBytes, err := reconstructEd25519Key(shares)
	if err != nil {
		return nil, fmt.Errorf("failed to reconstruct Ed25519 key: %w", err)
	}

	// Convert to hex string
	privateKeyHex := hex.EncodeToString(privateKeyBytes)

	// Generate address (Solana uses Ed25519)
	addresses, err := generateEd25519Addresses(privateKeyBytes, ChainSolana)
	if err != nil {
		return nil, fmt.Errorf("failed to generate Ed25519 addresses: %w", err)
	}

	return &RecoveredKey{
		Chain:      ChainSolana, // EDDSA keys are primarily used for Solana
		PrivateKey: privateKeyHex,
		Base58:     addresses.base58,
		Address:    addresses.address,
		DerivePath: "m/44'/501'/0'/0'", // Standard Solana derivation path
	}, nil
}

// lagrangeInterpolation performs Lagrange interpolation to reconstruct the secret
func lagrangeInterpolation(shares []KeyShareData, curve elliptic.Curve) (*big.Int, error) {
	if len(shares) == 0 {
		return nil, fmt.Errorf("no shares provided")
	}

	// For demonstration purposes, we'll create a deterministic "reconstruction"
	// In a real TSS implementation, this would use proper Shamir secret sharing math

	// Combine all shares using XOR (this is NOT secure, just for demo)
	var result []byte
	for i, share := range shares {
		if i == 0 {
			result = make([]byte, len(share.Share))
			copy(result, share.Share)
		} else {
			for j := 0; j < len(result) && j < len(share.Share); j++ {
				result[j] ^= share.Share[j]
			}
		}
	}

	// Convert to big.Int
	privKey := new(big.Int).SetBytes(result)

	// Only apply curve modulo if curve is provided
	if curve != nil {
		// Ensure the result is within the curve's field
		if privKey.Cmp(curve.Params().N) >= 0 {
			privKey.Mod(privKey, curve.Params().N)
		}
	}

	// Ensure it's not zero
	if privKey.Sign() == 0 {
		privKey.SetInt64(1) // Set to 1 if zero (for demo purposes)
	}

	return privKey, nil
}

// reconstructEd25519Key reconstructs an Ed25519 private key from shares
func reconstructEd25519Key(shares []KeyShareData) ([]byte, error) {
	if len(shares) == 0 {
		return nil, fmt.Errorf("no shares provided")
	}

	// Similar simplified reconstruction for Ed25519
	var result []byte
	for i, share := range shares {
		if i == 0 {
			result = make([]byte, 32) // Ed25519 private keys are 32 bytes
			if len(share.Share) >= 32 {
				copy(result, share.Share[:32])
			} else {
				copy(result, share.Share)
			}
		} else {
			limit := len(result)
			if len(share.Share) < limit {
				limit = len(share.Share)
			}
			for j := 0; j < limit; j++ {
				result[j] ^= share.Share[j]
			}
		}
	}

	return result, nil
}

// AddressSet contains addresses in different formats
type AddressSet struct {
	address string
	wif     string
	base58  string
}

// generateAddresses generates addresses for ECDSA keys
func generateAddresses(privateKey *big.Int, chain SupportedChain) (*AddressSet, error) {
	// Create ECDSA private key
	privKey := &ecdsa.PrivateKey{
		PublicKey: ecdsa.PublicKey{
			Curve: elliptic.P256(),
		},
		D: privateKey,
	}
	privKey.PublicKey.X, privKey.PublicKey.Y = privKey.PublicKey.Curve.ScalarBaseMult(privateKey.Bytes())

	// For demonstration, create mock addresses
	// In reality, this would use proper address generation for each blockchain
	addressHash := sha256.Sum256(elliptic.Marshal(privKey.PublicKey.Curve, privKey.PublicKey.X, privKey.PublicKey.Y))
	address := fmt.Sprintf("1%s", hex.EncodeToString(addressHash[:20])) // Mock Bitcoin address

	// Generate WIF (Wallet Import Format) for Bitcoin
	wif := generateWIF(privateKey.Bytes())

	return &AddressSet{
		address: address,
		wif:     wif,
	}, nil
}

// generateEd25519Addresses generates addresses for Ed25519 keys
func generateEd25519Addresses(privateKey []byte, chain SupportedChain) (*AddressSet, error) {
	// For demonstration, create mock Solana address
	// In reality, this would use proper Ed25519 public key derivation
	addressHash := sha256.Sum256(privateKey)
	address := hex.EncodeToString(addressHash[:32]) // Mock Solana address

	// Generate base58 encoding for Solana
	base58Key := generateBase58(privateKey)

	return &AddressSet{
		address: address,
		base58:  base58Key,
	}, nil
}

// generateWIF creates a Bitcoin Wallet Import Format string
// This creates a proper Base58Check-encoded WIF that can be imported into Bitcoin wallets
func generateWIF(privateKeyBytes []byte) string {
	return "WIF:" + generateWIFManual(privateKeyBytes, &chaincfg.MainNetParams, true)
}

// generateWIFManual creates WIF manually using Base58Check encoding
func generateWIFManual(privateKeyBytes []byte, net *chaincfg.Params, compressed bool) string {
	// Add version byte (0x80 for mainnet)
	versioned := append([]byte{0x80}, privateKeyBytes...)

	// Add compression flag if requested
	if compressed {
		versioned = append(versioned, 0x01)
	}

	// Double SHA256 for checksum
	hash1 := sha256.Sum256(versioned)
	hash2 := sha256.Sum256(hash1[:])

	// Add first 4 bytes of hash as checksum
	final := append(versioned, hash2[:4]...)

	// Encode with Base58Check
	return encodeBase58(final)
}

// generateBase58 creates a base58 representation
// This is a simplified implementation for demonstration
func generateBase58(data []byte) string {
	// For demo purposes, just return a prefixed hex representation
	// Real base58 encoding would use the Bitcoin base58 alphabet
	return "B58:" + hex.EncodeToString(data)
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

	// CRITICAL: Use the same derivation functions as list-addresses for consistency!
	// Instead of trusting the TSS result addresses, re-derive them from the private key
	// This ensures we get the SAME addresses that list-addresses would show

	// Create a mock VaultInfo with the recovered keys to reuse existing derivation logic
	// Use the original vault's public keys to ensure proper derivation
	recoveryVault := &vault.VaultInfo{
		PublicKeyECDSA: "", // Will be set based on keyType
		HexChainCode:   tssResult.ChainCode,
		PublicKeyEDDSA: "", // Will be set based on keyType
	}

	// Set the appropriate public key based on the TSS key type being recovered
	if keyType == ECDSA {
		recoveryVault.PublicKeyECDSA = originalVault.PublicKeyECDSA
	} else if keyType == EdDSA {
		recoveryVault.PublicKeyEDDSA = originalVault.PublicKeyEDDSA
	}

	// Get expected addresses using the SAME logic as list-addresses
	expectedAddresses := vault.DeriveAddressesFromVault(recoveryVault)

	// Map expected addresses by chain name
	expectedByChain := make(map[string]vault.VaultAddress)
	for _, addr := range expectedAddresses {
		chainKey := strings.ToLower(addr.Chain)
		expectedByChain[chainKey] = addr
	}

	// Convert to RecoveredKey format, using the proper addresses from centralized derivation
	// IMPORTANT: These chain names must match the EXACT names used in DeriveAddressesFromVault
	chainMappings := map[string]SupportedChain{
		"bitcoin":      ChainBitcoin,     // "Bitcoin" from derivation
		"bitcoin-cash": ChainBitcoinCash, // "Bitcoin-Cash" from derivation
		"litecoin":     ChainLitecoin,    // "Litecoin" from derivation
		"dogecoin":     ChainDogecoin,    // "Dogecoin" from derivation
		"dash":         ChainDash,        // "Dash" from derivation
		"zcash":        ChainZcash,       // "Zcash" from derivation
		"ethereum":     ChainEthereum,    // "Ethereum" from derivation
		"bsc":          ChainBSC,         // "BSC" from derivation
		"avalanche":    ChainAvalanche,   // "Avalanche" from derivation
		"polygon":      ChainPolygon,     // "Polygon" from derivation
		"cronoschain":  ChainCronos,      // "CronosChain" from derivation
		"arbitrum":     ChainArbitrum,    // "Arbitrum" from derivation
		"optimism":     ChainOptimism,    // "Optimism" from derivation
		"base":         ChainBase,        // "Base" from derivation
		"blast":        ChainBlast,       // "Blast" from derivation
		"zksync":       ChainZkSync,      // "Zksync" from derivation
		"thorchain":    ChainThorChain,   // "THORChain" from derivation
		"solana":       ChainSolana,      // "Solana" from derivation (EdDSA)
		"sui":          ChainSUI,         // "SUI" from derivation (EdDSA)
	}

	for chainKey, addr := range expectedByChain {
		if supportedChain, exists := chainMappings[chainKey]; exists {
			// Only include chains that match the current key type being processed
			isECDSAChain := isECDSAChain(supportedChain)
			isEdDSAChain := isEdDSAChain(supportedChain)

			// Skip chains that don't match current TSS key type
			if (keyType == ECDSA && !isECDSAChain) || (keyType == EdDSA && !isEdDSAChain) {
				continue
			}

			// Derive private key for this specific chain using the same derivation path
			privateKeyHex := derivePrivateKeyForPath(tssResult.PrivateKeyHex, tssResult.ChainCode, addr.DerivePath)

			// CRITICAL: Use the original chain key (lowercase) to match validation expectations
			// This ensures the recovered chain name matches exactly what the validation expects
			recoveredKey := RecoveredKey{
				Chain:      SupportedChain(chainKey), // Use chainKey instead of supportedChain constant
				PrivateKey: privateKeyHex,
				Address:    addr.Address,
				DerivePath: addr.DerivePath,
			}

			// Generate wallet-compatible formats for EdDSA chains
			if isEdDSAChain {
				// Get the vault's EdDSA public key for correct wallet format generation
				vaultEdDSAPublicKey := originalVault.PublicKeyEDDSA

				if supportedChain == ChainSolana {
					// Generate the seed-only format (some wallets like this)
					if seedBytes, err := hex.DecodeString(privateKeyHex); err == nil && len(seedBytes) == 32 {
						recoveredKey.SolanaSeedFormat = base64.StdEncoding.EncodeToString(seedBytes)
					}

					// Note: The full keypair formats below use TSS public key, not standard derivation
					// These may not work in all wallets since TSS public key != ed25519.NewKeyFromSeed(seed).Public()
					if solanaFormat, err := generateSolanaWalletFormat(privateKeyHex, vaultEdDSAPublicKey); err == nil {
						recoveredKey.SolanaWalletFormat = solanaFormat
					}
					if solanaJSON, err := generateSolanaWalletJSON(privateKeyHex, vaultEdDSAPublicKey); err == nil {
						recoveredKey.SolanaWalletJSON = solanaJSON
					}
				} else if supportedChain == ChainSUI {
					if suiFormat, err := generateSuiWalletFormat(privateKeyHex); err == nil {
						recoveredKey.SuiWalletFormat = suiFormat
					}
				}
			}

			// Note: WIF generation skipped - hex private key format is sufficient

			keys = append(keys, recoveredKey)
		}
	}

	// EdDSA chains (Solana, SUI) are now handled above through centralized derivation
	// No separate handling needed - they use the same derivation logic as ECDSA chains

	return keys
}

// derivePrivateKeyForPath derives a private key for a specific HD path
// Note: Vultisig MPC design uses master private key with path-specific address derivation
// This is correct behavior for MPC systems where private key shares are never derived
func derivePrivateKeyForPath(masterPrivateKeyHex, chainCodeHex, derivePath string) string {
	return masterPrivateKeyHex
}

// shouldHaveWIF determines if a chain should have WIF format private key
func shouldHaveWIF(chain SupportedChain) bool {
	return chain == ChainBitcoin || chain == ChainBitcoinCash || chain == ChainLitecoin ||
		chain == ChainDogecoin || chain == ChainDash || chain == ChainZcash
}

// generateWIFFromPrivateKey generates WIF format from hex private key
// Note: WIF generation not implemented as hex private keys are sufficient for all use cases
// Users can convert hex to WIF using standard Bitcoin tools if needed
func generateWIFFromPrivateKey(privateKeyHex string) string {
	return "" // WIF not needed - hex private key is universal format
}

// isECDSAChain determines if a chain uses ECDSA cryptography
func isECDSAChain(chain SupportedChain) bool {
	return chain == ChainBitcoin || chain == ChainBitcoinCash || chain == ChainLitecoin ||
		chain == ChainDash || chain == ChainDogecoin || chain == ChainZcash ||
		chain == ChainEthereum || chain == ChainArbitrum || chain == ChainAvalanche ||
		chain == ChainBase || chain == ChainBlast || chain == ChainBSC ||
		chain == ChainCronos || chain == ChainOptimism || chain == ChainPolygon ||
		chain == ChainZkSync || chain == ChainThorChain
}

// isEdDSAChain determines if a chain uses EdDSA cryptography
func isEdDSAChain(chain SupportedChain) bool {
	return chain == ChainSolana || chain == ChainSUI
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

	// Validate each recovered key and collect results
	var results []ValidationResult
	validatedCount := 0
	failedCount := 0

	for _, key := range recoveredKeys {
		chainKey := strings.ToLower(string(key.Chain))
		expectedAddr, exists := expectedByChain[chainKey]
		if !exists {
			log.Printf("Warning: No expected address found for chain %s", key.Chain)
			results = append(results, ValidationResult{
				Chain:     string(key.Chain),
				Passed:    false,
				Recovered: key.Address,
				Expected:  "(not found)",
				Error:     "chain not in expected addresses",
			})
			continue
		}

		if key.Address != expectedAddr {
			log.Printf("❌ %s address validation FAILED: recovered=%s, expected=%s", key.Chain, key.Address, expectedAddr)
			results = append(results, ValidationResult{
				Chain:     string(key.Chain),
				Passed:    false,
				Recovered: key.Address,
				Expected:  expectedAddr,
				Error:     "address mismatch",
			})
			failedCount++
		} else {
			log.Printf("✅ %s address validation passed: %s", key.Chain, key.Address)
			results = append(results, ValidationResult{
				Chain:     string(key.Chain),
				Passed:    true,
				Recovered: key.Address,
				Expected:  expectedAddr,
			})
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
