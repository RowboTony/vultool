package main

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/rowbotony/vultool/internal/recovery"
)

func main() {
	fmt.Println("=== Debug TSS Recovery ===")
	fmt.Println()

	vaultFiles := []string{
		"test/fixtures/qa-fast-share1of2.vult",
		"test/fixtures/qa-fast-share2of2.vult",
	}
	password := "vultcli01"
	threshold := 2

	fmt.Println("Step 1: Attempting ECDSA recovery...")
	ecdsaResult, err := recovery.ReconstructTSSKey(vaultFiles, password, recovery.ECDSA)
	if err != nil {
		fmt.Printf("ECDSA recovery error: %v\n", err)
	} else {
		fmt.Println("ECDSA recovery successful!")
		jsonBytes, _ := json.MarshalIndent(ecdsaResult, "  ", "  ")
		fmt.Printf("Result:\n%s\n", jsonBytes)
	}

	fmt.Println("\nStep 2: Attempting EdDSA recovery...")
	eddsaResult, err := recovery.ReconstructTSSKey(vaultFiles, password, recovery.EdDSA)
	if err != nil {
		fmt.Printf("EdDSA recovery error: %v\n", err)
	} else {
		fmt.Println("EdDSA recovery successful!")
		jsonBytes, _ := json.MarshalIndent(eddsaResult, "  ", "  ")
		fmt.Printf("Result:\n%s\n", jsonBytes)
	}

	fmt.Println("\nStep 3: Attempting full recovery via RecoverPrivateKeys...")
	keys, err := recovery.RecoverPrivateKeys(vaultFiles, threshold, password)
	if err != nil {
		log.Fatalf("Full recovery failed: %v", err)
	}

	fmt.Printf("\nRecovered %d keys:\n", len(keys))
	for _, key := range keys {
		fmt.Printf("\n%s:\n", key.Chain)
		fmt.Printf("  Address: %s\n", key.Address)
		if key.WIF != "" {
			fmt.Printf("  WIF: %s...\n", key.WIF[:10])
		}
		if key.PrivateKey != "" {
			fmt.Printf("  PrivKey: %s...\n", key.PrivateKey[:10])
		}
	}
}
