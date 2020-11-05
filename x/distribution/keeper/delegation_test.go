package keeper

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/distribution/types"
	"github.com/cosmos/cosmos-sdk/x/staking"
)

// Allocate tokens to validator (1 delegator) with 50% commission.
func TestCalculateRewardsBasic(t *testing.T) {
	ctx, _, k, sk, _ := CreateTestInputDefault(t, false, 1000)
	sh := staking.NewHandler(sk)

	// create validator with 50% commission
	commission := staking.NewCommissionRates(sdk.NewDecWithPrec(5, 1), sdk.NewDecWithPrec(5, 1), sdk.NewDec(0))
	msg := staking.NewMsgCreateValidator(
		valOpAddr1, valConsPk1, sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(100)), staking.Description{}, commission, minSelfDelegation,
	)

	res, err := sh(ctx, msg)
	require.NoError(t, err)
	require.NotNil(t, res)

	// end block to bond validator
	staking.EndBlocker(ctx, sk)

	// next block
	ctx = ctx.WithBlockHeight(ctx.BlockHeight() + 1)

	// fetch validator and delegation
	val := sk.Validator(ctx, valOpAddr1)
	del := sk.Delegation(ctx, sdk.AccAddress(valOpAddr1), valOpAddr1)

	// historical count should be 2 (once for validator init, once for delegation init)
	require.Equal(t, uint64(2), k.GetValidatorHistoricalReferenceCount(ctx))

	// end period
	endingPeriod := k.incrementValidatorPeriod(ctx, val)

	// historical count should be 2 still
	require.Equal(t, uint64(2), k.GetValidatorHistoricalReferenceCount(ctx))

	// calculate delegation rewards
	rewards, _ := k.calculateDelegationRewards(ctx, val, del, endingPeriod)

	// rewards should be zero
	require.True(t, rewards.IsZero())

	// allocate some rewards
	initial := int64(10)
	tokens := sdk.DecCoins{{Denom: sdk.DefaultBondDenom, Amount: sdk.NewDec(initial)}}
	k.AllocateTokensToValidator(ctx, val, tokens, sdk.DecCoins{})

	// end period
	endingPeriod = k.incrementValidatorPeriod(ctx, val)

	// calculate delegation rewards
	rewards, _ = k.calculateDelegationRewards(ctx, val, del, endingPeriod)

	// rewards should be half the tokens
	require.Equal(t, sdk.DecCoins{{Denom: sdk.DefaultBondDenom, Amount: sdk.NewDec(initial / 2)}}, rewards)

	// commission should be the other half
	require.Equal(t, sdk.DecCoins{{Denom: sdk.DefaultBondDenom, Amount: sdk.NewDec(initial / 2)}}, k.GetValidatorAccumulatedCommission(ctx, valOpAddr1))
}

// Allocate tokens to validator (1 delegator) with 50% commission and slash with 50% fraction.
func TestCalculateRewardsAfterSlash(t *testing.T) {
	ctx, _, k, sk, _ := CreateTestInputDefault(t, false, 1000)
	sh := staking.NewHandler(sk)

	// create validator with 50% commission
	commission := staking.NewCommissionRates(sdk.NewDecWithPrec(5, 1), sdk.NewDecWithPrec(5, 1), sdk.NewDec(0))
	valPower := int64(100)
	valTokens := sdk.TokensFromConsensusPower(valPower)
	msg := staking.NewMsgCreateValidator(valOpAddr1, valConsPk1,
		sdk.NewCoin(sdk.DefaultBondDenom, valTokens), staking.Description{}, commission, minSelfDelegation)

	res, err := sh(ctx, msg)
	require.NoError(t, err)
	require.NotNil(t, res)

	// end block to bond validator
	staking.EndBlocker(ctx, sk)

	// next block
	ctx = ctx.WithBlockHeight(ctx.BlockHeight() + 1)

	// fetch validator and delegation
	val := sk.Validator(ctx, valOpAddr1)
	del := sk.Delegation(ctx, sdk.AccAddress(valOpAddr1), valOpAddr1)

	// end period
	endingPeriod := k.incrementValidatorPeriod(ctx, val)

	// calculate delegation rewards
	rewards, _ := k.calculateDelegationRewards(ctx, val, del, endingPeriod)

	// rewards should be zero
	require.True(t, rewards.IsZero())

	// start out block height
	ctx = ctx.WithBlockHeight(ctx.BlockHeight() + 3)

	// slash the validator by 50%
	sk.Slash(ctx, valConsAddr1, ctx.BlockHeight(), valPower, sdk.NewDecWithPrec(5, 1))

	// retrieve validator
	val = sk.Validator(ctx, valOpAddr1)

	// increase block height
	ctx = ctx.WithBlockHeight(ctx.BlockHeight() + 3)

	// allocate some rewards
	initial := sdk.TokensFromConsensusPower(10)
	tokens := sdk.DecCoins{{Denom: sdk.DefaultBondDenom, Amount: initial.ToDec()}}
	k.AllocateTokensToValidator(ctx, val, tokens, sdk.DecCoins{})

	// end period
	endingPeriod = k.incrementValidatorPeriod(ctx, val)

	// calculate delegation rewards
	rewards, _ = k.calculateDelegationRewards(ctx, val, del, endingPeriod)

	// rewards should be half the tokens
	require.Equal(t, sdk.DecCoins{{Denom: sdk.DefaultBondDenom, Amount: initial.QuoRaw(2).ToDec()}}, rewards)

	// commission should be the other half
	require.Equal(t, sdk.DecCoins{{Denom: sdk.DefaultBondDenom, Amount: initial.QuoRaw(2).ToDec()}},
		k.GetValidatorAccumulatedCommission(ctx, valOpAddr1))
}

// Slash, allocate, slash with 50% (first slash has no rewards, second should lower rewards to 1/2).
func TestCalculateRewardsAfterManySlashes(t *testing.T) {
	ctx, _, k, sk, _ := CreateTestInputDefault(t, false, 1000)
	sh := staking.NewHandler(sk)

	// create validator with 50% commission
	power := int64(100)
	valTokens := sdk.TokensFromConsensusPower(power)
	commission := staking.NewCommissionRates(sdk.NewDecWithPrec(5, 1), sdk.NewDecWithPrec(5, 1), sdk.NewDec(0))
	msg := staking.NewMsgCreateValidator(valOpAddr1, valConsPk1,
		sdk.NewCoin(sdk.DefaultBondDenom, valTokens), staking.Description{}, commission, minSelfDelegation)

	res, err := sh(ctx, msg)
	require.NoError(t, err)
	require.NotNil(t, res)

	// end block to bond validator
	staking.EndBlocker(ctx, sk)

	// next block
	ctx = ctx.WithBlockHeight(ctx.BlockHeight() + 1)

	// fetch validator and delegation
	val := sk.Validator(ctx, valOpAddr1)
	del := sk.Delegation(ctx, sdk.AccAddress(valOpAddr1), valOpAddr1)

	// end period
	endingPeriod := k.incrementValidatorPeriod(ctx, val)

	// calculate delegation rewards
	rewards, _ := k.calculateDelegationRewards(ctx, val, del, endingPeriod)

	// rewards should be zero
	require.True(t, rewards.IsZero())

	// start out block height
	ctx = ctx.WithBlockHeight(ctx.BlockHeight() + 3)

	// slash the validator by 50%
	sk.Slash(ctx, valConsAddr1, ctx.BlockHeight(), power, sdk.NewDecWithPrec(5, 1))

	// fetch the validator again
	val = sk.Validator(ctx, valOpAddr1)

	// increase block height
	ctx = ctx.WithBlockHeight(ctx.BlockHeight() + 3)

	// allocate some rewards
	initial := sdk.TokensFromConsensusPower(10)
	tokens := sdk.DecCoins{{Denom: sdk.DefaultBondDenom, Amount: initial.ToDec()}}
	k.AllocateTokensToValidator(ctx, val, tokens, sdk.DecCoins{})

	// slash the validator by 50% again
	sk.Slash(ctx, valConsAddr1, ctx.BlockHeight(), power/2, sdk.NewDecWithPrec(5, 1))

	// fetch the validator again
	val = sk.Validator(ctx, valOpAddr1)

	// increase block height
	ctx = ctx.WithBlockHeight(ctx.BlockHeight() + 3)

	// allocate some more rewards
	k.AllocateTokensToValidator(ctx, val, tokens, sdk.DecCoins{})

	// end period
	endingPeriod = k.incrementValidatorPeriod(ctx, val)

	// calculate delegation rewards
	rewards, _ = k.calculateDelegationRewards(ctx, val, del, endingPeriod)

	// rewards should be half the tokens
	require.Equal(t, sdk.DecCoins{{Denom: sdk.DefaultBondDenom, Amount: initial.ToDec()}}, rewards)

	// commission should be the other half
	require.Equal(t, sdk.DecCoins{{Denom: sdk.DefaultBondDenom, Amount: initial.ToDec()}},
		k.GetValidatorAccumulatedCommission(ctx, valOpAddr1))
}

