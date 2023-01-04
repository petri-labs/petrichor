#!/bin/bash

printf "#1) Submit proposal to delete the upetrix Petrichor...\n\n"
petrichord tx gov submit-legacy-proposal delete-petrichor upetrix --from=demowallet1 --home ./data/petrichor --keyring-backend=test --broadcast-mode=block --gas 1000000 -y > /dev/null 2>&1

PROPOSAL_ID=$(petrichord query gov proposals --count-total --output json --home ./data/petrichor | jq .pagination.total -r)

printf "\n#2) Deposit funds to proposal $PROPOSAL_ID...\n\n"
petrichord tx gov deposit $PROPOSAL_ID 10000000stake --from=demowallet1 --home ./data/petrichor --keyring-backend=test --broadcast-mode=block --gas 1000000 -y > /dev/null 2>&1

printf "\n#3) Vote to pass the proposal...\n\n"
petrichord tx gov vote $PROPOSAL_ID yes --from=val1 --home ./data/petrichor --keyring-backend=test --broadcast-mode=block --gas 1000000 -y > /dev/null 2>&1

printf "\n#4) Query proposals...\n\n"
petrichord query gov proposal $PROPOSAL_ID --home ./data/petrichor

printf "\n#5) Query petrichors...\n\n"
petrichord query petrichor petrichors --home ./data/petrichor

printf "\n#6) Waiting for gov proposal to pass...\n\n"
sleep 8

printf "\n#7) Query petrichors after proposal passed...\n\n"
petrichord query petrichor petrichors --home ./data/petrichor