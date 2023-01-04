package keeper

import (
	"context"
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/petrinetwork/petrichor/x/petrichor/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (k Keeper) CreatePetrichor(ctx context.Context, req *types.MsgCreatePetrichorProposal) error {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	_, found := k.GetAssetByDenom(sdkCtx, req.Denom)

	if found {
		return status.Errorf(codes.AlreadyExists, "Asset with denom: %s already exists", req.Denom)
	}

	rewardStartTime := sdkCtx.BlockTime().Add(k.RewardDelayTime(sdkCtx))
	asset := types.PetrichorAsset{
		Denom:                req.Denom,
		RewardWeight:         req.RewardWeight,
		TakeRate:             req.TakeRate,
		TotalTokens:          sdk.ZeroInt(),
		TotalValidatorShares: sdk.ZeroDec(),
		RewardStartTime:      rewardStartTime,
		RewardChangeRate:     req.RewardChangeRate,
		RewardChangeInterval: req.RewardChangeInterval,
		LastRewardChangeTime: rewardStartTime,
	}
	k.SetAsset(sdkCtx, asset)
	return nil
}

func (k Keeper) UpdatePetrichor(ctx context.Context, req *types.MsgUpdatePetrichorProposal) error {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	asset, found := k.GetAssetByDenom(sdkCtx, req.Denom)

	if !found {
		return status.Errorf(codes.NotFound, "Asset with denom: %s does not exist", req.Denom)
	}

	asset.RewardWeight = req.RewardWeight
	asset.TakeRate = req.TakeRate
	asset.RewardChangeRate = req.RewardChangeRate
	asset.RewardChangeInterval = req.RewardChangeInterval

	err := k.UpdatePetrichorAsset(sdkCtx, asset)
	if err != nil {
		return err
	}

	return nil
}

func (k Keeper) DeletePetrichor(ctx context.Context, req *types.MsgDeletePetrichorProposal) error {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	asset, found := k.GetAssetByDenom(sdkCtx, req.Denom)

	if !found {
		return status.Errorf(codes.NotFound, "Asset with denom: %s does not exist", req.Denom)
	}

	if asset.TotalTokens.GT(math.ZeroInt()) {
		return status.Errorf(codes.Internal, "Asset cannot be deleted because there are still %s delegations associated with it", asset.TotalTokens)
	}

	k.DeleteAsset(sdkCtx, req.Denom)

	return nil
}
