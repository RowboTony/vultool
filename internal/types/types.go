package types

import "fmt"

// SupportedChain represents a blockchain we can derive keys for
type SupportedChain string

const (
	// ECDSA-based chains
	ChainBitcoin     SupportedChain = "bitcoin"
	ChainBitcoinCash SupportedChain = "bitcoincash"
	ChainLitecoin    SupportedChain = "litecoin"
	ChainDogecoin    SupportedChain = "dogecoin"
	ChainDash        SupportedChain = "dash"
	ChainZcash       SupportedChain = "zcash"
	ChainEthereum    SupportedChain = "ethereum"
	ChainBSC         SupportedChain = "bsc"
	ChainAvalanche   SupportedChain = "avalanche"
	ChainPolygon     SupportedChain = "polygon"
	ChainCronosChain SupportedChain = "cronoschain"
	ChainArbitrum    SupportedChain = "arbitrum"
	ChainOptimism    SupportedChain = "optimism"
	ChainBase        SupportedChain = "base"
	ChainBlast       SupportedChain = "blast"
	ChainZksync      SupportedChain = "zksync"
	ChainThorChain   SupportedChain = "thorchain"
	// EdDSA-based chains
	ChainSolana SupportedChain = "solana"
	ChainSUI    SupportedChain = "sui"
)

// DerivationPath represents an HD derivation path
type DerivationPath struct {
	Path        string         `json:"path"`
	Chain       SupportedChain `json:"chain"`
	Description string         `json:"description"`
	Purpose     string         `json:"purpose"` // e.g., "receiving", "change", "legacy"
}

