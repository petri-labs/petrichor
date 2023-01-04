#!/bin/bash

VAL_WALLET_ADDRESS=$(petrichord --home ./data/petrichor keys show val1 --keyring-backend test -a)
DEMO_WALLET_ADDRESS=$(petrichord --home ./data/petrichor keys show demowallet1 --keyring-backend test -a)

VAL_ADDR=$(petrichord query staking validators --output json | jq .validators[0].operator_address --raw-output)
TOKEN_DENOM=upetrix

printf "\n\n#1) Query wallet balances...\n\n"
petrichord query bank balances $DEMO_WALLET_ADDRESS --home ./data/petrichor

#printf "\n\n#2) Query rewards x/petrichor...\n\n"
#petrichord query petrichor rewards $DEMO_WALLET_ADDRESS $VAL_ADDR $TOKEN_DENOM --home ./data/petrichor

#printf "\n\n#3) Query native staked rewards...\n\n"
#petrichord query distribution rewards $DEMO_WALLET_ADDRESS $VAL_ADDR --home ./data/petrichor

#printf "\n\n#4) Claim rewards from validator...\n\n"
#petrichord tx distribution withdraw-rewards $VAL_ADDR --from=demowallet1 --home ./data/petrichor --keyring-backend=test --broadcast-mode=block --gas 1000000 -y #> /dev/null 2>&1

printf "\n\n#2) Claim rewards from x/petrichor $TOKEN_DENOM...\n\n"
petrichord tx petrichor claim-rewards $VAL_ADDR $TOKEN_DENOM --from=demowallet1 --home ./data/petrichor --keyring-backend=test --broadcast-mode=block --gas 1000000 -y > /dev/null 2>&1

#printf "\n\n#6) Query rewards x/petrichor...\n\n"
#petrichord query petrichor rewards $DEMO_WALLET_ADDRESS $VAL_ADDR $TOKEN_DENOM --home ./data/petrichor

#printf "\n\n#7) Query native staked rewards...\n\n"
#petrichord query distribution rewards $DEMO_WALLET_ADDRESS $VAL_ADDR --home ./data/petrichor

printf "\n\n#3) Query wallet balances after claim...\n\n"
petrichord query bank balances $DEMO_WALLET_ADDRESS --home ./data/petrichor
