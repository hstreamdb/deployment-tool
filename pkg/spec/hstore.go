package spec

const (
	StoreDefaultContainerName = "deploy_hstore"
	StoreDefaultImage         = "hstreamdb/hstream"
	StoreDefaultBinPath       = "/usr/local/bin/logdeviced"
	StoreDefaultCfgDir        = "/hstream/deploy/store"
	StoreDefaultDataDir       = "/hstream/data/store"
)

type HStoreSpec struct {
	Host          string       `yaml:"host"`
	Image         string       `yaml:"image"`
	SSHPort       int          `yaml:"ssh_port" default:"22"`
	RemoteCfgPath string       `yaml:"remote_config_path"`
	DataDir       string       `yaml:"data_dir"`
	Role          string       `yaml:"role" default:"Both"`
	EnableAdmin   bool         `yaml:"enable_admin"`
	AdminPort     int          `yaml:"admin_port" default:"6440"`
	StoreOps      StoreOps     `yaml:",inline"`
	ContainerCfg  ContainerCfg `yaml:"container_config"`
}

func (h *HStoreSpec) SetDefaultDataDir() {
	h.DataDir = StoreDefaultDataDir
}

func (h *HStoreSpec) SetDefaultImage() {
	h.Image = StoreDefaultImage
}

func (h *HStoreSpec) SetDefaultRemoteCfgPath() {
	h.RemoteCfgPath = StoreDefaultCfgDir
}

type StoreOps struct {
	Disk   uint `yaml:"disk" default:"1"`
	Shards uint `yaml:"shards" default:"1"`
}

type HAdminSpec struct {
	Host         string       `yaml:"host"`
	Image        string       `yaml:"image"`
	SSHPort      int          `yaml:"ssh_port" default:"22"`
	Replica      string       `yaml:"meta_replica"`
	Embed        bool         `yaml:"embed"`
	ContainerCfg ContainerCfg `yaml:"container_config"`
}
