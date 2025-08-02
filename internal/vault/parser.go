package vault

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"syscall"

	"github.com/golang/protobuf/proto"
	v1 "github.com/vultisig/commondata/go/vultisig/vault/v1"
	"golang.org/x/term"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// VaultInfo contains parsed vault information
type VaultInfo struct {
	Name           string              `json:"name"`
	PublicKeyECDSA string              `json:"public_key_ecdsa"`
	PublicKeyEDDSA string              `json:"public_key_eddsa"`
	HexChainCode   string              `json:"hex_chain_code"`
	LocalPartyKey  string              `json:"local_party_key"`
	IsEncrypted    bool                `json:"is_encrypted"`
	Version        int32               `json:"version"`
	KeyShares      []KeyShareInfo      `json:"key_shares,omitempty"`
	Metadata       map[string]string   `json:"metadata,omitempty"`
	CreatedAt      int64               `json:"created_at,omitempty"`
	FilePath       string              `json:"file_path"`
}

// KeyShareInfo contains information about a key share
type KeyShareInfo struct {
	PublicKey string `json:"public_key"`
	KeyType   string `json:"key_type"` // ECDSA or EDDSA
}

// ParseVaultFile parses a .vult file and returns vault information
// Uses interactive password prompt for encrypted vaults
func ParseVaultFile(filePath string) (*VaultInfo, error) {
	return ParseVaultFileWithPassword(filePath, "")
}

// ParseVaultFileWithPassword parses a .vult file with optional password parameter
// If password is empty and vault is encrypted, falls back to interactive prompt
func ParseVaultFileWithPassword(filePath, password string) (*VaultInfo, error) {
	// Get absolute path
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		return nil, fmt.Errorf("error getting absolute path: %w", err)
	}

	// Check if file exists
	if _, err := os.Stat(absPath); err != nil {
		return nil, fmt.Errorf("error accessing file %s: %w", absPath, err)
	}

	// Validate file path for security
	if err := validateSafePath(absPath); err != nil {
		return nil, fmt.Errorf("unsafe file path: %w", err)
	}

	// Read file content
	fileContent, err := os.ReadFile(absPath)
	if err != nil {
		return nil, fmt.Errorf("error reading file: %w", err)
	}

	// Decode base64 content
	rawContent, err := base64.StdEncoding.DecodeString(string(fileContent))
	if err != nil {
		return nil, fmt.Errorf("error decoding base64 content: %w", err)
	}

	// Unmarshal vault container
	var vaultContainer v1.VaultContainer
	if err := proto.Unmarshal(rawContent, &vaultContainer); err != nil {
		return nil, fmt.Errorf("error unmarshalling vault container: %w", err)
	}

	// Handle encrypted vs unencrypted vaults
	var vault *v1.Vault
	if vaultContainer.IsEncrypted {
		vault, err = decryptVaultWithPassword(&vaultContainer, absPath, password)
		if err != nil {
			return nil, fmt.Errorf("error decrypting vault: %w", err)
		}
	} else {
		vaultData, err := base64.StdEncoding.DecodeString(vaultContainer.Vault)
		if err != nil {
			return nil, fmt.Errorf("error decoding vault data: %w", err)
		}
		vault = &v1.Vault{}
		if err := proto.Unmarshal(vaultData, vault); err != nil {
			return nil, fmt.Errorf("error unmarshalling vault: %w", err)
		}
	}

	// Build vault info
	vaultInfo := &VaultInfo{
		Name:           vault.Name,
		PublicKeyECDSA: vault.PublicKeyEcdsa,
		PublicKeyEDDSA: vault.PublicKeyEddsa,
		HexChainCode:   vault.HexChainCode,
		LocalPartyKey:  vault.LocalPartyId,
		IsEncrypted:    vaultContainer.IsEncrypted,
		Version:        0, // Version field doesn't exist in v1.Vault
		CreatedAt:      getTimestamp(vault.CreatedAt),
		FilePath:       absPath,
		Metadata:       make(map[string]string),
	}

	// Extract key share information
	for _, keyShare := range vault.KeyShares {
		keyType := "ECDSA"
		if keyShare.PublicKey == vault.PublicKeyEddsa {
			keyType = "EDDSA"
		}
		vaultInfo.KeyShares = append(vaultInfo.KeyShares, KeyShareInfo{
			PublicKey: keyShare.PublicKey,
			KeyType:   keyType,
		})
	}

	return vaultInfo, nil
}

