#!/bin/bash

DEMO_WALLET_ADDRESS=$(petrichord --home ./data/petrichor keys show demowallet1 --keyring-backend test -a)
VAL_ADDR=$(petrichord query staking validators --output json | jq .validators[0].operator_address --raw-output)
COIN_DENOM=upetrix
COIN_AMOUNT=$(petrichord query petrichor delegation $DEMO_WALLET_ADDRESS $VAL_ADDR $COIN_DENOM --home ./data/petrichor --output json | jq .delegation.balance.amount --raw-output | sed 's/\.[0-9]*//')
COINS=$COIN_AMOUNT$COIN_DENOM

# FIX: failed to execute message; message index: 0: invalid shares amount: invalid
printf "#1) Undelegate 5000000000$COIN_DENOM from x/petrichor $COIN_DENOM...\n\n"
petrichord tx petrichor undelegate $VAL_ADDR $COINS --from=demowallet1 --home ./data/petrichor --keyring-backend=test --broadcast-mode=block --gas 1000000 -y > /dev/null 2>&1

printf "\n#2) Query delegations from x/petrichor $COIN_DENOM...\n\n"
petrichord query petrichor petrichor $COIN_DENOM

printf "\n#3) Query delegation on x/petrichor by delegator, validator and $COIN_DENOM...\n\n"
petrichord query petrichor delegation $DEMO_WALLET_ADDRESS $VAL_ADDR $COIN_DENOM --home ./data/petrichor
