package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// RewardPoolName defines reward pool name.
type RewardPoolName string

const (
	LiquidityProvidersPoolName RewardPoolName = "LiquidityProvidersPool"
	FoundationPoolName         RewardPoolName = "FoundationPool"
	PublicTreasuryPoolName     RewardPoolName = "PublicTreasuryPool"
	HARPName                   RewardPoolName = "HARP"
)

// IsValid checks RewardPoolName enum.
func (n RewardPoolName) IsValid() bool {
	switch n {
	case LiquidityProvidersPoolName, FoundationPoolName, PublicTreasuryPoolName, HARPName:
		return true
	default:
		return false
	}
}

// RewardPools contains collected rewards distributed by pools.
// RewardPools aren't module accounts, however its coins are held in the distribution module account.
// Thus the RewardPools must be reduced separately from the SendCoinsFromModuleToAccount call.
type RewardPools struct {
	// Pool for chain liquidity providers rewards
	LiquidityProvidersPool sdk.DecCoins `json:"liquidity_providers_pool" yaml:"liquidity_providers_pool"`
	// Chain maintainers controlled pool
	FoundationPool sdk.DecCoins `json:"foundation_pool" yaml:"foundation_pool"`
	// Community controlled pool
	PublicTreasuryPool sdk.DecCoins `json:"treasury_pool" yaml:"treasury_pool"`
	// High Availability Reward Pool used to reward top validators
	HARP sdk.DecCoins `json:"harp_pool" yaml:"harp_pool"`
}

// InitialRewardPools returns the initial RewardPools state.
func InitialRewardPools() RewardPools {
	return RewardPools{
		LiquidityProvidersPool: sdk.DecCoins{},
		FoundationPool:         sdk.DecCoins{},
		PublicTreasuryPool:     sdk.DecCoins{},
		HARP:                   sdk.DecCoins{},
	}
}

// ValidateGenesis validates the pools for a genesis state.
func (p RewardPools) ValidateGenesis() error {
	if p.LiquidityProvidersPool.IsAnyNegative() {
		return fmt.Errorf("negative LiquidityProvidersPool in distribution RewardPools: %s", p.LiquidityProvidersPool)
	}
	if p.FoundationPool.IsAnyNegative() {
		return fmt.Errorf("negative FoundationPool in distribution RewardPools: %s", p.FoundationPool)
	}
	if p.PublicTreasuryPool.IsAnyNegative() {
		return fmt.Errorf("negative PublicTreasuryPool in distribution RewardPools: %s", p.PublicTreasuryPool)
	}
	if p.HARP.IsAnyNegative() {
		return fmt.Errorf("negative HARP in distribution RewardPools: %s", p.HARP)
	}

	return nil
}

// TotalCoins returns sum of all pools.
func (p RewardPools) TotalCoins() sdk.DecCoins {
	coins := p.LiquidityProvidersPool
	coins = coins.Add(p.FoundationPool...)
	coins = coins.Add(p.PublicTreasuryPool...)
	coins = coins.Add(p.HARP...)

	return coins
}
