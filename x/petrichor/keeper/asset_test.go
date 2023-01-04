package keeper_test

import (
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	abcitypes "github.com/tendermint/tendermint/abci/types"
	test_helpers "github.com/petrinetwork/petrichor/app"
	"github.com/petrinetwork/petrichor/x/petrichor"
	"github.com/petrinetwork/petrichor/x/petrichor/types"
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/staking/teststaking"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/stretchr/testify/require"
)

func TestRebalancingAfterRewardsRateChange(t *testing.T) {
	app, ctx := createTestContext(t)
	startTime := time.Now()
	ctx = ctx.WithBlockTime(startTime)
	app.PetrichorKeeper.InitGenesis(ctx, &types.GenesisState{
		Params: types.DefaultParams(),
		Assets: []types.PetrichorAsset{
			types.NewPetrichorAsset(PETRICHOR_TOKEN_DENOM, sdk.NewDec(2), sdk.ZeroDec(), startTime),
			types.NewPetrichorAsset(PETRICHOR_2_TOKEN_DENOM, sdk.NewDec(10), sdk.ZeroDec(), startTime),
		},
	})

	// Set tax and rewards to be zero for easier calculation
	distParams := app.DistrKeeper.GetParams(ctx)
	distParams.CommunityTax = sdk.ZeroDec()
	distParams.BaseProposerReward = sdk.ZeroDec()
	distParams.BonusProposerReward = sdk.ZeroDec()
	app.DistrKeeper.SetParams(ctx, distParams)

	// Accounts
	addrs := test_helpers.AddTestAddrsIncremental(app, ctx, 4, sdk.NewCoins(
		sdk.NewCoin(PETRICHOR_TOKEN_DENOM, sdk.NewInt(1000_000)),
		sdk.NewCoin(PETRICHOR_2_TOKEN_DENOM, sdk.NewInt(1000_000)),
	))
	pks := test_helpers.CreateTestPubKeys(2)
	powerReduction := app.StakingKeeper.PowerReduction(ctx)

	valAddr1 := sdk.ValAddress(addrs[0])
	_val1 := teststaking.NewValidator(t, valAddr1, pks[0])
	_val1.Commission = stakingtypes.Commission{
		CommissionRates: stakingtypes.CommissionRates{
			Rate:          sdk.NewDec(0),
			MaxRate:       sdk.NewDec(0),
			MaxChangeRate: sdk.NewDec(0),
		},
		UpdateTime: time.Now(),
	}
	_val1.Status = stakingtypes.Bonded
	test_helpers.RegisterNewValidator(t, app, ctx, _val1)
	val1, err := app.PetrichorKeeper.GetPetrichorValidator(ctx, valAddr1)
	require.NoError(t, err)

	user1 := addrs[2]

	// Start by delegating
	_, err = app.PetrichorKeeper.Delegate(ctx, user1, val1, sdk.NewCoin(PETRICHOR_TOKEN_DENOM, sdk.NewInt(1000_000)))
	require.NoError(t, err)

	assets := app.PetrichorKeeper.GetAllAssets(ctx)
	err = app.PetrichorKeeper.RebalanceBondTokenWeights(ctx, assets)
	require.NoError(t, err)
	require.Equal(t, sdk.NewInt(3_000_000), app.StakingKeeper.TotalBondedTokens(ctx))

	// Expecting voting power for the petrichor module
	val, found := app.StakingKeeper.GetValidator(ctx, valAddr1)
	require.True(t, found)
	require.Equal(t, int64(2), val.ConsensusPower(powerReduction))

	// Update but did not change reward weight
	err = app.PetrichorKeeper.UpdatePetrichorAsset(ctx, types.PetrichorAsset{
		Denom:                PETRICHOR_TOKEN_DENOM,
		RewardWeight:         sdk.NewDec(2),
		TakeRate:             sdk.NewDec(10),
		RewardChangeRate:     sdk.NewDec(0),
		RewardChangeInterval: 0,
	})
	require.NoError(t, err)

	// Expect no snapshots to be created
	iter := app.PetrichorKeeper.IterateWeightChangeSnapshot(ctx, PETRICHOR_TOKEN_DENOM, val.GetOperator(), 0)
	require.False(t, iter.Valid())

	err = app.PetrichorKeeper.UpdatePetrichorAsset(ctx, types.PetrichorAsset{
		Denom:                PETRICHOR_TOKEN_DENOM,
		RewardWeight:         sdk.NewDec(20),
		TakeRate:             sdk.NewDec(0),
		RewardChangeRate:     sdk.NewDec(0),
		RewardChangeInterval: 0,
	})
	require.NoError(t, err)

	// Expect a snapshot to be created
	iter = app.PetrichorKeeper.IterateWeightChangeSnapshot(ctx, PETRICHOR_TOKEN_DENOM, val.GetOperator(), 0)
	require.True(t, iter.Valid())

	assets = app.PetrichorKeeper.GetAllAssets(ctx)
	err = app.PetrichorKeeper.RebalanceBondTokenWeights(ctx, assets)
	require.NoError(t, err)
	require.Equal(t, sdk.NewInt(21_000_000), app.StakingKeeper.TotalBondedTokens(ctx))

	// Expecting voting power to increase
	val, found = app.StakingKeeper.GetValidator(ctx, valAddr1)
	require.True(t, found)
	require.Equal(t, int64(20), val.ConsensusPower(powerReduction))

	err = app.PetrichorKeeper.UpdatePetrichorAsset(ctx, types.PetrichorAsset{
		Denom:                PETRICHOR_TOKEN_DENOM,
		RewardWeight:         sdk.NewDec(1),
		TakeRate:             sdk.NewDec(0),
		RewardChangeRate:     sdk.NewDec(0),
		RewardChangeInterval: 0,
	})
	require.NoError(t, err)

	assets = app.PetrichorKeeper.GetAllAssets(ctx)
	err = app.PetrichorKeeper.RebalanceBondTokenWeights(ctx, assets)
	require.NoError(t, err)
	require.Equal(t, sdk.NewInt(2_000_000), app.StakingKeeper.TotalBondedTokens(ctx))

	// Expecting voting power to decrease
	val, found = app.StakingKeeper.GetValidator(ctx, valAddr1)
	require.True(t, found)
	require.Equal(t, int64(1), val.ConsensusPower(powerReduction))

	_, stop := petrichor.RunAllInvariants(ctx, app.PetrichorKeeper)
	require.False(t, stop)

}

