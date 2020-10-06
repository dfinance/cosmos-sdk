package keeper

import (
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/staking/types"
)

// return a given amount of all the delegator unbonding-delegations
func (k Keeper) GetUnbondingDelegations(ctx sdk.Context, delegator sdk.AccAddress,
	maxRetrieve uint16) (unbondingDelegations []types.UnbondingDelegation) {

	unbondingDelegations = make([]types.UnbondingDelegation, maxRetrieve)

	store := ctx.KVStore(k.storeKey)
	delegatorPrefixKey := types.GetUBDsKey(delegator)
	iterator := sdk.KVStorePrefixIterator(store, delegatorPrefixKey)
	defer iterator.Close()

	i := 0
	for ; iterator.Valid() && i < int(maxRetrieve); iterator.Next() {
		unbondingDelegation := types.MustUnmarshalUBD(k.cdc, iterator.Value())
		unbondingDelegations[i] = unbondingDelegation
		i++
	}
	return unbondingDelegations[:i] // trim if the array length < maxRetrieve
}

// return a unbonding delegation
func (k Keeper) GetUnbondingDelegation(ctx sdk.Context,
	delAddr sdk.AccAddress, valAddr sdk.ValAddress) (ubd types.UnbondingDelegation, found bool) {

	store := ctx.KVStore(k.storeKey)
	key := types.GetUBDKey(delAddr, valAddr)
	value := store.Get(key)
	if value == nil {
		return ubd, false
	}

	ubd = types.MustUnmarshalUBD(k.cdc, value)
	return ubd, true
}

// return all unbonding delegations from a particular validator
func (k Keeper) GetUnbondingDelegationsFromValidator(ctx sdk.Context, valAddr sdk.ValAddress) (ubds []types.UnbondingDelegation) {
	store := ctx.KVStore(k.storeKey)
	iterator := sdk.KVStorePrefixIterator(store, types.GetUBDsByValIndexKey(valAddr))
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		key := types.GetUBDKeyFromValIndexKey(iterator.Key())
		value := store.Get(key)
		ubd := types.MustUnmarshalUBD(k.cdc, value)
		ubds = append(ubds, ubd)
	}
	return ubds
}

// iterate through all of the unbonding delegations
func (k Keeper) IterateUnbondingDelegations(ctx sdk.Context, fn func(index int64, ubd types.UnbondingDelegation) (stop bool)) {
	store := ctx.KVStore(k.storeKey)
	iterator := sdk.KVStorePrefixIterator(store, types.UnbondingDelegationKey)
	defer iterator.Close()

	for i := int64(0); iterator.Valid(); iterator.Next() {
		ubd := types.MustUnmarshalUBD(k.cdc, iterator.Value())
		if stop := fn(i, ubd); stop {
			break
		}
		i++
	}
}

// HasMaxUnbondingDelegationEntries - check if unbonding delegation has maximum number of entries
func (k Keeper) HasMaxUnbondingDelegationEntries(ctx sdk.Context,
	delegatorAddr sdk.AccAddress, validatorAddr sdk.ValAddress) bool {

	ubd, found := k.GetUnbondingDelegation(ctx, delegatorAddr, validatorAddr)
	if !found {
		return false
	}
	return len(ubd.Entries) >= int(k.MaxEntries(ctx))
}

// set the unbonding delegation and associated index
func (k Keeper) SetUnbondingDelegation(ctx sdk.Context, ubd types.UnbondingDelegation) {
	store := ctx.KVStore(k.storeKey)
	bz := types.MustMarshalUBD(k.cdc, ubd)
	key := types.GetUBDKey(ubd.DelegatorAddress, ubd.ValidatorAddress)
	store.Set(key, bz)
	store.Set(types.GetUBDByValIndexKey(ubd.DelegatorAddress, ubd.ValidatorAddress), []byte{}) // index, store empty bytes
}

// remove the unbonding delegation object and associated index
func (k Keeper) RemoveUnbondingDelegation(ctx sdk.Context, ubd types.UnbondingDelegation) {
	store := ctx.KVStore(k.storeKey)
	key := types.GetUBDKey(ubd.DelegatorAddress, ubd.ValidatorAddress)
	store.Delete(key)
	store.Delete(types.GetUBDByValIndexKey(ubd.DelegatorAddress, ubd.ValidatorAddress))
}

