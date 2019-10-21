package iris

import (
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/auth/client/utils"
	"github.com/cosmos/cosmos-sdk/x/ibc"
	mockbank "github.com/cosmos/cosmos-sdk/x/ibc/mock/bank"
)

func SendTx(node string, msg sdk.Msg, name, passphrase string) error {
	cdc := MakeCodec()
	txBldr := auth.NewTxBuilderFromCLI().WithTxEncoder(utils.GetTxEncoder(cdc))
	ctx := context.NewCLIContext().WithCodec(cdc).WithBroadcastMode(flags.BroadcastBlock)

	homeDir := "/Users/bianjie/ibc-testnets/ibc-b/n0/iriscli"
	ctx = ctx.WithNodeURI(node).
		WithChainID("chain-b").
		WithFromName(name).WithFromAddress(msg.GetSigners()[0])

	ctx.HomeDir = homeDir
	ctx.OutputFormat = "text"

	keyBase, err := client.NewKeyBaseFromDir(homeDir)
	if err != nil {
		return err
	}

	txBldr = txBldr.WithChainID("chain-b").WithKeybase(keyBase)

	err = CompleteAndBroadcastTxCLI(txBldr, ctx, []sdk.Msg{msg}, passphrase)
	return err
}

func MakeCodec() *codec.Codec {
	var cdc = codec.New()

	ibc.AppModuleBasic{}.RegisterCodec(cdc)
	sdk.RegisterCodec(cdc)
	auth.RegisterCodec(cdc)
	mockbank.RegisterCdc(cdc)
	codec.RegisterCrypto(cdc)
	codec.RegisterEvidences(cdc)

	return cdc.Seal()
}

func CompleteAndBroadcastTxCLI(txBldr auth.TxBuilder, cliCtx context.CLIContext, msgs []sdk.Msg, passphrase string) error {
	txBldr, err := utils.PrepareTxBuilder(txBldr, cliCtx)
	if err != nil {
		return err
	}

	fromName := cliCtx.GetFromName()

	if cliCtx.Simulate {
		return nil
	}

	// build and sign the transaction
	txBytes, err := txBldr.BuildAndSign(fromName, passphrase, msgs)
	if err != nil {
		return err
	}

	// broadcast to a Tendermint node
	res, err := cliCtx.BroadcastTx(txBytes)
	if err != nil {
		return err
	}

	return cliCtx.PrintOutput(res)
}
