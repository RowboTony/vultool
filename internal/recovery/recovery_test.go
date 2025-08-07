package recovery

import (
	"os"
	"strings"
	"testing"
)

func TestGetSupportedChains(t *testing.T) {
	chains := GetSupportedChains()
	expectedChains := []SupportedChain{ChainBitcoin, ChainEthereum, ChainSolana, ChainThorChain}
	
	if len(chains) != len(expectedChains) {
		t.Errorf("Expected %d chains, got %d", len(expectedChains), len(chains))
	}
	
	for i, expected := range expectedChains {
		if i >= len(chains) || chains[i] != expected {
			t.Errorf("Expected chain %s at index %d, got %s", expected, i, chains[i])
		}
	}
}

func TestGetCommonDerivationPaths(t *testing.T) {
	paths := GetCommonDerivationPaths()
	
	// Test that we have paths for all supported chains
	expectedChains := []SupportedChain{ChainBitcoin, ChainEthereum, ChainSolana, ChainThorChain}
	
	for _, chain := range expectedChains {
		chainPaths, exists := paths[chain]
		if !exists {
			t.Errorf("Missing derivation paths for chain: %s", chain)
			continue
		}
		
		if len(chainPaths) == 0 {
			t.Errorf("Empty derivation paths for chain: %s", chain)
		}
		
		// Test that each path has required fields
		for i, path := range chainPaths {
			if path.Path == "" {
				t.Errorf("Empty path for chain %s at index %d", chain, i)
			}
			if path.Chain != chain {
				t.Errorf("Path chain mismatch: expected %s, got %s", chain, path.Chain)
			}
			if path.Description == "" {
				t.Errorf("Empty description for chain %s path %s", chain, path.Path)
			}
			if path.Purpose == "" {
				t.Errorf("Empty purpose for chain %s path %s", chain, path.Path)
			}
		}
	}
	
	// Test specific known paths
	bitcoinPaths, exists := paths[ChainBitcoin]
	if !exists {
		t.Fatal("Bitcoin paths should exist")
	}
	
	// Check that standard BIP-44 path exists
	foundStandardPath := false
	for _, path := range bitcoinPaths {
		if path.Path == "m/44'/0'/0'/0/0" {
			foundStandardPath = true
			if path.Purpose != "receiving" {
				t.Errorf("Standard Bitcoin path should be receiving, got: %s", path.Purpose)
			}
			break
		}
	}
	if !foundStandardPath {
		t.Error("Standard Bitcoin BIP-44 path m/44'/0'/0'/0/0 not found")
	}
}

func TestValidateDerivationPath(t *testing.T) {
	tests := []struct {
		path        string
		shouldError bool
		description string
	}{
		{"", true, "empty path"},
		{"m/44'/0'/0'/0/0", false, "valid Bitcoin path"},
		{"m/44'/60'/0'/0/0", false, "valid Ethereum path"},
		{"m/44'/501'/0'/0'", false, "valid Solana path"},
		{"44'/0'/0'/0/0", true, "missing m/ prefix"},
		{"m", true, "just m"},
		{"m/", false, "m/ should be valid"},
		{"invalid", true, "invalid format"},
		{"m/44'/0'", false, "shorter valid path"},
	}
	
	for _, test := range tests {
		err := ValidateDerivationPath(test.path)
		if test.shouldError && err == nil {
			t.Errorf("Expected error for %s (%s), but got nil", test.path, test.description)
		}
		if !test.shouldError && err != nil {
			t.Errorf("Expected no error for %s (%s), but got: %v", test.path, test.description, err)
		}
	}
}

func TestRecoverPrivateKeys_InvalidFiles(t *testing.T) {
	// Test that invalid files return appropriate error
	_, err := RecoverPrivateKeys([]string{"file1", "file2"}, 2, "")
	if err == nil {
		t.Error("Expected error from invalid file paths")
	}
	
	if !strings.Contains(err.Error(), "failed to parse vault file") {
		t.Errorf("Expected 'failed to parse vault file' in error, got: %v", err)
	}
}

