# cosmos-relayer
Relayer for Cosmos Ecosystem

### Install
go mod tidy && make install

### Run
```bash
  relayer start [chainId-a] [node-a] [name-a] [password-a] [home-a] [client-id-a] [chainId-b] [node-b] [name-b] [password-a] [home-b] [client-id-b]

# example
relayer start "chain-iris" "tcp://localhost:26657" "n0" "12345678" "ibc-handshake/ibc-iris/n0/iriscli/" "client-to-gaia" "chain-gaia" "tcp://localhost:26557" "n0" "12345678" "ibc-handshake/ibc-gaia/n0/gaiacli/" "client-to-iris"
```