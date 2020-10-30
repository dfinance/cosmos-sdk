package v0_39_1_0

import v03902 "github.com/cosmos/cosmos-sdk/x/distribution/legacy/v0_39-0_2"

// Migrate accepts exported genesis state from
// Dfinance v0.2 based on Cosmos SDK v0.39.1 and migrates it to
// Dfinance v1.0 based on Cosmos SDK v0.39.1 genesis state.
func Migrate(oldState v03902.GenesisState) GenesisState {
	return GenesisState{
		Params:                          oldState.Params,
		RewardPools:                     oldState.RewardPools,
		DelegatorWithdrawInfos:          oldState.DelegatorWithdrawInfos,
		PreviousProposer:                oldState.PreviousProposer,
		OutstandingRewards:              oldState.OutstandingRewards,
		ValidatorAccumulatedCommissions: oldState.ValidatorAccumulatedCommissions,
		ValidatorHistoricalRewards:      oldState.ValidatorHistoricalRewards,
		ValidatorCurrentRewards:         oldState.ValidatorCurrentRewards,
		DelegatorStartingInfos:          oldState.DelegatorStartingInfos,
		ValidatorSlashEvents:            oldState.ValidatorSlashEvents,
		ValidatorLockedRewards:          oldState.ValidatorLockedRewards,
		RewardBankPool:                  migrateRewardBankPool(oldState.RewardBankPool),
		RewardsUnlockQueue:              oldState.RewardsUnlockQueue,
	}
}

// migrateRewardBankPool migrates old RewardsBankPoolRecord entries setting validatorAddr to a non-existing one.
func migrateRewardBankPool(oldPoolRecords []v03902.RewardsBankPoolRecord) []RewardsBankPoolRecord {
	newPoolRecords := make([]RewardsBankPoolRecord, 0, len(oldPoolRecords))
	for _, poolRecord := range oldPoolRecords {
		newPoolRecords = append(newPoolRecords, RewardsBankPoolRecord{
			DelAddress: poolRecord.AccAddress,
			ValAddress: CommonRewardBankValAddr,
			Coins:      poolRecord.Coins,
		})
	}

	return newPoolRecords
}
