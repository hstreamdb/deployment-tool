package spec

import "path"

const (
	StoreDefaultContainerName = "deploy_hstore"
	StoreDefaultImage         = "hstreamdb/hstream"
	StoreDefaultBinPath       = "/usr/local/bin/logdeviced"
	StoreDefaultCfgDir        = "deploy/store"
	StoreDefaultDataDir       = "data/store"

	AdminDefaultContainerName = "deploy_hadmin"
	AdminDefaultImage         = "hstreamdb/hstream"
	AdminDefaultBinPath       = "/usr/local/bin/ld-admin-server"
	AdminDefaultCfgDir        = "deploy/admin"
	AdminDefaultDataDir       = "data/admin"
)

type HStoreSpec struct {
	Host             string       `yaml:"host"`
	Image            string       `yaml:"image"`
	SSHPort          int          `yaml:"ssh_port" default:"22"`
	Role             string       `yaml:"role" default:"Both"`
	Location         string       `yaml:"location"`
	EnableAdmin      bool         `yaml:"enable_admin"`
	Port             int          `yaml:"port" default:"6440"`
	EnablePrometheus bool         `yaml:"enable_prometheus"`
	PromListenAddr   string       `yaml:"prometheus_listen_addr" default:"0.0.0.0:6300"`
	StoreOps         StoreOps     `yaml:",inline"`
	ContainerCfg     ContainerCfg `yaml:"container_config"`
	RemoteCfgPath    string
	DataDir          string
}

func (h *HStoreSpec) SetDataDir(prefix string) {
	h.DataDir = path.Join(prefix, StoreDefaultDataDir)
}

func (h *HStoreSpec) SetDefaultImage() {
	h.Image = StoreDefaultImage
}

func (h *HStoreSpec) SetRemoteCfgPath(prefix string) {
	h.RemoteCfgPath = path.Join(prefix, StoreDefaultCfgDir)
}

type StoreOps struct {
	Disk   uint `yaml:"disk" default:"1"`
	Shards uint `yaml:"shards" default:"1"`
}

type HAdminSpec struct {
	Host             string       `yaml:"host"`
	Image            string       `yaml:"image"`
	SSHPort          int          `yaml:"ssh_port" default:"22"`
	Port             int          `yaml:"port" default:"6440"`
	EnablePrometheus bool         `yaml:"enable_prometheus"`
	PromListenAddr   string       `yaml:"prometheus_listen_addr" default:"0.0.0.0:6300"`
	ContainerCfg     ContainerCfg `yaml:"container_config"`
	RemoteCfgPath    string
	DataDir          string
}

func (h *HAdminSpec) SetDataDir(prefix string) {
	h.DataDir = path.Join(prefix, AdminDefaultDataDir)
}

func (h *HAdminSpec) SetDefaultImage() {
	h.Image = AdminDefaultImage
}

func (h *HAdminSpec) SetRemoteCfgPath(prefix string) {
	h.RemoteCfgPath = path.Join(prefix, AdminDefaultCfgDir)
}
