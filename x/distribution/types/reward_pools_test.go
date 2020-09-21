package types

import (
	"testing"

	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func TestValidateGenesis(t *testing.T) {
	// ok
	{
		fp := InitialRewardPools()
		require.Nil(t, fp.ValidateGenesis())
	}

	// fail
	{
		fp := RewardPools{LiquidityProvidersPool: sdk.DecCoins{{Denom: "stake", Amount: sdk.NewDec(-1)}}}
		require.NotNil(t, fp.ValidateGenesis())
	}

	// fail
	{
		fp := RewardPools{FoundationPool: sdk.DecCoins{{Denom: "stake", Amount: sdk.NewDec(-1)}}}
		require.NotNil(t, fp.ValidateGenesis())
	}

	// fail
	{
		fp := RewardPools{PublicTreasuryPool: sdk.DecCoins{{Denom: "stake", Amount: sdk.NewDec(-1)}}}
		require.NotNil(t, fp.ValidateGenesis())
	}

	// fail
	{
		fp := RewardPools{HARP: sdk.DecCoins{{Denom: "stake", Amount: sdk.NewDec(-1)}}}
		require.NotNil(t, fp.ValidateGenesis())
	}
}
