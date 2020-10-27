package v0_39_0_2

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// GenesisState for Dfinance v0.2 Testnet based on Cosmos SDK v0.39.1.
// This is a starting point for all future Dfinance migrations.
type (
	GenesisState struct {
		Params                          Params                                 `json:"params"`
		RewardPools                     RewardPools                            `json:"reward_pools"`
		DelegatorWithdrawInfos          []DelegatorWithdrawInfo                `json:"delegator_withdraw_infos"`
		PreviousProposer                sdk.ConsAddress                        `json:"previous_proposer"`
		OutstandingRewards              []ValidatorOutstandingRewardsRecord    `json:"outstanding_rewards"`
		ValidatorAccumulatedCommissions []ValidatorAccumulatedCommissionRecord `json:"validator_accumulated_commissions"`
		ValidatorHistoricalRewards      []ValidatorHistoricalRewardsRecord     `json:"validator_historical_rewards"`
		ValidatorCurrentRewards         []ValidatorCurrentRewardsRecord        `json:"validator_current_rewards"`
		DelegatorStartingInfos          []DelegatorStartingInfoRecord          `json:"delegator_starting_infos"`
		ValidatorSlashEvents            []ValidatorSlashEventRecord            `json:"validator_slash_events"`
		ValidatorLockedRewards          []ValidatorLockedRewardsRecord         `json:"validator_locked_rewards"`
		RewardBankPool                  []RewardsBankPoolRecord                `json:"reward_bank_pool"`
		RewardsUnlockQueue              []RewardsUnlockQueueRecord             `json:"rewards_unlock_queue"`
	}

	Params struct {
		ValidatorsPoolTax          sdk.Dec          `json:"validators_pool_tax"`
		LiquidityProvidersPoolTax  sdk.Dec          `json:"liquidity_providers_pool_tax"`
		PublicTreasuryPoolTax      sdk.Dec          `json:"public_treasury_pool_tax"`
		HARPTax                    sdk.Dec          `json:"harp_tax"`
		PublicTreasuryPoolCapacity sdk.Int          `json:"public_treasury_pool_capacity"`
		BaseProposerReward         sdk.Dec          `json:"base_proposer_reward"`
		BonusProposerReward        sdk.Dec          `json:"bonus_proposer_reward"`
		LockedRatio                sdk.Dec          `json:"locked_ratio"`
		LockedDuration             time.Duration    `json:"locked_dur"`
		WithdrawAddrEnabled        bool             `json:"withdraw_addr_enabled"`
		FoundationNominees         []sdk.AccAddress `json:"foundation_nominees"`
	}

	RewardPools struct {
		LiquidityProvidersPool sdk.DecCoins `json:"liquidity_providers_pool"`
		FoundationPool         sdk.DecCoins `json:"foundation_pool"`
		PublicTreasuryPool     sdk.DecCoins `json:"treasury_pool"`
		HARP                   sdk.DecCoins `json:"harp_pool"`
	}

	DelegatorWithdrawInfo struct {
		DelegatorAddress sdk.AccAddress `json:"delegator_address"`
		WithdrawAddress  sdk.AccAddress `json:"withdraw_address"`
	}

	ValidatorOutstandingRewardsRecord struct {
		ValidatorAddress   sdk.ValAddress `json:"validator_address"`
		OutstandingRewards sdk.DecCoins   `json:"outstanding_rewards"`
	}

	ValidatorAccumulatedCommissionRecord struct {
		ValidatorAddress sdk.ValAddress                 `json:"validator_address"`
		Accumulated      ValidatorAccumulatedCommission `json:"accumulated"`
	}

	ValidatorAccumulatedCommission = sdk.DecCoins

	ValidatorHistoricalRewardsRecord struct {
		ValidatorAddress sdk.ValAddress             `json:"validator_address"`
		Period           uint64                     `json:"period"`
		Rewards          ValidatorHistoricalRewards `json:"rewards"`
	}

	ValidatorHistoricalRewards struct {
		CumulativeBondingRewardRatio sdk.DecCoins `json:"cumulative_bonding_reward_ratio"`
		CumulativeLPRewardRatio      sdk.DecCoins `json:"cumulative_lp_reward_ratio"`
		ReferenceCount               uint16       `json:"reference_count"`
	}

	ValidatorCurrentRewardsRecord struct {
		ValidatorAddress sdk.ValAddress          `json:"validator_address" yaml:"validator_address"`
		Rewards          ValidatorCurrentRewards `json:"rewards" yaml:"rewards"`
	}

	ValidatorCurrentRewards struct {
		BondingRewards sdk.DecCoins `json:"bonding_rewards"`
		LPRewards      sdk.DecCoins `json:"lp_rewards"`
		Period         uint64       `json:"period"`
	}

	DelegatorStartingInfoRecord struct {
		DelegatorAddress sdk.AccAddress        `json:"delegator_address"`
		ValidatorAddress sdk.ValAddress        `json:"validator_address"`
		StartingInfo     DelegatorStartingInfo `json:"starting_info"`
	}

	DelegatorStartingInfo struct {
		PreviousPeriod uint64  `json:"previous_period"`
		BondingStake   sdk.Dec `json:"bonding_stake"`
		LPStake        sdk.Dec `json:"lp_stake"`
		Height         uint64  `json:"creation_height"`
	}

	ValidatorSlashEventRecord struct {
		ValidatorAddress sdk.ValAddress      `json:"validator_address"`
		Height           uint64              `json:"height"`
		Period           uint64              `json:"period"`
		Event            ValidatorSlashEvent `json:"validator_slash_event"`
	}

	ValidatorSlashEvent struct {
		ValidatorPeriod uint64  `json:"validator_period"`
		Fraction        sdk.Dec `json:"fraction"`
	}

	ValidatorLockedRewardsRecord struct {
		ValidatorAddress sdk.ValAddress              `json:"validator_address"`
		LockedInfo       ValidatorLockedRewardsState `json:"locked_info"`
	}

	ValidatorLockedRewardsState struct {
		LockHeight  int64     `json:"lock_height"`
		LockedAt    time.Time `json:"locked_at"`
		UnlocksAt   time.Time `json:"unlocks_at"`
		LockedRatio sdk.Dec   `json:"locked_ratio"`
		AutoRenewal bool      `json:"auto_renewal"`
	}

	RewardsBankPoolRecord struct {
		AccAddress sdk.AccAddress `json:"acc_address"`
		Coins      sdk.Coins      `json:"coins"`
	}

	RewardsUnlockQueueRecord struct {
		Timestamp          time.Time        `json:" timestamp"`
		ValidatorAddresses []sdk.ValAddress `json:"validator_addresses"`
	}
)