// SetUnbondingDelegationEntry adds an entry to the unbonding delegation at
// the given addresses. It creates the unbonding delegation if it does not exist
func (k Keeper) SetUnbondingDelegationEntry(ctx sdk.Context,
	delegatorAddr sdk.AccAddress, validatorAddr sdk.ValAddress,
	creationHeight int64, minTime time.Time,
	delOpType types.DelegationOpType, balance sdk.Int,
) types.UnbondingDelegation {

	ubd, found := k.GetUnbondingDelegation(ctx, delegatorAddr, validatorAddr)
	if found {
		ubd.AddEntry(creationHeight, minTime, delOpType, balance)
	} else {
		ubd = types.NewUnbondingDelegation(delegatorAddr, validatorAddr, creationHeight, minTime, delOpType, balance)
	}
	k.SetUnbondingDelegation(ctx, ubd)

	return ubd
}

// unbonding delegation queue timeslice operations

// gets a specific unbonding queue timeslice. A timeslice is a slice of DVPairs
// corresponding to unbonding delegations that expire at a certain time.
func (k Keeper) GetUBDQueueTimeSlice(ctx sdk.Context, timestamp time.Time) (dvPairs []types.DVPair) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.GetUnbondingDelegationTimeKey(timestamp))
	if bz == nil {
		return []types.DVPair{}
	}
	k.cdc.MustUnmarshalBinaryLengthPrefixed(bz, &dvPairs)
	return dvPairs
}

// Sets a specific unbonding queue timeslice.
func (k Keeper) SetUBDQueueTimeSlice(ctx sdk.Context, timestamp time.Time, keys []types.DVPair) {
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshalBinaryLengthPrefixed(keys)
	store.Set(types.GetUnbondingDelegationTimeKey(timestamp), bz)
}

// Insert an unbonding delegation to the appropriate timeslice in the unbonding queue
func (k Keeper) InsertUBDQueue(ctx sdk.Context, ubd types.UnbondingDelegation,
	completionTime time.Time) {

	timeSlice := k.GetUBDQueueTimeSlice(ctx, completionTime)
	dvPair := types.DVPair{DelegatorAddress: ubd.DelegatorAddress, ValidatorAddress: ubd.ValidatorAddress}
	if len(timeSlice) == 0 {
		k.SetUBDQueueTimeSlice(ctx, completionTime, []types.DVPair{dvPair})
	} else {
		timeSlice = append(timeSlice, dvPair)
		k.SetUBDQueueTimeSlice(ctx, completionTime, timeSlice)
	}
}

// Returns all the unbonding queue timeslices from time 0 until endTime
func (k Keeper) UBDQueueIterator(ctx sdk.Context, endTime time.Time) sdk.Iterator {
	store := ctx.KVStore(k.storeKey)
	return store.Iterator(types.UnbondingQueueKey,
		sdk.InclusiveEndBytes(types.GetUnbondingDelegationTimeKey(endTime)))
}

// Returns a concatenated list of all the timeslices inclusively previous to
// currTime, and deletes the timeslices from the queue
func (k Keeper) DequeueAllMatureUBDQueue(ctx sdk.Context, currTime time.Time) (matureUnbonds []types.DVPair) {
	store := ctx.KVStore(k.storeKey)
	// gets an iterator for all timeslices from time 0 until the current Blockheader time
	unbondingTimesliceIterator := k.UBDQueueIterator(ctx, currTime)
	defer unbondingTimesliceIterator.Close()

	for ; unbondingTimesliceIterator.Valid(); unbondingTimesliceIterator.Next() {
		timeslice := []types.DVPair{}
		value := unbondingTimesliceIterator.Value()
		k.cdc.MustUnmarshalBinaryLengthPrefixed(value, &timeslice)
		matureUnbonds = append(matureUnbonds, timeslice...)
		store.Delete(unbondingTimesliceIterator.Key())
	}
	return matureUnbonds
}

