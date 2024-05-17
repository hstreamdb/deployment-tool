package spec

import "path"

const (
	ElasticSearchDefaultContainerName = "deploy_elastic_search"
	ElasticSearchDefaultImage         = "docker.elastic.co/elasticsearch/elasticsearch:8.13.3"
	ElasticSearchDefaultCfgDir        = "deploy/elasticsearch"
	ElasticSearchDefaultDataDir       = "data/elasticsearch"

	KibanaDefaultContainerName = "deploy_kibana"
	KibanaDefaultImage         = "docker.elastic.co/kibana/kibana:8.13.3"
	KibanaDefaultCfgDir        = "deploy/kibana"
	KibanaDefaultDataDir       = "data/kibana"

	FilebeatDefaultContainerName = "deploy_filebeat"
	FilebeatDefaultImage         = "docker.elastic.co/beats/filebeat-oss:8.13.3"
	FilebeatDefaultCfgDir        = "deploy/filebeat"
	FilebeatDefaultDataDir       = "data/filebeat"

	VectorDefaultContainerName = "deploy_vector"
	VectorDefaultImage         = "timberio/vector:latest-debian"
	VectorDefaultCfgDir        = "deploy/vector"
	VectorDefaultDataDir       = "data/vector"
)

type ElasticSearchSpec struct {
	Host          string            `yaml:"host"`
	SSHPort       int               `yaml:"ssh_port" default:"22"`
	Port          int               `yaml:"port" default:"9200"`
	Image         string            `yaml:"image"`
	ContainerCfg  ContainerCfg      `yaml:"container_config"`
	JavaOpts      string            `yaml:"es_java_opts"`
	ESConfigs     map[string]string `yaml:"es_configs"`
	DataDir       string
	RemoteCfgPath string
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
	ContainerCfg  ContainerCfg `yaml:"container_config"`
	RemoteCfgPath string
	DataDir       string
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
	ContainerCfg  ContainerCfg `yaml:"container_config"`
	RemoteCfgPath string
	DataDir       string
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

type VectorSpec struct {
	Host          string       `yaml:"host"`
	SSHPort       int          `yaml:"ssh_port" default:"22"`
	Image         string       `yaml:"image"`
	SinceNow      bool         `yaml:"since_now"`
	ContainerCfg  ContainerCfg `yaml:"container_config"`
	RemoteCfgPath string
	DataDir       string
}

func (v *VectorSpec) SetRemoteCfgPath(prefix string) {
	v.RemoteCfgPath = path.Join(prefix, VectorDefaultCfgDir)
}

func (v *VectorSpec) SetDataDir(prefix string) {
	v.DataDir = path.Join(prefix, VectorDefaultDataDir)
}

func (v *VectorSpec) SetDefaultImage() {
	v.Image = VectorDefaultImage
}
