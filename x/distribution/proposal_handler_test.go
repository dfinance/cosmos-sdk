package distribution

import (
	"testing"

	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/distribution/types"
)

func testPublicTreasurySpendProposal(recipient sdk.AccAddress, amount sdk.Coins) types.PublicTreasuryPoolSpendProposal {
	return types.NewPublicTreasuryPoolSpendProposal(
		"Test",
		"description",
		recipient,
		amount,
	)
}

func testTaxParamsUpdateProposal(validatorsTax, liquidityTax, treasuryTax, harpTax sdk.Dec) types.TaxParamsUpdateProposal {
	return types.NewTaxParamsUpdateProposal(
		"Test",
		"description",
		validatorsTax,
		liquidityTax,
		treasuryTax,
		harpTax,
	)
}

func TestPublicTreasurySpendProposalHandlerPassed(t *testing.T) {
	ctx, accountKeeper, keeper, _, supplyKeeper := CreateTestInputDefault(t, false, 10)
	recipient := delAddr1

	// add coins to the module account
	macc := keeper.GetDistributionAccount(ctx)
	err := macc.SetCoins(macc.GetCoins().Add(amount...))
	require.NoError(t, err)

	supplyKeeper.SetModuleAccount(ctx, macc)

	account := accountKeeper.NewAccountWithAddress(ctx, recipient)
	require.True(t, account.GetCoins().IsZero())
	accountKeeper.SetAccount(ctx, account)

	rewardPools := keeper.GetRewardPools(ctx)
	rewardPools.PublicTreasuryPool = sdk.NewDecCoinsFromCoins(amount...)
	keeper.SetRewardPools(ctx, rewardPools)

	tp := testPublicTreasurySpendProposal(recipient, amount)
	hdlr := NewProposalHandler(keeper)
	require.NoError(t, hdlr(ctx, tp))
	require.Equal(t, accountKeeper.GetAccount(ctx, recipient).GetCoins(), amount)
}

func TestPublicTreasurySpendProposalHandlerFailed(t *testing.T) {
	ctx, accountKeeper, keeper, _, _ := CreateTestInputDefault(t, false, 10)
	recipient := delAddr1

	account := accountKeeper.NewAccountWithAddress(ctx, recipient)
	require.True(t, account.GetCoins().IsZero())
	accountKeeper.SetAccount(ctx, account)

	tp := testPublicTreasurySpendProposal(recipient, amount)
	hdlr := NewProposalHandler(keeper)
	require.Error(t, hdlr(ctx, tp))
	require.True(t, accountKeeper.GetAccount(ctx, recipient).GetCoins().IsZero())
}

func TestTaxParamsUpdateHandler(t *testing.T) {
	ctx, _, keeper, _, _ := CreateTestInputDefault(t, false, 10)

	tp := testTaxParamsUpdateProposal(
		sdk.NewDecWithPrec(40, 2),
		sdk.NewDecWithPrec(35, 2),
		sdk.NewDecWithPrec(15, 2),
		sdk.NewDecWithPrec(10, 2),
	)
	hdlr := NewProposalHandler(keeper)
	require.NoError(t, hdlr(ctx, tp))

	newParams := keeper.GetParams(ctx)
	require.True(t, newParams.ValidatorsPoolTax.Equal(tp.ValidatorsPoolTax))
	require.True(t, newParams.LiquidityProvidersPoolTax.Equal(tp.LiquidityProvidersPoolTax))
	require.True(t, newParams.PublicTreasuryPoolTax.Equal(tp.PublicTreasuryPoolTax))
	require.True(t, newParams.HARPTax.Equal(tp.HARPTax))
}
