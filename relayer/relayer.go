package relayer

import (
	"context"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	ics23 "github.com/cosmos/cosmos-sdk/x/ibc/23-commitment"
	"github.com/cosmos/cosmos-sdk/x/ibc/23-commitment/merkle"
	bankmock "github.com/cosmos/cosmos-sdk/x/ibc/mock/bank"
	abciTypes "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/rpc/client"
	"github.com/tendermint/tendermint/types"
)

type Relayer struct {
	Node1 Node
	Node2 Node
}

// NewRelayer returns a relayer which provide the service for one to one blockchain
func NewRelayer(node1, node2 Node) Relayer {
	return Relayer{
		Node1: node1,
		Node2: node2,
	}
}

func (r Relayer) Start() {
	var subscribe = func(toNode Node, remote, subscriber string) error {
		c := client.NewHTTP(remote, "/websocket")
		err := c.Start()
		if err != nil {
			return err
		}

		out, err := c.Subscribe(context.Background(), subscriber, types.EventQueryTx.String())
		if err != nil {
			return err
		}
		go func() {
			for resultEvent := range out {
				toNode.LoadConfig()
				data := resultEvent.Data.(types.EventDataTx)
				r.handleEvent(toNode, data.Result.Events, uint64(data.Height))
			}
		}()
		return nil
	}

	err := subscribe(r.Node2, r.Node1.Ctx.NodeURI, fmt.Sprintf("%s->%s", r.Node1.Ctx.ChainID, r.Node2.Ctx.ChainID))
	if err != nil {
		panic(err)
	}

	err = subscribe(r.Node1, r.Node2.Ctx.NodeURI, fmt.Sprintf("%s->%s", r.Node2.Ctx.ChainID, r.Node1.Ctx.ChainID))
	if err != nil {
		panic(err)
	}
	fmt.Println("Start relayer success")
	select {}
}

func (r Relayer) handleEvent(node Node, events []abciTypes.Event, height uint64) {
	for _, e := range events {
		switch e.Type {
		case "send_packet":
			r.handlePacket(node, e, height)
		default:
		}

	}
}

func (r Relayer) handlePacket(node Node, event abciTypes.Event, height uint64) {
	for _, ab := range event.Attributes {
		switch string(ab.Key) {
		case "Packet":
			r.sendPacket(node, ab.Value, height)
		}
	}
}

func (r Relayer) sendPacket(node Node, packetBz []byte, height uint64) {
	var packet bankmock.Packet

	if err := packet.UnmarshalJSON(packetBz); err != nil {
		fmt.Println(fmt.Errorf("error unmarshalling packet: %v", packetBz))
		return
	}

	proof := merkle.Proof{}
	msg := bankmock.NewMsgRecvTransferPacket(packet, []ics23.Proof{proof}, height, node.Ctx.FromAddress)
	if err := msg.ValidateBasic(); err != nil {
		fmt.Println(fmt.Errorf("err recv packet msg: %v", err.ABCILog()))
		return
	}

	err := node.SendTx([]sdk.Msg{msg})
	if err != nil {
		fmt.Println(fmt.Errorf("broadcast tx error: %v", err))
		return
	}
}
