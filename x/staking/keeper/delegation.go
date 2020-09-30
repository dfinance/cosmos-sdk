package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/staking/types"
)

// return a specific delegation
func (k Keeper) GetDelegation(ctx sdk.Context,
	delAddr sdk.AccAddress, valAddr sdk.ValAddress) (
	delegation types.Delegation, found bool) {

	store := ctx.KVStore(k.storeKey)
	key := types.GetDelegationKey(delAddr, valAddr)
	value := store.Get(key)
	if value == nil {
		return delegation, false
	}

	delegation = types.MustUnmarshalDelegation(k.cdc, value)
	return delegation, true
}

// IterateAllDelegations iterate through all of the delegations
func (k Keeper) IterateAllDelegations(ctx sdk.Context, cb func(delegation types.Delegation) (stop bool)) {
	store := ctx.KVStore(k.storeKey)
	iterator := sdk.KVStorePrefixIterator(store, types.DelegationKey)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		delegation := types.MustUnmarshalDelegation(k.cdc, iterator.Value())
		if cb(delegation) {
			break
		}
	}
}

// GetAllDelegations returns all delegations used during genesis dump
func (k Keeper) GetAllDelegations(ctx sdk.Context) (delegations []types.Delegation) {
	k.IterateAllDelegations(ctx, func(delegation types.Delegation) bool {
		delegations = append(delegations, delegation)
		return false
	})
	return delegations
}

// return all delegations to a specific validator. Useful for querier.
func (k Keeper) GetValidatorDelegations(ctx sdk.Context, valAddr sdk.ValAddress) (delegations []types.Delegation) { //nolint:interfacer
	store := ctx.KVStore(k.storeKey)
	iterator := sdk.KVStorePrefixIterator(store, types.DelegationKey)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		delegation := types.MustUnmarshalDelegation(k.cdc, iterator.Value())
		if delegation.GetValidatorAddr().Equals(valAddr) {
			delegations = append(delegations, delegation)
		}
	}
	return delegations
}

// return a given amount of all the delegations from a delegator
func (k Keeper) GetDelegatorDelegations(ctx sdk.Context, delegator sdk.AccAddress,
	maxRetrieve uint16) (delegations []types.Delegation) {

	delegations = make([]types.Delegation, maxRetrieve)

	store := ctx.KVStore(k.storeKey)
	delegatorPrefixKey := types.GetDelegationsKey(delegator)
	iterator := sdk.KVStorePrefixIterator(store, delegatorPrefixKey)
	defer iterator.Close()

	i := 0
	for ; iterator.Valid() && i < int(maxRetrieve); iterator.Next() {
		delegation := types.MustUnmarshalDelegation(k.cdc, iterator.Value())
		delegations[i] = delegation
		i++
	}
	return delegations[:i] // trim if the array length < maxRetrieve
}

// HasValidatorDelegationsOverflow checks if current validator total staked coins are GT than the limit.
func (k Keeper) HasValidatorDelegationsOverflow(ctx sdk.Context, selfStaked, totalStaked sdk.Int) (overflow bool, maxDelegatedLimit sdk.Int) {
	// True for unbonding / unbonded validators
	if selfStaked.IsZero() {
		return
	}

	maxDelegationsRatio := k.MaxDelegationsRatio(ctx)
	maxDelegatedLimit = sdk.NewDecFromInt(selfStaked).Mul(maxDelegationsRatio).TruncateInt()
	if totalStaked.GT(maxDelegatedLimit) {
		overflow = true
	}

	return
}

// set a delegation
func (k Keeper) SetDelegation(ctx sdk.Context, delegation types.Delegation) {
	store := ctx.KVStore(k.storeKey)
	b := types.MustMarshalDelegation(k.cdc, delegation)
	store.Set(types.GetDelegationKey(delegation.DelegatorAddress, delegation.ValidatorAddress), b)
}

// remove a delegation
func (k Keeper) RemoveDelegation(ctx sdk.Context, delegation types.Delegation) {
	// TODO: Consider calling hooks outside of the store wrapper functions, it's unobvious.
	k.BeforeDelegationRemoved(ctx, delegation.DelegatorAddress, delegation.ValidatorAddress)
	store := ctx.KVStore(k.storeKey)
	store.Delete(types.GetDelegationKey(delegation.DelegatorAddress, delegation.ValidatorAddress))
}

