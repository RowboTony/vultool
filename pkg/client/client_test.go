package client

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Test file paths relative to the pkg/client directory
const (
	testUnencryptedGG20   = "../../test/fixtures/testGG20-part1of2.vult"
	testUnencryptedDKLS   = "../../test/fixtures/testDKLS-1of2.vult"
	testEncryptedVault    = "../../test/fixtures/qa-fast-share2of2.vult"
	testEncryptedPassword = "vulticli01"
)

func TestParseVaultFile_Success(t *testing.T) {
	vaultInfo, err := ParseVaultFile(testUnencryptedGG20)
	if err != nil {
		t.Fatalf("Failed to parse vault file: %v", err)
	}

	// Validate basic information
	if vaultInfo.Name == "" {
		t.Error("Vault name should not be empty")
	}
	if vaultInfo.PublicKeyECDSA == "" {
		t.Error("ECDSA public key should not be empty")
	}
	if vaultInfo.LocalPartyKey == "" {
		t.Error("Local party key should not be empty")
	}
	if len(vaultInfo.KeyShares) == 0 {
		t.Error("Key shares should not be empty")
	}
}

func TestParseVaultFile_NonExistent(t *testing.T) {
	_, err := ParseVaultFile("nonexistent.vult")
	if err == nil {
		t.Fatal("Expected error for non-existent file")
	}
	if !strings.Contains(err.Error(), "error accessing file") {
		t.Errorf("Expected file access error, got: %v", err)
	}
}

func TestParseVaultFileWithPassword_Encrypted(t *testing.T) {
	vaultInfo, err := ParseVaultFileWithPassword(testEncryptedVault, testEncryptedPassword)
	if err != nil {
		t.Fatalf("Failed to parse encrypted vault: %v", err)
	}

	// Validate basic information
	if vaultInfo.Name == "" {
		t.Error("Vault name should not be empty")
	}
	if !vaultInfo.IsEncrypted {
		t.Error("Vault should be marked as encrypted")
	}
	if vaultInfo.PublicKeyECDSA == "" {
		t.Error("ECDSA public key should not be empty")
	}
	if len(vaultInfo.KeyShares) == 0 {
		t.Error("Key shares should not be empty")
	}
}

func TestParseVaultFileWithPassword_WrongPassword(t *testing.T) {
	_, err := ParseVaultFileWithPassword(testEncryptedVault, "wrongpassword")
	if err == nil {
		t.Fatal("Expected error with wrong password")
	}
	if !strings.Contains(err.Error(), "decrypt") {
		t.Errorf("Expected decryption error, got: %v", err)
	}
}

func TestValidateVault_ValidVault(t *testing.T) {
	vaultInfo, err := ParseVaultFile(testUnencryptedGG20)
	if err != nil {
		t.Fatalf("Failed to parse vault: %v", err)
	}

	issues := ValidateVault(vaultInfo)
	if len(issues) > 0 {
		t.Errorf("Expected no validation issues for valid vault, got: %v", issues)
	}
}

