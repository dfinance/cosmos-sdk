package keeper

import (
	"testing"

	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/distribution/types"
	"github.com/cosmos/cosmos-sdk/x/staking"
)

// Allocate tokens to validator and check current commission and rewards amount.
func TestAllocateTokensToValidatorWithCommission(t *testing.T) {
	ctx, _, k, sk, _ := CreateTestInputDefault(t, false, 1000)
	sh := staking.NewHandler(sk)

	// create validator with 50% commission
	commission := staking.NewCommissionRates(sdk.NewDecWithPrec(5, 1), sdk.NewDecWithPrec(5, 1), sdk.NewDec(0))
	msg := staking.NewMsgCreateValidator(
		valOpAddr1, valConsPk1,
		sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(100)), staking.Description{}, commission, sdk.OneInt(),
	)

	res, err := sh(ctx, msg)
	require.NoError(t, err)
	require.NotNil(t, res)

	val := sk.Validator(ctx, valOpAddr1)

	// allocate tokens
	tokens := sdk.DecCoins{
		{Denom: sdk.DefaultBondDenom, Amount: sdk.NewDec(10)},
	}
	k.AllocateTokensToValidator(ctx, val, tokens)

	// check commission
	expected := sdk.DecCoins{
		{Denom: sdk.DefaultBondDenom, Amount: sdk.NewDec(5)},
	}
	require.Equal(t, expected, k.GetValidatorAccumulatedCommission(ctx, val.GetOperator()))

	// check current rewards
	require.Equal(t, expected, k.GetValidatorCurrentRewards(ctx, val.GetOperator()).Rewards)
}

