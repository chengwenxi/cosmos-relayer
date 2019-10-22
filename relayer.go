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
		Use:     "start [chainId-a] [node-a] [name-a] [passphrase-a] [home-a] [chainId-b] [node-b] [name-b] [passphrase-b] [home-b]",
		Short:   "Add a replayer for two blockchains",
		Args:    cobra.ExactArgs(8),
		Example: `relayer start "chain-a" "tcp://localhost:26657" "n0" "ibc-testnets/ibc-a/n0/iriscli/" "chain-b" "tcp://localhost:26557" "n1" "ibc-testnets/ibc-b/n0/iriscli/"`,
		RunE: func(cmd *cobra.Command, args []string) error {
			pass1, err := relayer.ReadPassphraseFromStdin(args[2])
			if err != nil {
				fmt.Println(err)
				return err
			}
			node0, err := relayer.NewNode(args[0], args[1], args[2], pass1, args[3])
			if err != nil {
				fmt.Println(err)
				return err
			}

			pass2, err := relayer.ReadPassphraseFromStdin(args[6])
			if err != nil {
				fmt.Println(err)
				return err
			}
			node1, err := relayer.NewNode(args[4], args[5], args[6], pass2, args[7])
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
