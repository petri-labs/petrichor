#!/usr/bin/make -f

localnet-petrichor-rmi:
	docker rmi terra-money/localnet-petrichor 2>/dev/null; true

localnet-build-env: localnet-petrichor-rmi
	docker build --tag terra-money/localnet-petrichor -f scripts/containers/Dockerfile \
    		$(shell git rev-parse --show-toplevel)
	
localnet-build-nodes:
	docker run --rm -v $(CURDIR)/.testnets:/petrichor terra-money/localnet-petrichor \
		testnet init-files --v 3 -o /petrichor --starting-ip-address 192.168.5.20 --keyring-backend=test --chain-id=petrichor-testnet-1
	docker-compose up -d

localnet-stop:
	docker-compose down

localnet-start: localnet-stop localnet-build-env localnet-build-nodes

.PHONY: localnet-start localnet-stop localnet-build-env localnet-build-nodes
