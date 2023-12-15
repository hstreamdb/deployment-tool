package spec

import "path"

const (
	elasticDockerRegistry = "docker.elastic.co/"
	elasticVersion        = ":7.10.2"
)

const (
	ElasticSearchDefaultContainerName = "deploy_elastic_search"
	ElasticSearchDefaultImage         = elasticDockerRegistry + "elasticsearch/elasticsearch-oss" + elasticVersion
	ElasticSearchDefaultCfgDir        = "deploy/elasticsearch"
	ElasticSearchDefaultDataDir       = "data/elasticsearch"

	KibanaDefaultContainerName = "deploy_kibana"
	KibanaDefaultImage         = elasticDockerRegistry + "kibana/kibana-oss" + elasticVersion
	KibanaDefaultCfgDir        = "deploy/kibana"
	KibanaDefaultDataDir       = "data/kibana"

	FilebeatDefaultContainerName = "deploy_filebeat"
	FilebeatDefaultImage         = elasticDockerRegistry + "beats/filebeat-oss" + elasticVersion
	FilebeatDefaultCfgDir        = "deploy/filebeat"
	FilebeatDefaultDataDir       = "data/filebeat"
)

type ElasticSearchSpec struct {
	Host          string       `yaml:"host"`
	SSHPort       int          `yaml:"ssh_port" default:"22"`
	Port          int          `yaml:"port" default:"9200"`
	Image         string       `yaml:"image"`
	DataDir       string       `yaml:"data_dir"`
	RemoteCfgPath string       `yaml:"remote_config_path"`
	ContainerCfg  ContainerCfg `yaml:"container_config"`
	IsOss         *bool        `yaml:"is_oss,omitempty"`
}

func (es *ElasticSearchSpec) SetDataDir(prefix string) {
	es.DataDir = path.Join(prefix, ElasticSearchDefaultDataDir)
}

func (es *ElasticSearchSpec) SetRemoteCfgPath(prefix string) {
	es.RemoteCfgPath = path.Join(prefix, ElasticSearchDefaultCfgDir)
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

	IsOss *bool `yaml:"is_oss,omitempty"`
}

func (k *KibanaSpec) SetDataDir(prefix string) {
	k.DataDir = path.Join(prefix, KibanaDefaultDataDir)
}

func (k *KibanaSpec) SetDefaultImage() {
	k.Image = KibanaDefaultImage
}

func (k *KibanaSpec) SetRemoteCfgPath(prefix string) {
	k.RemoteCfgPath = path.Join(prefix, KibanaDefaultCfgDir)
}

type FilebeatSpec struct {
	Host          string       `yaml:"host"`
	SSHPort       int          `yaml:"ssh_port" default:"22"`
	Image         string       `yaml:"image"`
	RemoteCfgPath string       `yaml:"remote_config_path"`
	DataDir       string       `yaml:"data_dir"`
	ContainerCfg  ContainerCfg `yaml:"container_config"`
}

func (fb *FilebeatSpec) SetRemoteCfgPath(prefix string) {
	fb.RemoteCfgPath = path.Join(prefix, FilebeatDefaultCfgDir)
}

func (fb *FilebeatSpec) SetDataDir(prefix string) {
	fb.DataDir = path.Join(prefix, FilebeatDefaultDataDir)
}

func (fb *FilebeatSpec) SetDefaultImage() {
	fb.Image = FilebeatDefaultImage
}
