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

type StopOpts struct {
	user         string
	usePassword  bool
	identityFile string
	configPath   string
	debugMode    bool
}

func newStop() *cobra.Command {
	opts := StopOpts{}
	cmd := &cobra.Command{
		Use:   "stop",
		Short: "Stop HStreamDB cluster.",
		RunE: func(cmd *cobra.Command, args []string) error {
			contant, err := os.ReadFile(opts.configPath)
			log.Debugf("opts: %+v\n", opts)
			if err != nil {
				return err
			}

			config := &spec.ComponentsSpec{}
			if err = yaml.Unmarshal(contant, config); err != nil {
				return err
			}
			services, err = service.NewServices(config)
			if err != nil {
				return err
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
				return err
			}

			var executor ext.Executor
			if opts.debugMode {
				executor = ext.NewDebugExecutor(user, password, identityFile)
			} else {
				executor = ext.NewSSHExecutor(user, password, identityFile)
			}

			if err = task.StopCluster(executor, services); err != nil {
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
