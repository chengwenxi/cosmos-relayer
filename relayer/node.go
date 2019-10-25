package relayer

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/x/ibc"
	"github.com/cosmos/cosmos-sdk/x/ibc/02-client/types/tendermint"
	"github.com/cosmos/cosmos-sdk/x/ibc/23-commitment/merkle"
	"github.com/tendermint/tendermint/types"

	"github.com/chengwenxi/cosmos-relayer/config"
	"github.com/cosmos/cosmos-sdk/client"
	ctx "github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/client/flags"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/auth/client/utils"
	ics04 "github.com/cosmos/cosmos-sdk/x/ibc/04-channel/types"
	bankmock "github.com/cosmos/cosmos-sdk/x/ibc/mock/bank"
	"github.com/tendermint/tendermint/libs/log"
)

type Node struct {
	ctx.CLIContext
	auth.TxBuilder
	Passphrase           string
	ClientId             string
	ChannelId            string
	CounterpartyClientId string
	logger               log.Logger
	prefix               config.Bech32Prefix
}

func NewNode(chainId, node, name, passphrase, home string) (*Node, error) {
	var cdc = makeCodec()
	cliCtx := ctx.NewCLIContext().
		WithCodec(cdc).
		WithBroadcastMode(flags.BroadcastBlock)

	keyBase, err := client.NewKeyBaseFromDir(home)
	if err != nil {
		return &Node{}, err
	}

	info, err := keyBase.Get(name)
	if err != nil {
		return &Node{}, err
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

	return &Node{
		CLIContext: cliCtx,
		TxBuilder:  builder,
		Passphrase: passphrase,
	}, nil

}

func (n *Node) WithLogger(logger log.Logger) *Node {
	n.logger = logger
	return n
}
func (n *Node) WithClientId(clientId string) *Node {
	n.ClientId = clientId
	return n
}
func (n *Node) WithChannelId(channelId string) *Node {
	n.ChannelId = channelId
	return n
}

func (n *Node) WithCounterpartyClientId(clientId string) *Node {
	n.CounterpartyClientId = clientId
	return n
}
func (n *Node) WithPrefix(prefix config.Bech32Prefix) *Node {
	n.prefix = prefix
	return n
}

func (n Node) SendTx(msgs []sdk.Msg) error {
	n.resetPrefix()
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

	n.logger.Info("Relay packet success", "targetChain", n.CLIContext.ChainID, "height", res.Height, "txHash", res.TxHash)
	return nil
}

func (n Node) GetHeader(h int64) (header tendermint.Header, err error) {
	client := n.Client

	commit, err := client.Commit(&h)
	if err != nil {
		n.logger.Error("Get commit error", "error", err.Error())
		return
	}

	prevHeight := h - 1
	validators, err := client.Validators(&prevHeight)
	if err != nil {
		n.logger.Error("Get prev validators error", "error", err.Error(), "height", prevHeight)
		return
	}

	nextValidators, err := client.Validators(&h)
	if err != nil {
		n.logger.Error("Get validators error", "error", err.Error(), "height", h)
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

func (n *Node) resetPrefix() {
	config := sdk.GetConfig()
	config.SetBech32PrefixForAccount(n.prefix.AccountAddr, n.prefix.AccountPub)
	config.SetBech32PrefixForValidator(n.prefix.ValidatorAddr, n.prefix.ValidatorPub)
	config.SetBech32PrefixForConsensusNode(n.prefix.ConsensusAddr, n.prefix.ConsensusPub)
}

func makeCodec() *codec.Codec {
	var cdc = codec.New()

	ibc.AppModuleBasic{}.RegisterCodec(cdc)
	sdk.RegisterCodec(cdc)
	auth.RegisterCodec(cdc)
	bankmock.RegisterCdc(cdc)
	codec.RegisterCrypto(cdc)
	codec.RegisterEvidences(cdc)

	return cdc.Seal()
}
