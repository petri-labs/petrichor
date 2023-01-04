#!/bin/bash

COIN_DENOM=ibc/4627AD2524E3E0523047E35BB76CC90E37D9D57ACF14F0FCBCEB2480705F3CB8

printf "#1) Submit proposal to create $COIN_DENOM Petrichor...\n\n"
petrichord tx gov submit-legacy-proposal create-petrichor $COIN_DENOM 0.5 0.5 --from=aztestval --node=tcp://3.75.187.158:26657 --chain-id=petrichor-testnet-1 --broadcast-mode=block -y
PROPOSAL_ID=$(petrichord query gov proposals --count-total --output json --node=tcp://3.75.187.158:26657 --chain-id=petrichor-testnet-1 | jq .proposals[0].id -r)

printf "\n#2) Deposit funds to proposal $PROPOSAL_ID...\n\n"
petrichord tx gov deposit $PROPOSAL_ID 10000000stake --from=aztestval --node=tcp://3.75.187.158:26657 --chain-id=petrichor-testnet-1 --broadcast-mode=block -y

printf "\n#3) Vote to pass the proposal...\n\n"
petrichord tx gov vote $PROPOSAL_ID yes --from=aztestval --node=tcp://3.75.187.158:26657 --chain-id=petrichor-testnet-1 --broadcast-mode=block -y
petrichord tx gov vote $PROPOSAL_ID yes --from=aztestval2 --node=tcp://3.75.187.158:26657 --chain-id=petrichor-testnet-1 --broadcast-mode=block -y
petrichord tx gov vote $PROPOSAL_ID yes --from=aztestval3 --node=tcp://3.75.187.158:26657 --chain-id=petrichor-testnet-1 --broadcast-mode=block -y

printf "\n#4) Query proposals...\n\n"
petrichord query gov proposal $PROPOSAL_ID --node=tcp://3.75.187.158:26657 --chain-id=petrichor-testnet-1

printf "\n#5) Query petrichors...\n\n"
petrichord query petrichor petrichors --node=tcp://3.75.187.158:26657 --chain-id=petrichor-testnet-1

printf "\n#6) Waiting for gov proposal to pass...\n\n"
sleep 8

printf "\n#7) Query petrichors after proposal passed...\n\n"
petrichord query petrichor petrichors --node=tcp://3.75.187.158:26657 --chain-id=petrichor-testnet-1