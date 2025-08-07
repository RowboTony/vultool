package vault

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Test file paths
const (
	testUnencryptedGG20   = "../../test/fixtures/testGG20-part1of2.vult"
	testUnencryptedDKLS   = "../../test/fixtures/testDKLS-1of2.vult"
	testEncryptedVault    = "../../test/fixtures/qa-fast-share2of2.vult"
	testEncryptedPassword = "vulticli01"
	testVault1            = "../../test/fixtures/testGG20-part1of2.vult"
	testVault2            = "../../test/fixtures/testGG20-part2of2.vult"
)

func TestParseVaultFile_UnencryptedGG20(t *testing.T) {
	vaultInfo, err := ParseVaultFile(testUnencryptedGG20)
	if err != nil {
		t.Fatalf("Failed to parse unencrypted GG20 vault: %v", err)
	}

	// Validate basic vault information
	if vaultInfo.Name == "" {
		t.Error("Vault name should not be empty")
	}
	if vaultInfo.IsEncrypted {
		t.Error("Vault should not be encrypted")
	}
	if vaultInfo.PublicKeyECDSA == "" {
		t.Error("ECDSA public key should not be empty")
	}
	if vaultInfo.PublicKeyEDDSA == "" {
		t.Error("EDDSA public key should not be empty")
	}
	if vaultInfo.HexChainCode == "" {
		t.Error("Hex chain code should not be empty")
	}
	if vaultInfo.LocalPartyKey == "" {
		t.Error("Local party key should not be empty")
	}
	if len(vaultInfo.KeyShares) == 0 {
		t.Error("Key shares should not be empty")
	}

	// Check file path (should be absolute path)
	if !strings.Contains(vaultInfo.FilePath, "testGG20-part1of2.vult") {
		t.Errorf("File path should contain test file name, got %s", vaultInfo.FilePath)
	}
}

func TestParseVaultFile_UnencryptedDKLS(t *testing.T) {
	vaultInfo, err := ParseVaultFile(testUnencryptedDKLS)
	if err != nil {
		t.Fatalf("Failed to parse unencrypted DKLS vault: %v", err)
	}

	// Validate basic vault information
	if vaultInfo.Name == "" {
		t.Error("Vault name should not be empty")
	}
	if vaultInfo.IsEncrypted {
		t.Error("Vault should not be encrypted")
	}
	if vaultInfo.PublicKeyECDSA == "" {
		t.Error("ECDSA public key should not be empty")
	}
	if vaultInfo.PublicKeyEDDSA == "" {
		t.Error("EDDSA public key should not be empty")
	}
	if vaultInfo.HexChainCode == "" {
		t.Error("Hex chain code should not be empty")
	}
	if vaultInfo.LocalPartyKey == "" {
		t.Error("Local party key should not be empty")
	}
	if len(vaultInfo.KeyShares) == 0 {
		t.Error("Key shares should not be empty")
	}

	// Check file path (should be absolute path)
	if !strings.Contains(vaultInfo.FilePath, "testDKLS-1of2.vult") {
		t.Errorf("File path should contain test file name, got %s", vaultInfo.FilePath)
	}
}

func TestParseVaultFileWithPassword_Encrypted(t *testing.T) {
	vaultInfo, err := ParseVaultFileWithPassword(testEncryptedVault, testEncryptedPassword)
	if err != nil {
		t.Fatalf("Failed to parse encrypted vault: %v", err)
	}

	// Validate basic vault information
	if vaultInfo.Name == "" {
		t.Error("Vault name should not be empty")
	}
	if !vaultInfo.IsEncrypted {
		t.Error("Vault should be encrypted")
	}
	if vaultInfo.PublicKeyECDSA == "" {
		t.Error("ECDSA public key should not be empty")
	}
	if vaultInfo.PublicKeyEDDSA == "" {
		t.Error("EDDSA public key should not be empty")
	}
	if vaultInfo.HexChainCode == "" {
		t.Error("Hex chain code should not be empty")
	}
	if vaultInfo.LocalPartyKey == "" {
		t.Error("Local party key should not be empty")
	}
	if len(vaultInfo.KeyShares) == 0 {
		t.Error("Key shares should not be empty")
	}

	// Check file path (should be absolute path)
	if !strings.Contains(vaultInfo.FilePath, "qa-fast-share2of2.vult") {
		t.Errorf("File path should contain test file name, got %s", vaultInfo.FilePath)
	}
}

func TestParseVaultFileWithPassword_WrongPassword(t *testing.T) {
	_, err := ParseVaultFileWithPassword(testEncryptedVault, "wrongpassword")
	if err == nil {
		t.Fatal("Expected error when using wrong password")
	}
	if !strings.Contains(err.Error(), "decrypt") {
		t.Errorf("Expected decryption error, got: %v", err)
	}
}

