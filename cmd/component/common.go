package component

import (
	ext "github.com/hstreamdb/deployment-tool/pkg/executor"
	"github.com/hstreamdb/deployment-tool/pkg/service"
	"github.com/hstreamdb/deployment-tool/pkg/spec"
	"github.com/hstreamdb/deployment-tool/pkg/utils"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
	"os"
)

// usage template was derived from https://github.com/spf13/cobra/blob/main/command.go#L539-L569
const UsageTpl = `
Usage:{{if .Runnable}}
 {{.UseLine}}{{end}}{{if gt (len .Aliases) 0}}

Aliases:
 {{.NameAndAliases}}{{end}}{{if .HasExample}}

Examples:
{{.Example}}{{end}}{{if .HasAvailableSubCommands}}{{$cmds := .Commands}}{{if eq (len .Groups) 0}}

Available Commands:{{range $cmds}}{{if (or .IsAvailableCommand (eq .Name "help"))}}
 {{rpad .Name .NamePadding }} {{.Short}}{{end}}{{end}}{{else}}{{range $group := .Groups}}

{{.Title}}{{range $cmds}}{{if (and (eq .GroupID $group.ID) (or .IsAvailableCommand (eq .Name "help")))}}
 {{rpad .Name .NamePadding }} {{.Short}}{{end}}{{end}}{{end}}{{if not .AllChildCommandsHaveGroup}}

Additional Commands:{{range $cmds}}{{if (and (eq .GroupID "") (or .IsAvailableCommand (eq .Name "help")))}}
 {{rpad .Name .NamePadding }} {{.Short}}{{end}}{{end}}{{end}}{{end}}{{end}}{{if .HasAvailableLocalFlags}}

Flags:
{{.LocalFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}{{if .HasAvailableInheritedFlags}}

Global Flags:
{{.InheritedFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}{{if .HasHelpSubCommands}}

Additional help topics:{{range .Commands}}{{if .IsAdditionalHelpTopicCommand}}
 {{rpad .CommandPath .CommandPathPadding}} {{.Short}}{{end}}{{end}}{{end}}{{if .HasAvailableSubCommands}}

Use "{{.CommandPath}} [command] --help" for more information about a command.{{end}}
`

type commonOpts struct {
	user         string
	usePassword  bool
	identityFile string
	configPath   string
	debugMode    bool
}

func getServices(cmd *cobra.Command, opts commonOpts) (*service.Services, ext.Executor, error) {
	var (
		executor ext.Executor
		services *service.Services
	)
	contents, err := os.ReadFile(opts.configPath)
	log.Debugf("opts: %+v\n", opts)
	if err != nil {
		return nil, nil, err
	}

	config := &spec.ComponentsSpec{}
	if err = yaml.Unmarshal(contents, config); err != nil {
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
