package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/staking/types"
)

// IsAccountBanned checks if account is banned (all staking ops are denied).
func (k Keeper) IsAccountBanned(ctx sdk.Context, accAddr sdk.AccAddress) bool {
	store := ctx.KVStore(k.storeKey)
	key := types.GetBannedAccKey(accAddr)

	return store.Has(key)
}

// BanAccount bans an account to prevent all staking ops from it.
func (k Keeper) BanAccount(ctx sdk.Context, accAddr sdk.AccAddress, banHeight int64) {
	store := ctx.KVStore(k.storeKey)
	key := types.GetBannedAccKey(accAddr)

	info := types.BannedAccInfo{
		Height: banHeight,
	}

	bz := k.cdc.MustMarshalBinaryLengthPrefixed(info)
	store.Set(key, bz)
}

// UnbanAccount removes an account ban allowing all staking ops for it.
func (k Keeper) UnbanAccount(ctx sdk.Context, accAddr sdk.AccAddress) {
	store := ctx.KVStore(k.storeKey)
	key := types.GetBannedAccKey(accAddr)
	store.Delete(key)
}

// IterateScheduledUnbondQueue iterates over ScheduledUnbondQueue.
func (k Keeper) IterateBannedAccounts(ctx sdk.Context, handler func(accAddr sdk.AccAddress, banHeight int64) (stop bool)) {
	store := ctx.KVStore(k.storeKey)
	iterator := sdk.KVStorePrefixIterator(store, types.BannedAccKey)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		accAddr := types.ParseBannedAccKey(iterator.Key())
		var banHeight int64
		k.cdc.MustUnmarshalBinaryLengthPrefixed(iterator.Value(), &banHeight)

		if handler(accAddr, banHeight) {
			break
		}
	}
}
