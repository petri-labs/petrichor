package simulation

import (
	"fmt"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/tendermint/tendermint/libs/json"
	"github.com/petrinetwork/petrichor/x/petrichor/types"
	"math/rand"
	"time"
)

func genRewardDelayTime(r *rand.Rand) time.Duration {
	return time.Duration(simulation.RandIntBetween(r, 60, 60*60*24*3*2)) * time.Second
}

func genTakeRateClaimInterval(r *rand.Rand) time.Duration {
	return time.Duration(simulation.RandIntBetween(r, 1, 60*60)) * time.Second
}

func genNumOfPetrichorAssets(r *rand.Rand) int {
	return simulation.RandIntBetween(r, 0, 50)
}

func RandomizedGenesisState(simState *module.SimulationState) {
	var (
		rewardDelayTime     time.Duration
		rewardClaimInterval time.Duration
		numOfPetrichorAssets int
	)

	r := simState.Rand
	rewardDelayTime = genRewardDelayTime(r)
	rewardClaimInterval = genTakeRateClaimInterval(r)
	numOfPetrichorAssets = genNumOfPetrichorAssets(r)

	var petrichorAssets []types.PetrichorAsset
	for i := 0; i < numOfPetrichorAssets; i += 1 {
		rewardRate := simulation.RandomDecAmount(r, sdk.NewDec(5))
		takeRate := simulation.RandomDecAmount(r, sdk.MustNewDecFromStr("0.5"))
		startTime := time.Now().Add(time.Duration(simulation.RandIntBetween(r, 60, 60*60*24*3*2)) * time.Second)
		petrichorAssets = append(petrichorAssets, types.NewPetrichorAsset(fmt.Sprintf("ASSET%d", i), rewardRate, takeRate, startTime))
	}

	petrichorGenesis := types.GenesisState{
		Params: types.Params{
			RewardDelayTime:       rewardDelayTime,
			TakeRateClaimInterval: rewardClaimInterval,
			LastTakeRateClaimTime: simState.GenTimestamp,
		},
		Assets: petrichorAssets,
	}

	bz, err := json.MarshalIndent(&petrichorGenesis, "", " ")
	if err != nil {
		panic(err)
	}
	fmt.Printf("Selected randomly generated petrichor parameters:\n%s\n", bz)

	simState.GenState[types.ModuleName] = simState.Cdc.MustMarshalJSON(&petrichorGenesis)
}
