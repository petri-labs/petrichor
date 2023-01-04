package types

import (
	cosmosmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"time"
)

func NewPetrichorAsset(denom string, rewardWeight sdk.Dec, takeRate sdk.Dec, rewardStartTime time.Time) PetrichorAsset {
	return PetrichorAsset{
		Denom:                denom,
		RewardWeight:         rewardWeight,
		TakeRate:             takeRate,
		TotalTokens:          sdk.ZeroInt(),
		TotalValidatorShares: sdk.ZeroDec(),
		RewardStartTime:      rewardStartTime,
		RewardChangeRate:     sdk.OneDec(),
		RewardChangeInterval: time.Duration(0),
		LastRewardChangeTime: rewardStartTime,
	}
}

func ConvertNewTokenToShares(totalTokens sdk.Dec, totalShares sdk.Dec, newTokens cosmosmath.Int) (shares sdk.Dec) {
	if totalShares.IsZero() {
		return sdk.NewDecFromInt(newTokens)
	}
	return totalShares.Quo(totalTokens).MulInt(newTokens)
}

func ConvertNewShareToDecToken(totalTokens sdk.Dec, totalShares sdk.Dec, shares sdk.Dec) (token sdk.Dec) {
	if totalShares.IsZero() {
		return totalTokens
	}
	return shares.Quo(totalShares).Mul(totalTokens)
}

func GetDelegationTokens(del Delegation, val PetrichorValidator, asset PetrichorAsset) sdk.Coin {
	valTokens := val.TotalDecTokensWithAsset(asset)
	totalDelegationShares := val.TotalDelegationSharesWithDenom(asset.Denom)
	delTokens := ConvertNewShareToDecToken(valTokens, totalDelegationShares, del.Shares)
	return sdk.NewCoin(asset.Denom, delTokens.TruncateInt())
}

func GetDelegationSharesFromTokens(val PetrichorValidator, asset PetrichorAsset, token cosmosmath.Int) sdk.Dec {
	valTokens := val.TotalTokensWithAsset(asset)
	totalDelegationShares := val.TotalDelegationSharesWithDenom(asset.Denom)
	if totalDelegationShares.TruncateInt().Equal(sdk.ZeroInt()) {
		return sdk.NewDecFromInt(token)
	}
	return ConvertNewTokenToShares(valTokens, totalDelegationShares, token)
}

func GetValidatorShares(asset PetrichorAsset, token cosmosmath.Int) sdk.Dec {
	return ConvertNewTokenToShares(sdk.NewDecFromInt(asset.TotalTokens), asset.TotalValidatorShares, token)
}

func (a PetrichorAsset) HasPositiveDecay() bool {
	return a.RewardChangeInterval > 0 && a.RewardChangeRate.IsPositive()
}
