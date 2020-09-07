package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/distribution/types"
)

// DistributeFromPublicTreasuryPool distributes funds from the distribution module account to
// a receiver address while updating the PublicTreasuryPool.
func (k Keeper) DistributeFromPublicTreasuryPool(ctx sdk.Context, amount sdk.Coins, receiveAddr sdk.AccAddress) error {
	pools := k.GetRewardPools(ctx)

	newPool, negative := pools.PublicTreasuryPool.SafeSub(sdk.NewDecCoinsFromCoins(amount...))
	if negative {
		return types.ErrBadDistribution
	}
	pools.PublicTreasuryPool = newPool

	err := k.supplyKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, receiveAddr, amount)
	if err != nil {
		return err
	}

	k.SetRewardPools(ctx, pools)

	return nil
}

// FundPublicTreasuryPool allows an account to directly fund the public treasury fund pool.
// The amount is first added to the distribution module account and then directly added to the pool.
func (k Keeper) FundPublicTreasuryPool(ctx sdk.Context, amount sdk.Coins, sender sdk.AccAddress) error {
	if err := k.supplyKeeper.SendCoinsFromAccountToModule(ctx, sender, types.ModuleName, amount); err != nil {
		return err
	}

	k.AppendToPublicTreasuryPool(ctx, sdk.NewDecCoinsFromCoins(amount...))

	return nil
}

// AppendToFoundationPool adds coins to the FoundationPool.
func (k Keeper) AppendToFoundationPool(ctx sdk.Context, coins sdk.DecCoins) {
	pools := k.GetRewardPools(ctx)
	pools.FoundationPool = pools.FoundationPool.Add(coins...)
	k.SetRewardPools(ctx, pools)
}

// AppendToPublicTreasuryPool adds coins to the PublicTreasuryPool.
func (k Keeper) AppendToPublicTreasuryPool(ctx sdk.Context, coins sdk.DecCoins) {
	pools := k.GetRewardPools(ctx)
	pools.PublicTreasuryPool = pools.PublicTreasuryPool.Add(coins...)
	k.SetRewardPools(ctx, pools)
}

// DistributeFromFoundationPoolToWallet distributes funds from the distribution module account to
// a receiver address while updating the FoundationPool.
func (k Keeper) DistributeFromFoundationPoolToWallet(ctx sdk.Context, amount sdk.Coins, receiveAddr sdk.AccAddress) error {
	pools := k.GetRewardPools(ctx)

	newPool, negative := pools.FoundationPool.SafeSub(sdk.NewDecCoinsFromCoins(amount...))
	if negative {
		return sdkerrors.Wrap(types.ErrBadDistribution, "FoundationPool sub")
	}
	pools.FoundationPool = newPool

	err := k.supplyKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, receiveAddr, amount)
	if err != nil {
		return err
	}

	k.SetRewardPools(ctx, pools)

	return nil
}

// DistributeFromFoundationPoolToPool transfers FoundationPool funds to an other distribution pool.
func (k Keeper) DistributeFromFoundationPoolToPool(ctx sdk.Context, amount sdk.Coins, receivePool types.RewardPoolName) error {
	if amount.IsAnyNegative() {
		return sdkerrors.Wrapf(types.ErrBadDistribution, "negative amount: %s", amount)
	}

	pools := k.GetRewardPools(ctx)
	poolsSupply := pools.TotalCoins()

	amountDecCoins := sdk.NewDecCoinsFromCoins(amount...)
	newPool, negative := pools.FoundationPool.SafeSub(amountDecCoins)
	if negative {
		return sdkerrors.Wrap(types.ErrBadDistribution, "FoundationPool sub")
	}
	pools.FoundationPool = newPool

	switch receivePool {
	case types.LiquidityProvidersPoolName:
		pools.LiquidityProvidersPool = pools.LiquidityProvidersPool.Add(amountDecCoins...)
	case types.PublicTreasuryPoolName:
		pools.PublicTreasuryPool = pools.PublicTreasuryPool.Add(amountDecCoins...)
	case types.HARPName:
		pools.HARP = pools.HARP.Add(amountDecCoins...)
	default:
		return sdkerrors.Wrapf(types.ErrBadDistribution, "unknown receivePool: %s", receivePool)
	}

	if !pools.TotalCoins().IsEqual(poolsSupply) {
		return sdkerrors.Wrap(types.ErrBadDistribution, "sanity check failed")
	}
	k.SetRewardPools(ctx, pools)

	return nil
}
