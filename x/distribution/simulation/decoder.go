package simulation

import (
	"bytes"
	"fmt"
	"time"

	tmkv "github.com/tendermint/tendermint/libs/kv"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/distribution/types"
)

// DecodeStore unmarshals the KVPair's Value to the corresponding distribution type
func DecodeStore(cdc *codec.Codec, kvA, kvB tmkv.Pair) string {
	switch {
	case bytes.Equal(kvA.Key[:1], types.RewardPoolsKey):
		var rewardPoolsA, rewardPoolsB types.RewardPools
		cdc.MustUnmarshalBinaryLengthPrefixed(kvA.Value, &rewardPoolsA)
		cdc.MustUnmarshalBinaryLengthPrefixed(kvB.Value, &rewardPoolsB)
		return fmt.Sprintf("%v\n%v", rewardPoolsA, rewardPoolsB)

	case bytes.Equal(kvA.Key[:1], types.ProposerKey):
		return fmt.Sprintf("%v\n%v", sdk.ConsAddress(kvA.Value), sdk.ConsAddress(kvB.Value))

	case bytes.Equal(kvA.Key[:1], types.ValidatorOutstandingRewardsPrefix):
		var rewardsA, rewardsB types.ValidatorOutstandingRewards
		cdc.MustUnmarshalBinaryLengthPrefixed(kvA.Value, &rewardsA)
		cdc.MustUnmarshalBinaryLengthPrefixed(kvB.Value, &rewardsB)
		return fmt.Sprintf("%v\n%v", rewardsA, rewardsB)

	case bytes.Equal(kvA.Key[:1], types.DelegatorWithdrawAddrPrefix):
		return fmt.Sprintf("%v\n%v", sdk.AccAddress(kvA.Value), sdk.AccAddress(kvB.Value))

	case bytes.Equal(kvA.Key[:1], types.DelegatorStartingInfoPrefix):
		var infoA, infoB types.DelegatorStartingInfo
		cdc.MustUnmarshalBinaryLengthPrefixed(kvA.Value, &infoA)
		cdc.MustUnmarshalBinaryLengthPrefixed(kvB.Value, &infoB)
		return fmt.Sprintf("%v\n%v", infoA, infoB)

	case bytes.Equal(kvA.Key[:1], types.ValidatorHistoricalRewardsPrefix):
		var rewardsA, rewardsB types.ValidatorHistoricalRewards
		cdc.MustUnmarshalBinaryLengthPrefixed(kvA.Value, &rewardsA)
		cdc.MustUnmarshalBinaryLengthPrefixed(kvB.Value, &rewardsB)
		return fmt.Sprintf("%v\n%v", rewardsA, rewardsB)

	case bytes.Equal(kvA.Key[:1], types.ValidatorCurrentRewardsPrefix):
		var rewardsA, rewardsB types.ValidatorCurrentRewards
		cdc.MustUnmarshalBinaryLengthPrefixed(kvA.Value, &rewardsA)
		cdc.MustUnmarshalBinaryLengthPrefixed(kvB.Value, &rewardsB)
		return fmt.Sprintf("%v\n%v", rewardsA, rewardsB)

	case bytes.Equal(kvA.Key[:1], types.ValidatorAccumulatedCommissionPrefix):
		var commissionA, commissionB types.ValidatorAccumulatedCommission
		cdc.MustUnmarshalBinaryLengthPrefixed(kvA.Value, &commissionA)
		cdc.MustUnmarshalBinaryLengthPrefixed(kvB.Value, &commissionB)
		return fmt.Sprintf("%v\n%v", commissionA, commissionB)

	case bytes.Equal(kvA.Key[:1], types.ValidatorSlashEventPrefix):
		var eventA, eventB types.ValidatorSlashEvent
		cdc.MustUnmarshalBinaryLengthPrefixed(kvA.Value, &eventA)
		cdc.MustUnmarshalBinaryLengthPrefixed(kvB.Value, &eventB)
		return fmt.Sprintf("%v\n%v", eventA, eventB)

	case bytes.Equal(kvA.Key[:1], types.DelegatorRewardsBankCoinsPrefix):
		var coinsA, coinsB sdk.Coins
		var delAddrA, delAddrB sdk.AccAddress
		var valAddrA, valAddrB sdk.ValAddress
		cdc.MustUnmarshalBinaryLengthPrefixed(kvA.Value, &coinsA)
		cdc.MustUnmarshalBinaryLengthPrefixed(kvB.Value, &coinsB)
		delAddrA, valAddrA = types.GetDelegatorRewardsBankCoinsAddress(kvA.Key)
		delAddrB, valAddrB = types.GetDelegatorRewardsBankCoinsAddress(kvB.Key)
		return fmt.Sprintf("%s-%s: %v\n%s-%s: %v", delAddrA, valAddrA, coinsA, delAddrB, valAddrB, coinsB)

	case bytes.Equal(kvA.Key[:1], types.RewardsUnlockQueueKey):
		var tsA, tsB time.Time
		var valAddrsA, valAddrsB []sdk.ValAddress
		tsA = types.ParseRewardsUnlockQueueTimeKey(kvA.Key)
		tsB = types.ParseRewardsUnlockQueueTimeKey(kvB.Key)
		cdc.MustUnmarshalBinaryLengthPrefixed(kvA.Value, &valAddrsA)
		cdc.MustUnmarshalBinaryLengthPrefixed(kvB.Value, &valAddrsB)
		return fmt.Sprintf("%v: %v\n%v: %v", tsA, valAddrsA, tsB, valAddrsB)

	default:
		panic(fmt.Sprintf("invalid distribution key prefix %X", kvA.Key[:1]))
	}
}
