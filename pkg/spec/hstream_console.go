package spec

const (
	ConsoleDefaultImage         = "hstreamdb/hstream-console"
	ConsoleDefaultContainerName = "deploy_hstream_console"
	ConsoleDefaultCfgDir        = "/hstream/deploy/hstream_console"
	ConsoleDefaultDataDir       = "/hstream/data/hstream_console"
)

type HStreamConsoleSpec struct {
	Host          string       `yaml:"host"`
	Port          int          `yaml:"port" default:"5177"`
	Image         string       `yaml:"image"`
	SSHPort       int          `yaml:"ssh_port" default:"22"`
	RemoteCfgPath string       `yaml:"remote_config_path"`
	DataDir       string       `yaml:"data_dir"`
	ContainerCfg  ContainerCfg `yaml:"container_config"`
}

func (h *HStreamConsoleSpec) SetDefaultDataDir() {
	h.DataDir = ConsoleDefaultDataDir
}

func (h *HStreamConsoleSpec) SetDefaultImage() {
	h.Image = ConsoleDefaultImage
}

func (h *HStreamConsoleSpec) SetDefaultRemoteCfgPath() {
	h.RemoteCfgPath = ConsoleDefaultCfgDir
}
