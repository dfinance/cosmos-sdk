package keeper

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/distribution/types"
	"github.com/cosmos/cosmos-sdk/x/staking"
)

func TestLockedRewards(t *testing.T) {
	ctx, _, k, sk, _ := CreateTestInputDefault(t, false, 1000)
	sh := staking.NewHandler(sk)

	// create two identical validators
	{
		commission := staking.NewCommissionRates(sdk.NewDecWithPrec(2, 1), sdk.NewDecWithPrec(2, 1), sdk.NewDec(0))

		msg := staking.NewMsgCreateValidator(valOpAddr1, valConsPk1,
			sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(100)), staking.Description{}, commission, minSelfDelegation)

		res, err := sh(ctx, msg)
		require.NoError(t, err)
		require.NotNil(t, res)

		msg = staking.NewMsgCreateValidator(valOpAddr2, valConsPk2,
			sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(100)), staking.Description{}, commission, minSelfDelegation)

		res, err = sh(ctx, msg)
		require.NoError(t, err)
		require.NotNil(t, res)
	}

	// end block to bond validator
	{
		staking.EndBlocker(ctx, sk)
		ctx = ctx.WithBlockHeight(ctx.BlockHeight() + 1)
	}

	// try enable locking from the wrong operator
	{
		msg := types.NewMsgLockValidatorRewards(valAccAddr2, valOpAddr1)
		require.Error(t, msg.ValidateBasic())
	}

	// ensure distribution power is equal
	{
		dPower1 := k.GetDistributionPower(ctx, valOpAddr1, 100)
		dPower2 := k.GetDistributionPower(ctx, valOpAddr2, 100)
		require.Equal(t, dPower1, dPower2)
	}

	// lock rewards for the 1st validator
	var rewardsUnlockedAt1 time.Time
	{
		msg := types.NewMsgLockValidatorRewards(valAccAddr1, valOpAddr1)
		require.NoError(t, msg.ValidateBasic())

		unlocksAt, err := k.LockValidatorRewards(ctx, msg.ValidatorAddress)
		require.NoError(t, err)
		require.False(t, unlocksAt.IsZero())

		state1, found1 := k.GetValidatorLockedState(ctx, valOpAddr1)
		require.True(t, found1)
		require.True(t, state1.IsLocked())

		state2, found2 := k.GetValidatorLockedState(ctx, valOpAddr2)
		require.True(t, found2)
		require.False(t, state2.IsLocked())

		rewardsUnlockedAt1 = state1.UnlocksAt
	}

	// ensure distribution power is different
	{
		dPower1 := k.GetDistributionPower(ctx, valOpAddr1, 100)
		dPower2 := k.GetDistributionPower(ctx, valOpAddr2, 100)
		require.Greater(t, dPower1, dPower2)
	}

	// emulate some time has passed and lock is still here
	{
		ctx = ctx.WithBlockHeight(ctx.BlockHeight() + 1).WithBlockTime(ctx.BlockTime().Add(1 * time.Hour))

		k.ProcessAllMatureRewardsUnlockQueueItems(ctx)

		state, found := k.GetValidatorLockedState(ctx, valOpAddr1)
		require.True(t, found)
		require.True(t, state.IsLocked())
	}

	// try to withdraw rewards from the 1st validator
	{
		_, err := k.WithdrawDelegationRewards(ctx, valAccAddr1, valOpAddr1)
		require.Error(t, err)
	}

	// emulate unlock period is over
	{
		ctx = ctx.WithBlockHeight(ctx.BlockHeight() + 1).WithBlockTime(rewardsUnlockedAt1)

		k.ProcessAllMatureRewardsUnlockQueueItems(ctx)

		state, found := k.GetValidatorLockedState(ctx, valOpAddr1)
		require.True(t, found)
		require.False(t, state.IsLocked())
	}

	// check withdraw rewards is unlocked
	{
		_, err := k.WithdrawDelegationRewards(ctx, valAccAddr1, valOpAddr1)
		require.NoError(t, err)
	}

	// ensure distribution power is equal again
	{
		dPower1 := k.GetDistributionPower(ctx, valOpAddr1, 100)
		dPower2 := k.GetDistributionPower(ctx, valOpAddr2, 100)
		require.Equal(t, dPower1, dPower2)
	}
}
