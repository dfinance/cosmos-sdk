package keeper

import (
	"bytes"
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/staking/types"
)

// Cache the amino decoding of validators, as it can be the case that repeated slashing calls
// cause many calls to GetValidator, which were shown to throttle the state machine in our
// simulation. Note this is quite biased though, as the simulator does more slashes than a
// live chain should, however we require the slashing to be fast as noone pays gas for it.
type cachedValidator struct {
	val        types.Validator
	marshalled string // marshalled amino bytes for the validator object (not operator address)
}

func newCachedValidator(val types.Validator, marshalled string) cachedValidator {
	return cachedValidator{
		val:        val,
		marshalled: marshalled,
	}
}

// get a single validator
func (k Keeper) GetValidator(ctx sdk.Context, addr sdk.ValAddress) (validator types.Validator, found bool) {
	store := ctx.KVStore(k.storeKey)
	value := store.Get(types.GetValidatorKey(addr))
	if value == nil {
		return validator, false
	}

	// If these amino encoded bytes are in the cache, return the cached validator
	strValue := string(value)
	if val, ok := k.validatorCache[strValue]; ok {
		valToReturn := val.val
		// Doesn't mutate the cache's value
		valToReturn.OperatorAddress = addr
		return valToReturn, true
	}

	// amino bytes weren't found in cache, so amino unmarshal and add it to the cache
	validator = types.MustUnmarshalValidator(k.cdc, value)
	cachedVal := newCachedValidator(validator, strValue)
	k.validatorCache[strValue] = newCachedValidator(validator, strValue)
	k.validatorCacheList.PushBack(cachedVal)

	// if the cache is too big, pop off the last element from it
	if k.validatorCacheList.Len() > aminoCacheSize {
		valToRemove := k.validatorCacheList.Remove(k.validatorCacheList.Front()).(cachedValidator)
		delete(k.validatorCache, valToRemove.marshalled)
	}

	validator = types.MustUnmarshalValidator(k.cdc, value)
	return validator, true
}

func (k Keeper) mustGetValidator(ctx sdk.Context, addr sdk.ValAddress) types.Validator {
	validator, found := k.GetValidator(ctx, addr)
	if !found {
		panic(fmt.Sprintf("validator record not found for address: %X\n", addr))
	}
	return validator
}

// get a single validator by consensus address
func (k Keeper) GetValidatorByConsAddr(ctx sdk.Context, consAddr sdk.ConsAddress) (validator types.Validator, found bool) {
	store := ctx.KVStore(k.storeKey)
	opAddr := store.Get(types.GetValidatorByConsAddrKey(consAddr))
	if opAddr == nil {
		return validator, false
	}
	return k.GetValidator(ctx, opAddr)
}

func (k Keeper) mustGetValidatorByConsAddr(ctx sdk.Context, consAddr sdk.ConsAddress) types.Validator {
	validator, found := k.GetValidatorByConsAddr(ctx, consAddr)
	if !found {
		panic(fmt.Errorf("validator with consensus-Address %s not found", consAddr))
	}
	return validator
}

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
func (k Keeper) SetValidatorStakingStateDelegation(ctx sdk.Context, valAddr sdk.ValAddress, delAddr sdk.AccAddress, delShares sdk.Dec) types.ValidatorStakingState {
	state := k.GetValidatorStakingState(ctx, valAddr)
	state = state.SetDelegator(valAddr, delAddr, delShares)
	k.SetValidatorStakingState(ctx, valAddr, state)

	return state
}

// RemoveValidatorStakingStateDelegation removes validator staking state delegation info.
func (k Keeper) RemoveValidatorStakingStateDelegation(ctx sdk.Context, valAddr sdk.ValAddress, delAddr sdk.AccAddress) types.ValidatorStakingState {
	state := k.GetValidatorStakingState(ctx, valAddr)
	state = state.RemoveDelegator(delAddr)
	k.SetValidatorStakingState(ctx, valAddr, state)

	return state
}

