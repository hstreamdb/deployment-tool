package main

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

var (
	services *service.Services
)

type DeployOpts struct {
	user         string
	usePassword  bool
	identityFile string
	configPath   string
	debugMode    bool
}

func newDeploy() *cobra.Command {
	opts := DeployOpts{}
	cmd := &cobra.Command{
		Use:   "start",
		Short: "Deploy a HStreamDB Cluster.",
		RunE: func(cmd *cobra.Command, args []string) error {
			contents, err := os.ReadFile(opts.configPath)
			log.Debugf("opts: %+v\n", opts)
			if err != nil {
				return err
			}

			config := spec.ComponentsSpec{}
			if err = yaml.Unmarshal(contents, &config); err != nil {
				return err
			}
			services, err = service.NewServices(config)
			if err != nil {
				return err
			}

			user := services.Global.User
			if cmd.Flags().Changed("user") {
				user = opts.user
			}
			keyPath := services.Global.KeyPath
			if cmd.Flags().Changed("identity-file") {
				keyPath = opts.identityFile
			}

			identityFile, password, err := utils.CheckSSHAuthentication(keyPath, opts.usePassword)
			if err != nil {
				return err
			}

			var executor ext.Executor
			if opts.debugMode {
				log.SetLevel(log.DebugLevel)
				executor = ext.NewDebugExecutor(user, password, identityFile)
			} else {
				executor = ext.NewSSHExecutor(user, password, identityFile)
			}

			if err = task.SetUpCluster(executor, services); err != nil {
				return err
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&opts.configPath, "config", "c", "template/config.yaml", "Cluster config path.")
	cmd.Flags().StringVarP(&opts.user, "user", "u", "", "User name to login via ssh.")
	cmd.Flags().BoolVarP(&opts.usePassword, "use-password", "p", false, "Use password authentication for ssh.")
	cmd.Flags().StringVarP(&opts.identityFile, "identity-file", "i", "", "The path of the SSH identity file.")
	cmd.Flags().BoolVarP(&opts.debugMode, "debug", "d", false, "Debug mode")
	return cmd
}
