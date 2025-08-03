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

func main() {
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

	rootCmd.AddCommand(inspectCmd)

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
