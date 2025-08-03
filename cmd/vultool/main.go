package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

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

				if err := vault.ExportToJSON(vaultInfo, file, true); err != nil {
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

			// Check if YAML output was requested
			useYAML, err := cmd.Flags().GetBool("yaml")
			if err != nil {
				fmt.Printf("Error reading yaml flag: %v\n", err)
				return
			}

			if useYAML {
				// Output YAML to stdout
				if err := vault.ExportToYAML(vaultInfo, os.Stdout); err != nil {
					fmt.Printf("Error outputting YAML: %v\n", err)
					return
				}
			} else {
				// Output JSON to stdout (default)
				if err := vault.ExportToJSON(vaultInfo, os.Stdout, true); err != nil {
					fmt.Printf("Error outputting JSON: %v\n", err)
					return
				}
			}
		},
	}
	decodeCmd.Flags().StringVarP(&vaultFile, "vault", "f", "", "Path to the .vult vault file (required)")
	decodeCmd.Flags().StringVar(&password, "password", "", "Password for encrypted vault files (alternative to interactive prompt)")
	decodeCmd.Flags().Bool("yaml", false, "Output in YAML format instead of JSON")
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
			fmt.Println(vault.FormatDiff(diff, true))
		},
	}
	diffCmd.Flags().StringVar(&password, "password", "", "Password for encrypted vault files (alternative to interactive prompt)")

	// Add all commands to root
	rootCmd.AddCommand(inspectCmd)
	rootCmd.AddCommand(infoCmd)
	rootCmd.AddCommand(decodeCmd)
	rootCmd.AddCommand(verifyCmd)
	rootCmd.AddCommand(diffCmd)

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