// Allocate tokens between two validators (commission: 1st - 50%, 2nd - 0%) and no pools distribution.
func TestAllocateTokensToManyValidators(t *testing.T) {
	// custom params disabling pools distribution
	distrParams := types.DefaultParams()
	distrParams.ValidatorsPoolTax = sdk.NewDecWithPrec(1, 0)
	distrParams.LiquidityProvidersPoolTax = sdk.ZeroDec()
	distrParams.PublicTreasuryPoolTax = sdk.ZeroDec()
	distrParams.HARPTax = sdk.ZeroDec()
	distrParams.BaseProposerReward = sdk.NewDecWithPrec(1, 2)
	distrParams.BonusProposerReward = sdk.NewDecWithPrec(4, 2)

	ctx, ak, _, k, sk, _, supplyKeeper := CreateTestInputAdvanced(t, false, 1000, distrParams)
	sh := staking.NewHandler(sk)

	// create validator with 50% commission
	commission := staking.NewCommissionRates(sdk.NewDecWithPrec(5, 1), sdk.NewDecWithPrec(5, 1), sdk.NewDec(0))
	msg := staking.NewMsgCreateValidator(valOpAddr1, valConsPk1,
		sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(100)), staking.Description{}, commission, sdk.OneInt())

	res, err := sh(ctx, msg)
	require.NoError(t, err)
	require.NotNil(t, res)

	// create second validator with 0% commission
	commission = staking.NewCommissionRates(sdk.NewDec(0), sdk.NewDec(0), sdk.NewDec(0))
	msg = staking.NewMsgCreateValidator(valOpAddr2, valConsPk2,
		sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(100)), staking.Description{}, commission, sdk.OneInt())

	res, err = sh(ctx, msg)
	require.NoError(t, err)
	require.NotNil(t, res)

	abciValA := abci.Validator{
		Address: valConsPk1.Address(),
		Power:   100,
	}
	abciValB := abci.Validator{
		Address: valConsPk2.Address(),
		Power:   100,
	}

	// assert initial state: zero outstanding rewards, zero pools, zero commission, zero current rewards
	require.True(t, k.GetValidatorOutstandingRewards(ctx, valOpAddr1).IsZero())
	require.True(t, k.GetValidatorOutstandingRewards(ctx, valOpAddr2).IsZero())
	require.True(t, k.GetRewardPools(ctx).LiquidityProvidersPool.IsZero())
	require.True(t, k.GetRewardPools(ctx).FoundationPool.IsZero())
	require.True(t, k.GetRewardPools(ctx).PublicTreasuryPool.IsZero())
	require.True(t, k.GetRewardPools(ctx).HARP.IsZero())
	require.True(t, k.GetValidatorAccumulatedCommission(ctx, valOpAddr1).IsZero())
	require.True(t, k.GetValidatorAccumulatedCommission(ctx, valOpAddr2).IsZero())
	require.True(t, k.GetValidatorCurrentRewards(ctx, valOpAddr1).Rewards.IsZero())
	require.True(t, k.GetValidatorCurrentRewards(ctx, valOpAddr2).Rewards.IsZero())

	// allocate tokens as if both had voted and second was proposer + pre allocated Foundation tokens
	fees := sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(100)))
	feeCollector := supplyKeeper.GetModuleAccount(ctx, k.feeCollectorName)
	require.NotNil(t, feeCollector)

	err = feeCollector.SetCoins(fees)
	require.NoError(t, err)
	ak.SetAccount(ctx, feeCollector)

	votes := []abci.VoteInfo{
		{
			Validator:       abciValA,
			SignedLastBlock: true,
		},
		{
			Validator:       abciValB,
			SignedLastBlock: true,
		},
	}
	k.AllocateTokens(ctx, 200, 200, valConsAddr2, votes, sdk.ZeroDec())

	// val1, val2: outstanding rewards (100% as no pools distribution involved)
	require.Equal(t, sdk.DecCoins{{Denom: sdk.DefaultBondDenom, Amount: sdk.NewDecWithPrec(475, 1)}}, k.GetValidatorOutstandingRewards(ctx, valOpAddr1))
	require.Equal(t, sdk.DecCoins{{Denom: sdk.DefaultBondDenom, Amount: sdk.NewDecWithPrec(525, 1)}}, k.GetValidatorOutstandingRewards(ctx, valOpAddr2))
	// empty pools (FoundationPool is zero as there shouldn't be any leftovers)
	require.True(t, k.GetRewardPools(ctx).FoundationPool.IsZero())
	require.True(t, k.GetRewardPools(ctx).PublicTreasuryPool.IsZero())
	require.True(t, k.GetRewardPools(ctx).LiquidityProvidersPool.IsZero())
	require.True(t, k.GetRewardPools(ctx).HARP.IsZero())
	// val1 commissions: 50% commission -> 47.5% / 2 = 23.75%
	require.Equal(t, sdk.DecCoins{{Denom: sdk.DefaultBondDenom, Amount: sdk.NewDecWithPrec(2375, 2)}}, k.GetValidatorAccumulatedCommission(ctx, valOpAddr1))
	// val2 commissions: zero
	require.True(t, k.GetValidatorAccumulatedCommission(ctx, valOpAddr2).IsZero())
	// val1 rewards: outstanding / 2
	require.Equal(t, sdk.DecCoins{{Denom: sdk.DefaultBondDenom, Amount: sdk.NewDecWithPrec(2375, 2)}}, k.GetValidatorCurrentRewards(ctx, valOpAddr1).Rewards)
	// val2 rewards: outstanding (as it has no commission)
	require.Equal(t, sdk.DecCoins{{Denom: sdk.DefaultBondDenom, Amount: sdk.NewDecWithPrec(525, 1)}}, k.GetValidatorCurrentRewards(ctx, valOpAddr2).Rewards)

	// check module funds invariant
	invMsg, invBroken := ModuleAccountInvariant(k)(ctx)
	require.False(t, invBroken, invMsg)
}

