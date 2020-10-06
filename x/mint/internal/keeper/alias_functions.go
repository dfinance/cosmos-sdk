package keeper

import sdk "github.com/cosmos/cosmos-sdk/types"

// StakingTokenSupply implements an alias call to the underlying staking keeper's.
// StakingTokenSupply to be used in BeginBlocker.
func (k Keeper) StakingTokenSupply(ctx sdk.Context) sdk.Int {
	return k.sk.StakingTokenSupply(ctx)
}

// BondedRatio returns relation of bonded staking tokens to all staking tokens (TotalSupply).
// Value is shifted by StakingTotalSupplyShift param.
// BondedRatio to be used in BeginBlocker.
//
func (k Keeper) BondedRatio(ctx sdk.Context) sdk.Dec {
	params := k.GetParams(ctx)

	bondedSupply := k.sk.TotalBondedTokens(ctx)
	totalSupply := k.sk.StakingTokenSupply(ctx)
	totalSupply = totalSupply.Add(params.StakingTotalSupplyShift)

	if totalSupply.IsPositive() {
		return bondedSupply.ToDec().QuoInt(totalSupply)
	}

	return sdk.ZeroDec()
}

// LockedRatio implements an alias call to the underlying distribution keeper's.
// LockedRatio to be used in BeginBlocker.
func (k Keeper) LockedRatio(ctx sdk.Context) sdk.Dec {
	return k.dk.LockedRatio(ctx)
}
