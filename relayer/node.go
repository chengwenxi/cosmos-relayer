package relayer

import (
	"github.com/chengwenxi/cosmos-relayer/chains/config"
	"github.com/cosmos/cosmos-sdk/client"
	ctx "github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/auth/client/utils"
	"github.com/cosmos/cosmos-sdk/x/ibc"
	mockbank "github.com/cosmos/cosmos-sdk/x/ibc/mock/bank"
	"strings"
)

type Node struct {
	Ctx        ctx.CLIContext
	Builder    auth.TxBuilder
	cdc        *codec.Codec
	Passphrase string
}

func NewNode(chainId, node, name, passphrase, home string) (Node, error) {
	var cdc = makeCodec()
	ctx := ctx.NewCLIContext().WithCodec(cdc).WithBroadcastMode(flags.BroadcastBlock)

	keyBase, err := client.NewKeyBaseFromDir(home)
	if err != nil {
		return Node{}, err
	}

	info, err := keyBase.Get(name)
	if err != nil {
		return Node{}, err
	}

	ctx = ctx.WithNodeURI(node).
		WithChainID(chainId).
		WithFromName(name).WithFromAddress(info.GetAddress())

	ctx.OutputFormat = "text"

	builder := auth.NewTxBuilderFromCLI().WithTxEncoder(utils.GetTxEncoder(cdc))
	builder = builder.WithChainID(chainId).WithKeybase(keyBase)
	return Node{
		Ctx:        ctx,
		Builder:    builder,
		cdc:        cdc,
		Passphrase: passphrase,
	}, nil

}

func (n Node) SendTx(msgs []sdk.Msg) error {
	txBldr, err := utils.PrepareTxBuilder(n.Builder, n.Ctx)
	if err != nil {
		return err
	}

	fromName := n.Ctx.GetFromName()

	if n.Ctx.Simulate {
		return nil
	}

	// build and sign the transaction
	txBytes, err := txBldr.BuildAndSign(fromName, n.Passphrase, msgs)
	if err != nil {
		return err
	}

	// broadcast to a Tendermint node
	res, err := n.Ctx.BroadcastTx(txBytes)
	if err != nil {
		return err
	}

	return n.Ctx.PrintOutput(res)
}
func (n Node) LoadConfig() {
	if strings.Contains(n.Ctx.ChainID, config.Iris) {
		//iris.SetNetworkType(iris.Testnet)
		config.LoadConfig(config.Iris)
	} else {
		config.LoadConfig(config.Cosmos)
	}
}

func makeCodec() *codec.Codec {
	var cdc = codec.New()

	ibc.AppModuleBasic{}.RegisterCodec(cdc)
	sdk.RegisterCodec(cdc)
	auth.RegisterCodec(cdc)
	mockbank.RegisterCdc(cdc)
	codec.RegisterCrypto(cdc)
	codec.RegisterEvidences(cdc)

	return cdc.Seal()
}
