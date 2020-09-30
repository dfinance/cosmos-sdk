package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/staking/types"
)

// GetValidatorStakingState gets validator staking state.
func (k Keeper) GetValidatorStakingState(ctx sdk.Context, addr sdk.ValAddress) (state types.ValidatorStakingState) {
	store := ctx.KVStore(k.storeKey)
	key := types.GetValidatorStakingStateKey(addr)

	bz := store.Get(key)
	if bz == nil {
		return types.NewValidatorStakingState()
	}
	k.cdc.MustUnmarshalBinaryLengthPrefixed(bz, &state)

	return
}

// SetValidatorStakingState sets validator staking state for specified validator address.
func (k Keeper) SetValidatorStakingState(ctx sdk.Context, addr sdk.ValAddress, state types.ValidatorStakingState) {
	store := ctx.KVStore(k.storeKey)
	key := types.GetValidatorStakingStateKey(addr)
	state.Sort()

	bz := k.cdc.MustMarshalBinaryLengthPrefixed(state)
	store.Set(key, bz)
}

// DeleteValidatorStakingState removes validator staking state for specified validator address.
func (k Keeper) DeleteValidatorStakingState(ctx sdk.Context, addr sdk.ValAddress) {
	store := ctx.KVStore(k.storeKey)
	key := types.GetValidatorStakingStateKey(addr)

	store.Delete(key)
}

// IterateValidatorStakingStates iterates over all validators staking state entries.
func (k Keeper) IterateValidatorStakingStates(ctx sdk.Context, handler func(valAddr sdk.ValAddress, state types.ValidatorStakingState) (stop bool)) {
	store := ctx.KVStore(k.storeKey)
	iterator := sdk.KVStorePrefixIterator(store, types.ValidatorsStakingStateKey)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		valAddr := types.ParseValidatorStakingStateKey(iterator.Key())
		var state types.ValidatorStakingState
		k.cdc.MustUnmarshalBinaryLengthPrefixed(iterator.Value(), &state)

		if handler(valAddr, state) {
			break
		}
	}
}

// SetValidatorStakingStateDelegation adds / sets validator staking state delegation info.
func (k Keeper) SetValidatorStakingStateDelegation(ctx sdk.Context,
	valAddr sdk.ValAddress, delAddr sdk.AccAddress,
	delBondingShares, delLPShares sdk.Dec,
) types.ValidatorStakingState {

	state := k.GetValidatorStakingState(ctx, valAddr)
	state = state.SetDelegator(valAddr, delAddr, delBondingShares, delLPShares)
	k.SetValidatorStakingState(ctx, valAddr, state)

	return state
}

// RemoveValidatorStakingStateDelegation removes validator staking state delegation info.
func (k Keeper) RemoveValidatorStakingStateDelegation(ctx sdk.Context,
	valAddr sdk.ValAddress, delAddr sdk.AccAddress,
) types.ValidatorStakingState {

	state := k.GetValidatorStakingState(ctx, valAddr)
	state = state.RemoveDelegator(delAddr)
	k.SetValidatorStakingState(ctx, valAddr, state)

	return state
}
