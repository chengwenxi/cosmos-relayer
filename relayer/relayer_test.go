package relayer

import (
	"context"
	"testing"

	"github.com/chengwenxi/cosmos-relayer/chains/iris"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/rpc/client"
	"github.com/tendermint/tendermint/types"
)

func Test_Subscribe(t *testing.T) {
	c := client.NewHTTP("tcp://localhost:26657", "/websocket")

	err := c.Start()
	require.Nil(t, err)

	out0, err := c.Subscribe(context.Background(), "iris", types.EventQueryTx.String())

	//out0, err := c.Subscribe(context.Background(), "iris", types.QueryForEvent("message").String())

	assert.NoError(t, err)

	iris.LoadConfig()

	accountAddress, err := sdk.AccAddressFromBech32("iaa1ndspclqujqeunuxypkkt287ttvjg9qnpcned7t")

	assert.NoError(t, err)

	relayer := Relayer{
		node1Address: accountAddress,
		node1:        "tcp://localhost:26557",
		chain1:       "chain-b",
	}

	go func() {
		for {
			select {
			case resultEvent := <-out0:
				println(resultEvent.Query)
				txResult := resultEvent.Data.(types.EventDataTx)
				relayer.handleEvent(txResult.Result.Events, uint64(txResult.Height))
			}
		}
	}()

	select {}
}
