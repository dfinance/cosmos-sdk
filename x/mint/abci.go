package mint

import (
	"fmt"
	"math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/mint/internal/types"
)

// BeginBlocker mints new tokens for the previous block.
func BeginBlocker(ctx sdk.Context, k Keeper) {
	// fetch stored minter & params
	minter := k.GetMinter(ctx)
	params := k.GetParams(ctx)

	// update and get blockDuration estimation
	k.AdjustAvgBLockDurEstimation(ctx)
	blocksPerYear, err := k.GetAvgBlocksPerYear(ctx)
	if err != nil {
		k.Logger(ctx).Info(fmt.Sprintf("mint skipped as blocksPerYear estimation is not available: %v", err))
		return
	}
	// sanity check
	if blocksPerYear > math.MaxInt64 {
		panic(fmt.Errorf("invalid blocksPerYear estimation: %d", blocksPerYear))
	}
	minter.BlocksPerYear = blocksPerYear

	// update annual params
	if k.CheckAnnualParamsAdjust(ctx) {
		params.InflationMin, params.InflationMax = minter.NextMinMaxInflation(params)
		k.SetParams(ctx, params)
		k.Logger(ctx).Info(fmt.Sprintf("Annual params update: %s", params))
	}

	// calculate inflation power
	bondedRatio, lockedRatio := k.BondedRatio(ctx), k.LockedRatio(ctx)
	inflationPower := minter.NextInflationPower(params, bondedRatio, lockedRatio)

	// recalculate inflation
	minter.Inflation = minter.NextInflationRate(params, inflationPower)
	minter.FoundationInflation = minter.NextFoundationInflationRate(params)

	// burn fees
	k.BurnFeeCoins(ctx)

	// recalculate annual provisions
	totalStakingSupply := k.StakingTokenSupply(ctx)
	minter.Provisions, minter.FoundationProvisions = minter.NextAnnualProvisions(params, totalStakingSupply)
	k.SetMinter(ctx, minter)

	// mint coins, update supply
	mintedCoin := minter.BlockProvision(params)
	mintedCoins := sdk.NewCoins(mintedCoin)
	if err := k.MintCoins(ctx, mintedCoins); err != nil {
		panic(fmt.Errorf("minting coins: %v", err))
	}
	if err = k.TransferCoinsToFeeCollector(ctx, mintedCoins); err != nil {
		panic(fmt.Errorf("sending minted coint to the fee collector account: %v", err))
	}

	// TODO: review event content
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeMint,
			sdk.NewAttribute(types.AttributeKeyInflation, minter.Inflation.String()),
			sdk.NewAttribute(types.AttributeKeyAnnualProvisions, minter.Provisions.String()),
			sdk.NewAttribute(sdk.AttributeKeyAmount, mintedCoin.Amount.String()),
		),
	)
}
