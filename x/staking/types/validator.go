package types

import (
	"bytes"
	"fmt"
	"time"

	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/crypto"
	tmtypes "github.com/tendermint/tendermint/types"
	yaml "gopkg.in/yaml.v2"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/staking/exported"
)

// Implements Validator interface
var _ exported.ValidatorI = Validator{}

// Validator defines the total amount of bond shares and their exchange rate to
// coins. Slashing results in a decrease in the exchange rate, allowing correct
// calculation of future undelegations without iterating over delegators.
// When coins are delegated to this validator, the validator is credited with a
// delegation whose number of bond shares is based on the amount of coins delegated
// divided by the current exchange rate. Voting power can be calculated as total
// bonded shares multiplied by exchange rate.
type Validator struct {
	// Address of the validator's operator; bech encoded in JSON
	OperatorAddress sdk.ValAddress `json:"operator_address" yaml:"operator_address"`
	// Consensus public key of the validator; bech encoded in JSON
	ConsPubKey crypto.PubKey `json:"consensus_pubkey" yaml:"consensus_pubkey"`
	// Has the validator been jailed from bonded status?
	Jailed bool `json:"jailed" yaml:"jailed"`
	// Has the validator been scheduled to force unbond due to low SelfStake amount compared to TotalDelegationsAmount
	ScheduledToUnbond bool `json:"scheduled_to_unbond" yaml:"scheduled_to_unbond"`
	// Validator status (bonded/unbonding/unbonded)
	Status sdk.BondStatus `json:"status" yaml:"status"`
	// Delegated bonding tokens (incl. self-delegation)
	Bonding ValidatorTokens `json:"bonding" yaml:"bonding"`
	// Delegated liquidity tokens
	LP ValidatorTokens `json:"lp" yaml:"lp"`
	// Description terms for the validator
	Description Description `json:"description" yaml:"description"`
	// If unbonding, height at which this validator has begun unbonding
	UnbondingHeight int64 `json:"unbonding_height" yaml:"unbonding_height"`
	// If unbonding, min time for the validator to complete unbonding
	UnbondingCompletionTime time.Time `json:"unbonding_time" yaml:"unbonding_time"`
	// If ScheduledToUnbond, height at which this schedule started
	ScheduledUnbondHeight int64 `json:"scheduled_unbond_height" yaml:"scheduled_unbond_height"`
	// Is ScheduledToUnbond, min time for the validator to begin force unbond
	ScheduledUnbondStartTime time.Time `json:"scheduled_unbond_time" yaml:"scheduled_unbond_time"`
	// Commission parameters
	Commission Commission `json:"commission" yaml:"commission"`
	// Validator's self declared minimum self delegation
	MinSelfDelegation sdk.Int `json:"min_self_delegation" yaml:"min_self_delegation"`
}

// TestEquivalent checks equality of vital fields of two validators.
func (v Validator) TestEquivalent(v2 Validator) bool {
	return v.ConsPubKey.Equals(v2.ConsPubKey) &&
		bytes.Equal(v.OperatorAddress, v2.OperatorAddress) &&
		v.Status.Equal(v2.Status) &&
		v.Bonding.Equal(v2.Bonding) && v.LP.Equal(v2.LP) &&
		v.Description == v2.Description &&
		v.Commission.Equal(v2.Commission)
}

// ABCIValidatorUpdate returns an abci.ValidatorUpdate from a staking validator type with the full validator power.
func (v Validator) ABCIValidatorUpdate() abci.ValidatorUpdate {
	return abci.ValidatorUpdate{
		PubKey: tmtypes.TM2PB.PubKey(v.ConsPubKey),
		Power:  v.ConsensusPower(),
	}
}

// ABCIValidatorUpdateZero returns an abci.ValidatorUpdate from a staking validator type with zero power used for validator updates.
func (v Validator) ABCIValidatorUpdateZero() abci.ValidatorUpdate {
	return abci.ValidatorUpdate{
		PubKey: tmtypes.TM2PB.PubKey(v.ConsPubKey),
		Power:  0,
	}
}

// SetInitialCommission attempts to set a validator's initial commission.
// An error is returned if the commission is invalid.
func (v Validator) SetInitialCommission(commission Commission) (Validator, error) {
	if err := commission.Validate(); err != nil {
		return v, err
	}
	v.Commission = commission

	return v, nil
}

// InvalidExRate checks if exchange rate is valid.
// In some situations, the exchange rate becomes invalid, e.g. if
// Validator loses all tokens due to slashing. In this case,
// make all future delegations invalid.
func (v Validator) InvalidExRate() bool {
	return v.Bonding.Tokens.IsZero() && v.Bonding.DelegatorShares.IsPositive()
}

// BondedTokens gets the bonded tokens which the validator holds.
func (v Validator) BondedTokens() sdk.Int {
	if v.IsBonded() {
		return v.Bonding.Tokens
	}

	return sdk.ZeroInt()
}

