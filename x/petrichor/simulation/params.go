package simulation

import (
	"fmt"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/x/simulation"
	"github.com/petrinetwork/petrichor/x/petrichor/types"
	"math/rand"
)

func ParamChanges(r *rand.Rand) []simtypes.ParamChange {
	return []simtypes.ParamChange{
		simulation.NewSimParamChange(types.ModuleName, string(types.RewardDelayTime),
			func(r *rand.Rand) string {
				return fmt.Sprintf("\"%d\"", genRewardDelayTime(r))
			},
		),
		simulation.NewSimParamChange(types.ModuleName, string(types.TakeRateClaimInterval),
			func(r *rand.Rand) string {
				return fmt.Sprintf("\"%d\"", genTakeRateClaimInterval(r))
			},
		),
	}
}
