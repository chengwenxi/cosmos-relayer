package relayer

import (
	"context"
	"fmt"

	"github.com/chengwenxi/cosmos-relayer/chains/iris"
	sdk "github.com/cosmos/cosmos-sdk/types"
	ics23 "github.com/cosmos/cosmos-sdk/x/ibc/23-commitment"
	"github.com/cosmos/cosmos-sdk/x/ibc/23-commitment/merkle"
	bankmock "github.com/cosmos/cosmos-sdk/x/ibc/mock/bank"
	"github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/rpc/client"
)

type Relayer struct {
	chain0       string
	node0        string
	node0Address sdk.AccAddress
	chain1       string
	node1        string
	node1Address sdk.AccAddress
}

// NewRelayer returns a relayer which provide the service for one to one blockchain
func NewRelayer(chain0, node0, chain1, node1 string) Relayer {
	return Relayer{
		chain0: chain0,
		node0:  node0,
		chain1: chain1,
		node1:  node1,
	}
}

func (r Relayer) Start() error {
	c0 := client.NewHTTP(r.node0, "/websocket")
	c1 := client.NewHTTP(r.node1, "/websocket")

	out0, err := c0.Subscribe(context.Background(), r.chain0, "")
	if err != nil {
		return err
	}

	go func() {
		for {
			select {
			case resultEvent := <-out0:
				println(resultEvent.Query)
			}
		}
	}()

	out1, err := c1.Subscribe(context.Background(), r.chain0, "")
	if err != nil {
		return err
	}

	go func() {
		for {
			select {
			case resultEvent := <-out1:
				println(resultEvent.Query)
			}
		}
	}()

	return nil
}

func (r Relayer) handleEvent(events []types.Event, height uint64) {
	for _, e := range events {
		switch e.Type {
		case "send_packet":
			r.handlePacket(e, height)
		default:
		}

	}
}

func (r Relayer) handlePacket(event types.Event, height uint64) {
	for _, ab := range event.Attributes {
		switch string(ab.Key) {
		case "Packet":
			r.sendPacket(ab.Value, height)
		}
	}
}

func (r Relayer) sendPacket(packetBz []byte, height uint64) {
	var packet bankmock.Packet

	if err := packet.UnmarshalJSON(packetBz); err != nil {
		fmt.Println(fmt.Errorf("error unmarshalling packet: %v", packetBz))
		return
	}

	proof := merkle.Proof{}
	msg := bankmock.NewMsgRecvTransferPacket(packet, []ics23.Proof{proof}, height, r.node1Address)
	if err := msg.ValidateBasic(); err != nil {
		fmt.Println(fmt.Errorf("err recv packet msg: %v", err.ABCILog()))
		return
	}

	err := iris.SendTx(r.node1, msg, "n1", "12345678")
	if err != nil {
		fmt.Println(fmt.Errorf("broadcast tx error: %v", err))
		return
	}
}