// ConsensusPower gets the consensus-engine power, a reduction of 10^6 from validator tokens is applied.
func (v Validator) ConsensusPower() int64 {
	if v.IsBonded() {
		return v.PotentialConsensusPower()
	}

	return 0
}

// LPPower gets the LP distribution / gov voting power fraction, a reduction of 10^6 from validator tokens is applied.
func (v Validator) LPPower() int64 {
	if v.IsBonded() {
		return v.PotentialLPPower()
	}

	return 0
}

// PotentialConsensusPower returns potential consensus-engine power.
func (v Validator) PotentialConsensusPower() int64 {
	return sdk.TokensToConsensusPower(v.Bonding.Tokens)
}

// PotentialLPPower returns potential LP distribution / gov voting power fraction.
func (v Validator) PotentialLPPower() int64 {
	return sdk.TokensToConsensusPower(v.LP.Tokens)
}

// UpdateStatus updates the location of the bonding shares within a validator to reflect the new status.
func (v Validator) UpdateStatus(newStatus sdk.BondStatus) Validator {
	v.Status = newStatus

	return v
}

// ScheduleValidatorForceUnbond set ScheduledToUnbond state/
func (v Validator) ScheduleValidatorForceUnbond(curBlockHeight int64, curBlockTime time.Time, unbondDelay time.Duration) Validator {
	v.ScheduledToUnbond = true
	v.ScheduledUnbondHeight = curBlockHeight
	v.ScheduledUnbondStartTime = curBlockTime.Add(unbondDelay)

	return v
}

// UnscheduleValidatorForceUnbond drops ScheduledToUnbond state/
func (v Validator) UnscheduleValidatorForceUnbond() Validator {
	v.ScheduledToUnbond = false
	v.ScheduledUnbondHeight = int64(0)
	v.ScheduledUnbondStartTime = time.Unix(0, 0).UTC()

	return v
}

// IsScheduledToUnbond checks if validator is schedulted to force unbond.
func (v Validator) IsScheduledToUnbond() bool {
	return v.ScheduledToUnbond
}

// TotalTokens sums bonding and liquidity tokens.
func (v Validator) TotalTokens() sdk.Int {
	return v.Bonding.Tokens.Add(v.LP.Tokens)
}

// GetTokens returns ValidatorTokens based on delegation operation type.
func (v Validator) GetTokens(delOpType DelegationOpType) ValidatorTokens {
	switch delOpType {
	case BondingDelOpType:
		return v.Bonding
	case LiquidityDelOpType:
		return v.LP
	default:
		panic(delOpType.Validate())
	}
}

// AddTokensFromDel adds tokens to a validator based on delegation operation type.
func (v Validator) AddTokensFromDel(delOpType DelegationOpType, amount sdk.Int) (Validator, sdk.Dec) {
	switch delOpType {
	case BondingDelOpType:
		updTokens, issuedShares := v.Bonding.AddTokensFromDel(amount)
		v.Bonding = updTokens
		return v, issuedShares
	case LiquidityDelOpType:
		updTokens, issuedShares := v.LP.AddTokensFromDel(amount)
		v.LP = updTokens
		return v, issuedShares
	default:
		panic(delOpType.Validate())
	}
}

// RemoveTokens removes tokens from a validator based on delegation operation type.
func (v Validator) RemoveTokens(delOpType DelegationOpType, tokens sdk.Int) Validator {
	switch delOpType {
	case BondingDelOpType:
		updTokens := v.Bonding.RemoveTokens(tokens)
		v.Bonding = updTokens
		return v
	case LiquidityDelOpType:
		updTokens := v.LP.RemoveTokens(tokens)
		v.LP = updTokens
		return v
	default:
		panic(delOpType.Validate())
	}
}

// RemoveDelShares removes delegator shares from a validator based on delegation operation type.
// NOTE: because token fractions are left in the valiadator,
//       the exchange rate of future shares of this validator can increase.
func (v Validator) RemoveDelShares(delOpType DelegationOpType, delShares sdk.Dec) (Validator, sdk.Int) {
	switch delOpType {
	case BondingDelOpType:
		updTokens, issuedTokens := v.Bonding.RemoveDelShares(delShares)
		v.Bonding = updTokens
		return v, issuedTokens
	case LiquidityDelOpType:
		updTokens, issuedTokens := v.LP.RemoveDelShares(delShares)
		v.LP = updTokens
		return v, issuedTokens
	default:
		panic(delOpType.Validate())
	}
}

