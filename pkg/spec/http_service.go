package spec

const (
	HttpServerDefaultContainerName = "deploy_http_server"
	HttpServerDefaultImage         = "hstreamdb/http-server"
	HttpServerDefaultCfgDir        = "/hstream/deploy/http-server"
	HttpServerDefaultDataDir       = "/hstream/data/http-server"
)

type HttpServerSpec struct {
	Host          string       `yaml:"host"`
	Image         string       `yaml:"image"`
	Port          int          `yaml:"port" default:"8081"`
	SSHPort       int          `yaml:"ssh_port" default:"22"`
	RemoteCfgPath string       `yaml:"remote_config_path"`
	DataDir       string       `yaml:"data_dir"`
	ContainerCfg  ContainerCfg `yaml:"container_config"`
}

func (h *HttpServerSpec) SetDefaultDataDir() {
	h.DataDir = HttpServerDefaultDataDir
}

func (h *HttpServerSpec) SetDefaultImage() {
	h.Image = HttpServerDefaultImage
}

func (h *HttpServerSpec) SetDefaultRemoteCfgPath() {
	h.RemoteCfgPath = HttpServerDefaultCfgDir
}
