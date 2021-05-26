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
		// Modify DepositParams.MinDeposit coin (empty - no modification)
		MinDeposit sdk.Coin
	}
)

func (opts *SquashOptions) SetParamsOp(minDepositCoinRaw string) error {
	op := paramsOperation{}
	if minDepositCoinRaw != "" {
		coin, err := sdk.ParseCoin(minDepositCoinRaw)
		if err != nil {
			return fmt.Errorf("minDepositCoin (%s): invalid: %w", minDepositCoinRaw, err)
		}
		op.MinDeposit = coin
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
func (keeper Keeper) PrepareForZeroHeight(ctx sdk.Context, opts SquashOptions) error {
	// paramsOps
	{
		if opts.paramsOps.MinDeposit.Denom != "" {
			params := keeper.GetDepositParams(ctx)
			params.MinDeposit = sdk.NewCoins(opts.paramsOps.MinDeposit)
			keeper.SetDepositParams(ctx, params)
		}
	}

	return nil
}
