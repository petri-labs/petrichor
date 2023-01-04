package petrichor

import (
	"github.com/petrinetwork/petrichor/x/petrichor/keeper"
	"github.com/petrinetwork/petrichor/x/petrichor/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
)

func NewPetrichorProposalHandler(k keeper.Keeper) govtypes.Handler {
	return func(ctx sdk.Context, content govtypes.Content) error {
		switch c := content.(type) {
		case *types.MsgCreatePetrichorProposal:
			return k.CreatePetrichor(ctx, c)
		case *types.MsgUpdatePetrichorProposal:
			return k.UpdatePetrichor(ctx, c)
		case *types.MsgDeletePetrichorProposal:
			return k.DeletePetrichor(ctx, c)

		default:
			return sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "unrecognized petrichor proposal content type: %T", c)
		}
	}
}
