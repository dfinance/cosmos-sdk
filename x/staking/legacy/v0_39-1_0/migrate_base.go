package v0_39_1_0

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	v03902 "github.com/cosmos/cosmos-sdk/x/staking/legacy/v0_39-0_2"
)

// MigrateBaseOptions defines base migration inputs.
type MigrateBaseOptions struct {
	ParamsMaxSelfDelegationLvl sdk.Int
}

// MigrateBase accepts exported genesis state from
// Dfinance v0.2 based on Cosmos SDK v0.39.1 and migrates it to
// Dfinance v1.0 based on Cosmos SDK v0.39.1 genesis state.
// Base migration performs only necessary operations, so state would be identical.
func MigrateBase(oldState v03902.GenesisState, migrateOpts MigrateBaseOptions) GenesisState {
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
func migrateParams(oldParams v03902.Params, migrateOpts MigrateBaseOptions) Params {
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
