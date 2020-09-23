package simulation

import (
	"bytes"
	"fmt"
	"time"

	tmkv "github.com/tendermint/tendermint/libs/kv"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/mint/internal/types"
)

// DecodeStore unmarshals the KVPair's Value to the corresponding mint type
func DecodeStore(cdc *codec.Codec, kvA, kvB tmkv.Pair) string {
	switch {
	case bytes.Equal(kvA.Key, types.MinterKey):
		var minterA, minterB types.Minter
		cdc.MustUnmarshalBinaryLengthPrefixed(kvA.Value, &minterA)
		cdc.MustUnmarshalBinaryLengthPrefixed(kvB.Value, &minterB)
		return fmt.Sprintf("%v\n%v", minterA, minterB)
	case bytes.Equal(kvA.Key, types.BlockDurFilterKey):
		var filterA, filterB types.BlockDurFilter
		if len(kvA.Value) != 0 {
			cdc.MustUnmarshalBinaryLengthPrefixed(kvA.Value, &filterA)
		}
		if len(kvB.Value) != 0 {
			cdc.MustUnmarshalBinaryLengthPrefixed(kvB.Value, &filterB)
		}
		return fmt.Sprintf("%v\n%v", filterA, filterB)
	case bytes.Equal(kvA.Key, types.AnnualUpdateTimestampKey):
		var timeA, timeB time.Time
		var err error
		if len(kvA.Value) != 0 {
			timeA, err = sdk.ParseTimeBytes(kvA.Value)
			if err != nil {
				panic(err)
			}
		}
		if len(kvB.Value) != 0 {
			timeB, err = sdk.ParseTimeBytes(kvB.Value)
			if err != nil {
				panic(err)
			}
		}
		return fmt.Sprintf("%v\n%v", timeA, timeB)
	default:
		panic(fmt.Sprintf("invalid mint key %X", kvA.Key))
	}
}
