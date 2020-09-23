package types

import sdk "github.com/cosmos/cosmos-sdk/types"

// RewardsBankPool is an intermediate module account used to store delegator rewards.
// BeforeDelegationSharesModified staking keeper hook force withdraws current delegator rewards.
// Instead we transfer them to the RewardsBankPool where they can be withdraw with delegator decision.
const RewardsBankPoolName = "rewards_bank_pool"

// RewardsBankDelegatorCoins used to store current delegator coins stored in the RewardsBankPool.
// Delegator address is encoded into the storage key.
type RewardsBankDelegatorCoins struct {
	Coins sdk.Coins `json:"coins"`
}
