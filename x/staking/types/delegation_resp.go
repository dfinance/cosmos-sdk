package types

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// DelegationResponse is equivalent to Delegation except that it contains a balance
// in addition to shares which is more suitable for client responses.
type DelegationResponse struct {
	Delegation
	BondingBalance sdk.Coin `json:"bonding_balance" yaml:"bonding_balance" swaggertype:"string"`
	LPBalance      sdk.Coin `json:"lp_balance" yaml:"lp_balance" swaggertype:"string"`
}

// NewDelegationResp creates a new DelegationResponse instance.
func NewDelegationResp(
	delegatorAddr sdk.AccAddress, validatorAddr sdk.ValAddress,
	bondingShares, lpShares sdk.Dec,
	bondingBalance, lpBalance sdk.Coin,
) DelegationResponse {

	return DelegationResponse{
		Delegation:     NewDelegation(delegatorAddr, validatorAddr, bondingShares, lpShares),
		BondingBalance: bondingBalance,
		LPBalance:      lpBalance,
	}
}

// String implements the Stringer interface for DelegationResponse.
func (d DelegationResponse) String() string {
	return fmt.Sprintf(`%s
  BondingBalance: %s
  LPBalance:      %s`,
		d.Delegation.String(), d.BondingBalance, d.LPBalance,
	)
}

type delegationRespAlias DelegationResponse

// MarshalJSON implements the json.Marshaler interface. This is so we can
// achieve a flattened structure while embedding other types.
func (d DelegationResponse) MarshalJSON() ([]byte, error) {
	return json.Marshal((delegationRespAlias)(d))
}

// UnmarshalJSON implements the json.Unmarshaler interface. This is so we can
// achieve a flattened structure while embedding other types.
func (d *DelegationResponse) UnmarshalJSON(bz []byte) error {
	return json.Unmarshal(bz, (*delegationRespAlias)(d))
}

// DelegationResponses is a collection of DelegationResp.
type DelegationResponses []DelegationResponse

// String implements the Stringer interface for DelegationResponses.
func (d DelegationResponses) String() (out string) {
	for _, del := range d {
		out += del.String() + "\n"
	}
	return strings.TrimSpace(out)
}

// RedelegationResponse is equivalent to a Redelegation except that its entries
// contain a balance in addition to shares which is more suitable for client
// responses.
type RedelegationResponse struct {
	Redelegation
	Entries []RedelegationEntryResponse `json:"entries" yaml:"entries"`
}

// NewRedelegationResponse crates a new RedelegationEntryResponse instance.
func NewRedelegationResponse(
	delegatorAddr sdk.AccAddress, validatorSrc, validatorDst sdk.ValAddress, entries []RedelegationEntryResponse,
) RedelegationResponse {

	return RedelegationResponse{
		Redelegation: Redelegation{
			DelegatorAddress:    delegatorAddr,
			ValidatorSrcAddress: validatorSrc,
			ValidatorDstAddress: validatorDst,
		},
		Entries: entries,
	}
}

// RedelegationEntryResponse is equivalent to a RedelegationEntry except that it
// contains a balance in addition to shares which is more suitable for client
// responses.
type RedelegationEntryResponse struct {
	RedelegationEntry
	Balance sdk.Int `json:"balance" yaml:"balance" swaggertype:"string" format:"integer"`
}

// NewRedelegationEntryResponse creates a new RedelegationEntryResponse instance.
func NewRedelegationEntryResponse(
	creationHeight int64, completionTime time.Time,
	opType DelegationOpType, sharesDst sdk.Dec, initialBalance, balance sdk.Int,
) RedelegationEntryResponse {

	return RedelegationEntryResponse{
		RedelegationEntry: NewRedelegationEntry(creationHeight, completionTime, opType, initialBalance, sharesDst),
		Balance:           balance,
	}
}

// String implements the Stringer interface for RedelegationResp.
func (r RedelegationResponse) String() string {
	out := fmt.Sprintf(`Redelegations between:
  Delegator:                 %s
  Source Validator:          %s
  Destination Validator:     %s
  Entries:
`,
		r.DelegatorAddress, r.ValidatorSrcAddress, r.ValidatorDstAddress,
	)

	for i, entry := range r.Entries {
		out += fmt.Sprintf(`    Redelegation Entry #%d:
      Operation type:            %s
      Creation height:           %v
      Min time to unbond (unix): %v
      Initial Balance:           %s
      Shares:                    %s
      Balance:                   %s
`,
			i, entry.OpType, entry.CreationHeight, entry.CompletionTime, entry.InitialBalance, entry.SharesDst, entry.Balance,
		)
	}

	return strings.TrimRight(out, "\n")
}

type redelegationRespAlias RedelegationResponse

// MarshalJSON implements the json.Marshaler interface. This is so we can
// achieve a flattened structure while embedding other types.
func (r RedelegationResponse) MarshalJSON() ([]byte, error) {
	return json.Marshal((redelegationRespAlias)(r))
}

// UnmarshalJSON implements the json.Unmarshaler interface. This is so we can
// achieve a flattened structure while embedding other types.
func (r *RedelegationResponse) UnmarshalJSON(bz []byte) error {
	return json.Unmarshal(bz, (*redelegationRespAlias)(r))
}

// RedelegationResponses are a collection of RedelegationResp.
type RedelegationResponses []RedelegationResponse

func (r RedelegationResponses) String() (out string) {
	for _, red := range r {
		out += red.String() + "\n"
	}
	return strings.TrimSpace(out)
}
