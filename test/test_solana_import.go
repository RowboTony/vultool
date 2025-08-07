package main

import (
	"crypto/ed25519"
	"encoding/base64"
	"encoding/hex"
	"fmt"
)

func main() {
	// The base64 wallet format from recovery
	walletBase64 := "Bvla7FP89MLPQ0nPq+aRLcE4zz1BJ8nTYEx6n8xSQOfG2irXsYco9kgddHpzNf1Spe7YLzw9laUd7tAzmcXAtg=="
	
	// Decode the base64
	walletBytes, err := base64.StdEncoding.DecodeString(walletBase64)
	if err != nil {
		panic(err)
	}
	
	fmt.Printf("Wallet bytes length: %d\n", len(walletBytes))
	fmt.Printf("Private key (first 32): %x\n", walletBytes[:32])
	fmt.Printf("Public key (last 32):  %x\n", walletBytes[32:])
	
	// The expected public key from the vault
	expectedPubKey := "c6da2ad7b18728f6481d747a7335fd52a5eed82f3c3d95a51deed03399c5c0b6"
	fmt.Printf("Expected public key:    %s\n", expectedPubKey)
	
	// Verify the public key matches
	actualPubKeyHex := hex.EncodeToString(walletBytes[32:])
	if actualPubKeyHex == expectedPubKey {
		fmt.Println("✅ Public key matches vault!")
	} else {
		fmt.Println("❌ Public key mismatch!")
	}
	
	// Generate a fresh keypair from the seed to see what standard ed25519 would produce
	seed := walletBytes[:32]
	freshPriv := ed25519.NewKeyFromSeed(seed)
	freshPub := freshPriv.Public().(ed25519.PublicKey)
	
	fmt.Printf("\nFresh Ed25519 generation from seed:\n")
	fmt.Printf("Fresh private key: %x\n", freshPriv)
	fmt.Printf("Fresh public key:  %x\n", freshPub)
	
	// The expected Solana address
	expectedAddress := "EPEg1C2pEwiEbPaBTuyydnvGpZoa6y3iXVVNzv7JYT8H"
	fmt.Printf("\nExpected Solana address: %s\n", expectedAddress)
	
	// Verify the Solana address by base58 encoding the public key
	solanaAddress := encodeBase58(walletBytes[32:])
	fmt.Printf("\nDerived Solana address:  %s\n", solanaAddress)
	if solanaAddress == expectedAddress {
		fmt.Println("✅ Solana address matches!")
	} else {
		fmt.Println("❌ Solana address mismatch!")
	}
}

// encodeBase58 encodes a byte slice to base58 using Bitcoin's alphabet
func encodeBase58(input []byte) string {
	alphabet := "123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz"
	
	if len(input) == 0 {
		return ""
	}
	
	// Count leading zeros
	zeros := 0
	for zeros < len(input) && input[zeros] == 0 {
		zeros++
	}
	
	// Convert to big integer
	var result []byte
	for _, b := range input {
		carry := int(b)
		for i := 0; i < len(result); i++ {
			carry += 256 * int(result[i])
			result[i] = byte(carry % 58)
			carry /= 58
		}
		for carry > 0 {
			result = append(result, byte(carry%58))
			carry /= 58
		}
	}
	
	// Convert to base58 characters
	for i := 0; i < len(result); i++ {
		result[i] = alphabet[result[i]]
	}
	
	// Add leading '1's for zeros
	for i := 0; i < zeros; i++ {
		result = append([]byte{'1'}, result...)
	}
	
	// Reverse
	for i, j := 0, len(result)-1; i < j; i, j = i+1, j-1 {
		result[i], result[j] = result[j], result[i]
	}
	
	return string(result)
}
