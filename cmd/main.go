package main

import (
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"os"
)

func main() {
	var rootCmd = &cobra.Command{
		Use:   "hdt",
		Short: "Deploy HStreamDB cluster.",
	}
	log.SetFormatter(&log.TextFormatter{})
	log.SetOutput(rootCmd.OutOrStdout())
	log.SetLevel(log.InfoLevel)

	rootCmd.AddCommand(
		newInit(),
		newDeploy(),
		newRemove(),
	)

	if err := rootCmd.Execute(); err != nil {
		log.Error(err)
		os.Exit(1)
	}
}
