package spec

const (
	KafkaServerBinConfigPath        = "/etc/hstream/config.yaml"
	KafkaServerDefaultImage         = "hstreamdb/hstream"
	KafkaServerDefaultContainerName = "deploy_hserver_kafka"
	KafkaServerDefaultBinPath       = "/usr/local/bin/hstream-kafka-server"
	KafkaServerDefaultCfgDir        = "/hstream/deploy/hserver_kafka"
	KafkaServerDefaultDataDir       = "/hstream/data/hserver_kafka"
)

type HServerKafkaSpec struct {
	Host               string            `yaml:"host"`
	AdvertisedAddress  string            `yaml:"advertised_address"`
	AdvertisedListener string            `yaml:"advertised_listener"`
	Port               int               `yaml:"port" default:"6570"`
	Image              string            `yaml:"image"`
	SSHPort            int               `yaml:"ssh_port" default:"22"`
	RemoteCfgPath      string            `yaml:"remote_config_path"`
	DataDir            string            `yaml:"data_dir"`
	Opts               map[string]string `yaml:"server_param"`
	ContainerCfg       ContainerCfg      `yaml:"container_config"`
}

func (h *HServerKafkaSpec) SetDefaultDataDir() {
	h.DataDir = KafkaServerDefaultDataDir
}

func (h *HServerKafkaSpec) SetDefaultImage() {
	h.Image = KafkaServerDefaultImage
}

func (h *HServerKafkaSpec) SetDefaultRemoteCfgPath() {
	h.RemoteCfgPath = KafkaServerDefaultCfgDir
}

type KafkaServerOpts struct {
	ServerLogLevel string `yaml:"server_log_level" default:"info"`
	StoreLogLevel  string `yaml:"store_log_level" default:"info"`
	Compression    string `yaml:"compression" default:"lz4"`
}
