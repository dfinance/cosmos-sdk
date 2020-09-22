package keeper

import (
	"fmt"

	"github.com/tendermint/tendermint/libs/log"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/mint/internal/types"
	"github.com/cosmos/cosmos-sdk/x/params"
)

// Keeper of the mint store
type Keeper struct {
	cdc              *codec.Codec
	storeKey         sdk.StoreKey
	paramSpace       params.Subspace
	sk               types.StakingKeeper
	supplyKeeper     types.SupplyKeeper
	feeCollectorName string
}

// NewKeeper creates a new mint Keeper instance
func NewKeeper(
	cdc *codec.Codec, key sdk.StoreKey, paramSpace params.Subspace,
	sk types.StakingKeeper, supplyKeeper types.SupplyKeeper, feeCollectorName string,
) Keeper {

	// ensure mint module account is set
	if addr := supplyKeeper.GetModuleAddress(types.ModuleName); addr == nil {
		panic("the mint module account has not been set")
	}

	return Keeper{
		cdc:              cdc,
		storeKey:         key,
		paramSpace:       paramSpace.WithKeyTable(types.ParamKeyTable()),
		sk:               sk,
		supplyKeeper:     supplyKeeper,
		feeCollectorName: feeCollectorName,
	}
}

// Logger returns a module-specific logger.
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

// MintCoins implements an alias call to the underlying supply keeper's
// MintCoins to be used in BeginBlocker.
func (k Keeper) MintCoins(ctx sdk.Context, newCoins sdk.Coins) error {
	if newCoins.Empty() {
		// skip as no coins need to be minted
		return nil
	}

	return k.supplyKeeper.MintCoins(ctx, types.ModuleName, newCoins)
}

// BurnFeeCoins burns collected fees withing FeeCollector pool by params.FeeBurningRatio.
func (k Keeper) BurnFeeCoins(ctx sdk.Context) {
	params := k.GetParams(ctx)
	mintDenom, burnRatio := params.MintDenom, params.FeeBurningRatio

	// calculate the burning amount
	feesCollected := k.supplyKeeper.GetModuleAccount(ctx, k.feeCollectorName).GetCoins()
	feesBurnAmt := sdk.NewDecFromInt(feesCollected.AmountOf(mintDenom)).Mul(burnRatio).TruncateInt()
	burnCoin := sdk.NewCoin(mintDenom, feesBurnAmt)

	// burn
	err := k.supplyKeeper.BurnCoins(ctx, k.feeCollectorName, sdk.NewCoins(burnCoin))
	if err != nil {
		panic(fmt.Errorf("burning fees %s for %s: %v", burnCoin.String(), k.feeCollectorName, err))
	}
}

// TransferCoinsToFeeCollector transfers coins from the Mint to the FeeCollector module account.
func (k Keeper) TransferCoinsToFeeCollector(ctx sdk.Context, coins sdk.Coins) error {
	return k.supplyKeeper.SendCoinsFromModuleToModule(ctx, types.ModuleName, k.feeCollectorName, coins)
}
