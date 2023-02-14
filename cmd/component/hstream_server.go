package component

import (
	ext "github.com/hstreamdb/deployment-tool/pkg/executor"
	"github.com/hstreamdb/deployment-tool/pkg/service"
	"github.com/hstreamdb/deployment-tool/pkg/spec"
	"github.com/hstreamdb/deployment-tool/pkg/task"
	"github.com/hstreamdb/deployment-tool/pkg/utils"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
	"os"
)

type serverOpts struct {
	user         string
	usePassword  bool
	identityFile string
	configPath   string
	debugMode    bool
}

func NewServerCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "server",
		Short: "Manage HStream Server instance.",
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
	cmd.AddCommand(newStartServerCmd())
	cmd.AddCommand(newRemoveServerCmd())

	return cmd
}

func newStartServerCmd() *cobra.Command {
	opts := serverOpts{}
	cmd := &cobra.Command{
		Use:          "start",
		Short:        "Start HStream Server Cluster",
		SilenceUsage: true,
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
	opts := serverOpts{}
	cmd := &cobra.Command{
		Use:          "remove",
		Short:        "Remove HStream Server Cluster",
		SilenceUsage: true,
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

func getServices(cmd *cobra.Command, opts serverOpts) (*service.Services, ext.Executor, error) {
	var (
		executor ext.Executor
		services *service.Services
	)
	contents, err := os.ReadFile(opts.configPath)
	log.Debugf("opts: %+v\n", opts)
	if err != nil {
		return nil, nil, err
	}

	config := spec.ComponentsSpec{}
	if err = yaml.Unmarshal(contents, &config); err != nil {
		return nil, nil, err
	}

	services, err = service.NewServices(config)
	if err != nil {
		return nil, nil, err
	}

	if cmd.Flags().Changed("user") {
		services.Global.User = opts.user
	}
	user := services.Global.User

	if cmd.Flags().Changed("identity-file") {
		services.Global.KeyPath = opts.identityFile
	}
	keyPath := services.Global.KeyPath
	identityFile, password, err := utils.CheckSSHAuthentication(keyPath, opts.usePassword)
	if err != nil {
		return nil, nil, err
	}

	if opts.debugMode {
		log.SetLevel(log.DebugLevel)
		executor = ext.NewDebugExecutor(user, password, identityFile)
	} else {
		executor = ext.NewSSHExecutor(user, password, identityFile)
	}
	return services, executor, nil
}
