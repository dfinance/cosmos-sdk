package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// ParamsHooks event hooks for params module
type ParamsHooks interface {
	AfterParamChanged(ctx sdk.Context, c ParamChange) error
	BeforeParamChanged(ctx sdk.Context, c ParamChange) error
}

type MultiParamsHooks []ParamsHooks

// NewMultiParamsHooks returns MultiParamsHooks container
func NewMultiParamsHooks(hooks ...ParamsHooks) MultiParamsHooks {
	return hooks
}

// nolint
func (h MultiParamsHooks) AfterParamChanged(ctx sdk.Context, c ParamChange) error {
	for i := range h {
		if err := h[i].AfterParamChanged(ctx, c); err != nil {
			return err
		}
	}

	return nil
}

// nolint
func (h MultiParamsHooks) BeforeParamChanged(ctx sdk.Context, c ParamChange) error {
	for i := range h {
		if err := h[i].BeforeParamChanged(ctx, c); err != nil {
			return err
		}
	}

	return nil
}