// Perform a delegation, set/update everything necessary within the store.
// tokenSrc indicates the bond status of the incoming funds.
func (k Keeper) Delegate(
	ctx sdk.Context, delAddr sdk.AccAddress,
	delOpType types.DelegationOpType, bondAmt sdk.Int, tokenSrc sdk.BondStatus,
	validator types.Validator, subtractAccount bool,
) (newShares sdk.Dec, err error) {

	// In some situations, the exchange rate becomes invalid, e.g. if
	// Validator loses all tokens due to slashing. In this case,
	// make all future delegations invalid.
	if validator.InvalidExRate() {
		return sdk.ZeroDec(), types.ErrDelegatorShareExRateInvalid
	}

	// Get or create the delegation object
	delegation, found := k.GetDelegation(ctx, delAddr, validator.OperatorAddress)
	if !found {
		delegation = types.NewDelegation(delAddr, validator.OperatorAddress, sdk.ZeroDec(), sdk.ZeroDec())
	}

	// call the appropriate hook if present
	if found {
		k.BeforeDelegationSharesModified(ctx, delAddr, validator.OperatorAddress)
	} else {
		k.BeforeDelegationCreated(ctx, delAddr, validator.OperatorAddress)
	}

	// if subtractAccount is true then we are
	// performing a delegation and not a redelegation, thus the source tokens are
	// all non bonded
	if subtractAccount {
		if tokenSrc == sdk.Bonded {
			panic("delegation token source cannot be bonded")
		}

		var recipientName string
		var coinDenom string

		if delOpType == types.BondingDelOpType {
			// bonding tokens
			coinDenom = k.BondDenom(ctx)
			switch {
			case validator.IsBonded():
				recipientName = types.BondedPoolName
			case validator.IsUnbonding(), validator.IsUnbonded():
				recipientName = types.NotBondedPoolName
			default:
				panic("invalid validator status")
			}
		} else {
			// liquidity tokens
			coinDenom = k.LPDenom(ctx)
			recipientName = types.LiquidityPoolName
		}

		coins := sdk.NewCoins(sdk.NewCoin(coinDenom, bondAmt))
		err := k.supplyKeeper.DelegateCoinsFromAccountToModule(ctx, delegation.DelegatorAddress, recipientName, coins)
		if err != nil {
			return sdk.Dec{}, err
		}
	} else if delOpType == types.BondingDelOpType {
		// potentially transfer tokens between pools
		// for bonding tokens only, as liquidity tokens have only one pool
		switch {
		case tokenSrc == sdk.Bonded && validator.IsBonded():
			// do nothing
		case (tokenSrc == sdk.Unbonded || tokenSrc == sdk.Unbonding) && !validator.IsBonded():
			// do nothing
		case (tokenSrc == sdk.Unbonded || tokenSrc == sdk.Unbonding) && validator.IsBonded():
			// transfer pools
			k.notBondedTokensToBonded(ctx, bondAmt)
		case tokenSrc == sdk.Bonded && !validator.IsBonded():
			// transfer pools
			k.bondedTokensToNotBonded(ctx, bondAmt)
		default:
			panic("unknown token source bond status")
		}
	}

	validator, newShares = k.AddValidatorTokensAndShares(ctx, validator, delOpType, bondAmt)

	// Update delegation
	delegation = delegation.AddShares(delOpType, newShares)
	k.SetDelegation(ctx, delegation)

	// Update validator staking state
	valStakingState := k.SetValidatorStakingStateDelegation(ctx,
		validator.OperatorAddress, delegation.DelegatorAddress,
		delegation.BondingShares, delegation.LPShares,
	)

	// Check if max delegations overflow occurs after this delegation
	valSelfStaked, valTotalStaked := valStakingState.GetSelfAndTotalStakes(validator)
	overflow, valStakeLimit := k.HasValidatorDelegationsOverflow(ctx, valSelfStaked, valTotalStaked)
	if overflow {
		return sdk.Dec{}, sdkerrors.Wrapf(
			types.ErrMaxDelegationsLimit,
			"current tokens limit for %s: %s",
			validator.OperatorAddress, valStakeLimit,
		)
	} else if validator.ScheduledToUnbond {
		// If overflow is fixed by this delegation, undo
		k.unscheduleValidatorForceUnbond(ctx, validator)
		k.Logger(ctx).Info(fmt.Sprintf(
			"Validator %s ScheduledUnbond status revoked due to delegation from %s: selfStaked / totalStaked / limit: %s / %s / %s",
			validator.OperatorAddress, delegation.DelegatorAddress, valSelfStaked, valTotalStaked, valStakeLimit,
		))
	}

	// Call the after-modification hook
	k.AfterDelegationModified(ctx, delegation.DelegatorAddress, delegation.ValidatorAddress)

	return newShares, nil
}

