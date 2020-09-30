package types

import (
	"fmt"
	"strings"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// ValidatorHistoricalRewards keeps historical rewards for a validator.
// Height is implicit within the store key.
// ReferenceCount =
//    number of outstanding delegations which ended the associated period (and might need to read that record)
//  + number of slashes which ended the associated period (and might need to read that record)
//  + one per validator for the zeroeth period, set on initialization
type ValidatorHistoricalRewards struct {
	// Sum from the zeroeth period until this period of rewards / tokens, per the spec
	CumulativeBondingRewardRatio sdk.DecCoins `json:"cumulative_bonding_reward_ratio" yaml:"cumulative_bonding_reward_ratio"`
	CumulativeLPRewardRatio      sdk.DecCoins `json:"cumulative_lp_reward_ratio" yaml:"cumulative_lp_reward_ratio"`
	// Indicates the number of objects which might need to reference this historical entry at any point
	ReferenceCount uint16 `json:"reference_count" yaml:"reference_count"`
}

// NewValidatorHistoricalRewards creates a new ValidatorHistoricalRewards.
func NewValidatorHistoricalRewards(
	cumulativeBondingRewardRatio, cumulativeLPRewardRatio sdk.DecCoins,
	referenceCount uint16,
) ValidatorHistoricalRewards {

	return ValidatorHistoricalRewards{
		CumulativeBondingRewardRatio: cumulativeBondingRewardRatio,
		CumulativeLPRewardRatio:      cumulativeLPRewardRatio,
		ReferenceCount:               referenceCount,
	}
}

// ValidatorCurrentRewards keeps current rewards and current period for a validator.
// Kept as a running counter and incremented each block as long as the validator's tokens remain constant.
type ValidatorCurrentRewards struct {
	// Current bonding tokens rewards
	BondingRewards sdk.DecCoins `json:"bonding_rewards" yaml:"bonding_rewards"`
	// Current liquidity tokens rewards
	LPRewards sdk.DecCoins `json:"lp_rewards" yaml:"lp_rewards"`
	// Current period
	Period uint64 `json:"period" yaml:"period"`
}

// NewValidatorCurrentRewards creates a new ValidatorCurrentRewards.
func NewValidatorCurrentRewards(bondingRewards, lpRewards sdk.DecCoins, period uint64) ValidatorCurrentRewards {
	return ValidatorCurrentRewards{
		BondingRewards: bondingRewards,
		LPRewards:      lpRewards,
		Period:         period,
	}
}

// ValidatorAccumulatedCommission keeps accumulated commission for a validator.
// Kept as a running counter, can be withdrawn at any time.
type ValidatorAccumulatedCommission = sdk.DecCoins

// InitialValidatorAccumulatedCommission returns the initial accumulated commission (zero).
func InitialValidatorAccumulatedCommission() ValidatorAccumulatedCommission {
	return ValidatorAccumulatedCommission{}
}

// ValidatorSlashEvent needed to calculate appropriate amounts of staking token
// for delegations which withdraw after a slash has occurred.
// Height is implicit within the store key.
type ValidatorSlashEvent struct {
	// Period when the slash occurred
	ValidatorPeriod uint64 `json:"validator_period" yaml:"validator_period"`
	// Slash fraction
	Fraction sdk.Dec `json:"fraction" yaml:"fraction"`
}

// NewValidatorSlashEvent creates a new ValidatorSlashEvent.
func NewValidatorSlashEvent(validatorPeriod uint64, fraction sdk.Dec) ValidatorSlashEvent {
	return ValidatorSlashEvent{
		ValidatorPeriod: validatorPeriod,
		Fraction:        fraction,
	}
}

func (vs ValidatorSlashEvent) String() string {
	return fmt.Sprintf(`Period:   %d
Fraction: %s`, vs.ValidatorPeriod, vs.Fraction)
}

// ValidatorSlashEvents is a collection of ValidatorSlashEvent.
type ValidatorSlashEvents []ValidatorSlashEvent

func (vs ValidatorSlashEvents) String() string {
	out := "Validator Slash Events:\n"
	for i, sl := range vs {
		out += fmt.Sprintf(`  Slash %d:
    Period:   %d
    Fraction: %s
`, i, sl.ValidatorPeriod, sl.Fraction)
	}
	return strings.TrimSpace(out)
}

// ValidatorOutstandingRewards keeps outstanding (un-withdrawn) rewards for a validator.
// It is inexpensive to track, allows simple sanity checks.
type ValidatorOutstandingRewards = sdk.DecCoins

// ValidatorLockedRewardsState contains locked rewards data.
type ValidatorLockedRewardsState struct {
	// Rewards lock block height
	LockHeight int64 `json:"lock_height" yaml:"lock_height"`
	// Rewards lock timestamp
	LockedAt time.Time `json:"locked_at" yaml:"locked_at"`
	// Rewards are locked until
	UnlocksAt time.Time `json:"unlocks_at" yaml:"unlocks_at"`
	// Locked shares to all shares relation (zero if there is no locking)
	LockedRatio sdk.Dec `json:"locked_ratio" yaml:"locked_ratio"`
	// Lock auto-renewal flag
	AutoRenewal bool `json:"auto_renewal" yaml:"auto_renewal"`
}

// GetDistributionPower calculates validator distribution power depending on the lock state.
func (l ValidatorLockedRewardsState) GetDistributionPower(stakingPower int64) int64 {
	if !l.IsLocked() {
		return stakingPower
	}
	lockedPower := sdk.NewDec(stakingPower).Mul(l.LockedRatio)

	return stakingPower + lockedPower.TruncateInt64()
}

// Lock locks current state.
func (l ValidatorLockedRewardsState) Lock(lockRatio sdk.Dec, lockDuration time.Duration, currTime time.Time, currHeight int64) ValidatorLockedRewardsState {
	l.LockedRatio = lockRatio
	l.LockedAt = currTime
	l.LockHeight = currHeight
	l.UnlocksAt = currTime.Add(lockDuration)
	l.AutoRenewal = true

	return l
}

// RenewLock renews current locked state.
func (l ValidatorLockedRewardsState) RenewLock(lockRatio sdk.Dec, lockDuration time.Duration, currTime time.Time) ValidatorLockedRewardsState {
	l.LockedRatio = lockRatio
	l.UnlocksAt = currTime.Add(lockDuration)
	l.AutoRenewal = true

	return l
}

// Unlock unlocks current state.
func (l ValidatorLockedRewardsState) Unlock() ValidatorLockedRewardsState {
	l.LockedRatio = sdk.ZeroDec()
	l.LockedAt = time.Time{}
	l.LockHeight = 0
	l.UnlocksAt = time.Time{}
	l.AutoRenewal = false

	return l
}

// DisableAutoRenewal drop the auto-renewal flag.
func (l ValidatorLockedRewardsState) DisableAutoRenewal() ValidatorLockedRewardsState {
	l.AutoRenewal = false

	return l
}

// IsLocked checks if locking is active.
func (l ValidatorLockedRewardsState) IsLocked() bool {
	return !l.LockedRatio.IsZero()
}

// NewValidatorLockedRewards creates a new ValidatorLockedRewardsState.
func NewValidatorLockedRewards() ValidatorLockedRewardsState {
	l := ValidatorLockedRewardsState{}
	l = l.Unlock()

	return l
}
