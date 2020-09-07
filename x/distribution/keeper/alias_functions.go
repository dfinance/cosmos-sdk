package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/distribution/types"
	"github.com/cosmos/cosmos-sdk/x/supply/exported"
)

// GetValidatorOutstandingRewardsCoins returns outstanding rewards for a validator.
func (k Keeper) GetValidatorOutstandingRewardsCoins(ctx sdk.Context, val sdk.ValAddress) sdk.DecCoins {
	return k.GetValidatorOutstandingRewards(ctx, val)
}

// GetRewardPoolsCoins returns sum of reward pools coins.
func (k Keeper) GetRewardPoolsCoins(ctx sdk.Context) sdk.DecCoins {
	pools := k.GetRewardPools(ctx)

	return pools.TotalCoins()
}

// GetDistributionAccount returns the distribution ModuleAccount.
func (k Keeper) GetDistributionAccount(ctx sdk.Context) exported.ModuleAccountI {
	return k.supplyKeeper.GetModuleAccount(ctx, types.ModuleName)
}
