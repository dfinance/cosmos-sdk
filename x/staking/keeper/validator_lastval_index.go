package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/staking/types"
)

// Last Validator Index operations

// Load the last validator power.
// Returns zero if the operator was not a validator last block.
func (k Keeper) GetLastValidatorPower(ctx sdk.Context, operator sdk.ValAddress) (power int64) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.GetLastValidatorPowerKey(operator))
	if bz == nil {
		return 0
	}
	k.cdc.MustUnmarshalBinaryLengthPrefixed(bz, &power)
	return
}

// Set the last validator power.
func (k Keeper) SetLastValidatorPower(ctx sdk.Context, operator sdk.ValAddress, power int64) {
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshalBinaryLengthPrefixed(power)
	store.Set(types.GetLastValidatorPowerKey(operator), bz)
}

// Delete the last validator power.
func (k Keeper) DeleteLastValidatorPower(ctx sdk.Context, operator sdk.ValAddress) {
	store := ctx.KVStore(k.storeKey)
	store.Delete(types.GetLastValidatorPowerKey(operator))
}

// returns an iterator for the consensus validators in the last block
func (k Keeper) LastValidatorsIterator(ctx sdk.Context) (iterator sdk.Iterator) {
	store := ctx.KVStore(k.storeKey)
	iterator = sdk.KVStorePrefixIterator(store, types.LastValidatorPowerKey)
	return iterator
}

// Iterate over last validator powers.
func (k Keeper) IterateLastValidatorPowers(ctx sdk.Context, handler func(operator sdk.ValAddress, power int64) (stop bool)) {
	store := ctx.KVStore(k.storeKey)
	iter := sdk.KVStorePrefixIterator(store, types.LastValidatorPowerKey)
	defer iter.Close()
	for ; iter.Valid(); iter.Next() {
		addr := sdk.ValAddress(iter.Key()[len(types.LastValidatorPowerKey):])
		var power int64
		k.cdc.MustUnmarshalBinaryLengthPrefixed(iter.Value(), &power)
		if handler(addr, power) {
			break
		}
	}
}

// get the group of the bonded validators
func (k Keeper) GetLastValidators(ctx sdk.Context) (validators []types.Validator) {
	store := ctx.KVStore(k.storeKey)

	// add the actual validator power sorted store
	maxValidators := k.MaxValidators(ctx)
	validators = make([]types.Validator, maxValidators)

	iterator := sdk.KVStorePrefixIterator(store, types.LastValidatorPowerKey)
	defer iterator.Close()

	i := 0
	for ; iterator.Valid(); iterator.Next() {

		// sanity check
		if i >= int(maxValidators) {
			panic("more validators than maxValidators found")
		}
		address := types.AddressFromLastValidatorPowerKey(iterator.Key())
		validator := k.mustGetValidator(ctx, address)

		validators[i] = validator
		i++
	}
	return validators[:i] // trim
}
