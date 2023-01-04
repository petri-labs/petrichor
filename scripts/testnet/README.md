# Scripts

This folder contains a sequence of helper scripts for creating an petrichor on testnet and automatic delegation. 

1. **gov.sh** submits a gov.json governance proposal, votes in favor of it and then queries the created petrichor.
2. **delegate.sh** delegates to the previously create petrichor and queries the modified petrichor.
3. **rewards.sh** claims available rewards and retrieves information about the process

> Note that these scripts must be executed in the specified order since they have dependencies on each other.
