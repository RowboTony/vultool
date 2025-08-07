package main

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"os/exec"
	"strings"
)

// Expected addresses from the user
const (
	expectedBTCAddress = "bc1qwzpjqun2rfga2fu0ld4wlk27tw2dk3ljxh2yyl"
	expectedETHAddress = "0x7e710a170D29EdB42D05b9417bE07DD8F1779CA3"
	expectedSOLAddress = "5DCrTjNsBUuhhWFpbH1LAuPenrxwHy319CPz7e6DUpRd"
)

type RecoveredKey struct {
	Chain   string `json:"chain"`
	Address string `json:"address"`
	WIF     string `json:"wif,omitempty"`
	PrivKey string `json:"privkey,omitempty"`
}

func main() {
	fmt.Println("=== Vultool Key Recovery Verification ===")
	fmt.Println()

	// Step 1: Recover keys using vultool
	fmt.Println("Step 1: Recovering keys from vault shares...")
	cmd := exec.Command("./vultool", "recover",
		"test/fixtures/qa-fast-share1of2.vult",
		"test/fixtures/qa-fast-share2of2.vult",
		"--threshold", "2",
		"--password", "vultcli01",
		"--json")

	output, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			log.Fatalf("Command failed: %v\nStderr: %s", err, exitErr.Stderr)
		}
		log.Fatalf("Failed to run vultool recover: %v", err)
	}

	// Parse the JSON output
	var recoveredKeys []RecoveredKey
	if err := json.Unmarshal(output, &recoveredKeys); err != nil {
		log.Fatalf("Failed to parse JSON output: %v\nOutput: %s", err, output)
	}

	fmt.Printf("Recovered %d keys\n\n", len(recoveredKeys))

	// Step 2: Find and verify addresses
	var btcKey, ethKey, solKey *RecoveredKey
	for i := range recoveredKeys {
		switch recoveredKeys[i].Chain {
		case "BTC":
			btcKey = &recoveredKeys[i]
		case "ETH":
			ethKey = &recoveredKeys[i]
		case "SOL":
			solKey = &recoveredKeys[i]
		}
	}

	fmt.Println("Step 2: Verifying recovered addresses...")
	fmt.Println()

	// Check Bitcoin
	if btcKey != nil {
		fmt.Printf("Bitcoin:\n")
		fmt.Printf("  Recovered: %s\n", btcKey.Address)
		fmt.Printf("  Expected:  %s\n", expectedBTCAddress)
		fmt.Printf("  Match: %v\n", btcKey.Address == expectedBTCAddress)
		if btcKey.WIF != "" {
			fmt.Printf("  WIF: %s...\n", btcKey.WIF[:10])
			// Verify WIF -> address derivation
			verifyBitcoinWIF(btcKey.WIF)
		}
		fmt.Println()
	} else {
		fmt.Println("Bitcoin: No key recovered")
	}

	// Check Ethereum
	if ethKey != nil {
		fmt.Printf("Ethereum:\n")
		fmt.Printf("  Recovered: %s\n", ethKey.Address)
		fmt.Printf("  Expected:  %s\n", expectedETHAddress)
		fmt.Printf("  Match: %v\n", strings.EqualFold(ethKey.Address, expectedETHAddress))
		if ethKey.PrivKey != "" {
			// Verify privkey -> address derivation
			verifyEthereumPrivKey(ethKey.PrivKey)
		}
		fmt.Println()
	} else {
		fmt.Println("Ethereum: No key recovered")
	}

	// Check Solana
	if solKey != nil {
		fmt.Printf("Solana:\n")
		fmt.Printf("  Recovered: %s\n", solKey.Address)
		fmt.Printf("  Expected:  %s\n", expectedSOLAddress)
		fmt.Printf("  Match: %v\n", solKey.Address == expectedSOLAddress)
		if solKey.PrivKey != "" {
			fmt.Printf("  PrivKey: %s...\n", solKey.PrivKey[:10])
		}
		fmt.Println()
	} else {
		fmt.Println("Solana: No key recovered")
	}

	// Step 3: Analysis
	fmt.Println("Step 3: Analysis")
	fmt.Println("================")

	allMatch := true
	if btcKey != nil && btcKey.Address != expectedBTCAddress {
		allMatch = false
		fmt.Println("❌ Bitcoin address does not match expected")
	}
	if ethKey != nil && !strings.EqualFold(ethKey.Address, expectedETHAddress) {
		allMatch = false
		fmt.Println("❌ Ethereum address does not match expected")
	}
	if solKey != nil && solKey.Address != expectedSOLAddress {
		allMatch = false
		fmt.Println("❌ Solana address does not match expected")
	}

	if allMatch && btcKey != nil && ethKey != nil && solKey != nil {
		fmt.Println("✅ All addresses match! The recovery is working correctly.")
	} else if !allMatch {
		fmt.Println("\n⚠️  Some addresses don't match. This suggests:")
		fmt.Println("  1. The current recovery implementation is using simplified/incorrect TSS reconstruction")
		fmt.Println("  2. The key derivation paths may be incorrect")
		fmt.Println("  3. The address encoding might be wrong")
		fmt.Println("\nThe proper fix requires:")
		fmt.Println("  - Using the official Vultisig mobile-tss-lib for TSS reconstruction")
		fmt.Println("  - Implementing proper VSS share combination")
		fmt.Println("  - Following the exact derivation paths used by Vultisig")
	}
}

func verifyBitcoinWIF(wif string) {
	// Basic WIF validation - decode and check if it produces a valid private key
	if len(wif) < 51 {
		fmt.Printf("    WIF appears invalid (too short): %d chars\n", len(wif))
		return
	}

	// Check WIF format (should start with 'K' or 'L' for mainnet compressed keys)
	if wif[0] != 'K' && wif[0] != 'L' && wif[0] != '5' {
		fmt.Printf("    WIF format issue: starts with '%c' (expected K/L for compressed, 5 for uncompressed)\n", wif[0])
	}
}

func verifyEthereumPrivKey(privKeyHex string) {
	// Derive Ethereum address from private key
	privKeyBytes, err := hex.DecodeString(strings.TrimPrefix(privKeyHex, "0x"))
	if err != nil {
		fmt.Printf("    Error decoding private key: %v\n", err)
		return
	}

	if len(privKeyBytes) != 32 {
		fmt.Printf("    Invalid private key length: %d bytes (expected 32)\n", len(privKeyBytes))
		return
	}

	// For verification, we'd need to:
	// 1. Get public key from private key using secp256k1
	// 2. Hash with Keccak256
	// 3. Take last 20 bytes
	// This requires elliptic curve operations which we'll skip for now
	fmt.Printf("    Private key length valid: 32 bytes\n")
}

// Helper functions for address derivation would go here
// For full implementation, we'd need:
// - secp256k1 for public key derivation
// - Proper Bitcoin SegWit address encoding
// - Ethereum Keccak256 hashing
// - Solana Ed25519 operations