func TestRebalancingWithUnbondedValidator(t *testing.T) {
	var err error
	app, ctx := createTestContext(t)
	bondDenom := app.StakingKeeper.BondDenom(ctx)
	startTime := time.Now()
	ctx = ctx.WithBlockTime(startTime).WithBlockHeight(1)
	app.PetrichorKeeper.InitGenesis(ctx, &types.GenesisState{
		Params: types.DefaultParams(),
		Assets: []types.PetrichorAsset{
			{
				Denom:        PETRICHOR_TOKEN_DENOM,
				RewardWeight: sdk.MustNewDecFromStr("0.1"),
				TakeRate:     sdk.NewDec(0),
				TotalTokens:  sdk.ZeroInt(),
			},
			{
				Denom:        PETRICHOR_2_TOKEN_DENOM,
				RewardWeight: sdk.MustNewDecFromStr("0.5"),
				TakeRate:     sdk.NewDec(0),
				TotalTokens:  sdk.ZeroInt(),
			},
		},
	})

	// Set tax and rewards to be zero for easier calculation
	distParams := app.DistrKeeper.GetParams(ctx)
	distParams.CommunityTax = sdk.ZeroDec()
	distParams.BaseProposerReward = sdk.ZeroDec()
	distParams.BonusProposerReward = sdk.ZeroDec()
	app.DistrKeeper.SetParams(ctx, distParams)

	// Accounts
	addrs := test_helpers.AddTestAddrsIncremental(app, ctx, 5, sdk.NewCoins(
		sdk.NewCoin(bondDenom, sdk.NewInt(10_000_000)),
		sdk.NewCoin(PETRICHOR_TOKEN_DENOM, sdk.NewInt(50_000_000)),
		sdk.NewCoin(PETRICHOR_2_TOKEN_DENOM, sdk.NewInt(50_000_000)),
	))
	pks := test_helpers.CreateTestPubKeys(2)

	// Increase the stake on genesis validator
	delegations := app.StakingKeeper.GetAllDelegations(ctx)
	require.Len(t, delegations, 1)
	valAddr0, err := sdk.ValAddressFromBech32(delegations[0].ValidatorAddress)
	require.NoError(t, err)
	val0, _ := app.StakingKeeper.GetValidator(ctx, valAddr0)
	_, err = app.StakingKeeper.Delegate(ctx, addrs[4], sdk.NewInt(9_000_000), stakingtypes.Unbonded, val0, true)
	require.NoError(t, err)

	// Creating two validators: 1 with 0% commission, 1 with 100% commission
	valAddr1 := sdk.ValAddress(addrs[0])
	_val1 := teststaking.NewValidator(t, valAddr1, pks[0])
	_val1.Commission = stakingtypes.Commission{
		CommissionRates: stakingtypes.CommissionRates{
			Rate:          sdk.NewDec(0),
			MaxRate:       sdk.NewDec(0),
			MaxChangeRate: sdk.NewDec(0),
		},
		UpdateTime: time.Now(),
	}
	_val1.Description.Moniker = "val1"
	test_helpers.RegisterNewValidator(t, app, ctx, _val1)
	val1, err := app.PetrichorKeeper.GetPetrichorValidator(ctx, valAddr1)
	require.NoError(t, err)

	valAddr2 := sdk.ValAddress(addrs[1])
	_val2 := teststaking.NewValidator(t, valAddr2, pks[1])
	_val2.Commission = stakingtypes.Commission{
		CommissionRates: stakingtypes.CommissionRates{
			Rate:          sdk.NewDec(1),
			MaxRate:       sdk.NewDec(1),
			MaxChangeRate: sdk.NewDec(0),
		},
		UpdateTime: time.Now(),
	}
	_val2.Description.Moniker = "val2"
	test_helpers.RegisterNewValidator(t, app, ctx, _val2)
	val2, err := app.PetrichorKeeper.GetPetrichorValidator(ctx, valAddr2)
	require.NoError(t, err)

	user1 := addrs[2]
	user2 := addrs[3]

	// Users add delegations
	_, err = app.PetrichorKeeper.Delegate(ctx, user1, val1, sdk.NewCoin(PETRICHOR_TOKEN_DENOM, sdk.NewInt(20_000_000)))
	require.NoError(t, err)
	_, err = app.PetrichorKeeper.Delegate(ctx, user1, val2, sdk.NewCoin(PETRICHOR_TOKEN_DENOM, sdk.NewInt(10_000_000)))
	require.NoError(t, err)
	_, err = app.PetrichorKeeper.Delegate(ctx, user2, val1, sdk.NewCoin(PETRICHOR_TOKEN_DENOM, sdk.NewInt(10_000_000)))
	require.NoError(t, err)
	_, err = app.PetrichorKeeper.Delegate(ctx, user2, val2, sdk.NewCoin(PETRICHOR_TOKEN_DENOM, sdk.NewInt(10_000_000)))
	require.NoError(t, err)

	_, err = app.PetrichorKeeper.Delegate(ctx, user1, val1, sdk.NewCoin(PETRICHOR_2_TOKEN_DENOM, sdk.NewInt(10_000_000)))
	require.NoError(t, err)
	_, err = app.PetrichorKeeper.Delegate(ctx, user1, val2, sdk.NewCoin(PETRICHOR_2_TOKEN_DENOM, sdk.NewInt(10_000_000)))
	require.NoError(t, err)
	_, err = app.PetrichorKeeper.Delegate(ctx, user2, val1, sdk.NewCoin(PETRICHOR_2_TOKEN_DENOM, sdk.NewInt(10_000_000)))
	require.NoError(t, err)
	_, err = app.PetrichorKeeper.Delegate(ctx, user2, val2, sdk.NewCoin(PETRICHOR_2_TOKEN_DENOM, sdk.NewInt(10_000_000)))
	require.NoError(t, err)
	require.NoError(t, err)

	assets := app.PetrichorKeeper.GetAllAssets(ctx)
	err = app.PetrichorKeeper.RebalanceBondTokenWeights(ctx, assets)
	require.NoError(t, err)
	_, err = app.StakingKeeper.ApplyAndReturnValidatorSetUpdates(ctx)
	require.NoError(t, err)
	require.Equal(t, sdk.NewInt(16_000_000), app.StakingKeeper.TotalBondedTokens(ctx))

	val1, _ = app.PetrichorKeeper.GetPetrichorValidator(ctx, valAddr1)
	val2, _ = app.PetrichorKeeper.GetPetrichorValidator(ctx, valAddr2)
	require.Greater(t, val1.Tokens.Int64(), val2.Tokens.Int64())

	// Set max validators to be 2 to trigger unbonding
	params := app.StakingKeeper.GetParams(ctx)
	params.MaxValidators = 2
	app.StakingKeeper.SetParams(ctx, params)
	_, err = app.StakingKeeper.ApplyAndReturnValidatorSetUpdates(ctx)
	require.NoError(t, err)
	require.Equal(t, sdk.NewInt(13_100_000), app.StakingKeeper.TotalBondedTokens(ctx))

	vals := app.StakingKeeper.GetBondedValidatorsByPower(ctx)
	require.Equal(t, 2, len(vals))
	require.Equal(t, "val1", vals[1].GetMoniker())

	assets = app.PetrichorKeeper.GetAllAssets(ctx)
	err = app.PetrichorKeeper.RebalanceBondTokenWeights(ctx, assets)
	require.NoError(t, err)
	require.Equal(t, sdk.NewInt(16_000_000).String(), app.StakingKeeper.TotalBondedTokens(ctx).String())

	app.PetrichorKeeper.GetPetrichorValidator(ctx, valAddr2)

	// Set max validators to be 3 to trigger rebonding
	params = app.StakingKeeper.GetParams(ctx)
	params.MaxValidators = 3
	app.StakingKeeper.SetParams(ctx, params)
	_, err = app.StakingKeeper.ApplyAndReturnValidatorSetUpdates(ctx)
	require.NoError(t, err)
	require.Equal(t, sdk.NewInt(18_900_000), app.StakingKeeper.TotalBondedTokens(ctx))

	assets = app.PetrichorKeeper.GetAllAssets(ctx)
	err = app.PetrichorKeeper.RebalanceBondTokenWeights(ctx, assets)
	require.NoError(t, err)
	require.Equal(t, sdk.NewInt(16_000_000).String(), app.StakingKeeper.TotalBondedTokens(ctx).String())

	_, stop := petrichor.RunAllInvariants(ctx, app.PetrichorKeeper)
	require.False(t, stop)
}