// Allocate tokens between three validators, rewards would be truncated so FoundationPool must be non-empty.
func TestAllocateTokensTruncation(t *testing.T) {
	// custom params disabling pools distribution
	distrParams := types.DefaultParams()
	distrParams.ValidatorsPoolTax = sdk.NewDecWithPrec(1, 0)
	distrParams.LiquidityProvidersPoolTax = sdk.ZeroDec()
	distrParams.PublicTreasuryPoolTax = sdk.ZeroDec()
	distrParams.HARPTax = sdk.ZeroDec()
	distrParams.BaseProposerReward = sdk.NewDecWithPrec(1, 2)
	distrParams.BonusProposerReward = sdk.NewDecWithPrec(4, 2)

	ctx, ak, _, k, sk, _, supplyKeeper := CreateTestInputAdvanced(t, false, 1000000, distrParams)
	sh := staking.NewHandler(sk)

	// create validator with 10% commission
	commission := staking.NewCommissionRates(sdk.NewDecWithPrec(1, 1), sdk.NewDecWithPrec(1, 1), sdk.NewDec(0))
	msg := staking.NewMsgCreateValidator(valOpAddr1, valConsPk1,
		sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(110)), staking.Description{}, commission, sdk.OneInt())
	res, err := sh(ctx, msg)
	require.NoError(t, err)
	require.NotNil(t, res)

	// create second validator with 10% commission
	commission = staking.NewCommissionRates(sdk.NewDecWithPrec(1, 1), sdk.NewDecWithPrec(1, 1), sdk.NewDec(0))
	msg = staking.NewMsgCreateValidator(valOpAddr2, valConsPk2,
		sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(100)), staking.Description{}, commission, sdk.OneInt())
	res, err = sh(ctx, msg)
	require.NoError(t, err)
	require.NotNil(t, res)

	// create third validator with 10% commission
	commission = staking.NewCommissionRates(sdk.NewDecWithPrec(1, 1), sdk.NewDecWithPrec(1, 1), sdk.NewDec(0))
	msg = staking.NewMsgCreateValidator(valOpAddr3, valConsPk3,
		sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(100)), staking.Description{}, commission, sdk.OneInt())
	res, err = sh(ctx, msg)
	require.NoError(t, err)
	require.NotNil(t, res)

	abciValA := abci.Validator{
		Address: valConsPk1.Address(),
		Power:   11,
	}
	abciValB := abci.Validator{
		Address: valConsPk2.Address(),
		Power:   10,
	}
	abciValС := abci.Validator{
		Address: valConsPk3.Address(),
		Power:   10,
	}

	// assert initial state: zero outstanding rewards, zero public treasury pool, zero commission, zero current rewards
	require.True(t, k.GetValidatorOutstandingRewards(ctx, valOpAddr1).IsZero())
	require.True(t, k.GetValidatorOutstandingRewards(ctx, valOpAddr2).IsZero())
	require.True(t, k.GetValidatorOutstandingRewards(ctx, valOpAddr3).IsZero())
	require.True(t, k.GetRewardPools(ctx).PublicTreasuryPool.IsZero())
	require.True(t, k.GetValidatorAccumulatedCommission(ctx, valOpAddr1).IsZero())
	require.True(t, k.GetValidatorAccumulatedCommission(ctx, valOpAddr2).IsZero())
	require.True(t, k.GetValidatorCurrentRewards(ctx, valOpAddr1).Rewards.IsZero())
	require.True(t, k.GetValidatorCurrentRewards(ctx, valOpAddr2).Rewards.IsZero())

	// allocate tokens as if both had voted and second was proposer
	fees := sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(634195840)))

	feeCollector := supplyKeeper.GetModuleAccount(ctx, k.feeCollectorName)
	require.NotNil(t, feeCollector)

	err = feeCollector.SetCoins(fees)
	require.NoError(t, err)

	ak.SetAccount(ctx, feeCollector)

	votes := []abci.VoteInfo{
		{
			Validator:       abciValA,
			SignedLastBlock: true,
		},
		{
			Validator:       abciValB,
			SignedLastBlock: true,
		},
		{
			Validator:       abciValС,
			SignedLastBlock: true,
		},
	}
	k.AllocateTokens(ctx, 31, 31, valConsAddr2, votes, sdk.ZeroDec())

	// check validators has outstanding rewards
	require.True(t, k.GetValidatorOutstandingRewards(ctx, valOpAddr1).IsValid())
	require.True(t, k.GetValidatorOutstandingRewards(ctx, valOpAddr1).IsAllPositive())
	require.True(t, k.GetValidatorOutstandingRewards(ctx, valOpAddr2).IsValid())
	require.True(t, k.GetValidatorOutstandingRewards(ctx, valOpAddr2).IsAllPositive())
	require.True(t, k.GetValidatorOutstandingRewards(ctx, valOpAddr3).IsValid())
	require.True(t, k.GetValidatorOutstandingRewards(ctx, valOpAddr3).IsAllPositive())

	// check FoundationPool has some leftovers
	require.True(t, k.GetRewardPools(ctx).FoundationPool.IsAllPositive())

	// check other pools are empty
	require.True(t, k.GetRewardPools(ctx).LiquidityProvidersPool.IsZero())
	require.True(t, k.GetRewardPools(ctx).PublicTreasuryPool.IsZero())
	require.True(t, k.GetRewardPools(ctx).HARP.IsZero())

	// check module funds invariant
	invMsg, invBroken := ModuleAccountInvariant(k)(ctx)
	require.False(t, invBroken, invMsg)
}

