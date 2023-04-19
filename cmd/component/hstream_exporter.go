package component

import (
	"github.com/hstreamdb/deployment-tool/pkg/task"
	"github.com/spf13/cobra"
)

func NewHStreamExporterCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "hstream-exporter",
		Short: "Manage HStream Exporter instance.",
		Args: func(cmd *cobra.Command, args []string) error {
			cmd.SetArgs(args[1:])
			return nil
		},
		//SilenceErrors:      true,
		DisableFlagParsing: true,
		SilenceUsage:       true,
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
	cmd.AddCommand(newStartExporterCmd())
	cmd.AddCommand(newRemoveExporterCmd())

	return cmd
}

func newStartExporterCmd() *cobra.Command {
	opts := commonOpts{}
	cmd := &cobra.Command{
		Use:          "start",
		Short:        "Start HStream Exporter Cluster",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			services, executor, err := getServices(cmd, opts)
			if err != nil {
				return err
			}
			return task.SetUpHStreamExporterService(executor, services)
		},
	}

	cmd.Flags().StringVarP(&opts.configPath, "config", "c", "template/config.yaml", "Cluster config path.")
	cmd.Flags().StringVarP(&opts.user, "user", "u", "", "User name to login via ssh.")
	cmd.Flags().BoolVarP(&opts.usePassword, "use-password", "p", false, "Use password authentication for ssh.")
	cmd.Flags().StringVarP(&opts.identityFile, "identity-file", "i", "", "The path of the SSH identity file.")
	cmd.Flags().BoolVarP(&opts.debugMode, "debug", "d", false, "Debug mode")
	return cmd
}

func newRemoveExporterCmd() *cobra.Command {
	opts := commonOpts{}
	cmd := &cobra.Command{
		Use:          "remove",
		Short:        "Remove HStream Exporter Cluster",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			services, executor, err := getServices(cmd, opts)
			if err != nil {
				return err
			}
			return task.RemoveHStreamExporterService(executor, services)
		},
	}

	cmd.Flags().StringVarP(&opts.configPath, "config", "c", "template/config.yaml", "Cluster config path.")
	cmd.Flags().StringVarP(&opts.user, "user", "u", "", "User name to login via ssh.")
	cmd.Flags().BoolVarP(&opts.usePassword, "use-password", "p", false, "Use password authentication for ssh.")
	cmd.Flags().StringVarP(&opts.identityFile, "identity-file", "i", "", "The path of the SSH identity file.")
	cmd.Flags().BoolVarP(&opts.debugMode, "debug", "d", false, "Debug mode")
	return cmd
}
