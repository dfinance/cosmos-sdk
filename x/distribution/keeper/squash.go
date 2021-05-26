package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/distribution/types"
)

type (
	// Operations order:
	//   1: slashOperations
	//   2: decCoinOperation
	//   3: main squash ops
	//   4: rewardOperations
	SquashOptions struct {
		// Slash event operations
		slashOps slashOperations
		// DecCoins / Coins operations (all rewards and pools)
		decCoinsOps []decCoinOperation
		// Reward operations (after the main squash)
		rewardOps rewardOperations
	}

	slashOperations struct {
		// Remove all slash events
		RemoveAll bool
	}

	decCoinOperation struct {
		// Coin denom
		Denom string
		// Remove coin balance
		// 1st priority
		Remove bool
		// Rename coin / move balance (empty - no renaming)
		// 2nd priority
		RenameTo string
	}

	rewardOperations struct {
		// Transfer all banked rewards to account
		BankTransfer bool
		// Transfer all validator commissions to operator
		CommissionTransfer bool
	}
)

func (opts *SquashOptions) SetSlashOp(removeAll bool) error {
	op := slashOperations{
		RemoveAll: removeAll,
	}
	opts.slashOps = op

	return nil
}

func (opts *SquashOptions) SetDecCoinOp(denomRaw string, remove bool, renameToRaw string) error {
	op := decCoinOperation{}
	op.Remove = remove

	if remove && renameToRaw != "" {
		return fmt.Errorf("remove op can not coexist with rename op")
	}

	if err := sdk.ValidateDenom(denomRaw); err != nil {
		return fmt.Errorf("denom (%s): invalid: %w", denomRaw, err)
	}
	op.Denom = denomRaw

	if renameToRaw != "" {
		if err := sdk.ValidateDenom(renameToRaw); err != nil {
			return fmt.Errorf("renameTo denom (%s): invalid: %w", renameToRaw, err)
		}
		op.RenameTo = renameToRaw
	}

	opts.decCoinsOps = append(opts.decCoinsOps, op)

	return nil
}

func (opts *SquashOptions) SetRewardOps(bankTransfer, commissionTransfer bool) error {
	opts.rewardOps.BankTransfer = bankTransfer
	opts.rewardOps.CommissionTransfer = commissionTransfer

	return nil
}

func NewEmptySquashOptions() SquashOptions {
	return SquashOptions{
		decCoinsOps: nil,
	}
}

