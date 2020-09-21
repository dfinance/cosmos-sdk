package client

import (
	"github.com/cosmos/cosmos-sdk/x/distribution/client/cli"
	"github.com/cosmos/cosmos-sdk/x/distribution/client/rest"
	govclient "github.com/cosmos/cosmos-sdk/x/gov/client"
)

// param change proposal handlers
var (
	PublicTreasurySpendProposalHandler = govclient.NewProposalHandler(
		cli.GetCmdSubmitPublicTreasurySpendProposal,
		rest.PublicTreasurySpendProposalRESTHandler,
	)

	TaxParamsUpdateProposalHandler = govclient.NewProposalHandler(
		cli.GetCmdSubmitTaxParamsUpdateProposal,
		rest.TaxParamsUpdateProposalRESTHandler,
	)
)
