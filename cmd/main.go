package main

import (
	"fmt"
	"github.com/spf13/cobra"
	"os"
)

func main() {
	var rootCmd = &cobra.Command{
		Use:   "hdt",
		Short: "Deploy HStreamDB cluster.",
	}

	rootCmd.AddCommand(
		newInit(),
		newDeploy(),
		newRemove(),
	)

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
