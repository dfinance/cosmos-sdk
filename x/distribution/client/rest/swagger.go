package rest

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	auth "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/distribution/types"
)

//nolint:deadcode,unused
type (
	TxWithdrawDelegatorReward struct {
		Msgs       []types.MsgWithdrawDelegatorReward `json:"msg"`
		Fee        auth.StdFee                        `json:"fee"`
		Signatures []auth.StdSignature                `json:"signatures"`
		Memo       string                             `json:"memo"`
	}

	TxSetWithdrawAddress struct {
		Msgs       []types.MsgSetWithdrawAddress `json:"msg"`
		Fee        auth.StdFee                   `json:"fee"`
		Signatures []auth.StdSignature           `json:"signatures"`
		Memo       string                        `json:"memo"`
	}

	TxFundPublicTreasuryPool struct {
		Msgs       []types.MsgFundPublicTreasuryPool `json:"msg"`
		Fee        auth.StdFee                       `json:"fee"`
		Signatures []auth.StdSignature               `json:"signatures"`
		Memo       string                            `json:"memo"`
	}

	QuerySwaggerValidatorDistInfoResp struct {
		Height int64 `json:"height"`
		Result struct {
			OperatorAddress     sdk.AccAddress `json:"operator_address" yaml:"operator_address" swaggertype:"string" format:"bech32" example:"wallet13jyjuz3kkdvqw8u4qfkwd94emdl3vx394kn07h"`
			ValidatorCommission sdk.DecCoins   `json:"validator_commission"`
			SelfBondRewards     sdk.DecCoins   `json:"self_bond_rewards"`
		} `json:"result"`
	}

	QueryDelegatorRewardsResp struct {
		Height int64                                    `json:"height"`
		Result types.QueryDelegatorTotalRewardsResponse `json:"result"`
	}

	QueryDelegationRewardsResp struct {
		Height int64                                `json:"height"`
		Result types.QueryDelegationRewardsResponse `json:"result"`
	}

	QueryDecCoinsResp struct {
		Height int64         `json:"height"`
		Result []sdk.DecCoin `json:"result"`
	}

	QueryAddressResp struct {
		Height int64  `json:"height"`
		Result string `json:"result" format:"bech32" example:"wallet13jyjuz3kkdvqw8u4qfkwd94emdl3vx394kn07h"`
	}

	QueryParamsResp struct {
		Height int64        `json:"height"`
		Result types.Params `json:"result"`
	}

	QueryExtendedValidatorResp struct {
		Height int64               `json:"height"`
		Result types.ValidatorResp `json:"result"`
	}

	QueryExtendedValidatorsResp struct {
		Height int64                 `json:"height"`
		Result []types.ValidatorResp `json:"result"`
	}
)
