package rest

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	auth "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/distribution/types"
)

//nolint:deadcode,unused
type (
	SwaggerValidatorDistInfo struct {
		OperatorAddress     sdk.AccAddress `json:"operator_address" yaml:"operator_address" swaggertype:"string" format:"bech32" example:"wallet13jyjuz3kkdvqw8u4qfkwd94emdl3vx394kn07h"`
		ValidatorCommission sdk.DecCoins   `json:"validator_commission"`
		SelfBondRewards     sdk.DecCoins   `json:"self_bond_rewards"`
	}

	TxWithdrawDelegatorReward struct {
		Msgs       []types.MsgWithdrawDelegatorReward `json:"msg" yaml:"msg"`
		Fee        auth.StdFee                        `json:"fee" yaml:"fee"`
		Signatures []auth.StdSignature                `json:"signatures" yaml:"signatures"`
		Memo       string                             `json:"memo" yaml:"memo"`
	}

	TxSetWithdrawAddress struct {
		Msgs       []types.MsgSetWithdrawAddress `json:"msg" yaml:"msg"`
		Fee        auth.StdFee                   `json:"fee" yaml:"fee"`
		Signatures []auth.StdSignature           `json:"signatures" yaml:"signatures"`
		Memo       string                        `json:"memo" yaml:"memo"`
	}

	TxFundPublicTreasuryPool struct {
		Msgs       []types.MsgFundPublicTreasuryPool `json:"msg" yaml:"msg"`
		Fee        auth.StdFee                       `json:"fee" yaml:"fee"`
		Signatures []auth.StdSignature               `json:"signatures" yaml:"signatures"`
		Memo       string                            `json:"memo" yaml:"memo"`
	}
)
