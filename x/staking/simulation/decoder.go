package simulation

import (
	"bytes"
	"fmt"
	"time"

	tmkv "github.com/tendermint/tendermint/libs/kv"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/staking/types"
)

// DecodeStore unmarshals the KVPair's Value to the corresponding staking type
func DecodeStore(cdc *codec.Codec, kvA, kvB tmkv.Pair) string {
	switch {
	case bytes.Equal(kvA.Key[:1], types.LastTotalPowerKey):
		var powerA, powerB sdk.Int
		cdc.MustUnmarshalBinaryLengthPrefixed(kvA.Value, &powerA)
		cdc.MustUnmarshalBinaryLengthPrefixed(kvB.Value, &powerB)
		return fmt.Sprintf("%v\n%v", powerA, powerB)

	case bytes.Equal(kvA.Key[:1], types.ValidatorsKey):
		var validatorA, validatorB types.Validator
		cdc.MustUnmarshalBinaryLengthPrefixed(kvA.Value, &validatorA)
		cdc.MustUnmarshalBinaryLengthPrefixed(kvB.Value, &validatorB)
		return fmt.Sprintf("%v\n%v", validatorA, validatorB)

	case bytes.Equal(kvA.Key[:1], types.LastValidatorPowerKey),
		bytes.Equal(kvA.Key[:1], types.ValidatorsByConsAddrKey),
		bytes.Equal(kvA.Key[:1], types.ValidatorsByPowerIndexKey):
		return fmt.Sprintf("%v\n%v", sdk.ValAddress(kvA.Value), sdk.ValAddress(kvB.Value))

	case bytes.Equal(kvA.Key[:1], types.DelegationKey):
		var delegationA, delegationB types.Delegation
		cdc.MustUnmarshalBinaryLengthPrefixed(kvA.Value, &delegationA)
		cdc.MustUnmarshalBinaryLengthPrefixed(kvB.Value, &delegationB)
		return fmt.Sprintf("%v\n%v", delegationA, delegationB)

	case bytes.Equal(kvA.Key[:1], types.UnbondingDelegationKey),
		bytes.Equal(kvA.Key[:1], types.UnbondingDelegationByValIndexKey):
		var ubdA, ubdB types.UnbondingDelegation
		cdc.MustUnmarshalBinaryLengthPrefixed(kvA.Value, &ubdA)
		cdc.MustUnmarshalBinaryLengthPrefixed(kvB.Value, &ubdB)
		return fmt.Sprintf("%v\n%v", ubdA, ubdB)

	case bytes.Equal(kvA.Key[:1], types.RedelegationKey),
		bytes.Equal(kvA.Key[:1], types.RedelegationByValSrcIndexKey):
		var redA, redB types.Redelegation
		cdc.MustUnmarshalBinaryLengthPrefixed(kvA.Value, &redA)
		cdc.MustUnmarshalBinaryLengthPrefixed(kvB.Value, &redB)
		return fmt.Sprintf("%v\n%v", redA, redB)

	case bytes.Equal(kvA.Key[:1], types.ValidatorsStakingStateKey):
		var stateA, stateB types.ValidatorStakingState
		var valAddrA, valAddrB sdk.ValAddress
		cdc.MustUnmarshalBinaryLengthPrefixed(kvA.Value, &stateA)
		cdc.MustUnmarshalBinaryLengthPrefixed(kvB.Value, &stateB)
		valAddrA = types.ParseValidatorStakingStateKey(kvA.Key)
		valAddrB = types.ParseValidatorStakingStateKey(kvB.Key)
		return fmt.Sprintf("%s: %v\n%s: %v", valAddrA, stateA, valAddrB, stateB)

	case bytes.Equal(kvA.Key[:1], types.ScheduledUnbondQueueKey):
		var valAddrsA, valAddrsB []sdk.ValAddress
		var timestampA, timestampB time.Time
		cdc.MustUnmarshalBinaryLengthPrefixed(kvA.Value, &valAddrsA)
		cdc.MustUnmarshalBinaryLengthPrefixed(kvB.Value, &valAddrsB)
		timestampA = types.ParseUnbondingDelegationTimeKey(kvA.Key)
		timestampB = types.ParseUnbondingDelegationTimeKey(kvB.Key)
		return fmt.Sprintf("%v: %v\n%v: %v", timestampA, valAddrsA, timestampB, valAddrsB)

	case bytes.Equal(kvA.Key[:1], types.BannedAccKey):
		var accAddrsA, accAddrsB sdk.AccAddress
		var heightA, heightB int64
		cdc.MustUnmarshalBinaryLengthPrefixed(kvA.Value, &heightA)
		cdc.MustUnmarshalBinaryLengthPrefixed(kvB.Value, &heightB)
		accAddrsA = types.ParseBannedAccKey(kvA.Key)
		accAddrsB = types.ParseBannedAccKey(kvB.Key)
		return fmt.Sprintf("%s: %d\n%s: %d", accAddrsA, heightA, accAddrsB, heightB)

	default:
		panic(fmt.Sprintf("invalid staking key prefix %X", kvA.Key[:1]))
	}
}
