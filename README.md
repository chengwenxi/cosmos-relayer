# cosmos-relayer
Relayer for Cosmos Ecosystem

### Install
go mod tidy && make install

### Run
```bash
relayer start [chainId-a] [node-a] [name-a] [passphrase-a] [home-a] [chainId-b] [node-b] [name-b] [passphrase-b] [home-b]

# example
relayer start "chain-a" "tcp://localhost:26657" "n0" "12345678" "ibc-testnets/ibc-a/n0/iriscli/" "chain-b" "tcp://localhost:26557" "n1" "12345678" "ibc-testnets/ibc-b/n0/iriscli/"
```