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
	"regexp"
	"strconv"
	"strings"
)

type ElasticSearch struct {
	spec            spec.ElasticSearchSpec
	ContainerName   string
	DisableSecurity bool
}

func isElasticSearchImageOss(esSpec spec.ElasticSearchSpec) bool {
	if esSpec.IsOss != nil {
		return *esSpec.IsOss
	} else {
		return strings.Contains(esSpec.Image, "elasticsearch-oss")
	}
}

func isKibanaImageOss(k spec.KibanaSpec) bool {
	if k.IsOss != nil {
		return *k.IsOss
	} else {
		return strings.Contains(k.Image, "kibana-oss")
	}
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
	args := append([]string{}, "sudo mkdir -p", cfgDir, dataDir, "-m 0775")
	args = append(args, fmt.Sprintf("&& sudo chown -R %[1]s:$(id -gn %[1]s) %[2]s %[3]s", globalCtx.User, cfgDir, dataDir))
	return &executor.ExecuteCtx{Target: es.spec.Host, Cmd: strings.Join(args, " ")}
}

func (es *ElasticSearch) Deploy(globalCtx *GlobalCtx) *executor.ExecuteCtx {
	args := spec.GetDockerExecCmd(globalCtx.containerCfg, es.spec.ContainerCfg, es.ContainerName, true)
	if es.DisableSecurity && !isElasticSearchImageOss(es.spec) {
		args = append(args, "-e xpack.security.enabled=false")
		args = append(args, "-e xpack.security.http.ssl.enabled=false")
	}
	args = append(args, fmt.Sprintf("-e network.host='%s'", es.spec.Host))
	args = append(args, fmt.Sprintf("-e http.port='%s'", strconv.Itoa(es.spec.Port)))
	args = append(args, "-e discovery.type=single-node")
	args = append(args, es.spec.Image)
	return &executor.ExecuteCtx{Target: es.spec.Host, Cmd: strings.Join(args, " ")}
}

func (es *ElasticSearch) Stop(globalCtx *GlobalCtx) *executor.ExecuteCtx {
	args := []string{"docker rm -f", es.ContainerName}
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
	args := append([]string{}, "sudo mkdir -p", cfgDir, dataDir, cfgDir+"/script", "-m 0775")
	args = append(args, fmt.Sprintf("&& sudo chown -R %[1]s:$(id -gn %[1]s) %[2]s %[3]s", globalCtx.User, cfgDir, dataDir))
	return &executor.ExecuteCtx{Target: k.spec.Host, Cmd: strings.Join(args, " ")}
}

func (k *Kibana) Deploy(globalCtx *GlobalCtx) *executor.ExecuteCtx {
	mountPoints := []spec.MountPoints{
		{filepath.Join(k.spec.RemoteCfgPath, "kibana.yml"), "/usr/share/kibana/config/kibana.yml"},
	}
	args := spec.GetDockerExecCmd(globalCtx.containerCfg, k.spec.ContainerCfg, k.ContainerName, true, mountPoints...)

	var imageTag string
	imageSplit := strings.Split(k.spec.Image, ":")
	if len(imageSplit) == 2 {
		imageTag = imageSplit[1]
	}
	if enableServerShutdownTimeout(imageTag) {
		args = append(args, "-e server.shutdownTimeout=5s")
	}

	if !isKibanaImageOss(k.spec) {
		args = append(args, "-e monitoring.ui.container.elasticsearch.enabled=true")
	}

	args = append(args, k.spec.Image)

	return &executor.ExecuteCtx{Target: k.spec.Host, Cmd: strings.Join(args, " ")}
}

func (k *Kibana) Stop(globalCtx *GlobalCtx) *executor.ExecuteCtx {
	args := []string{"docker rm -f", k.ContainerName}
	return &executor.ExecuteCtx{Target: k.spec.Host, Cmd: strings.Join(args, " ")}
}

