package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
)

// ModuleCdc is a generic sealed codec to be used throughout this module
var ModuleCdc *codec.Codec

// Register concrete types on codec codec
func RegisterCodec(cdc *codec.Codec) {
	cdc.RegisterConcrete(MsgSetFoundationAllocationRatio{}, "cosmos-sdk/MsgSetFoundationAllocationRatio", nil)
}

func init() {
	ModuleCdc = codec.New()
	codec.RegisterCrypto(ModuleCdc)
	RegisterCodec(ModuleCdc)
	ModuleCdc.Seal()
}
