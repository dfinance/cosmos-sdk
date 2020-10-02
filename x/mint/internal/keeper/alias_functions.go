package keeper

import sdk "github.com/cosmos/cosmos-sdk/types"

// StakingTokenSupply implements an alias call to the underlying staking keeper's
// StakingTokenSupply to be used in BeginBlocker.
func (k Keeper) StakingTokenSupply(ctx sdk.Context) sdk.Int {
	return k.sk.StakingTokenSupply(ctx)
}

// BondedRatio implements an alias call to the underlying staking keeper's
// BondedRatio to be used in BeginBlocker.
func (k Keeper) BondedRatio(ctx sdk.Context) sdk.Dec {
	return k.sk.BondedRatio(ctx)
}

// LockedRatio implements an alias call to the underlying distribution keeper's
// LockedRatio to be used in BeginBlocker.
func (k Keeper) LockedRatio(ctx sdk.Context) sdk.Dec {
	return k.dk.LockedRatio(ctx)
}