// set the main record holding validator details
func (k Keeper) SetValidator(ctx sdk.Context, validator types.Validator) {
	store := ctx.KVStore(k.storeKey)
	bz := types.MustMarshalValidator(k.cdc, validator)
	store.Set(types.GetValidatorKey(validator.OperatorAddress), bz)
}

// validator index
func (k Keeper) SetValidatorByConsAddr(ctx sdk.Context, validator types.Validator) {
	store := ctx.KVStore(k.storeKey)
	consAddr := sdk.ConsAddress(validator.ConsPubKey.Address())
	store.Set(types.GetValidatorByConsAddrKey(consAddr), validator.OperatorAddress)
}

// validator index
func (k Keeper) SetValidatorByPowerIndex(ctx sdk.Context, validator types.Validator) {
	// jailed validators are not kept in the power index
	if validator.Jailed {
		return
	}
	store := ctx.KVStore(k.storeKey)
	store.Set(types.GetValidatorsByPowerIndexKey(validator), validator.OperatorAddress)
}

// validator index
func (k Keeper) DeleteValidatorByPowerIndex(ctx sdk.Context, validator types.Validator) {
	store := ctx.KVStore(k.storeKey)
	store.Delete(types.GetValidatorsByPowerIndexKey(validator))
}

// validator index
func (k Keeper) SetNewValidatorByPowerIndex(ctx sdk.Context, validator types.Validator) {
	store := ctx.KVStore(k.storeKey)
	store.Set(types.GetValidatorsByPowerIndexKey(validator), validator.OperatorAddress)
}

// Update the tokens of an existing validator, update the validators power index key
func (k Keeper) AddValidatorTokensAndShares(ctx sdk.Context, validator types.Validator,
	tokensToAdd sdk.Int) (valOut types.Validator, addedShares sdk.Dec) {

	k.DeleteValidatorByPowerIndex(ctx, validator)
	validator, addedShares = validator.AddTokensFromDel(tokensToAdd)
	k.SetValidator(ctx, validator)
	k.SetValidatorByPowerIndex(ctx, validator)
	return validator, addedShares
}

// Update the tokens of an existing validator, update the validators power index key
func (k Keeper) RemoveValidatorTokensAndShares(ctx sdk.Context, validator types.Validator,
	sharesToRemove sdk.Dec) (valOut types.Validator, removedTokens sdk.Int) {

	k.DeleteValidatorByPowerIndex(ctx, validator)
	validator, removedTokens = validator.RemoveDelShares(sharesToRemove)
	k.SetValidator(ctx, validator)
	k.SetValidatorByPowerIndex(ctx, validator)
	return validator, removedTokens
}

// Update the tokens of an existing validator, update the validators power index key
func (k Keeper) RemoveValidatorTokens(ctx sdk.Context,
	validator types.Validator, tokensToRemove sdk.Int) types.Validator {

	k.DeleteValidatorByPowerIndex(ctx, validator)
	validator = validator.RemoveTokens(tokensToRemove)
	k.SetValidator(ctx, validator)
	k.SetValidatorByPowerIndex(ctx, validator)
	return validator
}

// UpdateValidatorCommission attempts to update a validator's commission rate.
// An error is returned if the new commission rate is invalid.
func (k Keeper) UpdateValidatorCommission(ctx sdk.Context,
	validator types.Validator, newRate sdk.Dec) (types.Commission, error) {

	commission := validator.Commission
	blockTime := ctx.BlockHeader().Time

	if err := commission.ValidateNewRate(newRate, blockTime); err != nil {
		return commission, err
	}

	commission.Rate = newRate
	commission.UpdateTime = blockTime

	return commission, nil
}