func TestRebalancingWithJailedValidator(t *testing.T) {
	var err error
	app, ctx := createTestContext(t)
	bondDenom := app.StakingKeeper.BondDenom(ctx)
	startTime := time.Now()
	ctx = ctx.WithBlockTime(startTime).WithBlockHeight(1)
	app.PetrichorKeeper.InitGenesis(ctx, &types.GenesisState{
		Params: types.DefaultParams(),
		Assets: []types.PetrichorAsset{
			{
				Denom:        PETRICHOR_TOKEN_DENOM,
				RewardWeight: sdk.MustNewDecFromStr("0.1"),
				TakeRate:     sdk.NewDec(0),
				TotalTokens:  sdk.ZeroInt(),
			},
			{
				Denom:        PETRICHOR_2_TOKEN_DENOM,
				RewardWeight: sdk.MustNewDecFromStr("0.5"),
				TakeRate:     sdk.NewDec(0),
				TotalTokens:  sdk.ZeroInt(),
			},
		},
	})

	// Set tax and rewards to be zero for easier calculation
	distParams := app.DistrKeeper.GetParams(ctx)
	distParams.CommunityTax = sdk.ZeroDec()
	distParams.BaseProposerReward = sdk.ZeroDec()
	distParams.BonusProposerReward = sdk.ZeroDec()
	app.DistrKeeper.SetParams(ctx, distParams)

	// Accounts
	addrs := test_helpers.AddTestAddrsIncremental(app, ctx, 5, sdk.NewCoins(
		sdk.NewCoin(bondDenom, sdk.NewInt(10_000_000)),
		sdk.NewCoin(PETRICHOR_TOKEN_DENOM, sdk.NewInt(50_000_000)),
		sdk.NewCoin(PETRICHOR_2_TOKEN_DENOM, sdk.NewInt(50_000_000)),
	))
	pks := test_helpers.CreateTestPubKeys(2)

	// Increase the stake on genesis validator
	delegations := app.StakingKeeper.GetAllDelegations(ctx)
	require.Len(t, delegations, 1)
	valAddr0, err := sdk.ValAddressFromBech32(delegations[0].ValidatorAddress)
	require.NoError(t, err)
	val0, _ := app.StakingKeeper.GetValidator(ctx, valAddr0)
	_, err = app.StakingKeeper.Delegate(ctx, addrs[4], sdk.NewInt(9_000_000), stakingtypes.Unbonded, val0, true)
	require.NoError(t, err)

	// Creating two validators: 1 with 0% commission, 1 with 100% commission
	valAddr1 := sdk.ValAddress(addrs[0])
	_val1 := teststaking.NewValidator(t, valAddr1, pks[0])
	_val1.Commission = stakingtypes.Commission{
		CommissionRates: stakingtypes.CommissionRates{
			Rate:          sdk.NewDec(0),
			MaxRate:       sdk.NewDec(0),
			MaxChangeRate: sdk.NewDec(0),
		},
		UpdateTime: time.Now(),
	}
	_val1.Description.Moniker = "val1"
	test_helpers.RegisterNewValidator(t, app, ctx, _val1)
	_, err = app.StakingKeeper.Delegate(ctx, addrs[0], sdk.NewInt(1_000_000), stakingtypes.Unbonded, _val1, true)
	require.NoError(t, err)
	val1, err := app.PetrichorKeeper.GetPetrichorValidator(ctx, valAddr1)
	require.NoError(t, err)

	valAddr2 := sdk.ValAddress(addrs[1])
	_val2 := teststaking.NewValidator(t, valAddr2, pks[1])
	_val2.Commission = stakingtypes.Commission{
		CommissionRates: stakingtypes.CommissionRates{
			Rate:          sdk.NewDec(1),
			MaxRate:       sdk.NewDec(1),
			MaxChangeRate: sdk.NewDec(0),
		},
		UpdateTime: time.Now(),
	}
	_val2.Description.Moniker = "val2"
	test_helpers.RegisterNewValidator(t, app, ctx, _val2)
	_, err = app.StakingKeeper.Delegate(ctx, addrs[1], sdk.NewInt(1_000_000), stakingtypes.Unbonded, _val2, true)
	require.NoError(t, err)
	val2, err := app.PetrichorKeeper.GetPetrichorValidator(ctx, valAddr2)
	require.NoError(t, err)

	user1 := addrs[2]
	user2 := addrs[3]

	// Users add delegations
	_, err = app.PetrichorKeeper.Delegate(ctx, user1, val1, sdk.NewCoin(PETRICHOR_TOKEN_DENOM, sdk.NewInt(20_000_000)))
	require.NoError(t, err)
	_, err = app.PetrichorKeeper.Delegate(ctx, user1, val2, sdk.NewCoin(PETRICHOR_TOKEN_DENOM, sdk.NewInt(10_000_000)))
	require.NoError(t, err)
	_, err = app.PetrichorKeeper.Delegate(ctx, user2, val1, sdk.NewCoin(PETRICHOR_TOKEN_DENOM, sdk.NewInt(10_000_000)))
	require.NoError(t, err)
	_, err = app.PetrichorKeeper.Delegate(ctx, user2, val2, sdk.NewCoin(PETRICHOR_TOKEN_DENOM, sdk.NewInt(10_000_000)))
	require.NoError(t, err)

	_, err = app.PetrichorKeeper.Delegate(ctx, user1, val1, sdk.NewCoin(PETRICHOR_2_TOKEN_DENOM, sdk.NewInt(10_000_000)))
	require.NoError(t, err)
	_, err = app.PetrichorKeeper.Delegate(ctx, user1, val2, sdk.NewCoin(PETRICHOR_2_TOKEN_DENOM, sdk.NewInt(10_000_000)))
	require.NoError(t, err)
	_, err = app.PetrichorKeeper.Delegate(ctx, user2, val1, sdk.NewCoin(PETRICHOR_2_TOKEN_DENOM, sdk.NewInt(10_000_000)))
	require.NoError(t, err)
	_, err = app.PetrichorKeeper.Delegate(ctx, user2, val2, sdk.NewCoin(PETRICHOR_2_TOKEN_DENOM, sdk.NewInt(10_000_000)))
	require.NoError(t, err)
	require.NoError(t, err)

	_, err = app.StakingKeeper.ApplyAndReturnValidatorSetUpdates(ctx)
	require.NoError(t, err)
	assets := app.PetrichorKeeper.GetAllAssets(ctx)
	err = app.PetrichorKeeper.RebalanceBondTokenWeights(ctx, assets)
	require.NoError(t, err)
	// 12 * 1.6 = 19.2
	require.Equal(t, sdk.NewInt(19_200_000), app.StakingKeeper.TotalBondedTokens(ctx))

	val1, _ = app.PetrichorKeeper.GetPetrichorValidator(ctx, valAddr1)
	val2, _ = app.PetrichorKeeper.GetPetrichorValidator(ctx, valAddr2)
	require.Greater(t, val1.Tokens.Int64(), val2.Tokens.Int64())

	// Jail validator
	cons2, _ := val2.GetConsAddr()
	app.SlashingKeeper.Jail(ctx, cons2)
	_, err = app.StakingKeeper.ApplyAndReturnValidatorSetUpdates(ctx)
	require.NoError(t, err)
	require.Equal(t, sdk.NewInt(14_720_000), app.StakingKeeper.TotalBondedTokens(ctx))

	vals := app.StakingKeeper.GetBondedValidatorsByPower(ctx)
	require.Equal(t, 2, len(vals))
	require.Equal(t, "val1", vals[1].GetMoniker())

	assets = app.PetrichorKeeper.GetAllAssets(ctx)
	err = app.PetrichorKeeper.RebalanceBondTokenWeights(ctx, assets)
	require.NoError(t, err)
	// 11 * 1.6 = 17.6
	require.Equal(t, sdk.NewInt(17_600_000).String(), app.StakingKeeper.TotalBondedTokens(ctx).String())

	app.PetrichorKeeper.GetPetrichorValidator(ctx, valAddr2)

	// Unjail validator
	err = app.SlashingKeeper.Unjail(ctx, valAddr2)
	require.NoError(t, err)
	_, err = app.StakingKeeper.ApplyAndReturnValidatorSetUpdates(ctx)
	require.NoError(t, err)
	require.Equal(t, sdk.NewInt(22_080_000), app.StakingKeeper.TotalBondedTokens(ctx))

	assets = app.PetrichorKeeper.GetAllAssets(ctx)
	err = app.PetrichorKeeper.RebalanceBondTokenWeights(ctx, assets)
	require.NoError(t, err)
	require.Equal(t, sdk.NewInt(19_200_000).String(), app.StakingKeeper.TotalBondedTokens(ctx).String())

	_, stop := petrichor.RunAllInvariants(ctx, app.PetrichorKeeper)
	require.False(t, stop)
}

