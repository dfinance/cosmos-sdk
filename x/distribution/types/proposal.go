package types

import (
	"fmt"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
)

const (
	ProposalTypePublicTreasuryPoolSpend = "PublicTreasuryPoolSpend"
	ProposalTypeTaxParamsUpdate         = "TaxParamsUpdate"
)

// Assert PublicTreasuryPoolSpendProposal, TaxParamsUpdateProposal implements govtypes.Content at compile-time
var _ govtypes.Content = PublicTreasuryPoolSpendProposal{}
var _ govtypes.Content = TaxParamsUpdateProposal{}

func init() {
	govtypes.RegisterProposalType(ProposalTypePublicTreasuryPoolSpend)
	govtypes.RegisterProposalType(ProposalTypeTaxParamsUpdate)
	govtypes.RegisterProposalTypeCodec(PublicTreasuryPoolSpendProposal{}, "cosmos-sdk/PublicTreasuryPoolSpendProposal")
	govtypes.RegisterProposalTypeCodec(TaxParamsUpdateProposal{}, "cosmos-sdk/TaxParamsUpdateProposal")
}

// PublicTreasuryPoolSpendProposal spends from the PublicTreasury pool to any account.
type PublicTreasuryPoolSpendProposal struct {
	Title       string         `json:"title" yaml:"title"`
	Description string         `json:"description" yaml:"description"`
	Recipient   sdk.AccAddress `json:"recipient" yaml:"recipient"`
	Amount      sdk.Coins      `json:"amount" yaml:"amount"`
}

// NewPublicTreasuryPoolSpendProposal creates a new PublicTreasury pool spend proposal.
func NewPublicTreasuryPoolSpendProposal(title, description string, recipient sdk.AccAddress, amount sdk.Coins) PublicTreasuryPoolSpendProposal {
	return PublicTreasuryPoolSpendProposal{title, description, recipient, amount}
}

// GetTitle returns the title of a PublicTreasury pool spend proposal.
func (csp PublicTreasuryPoolSpendProposal) GetTitle() string { return csp.Title }

// GetDescription returns the description of a PublicTreasury pool spend proposal.
func (csp PublicTreasuryPoolSpendProposal) GetDescription() string { return csp.Description }

// GetDescription returns the routing key of a PublicTreasury pool spend proposal.
func (csp PublicTreasuryPoolSpendProposal) ProposalRoute() string { return RouterKey }

// ProposalType returns the type of a PublicTreasury pool spend proposal.
func (csp PublicTreasuryPoolSpendProposal) ProposalType() string {
	return ProposalTypePublicTreasuryPoolSpend
}

// ValidateBasic runs basic stateless validity checks
func (csp PublicTreasuryPoolSpendProposal) ValidateBasic() error {
	if err := govtypes.ValidateAbstract(csp); err != nil {
		return err
	}

	if !csp.Amount.IsValid() {
		return sdkerrors.Wrap(ErrInvalidProposalAmount, "public treasury pool spend proposal")
	}
	if csp.Recipient.Empty() {
		return sdkerrors.Wrap(ErrEmptyProposalRecipient, "public treasury pool spend proposal")
	}

	return nil
}

// String implements the Stringer interface.
func (csp PublicTreasuryPoolSpendProposal) String() string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf(`PublicTreasuryPool Spend Proposal:
  Title:       %s
  Description: %s
  Recipient:   %s
  Amount:      %s
`, csp.Title, csp.Description, csp.Recipient, csp.Amount))
	return b.String()
}

// TaxParamsUpdateProposal updates Tax params wholly.
type TaxParamsUpdateProposal struct {
	Title                     string  `json:"title" yaml:"title"`
	Description               string  `json:"description" yaml:"description"`
	ValidatorsPoolTax         sdk.Dec `json:"validators_pool_tax" yaml:"validators_pool_tax"`
	LiquidityProvidersPoolTax sdk.Dec `json:"liquidity_providers_pool_tax" yaml:"liquidity_providers_pool_tax"`
	PublicTreasuryPoolTax     sdk.Dec `json:"public_treasury_pool_tax" yaml:"public_treasury_pool_tax"`
	HARPTax                   sdk.Dec `json:"harp_tax" yaml:"harp_tax"`
}

// NewTaxParamsUpdateProposal creates a new Tax params update proposal.
func NewTaxParamsUpdateProposal(
	title, description string,
	validatorsPoolTax, liquidityProvidersPoolTax, publicTreasuryPoolTax, harpTax sdk.Dec,
) TaxParamsUpdateProposal {
	return TaxParamsUpdateProposal{
		Title:                     title,
		Description:               description,
		ValidatorsPoolTax:         validatorsPoolTax,
		LiquidityProvidersPoolTax: liquidityProvidersPoolTax,
		PublicTreasuryPoolTax:     publicTreasuryPoolTax,
		HARPTax:                   harpTax,
	}
}

// GetTitle returns the title of a Tax params update proposal.
func (tup TaxParamsUpdateProposal) GetTitle() string { return tup.Title }

// GetDescription returns the description of a Tax params update proposal.
func (tup TaxParamsUpdateProposal) GetDescription() string { return tup.Description }

// GetDescription returns the routing key of a Tax params update proposal.
func (tup TaxParamsUpdateProposal) ProposalRoute() string { return RouterKey }

// ProposalType returns the type of a Tax params update proposal.
func (tup TaxParamsUpdateProposal) ProposalType() string { return ProposalTypeTaxParamsUpdate }

// ValidateBasic runs basic stateless validity checks.
func (tup TaxParamsUpdateProposal) ValidateBasic() error {
	if err := govtypes.ValidateAbstract(tup); err != nil {
		return err
	}

	if err := validateValidatorsPoolTax(tup.ValidatorsPoolTax); err != nil {
		return err
	}
	if err := validateLiquidityProvidersPoolTax(tup.LiquidityProvidersPoolTax); err != nil {
		return err
	}
	if err := validatePublicTreasuryPoolTax(tup.PublicTreasuryPoolTax); err != nil {
		return err
	}
	if err := validateParamKeyHARPTax(tup.HARPTax); err != nil {
		return err
	}

	if v := tup.ValidatorsPoolTax.Add(tup.LiquidityProvidersPoolTax).Add(tup.PublicTreasuryPoolTax).Add(tup.HARPTax); !v.Equal(sdk.OneDec()) {
		return fmt.Errorf("sum of all pool taxes must be 1.0: %s", v)
	}

	return nil
}

// String implements the Stringer interface.
func (tup TaxParamsUpdateProposal) String() string {
	return fmt.Sprintf(`TaxParams update Proposal:
  Title:                     %s
  Description:               %s
  ValidatorsPoolTax:         %s
  LiquidityProvidersPoolTax: %s
  PublicTreasuryPoolTax:     %s
  HARPTax:                   %s
`,
		tup.Title, tup.Description,
		tup.ValidatorsPoolTax, tup.LiquidityProvidersPoolTax, tup.PublicTreasuryPoolTax, tup.HARPTax,
	)
}
