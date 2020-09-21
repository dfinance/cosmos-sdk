package simulation

// DONTCOVER

import (
	"fmt"
	"math/rand"

	"github.com/cosmos/cosmos-sdk/codec"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/x/distribution/types"
)

// Simulation parameter constants
const (
	PublicTreasuryCapacity = "public_treasury_pool_capacity"
	BaseProposerReward     = "base_proposer_reward"
	BonusProposerReward    = "bonus_proposer_reward"
	WithdrawEnabled        = "withdraw_enabled"
)

// GetTaxes randomized taxes (sum must be 1.0)
func GetTaxes(r *rand.Rand) (validatorsTax, liquidityProvidersTax, publicTreasuryTax, harpTax sdk.Dec) {
	// 10% + rand(40%)
	validatorsTax = sdk.NewDecWithPrec(1, 1).Add(sdk.NewDecWithPrec(int64(r.Intn(40)), 2))
	liquidityProvidersTax = sdk.NewDecWithPrec(1, 1).Add(sdk.NewDecWithPrec(int64(r.Intn(40)), 2))

	// rand(100% - validatorsTax - liquidityProvidersTax)
	publicTreasuryTax = sdk.ZeroDec()
	publicTreasuryTaxRandLimit := sdk.OneDec().Sub(validatorsTax).Sub(liquidityProvidersTax)
	if publicTreasuryTaxRandLimit.GT(sdk.ZeroDec()) {
		publicTreasuryTaxRand := r.Intn(int(publicTreasuryTaxRandLimit.Mul(sdk.NewDecWithPrec(100, 0)).TruncateInt64()))
		publicTreasuryTax = sdk.NewDecWithPrec(int64(publicTreasuryTaxRand), 2)
	}

	// 100% - validatorsTax - liquidityProvidersTax - publicTreasuryTax
	harpTax = sdk.OneDec().Sub(validatorsTax).Sub(liquidityProvidersTax).Sub(publicTreasuryTax)

	return
}

// GenPublicTreasuryCapacity randomized PublicTreasuryCapacity
func GenPublicTreasuryCapacity(r *rand.Rand) sdk.Int {
	return sdk.NewInt(100000).Add(sdk.NewInt(int64(r.Intn(100000))))
}

// GenBaseProposerReward randomized BaseProposerReward
func GenBaseProposerReward(r *rand.Rand) sdk.Dec {
	return sdk.NewDecWithPrec(1, 2).Add(sdk.NewDecWithPrec(int64(r.Intn(30)), 2))
}

// GenBonusProposerReward randomized BonusProposerReward
func GenBonusProposerReward(r *rand.Rand) sdk.Dec {
	return sdk.NewDecWithPrec(1, 2).Add(sdk.NewDecWithPrec(int64(r.Intn(30)), 2))
}

// GenWithdrawEnabled returns a randomized WithdrawEnabled parameter.
func GenWithdrawEnabled(r *rand.Rand) bool {
	return r.Int63n(101) <= 95 // 95% chance of withdraws being enabled
}

// RandomizedGenState generates a random GenesisState for distribution
func RandomizedGenState(simState *module.SimulationState) {
	validatorsTax, liquidityProvidersTax, publicTreasuryTax, harpTax := GetTaxes(rand.New(rand.NewSource(1)))

	var publicTreasuryCapacity sdk.Int
	simState.AppParams.GetOrGenerate(
		simState.Cdc, PublicTreasuryCapacity, &publicTreasuryCapacity, simState.Rand,
		func(r *rand.Rand) { publicTreasuryCapacity = GenPublicTreasuryCapacity(r) },
	)

	var baseProposerReward sdk.Dec
	simState.AppParams.GetOrGenerate(
		simState.Cdc, BaseProposerReward, &baseProposerReward, simState.Rand,
		func(r *rand.Rand) { baseProposerReward = GenBaseProposerReward(r) },
	)

	var bonusProposerReward sdk.Dec
	simState.AppParams.GetOrGenerate(
		simState.Cdc, BonusProposerReward, &bonusProposerReward, simState.Rand,
		func(r *rand.Rand) { bonusProposerReward = GenBonusProposerReward(r) },
	)

	var withdrawEnabled bool
	simState.AppParams.GetOrGenerate(
		simState.Cdc, WithdrawEnabled, &withdrawEnabled, simState.Rand,
		func(r *rand.Rand) { withdrawEnabled = GenWithdrawEnabled(r) },
	)

	distrGenesis := types.GenesisState{
		RewardPools: types.InitialRewardPools(),
		Params: types.Params{
			ValidatorsPoolTax:          validatorsTax,
			LiquidityProvidersPoolTax:  liquidityProvidersTax,
			PublicTreasuryPoolTax:      publicTreasuryTax,
			PublicTreasuryPoolCapacity: publicTreasuryCapacity,
			HARPTax:                    harpTax,
			BaseProposerReward:         baseProposerReward,
			BonusProposerReward:        bonusProposerReward,
			WithdrawAddrEnabled:        withdrawEnabled,
		},
	}

	fmt.Printf("Selected randomly generated distribution parameters:\n%s\n", codec.MustMarshalJSONIndent(simState.Cdc, distrGenesis))
	simState.GenState[types.ModuleName] = simState.Cdc.MustMarshalJSON(distrGenesis)
}
