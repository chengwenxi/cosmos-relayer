package config

import (
	"io/ioutil"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

type RelayConfig struct {
	Title string `toml:"title"`
	Nodes []Node `toml:"nodes"`
}

type Node struct {
	ChainId              string       `toml:"chain_id"`
	Address              string       `toml:"address"`
	Escrow               Escrow       `toml:"escrow"`
	ClientId             string       `toml:"client_id"`
	ChannelId            string       `toml:"channel_id"`
	CounterpartyClientId string       `toml:"counterparty_client_id"`
	Bech32Prefix         Bech32Prefix `toml:"bech32_prefix"`
}

type Escrow struct {
	Name       string `toml:"name"`
	Passphrase string `toml:"passphrase"`
	Home       string `toml:"home"`
}

type Bech32Prefix struct {
	AccountAddr   string `toml:"account_addr"`
	AccountPub    string `toml:"account_pub"`
	ValidatorAddr string `toml:"validator_addr"`
	ValidatorPub  string `toml:"validator_pub"`
	ConsensusAddr string `toml:"consensus_addr"`
	ConsensusPub  string `toml:"consensus_pub"`
}

func Load(path string) *RelayConfig {
	var cfg *RelayConfig
	cfgFilePath := filepath.Join(path, fileNm)
	if _, err := toml.DecodeFile(cfgFilePath, &cfg); err != nil {
		panic(err)
	}
	return cfg
}

func Write(path string) error {
	cfgFilePath := filepath.Join(path, fileNm)
	err := ioutil.WriteFile(cfgFilePath, []byte(template), 0666)
	return err
}

const fileNm = "relay.toml"
const template = `
# This is a TOML config file.
# For more information, see https://github.com/toml-lang/toml

##### main base config options #####

# Relay server name
title = "Realy Configuration Example"

# Chain node configuration
[[nodes]]

# the chain id
chain_id = "chain-iris"
# the chain node rpc address
address = "tcp://localhost:26657"
# IBC client id
client_id = "client-to-gaia"
# IBC channel id
channel_id = "chann-to-gaia"
# IBC counterparty client id
counterparty_client_id = "client-to-iris"

# relay node escrow account
[nodes.escrow]

# escrow account name
name = "n0"
# escrow account password for sign transaction
passphrase = "12345678"
# escrow account location
home = "ibc-iris/n0/iriscli/"

[nodes.bech32_prefix]
account_addr = "iaa"
validator_addr = "iva"
consensus_addr = "ica"
account_pub = "iap"
validator_pub = "ivp"
consensus_pub = "icp"

#############other node configuration#####################

# Chain node configuration
[[nodes]]

# the chain id
chain_id = "chain-gaia"
# the chain node rpc address
address = "tcp://localhost:26557"
# IBC client id
client_id = "client-to-iris"
# IBC channel id
channel_id = "chann-to-iris"
# IBC counterparty client id
counterparty_client_id = "client-to-gaia"

# relay node escrow account
[nodes.escrow]

# escrow account name
name = "n0"
# escrow account password for sign transaction
passphrase = "12345678"
# escrow account location
home = "ibc-gaia/n0/gaiacli/"

[nodes.bech32_prefix]
account_addr = "cosmos"
account_pub = "cosmospub"
validator_addr = "cosmosvaloper"
validator_pub = "cosmosvaloperpub"
consensus_addr = "cosmosvalcons"
consensus_pub = "cosmosvalconspub"

`
