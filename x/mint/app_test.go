package mint

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	authExported "github.com/cosmos/cosmos-sdk/x/auth/exported"
	"github.com/cosmos/cosmos-sdk/x/mint/internal/types"
	"github.com/cosmos/cosmos-sdk/x/mock"
)

// Check that inflation is not changed until filter window is filled up.
func TestBlockDurEstimation(t *testing.T) {
	mApp, _, stakingKeeper, mintKeeper := getMockApp(t, DefaultParams())

	// create accounts
	genCoin := sdk.NewCoin(sdk.DefaultBondDenom, sdk.TokensFromConsensusPower(42))
	acc1 := &auth.BaseAccount{
		Address: addr1,
		Coins:   sdk.Coins{genCoin},
	}

	mock.SetGenesis(mApp, []authExported.Account{acc1})
	mock.CheckBalance(t, mApp, addr1, sdk.Coins{genCoin})

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

// Check fees are burned on inflation recalculation.
func TestFeesBurning(t *testing.T) {
	// alter params to disable inflation change
	mintParams := DefaultParams()
	mintParams.InflationMin = sdk.ZeroDec()
	mintParams.InfPwrBondedLockedRatio = sdk.ZeroDec()

	mApp, supplyKeeper, stakingKeeper, _ := getMockApp(t, mintParams)

	// create accounts
	genCoin := sdk.NewCoin(sdk.DefaultBondDenom, sdk.TokensFromConsensusPower(42))
	acc1 := &auth.BaseAccount{
		Address: addr1,
		Coins:   sdk.Coins{genCoin},
	}

	mock.SetGenesis(mApp, []authExported.Account{acc1})
	mock.CheckBalance(t, mApp, addr1, sdk.Coins{genCoin})

	getCurFees := func() sdk.Int {
		return supplyKeeper.GetModuleAccount(getCheckCtx(mApp), auth.FeeCollectorName).GetCoins().AmountOf(sdk.DefaultBondDenom)
	}

	// initial fees
	curFees := getCurFees()
	t.Logf("[0] fees: %s", curFees)
	require.True(t, curFees.IsZero())

	// create a validator
	bondCoin := sdk.NewCoin(sdk.DefaultBondDenom, sdk.TokensFromConsensusPower(10))
	createValidator(t, mApp, stakingKeeper, addr1, priv1, bondCoin)

	// emit inflation change
	for i := 0; i < 5; i++ {
		skipBlock(mApp)

		prevFees := curFees
		curFees = getCurFees()
		t.Logf("[%d] fees: %s -> %s", i+1, prevFees, curFees)

		if i > 0 {
			// check decreased in time
			burnAmt := sdk.NewDecFromInt(prevFees).Mul(mintParams.FeeBurningRatio).TruncateInt()
			require.Equal(t,
				prevFees.Sub(burnAmt).Int64(),
				curFees.Int64(),
			)
		}
	}
}

// Check annual params recalculation triggered.
func TestNewYearCatch(t *testing.T) {
	mApp, _, stakingKeeper, mintKeeper := getMockApp(t, DefaultParams())

	// create accounts
	genCoin := sdk.NewCoin(sdk.DefaultBondDenom, sdk.TokensFromConsensusPower(42))
	acc1 := &auth.BaseAccount{
		Address: addr1,
		Coins:   sdk.Coins{genCoin},
	}

	mock.SetGenesis(mApp, []authExported.Account{acc1})
	mock.CheckBalance(t, mApp, addr1, sdk.Coins{genCoin})

	// create a validator
	bondCoin := sdk.NewCoin(sdk.DefaultBondDenom, sdk.TokensFromConsensusPower(10))
	createValidator(t, mApp, stakingKeeper, addr1, priv1, bondCoin)

	// emulate one short block to fill the avgBlocksPerYear filter
	skipBlock(mApp)

	// get init params
	initParams := mintKeeper.GetParams(getCheckCtx(mApp))

	// rough block monthly skipper
	curBlockTime := time.Date(2000, 1, 2, 0, 0, 0, 0, time.UTC)
	skipMonth := func() {
		curBlockTime = curBlockTime.AddDate(0 , 1, 0)

		mApp.BeginBlock(abci.RequestBeginBlock{Header: getNextABCIHeaderWithTime(mApp, curBlockTime)})
		mApp.EndBlock(abci.RequestEndBlock{})
		mApp.Commit()
	}

	// emulate less than a year
	prevInflation := mintKeeper.GetMinter(getCheckCtx(mApp)).Inflation
	for i := 0; i < 11; i++ {
		skipMonth()

		// check params not changed
		curParams := mintKeeper.GetParams(getCheckCtx(mApp))
		require.Equal(t, initParams.InflationMin.String(), curParams.InflationMin.String(), "month [%d]", i+1)
		require.Equal(t, initParams.InflationMax.String(), curParams.InflationMax.String(), "month [%d]", i+1)

		prevInflation = mintKeeper.GetMinter(getCheckCtx(mApp)).Inflation
	}

	// params should change now
	prevUpdateTs := mintKeeper.GetAnnualUpdateTimestamp(getCheckCtx(mApp))
	skipMonth()
	curParams := mintKeeper.GetParams(getCheckCtx(mApp))
	require.NotEqual(t, initParams.InflationMin.String(), curParams.InflationMin.String(), "month [12]")
	require.NotEqual(t, initParams.InflationMax.String(), curParams.InflationMax.String(), "month [12]")

	// check min, max change
	curMin, curMax := curParams.InflationMin, curParams.InflationMax
	expectedMin := initParams.InflationMin.QuoInt64(2)
	expectedMax := initParams.InflationMax.QuoInt64(2).Add(initParams.InflationMax.Sub(prevInflation))

	t.Logf("Min change: %s -> %s", initParams.InflationMin, curMin)
	t.Logf("Max change (inf: %s): %s -> %s", prevInflation, initParams.InflationMax, curMax)
	require.True(t, expectedMin.Equal(curMin))
	require.True(t, expectedMax.Equal(curMax))

	// check nextAnnualUpdateTs has changed
	nextUpdateTs := mintKeeper.GetAnnualUpdateTimestamp(getCheckCtx(mApp))
	require.False(t, nextUpdateTs.IsZero())
	require.True(t, nextUpdateTs.After(prevUpdateTs))
	require.True(t, nextUpdateTs.Sub(prevUpdateTs) > types.AvgYearDur)
}
