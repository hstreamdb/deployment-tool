package service

import (
	"fmt"
	"github.com/hstreamdb/deployment-tool/pkg/executor"
	"github.com/hstreamdb/deployment-tool/pkg/spec"
	"github.com/hstreamdb/deployment-tool/pkg/template/config"
	"github.com/hstreamdb/deployment-tool/pkg/utils"
	"path"
	"path/filepath"
	"strconv"
	"strings"
)

type ElasticSearch struct {
	spec            spec.ElasticSearchSpec
	ContainerName   string
	DisableSecurity bool
}

func NewElasticSearch(esSpec spec.ElasticSearchSpec) *ElasticSearch {
	return &ElasticSearch{
		spec:          esSpec,
		ContainerName: spec.ElasticSearchDefaultContainerName,
		// FIXME: currently, only support `xpack.security.enabled=false`
		DisableSecurity: false,
	}
}

func (es *ElasticSearch) GetServiceName() string {
	return "elasticsearch"
}

func (es *ElasticSearch) Display() map[string]utils.DisplayedComponent {
	cfgDir, dataDir := es.getDirs()
	elasticsearch := utils.DisplayedComponent{
		Name:          "ElasticSearch",
		Host:          es.spec.Host,
		Ports:         strconv.Itoa(es.spec.Port),
		ContainerName: es.ContainerName,
		Image:         es.spec.Image,
		Paths:         strings.Join([]string{cfgDir, dataDir}, ","),
	}
	return map[string]utils.DisplayedComponent{"elasticsearch": elasticsearch}
}

func (es *ElasticSearch) InitEnv(globalCtx *GlobalCtx) *executor.ExecuteCtx {
	cfgDir, dataDir := es.getDirs()
	args := append([]string{}, "sudo mkdir -p", cfgDir, dataDir, "-m 0775")
	return &executor.ExecuteCtx{Target: es.spec.Host, Cmd: strings.Join(args, " ")}
}

func (es *ElasticSearch) Deploy(globalCtx *GlobalCtx) *executor.ExecuteCtx {
	mountPoints := []spec.MountPoints{
		{es.spec.DataDir, "/usr/share/elasticsearch/data"},
	}
	if len(globalCtx.LocalEsConfigFile) != 0 {
		mountPoints = append(mountPoints, spec.MountPoints{
			Local:  path.Join(es.spec.RemoteCfgPath, "elasticsearch.yml"),
			Remote: "/usr/share/elasticsearch/config/elasticsearch.yml"})
	}
	args := spec.GetDockerExecCmd(globalCtx.containerCfg, es.spec.ContainerCfg, es.ContainerName, true, mountPoints...)
	if es.DisableSecurity {
		args = append(args, "-e xpack.security.enabled=false")
	}
	args = append(args, "-e discovery.type=single-node")
	args = append(args, es.spec.Image)
	return &executor.ExecuteCtx{Target: es.spec.Host, Cmd: strings.Join(args, " ")}
}

func (es *ElasticSearch) Remove(globalCtx *GlobalCtx) *executor.ExecuteCtx {
	args := []string{"docker rm -f", es.ContainerName}
	args = append(args, "&&", "sudo rm -rf", es.spec.DataDir, es.spec.RemoteCfgPath)
	return &executor.ExecuteCtx{Target: es.spec.Host, Cmd: strings.Join(args, " ")}
}

func (es *ElasticSearch) SyncConfig(globalCtx *GlobalCtx) *executor.TransferCtx {
	if len(globalCtx.LocalEsConfigFile) != 0 {
		position := []executor.Position{
			{LocalDir: globalCtx.LocalEsConfigFile, RemoteDir: es.spec.RemoteCfgPath},
		}
		return &executor.TransferCtx{Target: es.spec.Host, Position: position}
	}
	return nil
}

func (es *ElasticSearch) getDirs() (string, string) {
	return es.spec.RemoteCfgPath, es.spec.DataDir
}

type Kibana struct {
	spec          spec.KibanaSpec
	ContainerName string
}

func NewKibana(kibanaSpec spec.KibanaSpec) *Kibana {
	return &Kibana{
		spec:          kibanaSpec,
		ContainerName: spec.KibanaDefaultContainerName,
	}
}

func (k *Kibana) GetServiceName() string {
	return "kibana"
}

func (k *Kibana) Display() map[string]utils.DisplayedComponent {
	cfgDir := k.spec.RemoteCfgPath
	kibana := utils.DisplayedComponent{
		Name:          "Kibana",
		Host:          k.spec.Host,
		Ports:         strconv.Itoa(k.spec.Port),
		ContainerName: k.ContainerName,
		Image:         k.spec.Image,
		Paths:         strings.Join([]string{cfgDir}, ","),
	}
	return map[string]utils.DisplayedComponent{"kibana": kibana}
}

