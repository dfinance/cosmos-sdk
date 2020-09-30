package types

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/staking/exported"
)

// Implements Delegation interface
var _ exported.DelegationI = Delegation{}

// DVPair is struct that just has a delegator-validator pair with no other data.
// It is intended to be used as a marshalable pointer. For example, a DVPair can be used to construct the
// key to getting an UnbondingDelegation from state.
type DVPair struct {
	DelegatorAddress sdk.AccAddress
	ValidatorAddress sdk.ValAddress
}

func (p DVPair) Equal(p2 DVPair) bool {
	if !p.DelegatorAddress.Equals(p2.DelegatorAddress) {
		return false
	}
	if !p.ValidatorAddress.Equals(p2.ValidatorAddress) {
		return false
	}

	return true
}

// Delegation represents the bond with tokens held by an account.
// It is owned by one delegator, and is associated with the voting power of one validator.
// Delegation target might be either bonding tokens or liquidity tokens.
// Object is deleted if bonding and liquidity shares are zero.
type Delegation struct {
	DelegatorAddress sdk.AccAddress `json:"delegator_address" yaml:"delegator_address"`
	ValidatorAddress sdk.ValAddress `json:"validator_address" yaml:"validator_address"`
	BondingShares    sdk.Dec        `json:"bonding_shares" yaml:"bonding_shares"`
	LPShares         sdk.Dec        `json:"lp_shares" yaml:"lp_shares"`
}

// Equal check equality of two Delegations.
// nolint
func (d Delegation) Equal(d2 Delegation) bool {
	return bytes.Equal(d.DelegatorAddress, d2.DelegatorAddress) &&
		bytes.Equal(d.ValidatorAddress, d2.ValidatorAddress) &&
		d.BondingShares.Equal(d2.BondingShares) &&
		d.LPShares.Equal(d2.LPShares)
}

// TotalShares sums bonding and liquidity shares.
func (d Delegation) TotalShares() sdk.Dec {
	return d.BondingShares.Add(d.LPShares)
}

// GetShares gets a shares value based on delegation operation.
// Panics on wrong DelegationOpType enum value.
func (d Delegation) GetShares(delOpType DelegationOpType) sdk.Dec {
	switch delOpType {
	case BondingDelOpType:
		return d.BondingShares
	case LiquidityDelOpType:
		return d.LPShares
	default:
		panic(delOpType.Validate())
	}
}

// AddShares add a value to delegation share based on delegation operation.
// Panics on wrong DelegationOpType enum value.
func (d Delegation) AddShares(delOpType DelegationOpType, shares sdk.Dec) Delegation {
	switch delOpType {
	case BondingDelOpType:
		d.BondingShares = d.BondingShares.Add(shares)
	case LiquidityDelOpType:
		d.LPShares = d.LPShares.Add(shares)
	default:
		panic(delOpType.Validate())
	}

	return d
}

// AddShares add a value to delegation share based on delegation operation.
// Panics on wrong DelegationOpType enum value.
func (d Delegation) SubShares(delOpType DelegationOpType, shares sdk.Dec) Delegation {
	switch delOpType {
	case BondingDelOpType:
		d.BondingShares = d.BondingShares.Sub(shares)
	case LiquidityDelOpType:
		d.LPShares = d.LPShares.Sub(shares)
	default:
		panic(delOpType.Validate())
	}

	return d
}

// nolint - for DelegationI
func (d Delegation) GetDelegatorAddr() sdk.AccAddress { return d.DelegatorAddress }
func (d Delegation) GetValidatorAddr() sdk.ValAddress { return d.ValidatorAddress }
func (d Delegation) GetBondingShares() sdk.Dec        { return d.BondingShares }
func (d Delegation) GetLPShares() sdk.Dec             { return d.LPShares }

// String returns a human readable string representation of a Delegation.
func (d Delegation) String() string {
	return fmt.Sprintf(`Delegation:
  Delegator:        %s
  Validator:        %s
  Bonding Shares:   %s
  Liquidity Shares: %s`,
		d.DelegatorAddress, d.ValidatorAddress, d.BondingShares, d.LPShares,
	)
}

// NewDelegation creates a new delegation object.
func NewDelegation(
	delegatorAddr sdk.AccAddress, validatorAddr sdk.ValAddress,
	bondingShares, lpShares sdk.Dec) Delegation {

	return Delegation{
		DelegatorAddress: delegatorAddr,
		ValidatorAddress: validatorAddr,
		BondingShares:    bondingShares,
		LPShares:         lpShares,
	}
}

// MustMarshalDelegation returns the delegation bytes.
// Panics if fails.
func MustMarshalDelegation(cdc *codec.Codec, delegation Delegation) []byte {
	return cdc.MustMarshalBinaryLengthPrefixed(delegation)
}

// MustUnmarshalDelegation returns the unmarshaled delegation from bytes.
// Panics if fails.
func MustUnmarshalDelegation(cdc *codec.Codec, value []byte) Delegation {
	delegation, err := UnmarshalDelegation(cdc, value)
	if err != nil {
		panic(err)
	}
	return delegation
}

// UnmarshalDelegation returns the unmarshaled delegation from bytes without a panic.
func UnmarshalDelegation(cdc *codec.Codec, value []byte) (delegation Delegation, err error) {
	err = cdc.UnmarshalBinaryLengthPrefixed(value, &delegation)
	return delegation, err
}

// Delegations is a collection of delegations.
type Delegations []Delegation

func (d Delegations) String() (out string) {
	for _, del := range d {
		out += del.String() + "\n"
	}
	return strings.TrimSpace(out)
}
