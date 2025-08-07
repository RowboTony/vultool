package main

import (
	"crypto/ed25519"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
)

func main() {
	// The recovered private key from TSS
	privateKeyHex := "06f95aec53fcf4c2cf4349cfabe6912dc138cf3d4127c9d3604c7a9fcc5240e7"
	vaultPublicKeyHex := "c6da2ad7b18728f6481d747a7335fd52a5eed82f3c3d95a51deed03399c5c0b6"
	
	privateKeyBytes, _ := hex.DecodeString(privateKeyHex)
	publicKeyBytes, _ := hex.DecodeString(vaultPublicKeyHex)
	
	fmt.Println("Different wallet import formats for Solana:")
	fmt.Println("============================================")
	
	// Format 1: Just the 32-byte seed in base64 (some wallets only want this)
	seedOnlyBase64 := base64.StdEncoding.EncodeToString(privateKeyBytes)
	fmt.Printf("\n1. Seed only (32 bytes base64):\n%s\n", seedOnlyBase64)
	
	// Format 2: Full 64-byte keypair in base64 (what we currently generate)
	keypair := make([]byte, 64)
	copy(keypair[:32], privateKeyBytes)
	copy(keypair[32:], publicKeyBytes)
	keypairBase64 := base64.StdEncoding.EncodeToString(keypair)
	fmt.Printf("\n2. Full keypair (64 bytes base64) - CURRENT:\n%s\n", keypairBase64)
	
	// Format 3: JSON array of the 64-byte keypair
	var jsonArray []int
	for _, b := range keypair {
		jsonArray = append(jsonArray, int(b))
	}
	jsonBytes, _ := json.Marshal(jsonArray)
	fmt.Printf("\n3. JSON array (64 bytes) - CURRENT:\n%s\n", string(jsonBytes))
	
	// Format 4: Using standard ed25519 to generate the keypair (NOT TSS)
	standardPriv := ed25519.NewKeyFromSeed(privateKeyBytes)
	standardPub := standardPriv.Public().(ed25519.PublicKey)
	standardKeypair := make([]byte, 64)
	copy(standardKeypair[:32], privateKeyBytes)
	copy(standardKeypair[32:], standardPub)
	standardBase64 := base64.StdEncoding.EncodeToString(standardKeypair)
	fmt.Printf("\n4. Standard Ed25519 keypair (64 bytes base64) - WRONG:\n%s\n", standardBase64)
	fmt.Printf("   This would create address from pubkey: %x\n", standardPub)
	
	// Format 5: Hex string of the seed (some wallets might accept this)
	fmt.Printf("\n5. Seed as hex string (32 bytes):\n%s\n", privateKeyHex)
	
	// Format 6: Full Ed25519 private key (64 bytes) from standard generation
	fullStandardPriv := ed25519.NewKeyFromSeed(privateKeyBytes)
	fullStandardBase64 := base64.StdEncoding.EncodeToString(fullStandardPriv)
	fmt.Printf("\n6. Standard Ed25519 full private key (64 bytes base64) - WRONG:\n%s\n", fullStandardBase64)
	
	fmt.Println("\n============================================")
	fmt.Printf("Expected Solana address: EPEg1C2pEwiEbPaBTuyydnvGpZoa6y3iXVVNzv7JYT8H\n")
	fmt.Printf("This comes from public key: %s\n", vaultPublicKeyHex)
}