func TestRebalancingWithDelayedRewardsStartTime(t *testing.T) {
	var err error
	app, ctx := createTestContext(t)
	bondDenom := app.StakingKeeper.BondDenom(ctx)
	startTime := time.Now()
	ctx = ctx.WithBlockTime(startTime).WithBlockHeight(1)

	app.PetrichorKeeper.InitGenesis(ctx, &types.GenesisState{
		Params: types.DefaultParams(),
		Assets: []types.PetrichorAsset{
			types.NewPetrichorAsset(PETRICHOR_TOKEN_DENOM, sdk.MustNewDecFromStr("0.5"), sdk.MustNewDecFromStr("0.1"), startTime.Add(time.Hour*24)),
			types.NewPetrichorAsset(PETRICHOR_2_TOKEN_DENOM, sdk.MustNewDecFromStr("0.2"), sdk.MustNewDecFromStr("0.1"), startTime.Add(time.Hour*24*2)),
		},
	})

	// Set tax and rewards to be zero for easier calculation
	distParams := app.DistrKeeper.GetParams(ctx)
	distParams.CommunityTax = sdk.ZeroDec()
	distParams.BaseProposerReward = sdk.ZeroDec()
	distParams.BonusProposerReward = sdk.ZeroDec()
	app.DistrKeeper.SetParams(ctx, distParams)

	// Accounts
	addrs := test_helpers.AddTestAddrsIncremental(app, ctx, 5, sdk.NewCoins(
		sdk.NewCoin(bondDenom, sdk.NewInt(10_000_000)),
		sdk.NewCoin(PETRICHOR_TOKEN_DENOM, sdk.NewInt(50_000_000)),
		sdk.NewCoin(PETRICHOR_2_TOKEN_DENOM, sdk.NewInt(50_000_000)),
	))
	pks := test_helpers.CreateTestPubKeys(2)

	// Increase the stake on genesis validator
	delegations := app.StakingKeeper.GetAllDelegations(ctx)
	require.Len(t, delegations, 1)
	valAddr0, err := sdk.ValAddressFromBech32(delegations[0].ValidatorAddress)
	require.NoError(t, err)
	val0, _ := app.StakingKeeper.GetValidator(ctx, valAddr0)
	_, err = app.StakingKeeper.Delegate(ctx, addrs[4], sdk.NewInt(9_000_000), stakingtypes.Unbonded, val0, true)
	require.NoError(t, err)

	// Creating two validators: 1 with 0% commission, 1 with 100% commission
	valAddr1 := sdk.ValAddress(addrs[0])
	_val1 := teststaking.NewValidator(t, valAddr1, pks[0])
	_val1.Commission = stakingtypes.Commission{
		CommissionRates: stakingtypes.CommissionRates{
			Rate:          sdk.NewDec(0),
			MaxRate:       sdk.NewDec(0),
			MaxChangeRate: sdk.NewDec(0),
		},
		UpdateTime: time.Now(),
	}
	_val1.Description.Moniker = "val1"
	test_helpers.RegisterNewValidator(t, app, ctx, _val1)
	_, err = app.StakingKeeper.Delegate(ctx, addrs[0], sdk.NewInt(1_000_000), stakingtypes.Unbonded, _val1, true)
	require.NoError(t, err)
	val1, err := app.PetrichorKeeper.GetPetrichorValidator(ctx, valAddr1)
	require.NoError(t, err)

	valAddr2 := sdk.ValAddress(addrs[1])
	_val2 := teststaking.NewValidator(t, valAddr2, pks[1])
	_val2.Commission = stakingtypes.Commission{
		CommissionRates: stakingtypes.CommissionRates{
			Rate:          sdk.NewDec(1),
			MaxRate:       sdk.NewDec(1),
			MaxChangeRate: sdk.NewDec(0),
		},
		UpdateTime: time.Now(),
	}
	_val2.Description.Moniker = "val2"
	test_helpers.RegisterNewValidator(t, app, ctx, _val2)
	_, err = app.StakingKeeper.Delegate(ctx, addrs[1], sdk.NewInt(1_000_000), stakingtypes.Unbonded, _val2, true)
	require.NoError(t, err)
	val2, err := app.PetrichorKeeper.GetPetrichorValidator(ctx, valAddr2)
	require.NoError(t, err)

	user1 := addrs[2]
	user2 := addrs[3]

	// Users add delegations
	_, err = app.PetrichorKeeper.Delegate(ctx, user1, val1, sdk.NewCoin(PETRICHOR_TOKEN_DENOM, sdk.NewInt(20_000_000)))
	require.NoError(t, err)
	_, err = app.PetrichorKeeper.Delegate(ctx, user1, val2, sdk.NewCoin(PETRICHOR_TOKEN_DENOM, sdk.NewInt(10_000_000)))
	require.NoError(t, err)
	_, err = app.PetrichorKeeper.Delegate(ctx, user2, val1, sdk.NewCoin(PETRICHOR_TOKEN_DENOM, sdk.NewInt(10_000_000)))
	require.NoError(t, err)
	_, err = app.PetrichorKeeper.Delegate(ctx, user2, val2, sdk.NewCoin(PETRICHOR_TOKEN_DENOM, sdk.NewInt(10_000_000)))
	require.NoError(t, err)

	_, err = app.PetrichorKeeper.Delegate(ctx, user1, val1, sdk.NewCoin(PETRICHOR_2_TOKEN_DENOM, sdk.NewInt(10_000_000)))
	require.NoError(t, err)
	_, err = app.PetrichorKeeper.Delegate(ctx, user1, val2, sdk.NewCoin(PETRICHOR_2_TOKEN_DENOM, sdk.NewInt(10_000_000)))
	require.NoError(t, err)
	_, err = app.PetrichorKeeper.Delegate(ctx, user2, val1, sdk.NewCoin(PETRICHOR_2_TOKEN_DENOM, sdk.NewInt(10_000_000)))
	require.NoError(t, err)
	_, err = app.PetrichorKeeper.Delegate(ctx, user2, val2, sdk.NewCoin(PETRICHOR_2_TOKEN_DENOM, sdk.NewInt(10_000_000)))
	require.NoError(t, err)
	_, err = app.StakingKeeper.ApplyAndReturnValidatorSetUpdates(ctx)
	require.NoError(t, err)

	// Expect that rewards rates are not updated due to ctx being before rewards start time
	require.NoError(t, err)
	assets := app.PetrichorKeeper.GetAllAssets(ctx)
	err = app.PetrichorKeeper.RebalanceBondTokenWeights(ctx, assets)
	require.NoError(t, err)
	require.Equal(t, sdk.NewInt(12_000_000), app.StakingKeeper.TotalBondedTokens(ctx))

	// Expect that rewards rates are updated only for petrichor 1
	ctx = ctx.WithBlockTime(ctx.BlockTime().Add(time.Hour * 24))
	assets = app.PetrichorKeeper.GetAllAssets(ctx)
	err = app.PetrichorKeeper.RebalanceBondTokenWeights(ctx, assets)
	require.NoError(t, err)
	// 12 * 1.5 = 18
	require.Equal(t, sdk.NewInt(18_000_000), app.StakingKeeper.TotalBondedTokens(ctx))

	// Expect that rewards rates are updated all petrichors
	ctx = ctx.WithBlockTime(ctx.BlockTime().Add(time.Hour * 48))
	assets = app.PetrichorKeeper.GetAllAssets(ctx)
	err = app.PetrichorKeeper.RebalanceBondTokenWeights(ctx, assets)
	require.NoError(t, err)
	// 12 * 1.7 = 18
	require.Equal(t, sdk.NewInt(20_400_000), app.StakingKeeper.TotalBondedTokens(ctx))

	_, stop := petrichor.RunAllInvariants(ctx, app.PetrichorKeeper)
	require.False(t, stop)
}

