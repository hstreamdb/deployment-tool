package spec

const (
	ServerBinConfigPath        = "/etc/hstream/config.yaml"
	ServerDefaultImage         = "docker.io/hstreamdb/hstream"
	ServerDefaultContainerName = "deploy_hserver"
	ServerDefaultBinPath       = "/usr/local/bin/hstream-server"
	ServerDefaultConfigPath    = "/hstream/deploy/hserver"
	ServerDefaultDataDir       = "/hstream/data/hserver"
)

type HServerSpec struct {
	Host          string       `yaml:"host"`
	Address       string       `yaml:"address"`
	Port          int          `yaml:"port" default:"6570"`
	InternalPort  int          `yaml:"internal_port" default:"6571"`
	Image         string       `yaml:"image"`
	SshPort       int          `yaml:"ssh_port" default:"22"`
	LocalCfgPath  string       `yaml:"local_config_path"`
	RemoteCfgPath string       `yaml:"remote_config_path"`
	DataDir       string       `yaml:"data_dir"`
	Opts          ServerOpts   `yaml:"server_config"`
	ContainerCfg  ContainerCfg `yaml:"container_config"`
}

type ServerOpts struct {
	ServerLogLevel string `yaml:"server_log_level" default:"info"`
	StoreLogLevel  string `yaml:"store_log_level" default:"info"`
	Compression    string `yaml:"compression" default:"lz4"`
}
