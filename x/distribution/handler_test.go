package distribution

import (
	"testing"

	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Test FoundationPool withdraw authorization (nominee check).
func TestHandleMsgWithdrawFoundationPool(t *testing.T) {
	nominee := delAddr1
	guest := delAddr2

	// set params with nominee
	params := DefaultParams()
	params.FoundationNominees = append(params.FoundationNominees, nominee)

	ctx, accountKeeper, _, keeper, _, _, supplyKeeper := CreateTestInputAdvanced(t, false, 10, params)

	// add coins to the module account
	{
		macc := keeper.GetDistributionAccount(ctx)
		err := macc.SetCoins(macc.GetCoins().Add(amount...))
		require.NoError(t, err)
		supplyKeeper.SetModuleAccount(ctx, macc)
	}

	// create accounts
	{
		nomineeAcc := accountKeeper.NewAccountWithAddress(ctx, nominee)
		require.True(t, nomineeAcc.GetCoins().IsZero())
		accountKeeper.SetAccount(ctx, nomineeAcc)

		guestAcc := accountKeeper.NewAccountWithAddress(ctx, guest)
		require.True(t, guestAcc.GetCoins().IsZero())
		accountKeeper.SetAccount(ctx, guestAcc)
	}

	// set FoundationPool supply
	{
		rewardPools := keeper.GetRewardPools(ctx)
		rewardPools.FoundationPool = sdk.NewDecCoinsFromCoins(amount...)
		keeper.SetRewardPools(ctx, rewardPools)
	}

	// check fail (non-nominee)
	{
		msg := NewMsgWithdrawFoundationPool(guest, guest, "", amount)
		_, err := handleMsgWithdrawFoundationPool(ctx, msg, keeper)
		require.Error(t, err)
	}

	// check ok
	{
		msg := NewMsgWithdrawFoundationPool(nominee, guest, "", amount)
		_, err := handleMsgWithdrawFoundationPool(ctx, msg, keeper)
		require.NoError(t, err)

		require.True(t, accountKeeper.GetAccount(ctx, guest).GetCoins().IsEqual(amount))
		require.True(t, supplyKeeper.GetModuleAccount(ctx, ModuleName).GetCoins().Empty())
	}
}