func TestConsumingRebalancingEvent(t *testing.T) {
	app, ctx := createTestContext(t)
	startTime := time.Now()
	ctx = ctx.WithBlockTime(startTime).WithBlockHeight(1)

	app.PetrichorKeeper.InitGenesis(ctx, &types.GenesisState{
		Params: types.DefaultParams(),
		Assets: []types.PetrichorAsset{
			types.NewPetrichorAsset(PETRICHOR_TOKEN_DENOM, sdk.MustNewDecFromStr("0.5"), sdk.MustNewDecFromStr("0.1"), startTime.Add(time.Hour*24)),
			types.NewPetrichorAsset(PETRICHOR_2_TOKEN_DENOM, sdk.MustNewDecFromStr("0.2"), sdk.MustNewDecFromStr("0.1"), startTime.Add(time.Hour*24*2)),
		},
	})

	app.PetrichorKeeper.QueueAssetRebalanceEvent(ctx)
	store := ctx.KVStore(app.PetrichorKeeper.StoreKey())
	key := types.AssetRebalanceQueueKey
	b := store.Get(key)
	require.NotNil(t, b)

	require.True(t, app.PetrichorKeeper.ConsumeAssetRebalanceEvent(ctx))
	b = store.Get(key)
	require.Nil(t, b)

	require.False(t, app.PetrichorKeeper.ConsumeAssetRebalanceEvent(ctx))
}