// nolint - for ValidatorI
func (v Validator) IsJailed() bool                { return v.Jailed }
func (v Validator) GetMoniker() string            { return v.Description.Moniker }
func (v Validator) GetStatus() sdk.BondStatus     { return v.Status }
func (v Validator) IsBonded() bool                { return v.GetStatus().Equal(sdk.Bonded) }
func (v Validator) IsUnbonded() bool              { return v.GetStatus().Equal(sdk.Unbonded) }
func (v Validator) IsUnbonding() bool             { return v.GetStatus().Equal(sdk.Unbonding) }
func (v Validator) GetOperator() sdk.ValAddress   { return v.OperatorAddress }
func (v Validator) GetConsPubKey() crypto.PubKey  { return v.ConsPubKey }
func (v Validator) GetConsAddr() sdk.ConsAddress  { return sdk.ConsAddress(v.ConsPubKey.Address()) }
func (v Validator) GetBondedTokens() sdk.Int      { return v.BondedTokens() }
func (v Validator) GetConsensusPower() int64      { return v.ConsensusPower() }
func (v Validator) GetCommission() sdk.Dec        { return v.Commission.Rate }
func (v Validator) GetMinSelfDelegation() sdk.Int { return v.MinSelfDelegation }

//
func (v Validator) GetBondingDelegatorShares() sdk.Dec { return v.Bonding.DelegatorShares }
func (v Validator) GetBondingTokens() sdk.Int          { return v.Bonding.Tokens }
func (v Validator) BondingTokensFromShares(shares sdk.Dec) sdk.Dec {
	return v.Bonding.TokensFromShares(shares)
}
func (v Validator) BondingTokensFromSharesTruncated(shares sdk.Dec) sdk.Dec {
	return v.Bonding.TokensFromSharesTruncated(shares)
}
func (v Validator) BondingTokensFromSharesRoundUp(shares sdk.Dec) sdk.Dec {
	return v.Bonding.TokensFromSharesRoundUp(shares)
}
func (v Validator) BondingSharesFromTokens(amount sdk.Int) (sdk.Dec, error) {
	return v.Bonding.SharesFromTokens(amount)
}
func (v Validator) BondingSharesFromTokensTruncated(amount sdk.Int) (sdk.Dec, error) {
	return v.Bonding.SharesFromTokensTruncated(amount)
}

//
func (v Validator) GetLPDelegatorShares() sdk.Dec { return v.LP.DelegatorShares }
func (v Validator) GetLPTokens() sdk.Int          { return v.LP.Tokens }
func (v Validator) LPTokensFromShares(shares sdk.Dec) sdk.Dec {
	return v.LP.TokensFromShares(shares)
}
func (v Validator) LPTokensFromSharesTruncated(shares sdk.Dec) sdk.Dec {
	return v.LP.TokensFromSharesTruncated(shares)
}
func (v Validator) LPTokensFromSharesRoundUp(shares sdk.Dec) sdk.Dec {
	return v.LP.TokensFromSharesRoundUp(shares)
}
func (v Validator) LPSharesFromTokens(amount sdk.Int) (sdk.Dec, error) {
	return v.LP.SharesFromTokens(amount)
}
func (v Validator) LPSharesFromTokensTruncated(amount sdk.Int) (sdk.Dec, error) {
	return v.LP.SharesFromTokensTruncated(amount)
}

// MarshalYAML implements custom marshal yaml function due to consensus pubkey.
func (v Validator) MarshalYAML() (interface{}, error) {
	bs, err := yaml.Marshal(struct {
		OperatorAddress          sdk.ValAddress
		ConsPubKey               string
		Jailed                   bool
		ScheduledToUnbond        bool
		Status                   sdk.BondStatus
		BondingDelegatorShares   sdk.Dec
		BondingTokens            sdk.Int
		LPDelegatorShares        sdk.Dec
		LPTokens                 sdk.Int
		Description              Description
		UnbondingHeight          int64
		UnbondingCompletionTime  time.Time
		ScheduledUnbondHeight    int64
		ScheduledUnbondStartTime time.Time
		Commission               Commission
		MinSelfDelegation        sdk.Int
	}{
		OperatorAddress:          v.OperatorAddress,
		ConsPubKey:               sdk.MustBech32ifyPubKey(sdk.Bech32PubKeyTypeConsPub, v.ConsPubKey),
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
		Commission:               v.Commission,
		MinSelfDelegation:        v.MinSelfDelegation,
	})
	if err != nil {
		return nil, err
	}

	return string(bs), nil
}

