package distribution

import (
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/distribution/types"
	"github.com/cosmos/cosmos-sdk/x/mint"
	"github.com/stretchr/testify/require"
)

// Test FoundationPool withdraw authorization (nominee check).
func TestHandleMsgWithdrawFoundationPool(t *testing.T) {
	nominee := delAddr1
	guest := delAddr2

	// set params with nominee
	params := DefaultParams()
	params.FoundationNominees = append(params.FoundationNominees, nominee)

	ctx, accountKeeper, _, keeper, _, _, supplyKeeper, mk := CreateTestInputAdvanced(t, false, 10, params)

	_ = mk // linter fix

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

func TestHandleMsgSetFoundationAllocationRatio(t *testing.T) {
	nominee := delAddr1
	guest := delAddr2

	// set params with nominee
	params := DefaultParams()
	params.FoundationNominees = append(params.FoundationNominees, nominee)

	ctx, accountKeeper, _, keeper, _, _, sk, mintKeeper := CreateTestInputAdvanced(t, false, 10, params)

	_ = sk // linter fix

	p := mintKeeper.GetParams(ctx)
	p.AvgBlockTimeWindow = 2
	stdRatio := mint.FoundationAllocationRatioMaxValue
	p.FoundationAllocationRatio = stdRatio
	mintKeeper.SetParams(ctx, p)

	ctx = ctx.WithBlockTime(time.Now())
	mintKeeper.AdjustAvgBLockDurEstimation(ctx)
	ctx = ctx.WithBlockTime(time.Now().Add(time.Second * 20))
	mintKeeper.AdjustAvgBLockDurEstimation(ctx)
	ctx = ctx.WithBlockHeight(10)

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
		msg := types.NewMsgSetFoundationAllocationRatio(guest, stdRatio)
		_, err := handleMsgSetFoundationAllocationRatio(ctx, msg, keeper, mintKeeper)
		require.Error(t, err)
	}

	// check wrong ratio, over than max value
	{
		func() {
			defer func() {
				if r := recover(); r != nil {
					require.Contains(t, r, "cannot be greater")
				}
			}()
			ratio := mint.FoundationAllocationRatioMaxValue.Add(sdk.NewDecWithPrec(1, 2))
			msg := types.NewMsgSetFoundationAllocationRatio(nominee, ratio)
			_, err := handleMsgSetFoundationAllocationRatio(ctx, msg, keeper, mintKeeper)
			require.Error(t, err)
		}()
	}

	// check wrong ratio, over than max value
	{
		func() {
			defer func() {
				if r := recover(); r != nil {
					require.Contains(t, r, "cannot be nagative")
				}
			}()
			ratio := sdk.NewDec(-1)
			msg := types.NewMsgSetFoundationAllocationRatio(nominee, ratio)
			_, err := handleMsgSetFoundationAllocationRatio(ctx, msg, keeper, mintKeeper)
			require.Error(t, err)
		}()
	}

	// check ChangeFoundationAllocationRatioTTL
	{
		abpy, err := mintKeeper.GetAvgBlocksPerYear(ctx)
		require.NoError(t, err)

		targetBlockHeight := int64(abpy * ChangeFoundationAllocationRatioTTL)

		// block limit - 1 block
		{
			ctx := ctx.WithBlockHeight(targetBlockHeight - 1)
			msg := types.NewMsgSetFoundationAllocationRatio(nominee, stdRatio)
			_, err = handleMsgSetFoundationAllocationRatio(ctx, msg, keeper, mintKeeper)
			require.NoError(t, err)
		}

		// block limit == block height
		{
			ctx := ctx.WithBlockHeight(targetBlockHeight)
			msg := types.NewMsgSetFoundationAllocationRatio(nominee, stdRatio)
			_, err = handleMsgSetFoundationAllocationRatio(ctx, msg, keeper, mintKeeper)
			require.Error(t, err)
			require.Contains(t, err.Error(), "is not allowed to change after")
		}

		// block limit > block height
		{
			ctx := ctx.WithBlockHeight(targetBlockHeight + 1)
			msg := types.NewMsgSetFoundationAllocationRatio(nominee, stdRatio)
			_, err = handleMsgSetFoundationAllocationRatio(ctx, msg, keeper, mintKeeper)
			require.Error(t, err)
			require.Contains(t, err.Error(), "is not allowed to change after")
		}
	}

	// check ok
	{
		msg := types.NewMsgSetFoundationAllocationRatio(nominee, stdRatio)
		_, err := handleMsgSetFoundationAllocationRatio(ctx, msg, keeper, mintKeeper)
		require.NoError(t, err)
		require.Equal(t, stdRatio, mintKeeper.GetParams(ctx).FoundationAllocationRatio)
	}
}
