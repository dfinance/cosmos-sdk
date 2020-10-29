package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/distribution/types"
	"github.com/cosmos/cosmos-sdk/x/staking/exported"
)

// PrepareForZeroHeight squashes current context state to fit zero-height (used on genesis export).
func (k Keeper) PrepareForZeroHeight(ctx sdk.Context) error {
	dels := k.stakingKeeper.GetAllSDKDelegations(ctx)

	// transfer all current rewards to the rewards bank
	// that makes all slash event and historical rewards ready to be deleted
	for _, del := range dels {
		val := k.stakingKeeper.Validator(ctx, del.ValidatorAddress)
		del := k.stakingKeeper.Delegation(ctx, del.DelegatorAddress, del.ValidatorAddress)
		if _, err := k.transferDelegationRewardsToRewardsBankPool(ctx, val, del); err != nil {
			return fmt.Errorf("transferring delegator %s rewards for validator %s to rewards bank pool: %w",
				del.GetDelegatorAddr(), val.GetOperator(), err)
		}
	}

	// clear validator slash events and historical rewards
	k.DeleteAllValidatorSlashEvents(ctx)
	k.DeleteAllValidatorHistoricalRewards(ctx)

	// partially reinitialize validators
	k.stakingKeeper.IterateValidators(ctx, func(_ int64, val exported.ValidatorI) (stop bool) {
		// set initial historical rewards (period 0) with reference count of 1
		k.SetValidatorHistoricalRewards(ctx, val.GetOperator(), 0, types.NewValidatorHistoricalRewards(sdk.DecCoins{}, sdk.DecCoins{}, 1))

		// update current rewards period (starting at period 1)
		curRewards := k.GetValidatorCurrentRewards(ctx, val.GetOperator())
		curRewards.Period = 1
		k.SetValidatorCurrentRewards(ctx, val.GetOperator(), curRewards)

		return false
	})

	// reinitialize all delegations (recreate DelegatorStartingInfo as they were deleted during the withdraw)
	for _, del := range dels {
		k.Hooks().BeforeDelegationCreated(ctx, del.DelegatorAddress, del.ValidatorAddress)
		k.Hooks().AfterDelegationModified(ctx, del.DelegatorAddress, del.ValidatorAddress)
	}

	// reset locked rewards state lock height
	k.IterateValidatorLockedRewards(ctx, func(valAddr sdk.ValAddress, info types.ValidatorLockedRewardsState) (stop bool) {
		info.LockHeight = 0
		k.SetValidatorLockedState(ctx, valAddr, info)
		return false
	})

	return nil
}
