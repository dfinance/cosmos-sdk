package types

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// x/mint module sentinel errors
var (
	ErrWrongFoundationAllocationRatio = sdkerrors.Register(ModuleName, 1, "foundation allocation ratio is wrong")
	ErrExceededTimeLimit              = sdkerrors.Register(ModuleName, 2, "exceeded time limit")
)
