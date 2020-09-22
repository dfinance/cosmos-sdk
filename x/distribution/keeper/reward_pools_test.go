package keeper

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/distribution/types"
)

// Test ingres -> PublicTreasuryPool transfer.
func TestFundPublicTreasuryPool(t *testing.T) {
	// nolint dogsled
	ctx, _, bk, keeper, _, _, _, _ := CreateTestInputAdvanced(t, false, 1000, types.DefaultParams())

	amount := sdk.NewCoins(sdk.NewInt64Coin("stake", 100))
	_ = bk.SetCoins(ctx, delAddr1, amount)

	initPool := keeper.GetRewardPools(ctx)
	assert.Empty(t, initPool.PublicTreasuryPool)

	err := keeper.FundPublicTreasuryPool(ctx, amount, delAddr1)
	assert.Nil(t, err)

	assert.Equal(t, initPool.PublicTreasuryPool.Add(sdk.NewDecCoinsFromCoins(amount...)...), keeper.GetRewardPools(ctx).PublicTreasuryPool)
	assert.Empty(t, bk.GetCoins(ctx, delAddr1))
}

// Test PublicTreasuryPool -> ingres transfer.
func TestDistributeFromPublicTreasuryPool(t *testing.T) {
	// nolint dogsled
	ctx, ak, _, k, _, _, sk, _ := CreateTestInputAdvanced(t, false, 1000, types.DefaultParams())

	// allocate distr module tokens
	poolCoins := sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(10000)))
	{
		dmacc := sk.GetModuleAccount(ctx, types.ModuleName)
		require.NotNil(t, dmacc)
		require.NoError(t, dmacc.SetCoins(poolCoins))
		ak.SetAccount(ctx, dmacc)
	}

	// set pool supply
	pools := k.GetRewardPools(ctx)
	require.True(t, pools.PublicTreasuryPool.Empty())
	pools.PublicTreasuryPool = sdk.NewDecCoinsFromCoins(poolCoins...)
	k.SetRewardPools(ctx, pools)

	// distribute
	receiverBalance := ak.GetAccount(ctx, delAddr1).GetCoins()
	err := k.DistributeFromPublicTreasuryPool(ctx, poolCoins, delAddr1)
	require.NoError(t, err)

	// check
	require.True(t, sk.GetModuleAccount(ctx, types.ModuleName).GetCoins().Empty())
	require.True(t, ak.GetAccount(ctx, delAddr1).GetCoins().IsEqual(receiverBalance.Add(poolCoins...)))
	require.True(t, k.GetRewardPools(ctx).PublicTreasuryPool.Empty())
	require.True(t, k.GetRewardPools(ctx).TotalCoins().Empty())
}

// Test FoundationPool -> other pools transfer.
func TestDistributeFromFoundationPoolToPool(t *testing.T) {
	// nolint dogsled
	ctx, ak, _, k, _, _, sk, _ := CreateTestInputAdvanced(t, false, 1000, types.DefaultParams())

	// allocate distr module tokens
	poolCoins := sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(90000)))
	{
		dmacc := sk.GetModuleAccount(ctx, types.ModuleName)
		require.NotNil(t, dmacc)
		require.NoError(t, dmacc.SetCoins(poolCoins))
		ak.SetAccount(ctx, dmacc)
	}

	// set pool supply
	pools := k.GetRewardPools(ctx)
	require.True(t, pools.FoundationPool.Empty())
	require.True(t, pools.LiquidityProvidersPool.Empty())
	require.True(t, pools.PublicTreasuryPool.Empty())
	require.True(t, pools.HARP.Empty())
	pools.FoundationPool = sdk.NewDecCoinsFromCoins(poolCoins...)
	k.SetRewardPools(ctx, pools)

	// distribute to three pools
	distributeAmt := poolCoins.AmountOf(sdk.DefaultBondDenom).Quo(sdk.NewInt(3))
	distributeCoins := sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, distributeAmt))
	require.NoError(t, k.DistributeFromFoundationPoolToPool(ctx, distributeCoins, types.LiquidityProvidersPoolName))
	require.NoError(t, k.DistributeFromFoundationPoolToPool(ctx, distributeCoins, types.PublicTreasuryPoolName))
	require.NoError(t, k.DistributeFromFoundationPoolToPool(ctx, distributeCoins, types.HARPName))

	// check
	require.True(t, sk.GetModuleAccount(ctx, types.ModuleName).GetCoins().IsEqual(poolCoins))
	require.True(t, k.GetRewardPools(ctx).TotalCoins().IsEqual(sdk.NewDecCoinsFromCoins(poolCoins...)))
	require.True(t, k.GetRewardPools(ctx).FoundationPool.Empty())
	require.True(t, k.GetRewardPools(ctx).LiquidityProvidersPool.IsEqual(sdk.NewDecCoinsFromCoins(distributeCoins...)))
	require.True(t, k.GetRewardPools(ctx).PublicTreasuryPool.IsEqual(sdk.NewDecCoinsFromCoins(distributeCoins...)))
	require.True(t, k.GetRewardPools(ctx).HARP.IsEqual(sdk.NewDecCoinsFromCoins(distributeCoins...)))
}

// Test FoundationPool -> ingres transfer.
func TestDistributeFromFoundationPoolToWallet(t *testing.T) {
	// nolint dogsled
	ctx, ak, _, k, _, _, sk, _ := CreateTestInputAdvanced(t, false, 1000, types.DefaultParams())

	// allocate distr module tokens
	poolCoins := sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(10000)))
	{
		dmacc := sk.GetModuleAccount(ctx, types.ModuleName)
		require.NotNil(t, dmacc)
		require.NoError(t, dmacc.SetCoins(poolCoins))
		ak.SetAccount(ctx, dmacc)
	}

	// set pool supply
	pools := k.GetRewardPools(ctx)
	require.True(t, pools.PublicTreasuryPool.Empty())
	pools.FoundationPool = sdk.NewDecCoinsFromCoins(poolCoins...)
	k.SetRewardPools(ctx, pools)

	// distribute
	receiverBalance := ak.GetAccount(ctx, delAddr1).GetCoins()
	err := k.DistributeFromFoundationPoolToWallet(ctx, poolCoins, delAddr1)
	require.NoError(t, err)

	// check
	require.True(t, sk.GetModuleAccount(ctx, types.ModuleName).GetCoins().Empty())
	require.True(t, ak.GetAccount(ctx, delAddr1).GetCoins().IsEqual(receiverBalance.Add(poolCoins...)))
	require.True(t, k.GetRewardPools(ctx).FoundationPool.Empty())
	require.True(t, k.GetRewardPools(ctx).TotalCoins().Empty())
}