// unbond a particular delegation and perform associated store operations
func (k Keeper) unbond(
	ctx sdk.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress,
	delOpType types.DelegationOpType, shares sdk.Dec,
) (amount sdk.Int, err error) {

	// check if a delegation object exists in the store
	delegation, found := k.GetDelegation(ctx, delAddr, valAddr)
	if !found {
		return amount, types.ErrNoDelegatorForAddress
	}
	delegationShares := delegation.GetShares(delOpType)

	// call the before-delegation-modified hook
	k.BeforeDelegationSharesModified(ctx, delAddr, valAddr)

	// ensure that we have enough shares to remove
	if delegationShares.LT(shares) {
		return amount, sdkerrors.Wrap(types.ErrNotEnoughDelegationShares, delegationShares.String())
	}

	// get validator
	validator, found := k.GetValidator(ctx, valAddr)
	if !found {
		return amount, types.ErrNoValidatorFound
	}
	validatorTokens := validator.GetTokens(delOpType)

	// subtract shares from delegation
	delegation = delegation.SubShares(delOpType, shares)
	delegationShares = delegation.GetShares(delOpType)

	// if the delegation is the operator of the validator and undelegating will decrease the validator's
	// self delegation below their minimum, trigger a jail validator
	// only applicable for bonding tokens
	isValidatorOperator := delegation.DelegatorAddress.Equals(validator.OperatorAddress)
	if delOpType.IsBonding() && isValidatorOperator && !validator.Jailed &&
		validatorTokens.TokensFromShares(delegationShares).TruncateInt().LT(validator.MinSelfDelegation) {

		k.jailValidator(ctx, validator)
		validator = k.mustGetValidator(ctx, validator.OperatorAddress)
	}

	// remove the delegation
	if delegation.TotalShares().IsZero() {
		k.RemoveDelegation(ctx, delegation)
	} else {
		k.SetDelegation(ctx, delegation)
		// call the after delegation modification hook
		k.AfterDelegationModified(ctx, delegation.DelegatorAddress, delegation.ValidatorAddress)
	}

	// remove the shares and coins from the validator
	// NOTE that the amount is later (in keeper.Delegation) moved between staking module pools
	validator, amount = k.RemoveValidatorTokensAndShares(ctx, validator, delOpType, shares)

	// Update validator staking state
	var valStakingState types.ValidatorStakingState
	if delegation.TotalShares().IsZero() {
		valStakingState = k.RemoveValidatorStakingStateDelegation(ctx, validator.OperatorAddress, delegation.DelegatorAddress)
	} else {
		valStakingState = k.SetValidatorStakingStateDelegation(ctx,
			validator.OperatorAddress, delegation.DelegatorAddress,
			delegation.BondingShares, delegation.LPShares,
		)
	}

	// Check if max delegations overflow occurs after this delegation (primary by undelegating selfStake)
	valSelfStaked, valTotalStaked := valStakingState.GetSelfAndTotalStakes(validator)
	overflow, valStakeLimit := k.HasValidatorDelegationsOverflow(ctx, valSelfStaked, valTotalStaked)
	if overflow {
		if !validator.ScheduledToUnbond {
			k.scheduleValidatorForceUnbond(ctx, validator)
			k.Logger(ctx).Info(fmt.Sprintf(
				"Validator %s ScheduledUnbond status set due to undelegation/redelegation from %s: selfStaked / totalStaked / limit: %s / %s / %s",
				validator.OperatorAddress, delegation.DelegatorAddress, valSelfStaked, valTotalStaked, valStakeLimit,
			))
		}
	} else {
		// Overflow might be fixed by lowering current delegations amount
		if validator.ScheduledToUnbond {
			k.unscheduleValidatorForceUnbond(ctx, validator)
			k.Logger(ctx).Info(fmt.Sprintf(
				"Validator %s ScheduledUnbond status revoked due to undelegation/redelegation from %s: selfStaked / totalStaked / limit: %s / %s / %s",
				validator.OperatorAddress, delegation.DelegatorAddress, valSelfStaked, valTotalStaked, valStakeLimit,
			))
		}
	}

	if validator.Bonding.DelegatorShares.IsZero() && validator.IsUnbonded() {
		// if not unbonded, we must instead remove validator in EndBlocker once it finishes its unbonding period
		k.RemoveValidator(ctx, validator.OperatorAddress)
	}

	return amount, nil
}

