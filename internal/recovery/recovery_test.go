package recovery

import (
	"os"
	"strings"
	"testing"
)

// TestRecoverPrivateKeys_GG20Integration - The main test that actually matters
// Tests end-to-end GG20 recovery with real vault files
func TestRecoverPrivateKeys_GG20Integration(t *testing.T) {
	// Test with actual GG20 test fixtures
	vaultFiles := []string{
		"../../test/fixtures/testGG20-part1of2.vult",
		"../../test/fixtures/testGG20-part2of2.vult",
	}

	// Skip if test files don't exist
	for _, file := range vaultFiles {
		if !fileExists(file) {
			t.Skipf("Test fixture not found: %s", file)
			return
		}
	}

	recoveredKeys, err := RecoverPrivateKeys(vaultFiles, 2, "")
	if err != nil {
		t.Fatalf("Recovery failed: %v", err)
	}

	if len(recoveredKeys) == 0 {
		t.Fatal("No keys recovered")
	}

	// Validate that we get both ECDSA and EdDSA keys
	foundBitcoin := false
	foundSolana := false
	foundSUI := false

	for _, key := range recoveredKeys {
		// Validate required fields
		if key.PrivateKey == "" {
			t.Error("Private key should not be empty")
		}
		if key.Address == "" {
			t.Error("Address should not be empty")
		}
		if key.DerivePath == "" {
			t.Error("Derive path should not be empty")
		}

		// Check for key chains
		switch key.Chain {
		case "bitcoin":
			foundBitcoin = true
		case "solana":
			foundSolana = true
		case "sui":
			foundSUI = true
			// Verify SUI address is correct (not a placeholder)
			expectedSUIAddr := "0xe36ca893894810713425724d15aedc3bf928013852cb1cd2d3676b1579f7501a"
			if key.Address != expectedSUIAddr {
				t.Errorf("SUI address should be %s, got %s", expectedSUIAddr, key.Address)
			}
		}
	}

	// Verify we recovered the essential chains
	if !foundBitcoin {
		t.Error("Should recover Bitcoin key from GG20 test fixture")
	}
	if !foundSolana {
		t.Error("Should recover Solana key from GG20 test fixture")
	}
	if !foundSUI {
		t.Error("Should recover SUI key from GG20 test fixture")
	}

	t.Logf("Successfully recovered %d keys", len(recoveredKeys))
}

// TestRecoverPrivateKeys_InsufficientShares - Basic error handling test
func TestRecoverPrivateKeys_InsufficientShares(t *testing.T) {
	_, err := RecoverPrivateKeys([]string{"file1"}, 2, "")
	if err == nil {
		t.Error("Expected error for insufficient shares")
	}

	if !strings.Contains(err.Error(), "insufficient shares") {
		t.Errorf("Expected 'insufficient shares' in error, got: %v", err)
	}
}

// TestRecoverPrivateKeys_InvalidFiles - Basic error handling test
func TestRecoverPrivateKeys_InvalidFiles(t *testing.T) {
	_, err := RecoverPrivateKeys([]string{"nonexistent1", "nonexistent2"}, 2, "")
	if err == nil {
		t.Error("Expected error from invalid file paths")
	}
	// Just verify we get some kind of file-related error
	if !strings.Contains(strings.ToLower(err.Error()), "file") {
		t.Errorf("Expected file-related error, got: %v", err)
	}
}

// Helper function to check if a file exists
func fileExists(filename string) bool {
	_, err := os.Stat(filename)
	return err == nil
}
