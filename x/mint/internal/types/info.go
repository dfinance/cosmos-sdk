package types

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// MintInfo contains current minter extended state.
type MintInfo struct {
	MinInflation              sdk.Dec   `json:"min_inflation" yaml:"min_inflation"`
	MaxInflation              sdk.Dec   `json:"max_inflation" yaml:"max_inflation"`
	InfPwrBondedLockedRatio   sdk.Dec   `json:"infpwr_bondedlocked_ratio" yaml:"infpwr_bondedlocked_ratio"`
	FoundationAllocationRatio sdk.Dec   `json:"foundation_allocation_ratio" yaml:"foundation_allocation_ratio"`
	StakingTotalSupplyShift   sdk.Int   `json:"staking_total_supply_shift" yaml:"staking_total_supply_shift"`
	InflationMain             sdk.Dec   `json:"inflation_main" yaml:"inflation_main"`
	InflationFoundation       sdk.Dec   `json:"inflation_foundation" yaml:"inflation_foundation"`
	AnnualProvisionMain       sdk.Dec   `json:"annual_provision_main" yaml:"annual_provision_main"`
	AnnualProvisionFoundation sdk.Dec   `json:"annual_provision_foundation" yaml:"annual_provision_foundation"`
	BlocksPerYearEstimation   uint64    `json:"blocks_per_year_estimation" yaml:"blocks_per_year_estimation"`
	BondedRatio               sdk.Dec   `json:"bonded_ratio" yaml:"bonded_ratio"`
	LockedRatio               sdk.Dec   `json:"locked_ratio" yaml:"locked_ratio"`
	NextAnnualUpdate          time.Time `json:"next_annual_update" yaml:"next_annual_update"`
}

// NewMintInfo creates a new NewMintInfo object.
func NewMintInfo(params Params, minter Minter, bondedRatio, lockedRatio sdk.Dec, nextAnnualUpdate time.Time) MintInfo {
	return MintInfo{
		MinInflation:              params.InflationMin,
		MaxInflation:              params.InflationMax,
		InfPwrBondedLockedRatio:   params.InfPwrBondedLockedRatio,
		FoundationAllocationRatio: params.FoundationAllocationRatio,
		StakingTotalSupplyShift:   params.StakingTotalSupplyShift,
		InflationMain:             minter.Inflation,
		InflationFoundation:       minter.FoundationInflation,
		AnnualProvisionMain:       minter.Provisions,
		AnnualProvisionFoundation: minter.FoundationProvisions,
		BlocksPerYearEstimation:   minter.BlocksPerYear,
		BondedRatio:               bondedRatio,
		LockedRatio:               lockedRatio,
		NextAnnualUpdate:          nextAnnualUpdate,
	}
}
