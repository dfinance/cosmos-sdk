package simulation

// DONTCOVER

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/cosmos/cosmos-sdk/codec"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/x/mint/internal/types"
)

// Simulation parameter constants
const (
	Inflation               = "inflation"
	InflationMax            = "inflation_max"
	InflationMin            = "inflation_min"
	FeeBurningRatio         = "fee_burning_ratio"
	InfPwrBondedLockedRatio = "infpwr_bondedlocked_ratio"
)

// GenInflation randomized Inflation
func GenInflation(r *rand.Rand) sdk.Dec {
	return sdk.NewDecWithPrec(int64(r.Intn(99)), 2)
}

// GenInflationMax randomized InflationMax
func GenInflationMax(r *rand.Rand) sdk.Dec {
	return sdk.NewDecWithPrec(50, 2)
}

// GenInflationMin randomized InflationMin
func GenInflationMin(r *rand.Rand) sdk.Dec {
	return sdk.NewDecWithPrec(1776, 4)
}

// GenFeeBurningRatio randomized FeeBurningRatio
func GenFeeBurningRatio(r *rand.Rand) sdk.Dec {
	return sdk.NewDecWithPrec(50, 2)
}

// GenInfPwrBondedLockedRatio randomized InfPwrBondedLockedRatio
func GenInfPwrBondedLockedRatio(r *rand.Rand) sdk.Dec {
	return sdk.NewDecWithPrec(4, 1)
}

// RandomizedGenState generates a random GenesisState for mint
func RandomizedGenState(simState *module.SimulationState) {
	// minter
	var inflation sdk.Dec
	simState.AppParams.GetOrGenerate(
		simState.Cdc, Inflation, &inflation, simState.Rand,
		func(r *rand.Rand) { inflation = GenInflation(r) },
	)

	// params
	mintDenom := sdk.DefaultBondDenom

	var inflationMax sdk.Dec
	simState.AppParams.GetOrGenerate(
		simState.Cdc, InflationMax, &inflationMax, simState.Rand,
		func(r *rand.Rand) { inflationMax = GenInflationMax(r) },
	)

	var inflationMin sdk.Dec
	simState.AppParams.GetOrGenerate(
		simState.Cdc, InflationMin, &inflationMin, simState.Rand,
		func(r *rand.Rand) { inflationMin = GenInflationMin(r) },
	)

	var feeBurningRatio sdk.Dec
	simState.AppParams.GetOrGenerate(
		simState.Cdc, FeeBurningRatio, &feeBurningRatio, simState.Rand,
		func(r *rand.Rand) { feeBurningRatio = GenFeeBurningRatio(r) },
	)

	var infPwrBondedLockedRatio sdk.Dec
	simState.AppParams.GetOrGenerate(
		simState.Cdc, InfPwrBondedLockedRatio, &infPwrBondedLockedRatio, simState.Rand,
		func(r *rand.Rand) { infPwrBondedLockedRatio = GenInfPwrBondedLockedRatio(r) },
	)

	foundationAllocationRatio := sdk.NewDecWithPrec(45, 2)
	avgBlocksTimeWindow := uint16(2)

	params := types.NewParams(
		mintDenom,
		inflationMax,
		inflationMin,
		feeBurningRatio,
		infPwrBondedLockedRatio,
		foundationAllocationRatio,
		avgBlocksTimeWindow,
		sdk.ZeroInt(),
	)

	mintGenesis := types.NewGenesisState(types.InitialMinter(inflation), params, types.BlockDurFilter{}, time.Time{})

	fmt.Printf("Selected randomly generated minting parameters:\n%s\n", codec.MustMarshalJSONIndent(simState.Cdc, mintGenesis))
	simState.GenState[types.ModuleName] = simState.Cdc.MustMarshalJSON(mintGenesis)
}
