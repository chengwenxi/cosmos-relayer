package main

import (
	"fmt"
	"os"

	"github.com/chengwenxi/cosmos-relayer/relayer"

	"github.com/spf13/cobra"
)

func main() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

var rootCmd = &cobra.Command{
	Use:          "relayer",
	Short:        "Relayer service which relays ibc messages between multi-cosmos blockchains",
	SilenceUsage: true,
}

func init() {
	rootCmd.AddCommand(addRelayerCmd())
}

func addRelayerCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "start [chainId-a] [node-a] [name-a] [password-a] [home-a] [client-id-a] [chainId-b] [node-b] [name-b] [password-b] [home-b] [client-id-b]",
		Short:   "Add a replayer for two blockchains",
		Args:    cobra.ExactArgs(12),
		Example: `relayer start "chain-a" "tcp://localhost:26657" "n0" "ibc-testnets/ibc-a/n0/iriscli/" "chain-b" "tcp://localhost:26557" "n1" "ibc-testnets/ibc-b/n0/iriscli/"`,
		RunE: func(cmd *cobra.Command, args []string) error {

			node0ChainId := args[0]
			node0Url := args[1]
			node0Name := args[2]
			node0Password := args[3]
			node0Home := args[4]
			node0ClientId := args[5]

			node1ChainId := args[6]
			node1Url := args[7]
			node1Name := args[8]
			node1Password := args[9]
			node1Home := args[10]
			node1ClientId := args[11]

			node0, err := relayer.NewNode(node0ChainId, node0Url, node0Name, node0Password, node0Home, node0ClientId, node1ClientId)
			if err != nil {
				fmt.Println(err)
				return err
			}

			node1, err := relayer.NewNode(node1ChainId, node1Url, node1Name, node1Password, node1Home, node1ClientId, node0ClientId)
			if err != nil {
				fmt.Println(err)
				return err
			}
			relayer := relayer.NewRelayer(node0, node1)
			relayer.Start()
			return nil
		},
	}
}