func TestRewardWeightDecay(t *testing.T) {
	var err error
	app, ctx := createTestContext(t)
	bondDenom := app.StakingKeeper.BondDenom(ctx)
	startTime := time.Now()
	ctx = ctx.WithBlockTime(startTime).WithBlockHeight(1)
	app.PetrichorKeeper.InitGenesis(ctx, &types.GenesisState{
		Params: types.DefaultParams(),
		Assets: []types.PetrichorAsset{},
	})

	// Accounts
	addrs := test_helpers.AddTestAddrsIncremental(app, ctx, 5, sdk.NewCoins(
		sdk.NewCoin(bondDenom, sdk.NewInt(10_000_000)),
		sdk.NewCoin(PETRICHOR_TOKEN_DENOM, sdk.NewInt(50_000_000)),
		sdk.NewCoin(PETRICHOR_2_TOKEN_DENOM, sdk.NewInt(50_000_000)),
	))

	// Increase the stake on genesis validator
	delegations := app.StakingKeeper.GetAllDelegations(ctx)
	require.Len(t, delegations, 1)
	valAddr0, err := sdk.ValAddressFromBech32(delegations[0].ValidatorAddress)
	require.NoError(t, err)
	val0, _ := app.StakingKeeper.GetValidator(ctx, valAddr0)
	_, err = app.StakingKeeper.Delegate(ctx, addrs[4], sdk.NewInt(9_000_000), stakingtypes.Unbonded, val0, true)
	require.NoError(t, err)

	// Pass a proposal to add a new asset with a decay rate
	decayInterval := time.Hour * 24 * 30
	app.PetrichorKeeper.CreatePetrichor(ctx, &types.MsgCreatePetrichorProposal{
		Title:                "",
		Description:          "",
		Denom:                PETRICHOR_TOKEN_DENOM,
		RewardWeight:         sdk.NewDec(1),
		TakeRate:             sdk.ZeroDec(),
		RewardChangeRate:     sdk.MustNewDecFromStr("0.5"),
		RewardChangeInterval: decayInterval,
	})
	asset, _ := app.PetrichorKeeper.GetAssetByDenom(ctx, PETRICHOR_TOKEN_DENOM)

	assets := app.PetrichorKeeper.GetAllAssets(ctx)
	// Running the decay hook now should do nothing
	app.PetrichorKeeper.RewardWeightChangeHook(ctx, assets)

	// Move block time to after change interval + one year
	ctx = ctx.WithBlockTime(asset.RewardStartTime.Add(decayInterval))

	// Running the decay hook should update reward weight
	app.PetrichorKeeper.RewardWeightChangeHook(ctx, assets)
	updatedAsset, _ := app.PetrichorKeeper.GetAssetByDenom(ctx, PETRICHOR_TOKEN_DENOM)
	require.Equal(t, types.PetrichorAsset{
		Denom:                PETRICHOR_TOKEN_DENOM,
		RewardWeight:         sdk.MustNewDecFromStr("0.5"),
		TakeRate:             asset.TakeRate,
		TotalTokens:          asset.TotalTokens,
		TotalValidatorShares: asset.TotalValidatorShares,
		RewardStartTime:      asset.RewardStartTime,
		RewardChangeRate:     asset.RewardChangeRate,
		RewardChangeInterval: asset.RewardChangeInterval,
		LastRewardChangeTime: ctx.BlockTime(),
	}, updatedAsset)

	// There should be a rebalancing event stored
	require.True(t, app.PetrichorKeeper.ConsumeAssetRebalanceEvent(ctx))

	ctx = ctx.WithBlockTime(ctx.BlockTime().Add(time.Hour * 10))
	// Updating the petrichor asset through proposal should queue another decay event
	app.PetrichorKeeper.UpdatePetrichor(ctx, &types.MsgUpdatePetrichorProposal{
		Title:                "",
		Description:          "",
		Denom:                PETRICHOR_TOKEN_DENOM,
		RewardWeight:         sdk.MustNewDecFromStr("0.5"),
		TakeRate:             sdk.ZeroDec(),
		RewardChangeRate:     sdk.ZeroDec(),
		RewardChangeInterval: 0,
	})

	// Updating petrichor asset again with a non-zero decay
	app.PetrichorKeeper.UpdatePetrichor(ctx, &types.MsgUpdatePetrichorProposal{
		Title:                "",
		Description:          "",
		Denom:                PETRICHOR_TOKEN_DENOM,
		RewardWeight:         sdk.MustNewDecFromStr("0.5"),
		TakeRate:             sdk.ZeroDec(),
		RewardChangeRate:     sdk.MustNewDecFromStr("0.1"),
		RewardChangeInterval: decayInterval,
	})

	// Add a new asset with an initial 0 decay
	err = app.PetrichorKeeper.CreatePetrichor(ctx, &types.MsgCreatePetrichorProposal{
		Title:                "",
		Description:          "",
		Denom:                PETRICHOR_2_TOKEN_DENOM,
		RewardWeight:         sdk.NewDec(1),
		TakeRate:             sdk.ZeroDec(),
		RewardChangeRate:     sdk.ZeroDec(),
		RewardChangeInterval: decayInterval,
	})
	require.NoError(t, err)

	ctx = ctx.WithBlockTime(ctx.BlockTime().Add(time.Hour))
	// Updating petrichor asset again with a non-zero decay
	err = app.PetrichorKeeper.UpdatePetrichor(ctx, &types.MsgUpdatePetrichorProposal{
		Title:                "",
		Description:          "",
		Denom:                PETRICHOR_2_TOKEN_DENOM,
		RewardWeight:         sdk.MustNewDecFromStr("0.5"),
		TakeRate:             sdk.ZeroDec(),
		RewardChangeRate:     sdk.MustNewDecFromStr("0.1"),
		RewardChangeInterval: decayInterval,
	})
	require.NoError(t, err)
}

