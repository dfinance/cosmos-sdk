package cli

import (
	"io/ioutil"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type (
	// PublicTreasuryPoolSpendProposalJSON defines a PublicTreasuryPoolSpendProposal with a deposit.
	PublicTreasuryPoolSpendProposalJSON struct {
		Title       string         `json:"title" yaml:"title"`
		Description string         `json:"description" yaml:"description"`
		Recipient   sdk.AccAddress `json:"recipient" yaml:"recipient"`
		Amount      sdk.Coins      `json:"amount" yaml:"amount"`
		Deposit     sdk.Coins      `json:"deposit" yaml:"deposit"`
	}

	TaxParamsUpdateProposalJSON struct {
		Title                     string    `json:"title" yaml:"title"`
		Description               string    `json:"description" yaml:"description"`
		ValidatorsPoolTax         sdk.Dec   `json:"validators_pool_tax" yaml:"validators_pool_tax"`
		LiquidityProvidersPoolTax sdk.Dec   `json:"liquidity_providers_pool_tax" yaml:"liquidity_providers_pool_tax"`
		PublicTreasuryPoolTax     sdk.Dec   `json:"public_treasury_pool_tax" yaml:"public_treasury_pool_tax"`
		HARPTax                   sdk.Dec   `json:"harp_tax" yaml:"harp_tax"`
		Deposit                   sdk.Coins `json:"deposit" yaml:"deposit"`
	}
)

// ParsePublicTreasuryPoolSpendProposalJSON reads and parses a PublicTreasuryPoolSpendProposalJSON from a file.
func ParsePublicTreasuryPoolSpendProposalJSON(cdc *codec.Codec, proposalFile string) (PublicTreasuryPoolSpendProposalJSON, error) {
	proposal := PublicTreasuryPoolSpendProposalJSON{}

	contents, err := ioutil.ReadFile(proposalFile)
	if err != nil {
		return proposal, err
	}

	if err := cdc.UnmarshalJSON(contents, &proposal); err != nil {
		return proposal, err
	}

	return proposal, nil
}

// ParseTaxParamsUpdateProposalJSON reads and parses a TaxParamsUpdateProposalJSON from a file.
func ParseTaxParamsUpdateProposalJSON(cdc *codec.Codec, proposalFile string) (TaxParamsUpdateProposalJSON, error) {
	proposal := TaxParamsUpdateProposalJSON{}

	contents, err := ioutil.ReadFile(proposalFile)
	if err != nil {
		return proposal, err
	}

	if err := cdc.UnmarshalJSON(contents, &proposal); err != nil {
		return proposal, err
	}

	return proposal, nil
}
