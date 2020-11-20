package rest

import (
	"github.com/cosmos/cosmos-sdk/x/mint/internal/types"
)

//nolint:deadcode,unused
type (
	QueryDecResp struct {
		Height int64  `json:"height"`
		Result string `json:"result" format:"number" example:"0.123"`
	}

	QueryMinterParamsResp struct {
		Height int64        `json:"height"`
		Result types.Params `json:"result"`
	}

	QueryMinterExtendedResp struct {
		Height int64          `json:"height"`
		Result types.MintInfo `json:"result"`
	}
)
