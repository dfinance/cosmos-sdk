package rest

import (
	auth "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/staking/types"
)

//nolint:deadcode,unused
type (
	TxUndelegate struct {
		Msgs       []types.MsgUndelegate `json:"msg" yaml:"msg"`
		Fee        auth.StdFee           `json:"fee" yaml:"fee"`
		Signatures []auth.StdSignature   `json:"signatures" yaml:"signatures"`
		Memo       string                `json:"memo" yaml:"memo"`
	}

	TxBeginRedelegate struct {
		Msgs       []types.MsgBeginRedelegate `json:"msg" yaml:"msg"`
		Fee        auth.StdFee                `json:"fee" yaml:"fee"`
		Signatures []auth.StdSignature        `json:"signatures" yaml:"signatures"`
		Memo       string                     `json:"memo" yaml:"memo"`
	}

	TxDelegate struct {
		Msgs       []types.MsgDelegate `json:"msg" yaml:"msg"`
		Fee        auth.StdFee         `json:"fee" yaml:"fee"`
		Signatures []auth.StdSignature `json:"signatures" yaml:"signatures"`
		Memo       string              `json:"memo" yaml:"memo"`
	}
)