// ValidateUnbondAmount validates that a given unbond or redelegation amount is
// valied based on upon the converted shares. If the amount is valid, the total
// amount of respective shares is returned, otherwise an error is returned.
func (k Keeper) ValidateUnbondAmount(
	ctx sdk.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress,
	delOpType types.DelegationOpType, amount sdk.Int,
) (shares sdk.Dec, err error) {

	validator, found := k.GetValidator(ctx, valAddr)
	if !found {
		return shares, types.ErrNoValidatorFound
	}
	valTokens := validator.GetTokens(delOpType)

	del, found := k.GetDelegation(ctx, delAddr, valAddr)
	if !found {
		return shares, types.ErrNoDelegation
	}

	shares, err = valTokens.SharesFromTokens(amount)
	if err != nil {
		return shares, err
	}

	sharesTruncated, err := valTokens.SharesFromTokensTruncated(amount)
	if err != nil {
		return shares, err
	}

	delShares := del.GetShares(delOpType)
	if sharesTruncated.GT(delShares) {
		return shares, types.ErrBadSharesAmount
	}

	// Cap the shares at the delegation's shares. Shares being greater could occur
	// due to rounding, however we don't want to truncate the shares or take the
	// minimum because we want to allow for the full withdraw of shares from a
	// delegation.
	if shares.GT(delShares) {
		shares = delShares
	}

	return shares, nil
}

// ForceRemoveDelegator removes all delegations and redelegations for the specified delegator.
// That action also updates corresponding unbonding queues.
func (k Keeper) ForceRemoveDelegator(ctx sdk.Context, delAddr sdk.AccAddress) error {
	// complete all redelegations
	if err := k.forceStopAllRedelegations(ctx, delAddr); err != nil {
		return err
	}

	// undelegate all delegations ignoring the limit
	// that would add undelegation to the UBQueue
	for _, delegation := range k.GetAllDelegatorDelegations(ctx, delAddr) {
		if delegation.BondingShares.IsPositive() {
			_, err := k.Undelegate(ctx, delAddr, delegation.ValidatorAddress,
				types.BondingDelOpType, delegation.BondingShares, true,
			)
			if err != nil {
				return fmt.Errorf("undelegating delegation BondingShares for validator %s: %w", delegation.ValidatorAddress, err)
			}
		}
		if delegation.LPShares.IsPositive() {
			_, err := k.Undelegate(ctx, delAddr, delegation.ValidatorAddress,
				types.LiquidityDelOpType, delegation.LPShares, true,
			)
			if err != nil {
				return fmt.Errorf("undelegating delegation LPShares for validator %s: %w", delegation.ValidatorAddress, err)
			}
		}
	}

	// complete all undelegations
	if err := k.forceStopAllUnbondingDelegations(ctx, delAddr); err != nil {
		return err
	}

	return nil
}

