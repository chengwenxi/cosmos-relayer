# cosmos-relayer

Relayer for Cosmos Ecosystem

## Install

```bash
go mod tidy && make install
```

## Run

1. Generate relayer configuration
```bash
relayer init --home=./
``` 


2. Modify the configuration file generated in the first step,and start relayer
```bash
relayer start --home=./
```
