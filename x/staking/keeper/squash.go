package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/staking/types"
)

// PrepareForZeroHeight squashes current context state to fit zero-height (used on genesis export).
func (k Keeper) PrepareForZeroHeight(ctx sdk.Context) error {
	// reset banned accounts height
	k.IterateBannedAccounts(ctx, func(accAddr sdk.AccAddress, _ int64) (stop bool) {
		k.BanAccount(ctx, accAddr, 0)
		return false
	})

	// reset redelegations  creation height
	k.IterateRedelegations(ctx, func(_ int64, red types.Redelegation) (stop bool) {
		for i := range red.Entries {
			red.Entries[i].CreationHeight = 0
		}
		k.SetRedelegation(ctx, red)
		return false
	})

	// reset unbonding delegations creation height
	k.IterateUnbondingDelegations(ctx, func(_ int64, ubd types.UnbondingDelegation) (stop bool) {
		for i := range ubd.Entries {
			ubd.Entries[i].CreationHeight = 0
		}
		k.SetUnbondingDelegation(ctx, ubd)
		return false
	})

	// reset validators bond height and scheduled unbond height
	for _, val := range k.GetAllValidators(ctx) {
		val.UnbondingHeight = 0
		val.ScheduledUnbondHeight = 0
		k.SetValidator(ctx, val)
	}

	_ = k.ApplyAndReturnValidatorSetUpdates(ctx)

	return nil
}
