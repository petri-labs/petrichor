package keeper_test

import (
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	"github.com/cosmos/cosmos-sdk/x/staking/teststaking"
	test_helpers "github.com/petrinetwork/petrichor/app"
	"github.com/petrinetwork/petrichor/x/petrichor/types"
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
)

func TestGenesis(t *testing.T) {
	app, ctx := createTestContext(t)
	app.PetrichorKeeper.InitGenesis(ctx, &types.GenesisState{
		Params: types.Params{
			RewardDelayTime:       time.Duration(1000000),
			TakeRateClaimInterval: time.Duration(1000000),
			LastTakeRateClaimTime: time.Unix(0, 0).UTC(),
		},
		Assets: []types.PetrichorAsset{
			types.NewPetrichorAsset("stake", sdk.NewDec(1), sdk.ZeroDec(), ctx.BlockTime()),
		},
	})

	delay := app.PetrichorKeeper.RewardDelayTime(ctx)
	require.Equal(t, time.Duration(1000000), delay)

	interval := app.PetrichorKeeper.RewardClaimInterval(ctx)
	require.Equal(t, time.Duration(1000000), interval)

	lastClaimTime := app.PetrichorKeeper.LastRewardClaimTime(ctx)
	require.Equal(t, time.Unix(0, 0).UTC(), lastClaimTime)

	assets := app.PetrichorKeeper.GetAllAssets(ctx)
	require.Equal(t, 1, len(assets))
	require.Equal(t, &types.PetrichorAsset{
		Denom:                "stake",
		RewardWeight:         sdk.NewDec(1.0),
		TakeRate:             sdk.NewDec(0.0),
		TotalTokens:          sdk.ZeroInt(),
		TotalValidatorShares: sdk.ZeroDec(),
		RewardStartTime:      ctx.BlockTime(),
		RewardChangeRate:     sdk.OneDec(),
		RewardChangeInterval: 0,
	}, assets[0])
}

func TestExportAndImportGenesis(t *testing.T) {
	app, ctx := createTestContext(t)
	ctx = ctx.WithBlockTime(time.Now()).WithBlockHeight(1)
	app.PetrichorKeeper.InitGenesis(ctx, &types.GenesisState{
		Params: types.Params{
			RewardDelayTime:       time.Duration(1000000),
			TakeRateClaimInterval: time.Duration(1000000),
			LastTakeRateClaimTime: time.Unix(0, 0).UTC(),
		},
		Assets: []types.PetrichorAsset{},
	})

	// All the addresses needed
	delegations := app.StakingKeeper.GetAllDelegations(ctx)
	require.Len(t, delegations, 1)
	delAddr, err := sdk.AccAddressFromBech32(delegations[0].DelegatorAddress)
	require.NoError(t, err)
	valAddr, err := sdk.ValAddressFromBech32(delegations[0].ValidatorAddress)
	require.NoError(t, err)
	val1, err := app.PetrichorKeeper.GetPetrichorValidator(ctx, valAddr)
	require.NoError(t, err)
	addrs := test_helpers.AddTestAddrsIncremental(app, ctx, 3, sdk.NewCoins(
		sdk.NewCoin(PETRICHOR_TOKEN_DENOM, sdk.NewInt(1000_000)),
		sdk.NewCoin(PETRICHOR_2_TOKEN_DENOM, sdk.NewInt(1000_000)),
	))
	valAddr2 := sdk.ValAddress(addrs[0])
	_val2 := teststaking.NewValidator(t, valAddr2, test_helpers.CreateTestPubKeys(1)[0])
	test_helpers.RegisterNewValidator(t, app, ctx, _val2)
	val2, err := app.PetrichorKeeper.GetPetrichorValidator(ctx, valAddr2)
	require.NoError(t, err)

	// Add petrichor asset
	err = app.PetrichorKeeper.CreatePetrichor(ctx, &types.MsgCreatePetrichorProposal{
		Title:                "",
		Description:          "",
		Denom:                PETRICHOR_TOKEN_DENOM,
		RewardWeight:         sdk.NewDec(1),
		TakeRate:             sdk.NewDec(0),
		RewardChangeRate:     sdk.MustNewDecFromStr("0.5"),
		RewardChangeInterval: time.Hour * 24,
	})
	require.NoError(t, err)

	// Delegate
	delegationCoin := sdk.NewCoin(PETRICHOR_TOKEN_DENOM, sdk.NewInt(1000_000_000))
	err = app.BankKeeper.MintCoins(ctx, minttypes.ModuleName, sdk.NewCoins(delegationCoin))
	require.NoError(t, err)
	err = app.BankKeeper.SendCoinsFromModuleToAccount(ctx, minttypes.ModuleName, delAddr, sdk.NewCoins(delegationCoin))
	require.NoError(t, err)
	_, err = app.PetrichorKeeper.Delegate(ctx, delAddr, val1, delegationCoin)
	require.NoError(t, err)

	// Redelegate
	_, err = app.PetrichorKeeper.Redelegate(ctx, delAddr, val1, val2, sdk.NewCoin(PETRICHOR_TOKEN_DENOM, sdk.NewInt(500_000_000)))
	require.NoError(t, err)

	// Undelegate
	_, err = app.PetrichorKeeper.Undelegate(ctx, delAddr, val1, sdk.NewCoin(PETRICHOR_TOKEN_DENOM, sdk.NewInt(500_000_000)))
	require.NoError(t, err)

	// Trigger update asset
	ctx = ctx.WithBlockTime(ctx.BlockTime().Add(time.Hour * 25)).WithBlockHeight(ctx.BlockHeight() + 1)
	err = app.PetrichorKeeper.UpdatePetrichorAsset(ctx, types.NewPetrichorAsset(PETRICHOR_TOKEN_DENOM, sdk.MustNewDecFromStr("0.5"), sdk.ZeroDec(), ctx.BlockTime()))
	require.NoError(t, err)

	genesisState := app.PetrichorKeeper.ExportGenesis(ctx)
	require.NotNil(t, genesisState.Params)
	require.Greater(t, len(genesisState.Assets), 0)
	require.Greater(t, len(genesisState.ValidatorInfos), 0)
	require.Greater(t, len(genesisState.Delegations), 0)
	require.Greater(t, len(genesisState.Undelegations), 0)
	require.Greater(t, len(genesisState.Redelegations), 0)
	require.Greater(t, len(genesisState.RewardWeightChangeSnaphots), 0)

	store := ctx.KVStore(app.PetrichorKeeper.StoreKey())
	iter := store.Iterator(nil, nil)

	// Init a new app
	app, ctx = createTestContext(t)
	ctx = ctx.WithBlockTime(time.Now()).WithBlockHeight(1)

	app.PetrichorKeeper.InitGenesis(ctx, genesisState)

	// Check all items in the petrichor store match
	iter2 := store.Iterator(nil, nil)
	for ; iter.Valid(); iter.Next() {
		require.Equal(t, iter.Key(), iter2.Key())
		require.Equal(t, iter.Value(), iter2.Value())
		iter2.Next()
	}
}
