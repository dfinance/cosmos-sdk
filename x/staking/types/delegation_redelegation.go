package types

import (
	"bytes"
	"fmt"
	"strings"
	"time"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// DVVTriplet is struct that just has a delegator-validator-validator triplet with no other data.
// It is intended to be used as a marshalable pointer. For example, a DVVTriplet can be used to construct the
// key to getting a Redelegation from state.
type DVVTriplet struct {
	DelegatorAddress    sdk.AccAddress
	ValidatorSrcAddress sdk.ValAddress
	ValidatorDstAddress sdk.ValAddress
}

func (t DVVTriplet) Equal(t2 DVVTriplet) bool {
	if !t.DelegatorAddress.Equals(t2.DelegatorAddress) {
		return false
	}
	if !t.ValidatorSrcAddress.Equals(t2.ValidatorSrcAddress) {
		return false
	}
	if !t.ValidatorDstAddress.Equals(t2.ValidatorDstAddress) {
		return false
	}

	return true
}

// Redelegation contains the list of a particular delegator's redelegating bonds from a particular source validator
// to a particular destination validator.
type Redelegation struct {
	DelegatorAddress    sdk.AccAddress      `json:"delegator_address" yaml:"delegator_address"`         // delegator
	ValidatorSrcAddress sdk.ValAddress      `json:"validator_src_address" yaml:"validator_src_address"` // validator redelegation source operator addr
	ValidatorDstAddress sdk.ValAddress      `json:"validator_dst_address" yaml:"validator_dst_address"` // validator redelegation destination operator addr
	Entries             []RedelegationEntry `json:"entries" yaml:"entries"`                             // redelegation entries
}

// RedelegationEntry - entry to a Redelegation.
// Entry type defines target tokens: bonding / liquidity.
type RedelegationEntry struct {
	// Height at which the redelegation took place
	CreationHeight int64 `json:"creation_height" yaml:"creation_height"`
	// Time at which the redelegation will complete
	CompletionTime time.Time `json:"completion_time" yaml:"completion_time"`
	// Operation type
	OpType DelegationOpType `json:"op_type" yaml:"op_type"`
	// Initial balance when redelegation started
	InitialBalance sdk.Int `json:"initial_balance" yaml:"initial_balance"`
	// Amount of destination-validator shares created by redelegation
	SharesDst sdk.Dec `json:"shares_dst" yaml:"shares_dst"`
}

// IsMature - is the current entry mature.
func (e RedelegationEntry) IsMature(currentTime time.Time) bool {
	return !e.CompletionTime.After(currentTime)
}

// AddEntry - append entry to the unbonding delegation.
func (d *Redelegation) AddEntry(
	creationHeight int64, minTime time.Time,
	opType DelegationOpType, balance sdk.Int, sharesDst sdk.Dec,
) {

	entry := NewRedelegationEntry(creationHeight, minTime, opType, balance, sharesDst)
	d.Entries = append(d.Entries, entry)
}

// RemoveEntry - remove entry at index i to the unbonding delegation.
func (d *Redelegation) RemoveEntry(i int64) {
	d.Entries = append(d.Entries[:i], d.Entries[i+1:]...)
}

// nolint
// inefficient but only used in tests
func (d Redelegation) Equal(d2 Redelegation) bool {
	bz1 := ModuleCdc.MustMarshalBinaryLengthPrefixed(&d)
	bz2 := ModuleCdc.MustMarshalBinaryLengthPrefixed(&d2)
	return bytes.Equal(bz1, bz2)
}

// String returns a human readable string representation of a Redelegation.
func (d Redelegation) String() string {
	out := fmt.Sprintf(`Redelegations between:
  Delegator:                 %s
  Source Validator:          %s
  Destination Validator:     %s
  Entries:
`,
		d.DelegatorAddress, d.ValidatorSrcAddress, d.ValidatorDstAddress,
	)

	for i, entry := range d.Entries {
		out += fmt.Sprintf(`    Redelegation Entry #%d:
      Operation type:            %s
      Creation height:           %v
      Min time to unbond (unix): %v
      Dest Shares:               %s`,
			i, entry.OpType, entry.CreationHeight, entry.CompletionTime, entry.SharesDst,
		)
	}

	return strings.TrimRight(out, "\n")
}

// NewRedelegation - create a new redelegation object.
func NewRedelegation(
	delegatorAddr sdk.AccAddress, validatorSrcAddr, validatorDstAddr sdk.ValAddress,
	creationHeight int64, minTime time.Time,
	opType DelegationOpType, balance sdk.Int, sharesDst sdk.Dec,
) Redelegation {

	entry := NewRedelegationEntry(creationHeight, minTime, opType, balance, sharesDst)

	return Redelegation{
		DelegatorAddress:    delegatorAddr,
		ValidatorSrcAddress: validatorSrcAddr,
		ValidatorDstAddress: validatorDstAddr,
		Entries:             []RedelegationEntry{entry},
	}
}

// NewRedelegationEntry - create a new redelegation object.
func NewRedelegationEntry(
	creationHeight int64, completionTime time.Time,
	opType DelegationOpType, balance sdk.Int, sharesDst sdk.Dec,
) RedelegationEntry {

	return RedelegationEntry{
		CreationHeight: creationHeight,
		CompletionTime: completionTime,
		OpType:         opType,
		InitialBalance: balance,
		SharesDst:      sharesDst,
	}
}

// Redelegations are a collection of Redelegation.
type Redelegations []Redelegation

func (d Redelegations) String() (out string) {
	for _, red := range d {
		out += red.String() + "\n"
	}
	return strings.TrimSpace(out)
}

// MustMarshalRED returns the Redelegation bytes. Panics if fails.
func MustMarshalRED(cdc *codec.Codec, red Redelegation) []byte {
	return cdc.MustMarshalBinaryLengthPrefixed(red)
}

// MustUnmarshalRED unmarshals a redelegation from a store value. Panics if fails.
func MustUnmarshalRED(cdc *codec.Codec, value []byte) Redelegation {
	red, err := UnmarshalRED(cdc, value)
	if err != nil {
		panic(err)
	}
	return red
}

// UnmarshalRED unmarshals a redelegation from a store value without a panic on error.
func UnmarshalRED(cdc *codec.Codec, value []byte) (red Redelegation, err error) {
	err = cdc.UnmarshalBinaryLengthPrefixed(value, &red)
	return red, err
}