// GetCommonDerivationPaths returns common HD derivation paths for supported chains
func GetCommonDerivationPaths() map[SupportedChain][]DerivationPath {
	return map[SupportedChain][]DerivationPath{
		// Bitcoin with multiple path types
		ChainBitcoin: {
			{Path: "m/44'/0'/0'/0/0", Chain: ChainBitcoin, Description: "First receiving address (P2PKH)", Purpose: "receiving"},
			{Path: "m/44'/0'/0'/1/0", Chain: ChainBitcoin, Description: "First change address (P2PKH)", Purpose: "change"},
			{Path: "m/49'/0'/0'/0/0", Chain: ChainBitcoin, Description: "P2SH-P2WPKH (SegWit v0)", Purpose: "segwit"},
			{Path: "m/84'/0'/0'/0/0", Chain: ChainBitcoin, Description: "P2WPKH (Native SegWit)", Purpose: "native_segwit"},
			{Path: "m/84'/0'/0'/0/1", Chain: ChainBitcoin, Description: "Second Native SegWit address", Purpose: "receiving"},
		},
		// Bitcoin Cash
		ChainBitcoinCash: {
			{Path: "m/44'/145'/0'/0/0", Chain: ChainBitcoinCash, Description: "Bitcoin Cash main address", Purpose: "receiving"},
			{Path: "m/44'/145'/0'/0/1", Chain: ChainBitcoinCash, Description: "Bitcoin Cash second address", Purpose: "receiving"},
			{Path: "m/44'/145'/0'/1/0", Chain: ChainBitcoinCash, Description: "Bitcoin Cash change address", Purpose: "change"},
		},
		// Litecoin
		ChainLitecoin: {
			{Path: "m/84'/2'/0'/0/0", Chain: ChainLitecoin, Description: "Litecoin Native SegWit", Purpose: "receiving"},
			{Path: "m/44'/2'/0'/0/0", Chain: ChainLitecoin, Description: "Litecoin Legacy P2PKH", Purpose: "receiving"},
			{Path: "m/49'/2'/0'/0/0", Chain: ChainLitecoin, Description: "Litecoin SegWit P2SH", Purpose: "segwit"},
		},
		// Dogecoin
		ChainDogecoin: {
			{Path: "m/44'/3'/0'/0/0", Chain: ChainDogecoin, Description: "Dogecoin main address", Purpose: "receiving"},
			{Path: "m/44'/3'/0'/0/1", Chain: ChainDogecoin, Description: "Dogecoin second address", Purpose: "receiving"},
			{Path: "m/44'/3'/0'/1/0", Chain: ChainDogecoin, Description: "Dogecoin change address", Purpose: "change"},
		},
		// Dash
		ChainDash: {
			{Path: "m/44'/5'/0'/0/0", Chain: ChainDash, Description: "Dash main address", Purpose: "receiving"},
			{Path: "m/44'/5'/0'/0/1", Chain: ChainDash, Description: "Dash second address", Purpose: "receiving"},
			{Path: "m/44'/5'/0'/1/0", Chain: ChainDash, Description: "Dash change address", Purpose: "change"},
		},
		// Zcash
		ChainZcash: {
			{Path: "m/44'/133'/0'/0/0", Chain: ChainZcash, Description: "Zcash transparent address", Purpose: "receiving"},
			{Path: "m/44'/133'/0'/0/1", Chain: ChainZcash, Description: "Zcash second address", Purpose: "receiving"},
			{Path: "m/44'/133'/0'/1/0", Chain: ChainZcash, Description: "Zcash change address", Purpose: "change"},
		},
		// Ethereum and EVM chains (they all use the same paths)
		ChainEthereum: {
			{Path: "m/44'/60'/0'/0/0", Chain: ChainEthereum, Description: "First Ethereum address", Purpose: "receiving"},
			{Path: "m/44'/60'/0'/0/1", Chain: ChainEthereum, Description: "Second Ethereum address", Purpose: "receiving"},
			{Path: "m/44'/60'/0'/0/2", Chain: ChainEthereum, Description: "Third Ethereum address", Purpose: "receiving"},
			{Path: "m/44'/60'/0'/0/3", Chain: ChainEthereum, Description: "Fourth Ethereum address", Purpose: "receiving"},
			{Path: "m/44'/60'/0'/0/4", Chain: ChainEthereum, Description: "Fifth Ethereum address", Purpose: "receiving"},
		},
		// BSC (Binance Smart Chain)
		ChainBSC: {
			{Path: "m/44'/60'/0'/0/0", Chain: ChainBSC, Description: "BSC main address", Purpose: "receiving"},
			{Path: "m/44'/60'/0'/0/1", Chain: ChainBSC, Description: "BSC second address", Purpose: "receiving"},
			{Path: "m/44'/60'/0'/0/2", Chain: ChainBSC, Description: "BSC third address", Purpose: "receiving"},
		},
		// Avalanche
		ChainAvalanche: {
			{Path: "m/44'/60'/0'/0/0", Chain: ChainAvalanche, Description: "Avalanche main address", Purpose: "receiving"},
			{Path: "m/44'/60'/0'/0/1", Chain: ChainAvalanche, Description: "Avalanche second address", Purpose: "receiving"},
			{Path: "m/44'/60'/0'/0/2", Chain: ChainAvalanche, Description: "Avalanche third address", Purpose: "receiving"},
		},
		// Polygon
		ChainPolygon: {
			{Path: "m/44'/60'/0'/0/0", Chain: ChainPolygon, Description: "Polygon main address", Purpose: "receiving"},
			{Path: "m/44'/60'/0'/0/1", Chain: ChainPolygon, Description: "Polygon second address", Purpose: "receiving"},
			{Path: "m/44'/60'/0'/0/2", Chain: ChainPolygon, Description: "Polygon third address", Purpose: "receiving"},
		},
		// Cronos Chain
		ChainCronosChain: {
			{Path: "m/44'/60'/0'/0/0", Chain: ChainCronosChain, Description: "Cronos main address", Purpose: "receiving"},
			{Path: "m/44'/60'/0'/0/1", Chain: ChainCronosChain, Description: "Cronos second address", Purpose: "receiving"},
		},
		// Arbitrum
		ChainArbitrum: {
			{Path: "m/44'/60'/0'/0/0", Chain: ChainArbitrum, Description: "Arbitrum main address", Purpose: "receiving"},
			{Path: "m/44'/60'/0'/0/1", Chain: ChainArbitrum, Description: "Arbitrum second address", Purpose: "receiving"},
		},
		// Optimism
		ChainOptimism: {
			{Path: "m/44'/60'/0'/0/0", Chain: ChainOptimism, Description: "Optimism main address", Purpose: "receiving"},
			{Path: "m/44'/60'/0'/0/1", Chain: ChainOptimism, Description: "Optimism second address", Purpose: "receiving"},
		},
		// Base
		ChainBase: {
			{Path: "m/44'/60'/0'/0/0", Chain: ChainBase, Description: "Base main address", Purpose: "receiving"},
			{Path: "m/44'/60'/0'/0/1", Chain: ChainBase, Description: "Base second address", Purpose: "receiving"},
		},
		// Blast
		ChainBlast: {
			{Path: "m/44'/60'/0'/0/0", Chain: ChainBlast, Description: "Blast main address", Purpose: "receiving"},
			{Path: "m/44'/60'/0'/0/1", Chain: ChainBlast, Description: "Blast second address", Purpose: "receiving"},
		},
		// zkSync
		ChainZksync: {
			{Path: "m/44'/60'/0'/0/0", Chain: ChainZksync, Description: "zkSync main address", Purpose: "receiving"},
			{Path: "m/44'/60'/0'/0/1", Chain: ChainZksync, Description: "zkSync second address", Purpose: "receiving"},
		},
		// THORChain
		ChainThorChain: {
			{Path: "m/44'/931'/0'/0/0", Chain: ChainThorChain, Description: "THORChain main address", Purpose: "receiving"},
			{Path: "m/44'/931'/0'/0/1", Chain: ChainThorChain, Description: "THORChain second address", Purpose: "receiving"},
		},
		// Solana
		ChainSolana: {
			{Path: "m/44'/501'/0'/0'", Chain: ChainSolana, Description: "Solana main account", Purpose: "receiving"},
			{Path: "m/44'/501'/1'/0'", Chain: ChainSolana, Description: "Solana second account", Purpose: "receiving"},
			{Path: "m/44'/501'/2'/0'", Chain: ChainSolana, Description: "Solana third account", Purpose: "receiving"},
		},
		// SUI
		ChainSUI: {
			{Path: "m/44'/784'/0'/0'/0'", Chain: ChainSUI, Description: "SUI main address", Purpose: "receiving"},
			{Path: "m/44'/784'/0'/0'/1'", Chain: ChainSUI, Description: "SUI second address", Purpose: "receiving"},
		},
	}
}

