package spec

import "path"

const (
	MetaStoreDefaultContainerName = "deploy_meta"
	MetaStoreDefaultImage         = "docker.io/zookeeper:3.6"
	MetaStoreDefaultCfgDir        = "deploy/metastore"
	MetaStoreDefaultDataDir       = "data/metastore"
)

type MetaStoreSpec struct {
	Host          string       `yaml:"host"`
	Image         string       `yaml:"image"`
	Port          int          `yaml:"port"`
	RaftPort      int          `yaml:"raft_port" default:"4002"`
	SSHPort       int          `yaml:"ssh_port" default:"22"`
	ContainerCfg  ContainerCfg `yaml:"container_config"`
	RemoteCfgPath string
	DataDir       string
}

func (m *MetaStoreSpec) SetDataDir(prefix string) {
	m.DataDir = path.Join(prefix, MetaStoreDefaultDataDir)
}

func (m *MetaStoreSpec) SetDefaultImage() {
	m.Image = MetaStoreDefaultImage
}

func (m *MetaStoreSpec) SetRemoteCfgPath(prefix string) {
	m.RemoteCfgPath = path.Join(prefix, MetaStoreDefaultCfgDir)
}