func TestParseVaultFile_NonExistentFile(t *testing.T) {
	_, err := ParseVaultFile("nonexistent.vult")
	if err == nil {
		t.Fatal("Expected error for non-existent file")
	}
	if !strings.Contains(err.Error(), "error accessing file") {
		t.Errorf("Expected file access error, got: %v", err)
	}
}

func TestParseVaultFile_InvalidPath(t *testing.T) {
	// Test directory traversal attempt - use a path that will actually fail validation
	_, err := ParseVaultFile("/etc/passwd")
	if err == nil {
		t.Fatal("Expected error for unsafe path")
	}
	// The error could be either unsafe path or file access error
	if !strings.Contains(err.Error(), "unsafe file path") && !strings.Contains(err.Error(), "access to system path") {
		t.Errorf("Expected unsafe path error, got: %v", err)
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
	// Create an invalid vault
	vaultInfo := &VaultInfo{
		Name:           "", // Empty name should cause validation error
		PublicKeyECDSA: "",
		PublicKeyEDDSA: "",
		HexChainCode:   "",
		LocalPartyKey:  "",
		KeyShares:      []KeyShareInfo{},
	}

	issues := ValidateVault(vaultInfo)
	expectedIssues := []string{
		"vault name is empty",
		"no public keys found",
		"no key shares found",
		"hex chain code is missing",
		"local party key is missing",
	}

	if len(issues) != len(expectedIssues) {
		t.Errorf("Expected %d validation issues, got %d: %v", len(expectedIssues), len(issues), issues)
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

func TestGetSummary(t *testing.T) {
	vaultInfo, err := ParseVaultFile(testUnencryptedGG20)
	if err != nil {
		t.Fatalf("Failed to parse vault: %v", err)
	}

	summary := GetSummary(vaultInfo)

	// Check that summary contains expected information
	expectedParts := []string{
		"Vault:",
		"File:",
		"Encrypted:",
		"Version:",
		"Local Party:",
		"ECDSA Public Key:",
		"EDDSA Public Key:",
		"Hex Chain Code:",
		"Key Shares:",
	}

	for _, part := range expectedParts {
		if !strings.Contains(summary, part) {
			t.Errorf("Summary missing expected part '%s'. Summary: %s", part, summary)
		}
	}

	// Check that encryption status is correct
	if !strings.Contains(summary, "Encrypted: false") {
		t.Errorf("Summary should show encryption status as false for unencrypted vault")
	}
}

func TestGetKeySharesInfo(t *testing.T) {
	vaultInfo, err := ParseVaultFile(testUnencryptedGG20)
	if err != nil {
		t.Fatalf("Failed to parse vault: %v", err)
	}

	keySharesInfo := GetKeySharesInfo(vaultInfo)

	// Check that key shares info contains expected information
	if !strings.Contains(keySharesInfo, "Key Share Information:") {
		t.Error("Key shares info should contain header")
	}

	if !strings.Contains(keySharesInfo, "Share 1:") {
		t.Error("Key shares info should contain first share")
	}

	if !strings.Contains(keySharesInfo, "ECDSA") || !strings.Contains(keySharesInfo, "EDDSA") {
		t.Error("Key shares info should contain key types")
	}
}

func TestDiffVaults_Identical(t *testing.T) {
	vault1, err := ParseVaultFile(testVault1)
	if err != nil {
		t.Fatalf("Failed to parse vault1: %v", err)
	}

	vault2, err := ParseVaultFile(testVault1) // Same file
	if err != nil {
		t.Fatalf("Failed to parse vault2: %v", err)
	}

	diff := DiffVaults(vault1, vault2)
	if !diff.Same {
		t.Errorf("Expected vaults to be identical, got diff: %+v", diff)
	}
	if len(diff.Details) > 0 {
		t.Errorf("Expected no diff details for identical vaults, got: %v", diff.Details)
	}
}

func TestDiffVaults_Different(t *testing.T) {
	vault1, err := ParseVaultFile(testVault1)
	if err != nil {
		t.Fatalf("Failed to parse vault1: %v", err)
	}

	vault2, err := ParseVaultFile(testVault2)
	if err != nil {
		t.Fatalf("Failed to parse vault2: %v", err)
	}

	diff := DiffVaults(vault1, vault2)
	if diff.Same {
		t.Errorf("Expected vaults to be different, got same: %+v", diff)
	}
	if len(diff.Details) == 0 {
		t.Errorf("Expected diff details for different vaults, got empty")
	}

	// Check that diff contains expected differences
	hasLocalPartyDiff := false
	for _, detail := range diff.Details {
		if strings.Contains(detail, "Local Party:") {
			hasLocalPartyDiff = true
			break
		}
	}
	if !hasLocalPartyDiff {
		t.Errorf("Expected local party difference in diff details: %v", diff.Details)
	}
}

func TestFormatDiff_Identical(t *testing.T) {
	vault1, err := ParseVaultFile(testVault1)
	if err != nil {
		t.Fatalf("Failed to parse vault: %v", err)
	}

	diff := DiffVaults(vault1, vault1)
	formatted := FormatDiff(diff, false)

	if !strings.Contains(formatted, "✓ Vaults are identical") {
		t.Errorf("Expected identical message in formatted diff: %s", formatted)
	}
}

func TestFormatDiff_Different(t *testing.T) {
	vault1, err := ParseVaultFile(testVault1)
	if err != nil {
		t.Fatalf("Failed to parse vault1: %v", err)
	}

	vault2, err := ParseVaultFile(testVault2)
	if err != nil {
		t.Fatalf("Failed to parse vault2: %v", err)
	}

	diff := DiffVaults(vault1, vault2)
	formatted := FormatDiff(diff, false)

	if !strings.Contains(formatted, "✗ Vaults differ:") {
		t.Errorf("Expected difference message in formatted diff: %s", formatted)
	}
}

func TestFormatDiff_WithColors(t *testing.T) {
	vault1, err := ParseVaultFile(testVault1)
	if err != nil {
		t.Fatalf("Failed to parse vault1: %v", err)
	}

	vault2, err := ParseVaultFile(testVault2)
	if err != nil {
		t.Fatalf("Failed to parse vault2: %v", err)
	}

	diff := DiffVaults(vault1, vault2)
	formatted := FormatDiff(diff, true)

	// Check for ANSI color codes
	if !strings.Contains(formatted, "\033[31m") { // Red for error
		t.Errorf("Expected color codes in formatted diff with colors enabled: %s", formatted)
	}
}

func TestValidateSafeOutputPath_ValidPath(t *testing.T) {
	// Test with a valid output path
	err := ValidateSafeOutputPath("output.json")
	if err != nil {
		t.Errorf("Expected no error for valid output path, got: %v", err)
	}

	// Test with a valid absolute path (using temp directory)
	tempDir := os.TempDir()
	validPath := filepath.Join(tempDir, "test-output.json")
	err = ValidateSafeOutputPath(validPath)
	if err != nil {
		t.Errorf("Expected no error for valid absolute path, got: %v", err)
	}
}

func TestValidateSafeOutputPath_DangerousPath(t *testing.T) {
	// Test with system directory (should fail)
	err := ValidateSafeOutputPath("/bin/malicious")
	if err == nil {
		t.Error("Expected error for dangerous output path")
	}
	if !strings.Contains(err.Error(), "system directory") {
		t.Errorf("Expected system directory error, got: %v", err)
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

func TestTruncateKey(t *testing.T) {
	// Test short key (should not be truncated)
	shortKey := "1234567890"
	result := truncateKey(shortKey)
	if result != shortKey {
		t.Errorf("Expected short key to remain unchanged, got: %s", result)
	}

	// Test long key (should be truncated)
	longKey := "1234567890abcdef1234567890abcdef"
	result = truncateKey(longKey)
	expected := "1234567890abcdef..."
	if result != expected {
		t.Errorf("Expected truncated key %s, got: %s", expected, result)
	}
}

func TestParseVaultContentDirect_UnencryptedVault(t *testing.T) {
	// Read a test vault file
	content, err := os.ReadFile(testUnencryptedGG20)
	if err != nil {
		t.Fatalf("Failed to read test file: %v", err)
	}

	vaultInfo, err := ParseVaultContentDirect(string(content), "test.vult")
	if err != nil {
		t.Fatalf("Failed to parse vault content directly: %v", err)
	}

	// Validate basic information
	if vaultInfo.Name == "" {
		t.Error("Vault name should not be empty")
	}
	if vaultInfo.FilePath != "test.vult" {
		t.Errorf("Expected file path 'test.vult', got: %s", vaultInfo.FilePath)
	}
}

func TestParseVaultContentDirect_EncryptedVault(t *testing.T) {
	// Read an encrypted test vault file
	content, err := os.ReadFile(testEncryptedVault)
	if err != nil {
		t.Fatalf("Failed to read test file: %v", err)
	}

	// Should fail for encrypted vaults
	_, err = ParseVaultContentDirect(string(content), "test.vult")
	if err == nil {
		t.Fatal("Expected error for encrypted vault in direct parsing mode")
	}
	if !strings.Contains(err.Error(), "encrypted vaults not supported") {
		t.Errorf("Expected encrypted vault error, got: %v", err)
	}
}

func TestParseVaultFromBytes(t *testing.T) {
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

func BenchmarkDiffVaults(b *testing.B) {
	vault1, err := ParseVaultFile(testVault1)
	if err != nil {
		b.Fatalf("Failed to parse vault1: %v", err)
	}

	vault2, err := ParseVaultFile(testVault2)
	if err != nil {
		b.Fatalf("Failed to parse vault2: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		DiffVaults(vault1, vault2)
	}
}