// decryptVault decrypts an encrypted vault using interactive password prompt
func decryptVault(container *v1.VaultContainer, filePath string) (*v1.Vault, error) {
	return decryptVaultWithPassword(container, filePath, "")
}

// decryptVaultWithPassword decrypts an encrypted vault with optional password parameter
// If password is empty, falls back to interactive prompt
func decryptVaultWithPassword(container *v1.VaultContainer, filePath, password string) (*v1.Vault, error) {
	var passwordBytes []byte
	var err error

	if password != "" {
		// Use provided password
		passwordBytes = []byte(password)
	} else {
		// Prompt for password interactively
		fmt.Printf("Enter password for encrypted vault (%s): ", filepath.Base(filePath))
		passwordBytes, err = term.ReadPassword(int(syscall.Stdin))
		if err != nil {
			return nil, fmt.Errorf("failed to read password: %w", err)
		}
		fmt.Println() // Print newline after password input
	}

	// Derive key from password
	hasher := sha256.New()
	hasher.Write(passwordBytes)
	key := hasher.Sum(nil)

	// Decrypt vault data
	decryptedData, err := decryptAES(container.Vault, key)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt vault: %w", err)
	}

	// Unmarshal decrypted vault
	var vault v1.Vault
	if err := proto.Unmarshal(decryptedData, &vault); err != nil {
		return nil, fmt.Errorf("failed to unmarshal decrypted vault: %w", err)
	}

	return &vault, nil
}

// decryptAES decrypts data using AES-GCM
func decryptAES(encryptedData string, key []byte) ([]byte, error) {
	// Decode base64 encrypted data
	ciphertext, err := base64.StdEncoding.DecodeString(encryptedData)
	if err != nil {
		return nil, fmt.Errorf("failed to decode encrypted data: %w", err)
	}

	// Create AES cipher
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("failed to create AES cipher: %w", err)
	}

	// Create GCM
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %w", err)
	}

	// Extract nonce and ciphertext
	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil, fmt.Errorf("ciphertext too short")
	}

	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]

	// Decrypt
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt: %w", err)
	}

	return plaintext, nil
}

// ValidateVault performs validation checks on a vault
func ValidateVault(vaultInfo *VaultInfo) []string {
	var issues []string

	if vaultInfo.Name == "" {
		issues = append(issues, "vault name is empty")
	}

	if vaultInfo.PublicKeyECDSA == "" && vaultInfo.PublicKeyEDDSA == "" {
		issues = append(issues, "no public keys found")
	}

	if len(vaultInfo.KeyShares) == 0 {
		issues = append(issues, "no key shares found")
	}

	if vaultInfo.HexChainCode == "" {
		issues = append(issues, "hex chain code is missing")
	}

	if vaultInfo.LocalPartyKey == "" {
		issues = append(issues, "local party key is missing")
	}

	return issues
}

// GetKeySharesInfo returns key shares information in human-readable format
func GetKeySharesInfo(vaultInfo *VaultInfo) string {
	var sb strings.Builder

	sb.WriteString("Key Share Information:\n")
	for i, share := range vaultInfo.KeyShares {
		sb.WriteString(fmt.Sprintf("  Share %d: %s (%s)\n", i+1, share.PublicKey, share.KeyType))
	}

	return sb.String()
}

// ExportToJSON exports vault info to JSON format
func ExportToJSON(vaultInfo *VaultInfo, writer io.Writer, pretty bool) error {
	var data []byte
	var err error

	if pretty {
		data, err = json.MarshalIndent(vaultInfo, "", "  ")
	} else {
		data, err = json.Marshal(vaultInfo)
	}

	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	_, err = writer.Write(data)
	if err != nil {
		return fmt.Errorf("failed to write JSON: %w", err)
	}

	return nil
}

