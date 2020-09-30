package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/cosmos/cosmos-sdk/x/distribution/types"
	"github.com/cosmos/cosmos-sdk/x/staking/exported"
)

// initializeDelegation initializes starting info for a new delegation.
// DelegatorStartingInfo includes bonding and lp stakes.
func (k Keeper) initializeDelegation(ctx sdk.Context, val sdk.ValAddress, del sdk.AccAddress) {
	// period has already been incremented - we want to store the period ended by this delegation action
	previousPeriod := k.GetValidatorCurrentRewards(ctx, val).Period - 1

	// increment reference count for the period we're going to track
	k.incrementReferenceCount(ctx, val, previousPeriod)

	validator := k.stakingKeeper.Validator(ctx, val)
	delegation := k.stakingKeeper.Delegation(ctx, del, val)

	// calculate delegation stake in tokens
	// we don't store directly, so multiply delegation shares * (tokens per share)
	// note: necessary to truncate so we don't allow withdrawing more rewards than owed
	bondingStake := validator.BondingTokensFromSharesTruncated(delegation.GetBondingShares())
	lpStake := validator.LPTokensFromSharesTruncated(delegation.GetLPShares())
	k.SetDelegatorStartingInfo(ctx,
		val, del,
		types.NewDelegatorStartingInfo(previousPeriod, bondingStake, lpStake, uint64(ctx.BlockHeight())),
	)
}

// calculateDelegationRewardsBetween calculates the rewards accrued by a delegation between two periods.
// Cumulative reward rate difference is used to get rewards from a stake.
// RewardsBankPool balance is not included in the calculation.
func (k Keeper) calculateDelegationRewardsBetween(ctx sdk.Context, val exported.ValidatorI,
	startingPeriod, endingPeriod uint64,
	bondingStake, lpStake sdk.Dec,
) (bondingRewards, lpRewards sdk.DecCoins) {

	// sanity check
	if startingPeriod > endingPeriod {
		panic("startingPeriod cannot be greater than endingPeriod")
	}

	// sanity check
	if bondingStake.IsNegative() {
		panic("bonding stake should not be negative")
	}
	if lpStake.IsNegative() {
		panic("LP stake should not be negative")
	}

	// return staking * (ending - starting)
	starting := k.GetValidatorHistoricalRewards(ctx, val.GetOperator(), startingPeriod)
	ending := k.GetValidatorHistoricalRewards(ctx, val.GetOperator(), endingPeriod)

	bondingDifference := ending.CumulativeBondingRewardRatio.Sub(starting.CumulativeBondingRewardRatio)
	if bondingDifference.IsAnyNegative() {
		panic("negative bonding rewards should not be possible")
	}

	lpDifference := ending.CumulativeLPRewardRatio.Sub(starting.CumulativeLPRewardRatio)
	if lpDifference.IsAnyNegative() {
		panic("negative LP rewards should not be possible")
	}

	// note: necessary to truncate so we don't allow withdrawing more rewards than owed
	bondingRewards = bondingDifference.MulDecTruncate(bondingStake)
	lpRewards = lpDifference.MulDecTruncate(lpStake)

	return
}

