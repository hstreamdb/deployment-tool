package component

import (
	"github.com/hstreamdb/deployment-tool/pkg/task"
	"github.com/spf13/cobra"
)

func NewServerCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "hdt server <command>",
		Short: "Manage HStream Server instance.",
		Args: func(cmd *cobra.Command, args []string) error {
			cmd.SetArgs(args[1:])
			return nil
		},
		SilenceErrors:         true,
		DisableFlagParsing:    true,
		SilenceUsage:          true,
		DisableFlagsInUseLine: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) <= 1 {
				return cmd.Help()
			}

			switch args[0] {
			case "-h", "--help":
				return cmd.Help()
			}
			return cmd.Execute()
		},
	}
	cmd.AddCommand(newStartServerCmd())
	cmd.AddCommand(newRemoveServerCmd())
	cmd.SetUsageTemplate(UsageTpl)

	return cmd
}

func newStartServerCmd() *cobra.Command {
	opts := commonOpts{}
	cmd := &cobra.Command{
		Use:           "start",
		Short:         "Start HStream Server Cluster",
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			services, executor, err := getServices(cmd, opts)
			if err != nil {
				return err
			}
			return task.SetUpHServerCluster(executor, services)
		},
	}

	cmd.Flags().StringVarP(&opts.configPath, "config", "c", "template/config.yaml", "Cluster config path.")
	cmd.Flags().StringVarP(&opts.user, "user", "u", "", "User name to login via ssh.")
	cmd.Flags().BoolVarP(&opts.usePassword, "use-password", "p", false, "Use password authentication for ssh.")
	cmd.Flags().StringVarP(&opts.identityFile, "identity-file", "i", "", "The path of the SSH identity file.")
	cmd.Flags().BoolVarP(&opts.debugMode, "debug", "d", false, "Debug mode")
	return cmd
}

func newRemoveServerCmd() *cobra.Command {
	opts := commonOpts{}
	cmd := &cobra.Command{
		Use:           "remove",
		Short:         "Remove HStream Server Cluster",
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			services, executor, err := getServices(cmd, opts)
			if err != nil {
				return err
			}
			return task.RemoveHServerCluster(executor, services)
		},
	}

	cmd.Flags().StringVarP(&opts.configPath, "config", "c", "template/config.yaml", "Cluster config path.")
	cmd.Flags().StringVarP(&opts.user, "user", "u", "", "User name to login via ssh.")
	cmd.Flags().BoolVarP(&opts.usePassword, "use-password", "p", false, "Use password authentication for ssh.")
	cmd.Flags().StringVarP(&opts.identityFile, "identity-file", "i", "", "The path of the SSH identity file.")
	cmd.Flags().BoolVarP(&opts.debugMode, "debug", "d", false, "Debug mode")
	return cmd
}
