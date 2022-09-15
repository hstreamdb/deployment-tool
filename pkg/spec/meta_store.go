package spec

const (
	MetaStoreDefaultContainerName = "deploy_meta"
	MetaStoreDefaultImage         = "docker.io/zookeeper:3.6"
	MetaStoreDefaultCfgDir        = "/hstream/deploy/metastore"
	MetaStoreDefaultDataDir       = "/hstream/data/metastore"
)

type MetaStoreSpec struct {
	Host          string       `yaml:"host"`
	Image         string       `yaml:"image"`
	SshPort       int          `yaml:"ssh_port" default:"22"`
	LocalCfgPath  string       `yaml:"local_config_path"`
	RemoteCfgPath string       `yaml:"remote_config_path"`
	DataDir       string       `yaml:"data_dir"`
	ContainerCfg  ContainerCfg `yaml:"container_config"`
}
