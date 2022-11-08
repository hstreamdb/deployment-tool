package service

import (
	"fmt"
	"github.com/hstreamdb/deployment-tool/pkg/executor"
	"github.com/hstreamdb/deployment-tool/pkg/spec"
	"github.com/hstreamdb/deployment-tool/pkg/template/config"
	"github.com/hstreamdb/deployment-tool/pkg/utils"
	"path/filepath"
	"strconv"
	"strings"
)

type ElasticSearch struct {
	spec            spec.ElasticSearchSpec
	ContainerName   string
	DisableSecurity bool
}

type Kibana struct {
	spec          spec.KibanaSpec
	ContainerName string
}

type Filebeat struct {
	spec              spec.FilebeatSpec
	ContainerName     string
	ElasticsearchHost string
	ElasticsearchPort string
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

func (es *ElasticSearch) getDirs() (string, string) {
	return es.spec.RemoteCfgPath, es.spec.DataDir
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
	args := append([]string{}, "sudo mkdir -p", cfgDir, dataDir)
	return &executor.ExecuteCtx{Target: es.spec.Host, Cmd: strings.Join(args, " ")}
}

func (es *ElasticSearch) Deploy(globalCtx *GlobalCtx) *executor.ExecuteCtx {
	args := spec.GetDockerExecCmd(globalCtx.containerCfg, es.spec.ContainerCfg, es.ContainerName, true)
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
	return nil
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
	cfgDir := k.spec.RemoteCfgPath
	args := append([]string{}, "sudo mkdir -p", cfgDir)
	return &executor.ExecuteCtx{Target: k.spec.Host, Cmd: strings.Join(args, " ")}
}

func (k *Kibana) Deploy(globalCtx *GlobalCtx) *executor.ExecuteCtx {
	args := spec.GetDockerExecCmd(globalCtx.containerCfg, k.spec.ContainerCfg, k.ContainerName, true)
	args = append(args, k.spec.Image)
	return &executor.ExecuteCtx{Target: k.spec.Host, Cmd: strings.Join(args, " ")}
}

func (k *Kibana) Remove(globalCtx *GlobalCtx) *executor.ExecuteCtx {
	args := []string{"docker rm -f", k.ContainerName}
	args = append(args, "&&", "sudo rm -rf", k.spec.RemoteCfgPath)
	return &executor.ExecuteCtx{Target: k.spec.Host, Cmd: strings.Join(args, " ")}
}

func (k *Kibana) SyncConfig(globalCtx *GlobalCtx) *executor.TransferCtx {
	return nil
}

func NewFilebeat(fbSpec spec.FilebeatSpec, elasticsearchHost, elasticsearchPort string) *Filebeat {
	return &Filebeat{
		spec:              fbSpec,
		ContainerName:     spec.FilebeatDefaultContainerName,
		ElasticsearchHost: elasticsearchHost,
		ElasticsearchPort: elasticsearchPort,
	}
}
func (fb *Filebeat) GetServiceName() string {
	return "filebeat"
}

func (fb *Filebeat) Display() map[string]utils.DisplayedComponent {
	cfgDir := fb.spec.RemoteCfgPath
	kibana := utils.DisplayedComponent{
		Name:          "Filebeat",
		Host:          fb.spec.Host,
		Ports:         "",
		ContainerName: fb.ContainerName,
		Image:         fb.spec.Image,
		Paths:         strings.Join([]string{cfgDir}, ","),
	}
	return map[string]utils.DisplayedComponent{"filebeat": kibana}
}

func (fb *Filebeat) InitEnv(globalCtx *GlobalCtx) *executor.ExecuteCtx {
	cfgDir := fb.spec.RemoteCfgPath
	args := append([]string{}, "sudo mkdir -p", cfgDir)
	return &executor.ExecuteCtx{Target: fb.spec.Host, Cmd: strings.Join(args, " ")}
}

func (fb *Filebeat) Deploy(globalCtx *GlobalCtx) *executor.ExecuteCtx {
	mountPoints := []spec.MountPoints{
		{"/var/lib/docker", "/var/lib/docker:ro"},
		{"/var/run/docker.sock", "/var/run/docker.sock"},
		{filepath.Join(fb.spec.RemoteCfgPath, "filebeat.yml"), "/usr/share/filebeat/filebeat.yml"},
	}
	args := spec.GetDockerExecCmd(globalCtx.containerCfg, fb.spec.ContainerCfg, fb.ContainerName, true, mountPoints...)
	args = append(args, fb.spec.Image)
	return &executor.ExecuteCtx{Target: fb.spec.Host, Cmd: strings.Join(args, " ")}
}

func (fb *Filebeat) Remove(globalCtx *GlobalCtx) *executor.ExecuteCtx {
	args := []string{"docker rm -f", fb.ContainerName}
	args = append(args, "&&", "sudo rm -rf", fb.spec.RemoteCfgPath)
	return &executor.ExecuteCtx{Target: fb.spec.Host, Cmd: strings.Join(args, " ")}
}

func (fb *Filebeat) SyncConfig(globalCtx *GlobalCtx) *executor.TransferCtx {
	cfg := config.FilebeatConfig{
		FilebeatHost:      fb.spec.Host,
		ElasticsearchHost: fb.ElasticsearchHost,
		ElasticsearchPort: fb.ElasticsearchPort,
	}
	genCfg, err := cfg.GenConfig()
	if err != nil {
		panic(fmt.Errorf("gen FilebeatConfig error: %s", err.Error()))
	}

	position := utils.ScpDir(filepath.Dir(genCfg), fb.spec.RemoteCfgPath)

	return &executor.TransferCtx{
		Target: fb.spec.Host, Position: position,
	}
}