// Undelegate unbonds an amount of delegator shares from a given validator. It
// will verify that the unbonding entries between the delegator and validator
// are not exceeded and unbond the staked tokens (based on shares) by creating
// an unbonding object and inserting it into the unbonding queue which will be
// processed during the staking EndBlocker.
// {ignoreUBLimit} ignores MaxUnbondingDelegationEntries and used for force undelegate operations.
func (k Keeper) Undelegate(
	ctx sdk.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress,
	delOpType types.DelegationOpType, sharesAmount sdk.Dec,
	ignoreUBLimit bool,
) (time.Time, error) {

	validator, found := k.GetValidator(ctx, valAddr)
	if !found {
		return time.Time{}, types.ErrNoDelegatorForAddress
	}

	if !ignoreUBLimit && k.HasMaxUnbondingDelegationEntries(ctx, delAddr, valAddr) {
		return time.Time{}, types.ErrMaxUnbondingDelegationEntries
	}

	returnAmount, err := k.unbond(ctx, delAddr, valAddr, delOpType, sharesAmount)
	if err != nil {
		return time.Time{}, err
	}

	// transfer the validator tokens to the not bonded pool
	if delOpType.IsBonding() && validator.IsBonded() {
		k.bondedTokensToNotBonded(ctx, returnAmount)
	}

	completionTime := ctx.BlockHeader().Time.Add(k.UnbondingTime(ctx))
	ubd := k.SetUnbondingDelegationEntry(ctx, delAddr, valAddr, ctx.BlockHeight(), completionTime, delOpType, returnAmount)
	k.InsertUBDQueue(ctx, ubd, completionTime)

	return completionTime, nil
}

// CompleteUnbondingWithAmount completes the unbonding of all mature entries in
// the retrieved unbonding delegation object and returns the total unbonding
// balance or an error upon failure.
func (k Keeper) CompleteUnbondingWithAmount(
	ctx sdk.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress, currTime time.Time,
) (sdk.Coins, error) {

	ubd, found := k.GetUnbondingDelegation(ctx, delAddr, valAddr)
	if !found {
		return nil, types.ErrNoUnbondingDelegation
	}

	bondDenom, lpDenom := k.BondDenom(ctx), k.LPDenom(ctx)
	balances := sdk.NewCoins()

	// loop through all the entries and complete unbonding mature entries
	for i := 0; i < len(ubd.Entries); i++ {
		entry := ubd.Entries[i]
		if entry.IsMature(currTime) {
			ubd.RemoveEntry(int64(i))
			i--

			// track undelegation only when remaining or truncated shares are non-zero
			if !entry.Balance.IsZero() {
				var senderName, denom string
				switch entry.OpType {
				case types.BondingDelOpType:
					senderName, denom = types.NotBondedPoolName, bondDenom
				case types.LiquidityDelOpType:
					senderName, denom = types.LiquidityPoolName, lpDenom
				default:
					panic(entry.OpType.Validate())
				}

				coin := sdk.NewCoin(denom, entry.Balance)
				err := k.supplyKeeper.UndelegateCoinsFromModuleToAccount(
					ctx, senderName, ubd.DelegatorAddress, sdk.NewCoins(coin),
				)
				if err != nil {
					return nil, err
				}

				balances = balances.Add(coin)
			}
		}
	}

	// set the unbonding delegation or remove it if there are no more entries
	if len(ubd.Entries) == 0 {
		k.RemoveUnbondingDelegation(ctx, ubd)
	} else {
		k.SetUnbondingDelegation(ctx, ubd)
	}

	return balances, nil
}

// CompleteUnbonding performs the same logic as CompleteUnbondingWithAmount except
// it does not return the total unbonding amount.
func (k Keeper) CompleteUnbonding(ctx sdk.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress) error {
	_, err := k.CompleteUnbondingWithAmount(ctx, delAddr, valAddr, ctx.BlockTime())
	return err
}
