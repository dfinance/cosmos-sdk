package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

type (
	SquashOptions struct {
		// Denomination operation
		denomOps []denomOperation
	}

	denomOperation struct {
		// Denom symbol
		Denom string
		// Remove denom total supply
		// 1st priority
		Remove bool
		// Change supply (0 - no change, negative - decrease, positive - increase)
		// 2nd priority
		ShiftSupplyAmount sdk.Int
		// Rename coin denom (empty - no renaming)
		// 3rd priority
		RenameTo string
	}
)

func (opts *SquashOptions) SetDenomOp(
	denomRaw string,
	remove bool, renameToRaw string, shiftSupplyAmountRaw string,
) error {

	op := denomOperation{}
	op.Remove = remove

	if remove && (renameToRaw != "" || shiftSupplyAmountRaw != "0") {
		return fmt.Errorf("remove op can not coexist with rename/shift ops")
	}

	if err := sdk.ValidateDenom(denomRaw); err != nil {
		return fmt.Errorf("denom (%s): invalid: %w", denomRaw, err)
	}
	op.Denom = denomRaw

	if renameToRaw != "" {
		if err := sdk.ValidateDenom(renameToRaw); err != nil {
			return fmt.Errorf("renameTo denom (%s): invalid: %w", renameToRaw, err)
		}
		op.RenameTo = renameToRaw
	}

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
	supply := k.GetSupply(ctx)

	// verify all denoms do present
	for i, op := range opts.denomOps {
		found := false
		for _, coin := range supply.GetTotal() {
			if coin.Denom == op.Denom {
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("denomOp[%d] (%s): not found", i, op.Denom)
		}
	}

	// remove ops
	for _, op := range opts.denomOps {
		if !op.Remove {
			continue
		}

		total := supply.GetTotal()
		for _, coin := range total {
			if coin.Denom != op.Denom {
				continue
			}
			total = total.Sub(sdk.NewCoins(coin))
			break
		}
		supply.SetTotal(total)
	}

	// shift ops
	for _, op := range opts.denomOps {
		if op.ShiftSupplyAmount.IsZero() {
			continue
		}

		if op.ShiftSupplyAmount.IsNegative() {
			coin := sdk.NewCoin(op.Denom, op.ShiftSupplyAmount.MulRaw(-1))
			supply = supply.Deflate(sdk.NewCoins(coin))
		} else {
			coin := sdk.NewCoin(op.Denom, op.ShiftSupplyAmount)
			supply = supply.Inflate(sdk.NewCoins(coin))
		}
	}

	// rename ops
	for _, op := range opts.denomOps {
		if op.RenameTo == "" {
			continue
		}

		oldCoin := sdk.NewCoin(op.Denom, sdk.ZeroInt())
		total := supply.GetTotal()
		for _, coin := range total {
			if coin.Denom == op.Denom {
				oldCoin.Amount = coin.Amount
				break
			}
		}
		newCoin := sdk.NewCoin(op.RenameTo, oldCoin.Amount)

		total = total.Sub(sdk.NewCoins(oldCoin))
		total = total.Add(newCoin)
		supply = supply.SetTotal(total)
	}

	k.SetSupply(ctx, supply)

	return nil
}
