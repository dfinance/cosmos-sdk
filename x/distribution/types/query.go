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
	Reward DelegationDelegatorReward `json:"reward" yaml:"reward"`
}

func (res QueryDelegationRewardsResponse) String() string {
	out := "Delegation Rewards:\n"
	out += res.Reward.String()

	return strings.TrimSpace(out)
}

// NewQueryDelegatorTotalRewardsResponse constructs a QueryDelegatorTotalRewardsResponse
func NewQueryDelegationRewardsResponse(reward DelegationDelegatorReward) QueryDelegationRewardsResponse {
	return QueryDelegationRewardsResponse{Reward: reward}
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
	for _, reward := range res.Rewards {
		out += reward.String()
	}
	out += fmt.Sprintf("  Total: %s\n", res.Total)

	return strings.TrimSpace(out)
}

// DelegationDelegatorReward defines the properties of a delegator's delegation reward.
type DelegationDelegatorReward struct {
	ValidatorAddress sdk.ValAddress `json:"validator_address" yaml:"validator_address" swaggertype:"string" format:"bech32" example:"wallet13jyjuz3kkdvqw8u4qfkwd94emdl3vx394kn07h"`
	// Current period reward amount
	Current sdk.DecCoins `json:"current" yaml:"current"`
	// Sum of current period reward amount and stored in rewards bank amount
	Total sdk.DecCoins `json:"total" yaml:"total"`
}

func (res DelegationDelegatorReward) String() string {
	return fmt.Sprintf(`  ValidatorAddress: %s
    Current: %s
    Total:   %s
`,
		res.ValidatorAddress, res.Current, res.Total,
	)
}

// NewDelegationDelegatorReward constructs a DelegationDelegatorReward.
func NewDelegationDelegatorReward(
	valAddr sdk.ValAddress,
	current, total sdk.DecCoins,
) DelegationDelegatorReward {

	return DelegationDelegatorReward{ValidatorAddress: valAddr, Current: current, Total: total}
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
