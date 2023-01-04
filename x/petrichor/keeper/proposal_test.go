package keeper_test

import (
	"github.com/petrinetwork/petrichor/x/petrichor/keeper"
	"github.com/petrinetwork/petrichor/x/petrichor/types"
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	"github.com/stretchr/testify/require"
)

func TestCreatePetrichor(t *testing.T) {
	// GIVEN
	app, ctx := createTestContext(t)
	startTime := time.Now()
	ctx.WithBlockTime(startTime).WithBlockHeight(1)
	queryServer := keeper.NewQueryServerImpl(app.PetrichorKeeper)
	rewardDuration := app.PetrichorKeeper.RewardDelayTime(ctx)

	// WHEN
	createErr := app.PetrichorKeeper.CreatePetrichor(ctx, &types.MsgCreatePetrichorProposal{
		Title:        "",
		Description:  "",
		Denom:        "upetri",
		RewardWeight: sdk.OneDec(),
		TakeRate:     sdk.OneDec(),
	})
	petrichorsRes, petrichorsErr := queryServer.Petrichors(ctx, &types.QueryPetrichorsRequest{})

	// THEN
	require.Nil(t, createErr)
	require.Nil(t, petrichorsErr)
	require.Equal(t, petrichorsRes, &types.QueryPetrichorsResponse{
		Petrichors: []types.PetrichorAsset{
			{
				Denom:                "upetri",
				RewardWeight:         sdk.NewDec(1),
				TakeRate:             sdk.NewDec(1),
				TotalTokens:          sdk.ZeroInt(),
				TotalValidatorShares: sdk.NewDec(0),
				RewardStartTime:      ctx.BlockTime().Add(rewardDuration),
				RewardChangeRate:     sdk.NewDec(0),
				RewardChangeInterval: 0,
				LastRewardChangeTime: ctx.BlockTime().Add(rewardDuration),
			},
		},
		Pagination: &query.PageResponse{
			NextKey: nil,
			Total:   1,
		},
	})
}

func TestCreatePetrichorFailWithDuplicatedDenom(t *testing.T) {
	// GIVEN
	app, ctx := createTestContext(t)
	startTime := time.Now()
	ctx.WithBlockTime(startTime).WithBlockHeight(1)
	app.PetrichorKeeper.InitGenesis(ctx, &types.GenesisState{
		Params: types.DefaultParams(),
		Assets: []types.PetrichorAsset{
			types.NewPetrichorAsset("upetri", sdk.NewDec(1), sdk.NewDec(0), startTime),
		},
	})

	// WHEN
	createErr := app.PetrichorKeeper.CreatePetrichor(ctx, &types.MsgCreatePetrichorProposal{
		Title:        "",
		Description:  "",
		Denom:        "upetri",
		RewardWeight: sdk.OneDec(),
		TakeRate:     sdk.OneDec(),
	})

	// THEN
	require.Error(t, createErr)
}

func TestUpdatePetrichor(t *testing.T) {
	// GIVEN
	app, ctx := createTestContext(t)
	startTime := time.Now()
	ctx.WithBlockTime(startTime).WithBlockHeight(1)
	app.PetrichorKeeper.InitGenesis(ctx, &types.GenesisState{
		Params: types.DefaultParams(),
		Assets: []types.PetrichorAsset{
			{
				Denom:                "upetri",
				RewardWeight:         sdk.NewDec(2),
				TakeRate:             sdk.OneDec(),
				TotalTokens:          sdk.ZeroInt(),
				TotalValidatorShares: sdk.NewDec(0),
			},
		},
	})
	queryServer := keeper.NewQueryServerImpl(app.PetrichorKeeper)

	// WHEN
	updateErr := app.PetrichorKeeper.UpdatePetrichor(ctx, &types.MsgUpdatePetrichorProposal{
		Title:                "",
		Description:          "",
		Denom:                "upetri",
		RewardWeight:         sdk.NewDec(6),
		TakeRate:             sdk.NewDec(7),
		RewardChangeInterval: 0,
		RewardChangeRate:     sdk.ZeroDec(),
	})
	petrichorsRes, petrichorsErr := queryServer.Petrichors(ctx, &types.QueryPetrichorsRequest{})

	// THEN
	require.Nil(t, updateErr)
	require.Nil(t, petrichorsErr)
	require.Equal(t, petrichorsRes, &types.QueryPetrichorsResponse{
		Petrichors: []types.PetrichorAsset{
			{
				Denom:                "upetri",
				RewardWeight:         sdk.NewDec(6),
				TakeRate:             sdk.NewDec(7),
				TotalTokens:          sdk.ZeroInt(),
				TotalValidatorShares: sdk.NewDec(0),
				RewardChangeRate:     sdk.NewDec(0),
				RewardChangeInterval: 0,
			},
		},
		Pagination: &query.PageResponse{
			NextKey: nil,
			Total:   1,
		},
	})
}

func TestDeletePetrichor(t *testing.T) {
	// GIVEN
	app, ctx := createTestContext(t)
	startTime := time.Now()
	ctx.WithBlockTime(startTime).WithBlockHeight(1)
	app.PetrichorKeeper.InitGenesis(ctx, &types.GenesisState{
		Params: types.DefaultParams(),
		Assets: []types.PetrichorAsset{
			{
				Denom:        "upetri",
				RewardWeight: sdk.NewDec(2),
				TakeRate:     sdk.OneDec(),
				TotalTokens:  sdk.ZeroInt(),
			},
		},
	})
	queryServer := keeper.NewQueryServerImpl(app.PetrichorKeeper)

	// WHEN
	deleteErr := app.PetrichorKeeper.DeletePetrichor(ctx, &types.MsgDeletePetrichorProposal{
		Denom: "upetri",
	})
	petrichorsRes, petrichorsErr := queryServer.Petrichors(ctx, &types.QueryPetrichorsRequest{})

	// THEN
	require.Nil(t, deleteErr)
	require.Nil(t, petrichorsErr)
	require.Equal(t, petrichorsRes, &types.QueryPetrichorsResponse{
		Petrichors: nil,
		Pagination: &query.PageResponse{
			NextKey: nil,
			Total:   0,
		},
	})
}
