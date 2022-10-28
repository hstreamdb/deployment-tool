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
	Port          int          `yaml:"port" default:"4001"`
	RaftPort      int          `yaml:"raft_port" default:"4002"`
	SSHPort       int          `yaml:"ssh_port" default:"22"`
	RemoteCfgPath string       `yaml:"remote_config_path"`
	DataDir       string       `yaml:"data_dir"`
	ContainerCfg  ContainerCfg `yaml:"container_config"`
}

func (m *MetaStoreSpec) SetDefaultDataDir() {
	m.DataDir = MetaStoreDefaultDataDir
}

func (m *MetaStoreSpec) SetDefaultImage() {
	m.Image = MetaStoreDefaultImage
}

func (m *MetaStoreSpec) SetDefaultRemoteCfgPath() {
	m.RemoteCfgPath = MetaStoreDefaultCfgDir
}
