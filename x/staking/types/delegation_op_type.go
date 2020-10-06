package types

import "fmt"

// DelegationOpType defines delegate / undelegate / redelegate operation type: bonding token / liquidity tokens.
type DelegationOpType string

const (
	BondingDelOpType   DelegationOpType = "bonding"
	LiquidityDelOpType DelegationOpType = "liquidity"
)

func (t DelegationOpType) IsBonding() bool {
	return t == BondingDelOpType
}

func (t DelegationOpType) IsLiquidity() bool {
	return t == LiquidityDelOpType
}

func (t DelegationOpType) Validate() error {
	switch t {
	case BondingDelOpType, LiquidityDelOpType:
		return nil
	default:
		return fmt.Errorf("unknown delegation operation: %s", string(t))
	}
}

func (t DelegationOpType) String() string {
	return string(t)
}
