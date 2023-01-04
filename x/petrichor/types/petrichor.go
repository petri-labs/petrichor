package types

func NewRewardWeightChangeSnapshot(asset PetrichorAsset, val PetrichorValidator) RewardWeightChangeSnapshot {
	return RewardWeightChangeSnapshot{
		PrevRewardWeight: asset.RewardWeight,
		RewardHistories:  val.GlobalRewardHistory,
	}
}