// GetSummary returns a human-readable summary of the vault
func GetSummary(vaultInfo *VaultInfo) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("Vault: %s\n", vaultInfo.Name))
	sb.WriteString(fmt.Sprintf("File: %s\n", vaultInfo.FilePath))
	sb.WriteString(fmt.Sprintf("Encrypted: %t\n", vaultInfo.IsEncrypted))
	sb.WriteString(fmt.Sprintf("Version: %d\n", vaultInfo.Version))
	sb.WriteString(fmt.Sprintf("Local Party: %s\n", vaultInfo.LocalPartyKey))

	if vaultInfo.PublicKeyECDSA != "" {
		sb.WriteString(fmt.Sprintf("ECDSA Public Key: %s\n", vaultInfo.PublicKeyECDSA))
	}
	if vaultInfo.PublicKeyEDDSA != "" {
		sb.WriteString(fmt.Sprintf("EDDSA Public Key: %s\n", vaultInfo.PublicKeyEDDSA))
	}
	if vaultInfo.HexChainCode != "" {
		sb.WriteString(fmt.Sprintf("Hex Chain Code: %s\n", vaultInfo.HexChainCode))
	}

	sb.WriteString(fmt.Sprintf("Key Shares: %d\n", len(vaultInfo.KeyShares)))
	for i, share := range vaultInfo.KeyShares {
		sb.WriteString(fmt.Sprintf("  Share %d: %s (%s)\n", i+1, share.PublicKey[:16]+"...", share.KeyType))
	}

	return sb.String()
}

// ParseVaultContentDirect parses vault content directly from string (for FFI)
func ParseVaultContentDirect(content, fileName string) (*VaultInfo, error) {
	// Decode base64 content
	rawContent, err := base64.StdEncoding.DecodeString(content)
	if err != nil {
		return nil, fmt.Errorf("error decoding base64 content: %w", err)
	}

	// Unmarshal vault container
	var vaultContainer v1.VaultContainer
	if err := proto.Unmarshal(rawContent, &vaultContainer); err != nil {
		return nil, fmt.Errorf("error unmarshalling vault container: %w", err)
	}

	// Handle encrypted vs unencrypted vaults
	var vault *v1.Vault
	if vaultContainer.IsEncrypted {
		// For FFI, we'll need to handle encryption differently
		// For now, return an error for encrypted vaults
		return nil, fmt.Errorf("encrypted vaults not supported in direct parsing mode")
	} else {
		vaultData, err := base64.StdEncoding.DecodeString(vaultContainer.Vault)
		if err != nil {
			return nil, fmt.Errorf("error decoding vault data: %w", err)
		}
		vault = &v1.Vault{}
		if err := proto.Unmarshal(vaultData, vault); err != nil {
			return nil, fmt.Errorf("error unmarshalling vault: %w", err)
		}
	}

	// Build vault info
	vaultInfo := &VaultInfo{
		Name:           vault.Name,
		PublicKeyECDSA: vault.PublicKeyEcdsa,
		PublicKeyEDDSA: vault.PublicKeyEddsa,
		HexChainCode:   vault.HexChainCode,
		LocalPartyKey:  vault.LocalPartyId,
		IsEncrypted:    vaultContainer.IsEncrypted,
		Version:        0, // Version field doesn't exist in v1.Vault
		CreatedAt:      getTimestamp(vault.CreatedAt),
		FilePath:       fileName, // Use provided filename instead of file path
		Metadata:       make(map[string]string),
	}

	// Extract key share information
	for _, keyShare := range vault.KeyShares {
		keyType := "ECDSA"
		if keyShare.PublicKey == vault.PublicKeyEddsa {
			keyType = "EDDSA"
		}
		vaultInfo.KeyShares = append(vaultInfo.KeyShares, KeyShareInfo{
			PublicKey: keyShare.PublicKey,
			KeyType:   keyType,
		})
	}

	return vaultInfo, nil
}

