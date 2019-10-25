package relayer

import (
	"context"
	"fmt"
	sdk "github.com/cosmos/cosmos-sdk/types"
	ics02 "github.com/cosmos/cosmos-sdk/x/ibc/02-client/types"
	ics23 "github.com/cosmos/cosmos-sdk/x/ibc/23-commitment"
	bankmock "github.com/cosmos/cosmos-sdk/x/ibc/mock/bank"
	abciTypes "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/types"
)

type Relayer struct {
	nodes map[string]Node
}

// NewRelayer returns a relayer which provide the service for one to one blockchain
func NewRelayer(node1, node2 Node) Relayer {
	nodes := map[string]Node{
		node1.Id: node1,
		node2.Id: node2,
	}
	return Relayer{nodes}
}

func (r Relayer) GetNode(id string) Node {
	return r.nodes[id]
}

func (r Relayer) Start() {
	var subscribe = func(toNode Node) error {
		counterpartyNode := r.GetNode(toNode.CounterpartyId)
		client := counterpartyNode.Client
		if err := client.Start(); err != nil {
			return err
		}

		subscriber := fmt.Sprintf("%s->%s", counterpartyNode.CLIContext.ChainID, toNode.CLIContext.ChainID)

		out, err := client.Subscribe(context.Background(), subscriber, types.EventQueryTx.String())
		if err != nil {
			return err
		}
		go func() {
			for resultEvent := range out {
				toNode.LoadConfig()
				data := resultEvent.Data.(types.EventDataTx)
				r.handleEvent(toNode, data)
			}
		}()
		return nil
	}

	for _, node := range r.nodes {
		if err := subscribe(node); err != nil {
			panic(err)
		}
	}
	fmt.Println("Start relayer success")
	select {}
}

func (r Relayer) handleEvent(node Node, data types.EventDataTx) {
	for _, e := range data.Result.Events {
		switch e.Type {
		case "send_packet":
			counterpartyNode := r.GetNode(node.CounterpartyId)
			println(fmt.Sprintf("listening for transactions[hash=%X] from the %s chain", data.Tx.Hash(), counterpartyNode.CLIContext.ChainID))
			r.handlePacket(node, e, data.Height)
		default:
		}

	}
}

func (r Relayer) handlePacket(node Node, event abciTypes.Event, height int64) {
	for _, ab := range event.Attributes {
		switch string(ab.Key) {
		case "Packet":
			r.sendPacket(node, ab.Value, height)
		}
	}
}

func (r Relayer) sendPacket(node Node, packetBz []byte, height int64) {
	var packet bankmock.Packet

	println(fmt.Sprintf("receive packet: %s", string(packetBz)))
	if err := packet.UnmarshalJSON(packetBz); err != nil {
		fmt.Println(fmt.Errorf("error unmarshalling packet: %v", packetBz))
		return
	}
	counterpartyNode := r.GetNode(node.CounterpartyId)

	waitForHeight(counterpartyNode, height+1)

	header, err := counterpartyNode.GetHeader(height + 1)
	if err != nil {
		return
	}
	msgUpdateClient := ics02.NewMsgUpdateClient(node.Id, header, node.FromAddress)

	proof, err := counterpartyNode.GetProof(packet, height)
	if err != nil {
		return
	}

	msg := bankmock.NewMsgRecvTransferPacket(packet, []ics23.Proof{proof}, uint64(height+1), node.FromAddress)
	if err := msg.ValidateBasic(); err != nil {
		fmt.Println(fmt.Errorf("err recv packet msg: %v", err.ABCILog()))
		return
	}

	err = node.SendTx([]sdk.Msg{msgUpdateClient, msg})
	if err != nil {
		fmt.Println(fmt.Errorf("broadcast tx error: %v", err))
		return
	}
}

func waitForHeight(node Node, height int64) {
	client := node.Client

	ctx := context.Background()
	subscriber := fmt.Sprintf("subscriber-height-%d", height)
	query := types.EventQueryNewBlock.String()

	out, err := client.Subscribe(ctx, subscriber, query)
	if err != nil {
		fmt.Println(fmt.Errorf("failed subscriber : %s,%v", subscriber, err))
		return
	}

	fmt.Println(fmt.Sprintf("waitting for block : %d", height))
	for event := range out {
		data := event.Data.(types.EventDataNewBlock)
		if data.Block.Height >= height {
			if err := client.Unsubscribe(ctx, subscriber, query); err != nil {
				fmt.Println(fmt.Errorf("failed unsubscribe : %s,%v", subscriber, err))
				return
			}
			break
		}
	}
}