// Allocate tokens per delegator (twice) with commission 50%.
func TestCalculateRewardsMultiDelegator(t *testing.T) {
	ctx, _, k, sk, _ := CreateTestInputDefault(t, false, 1000)
	sh := staking.NewHandler(sk)

	// create validator with 50% commission
	commission := staking.NewCommissionRates(sdk.NewDecWithPrec(5, 1), sdk.NewDecWithPrec(5, 1), sdk.NewDec(0))
	msg := staking.NewMsgCreateValidator(valOpAddr1, valConsPk1,
		sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(100)), staking.Description{}, commission, minSelfDelegation)

	res, err := sh(ctx, msg)
	require.NoError(t, err)
	require.NotNil(t, res)

	// end block to bond validator
	staking.EndBlocker(ctx, sk)

	// next block
	ctx = ctx.WithBlockHeight(ctx.BlockHeight() + 1)

	// fetch validator and delegation
	val := sk.Validator(ctx, valOpAddr1)
	del1 := sk.Delegation(ctx, sdk.AccAddress(valOpAddr1), valOpAddr1)

	// allocate some rewards
	initial := int64(20)
	tokens := sdk.DecCoins{{Denom: sdk.DefaultBondDenom, Amount: sdk.NewDec(initial)}}
	k.AllocateTokensToValidator(ctx, val, tokens, sdk.DecCoins{})

	// second delegation
	msg2 := staking.NewMsgDelegate(sdk.AccAddress(valOpAddr2), valOpAddr1, sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(100)))

	res, err = sh(ctx, msg2)
	require.NoError(t, err)
	require.NotNil(t, res)

	del2 := sk.Delegation(ctx, sdk.AccAddress(valOpAddr2), valOpAddr1)

	// fetch updated validator
	val = sk.Validator(ctx, valOpAddr1)

	// end block
	staking.EndBlocker(ctx, sk)

	// next block
	ctx = ctx.WithBlockHeight(ctx.BlockHeight() + 1)

	// allocate some more rewards
	k.AllocateTokensToValidator(ctx, val, tokens, sdk.DecCoins{})

	// end period
	endingPeriod := k.incrementValidatorPeriod(ctx, val)

	// calculate delegation rewards for del1
	rewards, _ := k.calculateDelegationRewards(ctx, val, del1, endingPeriod)

	// rewards for del1 should be 3/4 initial
	require.Equal(t, sdk.DecCoins{{Denom: sdk.DefaultBondDenom, Amount: sdk.NewDec(initial * 3 / 4)}}, rewards)

	// calculate delegation rewards for del2
	rewards, _ = k.calculateDelegationRewards(ctx, val, del2, endingPeriod)

	// rewards for del2 should be 1/4 initial
	require.Equal(t, sdk.DecCoins{{Denom: sdk.DefaultBondDenom, Amount: sdk.NewDec(initial * 1 / 4)}}, rewards)

	// commission should be equal to initial (50% twice)
	require.Equal(t, sdk.DecCoins{{Denom: sdk.DefaultBondDenom, Amount: sdk.NewDec(initial)}}, k.GetValidatorAccumulatedCommission(ctx, valOpAddr1))
}

// Allocate tokens to validator (1 delegator) with 50% commission and withdraw delegator rewards and validator commission.
func TestWithdrawDelegationRewardsBasic(t *testing.T) {
	balancePower := int64(1000)
	balanceTokens := sdk.TokensFromConsensusPower(balancePower)
	ctx, ak, k, sk, _ := CreateTestInputDefault(t, false, balancePower)
	sh := staking.NewHandler(sk)

	// set module account coins
	distrAcc := k.GetDistributionAccount(ctx)
	distrAcc.SetCoins(sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, balanceTokens)))
	k.supplyKeeper.SetModuleAccount(ctx, distrAcc)

	// create validator with 50% commission
	power := int64(100)
	valTokens := sdk.TokensFromConsensusPower(power)
	commission := staking.NewCommissionRates(sdk.NewDecWithPrec(5, 1), sdk.NewDecWithPrec(5, 1), sdk.NewDec(0))
	msg := staking.NewMsgCreateValidator(
		valOpAddr1, valConsPk1,
		sdk.NewCoin(sdk.DefaultBondDenom, valTokens),
		staking.Description{}, commission, minSelfDelegation,
	)

	res, err := sh(ctx, msg)
	require.NoError(t, err)
	require.NotNil(t, res)

	// assert correct initial balance
	expTokens := balanceTokens.Sub(valTokens)
	require.Equal(t,
		sdk.Coins{sdk.NewCoin(sdk.DefaultBondDenom, expTokens)},
		ak.GetAccount(ctx, sdk.AccAddress(valOpAddr1)).GetCoins(),
	)

	// end block to bond validator
	staking.EndBlocker(ctx, sk)

	// next block
	ctx = ctx.WithBlockHeight(ctx.BlockHeight() + 1)

	// fetch validator and delegation
	val := sk.Validator(ctx, valOpAddr1)

	// allocate some rewards
	initial := sdk.TokensFromConsensusPower(10)
	tokens := sdk.DecCoins{sdk.NewDecCoin(sdk.DefaultBondDenom, initial)}

	k.AllocateTokensToValidator(ctx, val, tokens, sdk.DecCoins{})

	// historical count should be 2 (initial + latest for delegation)
	require.Equal(t, uint64(2), k.GetValidatorHistoricalReferenceCount(ctx))

	// withdraw rewards
	_, err = k.WithdrawDelegationRewards(ctx, sdk.AccAddress(valOpAddr1), valOpAddr1)
	require.Nil(t, err)

	// historical count should still be 2 (added one record, cleared one)
	require.Equal(t, uint64(2), k.GetValidatorHistoricalReferenceCount(ctx))

	// assert correct balance
	exp := balanceTokens.Sub(valTokens).Add(initial.QuoRaw(2))
	require.Equal(t,
		sdk.Coins{sdk.NewCoin(sdk.DefaultBondDenom, exp)},
		ak.GetAccount(ctx, sdk.AccAddress(valOpAddr1)).GetCoins(),
	)

	// withdraw commission
	_, err = k.WithdrawValidatorCommission(ctx, valOpAddr1)
	require.Nil(t, err)

	// assert correct balance
	exp = balanceTokens.Sub(valTokens).Add(initial)
	require.Equal(t,
		sdk.Coins{sdk.NewCoin(sdk.DefaultBondDenom, exp)},
		ak.GetAccount(ctx, sdk.AccAddress(valOpAddr1)).GetCoins(),
	)
}

// Allocate tokens twice and slash the first allocation.
func TestCalculateRewardsAfterManySlashesInSameBlock(t *testing.T) {
	ctx, _, k, sk, _ := CreateTestInputDefault(t, false, 1000)
	sh := staking.NewHandler(sk)

	// create validator with 50% commission
	power := int64(100)
	valTokens := sdk.TokensFromConsensusPower(power)
	commission := staking.NewCommissionRates(sdk.NewDecWithPrec(5, 1), sdk.NewDecWithPrec(5, 1), sdk.NewDec(0))
	msg := staking.NewMsgCreateValidator(valOpAddr1, valConsPk1,
		sdk.NewCoin(sdk.DefaultBondDenom, valTokens), staking.Description{}, commission, minSelfDelegation)

	res, err := sh(ctx, msg)
	require.NoError(t, err)
	require.NotNil(t, res)

	// end block to bond validator
	staking.EndBlocker(ctx, sk)

	// next block
	ctx = ctx.WithBlockHeight(ctx.BlockHeight() + 1)

	// fetch validator and delegation
	val := sk.Validator(ctx, valOpAddr1)
	del := sk.Delegation(ctx, sdk.AccAddress(valOpAddr1), valOpAddr1)

	// end period
	endingPeriod := k.incrementValidatorPeriod(ctx, val)

	// calculate delegation rewards
	rewards, _ := k.calculateDelegationRewards(ctx, val, del, endingPeriod)

	// rewards should be zero
	require.True(t, rewards.IsZero())

	// start out block height
	ctx = ctx.WithBlockHeight(ctx.BlockHeight() + 3)

	// allocate some rewards
	initial := sdk.TokensFromConsensusPower(10).ToDec()
	tokens := sdk.DecCoins{{Denom: sdk.DefaultBondDenom, Amount: initial}}
	k.AllocateTokensToValidator(ctx, val, tokens, sdk.DecCoins{})

	// slash the validator by 50%
	sk.Slash(ctx, valConsAddr1, ctx.BlockHeight(), power, sdk.NewDecWithPrec(5, 1))

	// slash the validator by 50% again
	sk.Slash(ctx, valConsAddr1, ctx.BlockHeight(), power/2, sdk.NewDecWithPrec(5, 1))

	// fetch the validator again
	val = sk.Validator(ctx, valOpAddr1)

	// increase block height
	ctx = ctx.WithBlockHeight(ctx.BlockHeight() + 3)

	// allocate some more rewards
	k.AllocateTokensToValidator(ctx, val, tokens, sdk.DecCoins{})

	// end period
	endingPeriod = k.incrementValidatorPeriod(ctx, val)

	// calculate delegation rewards
	rewards, _ = k.calculateDelegationRewards(ctx, val, del, endingPeriod)

	// rewards should be half the tokens
	require.Equal(t, sdk.DecCoins{{Denom: sdk.DefaultBondDenom, Amount: initial}}, rewards)

	// commission should be the other half
	require.Equal(t, sdk.DecCoins{{Denom: sdk.DefaultBondDenom, Amount: initial}}, k.GetValidatorAccumulatedCommission(ctx, valOpAddr1))
}