// calculateDelegationRewards calculates the total rewards accrued by a delegation.
// Start period is taken from delegationStaringInfo.
// Delegator stake is reduced iterating over slash events.
// RewardsBankPool balance is not included in the calculation.
func (k Keeper) calculateDelegationRewards(ctx sdk.Context, val exported.ValidatorI, del exported.DelegationI,
	endingPeriod uint64,
) (bondingRewards, lpRewards sdk.DecCoins) {

	// fetch starting info for delegation
	startingInfo := k.GetDelegatorStartingInfo(ctx, del.GetValidatorAddr(), del.GetDelegatorAddr())

	if startingInfo.Height == uint64(ctx.BlockHeight()) {
		// started this height, no rewards yet
		return
	}

	startingPeriod := startingInfo.PreviousPeriod
	bondingStake := startingInfo.BondingStake
	lpStake := startingInfo.LPStake

	// Iterate through slashes and withdraw with calculated staking for
	// distribution periods. These period offsets are dependent on *when* slashes
	// happen - namely, in BeginBlock, after rewards are allocated...
	// Slashes which happened in the first block would have been before this
	// delegation existed, UNLESS they were slashes of a redelegation to this
	// validator which was itself slashed (from a fault committed by the
	// redelegation source validator) earlier in the same BeginBlock.
	startingHeight := startingInfo.Height
	// Slashes this block happened after reward allocation, but we have to account
	// for them for the stake sanity check below.
	// Slashing only affects bonding rewards.
	endingHeight := uint64(ctx.BlockHeight())
	if endingHeight > startingHeight {
		k.IterateValidatorSlashEventsBetween(ctx, del.GetValidatorAddr(), startingHeight, endingHeight,
			func(height uint64, event types.ValidatorSlashEvent) (stop bool) {
				endingPeriod := event.ValidatorPeriod
				if endingPeriod > startingPeriod {
					curBondingRewards, curLPRewards := k.calculateDelegationRewardsBetween(ctx, val, startingPeriod, endingPeriod, bondingStake, lpStake)

					bondingRewards = bondingRewards.Add(curBondingRewards...)
					lpRewards = lpRewards.Add(curLPRewards...)

					// Note: It is necessary to truncate so we don't allow withdrawing
					// more rewards than owed.
					bondingStake = bondingStake.MulTruncate(sdk.OneDec().Sub(event.Fraction))
					startingPeriod = endingPeriod
				}
				return false
			},
		)
	}

	// A total stake sanity check; Recalculated final stake should be less than or
	// equal to current stake here. We cannot use Equals because stake is truncated
	// when multiplied by slash fractions (see above). We could only use equals if
	// we had arbitrary-precision rationals.
	currentBondingStake := val.BondingTokensFromShares(del.GetBondingShares())
	if bondingStake.GT(currentBondingStake) {
		// Account for rounding inconsistencies between:
		//
		//     currentStake: calculated as in staking with a single computation
		//     stake:        calculated as an accumulation of stake
		//                   calculations across validator's distribution periods
		//
		// These inconsistencies are due to differing order of operations which
		// will inevitably have different accumulated rounding and may lead to
		// the smallest decimal place being one greater in stake than
		// currentStake. When we calculated slashing by period, even if we
		// round down for each slash fraction, it's possible due to how much is
		// being rounded that we slash less when slashing by period instead of
		// for when we slash without periods. In other words, the single slash,
		// and the slashing by period could both be rounding down but the
		// slashing by period is simply rounding down less, thus making stake >
		// currentStake
		//
		// A small amount of this error is tolerated and corrected for,
		// however any greater amount should be considered a breach in expected
		// behaviour.
		marginOfErr := sdk.SmallestDec().MulInt64(3)
		if bondingStake.LTE(currentBondingStake.Add(marginOfErr)) {
			bondingStake = currentBondingStake
		} else {
			panic(fmt.Sprintf("calculated final bonding stake for delegator %s greater than current stake"+
				"\n\tfinal stake:\t%s"+
				"\n\tcurrent stake:\t%s",
				del.GetDelegatorAddr(), bondingStake, currentBondingStake))
		}
	}

	// calculate rewards for final period
	curBondingRewards, curLPRewrds := k.calculateDelegationRewardsBetween(ctx, val, startingPeriod, endingPeriod, bondingStake, lpStake)
	bondingRewards = bondingRewards.Add(curBondingRewards...)
	lpRewards = lpRewards.Add(curLPRewrds...)

	return
}

// calculateDelegationTotalRewards sums current validator delegator rewards and stored RewardsBank coins.
func (k Keeper) calculateDelegationTotalRewards(ctx sdk.Context, val exported.ValidatorI, del exported.DelegationI, endingPeriod uint64) (rewards sdk.DecCoins) {
	// calculate rewards from the main pool
	curBondingDecCoins, curLPDecCoins := k.calculateDelegationRewards(ctx, val, del, endingPeriod)
	// get accumulated coins from the RewardsBankPool
	bankCoins := k.GetDelegatorRewardsBankCoins(ctx, del.GetDelegatorAddr())

	totalDecCoins := curBondingDecCoins.Add(curLPDecCoins...).Add(sdk.NewDecCoinsFromCoins(bankCoins...)...)

	return totalDecCoins
}

