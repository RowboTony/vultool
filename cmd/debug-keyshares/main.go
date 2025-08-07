package main

import (
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/rowbotony/vultool/internal/vault"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run cmd/debug-keyshares/main.go <vault-file> [password]")
		os.Exit(1)
	}

	vaultFile := os.Args[1]
	password := ""
	if len(os.Args) > 2 {
		password = os.Args[2]
	}

	// Parse vault file
	vaultInfo, err := vault.ParseVaultFileWithPassword(vaultFile, password)
	if err != nil {
		fmt.Printf("Error parsing vault: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("=== Vault: %s ===\n", vaultInfo.Name)
	fmt.Printf("Local Party ID: %s\n", vaultInfo.LocalPartyKey)
	fmt.Printf("ECDSA Public Key: %s\n", vaultInfo.PublicKeyECDSA)
	fmt.Printf("EDDSA Public Key: %s\n", vaultInfo.PublicKeyEDDSA)
	fmt.Printf("Hex Chain Code: %s\n", vaultInfo.HexChainCode)
	fmt.Printf("\n")

	// Now we need to get the raw keyshare data
	// We'll need to re-parse the vault to get the actual keyshare field
	fmt.Println("=== Key Shares ===")
	
	// We need to access the raw protobuf data to get the actual keyshare field
	// Let's create a debug function that preserves the raw keyshare data
	debugVaultWithKeyshares(vaultFile, password)
}

func debugVaultWithKeyshares(filePath, password string) {
	// This function will directly parse the protobuf to access raw keyshare data
	// We'll need to modify the vault parser or create a debug version
	
	fmt.Println("\nTo examine the actual keyshare data, we need to modify the vault parser")
	fmt.Println("to preserve the raw keyshare field from the protobuf.")
	fmt.Println("\nThe current VaultInfo structure only stores public keys, not the actual share data.")
	fmt.Println("\nNext steps:")
	fmt.Println("1. Modify internal/vault/parser.go to preserve keyshare data")
	fmt.Println("2. Parse the keyshare field (likely base64-encoded JSON or protobuf)")
	fmt.Println("3. Understand the TSS share structure")
	
	// For now, let's try to decode what we think might be in there
	fmt.Println("\n=== Attempting to decode keyshare patterns ===")
	
	// The keyshare field is likely one of:
	// 1. Base64-encoded JSON with TSS parameters
	// 2. Base64-encoded protobuf message
	// 3. Hex-encoded binary data
	
	// We'll need to actually get the raw keyshare data first
	// This requires modifying the parser
}

func tryDecodeAsJSON(data string) {
	decoded, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		return
	}
	
	var result interface{}
	if err := json.Unmarshal(decoded, &result); err == nil {
		pretty, _ := json.MarshalIndent(result, "", "  ")
		fmt.Printf("Decoded as JSON:\n%s\n", pretty)
	}
}

func tryDecodeAsHex(data string) {
	if strings.HasPrefix(strings.ToLower(data), "0x") {
		data = data[2:]
	}
	
	decoded, err := hex.DecodeString(data)
	if err == nil && len(decoded) > 0 {
		fmt.Printf("Decoded as hex (%d bytes):\n", len(decoded))
		fmt.Printf("First 32 bytes: %x\n", decoded[:min(32, len(decoded))])
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