// GenerateSequentialPaths generates sequential HD derivation paths for gap limit scanning
// This is essential for wallet recovery where users may have used non-consecutive addresses
func GenerateSequentialPaths(chain SupportedChain, count int) []DerivationPath {
	var paths []DerivationPath
	basePath := getBasePath(chain)

	if basePath == "" {
		return paths // Unsupported chain
	}

	for i := 0; i < count; i++ {
		path := fmt.Sprintf("%s%d", basePath, i)
		paths = append(paths, DerivationPath{
			Path:        path,
			Chain:       chain,
			Description: fmt.Sprintf("Address #%d (sequential)", i),
			Purpose:     "sequential",
		})
	}

	return paths
}

// getBasePath returns the base derivation path for sequential address generation
func getBasePath(chain SupportedChain) string {
	switch chain {
	// ECDSA-based chains
	case ChainBitcoin:
		return "m/84'/0'/0'/0/" // Native SegWit is most common modern standard
	case ChainBitcoinCash:
		return "m/44'/145'/0'/0/"
	case ChainLitecoin:
		return "m/84'/2'/0'/0/" // Native SegWit
	case ChainDogecoin:
		return "m/44'/3'/0'/0/"
	case ChainDash:
		return "m/44'/5'/0'/0/"
	case ChainZcash:
		return "m/44'/133'/0'/0/"
	case ChainEthereum, ChainBSC, ChainAvalanche, ChainPolygon, ChainCronosChain, ChainArbitrum, ChainOptimism, ChainBase, ChainBlast, ChainZksync:
		return "m/44'/60'/0'/0/" // All EVM chains use same path
	case ChainThorChain:
		return "m/44'/931'/0'/0/"
	// EdDSA-based chains
	case ChainSolana:
		return "m/44'/501'/" // Solana uses different format: m/44'/501'/0'/0', m/44'/501'/1'/0', etc.
	case ChainSUI:
		return "m/44'/784'/0'/0'/" // SUI has different format with trailing /
	default:
		return "" // Unsupported
	}
}

// GetSupportedChains returns a list of all supported blockchain chains
func GetSupportedChains() []SupportedChain {
	return []SupportedChain{
		ChainBitcoin,
		ChainEthereum,
		ChainSolana,
		ChainThorChain,
	}
}
