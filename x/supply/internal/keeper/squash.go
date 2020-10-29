package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

type (
	SquashOptions struct {
		denomOps []denomOperation
	}

	// Denomination operation
	denomOperation struct {
		// Denom symbol
		Denom string
		// Change supply (0 - no change, negative - decrease, positive - increase)
		ShiftSupplyAmount sdk.Int
	}
)

func (opts *SquashOptions) SetDenomOp(denomRaw, shiftSupplyAmountRaw string) error {
	op := denomOperation{}

	if err := sdk.ValidateDenom(denomRaw); err != nil {
		return fmt.Errorf("denom (%s): invalid: %w", denomRaw, err)
	}
	op.Denom = denomRaw

	shiftSupplyAmount, ok := sdk.NewIntFromString(shiftSupplyAmountRaw)
	if !ok {
		return fmt.Errorf("shiftSupplyAmount (%s): invalid sdk.Int", shiftSupplyAmountRaw)
	}
	op.ShiftSupplyAmount = shiftSupplyAmount

	opts.denomOps = append(opts.denomOps, op)

	return nil
}

func NewEmptySquashOptions() SquashOptions {
	return SquashOptions{
		denomOps: nil,
	}
}

// PrepareForZeroHeight squashes current context state to fit zero-height (used on genesis export).
func (k Keeper) PrepareForZeroHeight(ctx sdk.Context, opts SquashOptions) error {
	for _, denomOpt := range opts.denomOps {
		if !denomOpt.ShiftSupplyAmount.IsZero() {
			supply := k.GetSupply(ctx)
			if denomOpt.ShiftSupplyAmount.IsNegative() {
				coin := sdk.NewCoin(denomOpt.Denom, denomOpt.ShiftSupplyAmount.MulRaw(-1))
				supply = supply.Deflate(sdk.NewCoins(coin))
			} else {
				coin := sdk.NewCoin(denomOpt.Denom, denomOpt.ShiftSupplyAmount)
				supply = supply.Inflate(sdk.NewCoins(coin))
			}
			k.SetSupply(ctx, supply)
		}
	}

	return nil
}
