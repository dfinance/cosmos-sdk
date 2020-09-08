package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/cosmos/cosmos-sdk/x/distribution/types"
	"github.com/cosmos/cosmos-sdk/x/staking/exported"
)

// GetDistributionPower calculates distribution power based on validator stakingPower and lockedRewards ratio.
func (k Keeper) GetDistributionPower(ctx sdk.Context, valAddr sdk.ValAddress, stakingPower int64) int64 {
	lockedRewards := k.GetValidatorLockedRewards(ctx, valAddr)
	if lockedRewards.LockedRatio.IsZero() {
		return stakingPower
	}

	lockedPower := sdk.NewDec(stakingPower).Mul(lockedRewards.LockedRatio)
	distrPower := stakingPower + lockedPower.TruncateInt64()

	return distrPower
}

// initializeValidator initializes rewards for a new validator.
func (k Keeper) initializeValidator(ctx sdk.Context, val exported.ValidatorI) {
	// set initial historical rewards (period 0) with reference count of 1
	k.SetValidatorHistoricalRewards(ctx, val.GetOperator(), 0, types.NewValidatorHistoricalRewards(sdk.DecCoins{}, 1))

	// set current rewards (starting at period 1)
	k.SetValidatorCurrentRewards(ctx, val.GetOperator(), types.NewValidatorCurrentRewards(sdk.DecCoins{}, 1))

	// set accumulated commission
	k.SetValidatorAccumulatedCommission(ctx, val.GetOperator(), types.InitialValidatorAccumulatedCommission())

	// set outstanding rewards
	k.SetValidatorOutstandingRewards(ctx, val.GetOperator(), sdk.DecCoins{})

	// set empty locked rewards info
	k.SetValidatorLockedRewards(ctx, val.GetOperator(), types.NewValidatorLockedRewards(sdk.ZeroDec()))
}

// incrementValidatorPeriod increments validator period, returning the period just ended.
// Current rewards are converted to cumulative reward ratio, added to historical rewards of the previous period
// and saved as historical rewards for the current period (which ends).
// New period starts with empty rewards.
func (k Keeper) incrementValidatorPeriod(ctx sdk.Context, val exported.ValidatorI) uint64 {
	// fetch current rewards
	rewards := k.GetValidatorCurrentRewards(ctx, val.GetOperator())

	// calculate current ratio
	var current sdk.DecCoins
	if val.GetTokens().IsZero() {
		// can't calculate ratio for zero-token validators
		// ergo we instead add to the FoundationPool
		k.AppendToFoundationPool(ctx, rewards.Rewards)

		outstanding := k.GetValidatorOutstandingRewards(ctx, val.GetOperator())
		outstanding = outstanding.Sub(rewards.Rewards)
		k.SetValidatorOutstandingRewards(ctx, val.GetOperator(), outstanding)

		current = sdk.DecCoins{}
	} else {
		// note: necessary to truncate so we don't allow withdrawing more rewards than owed
		current = rewards.Rewards.QuoDecTruncate(val.GetTokens().ToDec())
	}

	// fetch historical rewards for last period
	historical := k.GetValidatorHistoricalRewards(ctx, val.GetOperator(), rewards.Period-1).CumulativeRewardRatio

	// decrement reference count
	k.decrementReferenceCount(ctx, val.GetOperator(), rewards.Period-1)

	// set new historical rewards with reference count of 1
	k.SetValidatorHistoricalRewards(ctx, val.GetOperator(), rewards.Period, types.NewValidatorHistoricalRewards(historical.Add(current...), 1))

	// set current rewards, incrementing period by 1
	k.SetValidatorCurrentRewards(ctx, val.GetOperator(), types.NewValidatorCurrentRewards(sdk.DecCoins{}, rewards.Period+1))

	return rewards.Period
}

// incrementReferenceCount increments the reference count for a historical rewards value.
func (k Keeper) incrementReferenceCount(ctx sdk.Context, valAddr sdk.ValAddress, period uint64) {
	historical := k.GetValidatorHistoricalRewards(ctx, valAddr, period)
	if historical.ReferenceCount > 2 {
		panic("reference count should never exceed 2")
	}
	historical.ReferenceCount++
	k.SetValidatorHistoricalRewards(ctx, valAddr, period, historical)
}

// decrementReferenceCount decrements the reference count for a historical rewards value.
// Value is deleted if zero references remain.
func (k Keeper) decrementReferenceCount(ctx sdk.Context, valAddr sdk.ValAddress, period uint64) {
	historical := k.GetValidatorHistoricalRewards(ctx, valAddr, period)
	if historical.ReferenceCount == 0 {
		panic("cannot set negative reference count")
	}
	historical.ReferenceCount--
	if historical.ReferenceCount == 0 {
		k.DeleteValidatorHistoricalReward(ctx, valAddr, period)
	} else {
		k.SetValidatorHistoricalRewards(ctx, valAddr, period, historical)
	}
}

// updateValidatorSlashFraction handles a new slash event.
// This ends the current rewards period and adds a slash event for it.
func (k Keeper) updateValidatorSlashFraction(ctx sdk.Context, valAddr sdk.ValAddress, fraction sdk.Dec) {
	// sanity check
	if fraction.GT(sdk.OneDec()) || fraction.IsNegative() {
		panic(fmt.Sprintf("fraction must be >=0 and <=1, current fraction: %v", fraction))
	}

	val := k.stakingKeeper.Validator(ctx, valAddr)

	// increment current period
	newPeriod := k.incrementValidatorPeriod(ctx, val)

	// increment reference count on period we need to track
	k.incrementReferenceCount(ctx, valAddr, newPeriod)

	slashEvent := types.NewValidatorSlashEvent(newPeriod, fraction)
	height := uint64(ctx.BlockHeight())

	k.SetValidatorSlashEvent(ctx, valAddr, height, newPeriod, slashEvent)
}

// removeValidator removes rewards for a validator.
// Commission rewards are transferred to an operator.
// Commission remainder and outstanding rewards are transferred to FoundationPool.
func (k Keeper) removeValidator(ctx sdk.Context, valAddr sdk.ValAddress) {
	// fetch outstanding
	outstanding := k.GetValidatorOutstandingRewards(ctx, valAddr)

	// force-withdraw commission
	commission := k.GetValidatorAccumulatedCommission(ctx, valAddr)
	if !commission.IsZero() {
		// subtract from outstanding
		outstanding = outstanding.Sub(commission)

		// split into integral & remainder
		coins, remainder := commission.TruncateDecimal()

		// remainder to FoundationPool
		k.AppendToFoundationPool(ctx, remainder)

		// add to validator account
		if !coins.IsZero() {
			accAddr := sdk.AccAddress(valAddr)
			withdrawAddr := k.GetDelegatorWithdrawAddr(ctx, accAddr)
			err := k.supplyKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, withdrawAddr, coins)
			if err != nil {
				panic(err)
			}
		}
	}

	// add outstanding to FoundationPool
	k.AppendToFoundationPool(ctx, outstanding)

	// delete outstanding
	k.DeleteValidatorOutstandingRewards(ctx, valAddr)

	// remove commission record
	k.DeleteValidatorAccumulatedCommission(ctx, valAddr)

	// clear slashes
	k.DeleteValidatorSlashEvents(ctx, valAddr)

	// clear historical rewards
	k.DeleteValidatorHistoricalRewards(ctx, valAddr)

	// clear current rewards
	k.DeleteValidatorCurrentRewards(ctx, valAddr)

	// clear locked rewards info
	k.DeleteValidatorLockedRewards(ctx, valAddr)
}
