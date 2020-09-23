package types

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// GenesisState - all staking state that must be provided at genesis
type GenesisState struct {
	Params               Params                 `json:"params" yaml:"params"`
	LastTotalPower       sdk.Int                `json:"last_total_power" yaml:"last_total_power"`
	LastValidatorPowers  []LastValidatorPower   `json:"last_validator_powers" yaml:"last_validator_powers"`
	Validators           Validators             `json:"validators" yaml:"validators"`
	Delegations          Delegations            `json:"delegations" yaml:"delegations"`
	UnbondingDelegations []UnbondingDelegation  `json:"unbonding_delegations" yaml:"unbonding_delegations"`
	Redelegations        []Redelegation         `json:"redelegations" yaml:"redelegations"`
	StakingStates        []StakingStateEntry    `json:"staking_states" yaml:"staking_states"`
	ScheduledUnbonds     []ScheduledUnbondEntry `json:"scheduled_unbonds" yaml:"scheduled_unbonds"`
	BannedAccounts       []BannedAccountEntry   `json:"banned_accounts" yaml:"banned_accounts"`
	Exported             bool                   `json:"exported" yaml:"exported"`
}

// LastValidatorPower required for validator set update logic
type LastValidatorPower struct {
	Address sdk.ValAddress
	Power   int64
}

// ScheduledUnbondEntry keeps ScheduledUnbondQueue state
type ScheduledUnbondEntry struct {
	Timestamp time.Time        `json:"timestamp" yaml:"timestamp"`
	ValAddrs  []sdk.ValAddress `json:"val_addrs" yaml:"val_addrs"`
}

// BannedAccountEntry keeps banned account info
type BannedAccountEntry struct {
	AccAddress sdk.AccAddress `json:"acc_address" yaml:"acc_address"`
	BanHeight  int64          `json:"ban_height" yaml:"ban_height"`
}

// StakingStateEntry keeps validator staking state.
type StakingStateEntry struct {
	ValAddr sdk.ValAddress        `json:"val_addr" yaml:"val_addr"`
	State   ValidatorStakingState `json:"state" yaml:"state"`
}

// NewGenesisState creates a new GenesisState instanc e
func NewGenesisState(params Params, validators []Validator, delegations []Delegation) GenesisState {
	return GenesisState{
		Params:      params,
		Validators:  validators,
		Delegations: delegations,
	}
}

// DefaultGenesisState gets the raw genesis raw message for testing
func DefaultGenesisState() GenesisState {
	return GenesisState{
		Params: DefaultParams(),
	}
}