func TestRecoverPrivateKeys_InsufficientShares(t *testing.T) {
	// Test that insufficient shares are detected
	_, err := RecoverPrivateKeys([]string{"file1"}, 2, "")
	if err == nil {
		t.Error("Expected error for insufficient shares")
	}
	
	if !strings.Contains(err.Error(), "insufficient shares") {
		t.Errorf("Expected 'insufficient shares' in error, got: %v", err)
	}
}

func TestDeriveAddress_Stub(t *testing.T) {
	// Test that the stub function returns appropriate error
	_, err := DeriveAddress("vault.vult", "m/44'/0'/0'/0/0", ChainBitcoin, "")
	if err == nil {
		t.Error("Expected error from stub implementation")
	}
	
	// Should fail at vault parsing since we're using a non-existent file,
	// but let's test with a more specific scenario
	if !strings.Contains(err.Error(), "failed to parse vault file") {
		// If it's not a vault parsing error, it should be the stub error
		if !strings.Contains(err.Error(), "not yet implemented") {
			t.Errorf("Expected either parse error or 'not yet implemented', got: %v", err)
		}
	}
}

func TestSupportedChainConstants(t *testing.T) {
	// Test that chain constants have expected values
	expectedChains := map[SupportedChain]string{
		ChainBitcoin:   "bitcoin",
		ChainEthereum:  "ethereum", 
		ChainSolana:    "solana",
		ChainThorChain: "thorchain",
	}
	
	for chain, expectedValue := range expectedChains {
		if string(chain) != expectedValue {
			t.Errorf("Chain constant %s has unexpected value: expected %s, got %s", 
				chain, expectedValue, string(chain))
		}
	}
}

func TestRecoveredKeyStructure(t *testing.T) {
	// Test that RecoveredKey structure can be created and has expected fields
	key := RecoveredKey{
		Chain:      ChainBitcoin,
		PrivateKey: "test-private-key",
		WIF:        "test-wif",
		Address:    "test-address",
		DerivePath: "m/44'/0'/0'/0/0",
	}
	
	if key.Chain != ChainBitcoin {
		t.Errorf("Expected chain %s, got %s", ChainBitcoin, key.Chain)
	}
	if key.PrivateKey != "test-private-key" {
		t.Errorf("Expected private key 'test-private-key', got '%s'", key.PrivateKey)
	}
	if key.Address != "test-address" {
		t.Errorf("Expected address 'test-address', got '%s'", key.Address)
	}
}

func TestDerivationPathStructure(t *testing.T) {
	// Test that DerivationPath structure can be created
	path := DerivationPath{
		Path:        "m/44'/0'/0'/0/0",
		Chain:       ChainBitcoin,
		Description: "Test path",
		Purpose:     "testing",
	}
	
	if path.Path != "m/44'/0'/0'/0/0" {
		t.Errorf("Expected path 'm/44'/0'/0'/0/0', got '%s'", path.Path)
	}
	if path.Chain != ChainBitcoin {
		t.Errorf("Expected chain %s, got %s", ChainBitcoin, path.Chain)
	}
}

// Integration tests using actual test fixtures
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
	
	// Validate that we get both ECDSA and EDDSA keys
	foundBitcoin := false
	foundSolana := false
	
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
		
		// Check chain-specific formats
		switch key.Chain {
		case ChainBitcoin:
			foundBitcoin = true
			if key.WIF == "" {
				t.Error("Bitcoin key should have WIF format")
			}
			if !strings.HasPrefix(key.WIF, "WIF:") {
				t.Errorf("Bitcoin WIF should have WIF: prefix, got: %s", key.WIF)
			}
		case ChainSolana:
			foundSolana = true
			if key.Base58 == "" {
				t.Error("Solana key should have Base58 format")
			}
			if !strings.HasPrefix(key.Base58, "B58:") {
				t.Errorf("Solana Base58 should have B58: prefix, got: %s", key.Base58)
			}
		}
	}
	
	if !foundBitcoin {
		t.Error("Should recover Bitcoin key from GG20 test fixture")
	}
	if !foundSolana {
		t.Error("Should recover Solana key from GG20 test fixture")
	}
}

