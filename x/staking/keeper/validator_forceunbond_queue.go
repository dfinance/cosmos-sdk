package keeper

import (
	"bytes"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/staking/types"
)

// Scheduled validator force unbond Queue operations

// GetScheduledUnbondQueueValidators gets validator addresses for specific timestamp scheduled unbond queue.
func (k Keeper) GetScheduledUnbondQueueValidators(ctx sdk.Context, timestamp time.Time) (valAddrs []sdk.ValAddress) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.GetScheduledUnbondQueueTimeKey(timestamp))
	if bz == nil {
		return []sdk.ValAddress{}
	}
	k.cdc.MustUnmarshalBinaryLengthPrefixed(bz, &valAddrs)

	return valAddrs
}

// SetScheduledUnbondQueueValidators sets validator addresses for specific timestamp scheduled unbond queue.
func (k Keeper) SetScheduledUnbondQueueValidators(ctx sdk.Context, timestamp time.Time, valAddrs []sdk.ValAddress) {
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshalBinaryLengthPrefixed(valAddrs)
	store.Set(types.GetScheduledUnbondQueueTimeKey(timestamp), bz)
}

// DeleteScheduledUnbondQueueValidators removes all validator addresses for specific timestamp scheduled unbond queue.
func (k Keeper) DeleteScheduledUnbondQueueValidators(ctx sdk.Context, timestamp time.Time) {
	store := ctx.KVStore(k.storeKey)
	store.Delete(types.GetScheduledUnbondQueueTimeKey(timestamp))
}

// InsertScheduledUnbondQueueValidator inserts a validator address for an appropriate timestamp scheduled unbond queue.
func (k Keeper) InsertScheduledUnbondQueueValidator(ctx sdk.Context, val types.Validator) {
	timestamp := val.ScheduledUnbondStartTime

	valAddrs := k.GetScheduledUnbondQueueValidators(ctx, timestamp)
	valAddrs = append(valAddrs, val.OperatorAddress)

	k.SetScheduledUnbondQueueValidators(ctx, val.ScheduledUnbondStartTime, valAddrs)
}

// DeleteScheduledUnbondQueueValidator removes a validator address for an appropriate timestamp scheduled unbond queue.
func (k Keeper) DeleteScheduledUnbondQueueValidator(ctx sdk.Context, val types.Validator) {
	timestamp := val.ScheduledUnbondStartTime
	valAddrs := k.GetScheduledUnbondQueueValidators(ctx, timestamp)
	newValAddrs := []sdk.ValAddress{}

	for _, valAddr := range valAddrs {
		if !bytes.Equal(valAddr, val.OperatorAddress) {
			newValAddrs = append(newValAddrs, valAddr)
		}
	}

	if len(newValAddrs) == 0 {
		k.DeleteScheduledUnbondQueueValidators(ctx, timestamp)
	} else {
		k.SetScheduledUnbondQueueValidators(ctx, timestamp, newValAddrs)
	}
}

// IterateScheduledUnbondQueue iterates over ScheduledUnbondQueue.
func (k Keeper) IterateScheduledUnbondQueue(ctx sdk.Context, handler func(timestamp time.Time, valAddrs []sdk.ValAddress) (stop bool)) {
	store := ctx.KVStore(k.storeKey)
	iterator := sdk.KVStorePrefixIterator(store, types.ScheduledUnbondQueueKey)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		timestamp := types.ParseUnbondingDelegationTimeKey(iterator.Key())
		var valAddrs []sdk.ValAddress
		k.cdc.MustUnmarshalBinaryLengthPrefixed(iterator.Value(), &valAddrs)

		if handler(timestamp, valAddrs) {
			break
		}
	}
}

// ScheduledUnbondQueueIterator returns iterator for all the scheduled unbond queue validator addresses from time 0 until endTime.
func (k Keeper) ScheduledUnbondQueueIterator(ctx sdk.Context, endTime time.Time) sdk.Iterator {
	store := ctx.KVStore(k.storeKey)
	endKey := types.GetScheduledUnbondQueueTimeKey(endTime)

	return store.Iterator(types.ScheduledUnbondQueueKey, sdk.InclusiveEndBytes(endKey))
}

// GetAllScheduledUnbondQueueMatureValidators returns all the scheduled unbond queue validator addresses from time 0 until endTime.
func (k Keeper) GetAllScheduledUnbondQueueMatureValidators(ctx sdk.Context, endTime time.Time) (valsAddrs []sdk.ValAddress) {
	iterator := k.ScheduledUnbondQueueIterator(ctx, endTime)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		curAddrs := []sdk.ValAddress{}
		k.cdc.MustUnmarshalBinaryLengthPrefixed(iterator.Value(), &curAddrs)
		valsAddrs = append(valsAddrs, curAddrs...)
	}

	return
}

// ProcessAllScheduledUnbondQueueMatureValidators processes all the scheduled unbond queue validator addresses from time 0 until current blockTime.
// Validator addresses are removed from the queue after the processing is done.
func (k Keeper) ProcessAllScheduledUnbondQueueMatureValidators(ctx sdk.Context) {
	store := ctx.KVStore(k.storeKey)
	iterator := k.ScheduledUnbondQueueIterator(ctx, ctx.BlockHeader().Time)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		valsAddrs := []sdk.ValAddress{}
		k.cdc.MustUnmarshalBinaryLengthPrefixed(iterator.Value(), &valsAddrs)

		for _, valAddr := range valsAddrs {
			val, found := k.GetValidator(ctx, valAddr)
			if !found {
				panic("validator in the scheduled unbond queue was not found")
			}
			if !val.ScheduledToUnbond {
				panic("unexpected validator in the scheduled unbond queue; ScheduledToUnbond flag was not set")
			}

			if updVal := k.completeForceUnbondValidator(ctx, val); updVal != nil {
				k.SetValidator(ctx, updVal.UnscheduleValidatorForceUnbond())
			}
		}

		store.Delete(iterator.Key())
	}
}