func TestCalculateRewardsMultiDelegatorMultiSlash(t *testing.T) {
	ctx, _, k, sk, _ := CreateTestInputDefault(t, false, 1000)
	sh := staking.NewHandler(sk)

	// create validator with 50% commission
	commission := staking.NewCommissionRates(sdk.NewDecWithPrec(5, 1), sdk.NewDecWithPrec(5, 1), sdk.NewDec(0))
	power := int64(100)
	valTokens := sdk.TokensFromConsensusPower(power)
	msg := staking.NewMsgCreateValidator(valOpAddr1, valConsPk1,
		sdk.NewCoin(sdk.DefaultBondDenom, valTokens), staking.Description{}, commission, minSelfDelegation)

	res, err := sh(ctx, msg)
	require.NoError(t, err)
	require.NotNil(t, res)

	// end block to bond validator
	staking.EndBlocker(ctx, sk)

	// next block
	ctx = ctx.WithBlockHeight(ctx.BlockHeight() + 1)

	// fetch validator and delegation
	val := sk.Validator(ctx, valOpAddr1)
	del1 := sk.Delegation(ctx, sdk.AccAddress(valOpAddr1), valOpAddr1)

	// allocate some rewards
	initial := sdk.TokensFromConsensusPower(30).ToDec()
	tokens := sdk.DecCoins{{Denom: sdk.DefaultBondDenom, Amount: initial}}
	k.AllocateTokensToValidator(ctx, val, tokens, sdk.DecCoins{})

	// slash the validator
	ctx = ctx.WithBlockHeight(ctx.BlockHeight() + 3)
	sk.Slash(ctx, valConsAddr1, ctx.BlockHeight(), power, sdk.NewDecWithPrec(5, 1))
	ctx = ctx.WithBlockHeight(ctx.BlockHeight() + 3)

	// second delegation
	delTokens := sdk.TokensFromConsensusPower(100)
	msg2 := staking.NewMsgDelegate(sdk.AccAddress(valOpAddr2), valOpAddr1,
		sdk.NewCoin(sdk.DefaultBondDenom, delTokens))

	res, err = sh(ctx, msg2)
	require.NoError(t, err)
	require.NotNil(t, res)

	del2 := sk.Delegation(ctx, sdk.AccAddress(valOpAddr2), valOpAddr1)

	// end block
	staking.EndBlocker(ctx, sk)

	// next block
	ctx = ctx.WithBlockHeight(ctx.BlockHeight() + 1)

	// allocate some more rewards
	k.AllocateTokensToValidator(ctx, val, tokens, sdk.DecCoins{})

	// slash the validator again
	ctx = ctx.WithBlockHeight(ctx.BlockHeight() + 3)
	sk.Slash(ctx, valConsAddr1, ctx.BlockHeight(), power, sdk.NewDecWithPrec(5, 1))
	ctx = ctx.WithBlockHeight(ctx.BlockHeight() + 3)

	// fetch updated validator
	val = sk.Validator(ctx, valOpAddr1)

	// end period
	endingPeriod := k.incrementValidatorPeriod(ctx, val)

	// calculate delegation rewards for del1
	rewards, _ := k.calculateDelegationRewards(ctx, val, del1, endingPeriod)

	// rewards for del1 should be 2/3 initial (half initial first period, 1/6 initial second period)
	require.Equal(t, sdk.DecCoins{{Denom: sdk.DefaultBondDenom, Amount: initial.QuoInt64(2).Add(initial.QuoInt64(6))}}, rewards)

	// calculate delegation rewards for del2
	rewards, _ = k.calculateDelegationRewards(ctx, val, del2, endingPeriod)

	// rewards for del2 should be initial / 3
	require.Equal(t, sdk.DecCoins{{Denom: sdk.DefaultBondDenom, Amount: initial.QuoInt64(3)}}, rewards)

	// commission should be equal to initial (twice 50% commission, unaffected by slashing)
	require.Equal(t, sdk.DecCoins{{Denom: sdk.DefaultBondDenom, Amount: initial}}, k.GetValidatorAccumulatedCommission(ctx, valOpAddr1))
}

func TestCalculateRewardsMultiDelegatorMultWithdraw(t *testing.T) {
	ctx, _, k, sk, _ := CreateTestInputDefault(t, false, 1000)
	sh := staking.NewHandler(sk)
	initial := int64(20)

	// set module account coins
	distrAcc := k.GetDistributionAccount(ctx)
	distrAcc.SetCoins(sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(1000))))
	k.supplyKeeper.SetModuleAccount(ctx, distrAcc)

	tokens := sdk.DecCoins{sdk.NewDecCoinFromDec(sdk.DefaultBondDenom, sdk.NewDec(initial))}

	// create validator with 50% commission
	commission := staking.NewCommissionRates(sdk.NewDecWithPrec(5, 1), sdk.NewDecWithPrec(5, 1), sdk.NewDec(0))
	msg := staking.NewMsgCreateValidator(valOpAddr1, valConsPk1,
		sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(100)), staking.Description{}, commission, minSelfDelegation)

	res, err := sh(ctx, msg)
	require.NoError(t, err)
	require.NotNil(t, res)

	// end block to bond validator
	staking.EndBlocker(ctx, sk)

	// next block
	ctx = ctx.WithBlockHeight(ctx.BlockHeight() + 1)

	// fetch validator and delegation
	val := sk.Validator(ctx, valOpAddr1)
	del1 := sk.Delegation(ctx, sdk.AccAddress(valOpAddr1), valOpAddr1)

	// allocate some rewards
	k.AllocateTokensToValidator(ctx, val, tokens, sdk.DecCoins{})

	// historical count should be 2 (validator init, delegation init)
	require.Equal(t, uint64(2), k.GetValidatorHistoricalReferenceCount(ctx))

	// second delegation
	msg2 := staking.NewMsgDelegate(sdk.AccAddress(valOpAddr2), valOpAddr1, sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(100)))
	res, err = sh(ctx, msg2)
	require.NoError(t, err)
	require.NotNil(t, res)

	// historical count should be 3 (second delegation init)
	require.Equal(t, uint64(3), k.GetValidatorHistoricalReferenceCount(ctx))

	// fetch updated validator
	val = sk.Validator(ctx, valOpAddr1)
	del2 := sk.Delegation(ctx, sdk.AccAddress(valOpAddr2), valOpAddr1)

	// end block
	staking.EndBlocker(ctx, sk)

	// next block
	ctx = ctx.WithBlockHeight(ctx.BlockHeight() + 1)

	// allocate some more rewards
	k.AllocateTokensToValidator(ctx, val, tokens, sdk.DecCoins{})

	// first delegator withdraws
	k.WithdrawDelegationRewards(ctx, sdk.AccAddress(valOpAddr1), valOpAddr1)

	// second delegator withdraws
	k.WithdrawDelegationRewards(ctx, sdk.AccAddress(valOpAddr2), valOpAddr1)

	// historical count should be 3 (validator init + two delegations)
	require.Equal(t, uint64(3), k.GetValidatorHistoricalReferenceCount(ctx))

	// validator withdraws commission
	k.WithdrawValidatorCommission(ctx, valOpAddr1)

	// end period
	endingPeriod := k.incrementValidatorPeriod(ctx, val)

	// calculate delegation rewards for del1
	rewards, _ := k.calculateDelegationRewards(ctx, val, del1, endingPeriod)

	// rewards for del1 should be zero
	require.True(t, rewards.IsZero())

	// calculate delegation rewards for del2
	rewards, _ = k.calculateDelegationRewards(ctx, val, del2, endingPeriod)

	// rewards for del2 should be zero
	require.True(t, rewards.IsZero())

	// commission should be zero
	require.True(t, k.GetValidatorAccumulatedCommission(ctx, valOpAddr1).IsZero())

	// next block
	ctx = ctx.WithBlockHeight(ctx.BlockHeight() + 1)

	// allocate some more rewards
	k.AllocateTokensToValidator(ctx, val, tokens, sdk.DecCoins{})

	// first delegator withdraws again
	k.WithdrawDelegationRewards(ctx, sdk.AccAddress(valOpAddr1), valOpAddr1)

	// end period
	endingPeriod = k.incrementValidatorPeriod(ctx, val)

	// calculate delegation rewards for del1
	rewards, _ = k.calculateDelegationRewards(ctx, val, del1, endingPeriod)

	// rewards for del1 should be zero
	require.True(t, rewards.IsZero())

	// calculate delegation rewards for del2
	rewards, _ = k.calculateDelegationRewards(ctx, val, del2, endingPeriod)

	// rewards for del2 should be 1/4 initial
	require.Equal(t, sdk.DecCoins{{Denom: sdk.DefaultBondDenom, Amount: sdk.NewDec(initial / 4)}}, rewards)

	// commission should be half initial
	require.Equal(t, sdk.DecCoins{{Denom: sdk.DefaultBondDenom, Amount: sdk.NewDec(initial / 2)}}, k.GetValidatorAccumulatedCommission(ctx, valOpAddr1))

	// next block
	ctx = ctx.WithBlockHeight(ctx.BlockHeight() + 1)

	// allocate some more rewards
	k.AllocateTokensToValidator(ctx, val, tokens, sdk.DecCoins{})

	// withdraw commission
	k.WithdrawValidatorCommission(ctx, valOpAddr1)

	// end period
	endingPeriod = k.incrementValidatorPeriod(ctx, val)

	// calculate delegation rewards for del1
	rewards, _ = k.calculateDelegationRewards(ctx, val, del1, endingPeriod)

	// rewards for del1 should be 1/4 initial
	require.Equal(t, sdk.DecCoins{{Denom: sdk.DefaultBondDenom, Amount: sdk.NewDec(initial / 4)}}, rewards)

	// calculate delegation rewards for del2
	rewards, _ = k.calculateDelegationRewards(ctx, val, del2, endingPeriod)

	// rewards for del2 should be 1/2 initial
	require.Equal(t, sdk.DecCoins{{Denom: sdk.DefaultBondDenom, Amount: sdk.NewDec(initial / 2)}}, rewards)

	// commission should be zero
	require.True(t, k.GetValidatorAccumulatedCommission(ctx, valOpAddr1).IsZero())
}

