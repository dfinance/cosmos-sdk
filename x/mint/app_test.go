package mint

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/crypto/secp256k1"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	authExported "github.com/cosmos/cosmos-sdk/x/auth/exported"
	"github.com/cosmos/cosmos-sdk/x/bank"
	"github.com/cosmos/cosmos-sdk/x/mint/internal/types"
	"github.com/cosmos/cosmos-sdk/x/mock"
	"github.com/cosmos/cosmos-sdk/x/staking"
	"github.com/cosmos/cosmos-sdk/x/supply"
	supplyExported "github.com/cosmos/cosmos-sdk/x/supply/exported"
)

var (
	maccPerms = map[string][]string{
		auth.FeeCollectorName:     nil,
		staking.NotBondedPoolName: {supply.Burner, supply.Staking},
		staking.BondedPoolName:    {supply.Burner, supply.Staking},
		types.ModuleName:          {supply.Minter},
	}

	priv1 = secp256k1.GenPrivKey()
	addr1 = sdk.AccAddress(priv1.PubKey().Address())
	priv2 = secp256k1.GenPrivKey()
	addr2 = sdk.AccAddress(priv2.PubKey().Address())
)

// getMockApp returns an initialized mock application for this module.
func getMockApp(t *testing.T) (*mock.App, staking.Keeper, Keeper) {
	mApp := mock.NewApp()

	supply.RegisterCodec(mApp.Cdc)
	staking.RegisterCodec(mApp.Cdc)

	keyStaking := sdk.NewKVStoreKey(staking.StoreKey)
	keySupply := sdk.NewKVStoreKey(supply.StoreKey)
	keyMint := sdk.NewKVStoreKey(StoreKey)

	feeCollector := supply.NewEmptyModuleAccount(auth.FeeCollectorName, maccPerms[auth.FeeCollectorName]...)
	notBondedPool := supply.NewEmptyModuleAccount(staking.NotBondedPoolName, maccPerms[staking.NotBondedPoolName]...)
	bondPool := supply.NewEmptyModuleAccount(staking.BondedPoolName, maccPerms[staking.BondedPoolName]...)
	minter := supply.NewEmptyModuleAccount(types.ModuleName, maccPerms[types.ModuleName]...)

	blacklistedAddrs := map[string]bool{
		feeCollector.GetAddress().String():  true,
		notBondedPool.GetAddress().String(): true,
		bondPool.GetAddress().String():      true,
		minter.GetAddress().String():        true,
	}

	bankKeeper := bank.NewBaseKeeper(mApp.AccountKeeper, mApp.ParamsKeeper.Subspace(bank.DefaultParamspace), blacklistedAddrs)
	supplyKeeper := supply.NewKeeper(mApp.Cdc, keySupply, mApp.AccountKeeper, bankKeeper, maccPerms)
	stakingKeeper := staking.NewKeeper(mApp.Cdc, keyStaking, supplyKeeper, mApp.ParamsKeeper.Subspace(staking.DefaultParamspace))
	keeper := NewKeeper(mApp.Cdc, keyMint, mApp.ParamsKeeper.Subspace(DefaultParamspace), &stakingKeeper, supplyKeeper, auth.FeeCollectorName)

	mApp.Router().AddRoute(staking.RouterKey, staking.NewHandler(stakingKeeper))
	mApp.SetBeginBlocker(getBeginBlocker(keeper))
	mApp.SetEndBlocker(getEndBlocker(stakingKeeper))
	mApp.SetInitChainer(getInitChainer(mApp, mApp.AccountKeeper, supplyKeeper, stakingKeeper, keeper,
		[]supplyExported.ModuleAccountI{feeCollector, notBondedPool, bondPool, minter}))

	require.NoError(t, mApp.CompleteSetup(keyStaking, keySupply, keyMint))

	return mApp, stakingKeeper, keeper
}

// getBeginBlocker returns a mint beginBlocker.
func getBeginBlocker(mintKeeper Keeper) sdk.BeginBlocker {
	return func(ctx sdk.Context, req abci.RequestBeginBlock) abci.ResponseBeginBlock {
		ctx = ctx.WithEventManager(sdk.NewEventManager())

		BeginBlocker(ctx, mintKeeper)

		return abci.ResponseBeginBlock{
			Events: ctx.EventManager().ABCIEvents(),
		}
	}
}

// getEndBlocker returns a staking endBlocker.
func getEndBlocker(stakingKeeper staking.Keeper) sdk.EndBlocker {
	return func(ctx sdk.Context, req abci.RequestEndBlock) abci.ResponseEndBlock {
		validatorUpdates := staking.EndBlocker(ctx, stakingKeeper)

		return abci.ResponseEndBlock{ValidatorUpdates: validatorUpdates}
	}
}

