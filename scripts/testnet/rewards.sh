#!/bin/bash

COIN_DENOM=ibc/4627AD2524E3E0523047E35BB76CC90E37D9D57ACF14F0FCBCEB2480705F3CB8
WALLET_ADDRESS=$(petrichord keys show aztestval -a)
VAL_ADDR=$(petrichord query staking validators --output json --node=tcp://3.75.187.158:26657 --chain-id=petrichor-testnet-1 | jq .validators[0].operator_address --raw-output)

printf "\n\n#1) Query $COIN_DENOM petrichor rewards...\n\n"
petrichord query petrichor rewards $WALLET_ADDRESS $VAL_ADDR $COIN_DENOM --node=tcp://3.75.187.158:26657 --chain-id=petrichor-testnet-1

printf "\n\n#2) Claim rewards from x/petrichor $COIN_DENOM...\n\n"
petrichord tx petrichor claim-rewards $VAL_ADDR $COIN_DENOM --from=aztestval --node=tcp://3.75.187.158:26657 --chain-id=petrichor-testnet-1 --gas=auto --broadcast-mode=block -y

printf "\n\n#3) Query $COIN_DENOM petrichor rewards after claim...\n\n"
petrichord query petrichor rewards $WALLET_ADDRESS $VAL_ADDR $COIN_DENOM --node=tcp://3.75.187.158:26657 --chain-id=petrichor-testnet-1
