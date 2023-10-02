module github.com/chengwenxi/cosmos-relayer

go 1.13

require (
	github.com/BurntSushi/toml v0.3.1
	github.com/cosmos/cosmos-sdk v0.34.4-0.20191015214354-791b2454139d
	github.com/spf13/cobra v0.0.5
	github.com/spf13/viper v1.5.0
	github.com/tendermint/tendermint v0.32.10
)

replace github.com/cosmos/cosmos-sdk => github.com/irisnet/cosmos-sdk v0.23.2-0.20191024053222-fb9fd55110ea
