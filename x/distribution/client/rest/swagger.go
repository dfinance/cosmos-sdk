package rest

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

//nolint:deadcode,unused
type (
	SwaggerValidatorDistInfo struct {
		OperatorAddress     sdk.AccAddress `json:"operator_address" yaml:"operator_address" swaggertype:"string" format:"bech32" example:"wallet13jyjuz3kkdvqw8u4qfkwd94emdl3vx394kn07h"`
		ValidatorCommission sdk.DecCoins   `json:"validator_commission"`
		SelfBondRewards     sdk.DecCoins   `json:"self_bond_rewards"`
	}
)
