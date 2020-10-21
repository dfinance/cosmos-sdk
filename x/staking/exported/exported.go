package exported

import (
	"time"

	"github.com/tendermint/tendermint/crypto"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// DelegationI delegation bond for a delegated proof of stake system
type DelegationI interface {
	GetDelegatorAddr() sdk.AccAddress // delegator sdk.AccAddress for the bond
	GetValidatorAddr() sdk.ValAddress // validator operator address
	GetBondingShares() sdk.Dec        // bonding tokens: amount of validator's shares held in this delegation
	GetLPShares() sdk.Dec             // liquidity tokens: amount of validator's shares held in this delegation
}

// ValidatorI expected validator functions
type ValidatorI interface {
	IsJailed() bool                                                // whether the validator is jailed
	GetMoniker() string                                            // moniker of the validator
	GetStatus() sdk.BondStatus                                     // status of the validator
	IsBonded() bool                                                // check if has a bonded status
	IsUnbonded() bool                                              // check if has status unbonded
	IsUnbonding() bool                                             // check if has status unbonding
	GetOperator() sdk.ValAddress                                   // operator address to receive/return validators coins
	GetConsPubKey() crypto.PubKey                                  // validation consensus pubkey
	GetConsAddr() sdk.ConsAddress                                  // validation consensus address
	GetBondedTokens() sdk.Int                                      // validator bonded tokens
	GetConsensusPower() int64                                      // validation power in tendermint
	LPPower() int64                                                // validator distribution / gov voting LP power fraction
	GetCommission() sdk.Dec                                        // validator commission rate
	GetMinSelfDelegation() sdk.Int                                 // validator minimum self delegation
	GetBondingDelegatorShares() sdk.Dec                            // bonding tokens: total outstanding delegator shares
	GetBondingTokens() sdk.Int                                     // bonding tokens: validation tokens
	BondingTokensFromShares(sdk.Dec) sdk.Dec                       // bonding tokens: token worth of provided delegator shares
	BondingTokensFromSharesTruncated(sdk.Dec) sdk.Dec              // bonding tokens: token worth of provided delegator shares, truncated
	BondingTokensFromSharesRoundUp(sdk.Dec) sdk.Dec                // bonding tokens: token worth of provided delegator shares, rounded up
	BondingSharesFromTokens(amt sdk.Int) (sdk.Dec, error)          // bonding tokens: shares worth of delegator's bond
	BondingSharesFromTokensTruncated(amt sdk.Int) (sdk.Dec, error) // bonding tokens: truncated shares worth of delegator's bond
	GetLPDelegatorShares() sdk.Dec                                 // liquidity tokens: total outstanding delegator shares
	GetLPTokens() sdk.Int                                          // liquidity tokens: validation tokens
	LPTokensFromShares(sdk.Dec) sdk.Dec                            // liquidity tokens: token worth of provided delegator shares
	LPTokensFromSharesTruncated(sdk.Dec) sdk.Dec                   // liquidity tokens: token worth of provided delegator shares, truncated
	LPTokensFromSharesRoundUp(sdk.Dec) sdk.Dec                     // liquidity tokens: token worth of provided delegator shares, rounded up
	LPSharesFromTokens(amt sdk.Int) (sdk.Dec, error)               // liquidity tokens: shares worth of delegator's bond
	LPSharesFromTokensTruncated(amt sdk.Int) (sdk.Dec, error)      // liquidity tokens: truncated shares worth of delegator's bond
	GetScheduledUnbondStartTime() time.Time                        // force unbond time (if not scheduled, result is time.Time{})
}
