package types

// BannedAccInfo keeps banned account info.
// Account address is stored within the storage key.
type BannedAccInfo struct {
	// BlockHeight at which account was banned
	Height int64 `json:"height"`
}
