package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// RouterKey is they name of the bank module
const RouterKey = ModuleName

// MsgSetFoundationAllocationRatio - high level transaction of the coin module
type MsgSetFoundationAllocationRatio struct {
	FromAddress sdk.AccAddress `json:"from_address" yaml:"from_address"`
	Ratio       sdk.Dec        `json:"ratio" yaml:"ratio"`
}

// NewMsgSetFoundationAllocationRatio - construct msg for change FoundationAllocationRatio.
func NewMsgSetFoundationAllocationRatio(fromAddr sdk.AccAddress, ratio sdk.Dec) MsgSetFoundationAllocationRatio {
	return MsgSetFoundationAllocationRatio{FromAddress: fromAddr, Ratio: ratio}
}

// Route Implements Msg.
func (msg MsgSetFoundationAllocationRatio) Route() string { return RouterKey }

// Type Implements Msg.
func (msg MsgSetFoundationAllocationRatio) Type() string { return "send" }

// ValidateBasic Implements Msg.
func (msg MsgSetFoundationAllocationRatio) ValidateBasic() error {
	if msg.FromAddress.Empty() {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "missing sender address")
	}
	if msg.Ratio.GT(sdk.NewDec(ChangeFoundationAllocationRatioMaxValue)) {
		return sdkerrors.Wrap(ErrWrongFoundationAllocationRatio, "ratio is greater than the maximum value for FoundationAllocationRatio")
	}
	if msg.Ratio.LT(sdk.NewDec(ChangeFoundationAllocationRatioMinValue)) {
		return sdkerrors.Wrap(ErrWrongFoundationAllocationRatio, "ratio is lower than the maximum value for FoundationAllocationRatio")
	}
	return nil
}

// GetSignBytes Implements Msg.
func (msg MsgSetFoundationAllocationRatio) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(msg))
}

// GetSigners Implements Msg.
func (msg MsgSetFoundationAllocationRatio) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.FromAddress}
}
