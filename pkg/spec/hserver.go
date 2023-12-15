package spec

import "path"

const (
	ServerBinConfigPath        = "/etc/hstream/config.yaml"
	ServerDefaultImage         = "hstreamdb/hstream"
	ServerDefaultContainerName = "deploy_hserver"
	ServerDefaultBinPath       = "/usr/local/bin/hstream-server"
	ServerGrpcHaskellBinPath   = "/usr/local/bin/hstream-server-old"
	ServerDefaultCfgDir        = "deploy/hserver"
	ServerDefaultDataDir       = "data/hserver"
)

type HServerSpec struct {
	Host string `yaml:"host"`
	// AdvertisedAddress only used before hstream v0.10.1
	// After v0.10.1, this field will be filled with the internal ip address
	AdvertisedAddress  string            `yaml:"advertised_address"`
	AdvertisedListener string            `yaml:"advertised_listener"`
	Port               int               `yaml:"port" default:"6570"`
	InternalPort       int               `yaml:"internal_port" default:"6571"`
	MonitorPort        int               `yaml:"monitor_port" default:"9700"`
	Image              string            `yaml:"image"`
	SSHPort            int               `yaml:"ssh_port" default:"22"`
	RemoteCfgPath      string            `yaml:"remote_config_path"`
	DataDir            string            `yaml:"data_dir"`
	Opts               map[string]string `yaml:"server_param"`
	ContainerCfg       ContainerCfg      `yaml:"container_config"`
}

func (h *HServerSpec) SetDataDir(prefix string) {
	h.DataDir = path.Join(prefix, ServerDefaultDataDir)
}

func (h *HServerSpec) SetDefaultImage() {
	h.Image = ServerDefaultImage
}

func (h *HServerSpec) SetRemoteCfgPath(prefix string) {
	h.RemoteCfgPath = path.Join(prefix, ServerDefaultCfgDir)
}

type ServerOpts struct {
	ServerLogLevel string `yaml:"server_log_level" default:"info"`
	StoreLogLevel  string `yaml:"store_log_level" default:"info"`
	Compression    string `yaml:"compression" default:"lz4"`
}
