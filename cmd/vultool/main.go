package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/rowbotony/vultool/internal/recovery"
	"github.com/rowbotony/vultool/internal/types"
	"github.com/rowbotony/vultool/internal/util"
	"github.com/rowbotony/vultool/internal/vault"
)

// Version is set at build time from VERSION file
// Build with: go build -ldflags "-X main.version=$(cat VERSION)"
var version = "dev"

// showFirstRunMessage displays a welcome message for first-time users
func showFirstRunMessage() {
	// Get user config directory
	configDir, err := os.UserConfigDir()
	if err != nil {
		// If we can't get config dir, skip the message rather than error
		return
	}

	// Create vultool config directory if it doesn't exist
	vultoolDir := filepath.Join(configDir, "vultool")
	// #nosec G301 - Standard config directory permissions
	if err := os.MkdirAll(vultoolDir, 0o750); err != nil {
		return
	}

	// Check if first-run marker exists
	firstRunFile := filepath.Join(vultoolDir, ".installed")
	if _, err := os.Stat(firstRunFile); err == nil {
		// File exists, not first run
		return
	}

	// Show welcome message
	fmt.Println("\n🎉 vultool installed successfully!")
	fmt.Printf("Version: %s\n", version)
	fmt.Println("\nNext steps:")
	fmt.Println("  vultool --help           # Show all available commands")
	fmt.Println("  vultool info -f file.vult    # Quick vault information")
	fmt.Println("  vultool inspect -f file.vult # Detailed vault inspection")
	fmt.Println("\nFor more examples, visit: https://github.com/rowbotony/vultool")
	fmt.Println()

	// Create the marker file to prevent showing this message again
	// #nosec G304 - firstRunFile is safely constructed from UserConfigDir
	if file, err := os.Create(firstRunFile); err == nil {
		if closeErr := file.Close(); closeErr != nil {
			// Log but don't error on close failure for marker file
			fmt.Printf("Warning: failed to close marker file: %v\n", closeErr)
		}
	}
}

