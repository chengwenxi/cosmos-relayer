# cosmos-relayer

Relayer for Cosmos Ecosystem

## Install

```bash
go mod tidy && make install
```

## Run

Command

```bash
relayer start \
    [chain-id-a] [node-a] [name-a] [password-a] [home-a] [client-id-a] \
    [chain-id-b] [node-b] [name-b] [password-b] [home-b] [client-id-b]
```

Example

```bash
relayer start \
    "iris" "tcp://localhost:26657" "n0" "12345678" "ibc-iris/n0/iriscli/" "client-to-gaia" \
    "cosmos" "tcp://localhost:26557" "n0" "12345678" "ibc-gaia/n0/gaiacli/" "client-to-iris"
```
