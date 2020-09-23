package types

import "time"

// GenesisState - minter state
type GenesisState struct {
	Minter         Minter         `json:"minter" yaml:"minter"`                     // minter object
	BlockDurFilter BlockDurFilter `json:"block_dur_filter" yaml:"block_dur_filter"` // block duration estimation filter
	AnnualUpdateTS time.Time      `json:"annual_update_ts" yaml:"annual_update_ts"` // annual params update timestamp
	Params         Params         `json:"params" yaml:"params"`                     // inflation params
}

// NewGenesisState creates a new GenesisState object
func NewGenesisState(minter Minter, params Params, blockDurFilter BlockDurFilter, annualUpdateTs time.Time) GenesisState {
	return GenesisState{
		Minter:         minter,
		BlockDurFilter: blockDurFilter,
		AnnualUpdateTS: annualUpdateTs,
		Params:         params,
	}
}

// DefaultGenesisState creates a default GenesisState object
func DefaultGenesisState() GenesisState {
	return GenesisState{
		Minter:         DefaultInitialMinter(),
		BlockDurFilter: BlockDurFilter{},
		AnnualUpdateTS: time.Time{},
		Params:         DefaultParams(),
	}
}

// ValidateGenesis validates the provided genesis state to ensure the
// expected invariants holds.
func ValidateGenesis(data GenesisState) error {
	if err := data.Params.Validate(); err != nil {
		return err
	}

	return ValidateMinter(data.Minter)
}
