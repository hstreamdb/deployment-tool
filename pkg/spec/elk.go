package spec

const (
	ElasticSearchDefaultContainerName = "deploy_elastic_search"
	ElasticSearchDefaultImage         = "docker.elastic.co/elasticsearch/elasticsearch:8.5.0"
	ElasticSearchDefaultCfgDir        = "/hstream/deploy/elasticsearch"
	ElasticSearchDefaultDataDir       = "/hstream/data/elasticsearch"

	KibanaDefaultContainerName = "deploy_kibana"
	KibanaDefaultImage         = "docker.elastic.co/kibana/kibana:8.5.0"
	KibanaDefaultCfgDir        = "/hstream/deploy/kibana"
	KibanaDefaultDataDir       = "/hstream/data/kibana"

	FilebeatDefaultContainerName = "deploy_filebeat"
	FilebeatDefaultImage         = "docker.elastic.co/beats/filebeat:8.5.0"
	FilebeatDefaultCfgDir        = "/hstream/deploy/filebeat"
	FilebeatDefaultDataDir       = "/hstream/data/filebeat"
)

type ElasticSearchSpec struct {
	Host          string       `yaml:"host"`
	SSHPort       int          `yaml:"ssh_port" default:"22"`
	Port          int          `yaml:"port" default:"9200"`
	Image         string       `yaml:"image"`
	DataDir       string       `yaml:"data_dir"`
	RemoteCfgPath string       `yaml:"remote_config_path"`
	ContainerCfg  ContainerCfg `yaml:"container_config"`
}

func (es *ElasticSearchSpec) SetDefaultDataDir() {
	es.DataDir = ElasticSearchDefaultDataDir
}

func (es *ElasticSearchSpec) SetDefaultRemoteCfgPath() {
	es.RemoteCfgPath = ElasticSearchDefaultCfgDir
}

func (es *ElasticSearchSpec) SetDefaultImage() {
	es.Image = ElasticSearchDefaultImage
}

type KibanaSpec struct {
	Host          string       `yaml:"host"`
	SSHPort       int          `yaml:"ssh_port" default:"22"`
	Port          int          `yaml:"port" default:"5601"`
	Image         string       `yaml:"image"`
	RemoteCfgPath string       `yaml:"remote_config_path"`
	DataDir       string       `yaml:"data_dir"`
	ContainerCfg  ContainerCfg `yaml:"container_config"`
}

func (k *KibanaSpec) SetDefaultDataDir() {
	k.DataDir = KibanaDefaultDataDir
}

func (k *KibanaSpec) SetDefaultImage() {
	k.Image = KibanaDefaultImage
}

func (k *KibanaSpec) SetDefaultRemoteCfgPath() {
	k.RemoteCfgPath = KibanaDefaultCfgDir
}

type FilebeatSpec struct {
	Host          string       `yaml:"host"`
	SSHPort       int          `yaml:"ssh_port" default:"22"`
	Image         string       `yaml:"image"`
	RemoteCfgPath string       `yaml:"remote_config_path"`
	DataDir       string       `yaml:"data_dir"`
	ContainerCfg  ContainerCfg `yaml:"container_config"`
}

func (fb *FilebeatSpec) SetDefaultRemoteCfgPath() {
	fb.RemoteCfgPath = FilebeatDefaultCfgDir
}

func (fb *FilebeatSpec) SetDefaultDataDir() {
	fb.DataDir = FilebeatDefaultDataDir
}

func (fb *FilebeatSpec) SetDefaultImage() {
	fb.Image = FilebeatDefaultImage
}
