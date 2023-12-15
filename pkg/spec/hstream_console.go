package spec

import "path"

const (
	ConsoleDefaultImage         = "hstreamdb/hstream-console"
	ConsoleDefaultContainerName = "deploy_hstream_console"
	ConsoleDefaultCfgDir        = "deploy/hstream_console"
	ConsoleDefaultDataDir       = "data/hstream_console"
)

type HStreamConsoleSpec struct {
	Host          string            `yaml:"host"`
	Port          int               `yaml:"port" default:"5177"`
	Image         string            `yaml:"image"`
	SSHPort       int               `yaml:"ssh_port" default:"22"`
	Option        map[string]string `yaml:"options"`
	RemoteCfgPath string            `yaml:"remote_config_path"`
	DataDir       string            `yaml:"data_dir"`
	ContainerCfg  ContainerCfg      `yaml:"container_config"`
}

func (h *HStreamConsoleSpec) SetDataDir(prefix string) {
	h.DataDir = path.Join(prefix, ConsoleDefaultDataDir)
}

func (h *HStreamConsoleSpec) SetDefaultImage() {
	h.Image = ConsoleDefaultImage
}

func (h *HStreamConsoleSpec) SetRemoteCfgPath(prefix string) {
	h.RemoteCfgPath = path.Join(prefix, ConsoleDefaultCfgDir)
}