func TestRewardWeightDecayOverTime(t *testing.T) {
	var err error
	app, ctx := createTestContext(t)
	bondDenom := app.StakingKeeper.BondDenom(ctx)
	startTime := time.Now().UTC()
	ctx = ctx.WithBlockTime(startTime).WithBlockHeight(1)
	app.PetrichorKeeper.InitGenesis(ctx, &types.GenesisState{
		Params: types.DefaultParams(),
		Assets: []types.PetrichorAsset{},
	})

	// Accounts
	addrs := test_helpers.AddTestAddrsIncremental(app, ctx, 5, sdk.NewCoins(
		sdk.NewCoin(bondDenom, sdk.NewInt(1_000_000_000_000)),
		sdk.NewCoin(PETRICHOR_TOKEN_DENOM, sdk.NewInt(5_000_000)),
		sdk.NewCoin(PETRICHOR_2_TOKEN_DENOM, sdk.NewInt(5_000_000)),
	))

	// Increase the stake on genesis validator
	delegations := app.StakingKeeper.GetAllDelegations(ctx)
	require.Len(t, delegations, 1)
	valAddr0, err := sdk.ValAddressFromBech32(delegations[0].ValidatorAddress)
	require.NoError(t, err)
	_val0, _ := app.StakingKeeper.GetValidator(ctx, valAddr0)
	_, err = app.StakingKeeper.Delegate(ctx, addrs[4], sdk.NewInt(9_000_000), stakingtypes.Unbonded, _val0, true)
	require.NoError(t, err)

	val0, _ := app.PetrichorKeeper.GetPetrichorValidator(ctx, _val0.GetOperator())
	require.NoError(t, err)

	// Pass a proposal to add a new asset with a decay rate
	decayInterval := time.Minute
	decayRate := sdk.MustNewDecFromStr("0.99998")
	app.PetrichorKeeper.CreatePetrichor(ctx, &types.MsgCreatePetrichorProposal{
		Title:                "",
		Description:          "",
		Denom:                PETRICHOR_TOKEN_DENOM,
		RewardWeight:         sdk.NewDec(1),
		TakeRate:             sdk.ZeroDec(),
		RewardChangeRate:     decayRate,
		RewardChangeInterval: decayInterval,
	})

	// Delegate to validator
	_, err = app.PetrichorKeeper.Delegate(ctx, addrs[1], val0, sdk.NewCoin(PETRICHOR_TOKEN_DENOM, sdk.NewInt(5_000_000)))
	require.NoError(t, err)
	//
	assets := app.PetrichorKeeper.GetAllAssets(ctx)
	err = app.PetrichorKeeper.RebalanceHook(ctx, assets)
	require.NoError(t, err)

	asset, _ := app.PetrichorKeeper.GetAssetByDenom(ctx, PETRICHOR_TOKEN_DENOM)

	// Simulate the chain running for 10 days with a defined block time
	blockTime := time.Second * 20
	for i := time.Duration(0); i <= time.Hour*24*10; i += blockTime {
		ctx = ctx.WithBlockTime(ctx.BlockTime().Add(blockTime)).WithBlockHeight(ctx.BlockHeight() + 1)
		assets = app.PetrichorKeeper.GetAllAssets(ctx)
		// Running the decay hook should update reward weight
		app.PetrichorKeeper.RewardWeightChangeHook(ctx, assets)
	}

	// time passed minus reward delay time (rewards and decay only start after the delay)
	totalDecayTime := (time.Hour * 24 * 10) - app.PetrichorKeeper.RewardDelayTime(ctx)
	intervals := uint64(totalDecayTime / decayInterval)
	asset, _ = app.PetrichorKeeper.GetAssetByDenom(ctx, PETRICHOR_TOKEN_DENOM)
	require.Equal(t, startTime.Add(app.PetrichorKeeper.RewardDelayTime(ctx)).Add(decayInterval*time.Duration(intervals)), asset.LastRewardChangeTime)
	require.True(t, decayRate.Power(intervals).Sub(asset.RewardWeight).LT(sdk.MustNewDecFromStr("0.0000000001")))
}

