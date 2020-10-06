package types

import (
	"bytes"
	"fmt"
	"strings"
	"time"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// UnbondingDelegation stores all of a single delegator's unbonding bonds for a single validator in an time-ordered list.
type UnbondingDelegation struct {
	DelegatorAddress sdk.AccAddress             `json:"delegator_address" yaml:"delegator_address"` // delegator
	ValidatorAddress sdk.ValAddress             `json:"validator_address" yaml:"validator_address"` // validator unbonding from operator addr
	Entries          []UnbondingDelegationEntry `json:"entries" yaml:"entries"`                     // unbonding delegation entries
}

// UnbondingDelegationEntry - entry to an UnbondingDelegation.
// Entry type defines target tokens: bonding / liquidity.
type UnbondingDelegationEntry struct {
	// Height which the unbonding took place
	CreationHeight int64 `json:"creation_height" yaml:"creation_height"`
	// Time at which the unbonding delegation will complete
	CompletionTime time.Time `json:"completion_time" yaml:"completion_time"`
	// Operation type
	OpType DelegationOpType `json:"op_type" yaml:"op_type"`
	// Tokens initially scheduled to receive at completion
	InitialBalance sdk.Int `json:"initial_balance" yaml:"initial_balance"`
	// Tokens to receive at completion
	Balance sdk.Int `json:"balance" yaml:"balance"`
}

// IsMature - is the current entry mature.
func (e UnbondingDelegationEntry) IsMature(currentTime time.Time) bool {
	return !e.CompletionTime.After(currentTime)
}

// AddEntry - append entry to the unbonding delegation.
func (d *UnbondingDelegation) AddEntry(
	creationHeight int64, minTime time.Time,
	opType DelegationOpType, balance sdk.Int,
) {

	entry := NewUnbondingDelegationEntry(creationHeight, minTime, opType, balance)
	d.Entries = append(d.Entries, entry)
}

// RemoveEntry - remove entry at index i to the unbonding delegation.
func (d *UnbondingDelegation) RemoveEntry(i int64) {
	d.Entries = append(d.Entries[:i], d.Entries[i+1:]...)
}

// Equal test equality of two UnbondingDelegation objects.
// Inefficient but only used in testing.
// nolint
func (d UnbondingDelegation) Equal(d2 UnbondingDelegation) bool {
	bz1 := ModuleCdc.MustMarshalBinaryLengthPrefixed(&d)
	bz2 := ModuleCdc.MustMarshalBinaryLengthPrefixed(&d2)
	return bytes.Equal(bz1, bz2)
}

// String returns a human readable string representation of an UnbondingDelegation.
func (d UnbondingDelegation) String() string {
	out := fmt.Sprintf(`Unbonding Delegations between:
  Delegator:                 %s
  Validator:                 %s
	Entries:
`,
		d.DelegatorAddress, d.ValidatorAddress,
	)

	for i, entry := range d.Entries {
		out += fmt.Sprintf(`    Unbonding Delegation %d:
      Operation type:            %s
      Creation Height:           %v
      Min time to unbond (unix): %v
      Expected balance:          %s`,
			i, entry.OpType, entry.CreationHeight, entry.CompletionTime, entry.Balance,
		)
	}

	return out
}

// NewUnbondingDelegation - create a new unbonding delegation object.
func NewUnbondingDelegation(delegatorAddr sdk.AccAddress, validatorAddr sdk.ValAddress,
	creationHeight int64, minTime time.Time,
	opType DelegationOpType, balance sdk.Int,
) UnbondingDelegation {

	entry := NewUnbondingDelegationEntry(creationHeight, minTime, opType, balance)
	return UnbondingDelegation{
		DelegatorAddress: delegatorAddr,
		ValidatorAddress: validatorAddr,
		Entries:          []UnbondingDelegationEntry{entry},
	}
}

// NewUnbondingDelegationEntry - create a new unbonding delegation object.
func NewUnbondingDelegationEntry(
	creationHeight int64, completionTime time.Time,
	opType DelegationOpType, balance sdk.Int,
) UnbondingDelegationEntry {

	return UnbondingDelegationEntry{
		CreationHeight: creationHeight,
		CompletionTime: completionTime,
		OpType:         opType,
		InitialBalance: balance,
		Balance:        balance,
	}
}

// UnbondingDelegations is a collection of UnbondingDelegation.
type UnbondingDelegations []UnbondingDelegation

func (ubds UnbondingDelegations) String() (out string) {
	for _, u := range ubds {
		out += u.String() + "\n"
	}
	return strings.TrimSpace(out)
}

// MustMarshalUBD returns serialized unbonding delegation.
func MustMarshalUBD(cdc *codec.Codec, ubd UnbondingDelegation) []byte {
	return cdc.MustMarshalBinaryLengthPrefixed(ubd)
}

// MustUnmarshalUBD returns deserialized unbonding delegation. Panic on error.
func MustUnmarshalUBD(cdc *codec.Codec, value []byte) UnbondingDelegation {
	ubd, err := UnmarshalUBD(cdc, value)
	if err != nil {
		panic(err)
	}
	return ubd
}

// UnmarshalUBD returns deserialized unbonding delegation without a panic on error.
func UnmarshalUBD(cdc *codec.Codec, value []byte) (ubd UnbondingDelegation, err error) {
	err = cdc.UnmarshalBinaryLengthPrefixed(value, &ubd)
	return ubd, err
}
