package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// CheckRatioVariable is ratio variable validation (0.0 <= value <= 1.0).
func CheckRatioVariable(name string, value sdk.Dec) error {
	if value.IsNegative() {
		return fmt.Errorf("%s: cannot be negative: %s", name, value)
	}
	if value.GT(sdk.OneDec()) {
		return fmt.Errorf("%s: too large: %s", name, value)
	}

	return nil
}
