package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

type ValidatorTokens struct {
	// Total shares issued to a validator's delegators
	DelegatorShares sdk.Dec `json:"delegator_shares" yaml:"delegator_shares" swaggertype:"string" format:"number" example:"0.123"`
	// Delegated tokens (incl. self-delegation for bonding tokens)
	Tokens sdk.Int `json:"tokens" yaml:"tokens" swaggertype:"string" format:"integer" example:"100"`
}

// TokensFromShares calculates the token worth of provided shares.
func (t ValidatorTokens) TokensFromShares(shares sdk.Dec) sdk.Dec {
	if t.DelegatorShares.IsZero() {
		return sdk.ZeroDec()
	}

	return (shares.MulInt(t.Tokens)).Quo(t.DelegatorShares)
}

// TokensFromSharesTruncated calculates the token worth of provided shares, truncated.
func (t ValidatorTokens) TokensFromSharesTruncated(shares sdk.Dec) sdk.Dec {
	if t.DelegatorShares.IsZero() {
		return sdk.ZeroDec()
	}

	return (shares.MulInt(t.Tokens)).QuoTruncate(t.DelegatorShares)
}

// TokensFromSharesRoundUp returns the token worth of provided shares, rounded up.
func (t ValidatorTokens) TokensFromSharesRoundUp(shares sdk.Dec) sdk.Dec {
	if t.DelegatorShares.IsZero() {
		return sdk.ZeroDec()
	}

	return (shares.MulInt(t.Tokens)).QuoRoundUp(t.DelegatorShares)
}

// SharesFromTokens returns the shares of a delegation given a bond amount.
// It returns an error if the validator has no tokens.
func (t ValidatorTokens) SharesFromTokens(amount sdk.Int) (sdk.Dec, error) {
	if t.Tokens.IsZero() {
		return sdk.ZeroDec(), ErrInsufficientShares
	}

	return t.DelegatorShares.MulInt(amount).QuoInt(t.Tokens), nil
}

// SharesFromTokensTruncated returns the truncated shares of a delegation given a bond amount.
// It returns an error if the validator has no tokens.
func (t ValidatorTokens) SharesFromTokensTruncated(amt sdk.Int) (sdk.Dec, error) {
	if t.Tokens.IsZero() {
		return sdk.ZeroDec(), ErrInsufficientShares
	}

	return t.DelegatorShares.MulInt(amt).QuoTruncate(t.Tokens.ToDec()), nil
}

// AddTokensFromDel adds tokens to a validator.
func (t ValidatorTokens) AddTokensFromDel(amount sdk.Int) (ValidatorTokens, sdk.Dec) {
	// calculate the shares to issue
	var issuedShares sdk.Dec
	if t.DelegatorShares.IsZero() {
		// the first delegation to a validator sets the exchange rate to one
		issuedShares = amount.ToDec()
	} else {
		shares, err := t.SharesFromTokens(amount)
		if err != nil {
			panic(err)
		}

		issuedShares = shares
	}

	t.Tokens = t.Tokens.Add(amount)
	t.DelegatorShares = t.DelegatorShares.Add(issuedShares)

	return t, issuedShares
}

// RemoveTokens removes tokens from a validator.
func (t ValidatorTokens) RemoveTokens(tokens sdk.Int) ValidatorTokens {
	if tokens.IsNegative() {
		panic(fmt.Sprintf("should not happen: trying to remove negative tokens %v", tokens))
	}
	if t.Tokens.LT(tokens) {
		panic(fmt.Sprintf("should not happen: only have %v tokens, trying to remove %v", t.Tokens, tokens))
	}
	t.Tokens = t.Tokens.Sub(tokens)

	return t
}

// RemoveDelShares removes delegator shares from a validator.
// NOTE: because token fractions are left in the valiadator,
//       the exchange rate of future shares of this validator can increase.
func (t ValidatorTokens) RemoveDelShares(delShares sdk.Dec) (ValidatorTokens, sdk.Int) {
	var issuedTokens sdk.Int

	remainingShares := t.DelegatorShares.Sub(delShares)
	if remainingShares.IsZero() {
		// last delegation share gets any trimmings
		issuedTokens = t.Tokens
		t.Tokens = sdk.ZeroInt()
	} else {
		// leave excess tokens in the validator
		// however fully use all the delegator shares
		issuedTokens = t.TokensFromShares(delShares).TruncateInt()
		t.Tokens = t.Tokens.Sub(issuedTokens)
		if t.Tokens.IsNegative() {
			panic("attempting to remove more tokens than available in validator")
		}
	}
	t.DelegatorShares = remainingShares

	return t, issuedTokens
}

// Equal compares two ValidatorTokens objects.
func (t ValidatorTokens) Equal(t2 ValidatorTokens) bool {
	return t.DelegatorShares.Equal(t2.DelegatorShares) && t.Tokens.Equal(t2.Tokens)
}

// String returns a human readable string representation of a ValidatorTokens.
func (t ValidatorTokens) String(name string) string {
	return fmt.Sprintf(`  %s:
    Delegator Shares: %s
    Tokens:           %s`,
		name, t.DelegatorShares, t.Tokens,
	)
}

// NewValidatorTokens creates an empty ValidatorTokens object.
func NewValidatorTokens() ValidatorTokens {
	return ValidatorTokens{
		DelegatorShares: sdk.ZeroDec(),
		Tokens:          sdk.ZeroInt(),
	}
}
