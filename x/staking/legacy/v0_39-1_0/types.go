package v0_39_1_0

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	v03902 "github.com/cosmos/cosmos-sdk/x/staking/legacy/v0_39-0_2"
)

const (
	ModuleName = "staking"
)

// GenesisState for Dfinance v1.0 Mainnet based on Cosmos SDK v0.39.1.
// Changelist:
//   - params.max_self_delegation_lvl added;
type (
	GenesisState struct {
		Params               Params                        `json:"params"`
		LastTotalPower       sdk.Int                       `json:"last_total_power"`
		LastValidatorPowers  []v03902.LastValidatorPower   `json:"last_validator_powers"`
		Validators           v03902.Validators             `json:"validators"`
		Delegations          v03902.Delegations            `json:"delegations"`
		UnbondingDelegations []v03902.UnbondingDelegation  `json:"unbonding_delegations"`
		Redelegations        []v03902.Redelegation         `json:"redelegations"`
		StakingStates        []v03902.StakingStateEntry    `json:"staking_states"`
		ScheduledUnbonds     []v03902.ScheduledUnbondEntry `json:"scheduled_unbonds"`
		BannedAccounts       []v03902.BannedAccountEntry   `json:"banned_accounts"`
		Exported             bool                          `json:"exported"`
	}

	Params struct {
		UnbondingTime            time.Duration `json:"unbonding_time"`
		MaxValidators            uint16        `json:"max_validators"`
		MaxEntries               uint16        `json:"max_entries"`
		HistoricalEntries        uint16        `json:"historical_entries"`
		BondDenom                string        `json:"bond_denom"`
		LPDenom                  string        `json:"lp_denom"`
		LPDistrRatio             sdk.Dec       `json:"lp_distr_ratio"`
		MinSelfDelegationLvl     sdk.Int       `json:"min_self_delegation_lvl"`
		MaxSelfDelegationLvl     sdk.Int       `json:"max_self_delegation_lvl"`
		MaxDelegationsRatio      sdk.Dec       `json:"max_delegations_ratio"`
		ScheduledUnbondDelayTime time.Duration `json:"scheduled_unbond_delay"`
	}
)
