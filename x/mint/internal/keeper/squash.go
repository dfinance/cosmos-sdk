package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

type (
	SquashOptions struct {
		// Params modification operation
		paramsOps paramsOperation
	}

	paramsOperation struct {
		// Modify mint denom (empty - no modification)
		mintDenom string
	}
)

func (opts *SquashOptions) SetParamsOp(mintDenomRaw string) error {
	op := paramsOperation{}
	if mintDenomRaw != "" {
		if err := sdk.ValidateDenom(mintDenomRaw); err != nil {
			return fmt.Errorf("mintDenom (%s): invalid: %w", mintDenomRaw, err)
		}
		op.mintDenom = mintDenomRaw
	}
	opts.paramsOps = op

	return nil
}

func NewEmptySquashOptions() SquashOptions {
	return SquashOptions{
		paramsOps: paramsOperation{},
	}
}

// PrepareForZeroHeight squashes current context state to fit zero-height (used on genesis export).
func (k Keeper) PrepareForZeroHeight(ctx sdk.Context, opts SquashOptions) error {
	// paramsOps
	{
		if opts.paramsOps.mintDenom != "" {
			params := k.GetParams(ctx)
			params.MintDenom = opts.paramsOps.mintDenom
			k.SetParams(ctx, params)
		}
	}

	return nil
}