// forceStopAllRedelegations modifies the RedelegationQueue removing all
// scheduled redelegation completions and completes them.
func (k Keeper) forceStopAllRedelegations(ctx sdk.Context, delAddr sdk.AccAddress) error {
	store := ctx.KVStore(k.storeKey)
	fakeCurrTime := ctx.BlockTime().Add(types.MaxUnbondingTime)
	rdTriplets := make([]types.DVVTriplet, 0)

	// we fake the endTime for RedelegationQueue to get all redelegation triplets
	iterator := k.RedelegationQueueIterator(ctx, fakeCurrTime)
	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		rcvTimeSlice := []types.DVVTriplet{}
		value := iterator.Value()
		k.cdc.MustUnmarshalBinaryLengthPrefixed(value, &rcvTimeSlice)

		// filter out redelegations for the specific delegator
		updTimeSlice := make([]types.DVVTriplet, 0, len(rcvTimeSlice))
		for _, triplet := range rcvTimeSlice {
			if triplet.DelegatorAddress.Equals(delAddr) {
				// add only if not duplicated
				found := false
				for _, addedTriplet := range rdTriplets {
					if addedTriplet.Equal(triplet) {
						found = true
						break
					}
				}
				if !found {
					rdTriplets = append(rdTriplets, triplet)
				}

				continue
			}

			updTimeSlice = append(updTimeSlice, triplet)
		}

		// update / remove timeSlice
		if len(updTimeSlice) != len(rcvTimeSlice) {
			if len(updTimeSlice) == 0 {
				store.Delete(iterator.Key())
			} else {
				store.Set(iterator.Key(), k.cdc.MustMarshalBinaryLengthPrefixed(updTimeSlice))
			}
		}
	}

	// complete redelegations
	for _, triplet := range rdTriplets {
		// we fake curTime for redelegation to be "mature"
		balances, err := k.CompleteRedelegationWithAmount(ctx, triplet.DelegatorAddress, triplet.ValidatorSrcAddress, triplet.ValidatorDstAddress, fakeCurrTime)
		if err != nil {
			return fmt.Errorf("completing redelegation for srcValidator %s and dstValidator %s: %w", triplet.ValidatorSrcAddress, triplet.ValidatorDstAddress, err)
		}

		ctx.EventManager().EmitEvent(
			sdk.NewEvent(
				types.EventTypeCompleteRedelegation,
				sdk.NewAttribute(sdk.AttributeKeyAmount, balances.String()),
				sdk.NewAttribute(types.AttributeKeyDelegator, triplet.DelegatorAddress.String()),
				sdk.NewAttribute(types.AttributeKeySrcValidator, triplet.ValidatorSrcAddress.String()),
				sdk.NewAttribute(types.AttributeKeyDstValidator, triplet.ValidatorDstAddress.String()),
			),
		)
	}

	return nil
}

// forceStopAllUnbondingDelegations modifies the UBQueue removing all
// scheduled undelegation completions and completes them.
func (k Keeper) forceStopAllUnbondingDelegations(ctx sdk.Context, delAddr sdk.AccAddress) error {
	store := ctx.KVStore(k.storeKey)
	fakeCurrTime := ctx.BlockTime().Add(types.MaxUnbondingTime)
	ubPairs := make([]types.DVPair, 0)

	// we fake the endTime for UBQueue to get all undelegation pairs
	iterator := k.UBDQueueIterator(ctx, fakeCurrTime)
	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		rcvTimeSlice := []types.DVPair{}
		value := iterator.Value()
		k.cdc.MustUnmarshalBinaryLengthPrefixed(value, &rcvTimeSlice)

		// filter out undelegations for the specific delegator
		updTimeSlice := make([]types.DVPair, 0, len(rcvTimeSlice))
		for _, pair := range rcvTimeSlice {
			if pair.DelegatorAddress.Equals(delAddr) {
				// add only if not duplicated
				found := false
				for _, addedPair := range ubPairs {
					if addedPair.Equal(pair) {
						found = true
						break
					}
				}
				if !found {
					ubPairs = append(ubPairs, pair)
				}

				continue
			}

			updTimeSlice = append(updTimeSlice, pair)
		}

		// update / remove timeSlice
		if len(updTimeSlice) != len(rcvTimeSlice) {
			if len(updTimeSlice) == 0 {
				store.Delete(iterator.Key())
			} else {
				store.Set(iterator.Key(), k.cdc.MustMarshalBinaryLengthPrefixed(updTimeSlice))
			}
		}
	}

	// complete unbondings
	for _, pair := range ubPairs {
		// we fake curTime for undelegation to be "mature"
		balances, err := k.CompleteUnbondingWithAmount(ctx, pair.DelegatorAddress, pair.ValidatorAddress, fakeCurrTime)
		if err != nil {
			return fmt.Errorf("completing unbonding delegation for validator %s: %w", pair.ValidatorAddress, err)
		}

		ctx.EventManager().EmitEvent(
			sdk.NewEvent(
				types.EventTypeCompleteUnbonding,
				sdk.NewAttribute(sdk.AttributeKeyAmount, balances.String()),
				sdk.NewAttribute(types.AttributeKeyValidator, pair.ValidatorAddress.String()),
				sdk.NewAttribute(types.AttributeKeyDelegator, pair.DelegatorAddress.String()),
			),
		)
	}

	return nil
}