// PrepareForZeroHeight squashes current context state to fit zero-height (used on genesis export).
func (k Keeper) PrepareForZeroHeight(ctx sdk.Context, opts SquashOptions) error {
	// slashOps
	{
		if opts.slashOps.RemoveAll {
			// TODO: implement
			// we can't remove all slash event as it will corrupt currentStake calculation (calculateDelegationRewards func)
			// rewriting history is a hell of a task
		}
	}

	// decCoinsOps
	{
		// remove ops
		removeCoin := func(denom string, coins sdk.Coins) sdk.Coins {
			coinToDel := sdk.NewCoin(denom, sdk.ZeroInt())
			for _, coin := range coins {
				if coin.Denom != denom {
					continue
				}
				coinToDel.Amount = coin.Amount
				break
			}
			coins = coins.Sub(sdk.NewCoins(coinToDel))
			return coins
		}
		removeDecCoin := func(denom string, coins sdk.DecCoins) sdk.DecCoins {
			coinToDel := sdk.NewDecCoin(denom, sdk.ZeroInt())
			for _, coin := range coins {
				if coin.Denom != denom {
					continue
				}
				coinToDel.Amount = coin.Amount
				break
			}
			coins = coins.Sub(sdk.NewDecCoins(coinToDel))
			return coins
		}
		for _, op := range opts.decCoinsOps {
			if !op.Remove {
				continue
			}

			k.IterateValidatorAccumulatedCommissions(ctx, func(val sdk.ValAddress, commission types.ValidatorAccumulatedCommission) (stop bool) {
				coins := removeDecCoin(op.Denom, commission)
				k.SetValidatorAccumulatedCommission(ctx, val, types.ValidatorAccumulatedCommission(coins))
				return false
			})
			k.IterateValidatorOutstandingRewards(ctx, func(val sdk.ValAddress, rewards types.ValidatorOutstandingRewards) (stop bool) {
				coins := removeDecCoin(op.Denom, rewards)
				k.SetValidatorOutstandingRewards(ctx, val, types.ValidatorOutstandingRewards(coins))
				return false
			})
			k.IterateValidatorCurrentRewards(ctx, func(val sdk.ValAddress, rewards types.ValidatorCurrentRewards) (stop bool) {
				rewards.BondingRewards, rewards.LPRewards = removeDecCoin(op.Denom, rewards.BondingRewards), removeDecCoin(op.Denom, rewards.LPRewards)
				k.SetValidatorCurrentRewards(ctx, val, rewards)
				return false
			})
			k.IterateDelegatorRewardsBankCoins(ctx, func(delAddr sdk.AccAddress, bankCoins sdk.Coins) (stop bool) {
				coins := removeCoin(op.Denom, bankCoins)
				k.SetDelegatorRewardsBankCoins(ctx, delAddr, coins)
				return false
			})
			k.IterateValidatorHistoricalRewards(ctx, func(val sdk.ValAddress, period uint64, rewards types.ValidatorHistoricalRewards) (stop bool) {
				rewards.CumulativeBondingRewardRatio, rewards.CumulativeLPRewardRatio = removeDecCoin(op.Denom, rewards.CumulativeBondingRewardRatio), removeDecCoin(op.Denom, rewards.CumulativeLPRewardRatio)
				k.SetValidatorHistoricalRewards(ctx, val, period, rewards)
				return false
			})

			rewardPools := k.GetRewardPools(ctx)
			rewardPools.LiquidityProvidersPool = removeDecCoin(op.Denom, rewardPools.LiquidityProvidersPool)
			rewardPools.PublicTreasuryPool = removeDecCoin(op.Denom, rewardPools.PublicTreasuryPool)
			rewardPools.HARP = removeDecCoin(op.Denom, rewardPools.HARP)
			rewardPools.FoundationPool = removeDecCoin(op.Denom, rewardPools.FoundationPool)
			k.SetRewardPools(ctx, rewardPools)
		}

		// rename ops
		renameCoin := func(oldDenom, newDenom string, coins sdk.Coins) sdk.Coins {
			oldCoin := sdk.NewCoin(oldDenom, sdk.ZeroInt())
			for _, coin := range coins {
				if coin.Denom != oldDenom {
					continue
				}
				oldCoin.Amount = coin.Amount
				break
			}
			newCoin := sdk.NewCoin(newDenom, oldCoin.Amount)

			coins = coins.Sub(sdk.NewCoins(oldCoin))
			coins = coins.Add(newCoin)
			return coins
		}
		renameDecCoin := func(oldDenom, newDenom string, coins sdk.DecCoins) sdk.DecCoins {
			oldCoin := sdk.NewDecCoin(oldDenom, sdk.ZeroInt())
			for _, coin := range coins {
				if coin.Denom != oldDenom {
					continue
				}
				oldCoin.Amount = coin.Amount
				break
			}
			newCoin := oldCoin
			newCoin.Denom = newDenom

			coins = coins.Sub(sdk.NewDecCoins(oldCoin))
			coins = coins.Add(newCoin)
			return coins
		}
		for _, op := range opts.decCoinsOps {
			if op.RenameTo == "" {
				continue
			}

			k.IterateValidatorAccumulatedCommissions(ctx, func(val sdk.ValAddress, commission types.ValidatorAccumulatedCommission) (stop bool) {
				coins := renameDecCoin(op.Denom, op.RenameTo, commission)
				k.SetValidatorAccumulatedCommission(ctx, val, types.ValidatorAccumulatedCommission(coins))
				return false
			})
			k.IterateValidatorOutstandingRewards(ctx, func(val sdk.ValAddress, rewards types.ValidatorOutstandingRewards) (stop bool) {
				coins := renameDecCoin(op.Denom, op.RenameTo, rewards)
				k.SetValidatorOutstandingRewards(ctx, val, types.ValidatorOutstandingRewards(coins))
				return false
			})
			k.IterateValidatorCurrentRewards(ctx, func(val sdk.ValAddress, rewards types.ValidatorCurrentRewards) (stop bool) {
				rewards.BondingRewards, rewards.LPRewards = renameDecCoin(op.Denom, op.RenameTo, rewards.BondingRewards), renameDecCoin(op.Denom, op.RenameTo, rewards.LPRewards)
				k.SetValidatorCurrentRewards(ctx, val, rewards)
				return false
			})
			k.IterateDelegatorRewardsBankCoins(ctx, func(delAddr sdk.AccAddress, bankCoins sdk.Coins) (stop bool) {
				coins := renameCoin(op.Denom, op.RenameTo, bankCoins)
				k.SetDelegatorRewardsBankCoins(ctx, delAddr, coins)
				return false
			})
			k.IterateValidatorHistoricalRewards(ctx, func(val sdk.ValAddress, period uint64, rewards types.ValidatorHistoricalRewards) (stop bool) {
				rewards.CumulativeBondingRewardRatio, rewards.CumulativeLPRewardRatio = renameDecCoin(op.Denom, op.RenameTo, rewards.CumulativeBondingRewardRatio), renameDecCoin(op.Denom, op.RenameTo, rewards.CumulativeLPRewardRatio)
				k.SetValidatorHistoricalRewards(ctx, val, period, rewards)
				return false
			})

			rewardPools := k.GetRewardPools(ctx)
			rewardPools.LiquidityProvidersPool = renameDecCoin(op.Denom, op.RenameTo, rewardPools.LiquidityProvidersPool)
			rewardPools.PublicTreasuryPool = renameDecCoin(op.Denom, op.RenameTo, rewardPools.PublicTreasuryPool)
			rewardPools.HARP = renameDecCoin(op.Denom, op.RenameTo, rewardPools.HARP)
			rewardPools.FoundationPool = renameDecCoin(op.Denom, op.RenameTo, rewardPools.FoundationPool)
			k.SetRewardPools(ctx, rewardPools)
		}
	}

	// main squash operation
	//{
	//	dels := k.stakingKeeper.GetAllSDKDelegations(ctx)
	//
	//	// transfer all current rewards to the rewards bank
	//	// that makes all slash event and historical rewards ready to be deleted
	//	for _, del := range dels {
	//		val := k.stakingKeeper.Validator(ctx, del.ValidatorAddress)
	//		if _, err := k.transferDelegationRewardsToRewardsBankPool(ctx, val, del); err != nil {
	//			return fmt.Errorf("transferring delegator %s rewards for validator %s to rewards bank pool: %w",
	//				del.GetDelegatorAddr(), val.GetOperator(), err)
	//		}
	//	}
	//
	//	// clear validator slash events and historical rewards
	//	k.DeleteAllValidatorSlashEvents(ctx)
	//	k.DeleteAllValidatorHistoricalRewards(ctx)
	//
	//	// partially reinitialize validators
	//	k.stakingKeeper.IterateValidators(ctx, func(_ int64, val exported.ValidatorI) (stop bool) {
	//		// set initial historical rewards (period 0) with reference count of 1
	//		k.SetValidatorHistoricalRewards(ctx, val.GetOperator(), 0, types.NewValidatorHistoricalRewards(sdk.DecCoins{}, sdk.DecCoins{}, 1))
	//
	//		// update current rewards period (starting at period 1)
	//		curRewards := k.GetValidatorCurrentRewards(ctx, val.GetOperator())
	//		curRewards.Period = 1
	//		k.SetValidatorCurrentRewards(ctx, val.GetOperator(), curRewards)
	//
	//		return false
	//	})
	//
	//	// reinitialize all delegations (recreate DelegatorStartingInfo as they were deleted during the withdraw)
	//	for _, del := range dels {
	//		k.Hooks().BeforeDelegationCreated(ctx, del.DelegatorAddress, del.ValidatorAddress)
	//		k.Hooks().AfterDelegationModified(ctx, del.DelegatorAddress, del.ValidatorAddress)
	//	}
	//
	//	// reset locked rewards state lock height
	//	k.IterateValidatorLockedRewards(ctx, func(valAddr sdk.ValAddress, info types.ValidatorLockedRewardsState) (stop bool) {
	//		info.LockHeight = 0
	//		k.SetValidatorLockedState(ctx, valAddr, info)
	//		return false
	//	})
	//}

	// after main squash operations
	{
		// Transfer all current (and banked) rewards to accounts
		if opts.rewardOps.BankTransfer {
			// Registered delegations (with StartingInfo)
			{
				var opErr error
				k.IterateDelegatorStartingInfos(ctx, func(val sdk.ValAddress, del sdk.AccAddress, info types.DelegatorStartingInfo) (stop bool) {
					if _, err := k.WithdrawDelegationRewards(ctx, del, val); err != nil {
						opErr = fmt.Errorf("transferring delegator %s rewards for validator %s to account: %w",
							del, val, err,
						)
						return true
					}

					return false
				})
				if opErr != nil {
					return opErr
				}
			}

			// Banked leftovers
			{
				var opErr error
				k.IterateDelegatorRewardsBankCoins(ctx, func(delAddr sdk.AccAddress, coins sdk.Coins) (stop bool) {
					withdrawAddr := k.GetDelegatorWithdrawAddr(ctx, delAddr)

					if err := k.supplyKeeper.SendCoinsFromModuleToAccount(ctx, types.RewardsBankPoolName, withdrawAddr, coins); err != nil {
						opErr = fmt.Errorf("transferring delegator %s rewards from rewardsBank: %w", delAddr, err)
						return true
					}
					k.DeleteDelegatorRewardsBankCoins(ctx, delAddr)

					return false
				})
				if opErr != nil {
					return opErr
				}
			}
		}

		// Transfer all validator commissions to operators
		if opts.rewardOps.CommissionTransfer {
			valAddrs := make([]sdk.ValAddress, 0)
			k.IterateValidatorAccumulatedCommissions(ctx, func(val sdk.ValAddress, value types.ValidatorAccumulatedCommission) (stop bool) {
				if !value.IsZero() {
					valAddrs = append(valAddrs, val)
				}
				return false
			})
			for _, valAddr := range valAddrs {
				if _, err := k.WithdrawValidatorCommission(ctx, valAddr); err != nil {
					return fmt.Errorf("transferring validator %s commissions to operator account: %w",
						valAddr, err)
				}
			}
		}
	}

	return nil
}
