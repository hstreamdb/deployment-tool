package spec

import (
	"fmt"
	"strings"
)

type GlobalCfg struct {
	User                       string       `yaml:"user"`
	KeyPath                    string       `yaml:"key_path"`
	SSHPort                    int          `yaml:"ssh_port" default:"22"`
	MetaReplica                int          `yaml:"meta_replica" default:"1"`
	MetaStoreConfigPath        string       `yaml:"meta_store_config_path"`
	HStoreConfigPath           string       `yaml:"hstore_config_path"`
	HServerConfigPath          string       `yaml:"hserver_config_path"`
	EnableHsGrpc               bool         `yaml:"enable_grpc_haskell"`
	DisableStoreNetworkCfgPath bool         `yaml:"disable_store_network_config_path"`
	EsConfigPath               string       `yaml:"elastic_search_config_path"`
	EnableDscpReflection       bool         `yaml:"enable_dscp_reflection"`
	DisableMonitorSuite        bool         `yaml:"disable_monitor_suite"`
	ContainerCfg               ContainerCfg `yaml:"container_config"`
}

type ContainerCfg struct {
	Cpu            string `yaml:"cpu_limit"`
	Memory         string `yaml:"memory_limit"`
	RemoveWhenExit bool   `yaml:"remove_when_exit"`
	DisableRestart bool   `yaml:"disable_restart"`
	Options        string `yaml:"options"`
}

func (c ContainerCfg) GetCmd() string {
	args := make([]string, 0, 4)
	if !c.DisableRestart {
		args = append(args, "--restart unless-stopped")
	}
	if c.RemoveWhenExit {
		args = append(args, "--rm")
	}
	if len(c.Cpu) != 0 {
		args = append(args, fmt.Sprintf("--cpus=%s", c.Cpu))
	}
	if len(c.Memory) != 0 {
		args = append(args, fmt.Sprintf("--memory=%s", c.Memory))
	}
	if len(c.Options) != 0 {
		args = append(args, c.Options)
	}
	return strings.Join(args, " ")
}
