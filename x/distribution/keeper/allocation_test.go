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

	// assert initial state: zero outstanding rewards, zero community pool, zero commission, zero current rewards
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
