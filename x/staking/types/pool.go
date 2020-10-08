package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// names used as root for pool module accounts
const (
	NotBondedPoolName = "not_bonded_tokens_pool"
	BondedPoolName    = "bonded_tokens_pool"
	LiquidityPoolName = "liquidity_tokens_pool"
)

// Pool - tracking bonded and not-bonded token supply of the bond denomination
type Pool struct {
	// Bonding tokens which are not bonded to a validators (unbonded or unbonding)
	NotBondedTokens sdk.Int `json:"not_bonded_tokens" yaml:"not_bonded_tokens" format:"string" type:"integer" swaggertype:"string" format:"integer" example:"500"`
	// Bonding tokens which are currently bonded to a validators
	BondedTokens sdk.Int `json:"bonded_tokens" yaml:"bonded_tokens" swaggertype:"string" format:"integer" example:"50000"`
	// Liquidity tokens which are currently bonded to a validators
	LiquidityTokens sdk.Int `json:"liquidity_tokens" yaml:"liquidity_tokens" swaggertype:"string" format:"integer" example:"10000"`
}

// NewPool creates a new Pool instance used for queries
func NewPool(notBonded, bonded, liquidity sdk.Int) Pool {
	return Pool{
		NotBondedTokens: notBonded,
		BondedTokens:    bonded,
		LiquidityTokens: liquidity,
	}
}

// String returns a human readable string representation of a pool.
func (p Pool) String() string {
	return fmt.Sprintf(`Pool:	
  Not Bonded Tokens:  %s	
  Bonded Tokens:      %s
  Liquidity Tokens:   %s`,
		p.NotBondedTokens, p.BondedTokens, p.LiquidityTokens,
	)
}
