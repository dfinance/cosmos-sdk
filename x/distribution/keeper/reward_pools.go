package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
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
