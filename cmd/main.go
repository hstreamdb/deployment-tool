package main

import (
	cmpt "github.com/hstreamdb/deployment-tool/cmd/component"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"os"
)

func main() {
	var rootCmd = &cobra.Command{
		Use: `hdt <command | component> [args...]`,
		Args: func(cmd *cobra.Command, args []string) error {
			return nil
		},
		Short: "Deploy HStreamDB cluster.",
		//SilenceErrors: true,
		SilenceUsage:       true,
		DisableFlagParsing: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) < 1 {
				return cmd.Help()
			}

			switch args[0] {
			case "-h", "--help":
				return cmd.Help()
			case "server":
				cmd := cmpt.NewServerCmd()
				if err := cmd.Execute(); err != nil {
					log.Error(err)
					os.Exit(1)
				}
				return nil
			}

			return nil
		},
	}
	log.SetFormatter(&log.TextFormatter{})
	log.SetOutput(rootCmd.OutOrStdout())
	log.SetLevel(log.InfoLevel)

	rootCmd.AddCommand(
		newInit(),
		newDeploy(),
		newRemove(),
		newStop(),
	)

	if err := rootCmd.Execute(); err != nil {
		log.Error(err)
		os.Exit(1)
	}
}