func main() {
	// Show welcome message for first-time users
	showFirstRunMessage()

	rootCmd := &cobra.Command{
		Use:     "vultool",
		Version: version,
		Short:   "Vultool - Standalone CLI for .vult file operations",
		Long:    `A standalone CLI tool for managing vault operations, compatible with Vultisig security models.`,
		Run: func(cmd *cobra.Command, args []string) {
			// Show help when no command is provided
			if err := cmd.Help(); err != nil {
				fmt.Printf("Error showing help: %v\n", err)
			}
		},
	}

	var (
		vaultFile     string
		exportFile    string
		validate      bool
		summary       bool
		showKeyshares bool
		password      string
	)

	inspectCmd := &cobra.Command{
		Use:   "inspect",
		Short: "Inspect and validate a vault file",
		Long:  `Inspect a .vult vault file, showing key shares, metadata, and more details, with validation options.`,
		Run: func(cmd *cobra.Command, args []string) {
			if vaultFile == "" {
				fmt.Println("Vault file is required.")
				return
			}
			absPath, err := filepath.Abs(vaultFile)
			if err != nil {
				fmt.Printf("Error getting absolute path: %v\n", err)
				return
			}

			vaultInfo, err := vault.ParseVaultFileWithPassword(absPath, password)
			if err != nil {
				fmt.Printf("Error parsing vault file: %v\n", err)
				return
			}

			if summary {
				fmt.Println(vault.GetSummary(vaultInfo))
				return
			}

			if showKeyshares {
				fmt.Println(vault.GetKeySharesInfo(vaultInfo))
				return
			}

			if validate {
				issues := vault.ValidateVault(vaultInfo)
				if len(issues) > 0 {
					fmt.Printf("Validation issues found:\n")
					for _, issue := range issues {
						fmt.Printf("  - %s\n", issue)
					}
					return
				} else {
					fmt.Println("✓ Vault validation passed - no issues found")
					return
				}
			}

			if exportFile != "" {
				// Validate output path for security
				if err := vault.ValidateSafeOutputPath(exportFile); err != nil {
					fmt.Printf("Unsafe export path: %v\n", err)
					return
				}

				// #nosec G304 - exportFile is validated by ValidateSafeOutputPath above
				file, err := os.Create(exportFile)
				if err != nil {
					fmt.Printf("Error creating export file: %v\n", err)
					return
				}
				defer func() {
					if closeErr := file.Close(); closeErr != nil {
						fmt.Printf("Warning: failed to close export file: %v\n", closeErr)
					}
				}()

				if err := util.OutputResult(vaultInfo, "json", file); err != nil {
					fmt.Printf("Error exporting to JSON: %v\n", err)
					return
				}
				fmt.Printf("Vault exported to: %s\n", exportFile)
				return
			}

			// Default: show summary if no specific flag is provided
			fmt.Println(vault.GetSummary(vaultInfo))
		},
	}
	inspectCmd.Flags().StringVarP(&vaultFile, "vault", "f", "", "Path to the .vult vault file (required)")
	inspectCmd.Flags().StringVar(&exportFile, "export", "", "Export vault metadata to JSON file")
	inspectCmd.Flags().BoolVar(&validate, "validate", false, "Run strict validation checks")
	inspectCmd.Flags().BoolVar(&summary, "summary", false, "Print high-level vault metadata")
	inspectCmd.Flags().BoolVar(&showKeyshares, "show-keyshares", false, "Output key share information")
	inspectCmd.Flags().StringVar(&password, "password", "", "Password for encrypted vault files (alternative to interactive prompt)")

	// Mark vault file as required
	if err := inspectCmd.MarkFlagRequired("vault"); err != nil {
		fmt.Printf("Error setting up CLI flags: %v\n", err)
		os.Exit(1)
	}

	// Add command aliases as specified in spec.md
	// info: alias to inspect --summary
	infoCmd := &cobra.Command{
		Use:   "info",
		Short: "Show concise vault information (alias for inspect --summary)",
		Long:  `Show a concise summary of vault information including protocol, key presence, threshold, and signer count.`,
		Run: func(cmd *cobra.Command, args []string) {
			if vaultFile == "" {
				fmt.Println("Vault file is required.")
				return
			}
			absPath, err := filepath.Abs(vaultFile)
			if err != nil {
				fmt.Printf("Error getting absolute path: %v\n", err)
				return
			}

			vaultInfo, err := vault.ParseVaultFileWithPassword(absPath, password)
			if err != nil {
				fmt.Printf("Error parsing vault file: %v\n", err)
				return
			}

			// Always show summary for info command
			fmt.Println(vault.GetSummary(vaultInfo))
		},
	}
	infoCmd.Flags().StringVarP(&vaultFile, "vault", "f", "", "Path to the .vult vault file (required)")
	infoCmd.Flags().StringVar(&password, "password", "", "Password for encrypted vault files (alternative to interactive prompt)")
	if err := infoCmd.MarkFlagRequired("vault"); err != nil {
		fmt.Printf("Error setting up info CLI flags: %v\n", err)
		os.Exit(1)
	}

	// decode: alias to inspect --json with YAML support
	decodeCmd := &cobra.Command{
		Use:   "decode",
		Short: "Decode vault to JSON or YAML format",
		Long:  `Decode and output the full vault protobuf data as JSON (default) or YAML for programmatic use.`,
		Run: func(cmd *cobra.Command, args []string) {
			if vaultFile == "" {
				fmt.Println("Vault file is required.")
				return
			}
			absPath, err := filepath.Abs(vaultFile)
			if err != nil {
				fmt.Printf("Error getting absolute path: %v\n", err)
				return
			}

			vaultInfo, err := vault.ParseVaultFileWithPassword(absPath, password)
			if err != nil {
				fmt.Printf("Error parsing vault file: %v\n", err)
				return
			}

			// Check output format flags
			useYAML, err := cmd.Flags().GetBool("yaml")
			if err != nil {
				fmt.Printf("Error reading yaml flag: %v\n", err)
				return
			}

			useTOML, err := cmd.Flags().GetBool("toml")
			if err != nil {
				fmt.Printf("Error reading toml flag: %v\n", err)
				return
			}

			// Determine output format - default to JSON
			format := "json"
			if useYAML {
				format = "yaml"
			} else if useTOML {
				format = "toml"
			}

			if err := util.OutputResult(vaultInfo, format, os.Stdout); err != nil {
				fmt.Printf("Error outputting %s: %v\n", format, err)
				return
			}
		},
	}
	decodeCmd.Flags().StringVarP(&vaultFile, "vault", "f", "", "Path to the .vult vault file (required)")
	decodeCmd.Flags().StringVar(&password, "password", "", "Password for encrypted vault files (alternative to interactive prompt)")
	decodeCmd.Flags().Bool("yaml", false, "Output in YAML format instead of JSON")
	decodeCmd.Flags().Bool("toml", false, "Output in TOML format (not yet implemented)")
	if err := decodeCmd.MarkFlagRequired("vault"); err != nil {
		fmt.Printf("Error setting up decode CLI flags: %v\n", err)
		os.Exit(1)
	}

	// verify: alias to inspect --validate
	verifyCmd := &cobra.Command{
		Use:   "verify",
		Short: "Verify vault integrity (alias for inspect --validate)",
		Long:  `Perform structural and cryptographic sanity checks on the vault file. Exits with code 0 if valid, 1 if invalid.`,
		Run: func(cmd *cobra.Command, args []string) {
			if vaultFile == "" {
				fmt.Println("Vault file is required.")
				return
			}
			absPath, err := filepath.Abs(vaultFile)
			if err != nil {
				fmt.Printf("Error getting absolute path: %v\n", err)
				os.Exit(1)
				return
			}

			vaultInfo, err := vault.ParseVaultFileWithPassword(absPath, password)
			if err != nil {
				fmt.Printf("Error parsing vault file: %v\n", err)
				os.Exit(1)
				return
			}

			// Run validation and exit with appropriate code
			issues := vault.ValidateVault(vaultInfo)
			if len(issues) > 0 {
				fmt.Printf("Validation issues found:\n")
				for _, issue := range issues {
					fmt.Printf("  - %s\n", issue)
				}
				os.Exit(1)
			} else {
				fmt.Println("✓ Vault validation passed - no issues found")
				os.Exit(0)
			}
		},
	}
	verifyCmd.Flags().StringVarP(&vaultFile, "vault", "f", "", "Path to the .vult vault file (required)")
	verifyCmd.Flags().StringVar(&password, "password", "", "Password for encrypted vault files (alternative to interactive prompt)")
	if err := verifyCmd.MarkFlagRequired("vault"); err != nil {
		fmt.Printf("Error setting up verify CLI flags: %v\n", err)
		os.Exit(1)
	}

	// diff: compare two vault files
	diffCmd := &cobra.Command{
		Use:   "diff",
		Short: "Compare two vault files",
		Long:  `Compare two .vult vault files and show differences in metadata and key shares.`,
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) != 2 {
				fmt.Println("Two vault files are required.")
				return
			}

			vaultFile1, vaultFile2 := args[0], args[1]

			absPath1, err := filepath.Abs(vaultFile1)
			if err != nil {
				fmt.Printf("Error getting absolute path: %v\n", err)
				return
			}

			absPath2, err := filepath.Abs(vaultFile2)
			if err != nil {
				fmt.Printf("Error getting absolute path: %v\n", err)
				return
			}

			vaultInfo1, err := vault.ParseVaultFileWithPassword(absPath1, password)
			if err != nil {
				fmt.Printf("Error parsing first vault file: %v\n", err)
				return
			}

			vaultInfo2, err := vault.ParseVaultFileWithPassword(absPath2, password)
			if err != nil {
				fmt.Printf("Error parsing second vault file: %v\n", err)
				return
			}

			diff := vault.DiffVaults(vaultInfo1, vaultInfo2)

			// Check if structured output was requested
			useJSON, _ := cmd.Flags().GetBool("json")
			useYAML, _ := cmd.Flags().GetBool("yaml")

			if useJSON {
				if err := util.OutputResult(diff, "json", os.Stdout); err != nil {
					fmt.Printf("Error outputting JSON: %v\n", err)
					return
				}
			} else if useYAML {
				if err := util.OutputResult(diff, "yaml", os.Stdout); err != nil {
					fmt.Printf("Error outputting YAML: %v\n", err)
					return
				}
			} else {
				// Default human-readable output
				fmt.Println(vault.FormatDiff(diff, true))
			}
		},
	}
	diffCmd.Flags().StringVar(&password, "password", "", "Password for encrypted vault files (alternative to interactive prompt)")
	diffCmd.Flags().Bool("json", false, "Output diff in JSON format")
	diffCmd.Flags().Bool("yaml", false, "Output diff in YAML format")

	// list-addresses: Derive and show all chain addresses from vault public keys
	listAddressesCmd := &cobra.Command{
		Use:   "list-addresses",
		Short: "List all blockchain addresses derived from vault public keys",
		Long: `Derive and display addresses for all supported blockchains from the vault's public keys.
This command uses proper cryptographic derivation to generate addresses for Bitcoin, Ethereum,
and all other supported chains directly from the vault's ECDSA and EdDSA public keys.`,
		Run: func(cmd *cobra.Command, args []string) {
			if vaultFile == "" {
				fmt.Println("Vault file is required.")
				return
			}
			absPath, err := filepath.Abs(vaultFile)
			if err != nil {
				fmt.Printf("Error getting absolute path: %v\n", err)
				return
			}

			vaultInfo, err := vault.ParseVaultFileWithPassword(absPath, password)
			if err != nil {
				fmt.Printf("Error parsing vault file: %v\n", err)
				return
			}

			// Derive addresses from vault public keys
			// This now works for ANY vault, not just hardcoded ones!
			addresses := vault.DeriveAddressesFromVault(vaultInfo)
			if len(addresses) == 0 {
				fmt.Println("No addresses could be derived from vault public keys")
				return
			}

			// Filter by chains if specified
			chainFilter, _ := cmd.Flags().GetStringSlice("chains")
			if len(chainFilter) > 0 {
				chainMap := make(map[string]bool)
				for _, chain := range chainFilter {
					chainMap[chain] = true
				}

				var filtered []vault.VaultAddress
				for _, addr := range addresses {
					if chainMap[addr.Chain] {
						filtered = append(filtered, addr)
					}
				}
				addresses = filtered
			}

			useJSON, _ := cmd.Flags().GetBool("json")
			useCSV, _ := cmd.Flags().GetBool("csv")

			if useJSON {
				if err := util.OutputResult(addresses, "json", os.Stdout); err != nil {
					fmt.Printf("Error outputting JSON: %v\n", err)
				}
			} else if useCSV {
				// Output CSV header
				fmt.Println("Chain,Ticker,Address,DerivePath")
				// Output each address as a CSV row
				for _, addr := range addresses {
					fmt.Printf("%s,%s,%s,%s\n",
						addr.Chain,
						addr.Ticker,
						addr.Address,
						addr.DerivePath)
				}
			} else {
				fmt.Printf("Vault: %s\n", vaultInfo.Name)
				fmt.Printf("Key Shares: %d\n", len(vaultInfo.KeyShares))
				if vaultInfo.IsEncrypted {
					fmt.Printf("Encrypted: Yes\n")
				}
				fmt.Println()

				fmt.Println("Addresses:")
				fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

				for _, addr := range addresses {
					fmt.Printf("%-15s %-6s %s\n", addr.Chain, addr.Ticker, addr.Address)
					if addr.DerivePath != "" {
						fmt.Printf("                      Path: %s\n", addr.DerivePath)
					}
					fmt.Println("────────────────────────────────────────────────────────────")
				}
			}
		},
	}
	listAddressesCmd.Flags().StringVarP(&vaultFile, "vault", "f", "", "Path to the .vult vault file (required)")
	listAddressesCmd.Flags().StringVar(&password, "password", "", "Password for encrypted vault files")
	listAddressesCmd.Flags().Bool("json", false, "Output in JSON format")
	listAddressesCmd.Flags().Bool("csv", false, "Output in CSV format")
	listAddressesCmd.Flags().StringSlice("chains", []string{}, "Filter by chain names (e.g., Bitcoin,Ethereum)")
	if err := listAddressesCmd.MarkFlagRequired("vault"); err != nil {
		fmt.Printf("Error setting up list-addresses CLI flags: %v\n", err)
		os.Exit(1)
	}

	// ===== MEDIC MILESTONE (v0.2) COMMANDS =====
	// These are STUB implementations for the v0.2 milestone

	// recover: combine threshold shares to reconstruct private keys
	recoverCmd := &cobra.Command{
		Use:   "recover",
		Short: "Combine ≥t shares to reconstruct private keys for WIF/hex export",
		Long: `Combine threshold shares from multiple .vult files to reconstruct the original private key.
Exports keys in various formats: WIF (Bitcoin), hex (Ethereum), base58 (Solana/THOR).

⚠️  WARNING: This command reconstructs the actual private key material.
Only use this for legitimate recovery purposes in a secure environment.`,
		Example: `  # Recover keys from 2-of-3 threshold shares
  vultool recover share1.vult share2.vult --threshold 2
  
  # Recover with password and export to file
  vultool recover *.vult --threshold 3 --password mypass --output keys.json
  
  # Recover specific chain only
  vultool recover share*.vult --threshold 2 --chain bitcoin`,
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) == 0 {
				fmt.Println("At least one vault file is required.")
				return
			}

			threshold, _ := cmd.Flags().GetInt("threshold")
			outputFile, _ := cmd.Flags().GetString("output")
			chainFilter, _ := cmd.Flags().GetString("chain")
			useJSON, _ := cmd.Flags().GetBool("json")

			// Validate threshold
			if threshold <= 0 || threshold > len(args) {
				fmt.Printf("Invalid threshold: must be between 1 and %d (number of provided files)\n", len(args))
				return
			}

			// Convert file arguments to absolute paths
			vaultFiles := make([]string, len(args))
			for i, file := range args {
				absPath, err := filepath.Abs(file)
				if err != nil {
					fmt.Printf("Error getting absolute path for %s: %v\n", file, err)
					return
				}
				vaultFiles[i] = absPath
			}

			fmt.Printf("🔄 Attempting to recover private keys from %d shares (threshold: %d)...\n", len(vaultFiles), threshold)
			if chainFilter != "" {
				fmt.Printf("   Filtering for chain: %s\n", chainFilter)
			}
			fmt.Println()

			// Call the recovery function (currently stubbed)
			recoveredKeys, err := recovery.RecoverPrivateKeys(vaultFiles, threshold, password)
			if err != nil {
				fmt.Printf("❌ Recovery failed: %v\n", err)
				return
			}

			// Filter by chain if specified
			if chainFilter != "" {
				var filtered []recovery.RecoveredKey
				for _, key := range recoveredKeys {
					if string(key.Chain) == chainFilter {
						filtered = append(filtered, key)
					}
				}
				recoveredKeys = filtered
			}

			// Output results
			if outputFile != "" {
				if err := vault.ValidateSafeOutputPath(outputFile); err != nil {
					fmt.Printf("Unsafe output path: %v\n", err)
					return
				}

				file, err := os.Create(outputFile)
				if err != nil {
					fmt.Printf("Error creating output file: %v\n", err)
					return
				}
				defer file.Close()

				if err := util.OutputResult(recoveredKeys, "json", file); err != nil {
					fmt.Printf("Error writing to output file: %v\n", err)
					return
				}
				fmt.Printf("✅ Recovery results written to: %s\n", outputFile)
			} else if useJSON {
				if err := util.OutputResult(recoveredKeys, "json", os.Stdout); err != nil {
					fmt.Printf("Error outputting JSON: %v\n", err)
				}
			} else {
				// Human-readable output
				fmt.Printf("✅ Successfully recovered %d keys:\n\n", len(recoveredKeys))
				for i, key := range recoveredKeys {
					fmt.Printf("Key %d (%s):\n", i+1, key.Chain)
					fmt.Printf("  Address:     %s\n", key.Address)
					fmt.Printf("  Private Key: %s\n", key.PrivateKey)

					// Display wallet-compatible formats for EdDSA chains
					if key.SolanaSeedFormat != "" {
						fmt.Printf("  Solana Seed Only (32-byte base64): %s\n", key.SolanaSeedFormat)
						fmt.Printf("  ⚠️  Note: Most wallets need standard Ed25519 keypair, not TSS format\n")
					}
					if key.SolanaWalletFormat != "" {
						fmt.Printf("  Solana TSS Format (64-byte base64): %s\n", key.SolanaWalletFormat)
					}
					if key.SolanaWalletJSON != "" {
						// Show complete JSON array for wallet import
						fmt.Printf("  Solana TSS Format (JSON array): %s\n", key.SolanaWalletJSON)
					}
					if key.SuiWalletFormat != "" {
						fmt.Printf("  Sui Wallet Format (base64): %s\n", key.SuiWalletFormat)
					}

					if key.WIF != "" {
						fmt.Printf("  WIF:         %s\n", key.WIF)
					}
					if key.Base58 != "" {
						fmt.Printf("  Base58:      %s\n", key.Base58)
					}
					if key.DerivePath != "" {
						fmt.Printf("  Derive Path: %s\n", key.DerivePath)
					}
					fmt.Println()
				}
			}
		},
	}
	recoverCmd.Flags().Int("threshold", 0, "Minimum number of shares required for recovery (required)")
	recoverCmd.Flags().StringVar(&password, "password", "", "Password for encrypted vault files")
	recoverCmd.Flags().String("output", "", "Output file for recovery results (JSON format)")
	recoverCmd.Flags().String("chain", "", "Filter results for specific blockchain (bitcoin, ethereum, solana, thorchain)")
	recoverCmd.Flags().Bool("json", false, "Output in JSON format")
	if err := recoverCmd.MarkFlagRequired("threshold"); err != nil {
		fmt.Printf("Error setting up recover CLI flags: %v\n", err)
		os.Exit(1)
	}

	// derive: read-only HD key derivation
	deriveCmd := &cobra.Command{
		Use:   "derive",
		Short: "Read-only HD key derivation for BTC/ETH/SOL using chain code",
		Long: `Perform read-only hierarchical deterministic (HD) key derivation from a single vault share.
Uses the vault's chain code to derive public keys and addresses without reconstructing private keys.

This is safe for generating receiving addresses from any single vault share.`,
		Example: `  # Derive Bitcoin address at standard path
  vultool derive -f vault.vult --path "m/44'/0'/0'/0/0" --chain bitcoin
  
  # Derive Ethereum address with custom path
  vultool derive -f vault.vult --path "m/44'/60'/0'/0/5" --chain ethereum
  
  # Output in JSON format
  vultool derive -f vault.vult --path "m/44'/501'/0'/0'" --chain solana --json`,
		Run: func(cmd *cobra.Command, args []string) {
			if vaultFile == "" {
				fmt.Println("Vault file is required.")
				return
			}

			derivePath, _ := cmd.Flags().GetString("path")
			chainStr, _ := cmd.Flags().GetString("chain")
			useJSON, _ := cmd.Flags().GetBool("json")

			if derivePath == "" {
				fmt.Println("Derivation path is required (use --path).")
				return
			}

			if chainStr == "" {
				fmt.Println("Chain is required (use --chain).")
				return
			}

			// Validate the derivation path
			if err := recovery.ValidateDerivationPath(derivePath); err != nil {
				fmt.Printf("Invalid derivation path: %v\n", err)
				return
			}

			// Convert chain string to enum
			var chain recovery.SupportedChain
			switch strings.ToLower(chainStr) {
			case "bitcoin", "btc":
				chain = recovery.ChainBitcoin
			case "ethereum", "eth":
				chain = recovery.ChainEthereum
			case "solana", "sol":
				chain = recovery.ChainSolana
			case "thorchain", "thor":
				chain = recovery.ChainThorChain
			default:
				fmt.Printf("Unsupported chain: %s. Supported chains: bitcoin, ethereum, solana, thorchain\n", chainStr)
				return
			}

			absPath, err := filepath.Abs(vaultFile)
			if err != nil {
				fmt.Printf("Error getting absolute path: %v\n", err)
				return
			}

			fmt.Printf("🔄 Deriving %s address at path %s...\n", chain, derivePath)

			// Call the derivation function (currently stubbed)
			derivedKey, err := recovery.DeriveAddress(absPath, derivePath, chain, password)
			if err != nil {
				fmt.Printf("❌ Derivation failed: %v\n", err)
				return
			}

			if useJSON {
				if err := util.OutputResult(derivedKey, "json", os.Stdout); err != nil {
					fmt.Printf("Error outputting JSON: %v\n", err)
				}
			} else {
				fmt.Printf("✅ Derived %s address:\n\n", chain)
				fmt.Printf("  Chain:       %s\n", derivedKey.Chain)
				fmt.Printf("  Address:     %s\n", derivedKey.Address)
				fmt.Printf("  Derive Path: %s\n", derivedKey.DerivePath)
				if derivedKey.PrivateKey != "" {
					fmt.Printf("  Private Key: %s\n", derivedKey.PrivateKey)
				}
			}
		},
	}
	deriveCmd.Flags().StringVarP(&vaultFile, "vault", "f", "", "Path to the .vult vault file (required)")
	deriveCmd.Flags().StringVar(&password, "password", "", "Password for encrypted vault files")
	deriveCmd.Flags().String("path", "", "HD derivation path (e.g., m/44'/0'/0'/0/0) (required)")
	deriveCmd.Flags().String("chain", "", "Target blockchain: bitcoin, ethereum, solana, thorchain (required)")
	deriveCmd.Flags().Bool("json", false, "Output in JSON format")
	if err := deriveCmd.MarkFlagRequired("vault"); err != nil {
		fmt.Printf("Error setting up derive CLI flags: %v\n", err)
		os.Exit(1)
	}
	if err := deriveCmd.MarkFlagRequired("path"); err != nil {
		fmt.Printf("Error setting up derive CLI flags: %v\n", err)
		os.Exit(1)
	}
	if err := deriveCmd.MarkFlagRequired("chain"); err != nil {
		fmt.Printf("Error setting up derive CLI flags: %v\n", err)
		os.Exit(1)
	}

	// list-addresses-paths: enumerate addresses along common derivation paths
	listAddressesPathsCmd := &cobra.Command{
		Use:   "list-paths",
		Short: "Enumerate addresses along common derivation paths for major chains",
		Long: `List common HD derivation paths and addresses for supported blockchains.
Useful for discovering which addresses are associated with a vault.

Shows predefined common paths covering different address types (Legacy, SegWit, etc.) for Bitcoin
and sequential addresses for Ethereum and other chains. The --count flag is not yet implemented.`,
		Example: `  # List all common derivation paths for all chains
  vultool list-paths -f vault.vult
  
  # List only Bitcoin common paths (different address types)
  vultool list-paths -f vault.vult --chain bitcoin
  
  # Generate 20 sequential Ethereum addresses for gap limit scanning
  vultool list-paths -f vault.vult --chain ethereum --sequential --count 20
  
  # Generate sequential paths for all chains (gap limit recovery)
  vultool list-paths -f vault.vult --sequential --count 10
  
  # Output in JSON format
  vultool list-paths -f vault.vult --json`,
		Run: func(cmd *cobra.Command, args []string) {
			if vaultFile == "" {
				fmt.Println("Vault file is required.")
				return
			}

			chainFilter, _ := cmd.Flags().GetString("chain")
			count, _ := cmd.Flags().GetInt("count")
			useJSON, _ := cmd.Flags().GetBool("json")
			showPaths, _ := cmd.Flags().GetBool("show-paths")
			sequential, _ := cmd.Flags().GetBool("sequential")

			// Get paths - either common paths or sequential paths
			var allPaths map[types.SupportedChain][]types.DerivationPath

			if sequential {
				// Generate sequential paths for gap limit scanning
				allPaths = make(map[types.SupportedChain][]types.DerivationPath)

				if count == 0 {
					count = 20 // Default gap limit
				}

				if chainFilter != "" {
					// Generate for specific chain
					var targetChain types.SupportedChain
					switch strings.ToLower(chainFilter) {
					case "bitcoin", "btc":
						targetChain = types.ChainBitcoin
					case "bitcoincash", "bch":
						targetChain = types.ChainBitcoinCash
					case "litecoin", "ltc":
						targetChain = types.ChainLitecoin
					case "dogecoin", "doge":
						targetChain = types.ChainDogecoin
					case "dash":
						targetChain = types.ChainDash
					case "zcash", "zec":
						targetChain = types.ChainZcash
					case "ethereum", "eth":
						targetChain = types.ChainEthereum
					case "bsc", "binance":
						targetChain = types.ChainBSC
					case "avalanche", "avax":
						targetChain = types.ChainAvalanche
					case "polygon", "matic":
						targetChain = types.ChainPolygon
					case "cronoschain", "cronos", "cro":
						targetChain = types.ChainCronosChain
					case "arbitrum", "arb":
						targetChain = types.ChainArbitrum
					case "optimism", "op":
						targetChain = types.ChainOptimism
					case "base":
						targetChain = types.ChainBase
					case "blast":
						targetChain = types.ChainBlast
					case "zksync":
						targetChain = types.ChainZksync
					case "thorchain", "thor", "rune":
						targetChain = types.ChainThorChain
					case "solana", "sol":
						targetChain = types.ChainSolana
					case "sui":
						targetChain = types.ChainSUI
					default:
						fmt.Printf("Unsupported chain: %s\n", chainFilter)
						return
					}

					paths := types.GenerateSequentialPaths(targetChain, count)
					if len(paths) > 0 {
						allPaths[targetChain] = paths
					}
				} else {
					// Generate for all supported chains
					supportedChains := []types.SupportedChain{
						types.ChainBitcoin, types.ChainEthereum, types.ChainSolana, types.ChainThorChain,
					}
					for _, chain := range supportedChains {
						paths := types.GenerateSequentialPaths(chain, count)
						if len(paths) > 0 {
							allPaths[chain] = paths
						}
					}
				}
			} else {
				// Use common derivation paths (original behavior)
				allPaths = types.GetCommonDerivationPaths()
			}

			// Filter by chain if specified
			if chainFilter != "" {
				var targetChain types.SupportedChain
				switch strings.ToLower(chainFilter) {
				case "bitcoin", "btc":
					targetChain = types.ChainBitcoin
				case "bitcoincash", "bch":
					targetChain = types.ChainBitcoinCash
				case "litecoin", "ltc":
					targetChain = types.ChainLitecoin
				case "dogecoin", "doge":
					targetChain = types.ChainDogecoin
				case "dash":
					targetChain = types.ChainDash
				case "zcash", "zec":
					targetChain = types.ChainZcash
				case "ethereum", "eth":
					targetChain = types.ChainEthereum
				case "bsc", "binance":
					targetChain = types.ChainBSC
				case "avalanche", "avax":
					targetChain = types.ChainAvalanche
				case "polygon", "matic":
					targetChain = types.ChainPolygon
				case "cronoschain", "cronos", "cro":
					targetChain = types.ChainCronosChain
				case "arbitrum", "arb":
					targetChain = types.ChainArbitrum
				case "optimism", "op":
					targetChain = types.ChainOptimism
				case "base":
					targetChain = types.ChainBase
				case "blast":
					targetChain = types.ChainBlast
				case "zksync":
					targetChain = types.ChainZksync
				case "thorchain", "thor", "rune":
					targetChain = types.ChainThorChain
				case "solana", "sol":
					targetChain = types.ChainSolana
				case "sui":
					targetChain = types.ChainSUI
				default:
					fmt.Printf("Unsupported chain: %s\n", chainFilter)
					fmt.Printf("Supported chains: bitcoin, bitcoincash, litecoin, dogecoin, dash, zcash, ethereum, bsc, avalanche, polygon, cronoschain, arbitrum, optimism, base, blast, zksync, thorchain, solana, sui\n")
					return
				}

				filtered := make(map[types.SupportedChain][]types.DerivationPath)
				if paths, exists := allPaths[targetChain]; exists {
					filtered[targetChain] = paths
				}
				allPaths = filtered
			}

			if showPaths {
				// Just show the paths without deriving addresses
				if useJSON {
					if err := util.OutputResult(allPaths, "json", os.Stdout); err != nil {
						fmt.Printf("Error outputting JSON: %v\n", err)
					}
				} else {
					fmt.Println("📋 Common HD Derivation Paths:")
					fmt.Println()
					for chain, paths := range allPaths {
						fmt.Printf("🔗 %s:\n", strings.Title(string(chain)))
						for _, path := range paths {
							fmt.Printf("   %-20s %s (%s)\n", path.Path, path.Description, path.Purpose)
						}
						fmt.Println()
					}
				}
				return
			}

			// Parse the vault file to get keys for derivation
			absPath, err := filepath.Abs(vaultFile)
			if err != nil {
				fmt.Printf("Error getting absolute path: %v\n", err)
				return
			}

			vaultInfo, err := vault.ParseVaultFileWithPassword(absPath, password)
			if err != nil {
				fmt.Printf("Error parsing vault file: %v\n", err)
				return
			}

			// Derive addresses for all the specified paths
			pathAddresses := vault.DerivePathAddresses(vaultInfo, allPaths, count)

			if len(pathAddresses) == 0 {
				fmt.Println("No addresses could be derived from vault for the specified paths")
				return
			}

			if useJSON {
				if err := util.OutputResult(pathAddresses, "json", os.Stdout); err != nil {
					fmt.Printf("Error outputting JSON: %v\n", err)
				}
			} else {
				fmt.Printf("📋 HD derivation paths and addresses for vault: %s\n\n", filepath.Base(vaultFile))

				currentChain := ""
				for _, addr := range pathAddresses {
					if addr.Chain != currentChain {
						if currentChain != "" {
							fmt.Println() // Add space between chains
						}
						fmt.Printf("🔗 %s (%s):\n", addr.Chain, addr.Ticker)
						currentChain = addr.Chain
					}

					fmt.Printf("   %-20s %s\n", addr.DerivePath, addr.Address)
				}
			}
		},
	}
	listAddressesPathsCmd.Flags().StringVarP(&vaultFile, "vault", "f", "", "Path to the .vult vault file (required)")
	listAddressesPathsCmd.Flags().StringVar(&password, "password", "", "Password for encrypted vault files")
	listAddressesPathsCmd.Flags().String("chain", "", "Filter for specific blockchain (bitcoin, ethereum, solana, thorchain)")
	listAddressesPathsCmd.Flags().Int("count", 0, "Number of sequential addresses to generate (default: 20 for --sequential, ignored otherwise)")
	listAddressesPathsCmd.Flags().Bool("sequential", false, "Generate sequential addresses for gap limit scanning instead of common paths")
	listAddressesPathsCmd.Flags().Bool("json", false, "Output in JSON format")
	listAddressesPathsCmd.Flags().Bool("show-paths", false, "Show derivation paths only (don't derive addresses)")
	if err := listAddressesPathsCmd.MarkFlagRequired("vault"); err != nil {
		fmt.Printf("Error setting up list-paths CLI flags: %v\n", err)
		os.Exit(1)
	}

	// Add all commands to root
	rootCmd.AddCommand(inspectCmd)
	rootCmd.AddCommand(infoCmd)
	rootCmd.AddCommand(decodeCmd)
	rootCmd.AddCommand(verifyCmd)
	rootCmd.AddCommand(diffCmd)

	// Add Medic milestone commands
	rootCmd.AddCommand(recoverCmd)
	rootCmd.AddCommand(deriveCmd)
	rootCmd.AddCommand(listAddressesCmd)
	rootCmd.AddCommand(listAddressesPathsCmd)

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
