package vault

import (
	"github.com/rowbotony/vultool/internal/types"
	"strings"
)

// DerivePathAddresses derives addresses for specified paths
func DerivePathAddresses(vaultInfo *VaultInfo, paths map[types.SupportedChain][]types.DerivationPath, count int) []VaultAddress {
	// For now, just return the default addresses from DeriveAddressesFromVault
	// This function can be enhanced later to support custom paths
	return DeriveAddressesFromVault(vaultInfo)
}

// getDefaultPath returns the default derivation path for a chain
func getDefaultPath(chain string) string {
	// Default paths for common chains
	switch strings.ToLower(chain) {
	case "bitcoin", "btc":
		return "m/84'/0'/0'/0/0"
	case "bitcoin-cash", "bch":
		return "m/44'/145'/0'/0/0"
	case "litecoin", "ltc":
		return "m/84'/2'/0'/0/0"
	case "dogecoin", "doge":
		return "m/44'/3'/0'/0/0"
	case "dash":
		return "m/44'/5'/0'/0/0"
	case "ethereum", "eth":
		return "m/44'/60'/0'/0/0"
	case "thorchain", "rune":
		return "m/44'/931'/0'/0/0"
	case "sui":
		return "m/44'/784'/0'/0'/0"
	default:
		// For EVM chains and others, use Ethereum path
		return "m/44'/60'/0'/0/0"
	}
}

// GetPathsForChain returns the derivation paths for a specific chain
func GetPathsForChain(chain string) []string {
	// For now, return just the default path
	// This can be expanded to support multiple paths per chain
	return []string{getDefaultPath(chain)}
}
