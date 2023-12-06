package main

import (
	"errors"
	cmpt "github.com/hstreamdb/deployment-tool/cmd/component"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"golang.org/x/exp/slices"
	"os"
)

var (
	hdtCmd = []string{"init", "start", "stop", "remove"}
)

func main() {
	var rootCmd = &cobra.Command{
		Use: `hdt <command> [args...]
 hdt <component> [args...]`,
		Args: func(cmd *cobra.Command, args []string) error {
			// Support usage of `hdt <component>`
			return nil
		},
		Short:                 "Deploy HStreamDB cluster.",
		SilenceErrors:         true,
		SilenceUsage:          true,
		DisableFlagParsing:    true,
		DisableFlagsInUseLine: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) < 1 {
				return cmd.Help()
			}

			switch args[0] {
			case "-h", "--help":
				return cmd.Help()
			case "server":
				cmd = cmpt.NewServerCmd()
			case "console":
				cmd = cmpt.NewConsoleCmd()
			case "hstream-exporter":
				cmd = cmpt.NewHStreamExporterCmd()
			default:
				if !slices.Contains(hdtCmd, args[0]) {
					return errors.New("unknown command")
				}
			}

			if err := cmd.Execute(); err != nil {
				log.Error(err)
				os.Exit(1)
			}
			return nil
		},
	}

	log.SetFormatter(&log.TextFormatter{
		FullTimestamp: true,
	})
	log.SetReportCaller(true)
	log.SetOutput(rootCmd.OutOrStdout())
	log.SetLevel(log.InfoLevel)

	rootCmd.AddCommand(
		newInit(),
		newDeploy(),
		newRemove(),
		newStop(),
	)

	rootCmd.SetUsageTemplate(cmpt.UsageTpl)
	if err := rootCmd.Execute(); err != nil {
		log.Error(err)
		os.Exit(1)
	}
}