// ParseVaultFromBytes parses vault content directly from bytes (for WASM)
func ParseVaultFromBytes(data []byte) (*VaultInfo, error) {
	// .vult files contain base64-encoded data, so convert bytes to string
	base64Content := string(data)
	return ParseVaultContentDirect(base64Content, "uploaded-file.vult")
}

// IsValidVultFile checks if the given content is a valid .vult file
// This function validates the structure regardless of file extension
func IsValidVultFile(content string) (bool, error) {
	// Step 1: Check if content is valid base64
	if !isValidBase64(content) {
		return false, fmt.Errorf("content is not valid base64")
	}

	// Step 2: Try to decode base64
	rawContent, err := base64.StdEncoding.DecodeString(content)
	if err != nil {
		return false, fmt.Errorf("failed to decode base64: %w", err)
	}

	// Step 3: Check if it's a valid protobuf VaultContainer
	var vaultContainer v1.VaultContainer
	if err := proto.Unmarshal(rawContent, &vaultContainer); err != nil {
		return false, fmt.Errorf("failed to unmarshal VaultContainer: %w", err)
	}

	// Step 4: Validate VaultContainer structure
	if vaultContainer.Vault == "" {
		return false, fmt.Errorf("VaultContainer has empty vault data")
	}

	// Step 5: Try to decode and validate the inner vault
	if !vaultContainer.IsEncrypted {
		// For unencrypted vaults, validate the inner vault structure
		vaultData, err := base64.StdEncoding.DecodeString(vaultContainer.Vault)
		if err != nil {
			return false, fmt.Errorf("failed to decode inner vault data: %w", err)
		}

		var vault v1.Vault
		if err := proto.Unmarshal(vaultData, &vault); err != nil {
			return false, fmt.Errorf("failed to unmarshal inner vault: %w", err)
		}

		// Step 6: Validate essential vault fields
		if err := validateVaultStructure(&vault); err != nil {
			return false, fmt.Errorf("vault structure validation failed: %w", err)
		}
	}
	// For encrypted vaults, we can't validate the inner structure without password
	// But we can confirm it's a valid VaultContainer with encrypted flag

	return true, nil
}

// IsValidVultFileBytes checks if the given byte content is a valid .vult file
func IsValidVultFileBytes(data []byte) (bool, error) {
	return IsValidVultFile(string(data))
}

// validateVaultStructure validates the essential structure of a Vault protobuf
func validateVaultStructure(vault *v1.Vault) error {
	if vault.Name == "" {
		return fmt.Errorf("vault name is empty")
	}

	if vault.PublicKeyEcdsa == "" && vault.PublicKeyEddsa == "" {
		return fmt.Errorf("no public keys found")
	}

	if len(vault.KeyShares) == 0 {
		return fmt.Errorf("no key shares found")
	}

	if vault.HexChainCode == "" {
		return fmt.Errorf("hex chain code is missing")
	}

	if vault.LocalPartyId == "" {
		return fmt.Errorf("local party ID is missing")
	}

	// Validate public key formats (hex strings)
	if vault.PublicKeyEcdsa != "" && !isValidHexString(vault.PublicKeyEcdsa) {
		return fmt.Errorf("invalid ECDSA public key format")
	}

	if vault.PublicKeyEddsa != "" && !isValidHexString(vault.PublicKeyEddsa) {
		return fmt.Errorf("invalid EDDSA public key format")
	}

	if !isValidHexString(vault.HexChainCode) {
		return fmt.Errorf("invalid hex chain code format")
	}

	// Validate key shares
	for i, keyShare := range vault.KeyShares {
		if keyShare.PublicKey == "" {
			return fmt.Errorf("key share %d has empty public key", i+1)
		}
		if keyShare.Keyshare == "" {
			return fmt.Errorf("key share %d has empty keyshare data", i+1)
		}
		if !isValidHexString(keyShare.PublicKey) {
			return fmt.Errorf("key share %d has invalid public key format", i+1)
		}
	}

	return nil
}

