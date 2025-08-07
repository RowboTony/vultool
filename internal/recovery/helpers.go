package recovery

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io"
	"os"

	v1 "github.com/vultisig/commondata/go/vultisig/vault/v1"
	"google.golang.org/protobuf/proto"
)

// readFileContent reads the content of a file
func readFileContent(filePath string) ([]byte, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	return io.ReadAll(file)
}

// decryptVaultWithPassword decrypts an encrypted vault container using the provided password
func decryptVaultWithPassword(vaultContainer *v1.VaultContainer, password string) (*v1.Vault, error) {
	if !vaultContainer.IsEncrypted {
		return nil, fmt.Errorf("vault is not encrypted")
	}

	// VaultContainer doesn't have a Salt field in the protobuf
	// Use SHA256 of password directly (matching Vultisig's approach)
	hash := sha256.Sum256([]byte(password))
	key := hash[:]

	// Decrypt the vault data
	encryptedData, err := base64.StdEncoding.DecodeString(vaultContainer.Vault)
	if err != nil {
		return nil, fmt.Errorf("failed to decode encrypted vault data: %w", err)
	}

	// Extract nonce (first 12 bytes) and ciphertext
	if len(encryptedData) < 12 {
		return nil, fmt.Errorf("encrypted data too short")
	}

	nonce := encryptedData[:12]
	ciphertext := encryptedData[12:]

	// Create cipher
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %w", err)
	}

	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %w", err)
	}

	// Decrypt
	plaintext, err := aesgcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt vault: %w", err)
	}

	// Unmarshal the decrypted vault
	var vault v1.Vault
	if err := proto.Unmarshal(plaintext, &vault); err != nil {
		return nil, fmt.Errorf("failed to unmarshal decrypted vault: %w", err)
	}

	return &vault, nil
}

// DecryptVault decrypts vault data with a password (legacy format support)
func DecryptVault(password string, encryptedData []byte) ([]byte, error) {
	// Generate key from password
	hash := sha256.Sum256([]byte(password))
	key := hash[:]

	// Extract nonce (first 12 bytes) and ciphertext
	if len(encryptedData) < 12 {
		return nil, fmt.Errorf("encrypted data too short")
	}

	nonce := encryptedData[:12]
	ciphertext := encryptedData[12:]

	// Create cipher
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %w", err)
	}

	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %w", err)
	}

	// Decrypt
	plaintext, err := aesgcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt: %w", err)
	}

	return plaintext, nil
}