func (k *Kibana) Remove(globalCtx *GlobalCtx) *executor.ExecuteCtx {
	args := []string{"docker rm -f", k.ContainerName}
	args = append(args, "&&", "sudo rm -rf", k.spec.RemoteCfgPath, k.spec.DataDir)
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

	var imageTag string
	imageSplit := strings.Split(k.spec.Image, ":")
	if len(imageSplit) == 2 {
		imageTag = imageSplit[1]
	}
	postfix := whichIndexPatternToUse(imageTag)
	localDir := fmt.Sprintf("template/kibana/export_%s.ndjson", postfix)

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
		Opts:      fmt.Sprintf("sudo chmod +x %s", remoteScriptPath),
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
	args := append([]string{}, "sudo mkdir -p", cfgDir, dataDir, "-m 0775")
	args = append(args, fmt.Sprintf("&& sudo chown -R %[1]s:$(id -gn %[1]s) %[2]s %[3]s", globalCtx.User, cfgDir, dataDir))
	return &executor.ExecuteCtx{Target: f.spec.Host, Cmd: strings.Join(args, " ")}
}

func (f *Filebeat) Deploy(globalCtx *GlobalCtx) *executor.ExecuteCtx {
	mountPoints := []spec.MountPoints{
		{"/var/lib/docker/containers", "/var/lib/docker/containers:ro"},
		{"/var/run/docker.sock", "/var/run/docker.sock:ro"},
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

func enableServerShutdownTimeout(kibanaImageTag string) bool {
	const assumeMsg = ", assume the version is higher than 7.13.0, set `ServerShutdownTimeout` to default value"

	kibanaImageTag = strings.TrimSpace(kibanaImageTag)

	if kibanaImageTag == "" {
		log.Warn("Kibana image tag is empty" + assumeMsg)
		return true
	}

	if kibanaImageTag == "latest" {
		log.Warnf("Use custom Kibana tag `%s`"+assumeMsg, kibanaImageTag)
		return true
	}

	ret, err := isVersionCompatible(kibanaImageTag, 7, 13, 0)
	if err != nil {
		log.Warnf("Can not parse Kibana image tag to version number: %s"+assumeMsg, err)
		return true
	}
	return ret
}

func whichIndexPatternToUse(kibanaImageTag string) string {
	const (
		available800 = "8.0.0"
		available760 = "7.6.0"

		subAssumeMsg = ", assume the version is higher than 8.0.0, use the 8.0.0 compatible index patterns"
		assumeMsg    = "Can not parse Kibana image tag to version number: %s" + subAssumeMsg
	)

	kibanaImageTag = strings.TrimSpace(kibanaImageTag)

	if kibanaImageTag == "" {
		log.Warn("Kibana image tag is empty" + subAssumeMsg)
		return available800
	}

	if kibanaImageTag == "latest" {
		log.Warnf("Use custom Kibana tag `%s`"+subAssumeMsg, kibanaImageTag)
		return available800
	}

	ret, err := isVersionCompatible(kibanaImageTag, 8, 0, 0)
	if err != nil {
		log.Warnf(assumeMsg, err)
		return available800
	}
	if ret {
		return available800
	}

	ret, err = isVersionCompatible(kibanaImageTag, 7, 6, 0)
	if err != nil {
		log.Warnf(assumeMsg, err)
		return available800
	}
	if ret {
		return available760
	} else {
		panic(fmt.Sprintf("The version `%s` is not supported by current hdt. Please use an image version higher than %s or %s (for `-oss` users)",
			kibanaImageTag,
			available800,
			available760))
	}

}

func isVersionCompatible(version string, requiredMajor int, requiredMinor int, requiredPatch int) (bool, error) {
	versionRegex, err := regexp.Compile(`^(\d+)\.(\d+)(?:\.(\d+))?$`)
	matches := versionRegex.FindStringSubmatch(version)

	if len(matches) > 0 {
		major, _ := strconv.Atoi(matches[1])
		minor, _ := strconv.Atoi(matches[2])
		patch := 0
		if matches[3] != "" {
			patch, _ = strconv.Atoi(matches[3])
		}

		if major > requiredMajor || (major == requiredMajor && minor >= requiredMinor) || (major == requiredMajor && minor == requiredMinor && patch >= requiredPatch) {
			return true, err
		}
	}

	return false, err
}