// bechValidator this is a helper struct used for JSON de- and encoding only.
type bechValidator struct {
	OperatorAddress          sdk.ValAddress `json:"operator_address" yaml:"operator_address"`                 // the bech32 address of the validator's operator
	ConsPubKey               string         `json:"consensus_pubkey" yaml:"consensus_pubkey"`                 // the bech32 consensus public key of the validator
	Jailed                   bool           `json:"jailed" yaml:"jailed"`                                     // has the validator been jailed from bonded status?
	ScheduledToUnbond        bool           `json:"scheduled_to_unbond" yaml:"scheduled_to_unbond"`           // has the validator been scheduled to force unbond due to low SelfStake amount compared to TotalDelegationsAmount
	Status                   sdk.BondStatus `json:"status" yaml:"status"`                                     // validator status (bonded/unbonding/unbonded)
	BondingDelegatorShares   sdk.Dec        `json:"bonding_delegator_shares" yaml:"bonding_delegator_shares"` // bondable tokens: total shares issued to a validator's delegators
	BondingTokens            sdk.Int        `json:"bonding_tokens" yaml:"bonding_tokens"`                     // bondable tokens: delegated tokens (incl. self-delegation)
	LPDelegatorShares        sdk.Dec        `json:"lp_delegator_shares" yaml:"lp_delegator_shares"`           // liquidity tokens: total shares issued to a validator's delegators
	LPTokens                 sdk.Int        `json:"lp_tokens" yaml:"lp_tokens"`                               // liquidity tokens: delegated tokens
	Description              Description    `json:"description" yaml:"description"`                           // description terms for the validator
	UnbondingHeight          int64          `json:"unbonding_height" yaml:"unbonding_height"`                 // if unbonding, height at which this validator has begun unbonding
	UnbondingCompletionTime  time.Time      `json:"unbonding_time" yaml:"unbonding_time"`                     // if unbonding, min time for the validator to complete unbonding
	ScheduledUnbondHeight    int64          `json:"scheduled_unbond_height" yaml:"scheduled_unbond_height"`   // if ScheduledToUnbond, height at which this schedule started
	ScheduledUnbondStartTime time.Time      `json:"scheduled_unbond_time" yaml:"scheduled_unbond_time"`       // is ScheduledToUnbond, min time for the validator to begin force unbond
	Commission               Commission     `json:"commission" yaml:"commission"`                             // commission parameters
	MinSelfDelegation        sdk.Int        `json:"min_self_delegation" yaml:"min_self_delegation"`           // minimum self delegation
}

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

// String returns a human readable string representation of a validator.
func (v Validator) String() string {
	bechConsPubKey, err := sdk.Bech32ifyPubKey(sdk.Bech32PubKeyTypeConsPub, v.ConsPubKey)
	if err != nil {
		panic(err)
	}
	return fmt.Sprintf(`Validator
  Operator Address:            %s
  Validator Consensus Pubkey:  %s
  Jailed:                      %v
  ScheduledToUnbond:           %v
  Status:                      %s
  %s
  %s
  Description:                 %s
  Unbonding Height:            %d
  Unbonding Completion Time:   %v
  Scheduled Unbond Height:     %d
  Scheduled Unbond Start Time: %v
  Minimum Self Delegation:     %v
  Commission:                  %s`,
		v.OperatorAddress, bechConsPubKey,
		v.Jailed, v.ScheduledToUnbond, v.Status,
		v.Bonding.String("Bonding"), v.LP.String("LP"),
		v.Description,
		v.UnbondingHeight, v.UnbondingCompletionTime,
		v.ScheduledUnbondHeight, v.ScheduledUnbondStartTime,
		v.MinSelfDelegation, v.Commission)
}

// NewValidator initializes a new validator.
func NewValidator(operator sdk.ValAddress, pubKey crypto.PubKey, description Description) Validator {
	return Validator{
		OperatorAddress:          operator,
		ConsPubKey:               pubKey,
		Jailed:                   false,
		ScheduledToUnbond:        false,
		Status:                   sdk.Unbonded,
		Bonding:                  NewValidatorTokens(),
		LP:                       NewValidatorTokens(),
		Description:              description,
		UnbondingHeight:          int64(0),
		UnbondingCompletionTime:  time.Unix(0, 0).UTC(),
		ScheduledUnbondHeight:    int64(0),
		ScheduledUnbondStartTime: time.Unix(0, 0).UTC(),
		Commission:               NewCommission(sdk.ZeroDec(), sdk.ZeroDec(), sdk.ZeroDec()),
		MinSelfDelegation:        sdk.OneInt(),
	}
}

// MustMarshalValidator serializes Validator for storage.
func MustMarshalValidator(cdc *codec.Codec, validator Validator) []byte {
	return cdc.MustMarshalBinaryLengthPrefixed(validator)
}

// MustMarshalValidator unserializes Validator from the storage.
func MustUnmarshalValidator(cdc *codec.Codec, value []byte) Validator {
	validator, err := UnmarshalValidator(cdc, value)
	if err != nil {
		panic(err)
	}
	return validator
}

// UnmarshalValidator unserializes Validator from the storage without a panic.
func UnmarshalValidator(cdc *codec.Codec, value []byte) (validator Validator, err error) {
	err = cdc.UnmarshalBinaryLengthPrefixed(value, &validator)
	return validator, err
}