// Test staking keeper hooks and RewardsBank pool.
func TestRewardsBank(t *testing.T) {
	ctx, ak, k, sk, _ := CreateTestInputDefault(t, false, 1000)
	sh := staking.NewHandler(sk)

	checkInvariants := func() {
		msg, broken := RewardsBankPoolInvariant(k)(ctx)
		require.False(t, broken, msg)
	}

	// set module account coins
	distrAcc := k.GetDistributionAccount(ctx)
	distrAcc.SetCoins(sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(1000))))
	k.supplyKeeper.SetModuleAccount(ctx, distrAcc)

	// create validator and endBlock to bond the validator
	{
		commission := staking.NewCommissionRates(sdk.NewDecWithPrec(1, 1), sdk.NewDecWithPrec(1, 1), sdk.NewDec(0))
		msg := staking.NewMsgCreateValidator(valOpAddr1, valConsPk1,
			sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(100)), staking.Description{}, commission, minSelfDelegation)

		res, err := sh(ctx, msg)
		require.NoError(t, err)
		require.NotNil(t, res)

		staking.EndBlocker(ctx, sk)
	}

	val := sk.Validator(ctx, valOpAddr1)
	del := sk.Delegation(ctx, sdk.AccAddress(valOpAddr1), valOpAddr1)

	// next block
	ctx = ctx.WithBlockHeight(ctx.BlockHeight() + 1)

	// allocate some rewards
	{
		tokens := sdk.DecCoins{sdk.NewDecCoinFromDec(sdk.DefaultBondDenom, sdk.NewDec(10))}
		k.AllocateTokensToValidator(ctx, val, tokens, sdk.DecCoins{})
	}

	// pre delegation checks
	var prevTotalRewards sdk.DecCoins
	var prevDistrAccCoins sdk.Coins
	{
		// RewardsBank delegator coins should be empty, as BeforeDelegationSharesModified hook wasn't triggered
		bankCoins := k.GetDelegatorRewardsBankCoins(ctx, del.GetDelegatorAddr(), val.GetOperator())
		require.True(t, bankCoins.Empty())

		// RewardsBankPool should be empty
		rewardsBankAcc := k.GetRewardsBankPoolAccount(ctx)
		require.True(t, rewardsBankAcc.GetCoins().Empty())

		// current total rewards shouldn't be empty
		cacheCtx, _ := ctx.CacheContext()
		endingPeriod := k.incrementValidatorPeriod(cacheCtx, val)
		rewards := k.calculateDelegationTotalRewards(cacheCtx, val, del, endingPeriod)
		require.False(t, rewards.Empty())
		prevTotalRewards = rewards

		prevDistrAccCoins = k.GetDistributionAccount(ctx).GetCoins()
	}

	checkInvariants()

	// delegate more to validator triggering the BeforeDelegationSharesModified hook
	{
		tokens := sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(10))
		msg := staking.NewMsgDelegate(del.GetDelegatorAddr(), val.GetOperator(), tokens)

		res, err := sh(ctx, msg)
		require.NoError(t, err)
		require.NotNil(t, res)

		staking.EndBlocker(ctx, sk)
	}

	// post delegation checks
	{
		// RewardsBank delegator coins should not be empty
		bankCoins := k.GetDelegatorRewardsBankCoins(ctx, del.GetDelegatorAddr(), val.GetOperator())
		require.False(t, bankCoins.Empty())

		// RewardsBankPool acc should not be empty
		rewardsBankAcc := k.GetRewardsBankPoolAccount(ctx)
		require.False(t, rewardsBankAcc.GetCoins().Empty())

		// Distr module acc balance should decrease (part was transferred to RewardsBankPool)
		curDistrAccCoins := k.GetDistributionAccount(ctx).GetCoins()
		require.True(t, curDistrAccCoins.AmountOf(sdk.DefaultBondDenom).LT(prevDistrAccCoins.AmountOf(sdk.DefaultBondDenom)))
		prevDistrAccCoins = curDistrAccCoins

		// current total rewards should not change (as there were no new allocations)
		cacheCtx, _ := ctx.CacheContext()
		endingPeriod := k.incrementValidatorPeriod(cacheCtx, val)
		curTotalRewards := k.calculateDelegationTotalRewards(cacheCtx, val, del, endingPeriod)
		curTotalRewards = k.addAccumulatedBankRewards(ctx, del.GetDelegatorAddr(), val.GetOperator(), curTotalRewards)
		require.True(t, curTotalRewards.IsEqual(prevTotalRewards))

		// current total rewards should be equal to current bankCoins (as current validator delegator rewards are empty)
		require.True(t, curTotalRewards.IsEqual(sdk.NewDecCoinsFromCoins(bankCoins...)))
	}

	checkInvariants()

	// next block
	ctx = ctx.WithBlockHeight(ctx.BlockHeight() + 1)

	// allocate more rewards to raise the main pool rewards
	{
		// update staking vars
		val = sk.Validator(ctx, del.GetValidatorAddr())
		del = sk.Delegation(ctx, del.GetDelegatorAddr(), del.GetValidatorAddr())

		tokens := sdk.DecCoins{sdk.NewDecCoinFromDec(sdk.DefaultBondDenom, sdk.NewDec(10))}
		k.AllocateTokensToValidator(ctx, val, tokens, sdk.DecCoins{})

		// check accumulated delegation bank rewards are not empty
		bankCoins := k.GetDelegatorRewardsBankCoins(ctx, del.GetDelegatorAddr(), val.GetOperator())
		require.False(t, bankCoins.Empty())

		// check current validator delegation rewards are not empty
		cacheCtx, _ := ctx.CacheContext()
		endingPeriod := k.incrementValidatorPeriod(cacheCtx, val)
		curRewards, _ := k.calculateDelegationRewards(cacheCtx, val, del, endingPeriod)
		require.False(t, curRewards.Empty())
	}

	checkInvariants()

	// transfer all delegators rewards
	prevDelBalance := ak.GetAccount(ctx, del.GetDelegatorAddr()).GetCoins()
	{
		// update staking vars
		val = sk.Validator(ctx, del.GetValidatorAddr())
		del = sk.Delegation(ctx, del.GetDelegatorAddr(), del.GetValidatorAddr())

		// transfer
		_, err := k.WithdrawDelegationRewards(ctx, del.GetDelegatorAddr(), val.GetOperator())
		require.NoError(t, err)

		// check delegator balance increased
		curDelBalance := ak.GetAccount(ctx, del.GetDelegatorAddr()).GetCoins()
		require.True(t, curDelBalance.AmountOf(sdk.DefaultBondDenom).GT(prevDelBalance.AmountOf(sdk.DefaultBondDenom)))

		// RewardsBank pool acc should be empty
		rewardsBankAcc := k.GetRewardsBankPoolAccount(ctx)
		require.True(t, rewardsBankAcc.GetCoins().Empty())

		// Distr module acc balance should decrease
		curDistrAccCoins := k.GetDistributionAccount(ctx).GetCoins()
		require.True(t, curDistrAccCoins.AmountOf(sdk.DefaultBondDenom).LT(prevDistrAccCoins.AmountOf(sdk.DefaultBondDenom)))

		// total delegator rewards should be empty
		cacheCtx, _ := ctx.CacheContext()
		endingPeriod := k.incrementValidatorPeriod(cacheCtx, val)
		curRewards, _ := k.calculateDelegationRewards(cacheCtx, val, del, endingPeriod)
		require.True(t, curRewards.Empty())
	}

	checkInvariants()

	// check delegator can't withdraw more
	{
		// update staking vars
		val = sk.Validator(ctx, del.GetValidatorAddr())
		del = sk.Delegation(ctx, del.GetDelegatorAddr(), del.GetValidatorAddr())

		// transfer
		rewards, err := k.WithdrawDelegationRewards(ctx, del.GetDelegatorAddr(), val.GetOperator())
		require.NoError(t, err)
		require.True(t, rewards.IsZero())
	}
}

