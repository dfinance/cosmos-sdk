package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

type (
	SquashOptions struct {
		// Squash order: 1st
		addAccountOps []addAccountOperation
	}

	// New account operation
	addAccountOperation struct {
		// Account address
		Address sdk.AccAddress
		// Account balance
		Coins sdk.Coins
	}
)

func (opts *SquashOptions) SetAddAccountOp(addressRaw, coinsRaw string) error {
	op := addAccountOperation{}

	addr, err := sdk.AccAddressFromBech32(addressRaw)
	if err != nil {
		return fmt.Errorf("address (%s): invalid AccAddress: %w", addressRaw, err)
	}
	op.Address = addr

	coins, err := sdk.ParseCoins(coinsRaw)
	if err != nil {
		return fmt.Errorf("coins (%s): sdk.Coins parsing failed: %w", coinsRaw, err)
	}
	op.Coins = coins

	opts.addAccountOps = append(opts.addAccountOps, op)

	return nil
}

func NewEmptySquashOptions() SquashOptions {
	return SquashOptions{
		addAccountOps: nil,
	}
}

// PrepareForZeroHeight squashes current context state to fit zero-height (used on genesis export).
func (ak AccountKeeper) PrepareForZeroHeight(ctx sdk.Context, opts SquashOptions) error {
	for i, accOpt := range opts.addAccountOps {
		acc := ak.NewAccountWithAddress(ctx, accOpt.Address)
		if err := acc.SetCoins(accOpt.Coins); err != nil {
			return fmt.Errorf("addAccountOps[%d]: SetCoins: %w", i, err)
		}
		ak.SetAccount(ctx, acc)
	}

	return nil
}