// remove the validator record and associated indexes
// except for the bonded validator index which is only handled in ApplyAndReturnTendermintUpdates
func (k Keeper) RemoveValidator(ctx sdk.Context, address sdk.ValAddress) {

	// first retrieve the old validator record
	validator, found := k.GetValidator(ctx, address)
	if !found {
		return
	}

	if !validator.IsUnbonded() {
		panic("cannot call RemoveValidator on bonded or unbonding validators")
	}
	if validator.Tokens.IsPositive() {
		panic("attempting to remove a validator which still contains tokens")
	}
	if validator.Tokens.IsPositive() {
		panic("validator being removed should never have positive tokens")
	}

	// delete the old validator record
	store := ctx.KVStore(k.storeKey)
	store.Delete(types.GetValidatorKey(address))
	store.Delete(types.GetValidatorByConsAddrKey(sdk.ConsAddress(validator.ConsPubKey.Address())))
	store.Delete(types.GetValidatorsByPowerIndexKey(validator))

	// call hooks
	k.AfterValidatorRemoved(ctx, validator.ConsAddress(), validator.OperatorAddress)
}

// get groups of validators

// get the set of all validators with no limits, used during genesis dump
func (k Keeper) GetAllValidators(ctx sdk.Context) (validators []types.Validator) {
	store := ctx.KVStore(k.storeKey)
	iterator := sdk.KVStorePrefixIterator(store, types.ValidatorsKey)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		validator := types.MustUnmarshalValidator(k.cdc, iterator.Value())
		validators = append(validators, validator)
	}
	return validators
}

// return a given amount of all the validators
func (k Keeper) GetValidators(ctx sdk.Context, maxRetrieve uint16) (validators []types.Validator) {
	store := ctx.KVStore(k.storeKey)
	validators = make([]types.Validator, maxRetrieve)

	iterator := sdk.KVStorePrefixIterator(store, types.ValidatorsKey)
	defer iterator.Close()

	i := 0
	for ; iterator.Valid() && i < int(maxRetrieve); iterator.Next() {
		validator := types.MustUnmarshalValidator(k.cdc, iterator.Value())
		validators[i] = validator
		i++
	}
	return validators[:i] // trim if the array length < maxRetrieve
}

// get the current group of bonded validators sorted by power-rank
func (k Keeper) GetBondedValidatorsByPower(ctx sdk.Context) []types.Validator {
	maxValidators := k.MaxValidators(ctx)
	validators := make([]types.Validator, maxValidators)

	iterator := k.ValidatorsPowerStoreIterator(ctx)
	defer iterator.Close()

	i := 0
	for ; iterator.Valid() && i < int(maxValidators); iterator.Next() {
		address := iterator.Value()
		validator := k.mustGetValidator(ctx, address)

		if validator.IsBonded() {
			validators[i] = validator
			i++
		}
	}
	return validators[:i] // trim
}

// returns an iterator for the current validator power store
func (k Keeper) ValidatorsPowerStoreIterator(ctx sdk.Context) sdk.Iterator {
	store := ctx.KVStore(k.storeKey)
	return sdk.KVStoreReversePrefixIterator(store, types.ValidatorsByPowerIndexKey)
}

//_______________________________________________________________________
// Last Validator Index

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

//_______________________________________________________________________
// Validator Queue

// gets a specific validator queue timeslice. A timeslice is a slice of ValAddresses corresponding to unbonding validators
// that expire at a certain time.
func (k Keeper) GetValidatorQueueTimeSlice(ctx sdk.Context, timestamp time.Time) (valAddrs []sdk.ValAddress) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.GetValidatorQueueTimeKey(timestamp))
	if bz == nil {
		return []sdk.ValAddress{}
	}
	k.cdc.MustUnmarshalBinaryLengthPrefixed(bz, &valAddrs)
	return valAddrs
}

// Sets a specific validator queue timeslice.
func (k Keeper) SetValidatorQueueTimeSlice(ctx sdk.Context, timestamp time.Time, keys []sdk.ValAddress) {
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshalBinaryLengthPrefixed(keys)
	store.Set(types.GetValidatorQueueTimeKey(timestamp), bz)
}