// Test RewardsBank operations with multiple delegations.
// nolint: staticcheck
func TestRewardsBankMulti(t *testing.T) {
	ctx, ak, k, sk, _ := CreateTestInputDefault(t, false, 1000)
	sh := staking.NewHandler(sk)

	checkInvariants := func() {
		msg, broken := RewardsBankPoolInvariant(k)(ctx)
		require.False(t, broken, msg)
	}

	getCurrentRewards := func(del staking.DelegationI, val staking.ValidatorI) sdk.DecCoins {
		cacheCtx, _ := ctx.CacheContext()
		endingPeriod := k.incrementValidatorPeriod(cacheCtx, val)
		rewards := k.calculateDelegationTotalRewards(cacheCtx, val, del, endingPeriod)
		return rewards
	}

	updateVal := func(val staking.ValidatorI) staking.ValidatorI {
		return sk.Validator(ctx, val.GetOperator())
	}

	updateDel := func(val staking.ValidatorI, del staking.DelegationI) staking.DelegationI {
		return sk.Delegation(ctx, del.GetDelegatorAddr(), val.GetOperator())
	}

	// set module account coins
	distrAcc := k.GetDistributionAccount(ctx)
	distrAcc.SetCoins(sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(1000))))
	k.supplyKeeper.SetModuleAccount(ctx, distrAcc)

	// create validators and endBlock to bond the validator
	{
		commission := staking.NewCommissionRates(sdk.NewDecWithPrec(1, 1), sdk.NewDecWithPrec(1, 1), sdk.NewDec(0))

		msg := staking.NewMsgCreateValidator(valOpAddr1, valConsPk1,
			sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(100)), staking.Description{}, commission, minSelfDelegation)
		res, err := sh(ctx, msg)
		require.NoError(t, err)
		require.NotNil(t, res)

		msg = staking.NewMsgCreateValidator(valOpAddr2, valConsPk2,
			sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(100)), staking.Description{}, commission, minSelfDelegation)
		res, err = sh(ctx, msg)
		require.NoError(t, err)
		require.NotNil(t, res)

		staking.EndBlocker(ctx, sk)
	}

	val1, val2 := sk.Validator(ctx, valOpAddr1), sk.Validator(ctx, valOpAddr2)
	del1, del2 := sk.Delegation(ctx, sdk.AccAddress(valOpAddr1), valOpAddr1), sk.Delegation(ctx, sdk.AccAddress(valOpAddr2), valOpAddr2)

	updateAll := func() {
		ctx = ctx.WithBlockHeight(ctx.BlockHeight() + 1)
		val1, val2 = updateVal(val1), updateVal(val2)
		del1, del2 = updateDel(val1, del1), updateDel(val2, del2)
	}

	// allocate some rewards to both validators (different amount)
	{
		tokens := sdk.DecCoins{sdk.NewDecCoinFromDec(sdk.DefaultBondDenom, sdk.NewDec(10))}
		k.AllocateTokensToValidator(ctx, val1, tokens, sdk.DecCoins{})

		tokens = sdk.DecCoins{sdk.NewDecCoinFromDec(sdk.DefaultBondDenom, sdk.NewDec(20))}
		k.AllocateTokensToValidator(ctx, val2, tokens, sdk.DecCoins{})

		updateAll()
	}

	// pre delegation modifications checks
	var prevDel1CurRewards, prevDel2CurRewards sdk.DecCoins
	prevDel1BankRewards, prevDel2BankRewards := sdk.Coins{}, sdk.Coins{}
	{
		// check rewards bank is empty
		{
			require.True(t, k.GetDelegatorRewardsBankCoins(ctx, del1.GetDelegatorAddr(), val1.GetOperator()).Empty())
			require.True(t, k.GetDelegatorRewardsBankCoins(ctx, del1.GetDelegatorAddr(), val2.GetOperator()).Empty())
			require.True(t, k.GetDelegatorRewardsBankCoins(ctx, del2.GetDelegatorAddr(), val2.GetOperator()).Empty())
			require.True(t, k.GetDelegatorRewardsBankCoins(ctx, del2.GetDelegatorAddr(), val1.GetOperator()).Empty())

			require.True(t, k.GetRewardsBankPoolAccount(ctx).GetCoins().Empty())
		}

		// current total rewards shouldn't be empty
		prevDel1CurRewards = getCurrentRewards(del1, val1)
		require.False(t, prevDel1CurRewards.Empty())

		prevDel2CurRewards = getCurrentRewards(del2, val2)
		require.False(t, prevDel2CurRewards.Empty())

		// del2 should have more rewards than del1
		require.True(t, prevDel2CurRewards.AmountOf(sdk.DefaultBondDenom).GT(prevDel1CurRewards.AmountOf(sdk.DefaultBondDenom)))
	}
	checkInvariants()

	// increase del1 shares
	{
		tokens := sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(10))
		msg := staking.NewMsgDelegate(del1.GetDelegatorAddr(), val1.GetOperator(), tokens)

		res, err := sh(ctx, msg)
		require.NoError(t, err)
		require.NotNil(t, res)

		staking.EndBlocker(ctx, sk)

		updateAll()
	}
	// checks
	{
		// del1
		{
			// bank should not be empty
			bankRewards := k.GetDelegatorRewardsBankCoins(ctx, del1.GetDelegatorAddr(), val1.GetOperator())
			require.False(t, bankRewards.IsZero())
			require.True(t, k.GetDelegatorRewardsBankCoins(ctx, del1.GetDelegatorAddr(), val2.GetOperator()).Empty())
			prevDel1BankRewards = bankRewards

			// bank rewards amount should be equal to prev current rewards
			require.True(t, bankRewards.AmountOf(sdk.DefaultBondDenom).Equal(prevDel1CurRewards.AmountOf(sdk.DefaultBondDenom).TruncateInt()))

			// current rewards should be empty
			curRewards := getCurrentRewards(del1, val1)
			require.True(t, curRewards.IsZero())
			prevDel1CurRewards = curRewards
		}

		// del2
		{
			// bank should be empty
			require.True(t, k.GetDelegatorRewardsBankCoins(ctx, del2.GetDelegatorAddr(), val2.GetOperator()).Empty())

			// current rewards should not be empty
			require.False(t, getCurrentRewards(del2, val2).IsZero())
		}
	}
	checkInvariants()

	// decrease del2 shares
	{
		udTokens := sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(50))
		msg := staking.NewMsgUndelegate(del2.GetDelegatorAddr(), val2.GetOperator(), udTokens)

		res, err := sh(ctx, msg)
		require.NoError(t, err)
		require.NotNil(t, res)

		staking.EndBlocker(ctx, sk)

		updateAll()
	}
	// check
	{
		// del1
		{
			// state should no change
			bankRewards := k.GetDelegatorRewardsBankCoins(ctx, del1.GetDelegatorAddr(), val1.GetOperator())
			require.True(t, bankRewards.IsEqual(prevDel1BankRewards))
			require.True(t, k.GetDelegatorRewardsBankCoins(ctx, del1.GetDelegatorAddr(), val2.GetOperator()).Empty())

			curRewards := getCurrentRewards(del1, val1)
			require.True(t, curRewards.IsEqual(prevDel1CurRewards))
		}

		// del2
		{
			// bank should not be empty
			bankRewards := k.GetDelegatorRewardsBankCoins(ctx, del2.GetDelegatorAddr(), val2.GetOperator())
			require.False(t, bankRewards.IsZero())
			require.True(t, k.GetDelegatorRewardsBankCoins(ctx, del2.GetDelegatorAddr(), val1.GetOperator()).Empty())
			prevDel2BankRewards = bankRewards

			// bank rewards amount should be equal to prev current rewards
			require.True(t, bankRewards.AmountOf(sdk.DefaultBondDenom).Equal(prevDel2CurRewards.AmountOf(sdk.DefaultBondDenom).TruncateInt()))

			// current rewards should be empty
			curRewards := getCurrentRewards(del2, val2)
			require.True(t, curRewards.IsZero())
			prevDel2CurRewards = curRewards
		}

		require.True(t, prevDel2BankRewards.IsAllGT(prevDel1BankRewards))
	}
	checkInvariants()

	// allocate some rewards to both validators
	{
		tokens := sdk.DecCoins{sdk.NewDecCoinFromDec(sdk.DefaultBondDenom, sdk.NewDec(10))}

		k.AllocateTokensToValidator(ctx, val1, tokens, sdk.DecCoins{})
		k.AllocateTokensToValidator(ctx, val2, tokens, sdk.DecCoins{})

		updateAll()
	}
	// check
	{
		curRewards := getCurrentRewards(del1, val1)
		require.False(t, curRewards.IsZero())
		prevDel1CurRewards = curRewards

		curRewards = getCurrentRewards(del2, val2)
		require.False(t, curRewards.IsZero())
		prevDel2CurRewards = curRewards
	}
	checkInvariants()

	// withdraw all
	del1InitialBalance, del2InitialBalance := ak.GetAccount(ctx, del1.GetDelegatorAddr()).GetCoins(), ak.GetAccount(ctx, del2.GetDelegatorAddr()).GetCoins()
	var del1WithdrawAmt, del2WithdrawAmt sdk.Int
	{
		coins, err := k.WithdrawDelegationRewards(ctx, del1.GetDelegatorAddr(), val1.GetOperator())
		require.NoError(t, err)
		del1WithdrawAmt = coins.AmountOf(sdk.DefaultBondDenom)

		coins, err = k.WithdrawDelegationRewards(ctx, del2.GetDelegatorAddr(), val2.GetOperator())
		require.NoError(t, err)
		del2WithdrawAmt = coins.AmountOf(sdk.DefaultBondDenom)

		updateAll()
	}
	// check
	{
		// banks should be empty
		require.True(t, k.GetDelegatorRewardsBankCoins(ctx, del1.GetDelegatorAddr(), val1.GetOperator()).Empty())
		require.True(t, k.GetDelegatorRewardsBankCoins(ctx, del2.GetDelegatorAddr(), val2.GetOperator()).Empty())
		require.True(t, k.GetDelegatorRewardsBankCoins(ctx, del1.GetDelegatorAddr(), val2.GetOperator()).Empty())
		require.True(t, k.GetDelegatorRewardsBankCoins(ctx, del2.GetDelegatorAddr(), val1.GetOperator()).Empty())

		// current rewards should be empty
		require.True(t, getCurrentRewards(del1, val1).IsZero())
		require.True(t, getCurrentRewards(del2, val2).IsZero())

		del1CurrentBalance, del2CurrentBalance := ak.GetAccount(ctx, del1.GetDelegatorAddr()).GetCoins(), ak.GetAccount(ctx, del2.GetDelegatorAddr()).GetCoins()
		del1DiffActual, del2DiffActual := del1CurrentBalance.Sub(del1InitialBalance).AmountOf(sdk.DefaultBondDenom), del2CurrentBalance.Sub(del2InitialBalance).AmountOf(sdk.DefaultBondDenom)

		del1DiffExpected := prevDel1CurRewards.AmountOf(sdk.DefaultBondDenom).TruncateInt().Add(prevDel1BankRewards.AmountOf(sdk.DefaultBondDenom))
		require.True(t, del1WithdrawAmt.Equal(del1DiffActual))
		require.True(t, del1DiffExpected.Equal(del1DiffActual))

		del2DiffExpected := prevDel2CurRewards.AmountOf(sdk.DefaultBondDenom).TruncateInt().Add(prevDel2BankRewards.AmountOf(sdk.DefaultBondDenom))
		require.True(t, del2WithdrawAmt.Equal(del2DiffActual))
		require.True(t, del2DiffExpected.Equal(del2DiffActual))
	}
	checkInvariants()
}

