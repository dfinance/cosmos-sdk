package v0_39_0_2

import (
	"time"

	"github.com/tendermint/tendermint/crypto"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// GenesisState for Dfinance v0.2 Testnet based on Cosmos SDK v0.39.1.
// This is a starting point for all future Dfinance migrations.
type (
	GenesisState struct {
		Params               Params                 `json:"params"`
		LastTotalPower       sdk.Int                `json:"last_total_power"`
		LastValidatorPowers  []LastValidatorPower   `json:"last_validator_powers"`
		Validators           Validators             `json:"validators"`
		Delegations          Delegations            `json:"delegations"`
		UnbondingDelegations []UnbondingDelegation  `json:"unbonding_delegations"`
		Redelegations        []Redelegation         `json:"redelegations"`
		StakingStates        []StakingStateEntry    `json:"staking_states"`
		ScheduledUnbonds     []ScheduledUnbondEntry `json:"scheduled_unbonds"`
		BannedAccounts       []BannedAccountEntry   `json:"banned_accounts"`
		Exported             bool                   `json:"exported"`
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
		MaxDelegationsRatio      sdk.Dec       `json:"max_delegations_ratio"`
		ScheduledUnbondDelayTime time.Duration `json:"scheduled_unbond_delay"`
	}

	LastValidatorPower struct {
		Address sdk.ValAddress
		Power   int64
	}

	Validators []Validator

	Validator struct {
		OperatorAddress          sdk.ValAddress  `json:"operator_address"`
		ConsPubKey               crypto.PubKey   `json:"consensus_pubkey"`
		Jailed                   bool            `json:"jailed"`
		ScheduledToUnbond        bool            `json:"scheduled_to_unbond"`
		Status                   sdk.BondStatus  `json:"status"`
		Bonding                  ValidatorTokens `json:"bonding"`
		LP                       ValidatorTokens `json:"lp"`
		Description              Description     `json:"description"`
		UnbondingHeight          int64           `json:"unbonding_height"`
		UnbondingCompletionTime  time.Time       `json:"unbonding_time"`
		ScheduledUnbondHeight    int64           `json:"scheduled_unbond_height"`
		ScheduledUnbondStartTime time.Time       `json:"scheduled_unbond_time"`
		Commission               Commission      `json:"commission"`
		MinSelfDelegation        sdk.Int         `json:"min_self_delegation"`
	}

	bechValidator struct {
		OperatorAddress          sdk.ValAddress `json:"operator_address"`
		ConsPubKey               string         `json:"consensus_pubkey"`
		Jailed                   bool           `json:"jailed"`
		ScheduledToUnbond        bool           `json:"scheduled_to_unbond"`
		Status                   sdk.BondStatus `json:"status"`
		BondingDelegatorShares   sdk.Dec        `json:"bonding_delegator_shares"`
		BondingTokens            sdk.Int        `json:"bonding_tokens"`
		LPDelegatorShares        sdk.Dec        `json:"lp_delegator_shares"`
		LPTokens                 sdk.Int        `json:"lp_tokens"`
		Description              Description    `json:"description"`
		UnbondingHeight          int64          `json:"unbonding_height" `
		UnbondingCompletionTime  time.Time      `json:"unbonding_time"`
		ScheduledUnbondHeight    int64          `json:"scheduled_unbond_height"`
		ScheduledUnbondStartTime time.Time      `json:"scheduled_unbond_time"`
		Commission               Commission     `json:"commission"`
		MinSelfDelegation        sdk.Int        `json:"min_self_delegation"`
	}

	Description struct {
		Moniker         string `json:"moniker"`
		Identity        string `json:"identity"`
		Website         string `json:"website"`
		SecurityContact string `json:"security_contact"`
		Details         string `json:"details"`
	}

	ValidatorTokens struct {
		DelegatorShares sdk.Dec `json:"delegator_shares"`
		Tokens          sdk.Int `json:"tokens"`
	}

	Commission struct {
		CommissionRates `json:"commission_rates"`
		UpdateTime      time.Time `json:"update_time"`
	}

	CommissionRates struct {
		Rate          sdk.Dec `json:"rate"`
		MaxRate       sdk.Dec `json:"max_rate"`
		MaxChangeRate sdk.Dec `json:"max_change_rate"`
	}

	Delegations []Delegation

	Delegation struct {
		DelegatorAddress sdk.AccAddress `json:"delegator_address"`
		ValidatorAddress sdk.ValAddress `json:"validator_address"`
		BondingShares    sdk.Dec        `json:"bonding_shares"`
		LPShares         sdk.Dec        `json:"lp_shares"`
	}

	UnbondingDelegation struct {
		DelegatorAddress sdk.AccAddress             `json:"delegator_address"`
		ValidatorAddress sdk.ValAddress             `json:"validator_address"`
		Entries          []UnbondingDelegationEntry `json:"entries"`
	}

	UnbondingDelegationEntry struct {
		CreationHeight int64            `json:"creation_height"`
		CompletionTime time.Time        `json:"completion_time"`
		OpType         DelegationOpType `json:"op_type"`
		InitialBalance sdk.Int          `json:"initial_balance"`
		Balance        sdk.Int          `json:"balance"`
	}

	DelegationOpType string

	Redelegation struct {
		DelegatorAddress    sdk.AccAddress      `json:"delegator_address"`
		ValidatorSrcAddress sdk.ValAddress      `json:"validator_src_address"`
		ValidatorDstAddress sdk.ValAddress      `json:"validator_dst_address"`
		Entries             []RedelegationEntry `json:"entries"`
	}

	RedelegationEntry struct {
		CreationHeight int64            `json:"creation_height"`
		CompletionTime time.Time        `json:"completion_time"`
		OpType         DelegationOpType `json:"op_type"`
		InitialBalance sdk.Int          `json:"initial_balance"`
		SharesDst      sdk.Dec          `json:"shares_dst"`
	}

	StakingStateEntry struct {
		ValAddr sdk.ValAddress        `json:"val_addr"`
		State   ValidatorStakingState `json:"state"`
	}

	ValidatorStakingState struct {
		Operator   DelegationTruncated   `json:"operator"`
		Delegators []DelegationTruncated `json:"delegators"`
	}

	DelegationTruncated struct {
		Address       sdk.AccAddress `json:"address"`
		BondingShares sdk.Dec        `json:"bonding_shares"`
		LPShares      sdk.Dec        `json:"lp_shares"`
	}

	ScheduledUnbondEntry struct {
		Timestamp time.Time        `json:"timestamp"`
		ValAddrs  []sdk.ValAddress `json:"val_addrs"`
	}

	BannedAccountEntry struct {
		AccAddress sdk.AccAddress `json:"acc_address"`
		BanHeight  int64          `json:"ban_height"`
	}
)

// MarshalJSON marshals the validator to JSON using Bech32.
func (v Validator) MarshalJSON() ([]byte, error) {
	bechConsPubKey, err := sdk.Bech32ifyPubKey(sdk.Bech32PubKeyTypeConsPub, v.ConsPubKey)
	if err != nil {
		return nil, err
	}

	return codec.Cdc.MarshalJSON(bechValidator{
		OperatorAddress:          v.OperatorAddress,
		ConsPubKey:               bechConsPubKey,
		Jailed:                   v.Jailed,
		ScheduledToUnbond:        v.ScheduledToUnbond,
		Status:                   v.Status,
		BondingDelegatorShares:   v.Bonding.DelegatorShares,
		BondingTokens:            v.Bonding.Tokens,
		LPDelegatorShares:        v.LP.DelegatorShares,
		LPTokens:                 v.LP.Tokens,
		Description:              v.Description,
		UnbondingHeight:          v.UnbondingHeight,
		UnbondingCompletionTime:  v.UnbondingCompletionTime,
		ScheduledUnbondHeight:    v.ScheduledUnbondHeight,
		ScheduledUnbondStartTime: v.ScheduledUnbondStartTime,
		MinSelfDelegation:        v.MinSelfDelegation,
		Commission:               v.Commission,
	})
}

// UnmarshalJSON unmarshals the validator from JSON using Bech32.
func (v *Validator) UnmarshalJSON(data []byte) error {
	bv := &bechValidator{}
	if err := codec.Cdc.UnmarshalJSON(data, bv); err != nil {
		return err
	}
	consPubKey, err := sdk.GetPubKeyFromBech32(sdk.Bech32PubKeyTypeConsPub, bv.ConsPubKey)
	if err != nil {
		return err
	}
	*v = Validator{
		OperatorAddress:   bv.OperatorAddress,
		ConsPubKey:        consPubKey,
		Jailed:            bv.Jailed,
		ScheduledToUnbond: bv.ScheduledToUnbond,
		Status:            bv.Status,
		Bonding: ValidatorTokens{
			DelegatorShares: bv.BondingDelegatorShares,
			Tokens:          bv.BondingTokens,
		},
		LP: ValidatorTokens{
			DelegatorShares: bv.LPDelegatorShares,
			Tokens:          bv.LPTokens,
		},
		Description:              bv.Description,
		UnbondingHeight:          bv.UnbondingHeight,
		UnbondingCompletionTime:  bv.UnbondingCompletionTime,
		ScheduledUnbondHeight:    bv.ScheduledUnbondHeight,
		ScheduledUnbondStartTime: bv.ScheduledUnbondStartTime,
		Commission:               bv.Commission,
		MinSelfDelegation:        bv.MinSelfDelegation,
	}
	return nil
}
