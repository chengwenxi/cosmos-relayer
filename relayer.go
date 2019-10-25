package main

import (
	"fmt"
	"github.com/chengwenxi/cosmos-relayer/config"
	"github.com/spf13/viper"
	"os"

	"github.com/chengwenxi/cosmos-relayer/relayer"

	"github.com/spf13/cobra"
)

var FlagConfigDir = "home"

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
	rootCmd.AddCommand(initCmd())
	rootCmd.AddCommand(startCmd())
}

func initCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "init",
		Short:   "Generate relayer configuration",
		Example: `relayer init --home=./"`,
		RunE: func(cmd *cobra.Command, args []string) error {
			err := config.Write(viper.GetString(FlagConfigDir))
			if err != nil {
				fmt.Println("init error", err.Error())
			}
			return nil
		},
	}
	cmd.Flags().String(FlagConfigDir, "", "configuration file path")
	_ = viper.BindPFlag(FlagConfigDir, cmd.Flags().Lookup(FlagConfigDir))
	_ = cmd.MarkFlagRequired(FlagConfigDir)
	return cmd
}

func startCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "start",
		Short:   "Add a relayer for two blockchains",
		Example: `relayer start --config-dir=./"`,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg := config.Load(viper.GetString(FlagConfigDir))
			relay := relayer.NewRelayerFromCfgFile(cfg)
			relay.Start()
			return nil
		},
	}
	cmd.Flags().String(FlagConfigDir, "", "configuration file path")
	_ = viper.BindPFlag(FlagConfigDir, cmd.Flags().Lookup(FlagConfigDir))
	_ = cmd.MarkFlagRequired(FlagConfigDir)
	return cmd
}