// Test rewards withdraw when delegation is reduced to zero.
func TestRewardsBankUndelegatingToZero(t *testing.T) {
	ctx, _, k, sk, _ := CreateTestInputDefault(t, false, 1000)
	sh := staking.NewHandler(sk)

	// set module account coins
	distrAcc := k.GetDistributionAccount(ctx)
	distrAcc.SetCoins(sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(1000))))
	k.supplyKeeper.SetModuleAccount(ctx, distrAcc)

	// create validator and endBlock to bond the validator
	{
		commission := staking.NewCommissionRates(sdk.NewDecWithPrec(1, 1), sdk.NewDecWithPrec(1, 1), sdk.NewDec(0))
		msg := staking.NewMsgCreateValidator(valOpAddr1, valConsPk1,
			sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(100)), staking.Description{}, commission, minSelfDelegation)

		res, err := sh(ctx, msg)
		require.NoError(t, err)
		require.NotNil(t, res)

		staking.EndBlocker(ctx, sk)
	}

	val := sk.Validator(ctx, valOpAddr1)
	delAddr := delAddr1

	// next block
	ctx = ctx.WithBlockHeight(ctx.BlockHeight() + 1)

	// delegate tokens
	delTokens := sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(10))
	{
		msg := staking.NewMsgDelegate(delAddr, val.GetOperator(), delTokens)

		res, err := sh(ctx, msg)
		require.NoError(t, err)
		require.NotNil(t, res)
		staking.EndBlocker(ctx, sk)

		del := sk.Delegation(ctx, delAddr, val.GetOperator())
		require.NotNil(t, del)

		val = sk.Validator(ctx, valOpAddr1)
		ctx = ctx.WithBlockHeight(ctx.BlockHeight() + 1)
	}

	// allocate some rewards
	{
		tokens := sdk.DecCoins{sdk.NewDecCoinFromDec(sdk.DefaultBondDenom, sdk.NewDec(10))}
		k.AllocateTokensToValidator(ctx, val, tokens, sdk.DecCoins{})

		val = sk.Validator(ctx, valOpAddr1)
		ctx = ctx.WithBlockHeight(ctx.BlockHeight() + 1)
	}

	// undelegate twice (till zero)
	{
		for i := 0; i < 2; i++ {
			udTokens := delTokens
			udTokens.Amount = udTokens.Amount.QuoRaw(2)

			msg := staking.NewMsgUndelegate(delAddr, val.GetOperator(), udTokens)

			res, err := sh(ctx, msg)
			require.NoError(t, err)
			require.NotNil(t, res)
			staking.EndBlocker(ctx, sk)

			val = sk.Validator(ctx, valOpAddr1)
			ctx = ctx.WithBlockHeight(ctx.BlockHeight() + 1)
		}
	}

	// transfer all rewards
	{
		rewards, err := k.WithdrawDelegationRewards(ctx, delAddr, val.GetOperator())
		require.NoError(t, err)

		// rewards should be zero, as current rewards were to small to truncate to a single Coin
		require.True(t, rewards.IsZero())
	}
}

// Test rewards distribution with LP stakes.
func TestLPRewardsWithLock(t *testing.T) {
	initPower := int64(1000)
	ctx, ak, k, sk, spk := CreateTestInputDefault(t, false, initPower)
	sh := staking.NewHandler(sk)

	allocateRewards := func(amtPower int64, distrPowerCmp, lpPowerCmp int) {
		// get validator 1 params
		val1, found := sk.GetValidator(ctx, valOpAddr1)
		require.True(t, found)
		distrPower1, lpPower1 := k.GetDistributionPower(ctx, val1.GetOperator(), val1.ConsensusPower(), val1.LPPower(), sk.LPDistrRatio(ctx))

		// get validator 2 params
		val2, found := sk.GetValidator(ctx, valOpAddr2)
		require.True(t, found)
		distrPower2, lpPower2 := k.GetDistributionPower(ctx, val2.GetOperator(), val2.ConsensusPower(), val2.LPPower(), sk.LPDistrRatio(ctx))

		switch distrPowerCmp {
		case -1:
			require.Greater(t, distrPower1, distrPower2)
		case 0:
			require.Equal(t, distrPower1, distrPower2)
		case 1:
			require.Greater(t, distrPower2, distrPower1)
		}

		switch lpPowerCmp {
		case -1:
			require.Greater(t, lpPower1, lpPower2)
		case 0:
			require.Equal(t, lpPower1, lpPower2)
		case 1:
			require.Greater(t, lpPower2, lpPower1)
		}

		// build votes
		abciVotes := types.ABCIVotes{
			{
				Validator:         val1,
				DistributionPower: distrPower1,
				LPPower:           lpPower1,
				SignedLastBlock:   false,
			},
			{
				Validator:         val2,
				DistributionPower: distrPower2,
				LPPower:           lpPower2,
				SignedLastBlock:   false,
			},
		}

		// set minted amount
		coin := sdk.NewCoin(sdk.DefaultBondDenom, sdk.TokensFromConsensusPower(amtPower))
		feeCollector := spk.GetModuleAccount(ctx, k.feeCollectorName)
		err := feeCollector.SetCoins(sdk.NewCoins(coin))
		require.NoError(t, err)
		spk.SetModuleAccount(ctx, feeCollector)

		// allocate
		k.AllocateTokens(
			ctx,
			distrPower1, lpPower1,
			distrPower1+distrPower2,
			lpPower1+lpPower2,
			val1.GetConsAddr(),
			abciVotes,
			sdk.ZeroDec(),
		)
	}

	getRewards := func(delAddr sdk.AccAddress, valAddr sdk.ValAddress) sdk.DecCoins {
		ctxCache, _ := ctx.CacheContext()

		val := sk.Validator(ctxCache, valAddr)
		require.NotNil(t, val)

		del := sk.Delegation(ctxCache, delAddr, valAddr)
		require.NotNil(t, del)

		endingPeriod := k.incrementValidatorPeriod(ctxCache, val)
		rewards := k.calculateDelegationTotalRewards(ctxCache, val, del, endingPeriod)
		rewards = k.addAccumulatedBankRewards(ctxCache, del.GetDelegatorAddr(), val.GetOperator(), rewards)
		if rewards == nil {
			return sdk.DecCoins{}
		}

		return rewards
	}

	endBlock := func() {
		staking.EndBlocker(ctx, sk)
		ctx = ctx.WithBlockHeight(ctx.BlockHeight() + 1)
	}

	checkInvariants := func() {
		msg, broken := AllInvariants(k)(ctx)
		require.False(t, broken, msg)
	}

	// set proposer rewards to zero (for equal rewards distribution)
	params := k.GetParams(ctx)
	params.BaseProposerReward = sdk.ZeroDec()
	params.BonusProposerReward = sdk.ZeroDec()
	k.SetParams(ctx, params)

	// add LPs to delegators
	{
		initTokens := sdk.TokensFromConsensusPower(initPower)
		coin := sdk.NewCoin(sk.LPDenom(ctx), initTokens)

		del1 := ak.GetAccount(ctx, delAddr1)
		err := del1.SetCoins(del1.GetCoins().Add(coin))
		require.NoError(t, err)
		ak.SetAccount(ctx, del1)

		del2 := ak.GetAccount(ctx, delAddr2)
		err = del2.SetCoins(del2.GetCoins().Add(coin))
		require.NoError(t, err)
		ak.SetAccount(ctx, del2)
	}

	// create validators
	{
		commission := staking.NewCommissionRates(sdk.NewDecWithPrec(1, 1), sdk.NewDecWithPrec(1, 1), sdk.NewDec(0))
		selfDelCoin := sdk.NewCoin(sdk.DefaultBondDenom, sdk.TokensFromConsensusPower(100))

		msg := staking.NewMsgCreateValidator(valOpAddr1, valConsPk1, selfDelCoin, staking.Description{}, commission, minSelfDelegation)
		res, err := sh(ctx, msg)
		require.NoError(t, err)
		require.NotNil(t, res)

		msg = staking.NewMsgCreateValidator(valOpAddr2, valConsPk2, selfDelCoin, staking.Description{}, commission, minSelfDelegation)
		res, err = sh(ctx, msg)
		require.NoError(t, err)
		require.NotNil(t, res)

		endBlock()
	}
	checkInvariants()

	// delegate bonding tokens by two delegator to both validators
	{
		delTokens := sdk.NewCoin(sdk.DefaultBondDenom, sdk.TokensFromConsensusPower(10))

		msg := staking.NewMsgDelegate(delAddr1, valOpAddr1, delTokens)
		res, err := sh(ctx, msg)
		require.NoError(t, err)
		require.NotNil(t, res)

		msg = staking.NewMsgDelegate(delAddr2, valOpAddr2, delTokens)
		res, err = sh(ctx, msg)
		require.NoError(t, err)
		require.NotNil(t, res)

		endBlock()
	}
	checkInvariants()

	// distribute rewards (LPPool shouldn't be distributed)
	{
		// distr and LP powers should be equal
		allocateRewards(10, 0, 0)
		endBlock()

		// check LPPool wasn't distributed
		pools := k.GetRewardPools(ctx)
		require.True(t, pools.LiquidityProvidersPool.IsAllPositive())

		// check delegators have rewards
		del1Rewards := getRewards(delAddr1, valOpAddr1)
		require.True(t, del1Rewards.IsAllPositive())

		del2Rewards := getRewards(delAddr2, valOpAddr2)
		require.True(t, del2Rewards.IsAllPositive())

		// ...and they are equal, as all shares are the same
		require.True(t, del1Rewards.IsEqual(del2Rewards), "%s / %s", del1Rewards.String(), del2Rewards.String())
	}
	checkInvariants()

	// delegate LP tokens by two delegator to both validators
	{
		delTokens := sdk.NewCoin(sdk.DefaultLiquidityDenom, sdk.TokensFromConsensusPower(10))

		msg := staking.NewMsgDelegate(delAddr1, valOpAddr1, delTokens)
		res, err := sh(ctx, msg)
		require.NoError(t, err)
		require.NotNil(t, res)

		msg = staking.NewMsgDelegate(delAddr2, valOpAddr2, delTokens)
		res, err = sh(ctx, msg)
		require.NoError(t, err)
		require.NotNil(t, res)

		endBlock()
	}
	checkInvariants()

	// distribute rewards (LPPool should be distributed)
	{
		// distr and LP powers should be equal
		allocateRewards(10, 0, 0)
		endBlock()

		// check LPPool was distributed
		pools := k.GetRewardPools(ctx)
		require.True(t, pools.LiquidityProvidersPool.IsZero())

		// check delegators have rewards
		del1Rewards := getRewards(delAddr1, valOpAddr1)
		require.True(t, del1Rewards.IsAllPositive())

		del2Rewards := getRewards(delAddr2, valOpAddr2)
		require.True(t, del2Rewards.IsAllPositive())

		// ...and they are equal, as all shares are the same
		require.True(t, del1Rewards.IsEqual(del2Rewards), "%s / %s", del1Rewards.String(), del2Rewards.String())
	}
	checkInvariants()

	// rebalance distribution by raising delegator 1 LP shares (delegator 1 is a winner now)
	{
		delTokens := sdk.NewCoin(sdk.DefaultLiquidityDenom, sdk.TokensFromConsensusPower(10))

		msg := staking.NewMsgDelegate(delAddr1, valOpAddr1, delTokens)
		res, err := sh(ctx, msg)
		require.NoError(t, err)
		require.NotNil(t, res)

		endBlock()
	}
	checkInvariants()

	// distribute rewards and check rewards are not the same
	{
		// as validator 1 has more LPs now, its power should be higher
		// the same for LP power
		// distr powers should be the same, but validator 1 has more LPs now
		allocateRewards(10, -1, -1)
		endBlock()

		// check LPPool was distributed
		pools := k.GetRewardPools(ctx)
		require.True(t, pools.LiquidityProvidersPool.IsZero())

		// check delegators have rewards
		del1Rewards := getRewards(delAddr1, valOpAddr1)
		require.True(t, del1Rewards.IsAllPositive())

		del2Rewards := getRewards(delAddr2, valOpAddr2)
		require.True(t, del2Rewards.IsAllPositive())

		require.True(t, del1Rewards.AmountOf(sdk.DefaultBondDenom).GT(del2Rewards.AmountOf(sdk.DefaultBondDenom)), "%s / %s", del1Rewards.String(), del2Rewards.String())
	}
	checkInvariants()

	// rebalance distribution by locking validator 2 rewards
	var unlockTime time.Time
	{
		ut, err := k.LockValidatorRewards(ctx, valOpAddr2)
		require.NoError(t, err)
		unlockTime = ut
	}
	checkInvariants()

	// distribute rewards and check rewards are not the same
	{
		// due to locking validator 2 distr power should be higher
		// locking should increase validator 2 power, but validator 1 should be a winner still
		allocateRewards(10, 1, -1)
		endBlock()

		// check LPPool was distributed
		pools := k.GetRewardPools(ctx)
		require.True(t, pools.LiquidityProvidersPool.IsZero())

		// check delegators have rewards
		del1Rewards := getRewards(delAddr1, valOpAddr1)
		require.True(t, del1Rewards.IsAllPositive())

		del2Rewards := getRewards(delAddr2, valOpAddr2)
		require.True(t, del2Rewards.IsAllPositive())
	}
	checkInvariants()

	// withdraw delegator 1 rewards
	{
		coins, err := k.WithdrawDelegationRewards(ctx, delAddr1, valOpAddr1)
		require.NoError(t, err)
		require.True(t, coins.IsAllPositive())

		// withdraw more (no rewards left)
		coins, err = k.WithdrawDelegationRewards(ctx, delAddr1, valOpAddr1)
		require.NoError(t, err)
		require.True(t, coins.IsZero())
	}
	checkInvariants()

	// withdraw delegator 2 rewards
	{
		_, err := k.WithdrawDelegationRewards(ctx, delAddr2, valOpAddr2)
		require.Error(t, err)

		// disable auto-lock
		err = k.DisableLockedRewardsAutoRenewal(ctx, valOpAddr2)
		require.NoError(t, err)

		// emulate lock stop
		ctx = ctx.WithBlockTime(unlockTime)
		k.ProcessAllMatureRewardsUnlockQueueItems(ctx)

		coins, err := k.WithdrawDelegationRewards(ctx, delAddr2, valOpAddr2)
		require.NoError(t, err)
		require.True(t, coins.IsAllPositive())

		// withdraw more (no rewards left)
		coins, err = k.WithdrawDelegationRewards(ctx, delAddr2, valOpAddr2)
		require.NoError(t, err)
		require.True(t, coins.IsZero())
	}
	checkInvariants()
}

