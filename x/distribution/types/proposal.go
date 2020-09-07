package types

import (
	"fmt"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
)

const (
	// ProposalTypePublicTreasuryPoolSpend defines the type for a PublicTreasuryPoolSpendProposal
	ProposalTypePublicTreasuryPoolSpend = "PublicTreasuryPoolSpend"
)

// Assert PublicTreasuryPoolSpendProposal implements govtypes.Content at compile-time
var _ govtypes.Content = PublicTreasuryPoolSpendProposal{}

func init() {
	govtypes.RegisterProposalType(ProposalTypePublicTreasuryPoolSpend)
	govtypes.RegisterProposalTypeCodec(PublicTreasuryPoolSpendProposal{}, "cosmos-sdk/PublicTreasuryPoolSpendProposal")
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
	err := govtypes.ValidateAbstract(csp)
	if err != nil {
		return err
	}
	if !csp.Amount.IsValid() {
		return ErrInvalidProposalAmount
	}
	if csp.Recipient.Empty() {
		return ErrEmptyProposalRecipient
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
