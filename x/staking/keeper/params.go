package keeper

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/params"
	"github.com/cosmos/cosmos-sdk/x/staking/types"
)

// Default parameter namespace
const (
	DefaultParamspace = types.ModuleName
)

// ParamTable for staking module
func ParamKeyTable() params.KeyTable {
	return params.NewKeyTable().RegisterParamSet(&types.Params{})
}

// UnbondingTime
func (k Keeper) UnbondingTime(ctx sdk.Context) (res time.Duration) {
	k.paramstore.Get(ctx, types.KeyUnbondingTime, &res)
	return
}

// MaxValidators - Maximum number of validators
func (k Keeper) MaxValidators(ctx sdk.Context) (res uint16) {
	k.paramstore.Get(ctx, types.KeyMaxValidators, &res)
	return
}

// MaxEntries - Maximum number of simultaneous unbonding
// delegations or redelegations (per pair/trio)
func (k Keeper) MaxEntries(ctx sdk.Context) (res uint16) {
	k.paramstore.Get(ctx, types.KeyMaxEntries, &res)
	return
}

// HistoricalEntries = number of historical info entries
// to persist in store
func (k Keeper) HistoricalEntries(ctx sdk.Context) (res uint16) {
	k.paramstore.Get(ctx, types.KeyHistoricalEntries, &res)
	return
}

// BondDenom - bondable coin denomination
func (k Keeper) BondDenom(ctx sdk.Context) (res string) {
	k.paramstore.Get(ctx, types.KeyBondDenom, &res)
	return
}

// LPDenom - liquidity coin denomination
func (k Keeper) LPDenom(ctx sdk.Context) (res string) {
	k.paramstore.Get(ctx, types.KeyLPDenom, &res)
	return
}

// LPDistrRatio - validator distribution and gov voting power LP ratio (BTokens + LPDistrRatio * LPTokens)
func (k Keeper) LPDistrRatio(ctx sdk.Context) (res sdk.Dec) {
	k.paramstore.Get(ctx, types.KeyLPDistrRatio, &res)
	return
}

// MinSelfDelegationLvl - min self-delegation level for validator
func (k Keeper) MinSelfDelegationLvl(ctx sdk.Context) (res sdk.Int) {
	k.paramstore.Get(ctx, types.KeyMinSelfDelegationLvl, &res)
	return
}

// MaxSelfDelegationLvl - max self-delegation level for validator
func (k Keeper) MaxSelfDelegationLvl(ctx sdk.Context) (res sdk.Int) {
	k.paramstore.Get(ctx, types.KeyMaxSelfDelegationLvl, &res)
	return
}

// MaxDelegationsRatio - max delegations ratio (MaxDelegationsAmount = SelfDelegation * MaxDelegationsRatio)
func (k Keeper) MaxDelegationsRatio(ctx sdk.Context) (res sdk.Dec) {
	k.paramstore.Get(ctx, types.KeyMaxDelegationsRatio, &res)
	return
}

// ScheduledUnbondDelay - force unbond scheduler delay
func (k Keeper) ScheduledUnbondDelay(ctx sdk.Context) (res time.Duration) {
	k.paramstore.Get(ctx, types.KeyScheduledUnbondDelayTime, &res)
	return
}

// Get all parameteras as types.Params
func (k Keeper) GetParams(ctx sdk.Context) types.Params {
	return types.NewParams(
		k.UnbondingTime(ctx),
		k.MaxValidators(ctx),
		k.MaxEntries(ctx),
		k.HistoricalEntries(ctx),
		k.BondDenom(ctx),
		k.LPDenom(ctx),
		k.LPDistrRatio(ctx),
		k.MinSelfDelegationLvl(ctx),
		k.MaxSelfDelegationLvl(ctx),
		k.MaxDelegationsRatio(ctx),
		k.ScheduledUnbondDelay(ctx),
	)
}

// set the params
func (k Keeper) SetParams(ctx sdk.Context, params types.Params) {
	k.paramstore.SetParamSet(ctx, &params)
}