// withdrawDelegationRewards calculates delegator rewards and withdraws it from a validator.
// Also decreases outstanding rewards and removes delegationStartingInfo.
// Result coins must be transferred to module / user account afterwards.
func (k Keeper) withdrawDelegationRewards(ctx sdk.Context, val exported.ValidatorI, del exported.DelegationI) (sdk.Coins, error) {
	// check existence of delegator starting info
	if !k.HasDelegatorStartingInfo(ctx, del.GetValidatorAddr(), del.GetDelegatorAddr()) {
		return nil, types.ErrEmptyDelegationDistInfo
	}

	// end current period and calculate rewards
	endingPeriod := k.incrementValidatorPeriod(ctx, val)
	bondingRewardsRaw, lpRewardsRaw := k.calculateDelegationRewards(ctx, val, del, endingPeriod)
	rewardsRaw := bondingRewardsRaw.Add(lpRewardsRaw...)
	outstanding := k.GetValidatorOutstandingRewards(ctx, del.GetValidatorAddr())

	// defensive edge case may happen on the very final digits
	// of the decCoins due to operation order of the distribution mechanism.
	rewards := rewardsRaw.Intersect(outstanding)
	if !rewards.IsEqual(rewardsRaw) {
		logger := k.Logger(ctx)
		logger.Info(fmt.Sprintf("missing rewards rounding error, delegator %v"+
			"withdrawing rewards from validator %v, should have received %v, got %v",
			val.GetOperator(), del.GetDelegatorAddr(), rewardsRaw, rewards))
	}

	// truncate coins, return remainder to FoundationPool
	coins, remainder := rewards.TruncateDecimal()

	// update the outstanding rewards and the FoundationPool only if the transaction was successful
	k.SetValidatorOutstandingRewards(ctx, del.GetValidatorAddr(), outstanding.Sub(rewards))
	k.AppendToFoundationPool(ctx, remainder)

	// decrement reference count of starting period
	startingInfo := k.GetDelegatorStartingInfo(ctx, del.GetValidatorAddr(), del.GetDelegatorAddr())
	startingPeriod := startingInfo.PreviousPeriod
	k.decrementReferenceCount(ctx, del.GetValidatorAddr(), startingPeriod)

	// remove delegator starting info
	k.DeleteDelegatorStartingInfo(ctx, del.GetValidatorAddr(), del.GetDelegatorAddr())

	return coins, nil
}

// transferDelegationRewardsToRewardsBankPool transfers current validator delegator rewards to the RewardsBankPool.
func (k Keeper) transferDelegationRewardsToRewardsBankPool(ctx sdk.Context, val exported.ValidatorI, del exported.DelegationI) (sdk.Coins, error) {
	// withdraw from the main pool
	curCoins, err := k.withdrawDelegationRewards(ctx, val, del)
	if err != nil {
		return nil, fmt.Errorf("withdrawDelegationRewards: %w", err)
	}

	// add coins to RewardsBankPool module account and update RewardsBankPool delegator coins
	if !curCoins.IsZero() {
		bankCoins := k.GetDelegatorRewardsBankCoins(ctx, del.GetDelegatorAddr())
		bankCoins = bankCoins.Add(curCoins...)
		k.SetDelegatorRewardsBankCoins(ctx, del.GetDelegatorAddr(), bankCoins)

		err := k.supplyKeeper.SendCoinsFromModuleToModule(ctx, types.ModuleName, types.RewardsBankPoolName, curCoins)
		if err != nil {
			return nil, fmt.Errorf("supplyKeeper.SendCoinsFromModuleToModule: %w", err)
		}
	}

	return curCoins, nil
}

// transferDelegationTotalRewardsToAccount transfers sum of current validator delegator rewards
// and stored RewardsBank coins to user account.
func (k Keeper) transferDelegationTotalRewardsToAccount(ctx sdk.Context, val exported.ValidatorI, del exported.DelegationI) (sdk.Coins, error) {
	withdrawAddr := k.GetDelegatorWithdrawAddr(ctx, del.GetDelegatorAddr())

	// withdraw from the main pool
	curCoins, err := k.withdrawDelegationRewards(ctx, val, del)
	if err != nil {
		return nil, fmt.Errorf("withdrawDelegationRewards: %w", err)
	}

	// get accumulated coins from the RewardsBankPool
	bankCoins := k.GetDelegatorRewardsBankCoins(ctx, del.GetDelegatorAddr())
	k.DeleteDelegatorRewardsBankCoins(ctx, del.GetDelegatorAddr())

	// add coins to user account from the main pool
	if !curCoins.IsZero() {
		err := k.supplyKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, withdrawAddr, curCoins)
		if err != nil {
			return nil, fmt.Errorf("supplyKeeper.SendCoinsFromModuleToAccount (Distribution macc): %w", err)
		}
	}

	// add coins to user account from the RewardsBank pool
	if !bankCoins.IsZero() {
		err := k.supplyKeeper.SendCoinsFromModuleToAccount(ctx, types.RewardsBankPoolName, withdrawAddr, bankCoins)
		if err != nil {
			return nil, fmt.Errorf("supplyKeeper.SendCoinsFromModuleToAccount (RewardsBankPool macc): %w", err)
		}
	}

	totalCoins := curCoins.Add(bankCoins...)

	return totalCoins, nil
}
