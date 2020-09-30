package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/distribution/types"
	"github.com/cosmos/cosmos-sdk/x/staking/exported"
)

// AllocateTokens handles distribution of the collected fees between pools and validators.
// Distribution is done based on previous block voting info.
// {dynamicFoundationPoolTax} is defined by mint module as a ratio between foundationAllocated token to all minted tokens.
func (k Keeper) AllocateTokens(
	ctx sdk.Context,
	proposerDistrPower, proposerLPPower,
	totalDistrPower, totalLPPower int64,
	proposer sdk.ConsAddress, votes types.ABCIVotes,
	dynamicFoundationPoolTax sdk.Dec,
) {

	// sanity check
	if dynamicFoundationPoolTax.IsNegative() || dynamicFoundationPoolTax.GT(sdk.OneDec()) {
		panic(fmt.Errorf("invalid dynamicFoundationPoolTax value: %s", dynamicFoundationPoolTax))
	}

	logger := k.Logger(ctx)
	params := k.GetParams(ctx)

	// fetch and clear the collected fees for distribution, since this is
	// called in BeginBlock, collected fees will be from the previous block
	// (and distributed to the previous proposer)
	feeCollector := k.supplyKeeper.GetModuleAccount(ctx, k.feeCollectorName)
	feesCollectedInt := feeCollector.GetCoins()
	feesCollected := sdk.NewDecCoinsFromCoins(feesCollectedInt...)

	// transfer collected fees to the distribution module account
	err := k.supplyKeeper.SendCoinsFromModuleToModule(ctx, k.feeCollectorName, types.ModuleName, feesCollectedInt)
	if err != nil {
		panic(err)
	}

	// temporary workaround to keep CanWithdrawInvariant happy
	// general discussions here: https://github.com/cosmos/cosmos-sdk/issues/2906#issuecomment-441867634
	if totalDistrPower == 0 {
		k.AppendToFoundationPool(ctx, feesCollected)
		return
	}

	// fund the FoundationPool using dynamic tax ratio
	pools := k.GetRewardPools(ctx)
	feesFoundationPart := feesCollected.MulDec(dynamicFoundationPoolTax)
	pools.FoundationPool = pools.FoundationPool.Add(feesFoundationPart...)
	rewardCoins := feesCollected.Sub(feesFoundationPart)

	// distribute collected fees (sub Foundation part) between pools
	validatorsPool := rewardCoins
	//
	liquidityProvidersReward := rewardCoins.MulDec(params.LiquidityProvidersPoolTax)
	pools.LiquidityProvidersPool = pools.LiquidityProvidersPool.Add(liquidityProvidersReward...)
	validatorsPool = validatorsPool.Sub(liquidityProvidersReward)
	//
	publicTreasuryReward := rewardCoins.MulDec(params.PublicTreasuryPoolTax)
	pools.PublicTreasuryPool = pools.PublicTreasuryPool.Add(publicTreasuryReward...)
	validatorsPool = validatorsPool.Sub(publicTreasuryReward)
	//
	harpReward := rewardCoins.MulDec(params.HARPTax)
	pools.HARP = pools.HARP.Add(harpReward...)
	validatorsPool = validatorsPool.Sub(harpReward)

	// feesRemainder are distribution leftovers due to truncations
	feesRemainder := validatorsPool

	// check publicTreasuryPool capacity: overflow is transferred to FoundationPool
	for i := 0; i < len(pools.PublicTreasuryPool); i++ {
		decCoin := pools.PublicTreasuryPool[i]

		diffDec := decCoin.Amount.Sub(sdk.NewDecFromInt(params.PublicTreasuryPoolCapacity))
		if diffDec.IsPositive() {
			foundationDecCoin := sdk.NewDecCoinFromDec(decCoin.Denom, diffDec)
			pools.FoundationPool = pools.FoundationPool.Add(foundationDecCoin)

			treasuryCoin := decCoin.Sub(foundationDecCoin)
			pools.PublicTreasuryPool[i] = treasuryCoin
		}
	}

	// check if LiquidityProvidersPool can be distributed
	lpPool, lpRemainder := sdk.DecCoins{}, sdk.DecCoins{}
	if totalLPPower > 0 {
		lpPool = pools.LiquidityProvidersPool
		lpRemainder = lpPool
		pools.LiquidityProvidersPool = sdk.DecCoins{}
	}

	// update pools
	k.SetRewardPools(ctx, pools)

	// distribute validatorsPool and lpPool

	// calculate previous proposer bonding reward relative to its distr power
	proposerPowerBondingRatio := sdk.NewDec(proposerDistrPower).Quo(sdk.NewDec(totalDistrPower))
	proposerBondingMultiplier := params.BaseProposerReward.Add(params.BonusProposerReward.Mul(proposerPowerBondingRatio))
	proposerBondingReward := validatorsPool.MulDecTruncate(proposerBondingMultiplier)

	// calculate previous proposer LP reward relative to its LP power
	proposerLPMultiplier := sdk.ZeroDec()
	proposerLPReward := sdk.DecCoins{}
	if proposerLPPower > 0 {
		proposerLPMultiplier = sdk.NewDec(proposerLPPower).QuoTruncate(sdk.NewDec(totalLPPower))
		proposerLPReward = lpPool.MulDecTruncate(proposerLPMultiplier)
	}

	// pay previous proposer
	proposerValidator := k.stakingKeeper.ValidatorByConsAddr(ctx, proposer)
	if proposerValidator != nil {
		ctx.EventManager().EmitEvent(
			sdk.NewEvent(
				types.EventTypeProposerReward,
				sdk.NewAttribute(sdk.AttributeKeyBondingAmount, proposerBondingReward.String()),
				sdk.NewAttribute(sdk.AttributeKeyLPAmount, proposerLPReward.String()),
				sdk.NewAttribute(types.AttributeKeyValidator, proposerValidator.GetOperator().String()),
			),
		)

		k.AllocateTokensToValidator(ctx, proposerValidator, proposerBondingReward, proposerLPReward)
		feesRemainder = feesRemainder.Sub(proposerBondingReward)
		lpRemainder = lpRemainder.Sub(proposerLPReward)
	} else {
		// previous proposer can be unknown if say, the unbonding period is 1 block, so
		// e.g. a validator undelegates at block X, it's removed entirely by
		// block X+1's endblock, then X+2 we need to refer to the previous
		// proposer for X+1, but we've forgotten about them.
		logger.Error(fmt.Sprintf(
			"WARNING: Attempt to allocate proposer rewards to unknown proposer %s. "+
				"This should happen only if the proposer unbonded completely within a single block, "+
				"which generally should not happen except in exceptional circumstances (or fuzz testing). "+
				"We recommend you investigate immediately.",
			proposer.String()))
	}

	// calculate previous voters rewards relative to their power (distribution / LP)
	voterBondingMultiplier := sdk.OneDec().Sub(proposerBondingMultiplier)
	voterLPMultiplier := sdk.OneDec().Sub(proposerLPMultiplier)
	for _, vote := range votes {
		// estimate bonding reward
		validatorBondingPowerRatio := sdk.NewDec(vote.DistributionPower).QuoTruncate(sdk.NewDec(totalDistrPower))
		validatorBondingReward := validatorsPool.MulDecTruncate(voterBondingMultiplier).MulDecTruncate(validatorBondingPowerRatio)

		// estimate LP reward
		validatorLPReward := sdk.DecCoins{}
		if vote.LPPower > 0 {
			validatorLPPowerRatio := sdk.NewDec(vote.LPPower).QuoTruncate(sdk.NewDec(totalLPPower))
			validatorLPReward = lpPool.MulDecTruncate(voterLPMultiplier).MulDecTruncate(validatorLPPowerRatio)
		}

		k.AllocateTokensToValidator(ctx, vote.Validator, validatorBondingReward, validatorLPReward)

		feesRemainder = feesRemainder.Sub(validatorBondingReward)
		lpRemainder = lpRemainder.Sub(validatorLPReward)
	}

	// transfer ValidatorsPool and LPPool remainders to FoundationPool
	k.AppendToFoundationPool(ctx, feesRemainder)
	k.AppendToFoundationPool(ctx, lpRemainder)
}

