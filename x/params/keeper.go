package params

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/params/subspace"
	"github.com/cosmos/cosmos-sdk/x/params/types"

	"github.com/tendermint/tendermint/libs/log"
)

// Implements ParametersHooks interface
var _ types.ParamsHooks = Keeper{}

// Keeper of the global paramstore
type Keeper struct {
	cdc              *codec.Codec
	key              sdk.StoreKey
	tkey             sdk.StoreKey
	spaces           map[string]*Subspace
	restrictedParams RestrictedParams
	hooks            ParametersHooks
}

// NewKeeper constructs a params keeper
func NewKeeper(cdc *codec.Codec, key, tkey sdk.StoreKey) Keeper {
	return Keeper{
		cdc:    cdc,
		key:    key,
		tkey:   tkey,
		spaces: make(map[string]*Subspace),
	}
}

// Logger returns a module-specific logger.
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

// Allocate subspace used for keepers
func (k Keeper) Subspace(s string) Subspace {
	_, ok := k.spaces[s]
	if ok {
		panic("subspace already occupied")
	}

	if s == "" {
		panic("cannot use empty string for subspace")
	}

	space := subspace.NewSubspace(k.cdc, k.key, k.tkey, s)
	k.spaces[s] = &space

	return space
}

// Get existing substore from keeper
func (k Keeper) GetSubspace(s string) (Subspace, bool) {
	space, ok := k.spaces[s]
	if !ok {
		return Subspace{}, false
	}
	return *space, ok
}

// SetRestrictedParams sets restricted params that will be rejected for a parameter change proposal.
func (k *Keeper) SetRestrictedParams(rp RestrictedParams) {
	k.restrictedParams = rp
}

// CheckRestrictions checks subspace and key in restricted list.
func (k Keeper) CheckRestrictions(subspace, key string) error {
	for _, param := range k.restrictedParams {
		if param.Subspace == subspace && param.Key == key {
			return sdkerrors.Wrapf(ErrDisallowedParameter, "subspace: %s, key: %s", param.Subspace, param.Key)
		}
	}

	return nil
}

// AfterParamChanged - call hook after changed a parameter
func (k Keeper) AfterParamChanged(ctx sdk.Context, c ParamChange) error {
	if k.hooks != nil {
		return k.hooks.AfterParamChanged(ctx, c)
	}

	return nil
}

// BeforeParamChanged - call hook after changed a parameter
func (k Keeper) BeforeParamChanged(ctx sdk.Context, c ParamChange) error {
	if k.hooks != nil {
		return k.hooks.BeforeParamChanged(ctx, c)
	}

	return nil
}

// SetHooks sets hooks for call on the param update
func (k *Keeper) SetHooks(h ParametersHooks) *Keeper {
	if k.hooks != nil {
		panic("cannot set hooks twice")
	}
	k.hooks = h
	return k
}
