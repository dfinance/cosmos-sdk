package types

import (
	"fmt"
	"strings"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// QueryDelegationRewardsResponse defines the properties of
// QueryDelegationRewards query's response.
type QueryDelegationRewardsResponse struct {
	// Current rewards for a specific validator
	Rewards DelegationDelegatorReward `json:"rewards" yaml:"rewards"`
	// All validators rewards accumulated on delegation modification events (shares change, undelegation, redelegation)
	// This truncated Int value would be transferred to the delegator account on withdraw_delegator_reward Tx
	Total sdk.DecCoins `json:"total" yaml:"total"`
}

func (res QueryDelegationRewardsResponse) String() string {
	out := "Delegation Rewards:\n"
	out += fmt.Sprintf(`Rewards:
	ValidatorAddress: %s
	Reward: %s`, res.Rewards.ValidatorAddress, res.Rewards.Reward)
	out += fmt.Sprintf("\n  Total: %s\n", res.Total)

	return strings.TrimSpace(out)
}

// NewQueryDelegatorTotalRewardsResponse constructs a QueryDelegatorTotalRewardsResponse
func NewQueryDelegationRewardsResponse(
	rewards DelegationDelegatorReward, total sdk.DecCoins,
) QueryDelegationRewardsResponse {

	return QueryDelegationRewardsResponse{Rewards: rewards, Total: total}
}

// QueryDelegatorTotalRewardsResponse defines the properties of
// QueryDelegatorTotalRewards query's response.
type QueryDelegatorTotalRewardsResponse struct {
	// Current rewards for all delegated validators
	Rewards []DelegationDelegatorReward `json:"rewards" yaml:"rewards"`
	// All validators rewards accumulated on delegations modification events (shares change, undelegation, redelegation)
	Total sdk.DecCoins `json:"total" yaml:"total"`
}

// NewQueryDelegatorTotalRewardsResponse constructs a QueryDelegatorTotalRewardsResponse
func NewQueryDelegatorTotalRewardsResponse(
	rewards []DelegationDelegatorReward, total sdk.DecCoins,
) QueryDelegatorTotalRewardsResponse {

	return QueryDelegatorTotalRewardsResponse{Rewards: rewards, Total: total}
}

func (res QueryDelegatorTotalRewardsResponse) String() string {
	out := "Delegator Total Rewards:\n"
	out += "  Rewards:"
	for _, reward := range res.Rewards {
		out += fmt.Sprintf(`  
	ValidatorAddress: %s
	Reward: %s`, reward.ValidatorAddress, reward.Reward)
	}
	out += fmt.Sprintf("\n  Total: %s\n", res.Total)

	return strings.TrimSpace(out)
}

// DelegationDelegatorReward defines the properties
// of a delegator's delegation reward.
type DelegationDelegatorReward struct {
	ValidatorAddress sdk.ValAddress `json:"validator_address" yaml:"validator_address" swaggertype:"string" format:"bech32" example:"wallet13jyjuz3kkdvqw8u4qfkwd94emdl3vx394kn07h"`
	Reward           sdk.DecCoins   `json:"reward" yaml:"reward"`
}

// NewDelegationDelegatorReward constructs a DelegationDelegatorReward.
func NewDelegationDelegatorReward(valAddr sdk.ValAddress,
	reward sdk.DecCoins) DelegationDelegatorReward {
	return DelegationDelegatorReward{ValidatorAddress: valAddr, Reward: reward}
}

// QueryLockedRewardsStateResponse defines the locked_rewards_state query response.
type QueryLockedRewardsStateResponse struct {
	Enabled      bool      `json:"enabled" yaml:"enabled"`
	AutoRenew    bool      `json:"auto_renew" yaml:"auto_renew"`
	LockedHeight int64     `json:"locked_height" yaml:"locked_height"`
	LockedAt     time.Time `json:"locked_at" yaml:"locked_at"`
	UnlocksAt    time.Time `json:"unlocks_at" yaml:"unlocks_at"`
}

// NewQueryLockedRewardsStateResponse constructs a QueryLockedRewardsStateResponse.
func NewQueryLockedRewardsStateResponse(state ValidatorLockedRewardsState) QueryLockedRewardsStateResponse {
	r := QueryLockedRewardsStateResponse{
		LockedHeight: state.LockHeight,
		LockedAt:     state.LockedAt,
		UnlocksAt:    state.UnlocksAt,
		AutoRenew:    state.AutoRenewal,
	}
	if state.IsLocked() {
		r.Enabled = true
	}

	return r
}
