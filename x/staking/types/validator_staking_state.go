package types

import (
	"bytes"
	"fmt"
	"sort"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// ValidatorStakingState keeps validator staking state.
// Used to avoid iterating over all validator delegations on delegate/undelegate/redelegate operations (gas reduction)
// ValidatorAddress is encoded within storage key.
type ValidatorStakingState struct {
	// Validator operator delegation data
	Operator DelegationTruncated `json:"operator"`
	// Delegators delegation data excluding the operators
	Delegators []DelegationTruncated `json:"delegators"`
}

// DelegationTruncated keeps essential data used by ValidatorStakingState.
type DelegationTruncated struct {
	Address       sdk.AccAddress `json:"address"`
	BondingShares sdk.Dec        `json:"bonding_shares"`
	LPShares      sdk.Dec        `json:"lp_shares"`
}

// Sort sorts state.Delegators.
// Sort is performed on every store.Set operation to achieve genesis consistency.
func (s ValidatorStakingState) Sort() {
	sort.Sort(s)
}

// Len implements sort interface for state.Delegators.
func (s ValidatorStakingState) Len() int {
	return len(s.Delegators)
}

// Less implements sort interface for state.Delegators.
func (s ValidatorStakingState) Less(i, j int) bool {
	return bytes.Compare(s.Delegators[i].Address, s.Delegators[j].Address) == -1
}

// Swap implements sort interface for state.Delegators.
func (s ValidatorStakingState) Swap(i, j int) {
	s.Delegators[i], s.Delegators[j] = s.Delegators[j], s.Delegators[i]
}

// SetDelegator adds / updates delegation info.
func (s ValidatorStakingState) SetDelegator(
	validatorAddr sdk.ValAddress, delegatorAddr sdk.AccAddress,
	bondingShares, lpShares sdk.Dec,
) ValidatorStakingState {

	if validatorAddr.Equals(delegatorAddr) {
		s.Operator.Address = delegatorAddr
		s.Operator.BondingShares = bondingShares
		s.Operator.LPShares = lpShares
		return s
	}

	for i := 0; i < len(s.Delegators); i++ {
		if s.Delegators[i].Address.Equals(delegatorAddr) {
			s.Delegators[i].BondingShares = bondingShares
			s.Delegators[i].LPShares = lpShares
			return s
		}
	}

	s.Delegators = append(s.Delegators, DelegationTruncated{
		Address:       delegatorAddr,
		BondingShares: bondingShares,
		LPShares:      lpShares,
	})

	return s
}

// RemoveDelegator removes delegation info.
// nolint: interfacer
func (s ValidatorStakingState) RemoveDelegator(delegatorAddr sdk.AccAddress) ValidatorStakingState {
	if s.Operator.Address.Equals(delegatorAddr) {
		s.Operator.Address = sdk.AccAddress{}
		s.Operator.BondingShares = sdk.ZeroDec()
		s.Operator.LPShares = sdk.ZeroDec()
		return s
	}

	rmIdx := -1
	for i := 0; i < len(s.Delegators); i++ {
		if s.Delegators[i].Address.Equals(delegatorAddr) {
			rmIdx = i
			break
		}
	}

	if rmIdx >= 0 {
		s.Delegators = append(s.Delegators[:rmIdx], s.Delegators[rmIdx+1:]...)
	}

	return s
}

// GetSelfAndTotalStakes returns selfStake and totalStakes values for bonding tokens.
func (s ValidatorStakingState) GetSelfAndTotalStakes(validator Validator) (selfStake, totalStakes sdk.Int) {
	selfStake, totalStakes = sdk.ZeroInt(), sdk.ZeroInt()

	getTokens := func(shares sdk.Dec) sdk.Int {
		return validator.BondingTokensFromShares(shares).TruncateInt()
	}

	if !s.Operator.BondingShares.IsZero() {
		selfStake = selfStake.Add(getTokens(s.Operator.BondingShares))
	}

	for i := 0; i < len(s.Delegators); i++ {
		delegation := &s.Delegators[i]
		if delegation.BondingShares.IsZero() {
			continue
		}

		totalStakes = totalStakes.Add(getTokens(delegation.BondingShares))
	}
	totalStakes = totalStakes.Add(selfStake)

	return
}

// InvariantCheck verifies that delegation exists in the state and it is correct.
// Used by module invariants check.
func (s ValidatorStakingState) InvariantCheck(validator Validator, delegation Delegation) error {
	const msgPrefixFmt = "broken validator %s staking state for delegator %s:\n"

	if s.Operator.Address.Equals(delegation.DelegatorAddress) {
		if !s.Operator.Address.Equals(validator.OperatorAddress) {
			return fmt.Errorf(msgPrefixFmt+"\tinvalid operator address: %s",
				validator.OperatorAddress, delegation.DelegatorAddress, s.Operator.Address,
			)
		}
		if !s.Operator.BondingShares.Equal(delegation.BondingShares) {
			return fmt.Errorf(msgPrefixFmt+"\tinvalid operator BondingShares: %s / %s",
				validator.OperatorAddress, delegation.DelegatorAddress, s.Operator.BondingShares, delegation.BondingShares,
			)
		}
		if !s.Operator.LPShares.Equal(delegation.LPShares) {
			return fmt.Errorf(msgPrefixFmt+"\tinvalid operator LPShares: %s / %s",
				validator.OperatorAddress, delegation.DelegatorAddress, s.Operator.LPShares, delegation.LPShares,
			)
		}
		return nil
	}

	for i := 0; i < len(s.Delegators); i++ {
		stateDel := &s.Delegators[i]
		if delegation.DelegatorAddress.Equals(stateDel.Address) {
			if !delegation.BondingShares.Equal(stateDel.BondingShares) {
				return fmt.Errorf(msgPrefixFmt+"\tinvalid BondingShares: %s / %s",
					validator.OperatorAddress, delegation.DelegatorAddress, stateDel.BondingShares, delegation.BondingShares,
				)
			}
			if !delegation.LPShares.Equal(stateDel.LPShares) {
				return fmt.Errorf(msgPrefixFmt+"\tinvalid LPShares: %s / %s",
					validator.OperatorAddress, delegation.DelegatorAddress, stateDel.LPShares, delegation.LPShares,
				)
			}
			return nil
		}
	}

	return fmt.Errorf(msgPrefixFmt+"not found",
		validator.OperatorAddress, delegation.DelegatorAddress,
	)
}

// NewValidatorStakingState creates an empty ValidatorStakingState object.
func NewValidatorStakingState() ValidatorStakingState {
	return ValidatorStakingState{
		Operator: DelegationTruncated{
			Address:       sdk.AccAddress{},
			BondingShares: sdk.ZeroDec(),
			LPShares:      sdk.ZeroDec(),
		},
		Delegators: make([]DelegationTruncated, 0),
	}
}