func TestValidateVault_InvalidVault(t *testing.T) {
	// Create an invalid vault info
	invalidVault := &VaultInfo{
		Name:           "", // Empty name should cause validation error
		PublicKeyECDSA: "",
		PublicKeyEDDSA: "",
		HexChainCode:   "",
		LocalPartyKey:  "",
		KeyShares:      []KeyShareInfo{},
	}

	issues := ValidateVault(invalidVault)
	expectedIssueCount := 5 // name, public keys, key shares, chain code, local party key
	if len(issues) != expectedIssueCount {
		t.Errorf("Expected %d validation issues, got %d: %v", expectedIssueCount, len(issues), issues)
	}

	// Check for specific expected issues
	expectedIssues := []string{
		"vault name is empty",
		"no public keys found",
		"no key shares found",
		"hex chain code is missing",
		"local party key is missing",
	}

	for _, expectedIssue := range expectedIssues {
		found := false
		for _, issue := range issues {
			if issue == expectedIssue {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected validation issue '%s' not found in: %v", expectedIssue, issues)
		}
	}
}

func TestIsValidVultFile_ValidContent(t *testing.T) {
	// Read a valid vault file
	content, err := os.ReadFile(testUnencryptedGG20)
	if err != nil {
		t.Fatalf("Failed to read test file: %v", err)
	}

	valid, err := IsValidVultFile(string(content))
	if err != nil {
		t.Errorf("Expected no error for valid vault file, got: %v", err)
	}
	if !valid {
		t.Error("Expected valid vault file to be recognized as valid")
	}
}

func TestIsValidVultFile_InvalidContent(t *testing.T) {
	// Test with invalid base64
	valid, err := IsValidVultFile("invalid-base64-content!")
	if err == nil {
		t.Error("Expected error for invalid base64 content")
	}
	if valid {
		t.Error("Expected invalid content to be recognized as invalid")
	}

	// Test with empty content
	valid, err = IsValidVultFile("")
	if err == nil {
		t.Error("Expected error for empty content")
	}
	if valid {
		t.Error("Expected empty content to be recognized as invalid")
	}
}

func TestValidateVultFileFromPath_ValidFile(t *testing.T) {
	// Get absolute path to avoid relative path validation issues
	absPath, err := filepath.Abs(testUnencryptedGG20)
	if err != nil {
		t.Fatalf("Failed to get absolute path: %v", err)
	}
	
	valid, err := ValidateVultFileFromPath(absPath)
	if err != nil {
		t.Errorf("Expected no error for valid vault file path, got: %v", err)
	}
	if !valid {
		t.Error("Expected valid vault file path to be recognized as valid")
	}
}

func TestValidateVultFileFromPath_InvalidPath(t *testing.T) {
	// Test with non-existent file
	valid, err := ValidateVultFileFromPath("nonexistent.vult")
	if err == nil {
		t.Error("Expected error for non-existent file")
	}
	if valid {
		t.Error("Expected non-existent file to be recognized as invalid")
	}

	// Test with dangerous path
	valid, err = ValidateVultFileFromPath("../../../etc/passwd")
	if err == nil {
		t.Error("Expected error for dangerous path")
	}
	if valid {
		t.Error("Expected dangerous path to be recognized as invalid")
	}
}

func TestParseVaultFromBytes_Success(t *testing.T) {
	// Read a test vault file
	content, err := os.ReadFile(testUnencryptedGG20)
	if err != nil {
		t.Fatalf("Failed to read test file: %v", err)
	}

	vaultInfo, err := ParseVaultFromBytes(content)
	if err != nil {
		t.Fatalf("Failed to parse vault from bytes: %v", err)
	}

	// Validate basic information
	if vaultInfo.Name == "" {
		t.Error("Vault name should not be empty")
	}
	if vaultInfo.FilePath != "uploaded-file.vult" {
		t.Errorf("Expected file path 'uploaded-file.vult', got: %s", vaultInfo.FilePath)
	}
	if vaultInfo.PublicKeyECDSA == "" {
		t.Error("ECDSA public key should not be empty")
	}
	if len(vaultInfo.KeyShares) == 0 {
		t.Error("Key shares should not be empty")
	}
}

func TestParseVaultFromBytes_InvalidData(t *testing.T) {
	// Test with invalid data
	_, err := ParseVaultFromBytes([]byte("invalid-vault-data"))
	if err == nil {
		t.Fatal("Expected error for invalid vault data")
	}
	if !strings.Contains(err.Error(), "base64") {
		t.Errorf("Expected base64 error, got: %v", err)
	}
}

// Test integration between different client functions
func TestClientIntegration(t *testing.T) {
	// Step 1: Parse a vault file
	vaultInfo, err := ParseVaultFile(testUnencryptedGG20)
	if err != nil {
		t.Fatalf("Failed to parse vault file: %v", err)
	}

	// Step 2: Validate it
	issues := ValidateVault(vaultInfo)
	if len(issues) > 0 {
		t.Errorf("Vault validation failed: %v", issues)
	}

	// Step 3: Validate the file path (use absolute path)
	absPath, err := filepath.Abs(testUnencryptedGG20)
	if err != nil {
		t.Fatalf("Failed to get absolute path: %v", err)
	}
	valid, err := ValidateVultFileFromPath(absPath)
	if err != nil {
		t.Errorf("File path validation failed: %v", err)
	}
	if !valid {
		t.Error("File path should be valid")
	}

	// Step 4: Read the file content and validate it
	content, err := os.ReadFile(testUnencryptedGG20)
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}

	valid, err = IsValidVultFile(string(content))
	if err != nil {
		t.Errorf("Content validation failed: %v", err)
	}
	if !valid {
		t.Error("Content should be valid")
	}

	// Step 5: Parse from bytes and compare
	vaultInfoFromBytes, err := ParseVaultFromBytes(content)
	if err != nil {
		t.Fatalf("Failed to parse vault from bytes: %v", err)
	}

	// Compare key fields (ignoring file path difference)
	if vaultInfo.Name != vaultInfoFromBytes.Name {
		t.Errorf("Vault name mismatch: file=%s, bytes=%s", vaultInfo.Name, vaultInfoFromBytes.Name)
	}
	if vaultInfo.PublicKeyECDSA != vaultInfoFromBytes.PublicKeyECDSA {
		t.Errorf("ECDSA key mismatch: file=%s, bytes=%s", vaultInfo.PublicKeyECDSA, vaultInfoFromBytes.PublicKeyECDSA)
	}
	if len(vaultInfo.KeyShares) != len(vaultInfoFromBytes.KeyShares) {
		t.Errorf("Key shares count mismatch: file=%d, bytes=%d", len(vaultInfo.KeyShares), len(vaultInfoFromBytes.KeyShares))
	}
}

