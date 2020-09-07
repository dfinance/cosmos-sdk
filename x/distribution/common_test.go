package distribution

import (
	"github.com/tendermint/tendermint/crypto/ed25519"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

var (
	delPk1   = ed25519.GenPrivKey().PubKey()
	delAddr1 = sdk.AccAddress(delPk1.Address())
	delPk2   = ed25519.GenPrivKey().PubKey()
	delAddr2 = sdk.AccAddress(delPk2.Address())

	amount = sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(1)))
)