// Allocate tokens and check it is distributed without a loss (between pools and validators).
func TestAllocateTokensPools(t *testing.T) {
	distrParams := types.DefaultParams()
	ctx, ak, _, k, sk, _, supplyKeeper := CreateTestInputAdvanced(t, false, 1000000, distrParams)
	sh := staking.NewHandler(sk)

	// create validator 1 with 10% commission
	{
		commission := staking.NewCommissionRates(sdk.NewDecWithPrec(1, 1), sdk.NewDecWithPrec(1, 1), sdk.NewDec(0))
		msg := staking.NewMsgCreateValidator(valOpAddr1, valConsPk1,
			sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(100)), staking.Description{}, commission, sdk.OneInt())
		res, err := sh(ctx, msg)
		require.NoError(t, err)
		require.NotNil(t, res)
	}

	// create validator 2 with 20% commission
	{
		commission := staking.NewCommissionRates(sdk.NewDecWithPrec(2, 1), sdk.NewDecWithPrec(2, 1), sdk.NewDec(0))
		msg := staking.NewMsgCreateValidator(valOpAddr2, valConsPk2,
			sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(100)), staking.Description{}, commission, sdk.OneInt())
		res, err := sh(ctx, msg)
		require.NoError(t, err)
		require.NotNil(t, res)
	}

	// prepare voting results (validators has the same power)
	// validator1 is a proposer
	abciValA := abci.Validator{
		Address: valConsPk1.Address(),
		Power:   10,
	}
	abciValB := abci.Validator{
		Address: valConsPk2.Address(),
		Power:   10,
	}
	votes := []abci.VoteInfo{
		{Validator: abciValA, SignedLastBlock: true},
		{Validator: abciValB, SignedLastBlock: false},
	}

	// assert initial state: zero outstanding rewards, zero pools, zero distr module balance
	{
		require.True(t, k.GetValidatorAccumulatedCommission(ctx, valOpAddr1).IsZero())
		require.True(t, k.GetValidatorOutstandingRewards(ctx, valOpAddr1).IsZero())
		require.True(t, k.GetValidatorCurrentRewards(ctx, valOpAddr1).Rewards.IsZero())
		//
		require.True(t, k.GetValidatorOutstandingRewards(ctx, valOpAddr2).IsZero())
		require.True(t, k.GetValidatorAccumulatedCommission(ctx, valOpAddr2).IsZero())
		require.True(t, k.GetValidatorCurrentRewards(ctx, valOpAddr2).Rewards.IsZero())
		//
		require.True(t, k.GetRewardPools(ctx).TotalCoins().IsZero())
		//
		supplyKeeper.GetModuleAccount(ctx, types.ModuleName).GetCoins().IsZero()
	}

	// allocate current fees
	feesAmt := sdk.NewInt(100000)
	{
		fees := sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, feesAmt))
		feeCollector := supplyKeeper.GetModuleAccount(ctx, k.feeCollectorName)
		require.NotNil(t, feeCollector)
		require.NoError(t, feeCollector.SetCoins(fees))
		ak.SetAccount(ctx, feeCollector)
	}

	// allocate tokens with foundation part of 1%
	foundationTax := sdk.NewDecWithPrec(1, 2)
	k.AllocateTokens(ctx, abciValA.Power, abciValA.Power+abciValB.Power, valConsAddr1, votes, foundationTax)

	// check reward pools distribution and only one coin exists in a pool
	rewardPools := k.GetRewardPools(ctx)
	foundationTaxDecAmt := sdk.NewDecFromInt(feesAmt).Mul(foundationTax)
	feesDistributionDecAmt := sdk.NewDecFromInt(feesAmt).Sub(foundationTaxDecAmt)

	{
		expLiquidityDecAmt := feesDistributionDecAmt.Mul(distrParams.LiquidityProvidersPoolTax)
		require.Len(t, rewardPools.LiquidityProvidersPool, 1)
		require.True(t,
			rewardPools.LiquidityProvidersPool.AmountOf(sdk.DefaultBondDenom).Equal(expLiquidityDecAmt),
			"LiquidityProvidersPool: %s / %s", rewardPools.LiquidityProvidersPool.AmountOf(sdk.DefaultBondDenom), expLiquidityDecAmt,
		)
	}

	{
		expTreasuryDecAmt := feesDistributionDecAmt.Mul(distrParams.PublicTreasuryPoolTax)
		require.Len(t, rewardPools.PublicTreasuryPool, 1)
		require.True(t,
			rewardPools.PublicTreasuryPool.AmountOf(sdk.DefaultBondDenom).Equal(expTreasuryDecAmt),
			"PublicTreasuryPool: %s / %s", rewardPools.PublicTreasuryPool.AmountOf(sdk.DefaultBondDenom), expTreasuryDecAmt,
		)
	}

	{
		expHARPDecAmt := feesDistributionDecAmt.Mul(distrParams.HARPTax)
		require.Len(t, rewardPools.HARP, 1)
		require.True(t,
			rewardPools.HARP.AmountOf(sdk.DefaultBondDenom).Equal(expHARPDecAmt),
			"HARP: %s / %s", rewardPools.HARP.AmountOf(sdk.DefaultBondDenom), expHARPDecAmt,
		)
	}

	{
		require.Len(t, rewardPools.FoundationPool, 1)
		require.True(t,
			rewardPools.FoundationPool.AmountOf(sdk.DefaultBondDenom).GTE(foundationTaxDecAmt),
			"FoundationPool (unconditional level): %s / %s", rewardPools.FoundationPool.AmountOf(sdk.DefaultBondDenom), foundationTaxDecAmt,
		)
	}

	// get validators rewards
	val1OutstangingRewardsAmt := k.GetValidatorOutstandingRewards(ctx, valOpAddr1).AmountOf(sdk.DefaultBondDenom)
	val1CommissionRewardsAmt := k.GetValidatorAccumulatedCommission(ctx, valOpAddr1).AmountOf(sdk.DefaultBondDenom)
	val1DelegatorsRewardsAmt := k.GetValidatorCurrentRewards(ctx, valOpAddr1).Rewards.AmountOf(sdk.DefaultBondDenom)
	{
		require.Len(t, k.GetValidatorOutstandingRewards(ctx, valOpAddr1), 1)
		require.Len(t, k.GetValidatorAccumulatedCommission(ctx, valOpAddr1), 1)
		require.Len(t, k.GetValidatorCurrentRewards(ctx, valOpAddr1).Rewards, 1)
	}
	//
	val2OutstangingRewardsAmt := k.GetValidatorOutstandingRewards(ctx, valOpAddr2).AmountOf(sdk.DefaultBondDenom)
	val2CommissionRewardsAmt := k.GetValidatorAccumulatedCommission(ctx, valOpAddr2).AmountOf(sdk.DefaultBondDenom)
	val2DelegatorsRewardsAmt := k.GetValidatorCurrentRewards(ctx, valOpAddr2).Rewards.AmountOf(sdk.DefaultBondDenom)
	{
		require.Len(t, k.GetValidatorOutstandingRewards(ctx, valOpAddr2), 1)
		require.Len(t, k.GetValidatorAccumulatedCommission(ctx, valOpAddr2), 1)
		require.Len(t, k.GetValidatorCurrentRewards(ctx, valOpAddr2).Rewards, 1)
	}

	// check ValidatorsPool distribution
	// actual sum of validators outstanding rewards might be less that ValidatorsPool distribution due to Truncate()
	truncErrorDecAmt := rewardPools.FoundationPool.AmountOf(sdk.DefaultBondDenom).Sub(foundationTaxDecAmt)
	expValidatorsPoolDecAmt := feesDistributionDecAmt.Mul(distrParams.ValidatorsPoolTax)
	{
		valsOutstaningRewards := val1OutstangingRewardsAmt.Add(val2OutstangingRewardsAmt)
		require.True(t,
			valsOutstaningRewards.Add(truncErrorDecAmt).Equal(expValidatorsPoolDecAmt),
			"ValidatorsPool (outstanding sum): %s / %s", valsOutstaningRewards, expValidatorsPoolDecAmt,
		)
	}

	// check FoundationPool exact value (including Truncations)
	require.True(t,
		rewardPools.FoundationPool.AmountOf(sdk.DefaultBondDenom).Equal(foundationTaxDecAmt.Add(truncErrorDecAmt)),
		"FoundationPool (exact): %s / %s", rewardPools.FoundationPool.AmountOf(sdk.DefaultBondDenom), foundationTaxDecAmt.Add(truncErrorDecAmt),
	)

	// check validator outstanding = commission + delegators rewards
	{
		require.True(t,
			val1OutstangingRewardsAmt.Equal(val1CommissionRewardsAmt.Add(val1DelegatorsRewardsAmt)),
			"Validator 1 outstanding check: %s / %s", val1OutstangingRewardsAmt, val1CommissionRewardsAmt.Add(val1DelegatorsRewardsAmt),
		)
		require.True(t,
			val2OutstangingRewardsAmt.Equal(val2CommissionRewardsAmt.Add(val2DelegatorsRewardsAmt)),
			"Validator 2 outstanding check: %s / %s", val2OutstangingRewardsAmt, val2CommissionRewardsAmt.Add(val2DelegatorsRewardsAmt),
		)
	}

	// check tokens distribution between validators
	{
		// estimate proposer bonus (base + bonus/2 as they have the same power), proposer / voters multiplier
		proposerBonusRatio := distrParams.BaseProposerReward.Add(distrParams.BonusProposerReward.QuoInt64(2))
		votersRatio := sdk.OneDec().Sub(proposerBonusRatio)
		val1RewardsRatio := proposerBonusRatio.Add(votersRatio.QuoInt64(2))
		val2RewardsRatio := votersRatio.QuoInt64(2)
		// estimate validators outstanding rewards
		expVal1OutstandingRewards := expValidatorsPoolDecAmt.MulTruncate(val1RewardsRatio)
		expVal2OutstandingRewards := expValidatorsPoolDecAmt.MulTruncate(val2RewardsRatio)
		// check
		require.True(t,
			val1OutstangingRewardsAmt.Equal(expVal1OutstandingRewards),
			"Validator 1 outstanding: %s, %s", val1OutstangingRewardsAmt, expVal1OutstandingRewards,
		)
		require.True(t,
			val2OutstangingRewardsAmt.Equal(expVal2OutstandingRewards),
			"Validator 2 outstanding: %s, %s", val2OutstangingRewardsAmt, expVal2OutstandingRewards,
		)
	}

	// check coins were transferred from feeCollector to distr module account
	{
		feeAcc := supplyKeeper.GetModuleAccount(ctx, k.feeCollectorName)
		require.True(t, feeAcc.GetCoins().IsZero())

		distrAcc := supplyKeeper.GetModuleAccount(ctx, types.ModuleName)
		require.Len(t, distrAcc.GetCoins(), 1)
		require.True(t, distrAcc.GetCoins().AmountOf(sdk.DefaultBondDenom).Equal(feesAmt))
	}
}