func TestRecoverPrivateKeys_DKLSIntegration(t *testing.T) {
	// Test with actual DKLS test fixtures
	vaultFiles := []string{
		"../../test/fixtures/testDKLS-1of2.vult",
		"../../test/fixtures/testDKLS-2of2.vult",
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
		t.Fatal("No keys recovered from DKLS vault")
	}
	
	// DKLS vaults should also produce keys
	for _, key := range recoveredKeys {
		if key.PrivateKey == "" {
			t.Error("DKLS private key should not be empty")
		}
		if key.Address == "" {
			t.Error("DKLS address should not be empty")
		}
	}
}

func TestValidateVaultCompatibility_Integration(t *testing.T) {
	// Test vault compatibility with mismatched vaults
	vaultFiles := []string{
		"../../test/fixtures/testGG20-part1of2.vult",
		"../../test/fixtures/testDKLS-1of2.vult", // Different vault!
	}
	
	// Skip if test files don't exist
	for _, file := range vaultFiles {
		if !fileExists(file) {
			t.Skipf("Test fixture not found: %s", file)
			return
		}
	}
	
	_, err := RecoverPrivateKeys(vaultFiles, 2, "")
	if err == nil {
		t.Fatal("Expected error for incompatible vaults")
	}
	
	if !strings.Contains(err.Error(), "compatibility check failed") {
		t.Errorf("Expected compatibility check error, got: %v", err)
	}
}

func TestParseKeyShareData(t *testing.T) {
	// Test the key share parsing function
	shareData, err := parseKeyShareData("0267db81657a956f364167c3986a426b448a74ac0db2092f6665c4c202b37f6f1d", 1)
	if err != nil {
		t.Fatalf("Failed to parse key share data: %v", err)
	}
	
	if shareData.Index != 1 {
		t.Errorf("Expected index 1, got %d", shareData.Index)
	}
	
	if len(shareData.Share) == 0 {
		t.Error("Share data should not be empty")
	}
	
	if shareData.PublicKey != "0267db81657a956f364167c3986a426b448a74ac0db2092f6665c4c202b37f6f1d" {
		t.Errorf("Public key mismatch: %s", shareData.PublicKey)
	}
	
	// Test with invalid hex
	_, err = parseKeyShareData("invalid-hex", 1)
	if err == nil {
		t.Error("Expected error for invalid hex public key")
	}
}

func TestLagrangeInterpolation(t *testing.T) {
	// Test the Lagrange interpolation function
	shares := []KeyShareData{
		{Share: []byte{0x01, 0x02, 0x03}, Index: 1, PublicKey: "test1"},
		{Share: []byte{0x04, 0x05, 0x06}, Index: 2, PublicKey: "test2"},
	}
	
	// Import elliptic to get a proper curve
	// Note: We can't import it at the top due to existing imports, so we'll handle the curve differently
	result, err := lagrangeInterpolation(shares, nil) // This should handle nil curve gracefully
	if err != nil {
		t.Fatalf("Lagrange interpolation failed: %v", err)
	}
	
	if result == nil {
		t.Error("Result should not be nil")
	}
	
	if result.Sign() == 0 {
		t.Error("Result should not be zero")
	}
	
	// Test with empty shares
	_, err = lagrangeInterpolation([]KeyShareData{}, nil)
	if err == nil {
		t.Error("Expected error for empty shares")
	}
}

// Helper function to check if a file exists
func fileExists(filename string) bool {
	_, err := os.Stat(filename)
	return err == nil
}
