package types

import sdk "github.com/cosmos/cosmos-sdk/types"

var (
	// MinterKey is used for the keeper store
	MinterKey = []byte{0x00}
	// BlockDurFilterKey is used to store avg block duration filter
	BlockDurFilterKey = []byte("BlockDurFilter")
	// AnnualUpdateTimestampKey is used to store timestamp of the next annual params update (new year)
	AnnualUpdateTimestampKey = []byte("AnnualUpdateTimestamp")
	// FoundationAllocationRatioMaxValue is used to validate max value of the FoundationAllocationRatio
	FoundationAllocationRatioMaxValue = sdk.NewDecWithPrec(45, 2)
)

// nolint
const (
	// ModuleName
	ModuleName = "mint"

	// DefaultParamspace params keeper
	DefaultParamspace = ModuleName

	// StoreKey is the default store key for mint
	StoreKey = ModuleName

	// QuerierRoute is the querier route for the minting store.
	QuerierRoute = StoreKey

	// Query endpoints supported by the minting querier
	QueryParameters             = "parameters"
	QueryInflation              = "inflation"
	QueryAnnualProvisions       = "annual_provisions"
	QueryBlocksPerYear          = "blocks_per_year"
	QueryNextAnnualParamsUpdate = "next_annual_params_update"
)