// Deletes a specific validator queue timeslice.
func (k Keeper) DeleteValidatorQueueTimeSlice(ctx sdk.Context, timestamp time.Time) {
	store := ctx.KVStore(k.storeKey)
	store.Delete(types.GetValidatorQueueTimeKey(timestamp))
}

// Insert an validator address to the appropriate timeslice in the validator queue
func (k Keeper) InsertValidatorQueue(ctx sdk.Context, val types.Validator) {
	timeSlice := k.GetValidatorQueueTimeSlice(ctx, val.UnbondingCompletionTime)
	timeSlice = append(timeSlice, val.OperatorAddress)
	k.SetValidatorQueueTimeSlice(ctx, val.UnbondingCompletionTime, timeSlice)
}

// Delete a validator address from the validator queue
func (k Keeper) DeleteValidatorQueue(ctx sdk.Context, val types.Validator) {
	timeSlice := k.GetValidatorQueueTimeSlice(ctx, val.UnbondingCompletionTime)
	newTimeSlice := []sdk.ValAddress{}
	for _, addr := range timeSlice {
		if !bytes.Equal(addr, val.OperatorAddress) {
			newTimeSlice = append(newTimeSlice, addr)
		}
	}
	if len(newTimeSlice) == 0 {
		k.DeleteValidatorQueueTimeSlice(ctx, val.UnbondingCompletionTime)
	} else {
		k.SetValidatorQueueTimeSlice(ctx, val.UnbondingCompletionTime, newTimeSlice)
	}
}

// Returns all the validator queue timeslices from time 0 until endTime
func (k Keeper) ValidatorQueueIterator(ctx sdk.Context, endTime time.Time) sdk.Iterator {
	store := ctx.KVStore(k.storeKey)
	return store.Iterator(types.ValidatorQueueKey, sdk.InclusiveEndBytes(types.GetValidatorQueueTimeKey(endTime)))
}

// Returns a concatenated list of all the timeslices before currTime, and deletes the timeslices from the queue
func (k Keeper) GetAllMatureValidatorQueue(ctx sdk.Context, currTime time.Time) (matureValsAddrs []sdk.ValAddress) {
	// gets an iterator for all timeslices from time 0 until the current Blockheader time
	validatorTimesliceIterator := k.ValidatorQueueIterator(ctx, ctx.BlockHeader().Time)
	defer validatorTimesliceIterator.Close()

	for ; validatorTimesliceIterator.Valid(); validatorTimesliceIterator.Next() {
		timeslice := []sdk.ValAddress{}
		k.cdc.MustUnmarshalBinaryLengthPrefixed(validatorTimesliceIterator.Value(), &timeslice)
		matureValsAddrs = append(matureValsAddrs, timeslice...)
	}

	return matureValsAddrs
}

// Unbonds all the unbonding validators that have finished their unbonding period
func (k Keeper) UnbondAllMatureValidatorQueue(ctx sdk.Context) {
	store := ctx.KVStore(k.storeKey)
	validatorTimesliceIterator := k.ValidatorQueueIterator(ctx, ctx.BlockHeader().Time)
	defer validatorTimesliceIterator.Close()

	for ; validatorTimesliceIterator.Valid(); validatorTimesliceIterator.Next() {
		timeslice := []sdk.ValAddress{}
		k.cdc.MustUnmarshalBinaryLengthPrefixed(validatorTimesliceIterator.Value(), &timeslice)

		for _, valAddr := range timeslice {
			val, found := k.GetValidator(ctx, valAddr)
			if !found {
				panic("validator in the unbonding queue was not found")
			}

			if !val.IsUnbonding() {
				panic("unexpected validator in unbonding queue; status was not unbonding")
			}
			val = k.unbondingToUnbonded(ctx, val)
			if val.GetDelegatorShares().IsZero() {
				k.RemoveValidator(ctx, val.OperatorAddress)
			}
		}

		store.Delete(validatorTimesliceIterator.Key())
	}
}

//_______________________________________________________________________
// Scheduled validator force unbond Queue

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
