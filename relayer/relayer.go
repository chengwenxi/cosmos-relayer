package relayer

import (
	"context"
	"encoding/hex"
	"fmt"
	"os"
	"strings"

	"github.com/chengwenxi/cosmos-relayer/config"

	sdk "github.com/cosmos/cosmos-sdk/types"
	ics02 "github.com/cosmos/cosmos-sdk/x/ibc/02-client/types"
	ics04 "github.com/cosmos/cosmos-sdk/x/ibc/04-channel/types"
	ics23 "github.com/cosmos/cosmos-sdk/x/ibc/23-commitment"
	bankmock "github.com/cosmos/cosmos-sdk/x/ibc/mock/bank"
	abciTypes "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/log"
	"github.com/tendermint/tendermint/types"
)

type Relayer struct {
	nodes  map[string]*Node
	logger log.Logger
}

func NewRelayerFromCfgFile(cfg *config.RelayConfig) Relayer {
	var nodes = make(map[string]*Node, len(cfg.Nodes))
	logger := log.NewTMLogger(log.NewSyncWriter(os.Stdout))
	for _, n := range cfg.Nodes {
		node, err := NewNode(n.ChainId, n.Address, n.Escrow.Name, n.Escrow.Passphrase, n.Escrow.Home)
		if err != nil {
			panic("Init node failed")
		}
		node.WithLogger(logger).
			WithClientId(n.ClientId).
			WithChannelId(n.ChannelId).
			WithCounterpartyClientId(n.CounterpartyClientId).
			WithPrefix(n.Bech32Prefix)
		nodes[n.ClientId] = node
	}

	return Relayer{
		nodes:  nodes,
		logger: logger,
	}
}

func (r Relayer) GetNode(clientId string) *Node {
	return r.nodes[clientId]
}

func (r Relayer) Start() {
	for _, node := range r.nodes {
		if err := r.addTask(node); err != nil {
			panic(err)
		}
	}
	r.logger.Info("Relayer start success")
	select {}
}

func (r Relayer) addTask(node *Node) error {
	counterpartyNode := r.GetNode(node.CounterpartyClientId)
	client := counterpartyNode.Client
	if err := client.Start(); err != nil {
		return err
	}

	subscriber := fmt.Sprintf("%s->%s", counterpartyNode.CLIContext.ChainID, node.CLIContext.ChainID)

	out, err := client.Subscribe(context.Background(), subscriber, types.EventQueryTx.String())
	if err != nil {
		return err
	}
	go func() {
		for resultEvent := range out {
			data := resultEvent.Data.(types.EventDataTx)
			r.handleEvent(node, data)
		}
	}()
	r.logger.Info("Subscribe event start", "subscriber", subscriber, "clientId", counterpartyNode.ClientId, "channelId", counterpartyNode.ChannelId)
	return nil
}

func (r Relayer) handleEvent(node *Node, data types.EventDataTx) {
	for _, e := range data.Result.Events {
		switch e.Type {
		case ics04.EventTypeSendPacket:
			counterpartyNode := r.GetNode(node.CounterpartyClientId)
			txHash := strings.ToUpper(hex.EncodeToString(data.Tx.Hash()))
			r.logger.Info("Listened transaction", "sourceChain", counterpartyNode.CLIContext.ChainID, "height", data.Height, "txHash", txHash)
			r.handlePacket(node, e, data.Height)
		default:
		}

	}
}

func (r Relayer) handlePacket(node *Node, event abciTypes.Event, height int64) {
	for _, ab := range event.Attributes {
		switch string(ab.Key) {
		case ics04.AttributeKeyPacket:
			r.sendPacket(node, ab.Value, height)
		}
	}
}

func (r Relayer) sendPacket(node *Node, packetBz []byte, height int64) {
	var packet bankmock.Packet
	counterpartyNode := r.GetNode(node.CounterpartyClientId)

	if err := packet.UnmarshalJSON(packetBz); err != nil {
		r.logger.Error("UnmarshalJSON packet error", "error", err.Error())
		return
	}

	r.logger.Info("Received packet", "sourceChain", counterpartyNode.CLIContext.ChainID, "packet", string(packetBz), "data", string(packet.Mdata))

	r.waitForHeight(counterpartyNode, height+1)

	header, err := counterpartyNode.GetHeader(height + 1)
	if err != nil {
		return
	}
	msgUpdateClient := ics02.NewMsgUpdateClient(node.ClientId, header, node.FromAddress)

	proof, err := counterpartyNode.GetProof(packet, height)
	if err != nil {
		return
	}

	msg := bankmock.NewMsgRecvTransferPacket(packet, []ics23.Proof{proof}, uint64(height+1), node.FromAddress)
	if err := msg.ValidateBasic(); err != nil {
		r.logger.Error("Validate msg error", "error", err.ABCILog())
		return
	}

	err = node.SendTx([]sdk.Msg{msgUpdateClient, msg})
	if err != nil {
		r.logger.Error("Broadcast tx error", "targetChain", node.CLIContext.ChainID, "errMsg", err.Error())
		return
	}
}

func (r Relayer) waitForHeight(node *Node, height int64) {
	client := node.Client

	ctx := context.Background()
	subscriber := fmt.Sprintf("subscriber-height-%d", height)
	query := types.EventQueryNewBlock.String()

	out, err := client.Subscribe(ctx, subscriber, query)
	if err != nil {
		r.logger.Error("Subscriber block event error", "sourceChain", node.CLIContext.ChainID, "height", height)
		return
	}

	r.logger.Info("Waiting for block to get proof", "sourceChain", node.CLIContext.ChainID, "height", height)
	for event := range out {
		data := event.Data.(types.EventDataNewBlock)
		if data.Block.Height >= height {
			if err := client.Unsubscribe(ctx, subscriber, query); err != nil {
				r.logger.Error("Unsubscribe block event error", "height", height)

				return
			}
			break
		}
	}
}
