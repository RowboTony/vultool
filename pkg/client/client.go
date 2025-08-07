// Package client provides a public API for vultool functionality
// This package is intended for consumption by other Go applications
package client

import (
	"github.com/rowbotony/vultool/internal/vault"
)

// VaultInfo represents parsed vault information for external clients
type VaultInfo = vault.VaultInfo

// KeyShareInfo represents key share information for external clients
type KeyShareInfo = vault.KeyShareInfo

// ParseVaultFile parses a .vult file and returns vault information
func ParseVaultFile(filePath string) (*VaultInfo, error) {
	return vault.ParseVaultFile(filePath)
}

// ParseVaultFileWithPassword parses a .vult file with a provided password
func ParseVaultFileWithPassword(filePath, password string) (*VaultInfo, error) {
	return vault.ParseVaultFileWithPassword(filePath, password)
}

// ValidateVault performs validation checks on a vault
func ValidateVault(vaultInfo *VaultInfo) []string {
	return vault.ValidateVault(vaultInfo)
}

// IsValidVultFile checks if the given content is a valid .vult file
func IsValidVultFile(content string) (bool, error) {
	return vault.IsValidVultFile(content)
}

// ValidateVultFileFromPath checks if a file at the given path is a valid .vult file
func ValidateVultFileFromPath(filePath string) (bool, error) {
	return vault.ValidateVultFileFromPath(filePath)
}

// ParseVaultFromBytes parses vault content directly from bytes
func ParseVaultFromBytes(data []byte) (*VaultInfo, error) {
	return vault.ParseVaultFromBytes(data)
}