// Test query and withdraw rewards that got stucked within validator where delegator has no longer delegations to.
func TestWithdrawDelegationRewardsMissingVal(t *testing.T) {
	ctx, _, k, sk, _ := CreateTestInputDefault(t, false, 1000)
	sh := staking.NewHandler(sk)

	// set module account coins
	distrAcc := k.GetDistributionAccount(ctx)
	distrAcc.SetCoins(sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(1000))))
	k.supplyKeeper.SetModuleAccount(ctx, distrAcc)

	updateVal := func(val staking.ValidatorI) staking.ValidatorI {
		return sk.Validator(ctx, val.GetOperator())
	}

	getBankRewards := func(delAdd sdk.AccAddress, val staking.ValidatorI) sdk.Coins {
		return k.GetDelegatorRewardsBankCoins(ctx, delAddr1, val.GetOperator())
	}

	getCurrentRewards := func(delAddr sdk.AccAddress, val staking.ValidatorI) sdk.DecCoins {
		cacheCtx, _ := ctx.CacheContext()
		del := sk.Delegation(ctx, delAddr, val.GetOperator())
		endingPeriod := k.incrementValidatorPeriod(cacheCtx, val)
		rewards := k.calculateDelegationTotalRewards(cacheCtx, val, del, endingPeriod)
		return rewards
	}

	queryRewards := func(delAddr sdk.AccAddress, val staking.ValidatorI) (result types.QueryDelegationRewardsResponse) {
		queryParams := types.QueryDelegationRewardsParams{
			DelegatorAddress: delAddr,
			ValidatorAddress: val.GetOperator(),
		}

		res, err := queryDelegationRewards(ctx, nil, abci.RequestQuery{Data: k.cdc.MustMarshalJSON(queryParams)}, k)
		require.NoError(t, err)

		k.cdc.MustUnmarshalJSON(res, &result)
		return
	}

	queryTotalRewards := func(delAddr sdk.AccAddress) (result types.QueryDelegatorTotalRewardsResponse) {
		queryParams := types.QueryDelegatorParams{
			DelegatorAddress: delAddr,
		}

		res, err := queryDelegatorTotalRewards(ctx, nil, abci.RequestQuery{Data: k.cdc.MustMarshalJSON(queryParams)}, k)
		require.NoError(t, err)

		k.cdc.MustUnmarshalJSON(res, &result)
		return
	}

	nextBlock := func() {
		ctx = ctx.WithBlockHeight(ctx.BlockHeight() + 1)
	}

	// create validators and endBlock to bond the validator
	{
		commission := staking.NewCommissionRates(sdk.NewDecWithPrec(1, 1), sdk.NewDecWithPrec(1, 1), sdk.NewDec(0))

		msg := staking.NewMsgCreateValidator(valOpAddr1, valConsPk1,
			sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(100)), staking.Description{}, commission, minSelfDelegation)
		res, err := sh(ctx, msg)
		require.NoError(t, err)
		require.NotNil(t, res)

		msg = staking.NewMsgCreateValidator(valOpAddr2, valConsPk2,
			sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(100)), staking.Description{}, commission, minSelfDelegation)
		res, err = sh(ctx, msg)
		require.NoError(t, err)
		require.NotNil(t, res)

		staking.EndBlocker(ctx, sk)
	}

	val1, val2 := sk.Validator(ctx, valOpAddr1), sk.Validator(ctx, valOpAddr2)
	delAddr := delAddr1

	// delegate to val1 and allocate some rewards
	{
		delTokens := sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(100))
		msg := staking.NewMsgDelegate(delAddr, val1.GetOperator(), delTokens)

		res, err := sh(ctx, msg)
		require.NoError(t, err)
		require.NotNil(t, res)
		staking.EndBlocker(ctx, sk)
		val1 = updateVal(val1)

		allocTokens := sdk.DecCoins{sdk.NewDecCoinFromDec(sdk.DefaultBondDenom, sdk.NewDec(100))}
		k.AllocateTokensToValidator(ctx, val1, allocTokens, sdk.DecCoins{})
		val1 = updateVal(val1)

		nextBlock()
	}
	// check rewards state
	{
		// bank rewards are empty
		require.True(t, getBankRewards(delAddr1, val1).IsZero())
		// current rewards aren't empty
		require.False(t, getCurrentRewards(delAddr1, val1).IsZero())
	}

	// modify the delegation to val1 (that fills the rewards bank)
	// allocate some rewards (that fills the current rewards)
	{
		tokens := sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(100))
		msg := staking.NewMsgDelegate(delAddr, val1.GetOperator(), tokens)

		res, err := sh(ctx, msg)
		require.NoError(t, err)
		require.NotNil(t, res)
		staking.EndBlocker(ctx, sk)
		val1 = updateVal(val1)

		allocTokens := sdk.DecCoins{sdk.NewDecCoinFromDec(sdk.DefaultBondDenom, sdk.NewDec(100))}
		k.AllocateTokensToValidator(ctx, val1, allocTokens, sdk.DecCoins{})
		val1 = updateVal(val1)

		nextBlock()
	}
	// check rewards state
	{
		// bank has rewards
		require.False(t, getBankRewards(delAddr1, val1).IsZero())
		// current rewards aren't empty
		require.False(t, getCurrentRewards(delAddr1, val1).IsZero())
	}

	// redelegate all del tokens from val1 to val2
	{
		rdTokens := sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(200))
		msg := staking.NewMsgBeginRedelegate(delAddr, val1.GetOperator(), val2.GetOperator(), rdTokens)
		res, err := sh(ctx, msg)
		require.NoError(t, err)
		require.NotNil(t, res)
		staking.EndBlocker(ctx, sk)
		val1 = updateVal(val1)
		val2 = updateVal(val2)

		nextBlock()
	}
	// check rewards state
	{
		// bank for val1 has rewards
		require.False(t, getBankRewards(delAddr1, val1).IsZero())
		// bank for val2 has no rewards yet
		require.True(t, getBankRewards(delAddr1, val2).IsZero())
	}

	// check queryDelegationRewards
	{
		// delegator has no delegation for val1, but Total shouldn't be empty
		val1Rewards := queryRewards(delAddr1, val1)
		require.True(t, val1Rewards.Reward.Current.IsZero())
		require.False(t, val1Rewards.Reward.Total.IsZero())

		// delegator has delegation for val2, but current and zero should be empty
		val2Rewards := queryRewards(delAddr1, val2)
		require.True(t, val2Rewards.Reward.Current.IsZero())
		require.True(t, val2Rewards.Reward.Total.IsZero())
	}

	// check queryDelegatorTotalRewards
	{
		delRewards := queryTotalRewards(delAddr1)
		require.Len(t, delRewards.Rewards, 2)
		require.True(t, delRewards.Total.IsEqual(delRewards.Rewards[0].Total.Add(delRewards.Rewards[1].Total...)))

		// existing delegations should go first
		require.True(t, delRewards.Rewards[0].ValidatorAddress.Equals(val2.GetOperator()))
		require.True(t, delRewards.Rewards[0].Current.IsZero())
		require.True(t, delRewards.Rewards[0].Total.IsZero())

		// non-existing delegations should go last
		require.True(t, delRewards.Rewards[1].ValidatorAddress.Equals(val1.GetOperator()))
		require.True(t, delRewards.Rewards[1].Current.IsZero())
		require.False(t, delRewards.Rewards[1].Total.IsZero())
	}

	// check withdraw rewards from val1 missing delegation
	{
		rewards, err := k.WithdrawDelegationRewards(ctx, delAddr, val1.GetOperator())
		require.NoError(t, err)
		require.False(t, rewards.IsZero())
	}

	// check withdraw rewards from val2 (no rewards)
	{
		rewards, err := k.WithdrawDelegationRewards(ctx, delAddr, val2.GetOperator())
		require.NoError(t, err)
		require.True(t, rewards.IsZero())
	}
}