// Allocate tokens and check that overflowing PublicTreasuryPool does transfer the diff to FoundationPool.
func TestAllocatePublicTreasuryOverflow(t *testing.T) {
	// define custom taxes (only PublicTreasury distribution)
	distrParams := types.DefaultParams()
	distrParams.ValidatorsPoolTax = sdk.ZeroDec()
	distrParams.LiquidityProvidersPoolTax = sdk.ZeroDec()
	distrParams.PublicTreasuryPoolTax = sdk.NewDecWithPrec(1, 0)
	distrParams.HARPTax = sdk.ZeroDec()

	ctx, ak, _, k, sk, _, supplyKeeper := CreateTestInputAdvanced(t, false, 1000000, distrParams)
	sh := staking.NewHandler(sk)

	// create validator with 0% commission
	{
		commission := staking.NewCommissionRates(sdk.ZeroDec(), sdk.ZeroDec(), sdk.NewDec(0))
		msg := staking.NewMsgCreateValidator(valOpAddr1, valConsPk1,
			sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(100)), staking.Description{}, commission, sdk.OneInt())
		res, err := sh(ctx, msg)
		require.NoError(t, err)
		require.NotNil(t, res)
	}

	// prepare voting results
	abciVal := abci.Validator{
		Address: valConsPk1.Address(),
		Power:   10,
	}
	votes := []abci.VoteInfo{
		{Validator: abciVal, SignedLastBlock: true},
	}

	// allocate current fees
	overflowAmt := sdk.NewInt(10000)
	feesAmt := distrParams.PublicTreasuryPoolCapacity.Add(overflowAmt)
	{
		fees := sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, feesAmt))
		feeCollector := supplyKeeper.GetModuleAccount(ctx, k.feeCollectorName)
		require.NotNil(t, feeCollector)
		require.NoError(t, feeCollector.SetCoins(fees))
		ak.SetAccount(ctx, feeCollector)
	}

	// assert initial state: zero outstanding rewards, zero pools, zero distr module balance
	{
		require.True(t, k.GetValidatorAccumulatedCommission(ctx, valOpAddr1).IsZero())
		require.True(t, k.GetValidatorOutstandingRewards(ctx, valOpAddr1).IsZero())
		require.True(t, k.GetValidatorCurrentRewards(ctx, valOpAddr1).Rewards.IsZero())
		//
		require.True(t, k.GetRewardPools(ctx).TotalCoins().IsZero())
		//
		supplyKeeper.GetModuleAccount(ctx, types.ModuleName).GetCoins().IsZero()
	}

	// allocate tokens without foundation part
	k.AllocateTokens(ctx, abciVal.Power, abciVal.Power, valConsAddr1, votes, sdk.ZeroDec())

	// check pools distribution
	rewardPools := k.GetRewardPools(ctx)

	// out of scope pools should be empty as well as the validator rewards
	require.True(t, rewardPools.LiquidityProvidersPool.IsZero())
	require.True(t, rewardPools.HARP.IsZero())
	require.True(t, k.GetValidatorOutstandingRewards(ctx, valOpAddr1).IsZero())

	// check overflow goes to FoundationPool
	foundationAmt := rewardPools.FoundationPool.AmountOf(sdk.DefaultBondDenom)
	require.True(t, foundationAmt.TruncateInt().Equal(overflowAmt))

	// check PublicTreasury capped to its capacity
	treasuryAmt := rewardPools.PublicTreasuryPool.AmountOf(sdk.DefaultBondDenom)
	require.True(t, treasuryAmt.TruncateInt().Equal(distrParams.PublicTreasuryPoolCapacity))

	// check coins were transferred from feeCollector to distr module account
	{
		feeAcc := supplyKeeper.GetModuleAccount(ctx, k.feeCollectorName)
		require.True(t, feeAcc.GetCoins().IsZero())

		distrAcc := supplyKeeper.GetModuleAccount(ctx, types.ModuleName)
		require.Len(t, distrAcc.GetCoins(), 1)
		require.True(t, distrAcc.GetCoins().AmountOf(sdk.DefaultBondDenom).Equal(feesAmt))
	}
}
