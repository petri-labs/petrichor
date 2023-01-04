package client

import (
	"github.com/petrinetwork/petrichor/x/petrichor/client/cli"

	govclient "github.com/cosmos/cosmos-sdk/x/gov/client"
)

var (
	CreatePetrichorProposalHandler = govclient.NewProposalHandler(cli.CreatePetrichor)
	UpdatePetrichorProposalHandler = govclient.NewProposalHandler(cli.UpdatePetrichor)
	DeletePetrichorProposalHandler = govclient.NewProposalHandler(cli.DeletePetrichor)
)
