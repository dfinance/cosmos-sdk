package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/mint/internal/types"
)

// AdjustAvgBLockDurEstimation adds a new blockTime tick to the BlockDurFilter.
func (k Keeper) AdjustAvgBLockDurEstimation(ctx sdk.Context) {
	params := k.GetParams(ctx)

	filter := k.getBlockDurFilter(ctx)
	if filter == nil {
		newFilter := types.NewBlockDurFilter(params.AvgBlockTimeWindow)
		filter = &newFilter
	}

	filter.Push(ctx.BlockTime(), params.AvgBlockTimeWindow)
	k.setBlockDurFilter(ctx, *filter)
}

// GetAvgBlocksPerYear returns blocksPerYear estimation if estimation can be calculated.
// BlockDurFilter has to be filled up in order to get the estimation.
func (k Keeper) GetAvgBlocksPerYear(ctx sdk.Context) (uint64, error) {
	filter := k.getBlockDurFilter(ctx)
	if filter == nil {
		return 0, fmt.Errorf("filter is not created")
	}

	params := k.GetParams(ctx)

	return filter.GetBlocksPerYear(params.AvgBlockTimeWindow)
}

// getBlockDurFilter gets the BlockDurFilter from the storage if exists.
func (k Keeper) getBlockDurFilter(ctx sdk.Context) *types.BlockDurFilter {
	store := ctx.KVStore(k.storeKey)
	if !store.Has(types.BlockDurFilterKey) {
		return nil
	}

	bz := store.Get(types.BlockDurFilterKey)
	filter := types.BlockDurFilter{}
	k.cdc.MustUnmarshalBinaryLengthPrefixed(bz, &filter)

	return &filter
}

// setBlockDurFilter sets the BlockDurFilter to the storage.
func (k Keeper) setBlockDurFilter(ctx sdk.Context, filter types.BlockDurFilter) {
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshalBinaryLengthPrefixed(filter)
	store.Set(types.BlockDurFilterKey, bz)
}
