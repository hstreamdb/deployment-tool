package service

import (
	"fmt"
	"github.com/hstreamdb/deployment-tool/pkg/executor"
	"github.com/hstreamdb/deployment-tool/pkg/spec"
	"github.com/hstreamdb/deployment-tool/pkg/template/config"
	"github.com/hstreamdb/deployment-tool/pkg/template/script"
	"github.com/hstreamdb/deployment-tool/pkg/utils"
	log "github.com/sirupsen/logrus"
	"os"
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
		DisableSecurity: true,
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
	args := append([]string{}, "mkdir -p", cfgDir, dataDir+"/data", "-m 0775")
	args = append(args, fmt.Sprintf("&& chown -R %[1]s:$(id -gn %[1]s) %[2]s %[3]s", globalCtx.User, cfgDir, dataDir))
	args = append(args, fmt.Sprintf("&& chgrp -R 0 %s %s", cfgDir, dataDir))
	return &executor.ExecuteCtx{Target: es.spec.Host, Cmd: strings.Join(args, " ")}
}

func (es *ElasticSearch) Deploy(globalCtx *GlobalCtx) *executor.ExecuteCtx {
	mountPoints := []spec.MountPoints{
		{"/mnt", "/mnt"},
		{es.spec.DataDir + "/data", "/usr/share/elasticsearch/data"},
	}
	args := spec.GetDockerExecCmd(globalCtx.containerCfg, es.spec.ContainerCfg, es.ContainerName, true, mountPoints...)
	if es.DisableSecurity {
		args = append(args, "-e xpack.security.enabled=false")
		args = append(args, "-e xpack.security.http.ssl.enabled=false")
	}
	args = append(args, fmt.Sprintf("-e network.host=%s", es.spec.Host))
	args = append(args, fmt.Sprintf("-e http.port=%d", es.spec.Port))
	args = append(args, "-e discovery.type=single-node")
	args = append(args, "--group-add 0")
	if len(es.spec.JavaOpts) != 0 {
		args = append(args, fmt.Sprintf("-e ES_JAVA_OPTS=\"%s\"", es.spec.JavaOpts))
	}
	for k, v := range es.spec.ESConfigs {
		args = append(args, fmt.Sprintf("-e %s=%s", k, v))
	}
	args = append(args, es.spec.Image)
	return &executor.ExecuteCtx{Target: es.spec.Host, Cmd: strings.Join(args, " ")}
}

func (es *ElasticSearch) Stop(globalCtx *GlobalCtx) *executor.ExecuteCtx {
	args := []string{"docker rm -f", es.ContainerName}
	return &executor.ExecuteCtx{Target: es.spec.Host, Cmd: strings.Join(args, " ")}
}

func (es *ElasticSearch) Remove(globalCtx *GlobalCtx) *executor.ExecuteCtx {
	args := []string{"docker rm -f", es.ContainerName}
	args = append(args, "&&", "rm -rf", es.spec.DataDir, es.spec.RemoteCfgPath)
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
	spec                 spec.KibanaSpec
	ContainerName        string
	ElasticSearchHost    string
	ElasticSearchPort    int
	CheckReadyScriptPath string
}

func (k *Kibana) GetSSHHost() int {
	return k.spec.SSHPort
}

