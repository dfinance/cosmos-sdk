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

// GetAnnualUpdateTimestamp reads current annual update timestamp (if exists).
func (k Keeper) GetAnnualUpdateTimestamp(ctx sdk.Context) time.Time {
	store := ctx.KVStore(k.storeKey)

	tBz := store.Get(types.AnnualUpdateTimestampKey)
	if tBz == nil {
		return time.Time{}
	}

	t := time.Time{}
	k.cdc.MustUnmarshalBinaryLengthPrefixed(tBz, &t)

	return t
}

// setNextAnnualUpdateTimestamp estimates and stores next annual update timestamp.
func (k Keeper) setNextAnnualUpdateTimestamp(ctx sdk.Context) {
	store := ctx.KVStore(k.storeKey)
	curTs := ctx.BlockTime()

	nextAnnualUpdateTs := curTs.AddDate(1, 0, 0)
	nextAnnualUpdateBz := k.cdc.MustMarshalBinaryLengthPrefixed(nextAnnualUpdateTs)
	store.Set(types.AnnualUpdateTimestampKey, nextAnnualUpdateBz)

	k.Logger(ctx).Info(fmt.Sprintf("Next annual params update is set to: %v", nextAnnualUpdateTs))
}
