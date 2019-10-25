package relayer

import (
	"fmt"
	"github.com/cosmos/cosmos-sdk/x/ibc/02-client/types/tendermint"
	"github.com/cosmos/cosmos-sdk/x/ibc/23-commitment/merkle"
	"github.com/tendermint/tendermint/types"
	"strings"

	"github.com/chengwenxi/cosmos-relayer/chains/config"
	"github.com/cosmos/cosmos-sdk/client"
	ctx "github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/client/flags"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/auth/client/utils"
	ics04 "github.com/cosmos/cosmos-sdk/x/ibc/04-channel/types"
	bankmock "github.com/cosmos/cosmos-sdk/x/ibc/mock/bank"
)

type Node struct {
	ctx.CLIContext
	auth.TxBuilder
	Passphrase     string
	Id             string
	CounterpartyId string
}

func NewNode(chainId, node, name, passphrase, home, id, counterpartyId string) (Node, error) {
	var cdc = makeCodec()
	cliCtx := ctx.NewCLIContext().
		WithCodec(cdc).
		WithBroadcastMode(flags.BroadcastBlock)

	keyBase, err := client.NewKeyBaseFromDir(home)
	if err != nil {
		return Node{}, err
	}

	info, err := keyBase.Get(name)
	if err != nil {
		return Node{}, err
	}

	cliCtx = cliCtx.WithNodeURI(node).
		WithChainID(chainId).
		WithFromName(name).
		WithFromAddress(info.GetAddress())

	cliCtx.OutputFormat = "json"

	builder := auth.NewTxBuilderFromCLI().
		WithTxEncoder(utils.GetTxEncoder(cdc)).
		WithChainID(chainId).
		WithKeybase(keyBase)

	return Node{
		CLIContext:     cliCtx,
		TxBuilder:      builder,
		Passphrase:     passphrase,
		Id:             id,
		CounterpartyId: counterpartyId,
	}, nil

}

func (n Node) SendTx(msgs []sdk.Msg) error {
	txBldr, err := utils.PrepareTxBuilder(n.TxBuilder, n.CLIContext)
	if err != nil {
		return err
	}
	fromName := n.GetFromName()

	if n.Simulate {
		return nil
	}

	// build and sign the transaction
	txBytes, err := txBldr.BuildAndSign(fromName, n.Passphrase, msgs)
	if err != nil {
		return err
	}

	// broadcast to a Tendermint node
	res, err := n.BroadcastTx(txBytes)
	if err != nil {
		return err
	}

	return n.PrintOutput(res)
}

func (n Node) GetHeader(h int64) (header tendermint.Header, err error) {
	client := n.Client

	commit, err := client.Commit(&h)
	if err != nil {
		fmt.Println(fmt.Errorf("get commit error: %v", err.Error()))
		return
	}

	prevHeight := h - 1
	validators, err := client.Validators(&prevHeight)
	if err != nil {
		fmt.Println(fmt.Errorf("get commit error: %v", err.Error()))
		return
	}

	nextValidators, err := client.Validators(&h)
	if err != nil {
		fmt.Println(fmt.Errorf("get validators error: %v", err.Error()))
		return
	}
	return tendermint.Header{
		SignedHeader:     commit.SignedHeader,
		ValidatorSet:     types.NewValidatorSet(validators.Validators),
		NextValidatorSet: types.NewValidatorSet(nextValidators.Validators),
	}, nil
}

func (n Node) GetProof(packet bankmock.Packet, h int64) (merkle.Proof, error) {
	key := append([]byte("channels/"), ics04.KeyPacketCommitment(packet.MsourcePort, packet.MsourceChannel, packet.Msequence)...)
	proof, err := n.QueryStoreProof(key, "ibc", h)
	return merkle.Proof{Proof: proof, Key: key}, err
}

func (n Node) LoadConfig() {
	if strings.Contains(n.CLIContext.ChainID, config.Cosmos) {
		config.LoadConfig(config.Cosmos)
	} else {
		//config.SetNetworkType(config.Testnet)
		config.LoadConfig(config.Iris)
	}
}
