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

	QueryParamsResp struct {
		Height int64        `json:"height"`
		Result types.Params `json:"result"`
	}

	QueryPoolResp struct {
		Height int64      `json:"height"`
		Result types.Pool `json:"result"`
	}

	QueryHistoricalInfoResp struct {
		Height int64                `json:"height"`
		Result types.HistoricalInfo `json:"result"`
	}

	QueryUnbondingDelegationsResp struct {
		Height int64                       `json:"height"`
		Result []types.UnbondingDelegation `json:"result"`
	}

	QueryUnbondingDelegationResp struct {
		Height int64                     `json:"height"`
		Result types.UnbondingDelegation `json:"result"`
	}

	QueryDelegationsResp struct {
		Height int64                      `json:"height"`
		Result []types.DelegationResponse `json:"result"`
	}

	QueryDelegationResp struct {
		Height int64            `json:"height"`
		Result types.Delegation `json:"result"`
	}

	QueryValidatorsResp struct {
		Height int64             `json:"height"`
		Result []types.Validator `json:"result"`
	}

	QueryValidatorResp struct {
		Height int64           `json:"height"`
		Result types.Validator `json:"result"`
	}

	QueryRedelegationsResp struct {
		Height int64                        `json:"height"`
		Result []types.RedelegationResponse `json:"result"`
	}
)
