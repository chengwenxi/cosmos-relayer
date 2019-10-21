package main

import (
	"github.com/chengwenxi/cosmos-relayer/relayer"
	"os"

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
		Use:     "add [chain0] [node0] [chain1] [node1]",
		Short:   "Add a replayer for two blockchains",
		Args:    cobra.ExactArgs(4),
		Example: "relayer add chain0 tcp://localhost:26557 chain1 tcp://localhost:26657",
		RunE: func(cmd *cobra.Command, args []string) error {
			relayer.NewRelayer(args[0], args[1], args[2], args[3])
			return nil
		},
	}
}