func NewKibana(kibanaSpec spec.KibanaSpec, elasticSearchHost string, elasticSearchPort int) *Kibana {
	return &Kibana{
		spec:              kibanaSpec,
		ContainerName:     spec.KibanaDefaultContainerName,
		ElasticSearchHost: elasticSearchHost,
		ElasticSearchPort: elasticSearchPort,
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
	args := append([]string{}, "mkdir -p", cfgDir, dataDir, cfgDir+"/script", "-m 0775")
	args = append(args, fmt.Sprintf("&& chown -R %[1]s:$(id -gn %[1]s) %[2]s %[3]s", globalCtx.User, cfgDir, dataDir))
	return &executor.ExecuteCtx{Target: k.spec.Host, Cmd: strings.Join(args, " ")}
}

func (k *Kibana) Deploy(globalCtx *GlobalCtx) *executor.ExecuteCtx {
	mountPoints := []spec.MountPoints{
		{filepath.Join(k.spec.RemoteCfgPath, "kibana.yml"), "/usr/share/kibana/config/kibana.yml"},
	}
	args := spec.GetDockerExecCmd(globalCtx.containerCfg, k.spec.ContainerCfg, k.ContainerName, true, mountPoints...)

	args = append(args, k.spec.Image)

	return &executor.ExecuteCtx{Target: k.spec.Host, Cmd: strings.Join(args, " ")}
}

func (k *Kibana) Stop(globalCtx *GlobalCtx) *executor.ExecuteCtx {
	args := []string{"docker rm -f", k.ContainerName}
	return &executor.ExecuteCtx{Target: k.spec.Host, Cmd: strings.Join(args, " ")}
}

func (k *Kibana) Remove(globalCtx *GlobalCtx) *executor.ExecuteCtx {
	args := []string{"docker rm -f", k.ContainerName}
	args = append(args, "&&", "rm -rf", k.spec.RemoteCfgPath, k.spec.DataDir)
	return &executor.ExecuteCtx{Target: k.spec.Host, Cmd: strings.Join(args, " ")}
}

func (k *Kibana) SyncConfig(globalCtx *GlobalCtx) *executor.TransferCtx {
	cfg := config.KibanaConfig{
		KibanaHost:        k.spec.Host,
		KibanaPort:        strconv.Itoa(k.spec.Port),
		ElasticSearchHost: k.ElasticSearchHost,
		ElasticSearchPort: strconv.Itoa(k.ElasticSearchPort),
	}
	genCfg, err := cfg.GenConfig()
	if err != nil {
		log.Errorf("gen KibanaConfig error: %s", err.Error())
		os.Exit(1)
	}

	positions := []executor.Position{
		{LocalDir: genCfg, RemoteDir: filepath.Join(k.spec.RemoteCfgPath, "kibana.yml")},
	}

	remotePath := filepath.Join(k.spec.RemoteCfgPath, "export.ndjson")

	localDir := fmt.Sprintf("template/kibana/export_%s.ndjson", "8.0.0")
	//localDir := "template/kibana/export.ndjson"

	positions = append(positions, executor.Position{
		LocalDir:  localDir,
		RemoteDir: remotePath,
	})

	chkCfg := script.KibanaReadyCheck{
		KibanaHost: k.spec.Host,
		KibanaPort: strconv.Itoa(k.spec.Port),
		FilePath:   remotePath,
		Timeout:    strconv.Itoa(120),
	}
	chkCmd, err := chkCfg.GenScript()
	if err != nil {
		log.Errorf("gen KibanaReadyCheck error: %s", err.Error())
		os.Exit(1)
	}
	scriptName := filepath.Base(chkCmd)
	cfgDir, _ := k.getDirs()
	remoteScriptPath := filepath.Join(cfgDir, "script", scriptName)
	k.CheckReadyScriptPath = remoteScriptPath
	positions = append(positions, executor.Position{
		LocalDir:  chkCmd,
		RemoteDir: remoteScriptPath,
		Opts:      fmt.Sprintf("chmod +x %s", remoteScriptPath),
	})

	return &executor.TransferCtx{
		Target: k.spec.Host, Position: positions,
	}
}

func (k *Kibana) CheckReady() *executor.ExecuteCtx {
	if len(k.CheckReadyScriptPath) == 0 {
		return nil
	}
	args := []string{"/usr/bin/env bash"}
	args = append(args, k.CheckReadyScriptPath)
	return &executor.ExecuteCtx{Target: k.spec.Host, Cmd: strings.Join(args, " ")}
}

func (k *Kibana) getDirs() (string, string) {
	return k.spec.RemoteCfgPath, k.spec.DataDir
}

type Filebeat struct {
	spec              spec.FilebeatSpec
	ContainerName     string
	ElasticsearchHost string
	ElasticsearchPort string
	KibanaHost        string
	KibanaPort        string
}

func NewFilebeat(fbSpec spec.FilebeatSpec, elasticsearchHost, elasticsearchPort, kibanaHost, kibanaPort string) *Filebeat {
	return &Filebeat{
		spec:              fbSpec,
		ContainerName:     spec.FilebeatDefaultContainerName,
		ElasticsearchHost: elasticsearchHost,
		ElasticsearchPort: elasticsearchPort,
		KibanaHost:        kibanaHost,
		KibanaPort:        kibanaPort,
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
	args := append([]string{}, "mkdir -p", cfgDir, dataDir, "-m 0775")
	args = append(args, fmt.Sprintf("&& chown -R %[1]s:$(id -gn %[1]s) %[2]s %[3]s", globalCtx.User, cfgDir, dataDir))
	return &executor.ExecuteCtx{Target: f.spec.Host, Cmd: strings.Join(args, " ")}
}

func (f *Filebeat) Deploy(globalCtx *GlobalCtx) *executor.ExecuteCtx {
	mountPoints := []spec.MountPoints{
		{"/var/lib/docker/containers", "/var/lib/docker/containers:ro"},
		{"/var/run/docker.sock", "/var/run/docker.sock:ro"},
		// FIXME: filebeat has a poor support for journald, see https://github.com/elastic/beats/issues/37086
		//{"/var/log/journal", "/var/log/journal:ro"},
		//{"/run/systemd", "/run/systemd"},
		{filepath.Join(f.spec.RemoteCfgPath, "filebeat.yml"), "/usr/share/filebeat/filebeat.yml:ro"},
	}
	args := spec.GetDockerExecCmd(globalCtx.containerCfg, f.spec.ContainerCfg, f.ContainerName, true, mountPoints...)
	args = append(args, "--user=root")

	args = append(args, f.spec.Image)
	args = append(args, "filebeat")
	args = append(args, "-e")
	args = append(args, "--strict.perms=false")
	args = append(args, fmt.Sprintf("-E setup.kibana.host=%s:%s", f.KibanaHost, f.KibanaPort))
	args = append(args, fmt.Sprintf("-E output.elasticsearch.hosts=[\"%s:%s\"]", f.ElasticsearchHost, f.ElasticsearchPort))
	return &executor.ExecuteCtx{Target: f.spec.Host, Cmd: strings.Join(args, " ")}
}

func (f *Filebeat) Stop(globalCtx *GlobalCtx) *executor.ExecuteCtx {
	args := []string{"docker rm -f", f.ContainerName}
	return &executor.ExecuteCtx{Target: f.spec.Host, Cmd: strings.Join(args, " ")}
}

func (f *Filebeat) Remove(globalCtx *GlobalCtx) *executor.ExecuteCtx {
	args := []string{"docker rm -f", f.ContainerName}
	args = append(args, "&&", "rm -rf", f.spec.RemoteCfgPath, f.spec.DataDir)
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
		log.Errorf("gen FilebeatConfig error: %s", err.Error())
		os.Exit(1)
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
