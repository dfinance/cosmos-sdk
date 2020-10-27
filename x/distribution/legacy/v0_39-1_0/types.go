package v0_39_1_0

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	v0_39_0_2 "github.com/cosmos/cosmos-sdk/x/distribution/legacy/v0_39-0_2"
)

// GenesisState for Dfinance v1.0 Mainnet based on Cosmos SDK v0.39.1.
// Changelist:
//   - reward_bank_pool: validator info added;
type (
	GenesisState struct {
		Params                          v0_39_0_2.Params                                 `json:"params"`
		RewardPools                     v0_39_0_2.RewardPools                            `json:"reward_pools"`
		DelegatorWithdrawInfos          []v0_39_0_2.DelegatorWithdrawInfo                `json:"delegator_withdraw_infos"`
		PreviousProposer                sdk.ConsAddress                                  `json:"previous_proposer"`
		OutstandingRewards              []v0_39_0_2.ValidatorOutstandingRewardsRecord    `json:"outstanding_rewards"`
		ValidatorAccumulatedCommissions []v0_39_0_2.ValidatorAccumulatedCommissionRecord `json:"validator_accumulated_commissions"`
		ValidatorHistoricalRewards      []v0_39_0_2.ValidatorHistoricalRewardsRecord     `json:"validator_historical_rewards"`
		ValidatorCurrentRewards         []v0_39_0_2.ValidatorCurrentRewardsRecord        `json:"validator_current_rewards"`
		DelegatorStartingInfos          []v0_39_0_2.DelegatorStartingInfoRecord          `json:"delegator_starting_infos"`
		ValidatorSlashEvents            []v0_39_0_2.ValidatorSlashEventRecord            `json:"validator_slash_events"`
		ValidatorLockedRewards          []v0_39_0_2.ValidatorLockedRewardsRecord         `json:"validator_locked_rewards"`
		RewardBankPool                  []RewardsBankPoolRecord                          `json:"reward_bank_pool"`
		RewardsUnlockQueue              []v0_39_0_2.RewardsUnlockQueueRecord             `json:"rewards_unlock_queue"`
	}

	RewardsBankPoolRecord struct {
		DelAddress sdk.AccAddress `json:"del_address"`
		ValAddress sdk.ValAddress `json:"val_address"`
		Coins      sdk.Coins      `json:"coins"`
	}
)

var (
	// CommonRewardBankValAddr is used as a default non-existing validator for accumulated bank rewards.
	CommonRewardBankValAddr sdk.ValAddress
)

func init() {
	zeroAddrBytes := make([]byte, sdk.AddrLen, sdk.AddrLen)
	CommonRewardBankValAddr = sdk.ValAddress(zeroAddrBytes)
}
