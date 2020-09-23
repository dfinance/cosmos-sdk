package mint

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// InitGenesis new mint genesis
func InitGenesis(ctx sdk.Context, keeper Keeper, data GenesisState) {
	keeper.SetMinter(ctx, data.Minter)
	keeper.SetParams(ctx, data.Params)
	if len(data.BlockDurFilter.Values) != 0 {
		keeper.SetBlockDurFilter(ctx, data.BlockDurFilter)
	}
	if !data.AnnualUpdateTS.IsZero() {
		keeper.SetAnnualUpdateTimestamp(ctx, data.AnnualUpdateTS)
	}
}

// ExportGenesis returns a GenesisState for a given context and keeper.
func ExportGenesis(ctx sdk.Context, keeper Keeper) GenesisState {
	minter := keeper.GetMinter(ctx)
	params := keeper.GetParams(ctx)
	blockDurFilter := BlockDurFilter{}
	if filter := keeper.GetBlockDurFilter(ctx); filter != nil {
		blockDurFilter = *filter
	}
	annualUpdateTS := time.Time{}
	if keeper.HasAnnualUpdateTimestamp(ctx) {
		annualUpdateTS = keeper.GetAnnualUpdateTimestamp(ctx)
	}

	return NewGenesisState(minter, params, blockDurFilter, annualUpdateTS)
}