// getInitChainer initializes the chainer of the mock app and sets the genesis state.
// It returns an empty ResponseInitChain.
func getInitChainer(mApp *mock.App,
	accountKeeper auth.AccountKeeper, supplyKeeper supply.Keeper, stakingKeeper staking.Keeper, mintKeeper Keeper,
	blacklistedAddrs []supplyExported.ModuleAccountI,
) sdk.InitChainer {
	return func(ctx sdk.Context, req abci.RequestInitChain) abci.ResponseInitChain {
		mApp.InitChainer(ctx, req)

		for _, macc := range blacklistedAddrs {
			supplyKeeper.SetModuleAccount(ctx, macc)
		}

		supplyGenesis := supply.DefaultGenesisState()
		supply.InitGenesis(ctx, supplyKeeper, accountKeeper, supplyGenesis)

		stakingGenesis := staking.DefaultGenesisState()
		validators := staking.InitGenesis(ctx, stakingKeeper, accountKeeper, supplyKeeper, stakingGenesis)

		mintGenesis := DefaultGenesisState()
		mintGenesis.Params.AvgBlockTimeWindow = 2
		InitGenesis(ctx, mintKeeper, mintGenesis)

		return abci.ResponseInitChain{Validators: validators}
	}
}

// getCheckCtx returns CheckTx context.
func getCheckCtx(mApp *mock.App) sdk.Context {
	return mApp.BaseApp.NewContext(true, abci.Header{})
}

// getNextABCIHeader returns next ABCI header with shifted blockTime.
func getNextABCIHeader(mApp *mock.App) abci.Header {
	startTime := time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)
	nextHeight := mApp.LastBlockHeight() + 1

	return abci.Header{
		Height: nextHeight,
		Time: startTime.Add(time.Duration(nextHeight) * 5 * time.Second),
	}
}

// skipBlock emulates noop block.
func skipBlock(mApp *mock.App) {
	mApp.BeginBlock(abci.RequestBeginBlock{Header: getNextABCIHeader(mApp)})
	mApp.EndBlock(abci.RequestEndBlock{})
	mApp.Commit()
}

// createValidator creates a new validator for account operator.
func createValidator(t *testing.T, mApp *mock.App, keeper staking.Keeper, accAddr sdk.AccAddress, accPrvKey secp256k1.PrivKeySecp256k1, bondCoin sdk.Coin) {
	commissionRates := staking.NewCommissionRates(sdk.ZeroDec(), sdk.ZeroDec(), sdk.ZeroDec())
	description := staking.NewDescription("foo_moniker", "", "", "", "")
	valAddr := sdk.ValAddress(accAddr)
	createValidatorMsg := staking.NewMsgCreateValidator(valAddr, accPrvKey.PubKey(), bondCoin, description, commissionRates, sdk.OneInt())

	acc := mApp.AccountKeeper.GetAccount(getCheckCtx(mApp), accAddr)

	mock.SignCheckDeliver(t, mApp.Cdc, mApp.BaseApp, getNextABCIHeader(mApp), []sdk.Msg{createValidatorMsg}, []uint64{0}, []uint64{0}, true, true, accPrvKey)

	mock.CheckBalance(t, mApp, addr1, acc.GetCoins().Sub(sdk.Coins{bondCoin}))

	_, valFound := keeper.GetValidator(getCheckCtx(mApp), valAddr)
	require.True(t, valFound)
}

// Check that inflation is not changed until filter window is filled up.
func TestBlockDurEstimation(t *testing.T) {
	mApp, stakingKeeper, mintKeeper := getMockApp(t)

	// create accounts
	genCoin := sdk.NewCoin(sdk.DefaultBondDenom, sdk.TokensFromConsensusPower(42))
	acc1 := &auth.BaseAccount{
		Address: addr1,
		Coins:   sdk.Coins{genCoin},
	}
	acc2 := &auth.BaseAccount{
		Address: addr2,
		Coins:   sdk.Coins{genCoin},
	}
	accs := []authExported.Account{acc1, acc2}

	mock.SetGenesis(mApp, accs)
	mock.CheckBalance(t, mApp, addr1, sdk.Coins{genCoin})
	mock.CheckBalance(t, mApp, addr2, sdk.Coins{genCoin})

	// genesis inflation
	initInflation := mintKeeper.GetMinter(getCheckCtx(mApp)).Inflation
	t.Logf("Init inf: %s", initInflation.String())

	// create a validator
	bondCoin := sdk.NewCoin(sdk.DefaultBondDenom, sdk.TokensFromConsensusPower(10))
	createValidator(t, mApp, stakingKeeper, addr1, priv1, bondCoin)

	// check inflation didn't change (window is 2, we have to wait for one more block)
	curInflation := mintKeeper.GetMinter(getCheckCtx(mApp)).Inflation
	require.Equal(t, initInflation.String(), curInflation.String())

	// check inflation changed (window is full)
	for i := 0; i < 2; i++ {
		skipBlock(mApp)
		curInflation = mintKeeper.GetMinter(getCheckCtx(mApp)).Inflation
		require.NotEqual(t, initInflation.String(), curInflation.String())
	}
}
