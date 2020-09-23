package keeper

import (
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/mint/internal/types"
)

// CheckAnnualParamsAdjust returns true if annual params should be adjusted as a new year has arrived.
// CheckAnnualParamsAdjust updates the nextAnnualUpdateTs if needed.
// CheckAnnualParamsAdjust is used by BeginBlock.
func (k Keeper) CheckAnnualParamsAdjust(ctx sdk.Context) bool {
	// set the initial year start (annual update) timestamp
	nextAnnualUpdateTs := k.GetAnnualUpdateTimestamp(ctx)
	if nextAnnualUpdateTs.IsZero() {
		k.setNextAnnualUpdateTimestamp(ctx)
		return false
	}

	// check if year has changed
	curTs := ctx.BlockTime()
	diff := curTs.Sub(nextAnnualUpdateTs)
	if diff > 0 {
		k.setNextAnnualUpdateTimestamp(ctx)
		return true
	}

	return false
}

// HasAnnualUpdateTimestamp checks if next annual update timestamp is set.
func (k Keeper) HasAnnualUpdateTimestamp(ctx sdk.Context) bool {
	store := ctx.KVStore(k.storeKey)
	return store.Has(types.AnnualUpdateTimestampKey)
}

// GetAnnualUpdateTimestamp reads current annual update timestamp (if exists).
func (k Keeper) GetAnnualUpdateTimestamp(ctx sdk.Context) time.Time {
	store := ctx.KVStore(k.storeKey)

	tBz := store.Get(types.AnnualUpdateTimestampKey)
	if tBz == nil {
		return time.Time{}
	}

	t, err := sdk.ParseTimeBytes(tBz)
	if err != nil {
		panic(fmt.Errorf("AnnualUpdateTimestamp parse: %w", err))
	}

	return t
}

// SetAnnualUpdateTimestamp sets next annual update timestamp.
func (k Keeper) SetAnnualUpdateTimestamp(ctx sdk.Context, timestamp time.Time) {
	store := ctx.KVStore(k.storeKey)

	tBz := sdk.FormatTimeBytes(timestamp)
	store.Set(types.AnnualUpdateTimestampKey, tBz)
}

// setNextAnnualUpdateTimestamp estimates and stores next annual update timestamp.
func (k Keeper) setNextAnnualUpdateTimestamp(ctx sdk.Context) {
	curTs := ctx.BlockTime()

	nextAnnualUpdateTs := curTs.AddDate(1, 0, 0)
	k.SetAnnualUpdateTimestamp(ctx, nextAnnualUpdateTs)

	k.Logger(ctx).Info(fmt.Sprintf("Next annual params update is set to: %v", nextAnnualUpdateTs))
}
