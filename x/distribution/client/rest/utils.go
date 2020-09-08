package rest

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/rest"
)

type (
	// PublicTreasuryPoolSpendProposalReq defines a public treasury pool spend proposal request body.
	PublicTreasuryPoolSpendProposalReq struct {
		BaseReq rest.BaseReq `json:"base_req" yaml:"base_req"`

		Title       string         `json:"title" yaml:"title"`
		Description string         `json:"description" yaml:"description"`
		Recipient   sdk.AccAddress `json:"recipient" yaml:"recipient"`
		Amount      sdk.Coins      `json:"amount" yaml:"amount"`
		Proposer    sdk.AccAddress `json:"proposer" yaml:"proposer"`
		Deposit     sdk.Coins      `json:"deposit" yaml:"deposit"`
	}

	// TaxParamsUpdateProposalReq defines a tax params update proposal request body.
	TaxParamsUpdateProposalReq struct {
		BaseReq rest.BaseReq `json:"base_req" yaml:"base_req"`

		Title                     string         `json:"title" yaml:"title"`
		Description               string         `json:"description" yaml:"description"`
		ValidatorsPoolTax         sdk.Dec        `json:"validators_pool_tax" yaml:"validators_pool_tax"`
		LiquidityProvidersPoolTax sdk.Dec        `json:"liquidity_providers_pool_tax" yaml:"liquidity_providers_pool_tax"`
		PublicTreasuryPoolTax     sdk.Dec        `json:"public_treasury_pool_tax" yaml:"public_treasury_pool_tax"`
		HARPTax                   sdk.Dec        `json:"harp_tax" yaml:"harp_tax"`
		Proposer                  sdk.AccAddress `json:"proposer" yaml:"proposer"`
		Deposit                   sdk.Coins      `json:"deposit" yaml:"deposit"`
	}
)
