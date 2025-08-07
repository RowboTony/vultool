package vault

// VaultAddress represents a blockchain address derived from vault public keys
type VaultAddress struct {
	Chain      string `json:"chain"`
	Ticker     string `json:"ticker"`
	Address    string `json:"address"`
	DerivePath string `json:"derive_path,omitempty"`
	IsNative   bool   `json:"is_native,omitempty"`
}
