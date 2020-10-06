package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// DelegatorStartingInfo is starting info for a delegator reward period.
// Tracks the previous validator period, the delegation's amount
// of staking token, and the creation height (to check later on
// if any slashes have occurred).
// NOTE that even though validators are slashed to whole staking tokens, the
// delegators within the validator may be left with less than a full token,
// thus sdk.Dec is used.
type DelegatorStartingInfo struct {
	// Period at which the delegation should withdraw starting from
	PreviousPeriod uint64 `json:"previous_period" yaml:"previous_period"`
	// Amount of staking bonding tokens delegated
	BondingStake sdk.Dec `json:"bonding_stake" yaml:"bonding_stake"`
	// Amount of staking LP tokens delegated
	LPStake sdk.Dec `json:"lp_stake" yaml:"lp_stake"`
	// Height at which delegation was created
	Height uint64 `json:"creation_height" yaml:"creation_height"`
}

// create a new DelegatorStartingInfo
func NewDelegatorStartingInfo(previousPeriod uint64, bondingStake, lpStake sdk.Dec, height uint64) DelegatorStartingInfo {
	return DelegatorStartingInfo{
		PreviousPeriod: previousPeriod,
		BondingStake:   bondingStake,
		LPStake:        lpStake,
		Height:         height,
	}
}
