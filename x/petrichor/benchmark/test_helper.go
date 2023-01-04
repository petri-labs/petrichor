package benchmark

import (
	"fmt"
	"math/rand"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/x/staking/teststaking"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	test_helpers "github.com/petrinetwork/petrichor/app"
	"github.com/petrinetwork/petrichor/x/petrichor/types"
	"time"
)

func SetupApp(t *testing.T, r *rand.Rand, numAssets int, numValidators int, numDelegators int) (app *test_helpers.App, ctx sdk.Context, assets []types.PetrichorAsset, valAddrs []sdk.AccAddress, delAddrs []sdk.AccAddress) {
	app = test_helpers.Setup(t, false)
	ctx = app.BaseApp.NewContext(false, tmproto.Header{})
	startTime := time.Now()
	ctx = ctx.WithBlockTime(startTime)
	for i := 0; i < numAssets; i += 1 {
		rewardWeight := simulation.RandomDecAmount(r, sdk.NewDec(1))
		takeRate := simulation.RandomDecAmount(r, sdk.MustNewDecFromStr("0.0001"))
		asset := types.NewPetrichorAsset(fmt.Sprintf("ASSET%d", i), rewardWeight, takeRate, startTime)
		asset.RewardChangeRate = sdk.OneDec().Sub(simulation.RandomDecAmount(r, sdk.MustNewDecFromStr("0.00001")))
		asset.RewardChangeInterval = time.Minute * 5
		assets = append(assets, asset)
	}
	params := types.NewParams()
	params.TakeRateClaimInterval = time.Minute * 5
	app.PetrichorKeeper.InitGenesis(ctx, &types.GenesisState{
		Params: params,
		Assets: assets,
	})

	// Accounts
	valAddrs = test_helpers.AddTestAddrsIncremental(app, ctx, numValidators, sdk.NewCoins())
	pks := test_helpers.CreateTestPubKeys(numValidators)

	for i := 0; i < numValidators; i += 1 {
		valAddr := sdk.ValAddress(valAddrs[i])
		_val := teststaking.NewValidator(t, valAddr, pks[i])
		_val.Commission = stakingtypes.Commission{
			CommissionRates: stakingtypes.CommissionRates{
				Rate:          sdk.NewDec(0),
				MaxRate:       sdk.NewDec(0),
				MaxChangeRate: sdk.NewDec(0),
			},
			UpdateTime: time.Now(),
		}
		_val.Status = stakingtypes.Bonded
		test_helpers.RegisterNewValidator(t, app, ctx, _val)
	}

	delAddrs = test_helpers.AddTestAddrsIncremental(app, ctx, numDelegators, sdk.NewCoins())
	return
}

func GenerateOperationSlots(operations ...int) func(r *rand.Rand) int {
	var slots []int
	for i := 0; i < len(operations); i += 1 {
		for o := 0; o < operations[i]; o += 1 {
			slots = append(slots, i)
		}
	}
	return func(r *rand.Rand) int {
		return slots[r.Intn(len(slots)-1)]
	}
}