func (k *Kibana) InitEnv(globalCtx *GlobalCtx) *executor.ExecuteCtx {
	cfgDir, dataDir := k.getDirs()
	args := append([]string{}, "sudo mkdir -p", cfgDir, dataDir, "-m 0775")
	return &executor.ExecuteCtx{Target: k.spec.Host, Cmd: strings.Join(args, " ")}
}

func (k *Kibana) Deploy(globalCtx *GlobalCtx) *executor.ExecuteCtx {
	args := spec.GetDockerExecCmd(globalCtx.containerCfg, k.spec.ContainerCfg, k.ContainerName, true)
	args = append(args, k.spec.Image)
	return &executor.ExecuteCtx{Target: k.spec.Host, Cmd: strings.Join(args, " ")}
}

func (k *Kibana) Remove(globalCtx *GlobalCtx) *executor.ExecuteCtx {
	args := []string{"docker rm -f", k.ContainerName}
	args = append(args, "&&", "sudo rm -rf", k.spec.RemoteCfgPath, k.spec.DataDir)
	return &executor.ExecuteCtx{Target: k.spec.Host, Cmd: strings.Join(args, " ")}
}

func (k *Kibana) SyncConfig(globalCtx *GlobalCtx) *executor.TransferCtx {
	return nil
}

func (k *Kibana) getDirs() (string, string) {
	return k.spec.RemoteCfgPath, k.spec.DataDir
}

type Filebeat struct {
	spec              spec.FilebeatSpec
	ContainerName     string
	ElasticsearchHost string
	ElasticsearchPort string
}

func NewFilebeat(fbSpec spec.FilebeatSpec, elasticsearchHost, elasticsearchPort string) *Filebeat {
	return &Filebeat{
		spec:              fbSpec,
		ContainerName:     spec.FilebeatDefaultContainerName,
		ElasticsearchHost: elasticsearchHost,
		ElasticsearchPort: elasticsearchPort,
	}
}
func (f *Filebeat) GetServiceName() string {
	return "filebeat"
}

func (f *Filebeat) Display() map[string]utils.DisplayedComponent {
	cfgDir := f.spec.RemoteCfgPath
	kibana := utils.DisplayedComponent{
		Name:          "Filebeat",
		Host:          f.spec.Host,
		Ports:         "",
		ContainerName: f.ContainerName,
		Image:         f.spec.Image,
		Paths:         strings.Join([]string{cfgDir}, ","),
	}
	return map[string]utils.DisplayedComponent{"filebeat": kibana}
}

func (f *Filebeat) InitEnv(globalCtx *GlobalCtx) *executor.ExecuteCtx {
	cfgDir, dataDir := f.getDirs()
	args := append([]string{}, "sudo mkdir -p", cfgDir, dataDir, "-m 0775")
	return &executor.ExecuteCtx{Target: f.spec.Host, Cmd: strings.Join(args, " ")}
}

func (f *Filebeat) Deploy(globalCtx *GlobalCtx) *executor.ExecuteCtx {
	mountPoints := []spec.MountPoints{
		{"/var/lib/docker/containers", "/var/lib/docker/containers:ro"},
		{"/var/run/docker.sock", "/var/run/docker.sock:ro"},
		{filepath.Join(f.spec.RemoteCfgPath, "filebeat.yml"), "/usr/share/filebeat/filebeat.yml:ro"},
	}
	args := spec.GetDockerExecCmd(globalCtx.containerCfg, f.spec.ContainerCfg, f.ContainerName, true, mountPoints...)
	args = append(args, f.spec.Image)
	return &executor.ExecuteCtx{Target: f.spec.Host, Cmd: strings.Join(args, " ")}
}

func (f *Filebeat) Remove(globalCtx *GlobalCtx) *executor.ExecuteCtx {
	args := []string{"docker rm -f", f.ContainerName}
	args = append(args, "&&", "sudo rm -rf", f.spec.RemoteCfgPath, f.spec.DataDir)
	return &executor.ExecuteCtx{Target: f.spec.Host, Cmd: strings.Join(args, " ")}
}

func (f *Filebeat) SyncConfig(globalCtx *GlobalCtx) *executor.TransferCtx {
	cfg := config.FilebeatConfig{
		FilebeatHost:      f.spec.Host,
		ElasticsearchHost: f.ElasticsearchHost,
		ElasticsearchPort: f.ElasticsearchPort,
	}
	genCfg, err := cfg.GenConfig()
	if err != nil {
		panic(fmt.Errorf("gen FilebeatConfig error: %s", err.Error()))
	}
	position := []executor.Position{
		{LocalDir: genCfg, RemoteDir: filepath.Join(f.spec.RemoteCfgPath, "filebeat.yml")},
	}

	return &executor.TransferCtx{
		Target: f.spec.Host, Position: position,
	}
}

func (f *Filebeat) getDirs() (string, string) {
	return f.spec.RemoteCfgPath, f.spec.DataDir
}
