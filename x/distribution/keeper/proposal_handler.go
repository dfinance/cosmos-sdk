package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/distribution/types"
)

// HandlePublicTreasuryPoolSpendProposal is a handler for executing a passed public treasury spend proposal.
func HandlePublicTreasuryPoolSpendProposal(ctx sdk.Context, k Keeper, p types.PublicTreasuryPoolSpendProposal) error {
	if k.blacklistedAddrs[p.Recipient.String()] {
		return sdkerrors.Wrapf(sdkerrors.ErrUnauthorized, "%s is blacklisted from receiving external funds", p.Recipient)
	}

	if err := k.DistributeFromPublicTreasuryPool(ctx, p.Amount, p.Recipient); err != nil {
		return err
	}

	logger := k.Logger(ctx)
	logger.Info(fmt.Sprintf("transferred %s from the public treasury pool to recipient %s", p.Amount, p.Recipient))

	return nil
}

// HandleTaxParamsUpdateProposal is a handler for executing a passed tax params update proposal.
func HandleTaxParamsUpdateProposal(ctx sdk.Context, k Keeper, p types.TaxParamsUpdateProposal) error {
	params := k.GetParams(ctx)
	params.ValidatorsPoolTax = p.ValidatorsPoolTax
	params.LiquidityProvidersPoolTax = p.LiquidityProvidersPoolTax
	params.PublicTreasuryPoolTax = p.PublicTreasuryPoolTax
	params.HARPTax = p.HARPTax
	k.SetParams(ctx, params)

	logger := k.Logger(ctx)
	logger.Info(fmt.Sprintf("tax params updated: %s", params))

	return nil
}