func TestClaimTakeRate(t *testing.T) {
	app, ctx := createTestContext(t)
	startTime := time.Now().UTC()
	ctx = ctx.WithBlockTime(startTime)
	ctx = ctx.WithBlockHeight(1)
	takeRateInterval := time.Minute * 5
	app.PetrichorKeeper.InitGenesis(ctx, &types.GenesisState{
		Params: types.Params{
			RewardDelayTime:       time.Minute * 60,
			TakeRateClaimInterval: takeRateInterval,
			LastTakeRateClaimTime: startTime,
		},
		Assets: []types.PetrichorAsset{
			types.NewPetrichorAsset(PETRICHOR_TOKEN_DENOM, sdk.NewDec(2), sdk.MustNewDecFromStr("0.5"), startTime),
			types.NewPetrichorAsset(PETRICHOR_2_TOKEN_DENOM, sdk.NewDec(10), sdk.NewDec(0), startTime),
		},
	})

	// Accounts
	feeCollectorAddr := app.AccountKeeper.GetModuleAddress(authtypes.FeeCollectorName)
	delegations := app.StakingKeeper.GetAllDelegations(ctx)
	valAddr1, err := sdk.ValAddressFromBech32(delegations[0].ValidatorAddress)
	require.NoError(t, err)
	val1, err := app.PetrichorKeeper.GetPetrichorValidator(ctx, valAddr1)
	require.NoError(t, err)
	addrs := test_helpers.AddTestAddrsIncremental(app, ctx, 1, sdk.NewCoins(
		sdk.NewCoin(PETRICHOR_TOKEN_DENOM, sdk.NewInt(1000_000_000)),
		sdk.NewCoin(PETRICHOR_2_TOKEN_DENOM, sdk.NewInt(1000_000_000)),
	))
	user1 := addrs[0]

	app.PetrichorKeeper.Delegate(ctx, user1, val1, sdk.NewCoin(PETRICHOR_TOKEN_DENOM, sdk.NewInt(1000_000_000)))
	app.PetrichorKeeper.Delegate(ctx, user1, val1, sdk.NewCoin(PETRICHOR_2_TOKEN_DENOM, sdk.NewInt(1000_000_000)))

	assets := app.PetrichorKeeper.GetAllAssets(ctx)
	err = app.PetrichorKeeper.RebalanceBondTokenWeights(ctx, assets)
	require.NoError(t, err)
	// Check total bonded amount
	require.Equal(t, sdk.NewInt(13_000_000), app.StakingKeeper.TotalBondedTokens(ctx))

	// Calling it immediately will not update anything
	coins, err := app.PetrichorKeeper.DeductAssetsHook(ctx, assets)
	require.Nil(t, coins)
	require.Nil(t, err)

	// Advance block time
	timePassed := time.Minute*5 + time.Second
	ctx = ctx.WithBlockTime(ctx.BlockTime().Add(timePassed))
	ctx = ctx.WithBlockHeight(2)
	coinsClaimed, _ := app.PetrichorKeeper.DeductAssetsHook(ctx, assets)
	coins = app.BankKeeper.GetAllBalances(ctx, feeCollectorAddr)
	require.Equal(t, coinsClaimed, coins)

	expectedAmount := sdk.MustNewDecFromStr("0.5").Mul(sdk.NewDec(timePassed.Nanoseconds() / takeRateInterval.Nanoseconds())).MulInt(sdk.NewInt(1000_000_000))
	require.Equal(t, expectedAmount.TruncateInt(), coins.AmountOf(PETRICHOR_TOKEN_DENOM))

	lastUpdate := app.PetrichorKeeper.LastRewardClaimTime(ctx)
	require.Equal(t, startTime.Add(takeRateInterval), lastUpdate)

	asset, found := app.PetrichorKeeper.GetAssetByDenom(ctx, PETRICHOR_TOKEN_DENOM)
	require.True(t, found)
	require.Equal(t, sdk.NewDec(2), asset.RewardWeight)

	// At the next begin block, tokens will be distributed from the fee pool
	cons, _ := val1.GetConsAddr()
	app.DistrKeeper.AllocateTokens(ctx, 1, 1, cons, []abcitypes.VoteInfo{
		{
			Validator: abcitypes.Validator{
				Address: cons,
				Power:   1,
			},
			SignedLastBlock: true,
		},
	})

	rewards := app.DistrKeeper.GetValidatorOutstandingRewards(ctx, valAddr1).Rewards
	community := app.DistrKeeper.GetFeePool(ctx).CommunityPool
	// This is case, validator 1 has 0% commission
	commission := app.DistrKeeper.GetValidatorAccumulatedCommission(ctx, valAddr1).Commission
	require.Equal(t, sdk.DecCoins(nil), commission)
	// And rewards + community pool should add up to total coins claimed
	require.Equal(t,
		sdk.NewDecFromInt(coinsClaimed.AmountOf(PETRICHOR_TOKEN_DENOM)),
		rewards.AmountOf(PETRICHOR_TOKEN_DENOM).Add(community.AmountOf(PETRICHOR_TOKEN_DENOM)),
	)
}

func TestClaimTakeRateToZero(t *testing.T) {
	app, ctx := createTestContext(t)
	startTime := time.Now().UTC()
	ctx = ctx.WithBlockTime(startTime)
	ctx = ctx.WithBlockHeight(1)
	takeRateInterval := time.Minute * 5
	asset := types.NewPetrichorAsset(PETRICHOR_TOKEN_DENOM, sdk.NewDec(2), sdk.MustNewDecFromStr("0.8"), startTime)
	app.PetrichorKeeper.InitGenesis(ctx, &types.GenesisState{
		Params: types.Params{
			RewardDelayTime:       time.Minute * 60,
			TakeRateClaimInterval: takeRateInterval,
			LastTakeRateClaimTime: startTime,
		},
		Assets: []types.PetrichorAsset{
			asset,
		},
	})

	// Accounts
	delegations := app.StakingKeeper.GetAllDelegations(ctx)
	valAddr1, err := sdk.ValAddressFromBech32(delegations[0].ValidatorAddress)
	require.NoError(t, err)
	val1, err := app.PetrichorKeeper.GetPetrichorValidator(ctx, valAddr1)
	require.NoError(t, err)
	addrs := test_helpers.AddTestAddrsIncremental(app, ctx, 1, sdk.NewCoins(
		sdk.NewCoin(PETRICHOR_TOKEN_DENOM, sdk.NewInt(1000_000_000)),
	))
	user1 := addrs[0]

	app.PetrichorKeeper.Delegate(ctx, user1, val1, sdk.NewCoin(PETRICHOR_TOKEN_DENOM, sdk.NewInt(1000_000_000)))
	assets := app.PetrichorKeeper.GetAllAssets(ctx)
	err = app.PetrichorKeeper.RebalanceBondTokenWeights(ctx, assets)
	require.NoError(t, err)

	timePassed := time.Minute * 5
	// Advance block time
	for i := 0; i < 1000; i++ {
		ctx = ctx.WithBlockTime(ctx.BlockTime().Add(timePassed))
		ctx = ctx.WithBlockHeight(ctx.BlockHeight() + 1)
		_, err = app.PetrichorKeeper.DeductAssetsHook(ctx, assets)
		require.NoError(t, err)
	}

	asset, _ = app.PetrichorKeeper.GetAssetByDenom(ctx, PETRICHOR_TOKEN_DENOM)
	require.True(t, asset.TotalTokens.GTE(sdk.OneInt()))
}