// isValidBase64 checks if a string is valid base64
func isValidBase64(s string) bool {
	// Remove whitespace and check if it's valid base64
	s = strings.ReplaceAll(s, "\n", "")
	s = strings.ReplaceAll(s, "\r", "")
	s = strings.ReplaceAll(s, " ", "")
	s = strings.ReplaceAll(s, "\t", "")

	if len(s) == 0 {
		return false
	}

	// Base64 length should be multiple of 4
	if len(s)%4 != 0 {
		return false
	}

	// Check if all characters are valid base64
	validBase64Regex := regexp.MustCompile(`^[A-Za-z0-9+/]*={0,2}$`)
	return validBase64Regex.MatchString(s)
}

// isValidHexString checks if a string is a valid hexadecimal string
func isValidHexString(s string) bool {
	if len(s) == 0 {
		return false
	}

	// Remove 0x prefix if present
	if strings.HasPrefix(s, "0x") {
		s = s[2:]
	}

	// Check if all characters are valid hex
	validHexRegex := regexp.MustCompile(`^[0-9a-fA-F]+$`)
	return validHexRegex.MatchString(s)
}

// ValidateVultFileFromPath checks if a file at the given path is a valid .vult file
func ValidateVultFileFromPath(filePath string) (bool, error) {
	// Validate file path for security
	if err := validateSafePath(filePath); err != nil {
		return false, fmt.Errorf("unsafe file path: %w", err)
	}

	// Read file content
	fileContent, err := os.ReadFile(filePath)
	if err != nil {
		return false, fmt.Errorf("failed to read file: %w", err)
	}

	return IsValidVultFile(string(fileContent))
}

// getTimestamp converts protobuf timestamp to Unix timestamp
func getTimestamp(ts *timestamppb.Timestamp) int64 {
	if ts == nil {
		return 0
	}
	return ts.GetSeconds()
}

// validateSafePath performs basic security checks on file paths
func validateSafePath(path string) error {
	// Clean the path to resolve any .. or . elements
	cleanPath := filepath.Clean(path)
	
	// Check for dangerous patterns
	dangerousPaths := []string{
		"/etc/passwd", "/etc/shadow", "/etc/hosts",
		"/proc/", "/sys/", "/dev/",
		"/Windows/System32", "/Windows/system32",
		"C:\\Windows\\System32", "c:\\windows\\system32",
	}
	
	// Convert to lowercase for case-insensitive comparison on Windows
	lowerPath := strings.ToLower(cleanPath)
	
	for _, dangerous := range dangerousPaths {
		if strings.Contains(lowerPath, strings.ToLower(dangerous)) {
			return fmt.Errorf("access to system path %q is not allowed", dangerous)
		}
	}
	
	// Check for directory traversal attempts
	if strings.Contains(cleanPath, "..") {
		return fmt.Errorf("directory traversal detected in path")
	}
	
	// Platform-specific additional checks
	if runtime.GOOS == "windows" {
		// Check for Windows-specific dangerous paths
		if strings.HasPrefix(lowerPath, "\\\\.\\") {
			return fmt.Errorf("device path access not allowed")
		}
	}
	
	return nil
}

// ValidateSafeOutputPath validates paths for file creation/writing
func ValidateSafeOutputPath(path string) error {
	// First run basic path validation
	if err := validateSafePath(path); err != nil {
		return err
	}
	
	// Additional checks for output files
	cleanPath := filepath.Clean(path)
	
	// Don't allow writing to system directories
	systemDirs := []string{
		"/bin", "/sbin", "/usr/bin", "/usr/sbin",
		"/boot", "/lib", "/lib64",
		"C:\\Program Files", "C:\\Windows",
	}
	
	lowerPath := strings.ToLower(cleanPath)
	for _, sysDir := range systemDirs {
		if strings.HasPrefix(lowerPath, strings.ToLower(sysDir)) {
			return fmt.Errorf("writing to system directory %q is not allowed", sysDir)
		}
	}
	
	return nil
}