func TestWithdrawWithForceUnbondedValidator(t *testing.T) {
	ctx, _, k, sk, _ := CreateTestInputDefault(t, false, 1000)
	sh := staking.NewHandler(sk)

	// set module account coins
	distrAcc := k.GetDistributionAccount(ctx)
	distrAcc.SetCoins(sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(1000))))
	k.supplyKeeper.SetModuleAccount(ctx, distrAcc)

	updateVal := func(val staking.ValidatorI) staking.ValidatorI {
		return sk.Validator(ctx, val.GetOperator())
	}

	getBankRewards := func(delAdd sdk.AccAddress, val staking.ValidatorI) sdk.Coins {
		return k.GetDelegatorRewardsBankCoins(ctx, delAddr1, val.GetOperator())
	}

	getCurrentRewards := func(delAddr sdk.AccAddress, val staking.ValidatorI) sdk.DecCoins {
		cacheCtx, _ := ctx.CacheContext()
		del := sk.Delegation(ctx, delAddr, val.GetOperator())
		endingPeriod := k.incrementValidatorPeriod(cacheCtx, val)
		rewards := k.calculateDelegationTotalRewards(cacheCtx, val, del, endingPeriod)
		return rewards
	}

	queryRewards := func(delAddr sdk.AccAddress, valAddr sdk.ValAddress) (result types.QueryDelegationRewardsResponse) {
		queryParams := types.QueryDelegationRewardsParams{
			DelegatorAddress: delAddr,
			ValidatorAddress: valAddr,
		}

		res, err := queryDelegationRewards(ctx, nil, abci.RequestQuery{Data: k.cdc.MustMarshalJSON(queryParams)}, k)
		require.NoError(t, err)

		k.cdc.MustUnmarshalJSON(res, &result)
		return
	}

	queryTotalRewards := func(delAddr sdk.AccAddress) (result types.QueryDelegatorTotalRewardsResponse) {
		queryParams := types.QueryDelegatorParams{
			DelegatorAddress: delAddr,
		}

		res, err := queryDelegatorTotalRewards(ctx, nil, abci.RequestQuery{Data: k.cdc.MustMarshalJSON(queryParams)}, k)
		require.NoError(t, err)

		k.cdc.MustUnmarshalJSON(res, &result)
		return
	}

	nextBlock := func() {
		ctx = ctx.WithBlockHeight(ctx.BlockHeight() + 1)
	}

	// create validators and endBlock to bond the validator
	{
		// overwrite PowerReduction as otherwise validator won't be bonded
		sdk.PowerReduction = sdk.NewInt(1)

		commission := staking.NewCommissionRates(sdk.NewDecWithPrec(1, 1), sdk.NewDecWithPrec(1, 1), sdk.NewDec(0))

		msg := staking.NewMsgCreateValidator(valOpAddr1, valConsPk1,
			sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(10)), staking.Description{}, commission, minSelfDelegation)
		res, err := sh(ctx, msg)
		require.NoError(t, err)
		require.NotNil(t, res)
		staking.EndBlocker(ctx, sk)
	}

	val := sk.Validator(ctx, valOpAddr1)
	delAddr := delAddr1

	// delegate and allocate some rewards
	{
		delTokens := sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(45))
		msg := staking.NewMsgDelegate(delAddr, val.GetOperator(), delTokens)

		res, err := sh(ctx, msg)
		require.NoError(t, err)
		require.NotNil(t, res)
		staking.EndBlocker(ctx, sk)
		val = updateVal(val)

		allocTokens := sdk.DecCoins{sdk.NewDecCoinFromDec(sdk.DefaultBondDenom, sdk.NewDec(100))}
		k.AllocateTokensToValidator(ctx, val, allocTokens, sdk.DecCoins{})
		val = updateVal(val)

		nextBlock()
	}
	// check rewards state
	{
		require.True(t, getBankRewards(delAddr1, val).IsZero())
		require.False(t, getCurrentRewards(delAddr1, val).IsZero())
	}

	// modify the delegation and allocate some rewards
	{
		tokens := sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(45))
		msg := staking.NewMsgDelegate(delAddr, val.GetOperator(), tokens)

		res, err := sh(ctx, msg)
		require.NoError(t, err)
		require.NotNil(t, res)
		staking.EndBlocker(ctx, sk)
		val = updateVal(val)

		allocTokens := sdk.DecCoins{sdk.NewDecCoinFromDec(sdk.DefaultBondDenom, sdk.NewDec(100))}
		k.AllocateTokensToValidator(ctx, val, allocTokens, sdk.DecCoins{})
		val = updateVal(val)

		nextBlock()
	}
	// check rewards state
	{
		// bank has rewards
		require.False(t, getBankRewards(delAddr1, val).IsZero())
		// current rewards aren't empty
		require.False(t, getCurrentRewards(delAddr1, val).IsZero())
	}

	// produce an "overflow" event to start force unbond
	var fubTime time.Time
	{
		require.Equal(t, val.GetStatus(), sdk.Bonded)

		udTokens := sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(5))
		msg := staking.NewMsgUndelegate(sdk.AccAddress(valOpAddr1), val.GetOperator(), udTokens)

		res, err := sh(ctx, msg)
		require.NoError(t, err)
		require.NotNil(t, res)
		staking.EndBlocker(ctx, sk)

		val = updateVal(val)

		nextBlock()
	}
	// check force unbond scheduled
	{
		fubTime = val.GetScheduledUnbondStartTime()
		require.False(t, fubTime.IsZero())
	}

	// emulate time for force unbond and delegation unbond
	var valAddr sdk.ValAddress
	{
		// wait for scheduled force unbond
		nextBlock()
		ctx = ctx.WithBlockTime(fubTime)
		staking.EndBlocker(ctx, sk)

		// check validator is unbonding
		val = updateVal(val)
		require.Equal(t, sdk.Unbonding, val.GetStatus())
		valAddr = val.GetOperator()

		// woit for delegations unbonding
		nextBlock()
		ctx = ctx.WithBlockTime(ctx.BlockTime().Add(sk.GetParams(ctx).UnbondingTime))
		staking.EndBlocker(ctx, sk)

		// check validator deleted
		val = updateVal(val)
		require.Nil(t, val)
		_, found := k.GetValidatorLockedState(ctx, valAddr)
		require.False(t, found)
	}

	// check queryDelegationRewards
	{
		rewards := queryRewards(delAddr, valAddr)
		require.True(t, rewards.Reward.Current.IsZero())
		require.False(t, rewards.Reward.Total.IsZero())
	}

	// check queryDelegatorTotalRewards
	{
		rewards := queryTotalRewards(delAddr)
		require.Len(t, rewards.Rewards, 1)
		require.True(t, rewards.Rewards[0].Current.IsZero())
		require.False(t, rewards.Rewards[0].Total.IsZero())
		require.True(t, rewards.Rewards[0].Total.IsEqual(rewards.Total))
	}

	// withdraw non-existing validator and non-existing delegation rewards
	{
		rewards, err := k.WithdrawDelegationRewards(ctx, delAddr, valAddr)
		require.NoError(t, err)
		require.False(t, rewards.IsZero())
	}

	// check queryDelegationRewards (all zeros)
	{
		rewards := queryRewards(delAddr, valAddr)
		require.True(t, rewards.Reward.Current.IsZero())
		require.True(t, rewards.Reward.Total.IsZero())
	}
}
