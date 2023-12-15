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
	Host          string       `yaml:"host"`
	Image         string       `yaml:"image"`
	SSHPort       int          `yaml:"ssh_port" default:"22"`
	RemoteCfgPath string       `yaml:"remote_config_path"`
	DataDir       string       `yaml:"data_dir"`
	Role          string       `yaml:"role" default:"Both"`
	Location      string       `yaml:"location"`
	EnableAdmin   bool         `yaml:"enable_admin"`
	Port          int          `yaml:"port" default:"6440"`
	StoreOps      StoreOps     `yaml:",inline"`
	ContainerCfg  ContainerCfg `yaml:"container_config"`
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
	Host          string       `yaml:"host"`
	Image         string       `yaml:"image"`
	SSHPort       int          `yaml:"ssh_port" default:"22"`
	Port          int          `yaml:"port" default:"6440"`
	RemoteCfgPath string       `yaml:"remote_config_path"`
	DataDir       string       `yaml:"data_dir"`
	ContainerCfg  ContainerCfg `yaml:"container_config"`
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
