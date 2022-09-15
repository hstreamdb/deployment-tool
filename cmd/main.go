package main

import (
	"fmt"
	"github.com/spf13/cobra"
	"os"
)

func main() {
	// 1. parse config -> [Instance]
	// 2. [Instance] -> [DeployTasks]
	// 3. deploy [DeployTasks]

	var rootCmd = &cobra.Command{
		Use: "dev-deploy",
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
