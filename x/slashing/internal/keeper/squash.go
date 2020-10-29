package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/slashing/internal/types"
)

// PrepareForZeroHeight squashes current context state to fit zero-height (used on genesis export).
func (k Keeper) PrepareForZeroHeight(ctx sdk.Context) error {
	// reset start height on signing infos
	k.IterateValidatorSigningInfos(
		ctx,
		func(addr sdk.ConsAddress, info types.ValidatorSigningInfo) (stop bool) {
			info.StartHeight = 0
			k.SetValidatorSigningInfo(ctx, addr, info)
			return false
		},
	)

	return nil
}
