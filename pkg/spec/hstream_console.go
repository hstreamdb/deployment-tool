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
	ContainerCfg  ContainerCfg      `yaml:"container_config"`
	RemoteCfgPath string
	DataDir       string
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