// AllocateTokensToValidator allocates tokens to a particular validator, splitting according to commission.
func (k Keeper) AllocateTokensToValidator(ctx sdk.Context, val exported.ValidatorI, bondingTokens, lpTokens sdk.DecCoins) {
	// split tokens between validator and delegators according to commission
	bondingCommission := bondingTokens.MulDec(val.GetCommission())
	bondingShared := bondingTokens.Sub(bondingCommission)

	// update current commission
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeCommission,
			sdk.NewAttribute(sdk.AttributeKeyBondingAmount, bondingCommission.String()),
			sdk.NewAttribute(types.AttributeKeyValidator, val.GetOperator().String()),
		),
	)
	currentCommission := k.GetValidatorAccumulatedCommission(ctx, val.GetOperator())
	currentCommission = currentCommission.Add(bondingCommission...)
	k.SetValidatorAccumulatedCommission(ctx, val.GetOperator(), currentCommission)

	// update current rewards
	currentRewards := k.GetValidatorCurrentRewards(ctx, val.GetOperator())
	currentRewards.BondingRewards = currentRewards.BondingRewards.Add(bondingShared...)
	currentRewards.LPRewards = currentRewards.LPRewards.Add(lpTokens...)
	k.SetValidatorCurrentRewards(ctx, val.GetOperator(), currentRewards)

	// update outstanding rewards
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeRewards,
			sdk.NewAttribute(sdk.AttributeKeyBondingAmount, bondingTokens.String()),
			sdk.NewAttribute(sdk.AttributeKeyLPAmount, lpTokens.String()),
			sdk.NewAttribute(types.AttributeKeyValidator, val.GetOperator().String()),
		),
	)
	outstanding := k.GetValidatorOutstandingRewards(ctx, val.GetOperator())
	outstanding = outstanding.Add(bondingTokens...).Add(lpTokens...)
	k.SetValidatorOutstandingRewards(ctx, val.GetOperator(), outstanding)
}
