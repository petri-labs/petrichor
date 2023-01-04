#!/bin/bash

CHAIN_DIR=./data
CHAINID=${CHAINID:-petrichor}
COIN_DENOM=upetrix
VAL_WALLET_ADDRESS=$(petrichord --home $CHAIN_DIR/$CHAINID keys show demowallet1 --keyring-backend test -a)
VAL_ADDR=$(petrichord query staking validators --output json | jq .validators[0].operator_address --raw-output)

printf "#1) Delegate 10000000000$COIN_DENOM thru x/petrichor $COIN_DENOM...\n\n"
petrichord tx petrichor delegate $VAL_ADDR 10000000000$COIN_DENOM --from=demowallet1 --home $CHAIN_DIR/$CHAINID --keyring-backend=test --broadcast-mode=block --gas 1000000 -y > /dev/null 2>&1

printf "\n#2) Query delegations from x/petrichor $COIN_DENOM...\n\n"
petrichord query petrichor petrichor $COIN_DENOM

printf "\n#3) Query delegation on x/petrichor by delegator, validator and $COIN_DENOM...\n\n"
petrichord query petrichor delegation $VAL_WALLET_ADDRESS $VAL_ADDR $COIN_DENOM
