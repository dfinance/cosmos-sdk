package types

import (
	"bytes"
	"sort"
	"strings"

	"github.com/cosmos/cosmos-sdk/x/staking/exported"
)

// Validators is a collection of Validator
type Validators []Validator

func (v Validators) String() (out string) {
	for _, val := range v {
		out += val.String() + "\n"
	}
	return strings.TrimSpace(out)
}

// ToSDKValidators -  convenience function convert []Validators to []sdk.Validators
func (v Validators) ToSDKValidators() (validators []exported.ValidatorI) {
	for _, val := range v {
		validators = append(validators, val)
	}
	return validators
}

// Sort Validators sorts validator array in ascending operator address order
func (v Validators) Sort() {
	sort.Sort(v)
}

// Implements sort interface
func (v Validators) Len() int {
	return len(v)
}

// Implements sort interface
func (v Validators) Less(i, j int) bool {
	return bytes.Compare(v[i].OperatorAddress, v[j].OperatorAddress) == -1
}

// Implements sort interface
func (v Validators) Swap(i, j int) {
	it := v[i]
	v[i] = v[j]
	v[j] = it
}