// Test with both GG20 and DKLS vault types
func TestDifferentVaultTypes(t *testing.T) {
	vaultTypes := []struct {
		name string
		path string
	}{
		{"GG20", testUnencryptedGG20},
		{"DKLS", testUnencryptedDKLS},
	}

	for _, vt := range vaultTypes {
		t.Run(vt.name, func(t *testing.T) {
			// Parse vault
			vaultInfo, err := ParseVaultFile(vt.path)
			if err != nil {
				t.Fatalf("Failed to parse %s vault: %v", vt.name, err)
			}

			// Validate vault
			issues := ValidateVault(vaultInfo)
			if len(issues) > 0 {
				t.Errorf("%s vault validation failed: %v", vt.name, issues)
			}

			// Check essential fields are present
			if vaultInfo.Name == "" {
				t.Errorf("%s vault name is empty", vt.name)
			}
			if vaultInfo.PublicKeyECDSA == "" {
				t.Errorf("%s vault ECDSA key is empty", vt.name)
			}
			if vaultInfo.PublicKeyEDDSA == "" {
				t.Errorf("%s vault EDDSA key is empty", vt.name)
			}
			if vaultInfo.HexChainCode == "" {
				t.Errorf("%s vault chain code is empty", vt.name)
			}
			if len(vaultInfo.KeyShares) == 0 {
				t.Errorf("%s vault has no key shares", vt.name)
			}
		})
	}
}

// Benchmark tests for performance validation
func BenchmarkParseVaultFile(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, err := ParseVaultFile(testUnencryptedGG20)
		if err != nil {
			b.Fatalf("Failed to parse vault: %v", err)
		}
	}
}

func BenchmarkValidateVault(b *testing.B) {
	vaultInfo, err := ParseVaultFile(testUnencryptedGG20)
	if err != nil {
		b.Fatalf("Failed to parse vault: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ValidateVault(vaultInfo)
	}
}

func BenchmarkIsValidVultFile(b *testing.B) {
	content, err := os.ReadFile(testUnencryptedGG20)
	if err != nil {
		b.Fatalf("Failed to read test file: %v", err)
	}
	contentStr := string(content)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := IsValidVultFile(contentStr)
		if err != nil {
			b.Fatalf("Validation failed: %v", err)
		}
	}
}

func BenchmarkParseVaultFromBytes(b *testing.B) {
	content, err := os.ReadFile(testUnencryptedGG20)
	if err != nil {
		b.Fatalf("Failed to read test file: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := ParseVaultFromBytes(content)
		if err != nil {
			b.Fatalf("Failed to parse vault from bytes: %v", err)
		}
	}
}
