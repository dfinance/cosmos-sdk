package types

import (
	"testing"

	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func TestNextMinMaxInflation(t *testing.T) {
	minter := DefaultInitialMinter()
	params := DefaultParams()

	tests := []struct {
		inMin, inMax, inInflation sdk.Dec
		outMin, outMax            sdk.Dec
	}{
		// increased max
		{
			sdk.NewDecWithPrec(10, 2),
			sdk.NewDecWithPrec(50, 2),
			sdk.NewDecWithPrec(5, 2),
			sdk.NewDecWithPrec(5, 2),
			sdk.NewDecWithPrec(70, 2),
		},
		// decreased max
		{
			sdk.NewDecWithPrec(4, 2),
			sdk.NewDecWithPrec(50, 2),
			sdk.NewDecWithPrec(45, 2),
			sdk.NewDecWithPrec(2, 2),
			sdk.NewDecWithPrec(30, 2),
		},
		// actualInflation reached its max
		{
			sdk.NewDecWithPrec(20, 2),
			sdk.NewDecWithPrec(50, 2),
			sdk.NewDecWithPrec(50, 2),
			sdk.NewDecWithPrec(10, 2),
			sdk.NewDecWithPrec(25, 2),
		},
	}

	for i, tc := range tests {
		params.InflationMin = tc.inMin
		params.InflationMax = tc.inMax
		minter.Inflation = tc.inInflation

		min, max := minter.NextMinMaxInflation(params)
		require.True(t, min.Equal(tc.outMin), "Test [%d]: minInflation diff: %s", i, min.Sub(tc.outMin))
		require.True(t, max.Equal(tc.outMax), "Test [%d]: maxInflation diff: %s", i, max.Sub(tc.outMax))
	}
}

func TestNextInflationRate(t *testing.T) {
	minter := DefaultInitialMinter()
	params := DefaultParams()

	tests := []struct {
		inMin, inMax, inPwr sdk.Dec
		outInf              sdk.Dec
	}{
		// zero min, max
		{
			sdk.ZeroDec(),
			sdk.ZeroDec(),
			sdk.NewDecWithPrec(1, 1),
			sdk.ZeroDec(),
		},
		// zero min
		{
			sdk.ZeroDec(),
			sdk.NewDecWithPrec(5, 1),
			sdk.NewDecWithPrec(3, 1),
			sdk.NewDecWithPrec(15, 2),
		},
		// zero inflationPower
		{
			sdk.NewDecWithPrec(2, 1),
			sdk.NewDecWithPrec(5, 1),
			sdk.ZeroDec(),
			sdk.NewDecWithPrec(2, 1),
		},
		// mixed
		{
			sdk.NewDecWithPrec(2, 1),
			sdk.NewDecWithPrec(5, 1),
			sdk.NewDecWithPrec(3, 1),
			sdk.NewDecWithPrec(29, 2),
		},
		// capped to max
		{
			sdk.NewDecWithPrec(2, 1),
			sdk.NewDecWithPrec(3, 1),
			sdk.NewDecWithPrec(15, 1),
			sdk.NewDecWithPrec(3, 1),
		},
	}

	for i, tc := range tests {
		params.InflationMin = tc.inMin
		params.InflationMax = tc.inMax

		inf := minter.NextInflationRate(params, tc.inPwr)
		require.True(t, inf.Equal(tc.outInf), "Test [%d]: diff: %s", i, inf.Sub(tc.outInf))
	}
}

func TestNextFoundationInflationRate(t *testing.T) {
	minter := DefaultInitialMinter()
	params := DefaultParams()

	tests := []struct {
		inMax, inAllocRatio, inInf sdk.Dec
		outFInf                    sdk.Dec
	}{
		// 1st min
		{
			sdk.NewDecWithPrec(5, 1),
			sdk.NewDecWithPrec(45, 2),
			sdk.NewDecWithPrec(4, 1),
			sdk.NewDecWithPrec(1, 1),
		},
		// 2nd min
		{
			sdk.NewDecWithPrec(6, 1),
			sdk.NewDecWithPrec(45, 2),
			sdk.NewDecWithPrec(2, 1),
			sdk.NewDecWithPrec(9, 2),
		},
	}

	for i, tc := range tests {
		params.InflationMax = tc.inMax
		params.FoundationAllocationRatio = tc.inAllocRatio
		minter.Inflation = tc.inInf

		fInf := minter.NextFoundationInflationRate(params)
		require.True(t, fInf.Equal(tc.outFInf), "Test [%d]: diff: %s", i, fInf.Sub(tc.outFInf))
	}
}

func TestNextAnnualProvisions(t *testing.T) {
	supply := sdk.NewInt(1000)
	minter := DefaultInitialMinter()
	minter.Inflation = sdk.NewDecWithPrec(2, 1)
	minter.FoundationInflation = sdk.NewDecWithPrec(5, 2)

	infProvision, foundationProvision := minter.NextAnnualProvisions(Params{}, supply)
	require.EqualValues(t, int64(200), infProvision.TruncateInt().Int64())
	require.EqualValues(t, int64(50), foundationProvision.TruncateInt().Int64())
}

func TestBlockProvision(t *testing.T) {
	params := DefaultParams()
	minter := DefaultInitialMinter()

	tests := []struct {
		inProv, inFProv sdk.Dec
		inBPY           uint64
		outCoinAmt      int64
	}{
		// 2.5 truncated
		{
			sdk.NewDecWithPrec(200, 0),
			sdk.NewDecWithPrec(50, 0),
			100,
			2,
		},
		// no truncation
		{
			sdk.NewDecWithPrec(200, 0),
			sdk.NewDecWithPrec(100, 0),
			100,
			3,
		},
		// zero provision
		{
			sdk.ZeroDec(),
			sdk.ZeroDec(),
			100,
			0,
		},
	}

	for i, tc := range tests {
		minter.Provisions = tc.inProv
		minter.FoundationProvisions = tc.inFProv
		minter.BlocksPerYear = tc.inBPY

		expCoin := sdk.NewCoin(params.MintDenom, sdk.NewInt(tc.outCoinAmt))
		outCoin := minter.BlockProvision(params)
		require.True(t, expCoin.IsEqual(outCoin), "Test [%d]", i)
	}
}

func TestNextInflationPower(t *testing.T) {
	// init base
	params := DefaultParams()
	minter := DefaultInitialMinter()
	params.InflationMin = sdk.NewDecWithPrec(1765, 4)
	params.InflationMax = sdk.NewDecWithPrec(5, 1)
	params.InfPwrBondedLockedRatio = sdk.NewDecWithPrec(4, 1)

	// test cases (surface sourced from MatLAB)
	testCases := []struct {
		BondedRatio  sdk.Dec
		LockedRatio  sdk.Dec
		ExpInflation sdk.Dec
	}{
		// Bonded: 0.0
		{
			BondedRatio:  sdk.NewDecWithPrec(0, 0),
			LockedRatio:  sdk.ZeroDec(),
			ExpInflation: sdk.NewDecWithPrec(305900, 6),
		},
		{
			BondedRatio:  sdk.NewDecWithPrec(0, 0),
			LockedRatio:  sdk.NewDecWithPrec(2, 1),
			ExpInflation: sdk.NewDecWithPrec(354425, 6),
		},
		{
			BondedRatio:  sdk.NewDecWithPrec(0, 0),
			LockedRatio:  sdk.NewDecWithPrec(4, 1),
			ExpInflation: sdk.NewDecWithPrec(402950, 6),
		},
		{
			BondedRatio:  sdk.NewDecWithPrec(0, 0),
			LockedRatio:  sdk.NewDecWithPrec(6, 1),
			ExpInflation: sdk.NewDecWithPrec(451475, 6),
		},
		{
			BondedRatio:  sdk.NewDecWithPrec(0, 0),
			LockedRatio:  sdk.NewDecWithPrec(8, 1),
			ExpInflation: sdk.NewDecWithPrec(500000, 6),
		},
		{
			BondedRatio:  sdk.NewDecWithPrec(0, 0),
			LockedRatio:  sdk.NewDecWithPrec(1, 0),
			ExpInflation: sdk.NewDecWithPrec(383540, 6),
		},
		// Bonded: 0.2
		{
			BondedRatio:  sdk.NewDecWithPrec(2, 1),
			LockedRatio:  sdk.NewDecWithPrec(0, 0),
			ExpInflation: sdk.NewDecWithPrec(289725, 6),
		},
		{
			BondedRatio:  sdk.NewDecWithPrec(2, 1),
			LockedRatio:  sdk.NewDecWithPrec(2, 1),
			ExpInflation: sdk.NewDecWithPrec(338250, 6),
		},
		{
			BondedRatio:  sdk.NewDecWithPrec(2, 1),
			LockedRatio:  sdk.NewDecWithPrec(4, 1),
			ExpInflation: sdk.NewDecWithPrec(386775, 6),
		},
		{
			BondedRatio:  sdk.NewDecWithPrec(2, 1),
			LockedRatio:  sdk.NewDecWithPrec(6, 1),
			ExpInflation: sdk.NewDecWithPrec(435300, 6),
		},
		{
			BondedRatio:  sdk.NewDecWithPrec(2, 1),
			LockedRatio:  sdk.NewDecWithPrec(8, 1),
			ExpInflation: sdk.NewDecWithPrec(483825, 6),
		},
		{
			BondedRatio:  sdk.NewDecWithPrec(2, 1),
			LockedRatio:  sdk.NewDecWithPrec(1, 0),
			ExpInflation: sdk.NewDecWithPrec(367365, 6),
		},
		// Bonded: 0.4
		{
			BondedRatio:  sdk.NewDecWithPrec(4, 1),
			LockedRatio:  sdk.NewDecWithPrec(0, 0),
			ExpInflation: sdk.NewDecWithPrec(273550, 6),
		},
		{
			BondedRatio:  sdk.NewDecWithPrec(4, 1),
			LockedRatio:  sdk.NewDecWithPrec(2, 1),
			ExpInflation: sdk.NewDecWithPrec(322075, 6),
		},
		{
			BondedRatio:  sdk.NewDecWithPrec(4, 1),
			LockedRatio:  sdk.NewDecWithPrec(4, 1),
			ExpInflation: sdk.NewDecWithPrec(370600, 6),
		},
		{
			BondedRatio:  sdk.NewDecWithPrec(4, 1),
			LockedRatio:  sdk.NewDecWithPrec(6, 1),
			ExpInflation: sdk.NewDecWithPrec(419125, 6),
		},
		{
			BondedRatio:  sdk.NewDecWithPrec(4, 1),
			LockedRatio:  sdk.NewDecWithPrec(8, 1),
			ExpInflation: sdk.NewDecWithPrec(467650, 6),
		},
		{
			BondedRatio:  sdk.NewDecWithPrec(4, 1),
			LockedRatio:  sdk.NewDecWithPrec(1, 0),
			ExpInflation: sdk.NewDecWithPrec(351190, 6),
		},
		// Bonded: 0.6
		{
			BondedRatio:  sdk.NewDecWithPrec(6, 1),
			LockedRatio:  sdk.NewDecWithPrec(0, 0),
			ExpInflation: sdk.NewDecWithPrec(257375, 6),
		},
		{
			BondedRatio:  sdk.NewDecWithPrec(6, 1),
			LockedRatio:  sdk.NewDecWithPrec(2, 1),
			ExpInflation: sdk.NewDecWithPrec(305900, 6),
		},
		{
			BondedRatio:  sdk.NewDecWithPrec(6, 1),
			LockedRatio:  sdk.NewDecWithPrec(4, 1),
			ExpInflation: sdk.NewDecWithPrec(354425, 6),
		},
		{
			BondedRatio:  sdk.NewDecWithPrec(6, 1),
			LockedRatio:  sdk.NewDecWithPrec(6, 1),
			ExpInflation: sdk.NewDecWithPrec(402950, 6),
		},
		{
			BondedRatio:  sdk.NewDecWithPrec(6, 1),
			LockedRatio:  sdk.NewDecWithPrec(8, 1),
			ExpInflation: sdk.NewDecWithPrec(451475, 6),
		},
		{
			BondedRatio:  sdk.NewDecWithPrec(6, 1),
			LockedRatio:  sdk.NewDecWithPrec(1, 0),
			ExpInflation: sdk.NewDecWithPrec(335015, 6),
		},
		// Bonded: 0.8
		{
			BondedRatio:  sdk.NewDecWithPrec(8, 1),
			LockedRatio:  sdk.NewDecWithPrec(0, 0),
			ExpInflation: sdk.NewDecWithPrec(241200, 6),
		},
		{
			BondedRatio:  sdk.NewDecWithPrec(8, 1),
			LockedRatio:  sdk.NewDecWithPrec(2, 1),
			ExpInflation: sdk.NewDecWithPrec(289725, 6),
		},
		{
			BondedRatio:  sdk.NewDecWithPrec(8, 1),
			LockedRatio:  sdk.NewDecWithPrec(4, 1),
			ExpInflation: sdk.NewDecWithPrec(338250, 6),
		},
		{
			BondedRatio:  sdk.NewDecWithPrec(8, 1),
			LockedRatio:  sdk.NewDecWithPrec(6, 1),
			ExpInflation: sdk.NewDecWithPrec(386775, 6),
		},
		{
			BondedRatio:  sdk.NewDecWithPrec(8, 1),
			LockedRatio:  sdk.NewDecWithPrec(8, 1),
			ExpInflation: sdk.NewDecWithPrec(435300, 6),
		},
		{
			BondedRatio:  sdk.NewDecWithPrec(8, 1),
			LockedRatio:  sdk.NewDecWithPrec(1, 0),
			ExpInflation: sdk.NewDecWithPrec(318840, 6),
		},
		// Bonded: 1.0
		{
			BondedRatio:  sdk.NewDecWithPrec(1, 0),
			LockedRatio:  sdk.NewDecWithPrec(0, 0),
			ExpInflation: sdk.NewDecWithPrec(176500, 6),
		},
		{
			BondedRatio:  sdk.NewDecWithPrec(1, 0),
			LockedRatio:  sdk.NewDecWithPrec(2, 1),
			ExpInflation: sdk.NewDecWithPrec(225025, 6),
		},
		{
			BondedRatio:  sdk.NewDecWithPrec(1, 0),
			LockedRatio:  sdk.NewDecWithPrec(4, 1),
			ExpInflation: sdk.NewDecWithPrec(273550, 6),
		},
		{
			BondedRatio:  sdk.NewDecWithPrec(1, 0),
			LockedRatio:  sdk.NewDecWithPrec(6, 1),
			ExpInflation: sdk.NewDecWithPrec(322075, 6),
		},
		{
			BondedRatio:  sdk.NewDecWithPrec(1, 0),
			LockedRatio:  sdk.NewDecWithPrec(8, 1),
			ExpInflation: sdk.NewDecWithPrec(370600, 6),
		},
		{
			BondedRatio:  sdk.NewDecWithPrec(1, 0),
			LockedRatio:  sdk.NewDecWithPrec(1, 0),
			ExpInflation: sdk.NewDecWithPrec(254140, 6),
		},
	}

	for i, testCase := range testCases {
		inflationPwr := minter.NextInflationPower(params, testCase.BondedRatio, testCase.LockedRatio)
		inflationCapped := minter.NextInflationRate(params, inflationPwr)

		require.True(t, testCase.ExpInflation.Equal(inflationCapped), "%d: BondedRatio / LockedRatio (%s / %s): expected / received: %s / %s", i, testCase.BondedRatio, testCase.LockedRatio, testCase.ExpInflation, inflationCapped)
	}
}
