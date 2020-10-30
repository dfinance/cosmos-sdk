package v0_39_1_0

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	v03902 "github.com/cosmos/cosmos-sdk/x/staking/legacy/v0_39-0_2"
)

// MigrateOptions defines migration inputs.
type MigrateOptions struct {
	ParamsMaxSelfDelegationLvl sdk.Int
}

// Migrate accepts exported genesis state from
// Dfinance v0.2 based on Cosmos SDK v0.39.1 and migrates it to
// Dfinance v1.0 based on Cosmos SDK v0.39.1 genesis state.
func Migrate(oldState v03902.GenesisState, migrateOpts MigrateOptions) GenesisState {
	return GenesisState{
		Params:               migrateParams(oldState.Params, migrateOpts),
		LastTotalPower:       oldState.LastTotalPower,
		LastValidatorPowers:  oldState.LastValidatorPowers,
		Validators:           oldState.Validators,
		Delegations:          oldState.Delegations,
		UnbondingDelegations: oldState.UnbondingDelegations,
		Redelegations:        oldState.Redelegations,
		StakingStates:        oldState.StakingStates,
		ScheduledUnbonds:     oldState.ScheduledUnbonds,
		BannedAccounts:       oldState.BannedAccounts,
		Exported:             oldState.Exported,
	}
}

// migrateParams adds a new param field.
func migrateParams(oldParams v03902.Params, migrateOpts MigrateOptions) Params {
	return Params{
		UnbondingTime:            oldParams.UnbondingTime,
		MaxValidators:            oldParams.MaxValidators,
		MaxEntries:               oldParams.MaxEntries,
		HistoricalEntries:        oldParams.HistoricalEntries,
		BondDenom:                oldParams.BondDenom,
		LPDenom:                  oldParams.LPDenom,
		LPDistrRatio:             oldParams.LPDistrRatio,
		MinSelfDelegationLvl:     oldParams.MinSelfDelegationLvl,
		MaxSelfDelegationLvl:     migrateOpts.ParamsMaxSelfDelegationLvl,
		MaxDelegationsRatio:      oldParams.MaxDelegationsRatio,
		ScheduledUnbondDelayTime: oldParams.ScheduledUnbondDelayTime,
	}
}
