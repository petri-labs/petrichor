package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"time"
)

const (
	ProposalTypeCreatePetrichor = "msg_create_petrichor_proposal"
	ProposalTypeUpdatePetrichor = "msg_update_petrichor_proposal"
	ProposalTypeDeletePetrichor = "msg_delete_petrichor_proposal"
)

var (
	_ govtypes.Content = &MsgCreatePetrichorProposal{}
	_ govtypes.Content = &MsgUpdatePetrichorProposal{}
	_ govtypes.Content = &MsgDeletePetrichorProposal{}
)

func init() {
	govtypes.RegisterProposalType(ProposalTypeCreatePetrichor)
	govtypes.RegisterProposalType(ProposalTypeUpdatePetrichor)
	govtypes.RegisterProposalType(ProposalTypeDeletePetrichor)
}
func NewMsgCreatePetrichorProposal(title, description, denom string, rewardWeight, takeRate sdk.Dec, rewardChangeRate sdk.Dec, rewardChangeInterval time.Duration) govtypes.Content {
	return &MsgCreatePetrichorProposal{
		Title:                title,
		Description:          description,
		Denom:                denom,
		RewardWeight:         rewardWeight,
		TakeRate:             takeRate,
		RewardChangeRate:     rewardChangeRate,
		RewardChangeInterval: rewardChangeInterval,
	}
}
func (m *MsgCreatePetrichorProposal) GetTitle() string       { return m.Title }
func (m *MsgCreatePetrichorProposal) GetDescription() string { return m.Description }
func (m *MsgCreatePetrichorProposal) ProposalRoute() string  { return RouterKey }
func (m *MsgCreatePetrichorProposal) ProposalType() string   { return ProposalTypeCreatePetrichor }

func (m *MsgCreatePetrichorProposal) ValidateBasic() error {

	if m.Denom == "" {
		return status.Errorf(codes.InvalidArgument, "Petrichor denom must have a value")
	}

	if m.RewardWeight.IsNil() || m.RewardWeight.LTE(sdk.ZeroDec()) {
		return status.Errorf(codes.InvalidArgument, "Petrichor rewardWeight must be a positive number")
	}

	if m.TakeRate.IsNil() || m.TakeRate.IsNegative() || m.TakeRate.GTE(sdk.OneDec()) {
		return status.Errorf(codes.InvalidArgument, "Petrichor takeRate must be more or equals to 0 but strictly less than 1")
	}

	if m.RewardChangeRate.IsZero() || m.RewardChangeRate.IsNegative() {
		return status.Errorf(codes.InvalidArgument, "Petrichor rewardChangeRate must be strictly a positive number")
	}

	return nil
}

func NewMsgUpdatePetrichorProposal(title, description, denom string, rewardWeight, takeRate sdk.Dec, rewardChangeRate sdk.Dec, rewardChangeInterval time.Duration) govtypes.Content {
	return &MsgUpdatePetrichorProposal{
		Title:                title,
		Description:          description,
		Denom:                denom,
		RewardWeight:         rewardWeight,
		TakeRate:             takeRate,
		RewardChangeRate:     rewardChangeRate,
		RewardChangeInterval: rewardChangeInterval,
	}
}
func (m *MsgUpdatePetrichorProposal) GetTitle() string       { return m.Title }
func (m *MsgUpdatePetrichorProposal) GetDescription() string { return m.Description }
func (m *MsgUpdatePetrichorProposal) ProposalRoute() string  { return RouterKey }
func (m *MsgUpdatePetrichorProposal) ProposalType() string   { return ProposalTypeUpdatePetrichor }

func (m *MsgUpdatePetrichorProposal) ValidateBasic() error {
	if m.Denom == "" {
		return status.Errorf(codes.InvalidArgument, "Petrichor denom must have a value")
	}

	if m.RewardWeight.IsNil() || m.RewardWeight.LTE(sdk.ZeroDec()) {
		return status.Errorf(codes.InvalidArgument, "Petrichor rewardWeight must be a positive number")
	}

	if m.TakeRate.IsNil() || m.TakeRate.IsNegative() || m.TakeRate.GTE(sdk.OneDec()) {
		return status.Errorf(codes.InvalidArgument, "Petrichor takeRate must be more or equals to 0 but strictly less than 1")
	}

	if m.RewardChangeRate.IsZero() || m.RewardChangeRate.IsNegative() {
		return status.Errorf(codes.InvalidArgument, "Petrichor rewardChangeRate must be strictly a positive number")
	}

	return nil
}

func NewMsgDeletePetrichorProposal(title, description, denom string) govtypes.Content {
	return &MsgDeletePetrichorProposal{
		Title:       title,
		Description: description,
		Denom:       denom,
	}
}
func (m *MsgDeletePetrichorProposal) GetTitle() string       { return m.Title }
func (m *MsgDeletePetrichorProposal) GetDescription() string { return m.Description }
func (m *MsgDeletePetrichorProposal) ProposalRoute() string  { return RouterKey }
func (m *MsgDeletePetrichorProposal) ProposalType() string   { return ProposalTypeDeletePetrichor }

func (m *MsgDeletePetrichorProposal) ValidateBasic() error {
	if m.Denom == "" {
		return status.Errorf(codes.InvalidArgument, "Petrichor denom must have a value")
	}
	return nil
}
