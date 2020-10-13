package types

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/staking"
)

// ValidatorResp contains staking.Validator extended with distribution info.
type ValidatorResp struct {
	// Address of the validator's operator; bech encoded in JSON
	OperatorAddress sdk.ValAddress `json:"operator_address" yaml:"operator_address"`
	// Consensus public key of the validator; bech encoded in JSON
	ConsPubKey string `json:"consensus_pubkey" yaml:"consensus_pubkey"`
	// Has the validator been jailed from bonded status?
	Jailed bool `json:"jailed" yaml:"jailed"`
	// Has the validator been scheduled to force unbond due to low SelfStake amount compared to TotalDelegationsAmount
	ScheduledToUnbond bool `json:"scheduled_to_unbond" yaml:"scheduled_to_unbond"`
	// Validator status (bonded/unbonding/unbonded)
	Status sdk.BondStatus `json:"status" yaml:"status"`
	// Bondable tokens: total shares issued to a validator's delegators
	BondingDelegatorShares sdk.Dec `json:"bonding_delegator_shares" yaml:"bonding_delegator_shares"`
	// Bondable tokens: delegated tokens (incl. self-delegation)
	BondingTokens sdk.Int `json:"bonding_tokens" yaml:"bonding_tokens"`
	// Liquidity tokens: total shares issued to a validator's delegators
	LPDelegatorShares sdk.Dec `json:"lp_delegator_shares" yaml:"lp_delegator_shares"`
	// Liquidity tokens: delegated tokens
	LPTokens sdk.Int `json:"lp_tokens" yaml:"lp_tokens"`
	// Description terms for the validator
	Description staking.Description `json:"description" yaml:"description"`
	// If unbonding, height at which this validator has begun unbonding
	UnbondingHeight int64 `json:"unbonding_height" yaml:"unbonding_height"`
	// If unbonding, min time for the validator to complete unbonding
	UnbondingCompletionTime time.Time `json:"unbonding_time" yaml:"unbonding_time"`
	// If ScheduledToUnbond, height at which this schedule started
	ScheduledUnbondHeight int64 `json:"scheduled_unbond_height" yaml:"scheduled_unbond_height"`
	// Is ScheduledToUnbond, min time for the validator to begin force unbond
	ScheduledUnbondStartTime time.Time `json:"scheduled_unbond_time" yaml:"scheduled_unbond_time"`
	// Commission parameters
	Commission staking.Commission `json:"commission" yaml:"commission"`
	// Validator's self declared minimum self delegation
	MinSelfDelegation sdk.Int `json:"min_self_delegation" yaml:"min_self_delegation"`
	// Max bonding delegations level
	MaxBondingDelegationsLvl sdk.Int `json:"max_bonding_delegations_lvl" yaml:"max_bonding_delegations_lvl"`
	// Bonding tokens rewards distribution power
	BondingDistributionPower int64 `json:"bonding_distribution_power" yaml:"bonding_distribution_power"`
	// LP tokens rewards distribution power
	LPDistributionPower int64 `json:"lp_distribution_power" yaml:"lp_distribution_power"`
	// Rewards locked flag
	RewardsLocked     bool      `json:"rewards_locked" yaml:"rewards_locked"`
	// Rewards unlock time (if locked)
	RewardsUnlockTime time.Time `json:"rewards_unlock_time" yaml:"rewards_unlock_time"`
}

// NewValidatorResp builds a new ValidatorResp object.
func NewValidatorResp(
	validator staking.Validator,
	stakingState staking.ValidatorStakingState,
	lockedState ValidatorLockedRewardsState,
	maxDelegationsRatio sdk.Dec,
	distPwr, lpPwr int64,
) (ValidatorResp, error) {

	bechConsPubKey, err := sdk.Bech32ifyPubKey(sdk.Bech32PubKeyTypeConsPub, validator.ConsPubKey)
	if err != nil {
		return ValidatorResp{}, err
	}

	var selfStake sdk.Int
	if !stakingState.Operator.BondingShares.IsZero() {
		selfStake = validator.BondingTokensFromShares(stakingState.Operator.BondingShares).TruncateInt()
	}
	maxDelegationsLvl := selfStake.ToDec().Mul(maxDelegationsRatio).TruncateInt()

	return ValidatorResp{
		OperatorAddress:          validator.OperatorAddress,
		ConsPubKey:               bechConsPubKey,
		Jailed:                   validator.Jailed,
		ScheduledToUnbond:        validator.ScheduledToUnbond,
		Status:                   validator.Status,
		BondingDelegatorShares:   validator.Bonding.DelegatorShares,
		BondingTokens:            validator.Bonding.Tokens,
		LPDelegatorShares:        validator.LP.DelegatorShares,
		LPTokens:                 validator.LP.Tokens,
		Description:              validator.Description,
		UnbondingHeight:          validator.UnbondingHeight,
		UnbondingCompletionTime:  validator.UnbondingCompletionTime,
		ScheduledUnbondHeight:    validator.ScheduledUnbondHeight,
		ScheduledUnbondStartTime: validator.ScheduledUnbondStartTime,
		Commission:               validator.Commission,
		MinSelfDelegation:        validator.MinSelfDelegation,
		MaxBondingDelegationsLvl: maxDelegationsLvl,
		BondingDistributionPower: distPwr,
		LPDistributionPower:      lpPwr,
		RewardsLocked:            lockedState.IsLocked(),
		RewardsUnlockTime:        lockedState.UnlocksAt,
	}, nil
}
