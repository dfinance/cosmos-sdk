package keeper

import (
	"fmt"

	abci "github.com/tendermint/tendermint/abci/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/distribution/types"
	"github.com/cosmos/cosmos-sdk/x/staking/exported"
)

// AllocateTokens handles distribution of the collected fees between pools and validators.
// Distribution is done based on previous block voting info.
// {dynamicFoundationPoolTax} is defined by mint module as a ratio between foundationAllocated token to all minted tokens.
func (k Keeper) AllocateTokens(
	ctx sdk.Context,
	proposerPower, totalPower int64,
	proposer sdk.ConsAddress, votes []abci.VoteInfo,
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
	if totalPower == 0 {
		k.AppendToFoundationPool(ctx, feesCollected)
		return
	}

	// fund the FoundationPool using dynamic tax ratio
	pools := k.GetRewardPools(ctx)
	feesFoundationPart := feesCollected.MulDec(dynamicFoundationPoolTax)
	pools.FoundationPool = pools.FoundationPool.Add(feesFoundationPart...)
	feesCollected = feesCollected.Sub(feesFoundationPart)

	// distribute collected fees (sub Foundation part) between pools
	pools.LiquidityProvidersPool = pools.LiquidityProvidersPool.Add(feesCollected.MulDec(params.LiquidityProvidersPoolTax)...)
	pools.PublicTreasuryPool = pools.PublicTreasuryPool.Add(feesCollected.MulDec(params.PublicTreasuryPoolTax)...)
	pools.HARP = pools.HARP.Add(feesCollected.MulDec(params.HARPTax)...)
	validatorsPool := feesCollected.MulDec(params.ValidatorsPoolTax)

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

	// distribute validatorsPool

	// calculate previous proposer reward relative to its power
	proposerPowerRatio := sdk.NewDec(proposerPower).Quo(sdk.NewDec(totalPower))
	proposerMultiplier := params.BaseProposerReward.Add(params.BonusProposerReward.MulTruncate(proposerPowerRatio))
	proposerReward := validatorsPool.MulDecTruncate(proposerMultiplier)

	// pay previous proposer
	proposerValidator := k.stakingKeeper.ValidatorByConsAddr(ctx, proposer)
	if proposerValidator != nil {
		ctx.EventManager().EmitEvent(
			sdk.NewEvent(
				types.EventTypeProposerReward,
				sdk.NewAttribute(sdk.AttributeKeyAmount, proposerReward.String()),
				sdk.NewAttribute(types.AttributeKeyValidator, proposerValidator.GetOperator().String()),
			),
		)

		k.AllocateTokensToValidator(ctx, proposerValidator, proposerReward)
		feesRemainder = feesRemainder.Sub(proposerReward)
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

	// calculate previous voters rewards relative to their power
	voterMultiplier := sdk.OneDec().Sub(proposerMultiplier)
	for _, vote := range votes {
		validator := k.stakingKeeper.ValidatorByConsAddr(ctx, vote.Validator.Address)

		validatorPowerRatio := sdk.NewDec(vote.Validator.Power).QuoTruncate(sdk.NewDec(totalPower))
		validatorReward := validatorsPool.MulDecTruncate(voterMultiplier).MulDecTruncate(validatorPowerRatio)
		k.AllocateTokensToValidator(ctx, validator, validatorReward)

		feesRemainder = feesRemainder.Sub(validatorReward)
	}

	// transfer ValidatorsPool remainder to FoundationPool
	k.AppendToFoundationPool(ctx, feesRemainder)
}

// AllocateTokensToValidator allocates tokens to a particular validator, splitting according to commission.
func (k Keeper) AllocateTokensToValidator(ctx sdk.Context, val exported.ValidatorI, tokens sdk.DecCoins) {
	// split tokens between validator and delegators according to commission
	commission := tokens.MulDec(val.GetCommission())
	shared := tokens.Sub(commission)

	// update current commission
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeCommission,
			sdk.NewAttribute(sdk.AttributeKeyAmount, commission.String()),
			sdk.NewAttribute(types.AttributeKeyValidator, val.GetOperator().String()),
		),
	)
	currentCommission := k.GetValidatorAccumulatedCommission(ctx, val.GetOperator())
	currentCommission = currentCommission.Add(commission...)
	k.SetValidatorAccumulatedCommission(ctx, val.GetOperator(), currentCommission)

	// update current rewards
	currentRewards := k.GetValidatorCurrentRewards(ctx, val.GetOperator())
	currentRewards.Rewards = currentRewards.Rewards.Add(shared...)
	k.SetValidatorCurrentRewards(ctx, val.GetOperator(), currentRewards)

	// update outstanding rewards
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeRewards,
			sdk.NewAttribute(sdk.AttributeKeyAmount, tokens.String()),
			sdk.NewAttribute(types.AttributeKeyValidator, val.GetOperator().String()),
		),
	)
	outstanding := k.GetValidatorOutstandingRewards(ctx, val.GetOperator())
	outstanding = outstanding.Add(tokens...)
	k.SetValidatorOutstandingRewards(ctx, val.GetOperator(), outstanding)
}
