package keeper

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

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
		bankCoins := k.GetDelegatorRewardsBankCoins(ctx, del.GetDelegatorAddr())
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
		bankCoins := k.GetDelegatorRewardsBankCoins(ctx, del.GetDelegatorAddr())
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
		bankCoins := k.GetDelegatorRewardsBankCoins(ctx, del.GetDelegatorAddr())
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
	}

	// allocate some rewards
	{
		tokens := sdk.DecCoins{sdk.NewDecCoinFromDec(sdk.DefaultBondDenom, sdk.NewDec(10))}
		k.AllocateTokensToValidator(ctx, val, tokens, sdk.DecCoins{})
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
		}
	}

	// transfer all rewards
	{
		_, err := k.WithdrawDelegationRewards(ctx, delAddr, val.GetOperator())
		require.Error(t, err)
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
