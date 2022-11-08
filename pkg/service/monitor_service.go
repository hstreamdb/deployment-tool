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

type MonitorSuite struct {
	Host                  string
	spec                  spec.MonitorSpec
	NodeContainerName     string
	CadvisorContainerName string
}

func NewMonitorSuite(host string, moSpec spec.MonitorSpec) *MonitorSuite {
	return &MonitorSuite{
		Host:                  host,
		spec:                  moSpec,
		NodeContainerName:     spec.NodeExporterDefaultContainerName,
		CadvisorContainerName: spec.CadvisorDefaultContainerName,
	}
}

func (m *MonitorSuite) GetServiceName() string {
	return "monitor suite"
}

func (m *MonitorSuite) Display() map[string]utils.DisplayedComponent {
	cfgDir, dataDir := m.getDirs()
	nodeContainer := utils.DisplayedComponent{
		Name:          "NodeExporter",
		Host:          m.Host,
		Ports:         strconv.Itoa(m.spec.NodeExporterPort),
		ContainerName: m.NodeContainerName,
		Image:         m.spec.NodeExporterImage,
		Paths:         strings.Join([]string{cfgDir, dataDir}, ","),
	}
	cadVisorContainer := utils.DisplayedComponent{
		Name:          "Cadvisor",
		Host:          m.Host,
		Ports:         strconv.Itoa(m.spec.CadvisorPort),
		ContainerName: m.CadvisorContainerName,
		Image:         m.spec.CadvisorImage,
		Paths:         strings.Join([]string{cfgDir, dataDir}, ","),
	}
	return map[string]utils.DisplayedComponent{
		"nodeExporter": nodeContainer,
		"cadVisor":     cadVisorContainer,
	}
}

func (m *MonitorSuite) InitEnv(globalCtx *GlobalCtx) *executor.ExecuteCtx {
	cfgDir, dataDir := m.getDirs()
	args := append([]string{}, "sudo mkdir -p", cfgDir, dataDir)
	return &executor.ExecuteCtx{Target: m.Host, Cmd: strings.Join(args, " ")}
}

func (m *MonitorSuite) Deploy(globalCtx *GlobalCtx) *executor.ExecuteCtx {
	nodeMountP := []spec.MountPoints{
		{Local: "/proc", Remote: "/host/proc:ro"},
		{Local: "/sys", Remote: "/host/sys:ro"},
		{Local: "/", Remote: "/rootfs:ro"},
	}
	startNodeExporter := spec.GetDockerExecCmd(globalCtx.containerCfg, m.spec.ContainerCfg, m.NodeContainerName, true, nodeMountP...)
	args := append(startNodeExporter, m.spec.NodeExporterImage)
	args = append(args, "--path.procfs=/host/proc", "--path.rootfs=/rootfs", "--path.sysfs=/host/sys", "&&")

	cardvisorMountP := []spec.MountPoints{
		{Local: "/", Remote: "/rootfs:ro"},
		{Local: "/var/run", Remote: "/var/run:ro"},
		{Local: "/sys", Remote: "/sys:ro"},
		{Local: "/var/lib/docker/", Remote: "/var/lib/docker:ro"},
		{Local: "/dev/disk/", Remote: "/dev/disk:ro"},
	}
	startCadvisor := spec.GetDockerExecCmd(globalCtx.containerCfg, m.spec.ContainerCfg, m.CadvisorContainerName, false, cardvisorMountP...)
	args = append(args, startCadvisor...)
	args = append(args, fmt.Sprintf("-p %d:8080", m.spec.CadvisorPort))
	args = append(args, "--detach=true", "--privileged=true", "--device /dev/kmsg")
	args = append(args, m.spec.CadvisorImage)
	return &executor.ExecuteCtx{Target: m.Host, Cmd: strings.Join(args, " ")}
}

func (m *MonitorSuite) Remove(globalCtx *GlobalCtx) *executor.ExecuteCtx {
	args := []string{"docker rm -f", m.NodeContainerName, m.CadvisorContainerName}
	args = append(args, "&&", "sudo rm -rf", m.spec.DataDir, m.spec.RemoteCfgPath)
	return &executor.ExecuteCtx{Target: m.Host, Cmd: strings.Join(args, " ")}
}

func (m *MonitorSuite) SyncConfig(globalCtx *GlobalCtx) *executor.TransferCtx {
	return nil
}

func (m *MonitorSuite) getDirs() (string, string) {
	return m.spec.RemoteCfgPath, m.spec.DataDir
}

type Prometheus struct {
	spec                spec.PrometheusSpec
	ContainerName       string
	MonitoredHosts      []string
	NodeExporterPort    int
	CadvisorPort        int
	HStreamExporterAddr []string
	AlertManagerAddr    []string
}

func NewPrometheus(promSpec spec.PrometheusSpec, monitorSuites []*MonitorSuite, hstreamExporterAddr []string, alertAddr []string) *Prometheus {
	hosts := make([]string, 0, len(monitorSuites))
	for _, suite := range monitorSuites {
		hosts = append(hosts, suite.Host)
	}
	return &Prometheus{
		spec:                promSpec,
		ContainerName:       spec.PrometheusDefaultContainerName,
		MonitoredHosts:      hosts,
		NodeExporterPort:    monitorSuites[0].spec.NodeExporterPort,
		CadvisorPort:        monitorSuites[0].spec.CadvisorPort,
		HStreamExporterAddr: hstreamExporterAddr,
		AlertManagerAddr:    alertAddr,
	}
}

func (p *Prometheus) GetServiceName() string {
	return "prometheus"
}

func (p *Prometheus) Display() map[string]utils.DisplayedComponent {
	cfgDir, dataDir := p.getDirs()
	prometheus := utils.DisplayedComponent{
		Name:          "Prometheus",
		Host:          p.spec.Host,
		Ports:         strconv.Itoa(p.spec.Port),
		ContainerName: p.ContainerName,
		Image:         p.spec.Image,
		Paths:         strings.Join([]string{cfgDir, dataDir}, ","),
	}
	return map[string]utils.DisplayedComponent{"prometheus": prometheus}
}

func (p *Prometheus) InitEnv(globalCtx *GlobalCtx) *executor.ExecuteCtx {
	cfgDir, dataDir := p.getDirs()
	args := append([]string{}, "sudo mkdir -p", cfgDir, dataDir)
	return &executor.ExecuteCtx{Target: p.spec.Host, Cmd: strings.Join(args, " ")}
}

func (p *Prometheus) Deploy(globalCtx *GlobalCtx) *executor.ExecuteCtx {
	mountPoints := []spec.MountPoints{
		{p.spec.RemoteCfgPath, "/etc/prometheus"},
	}
	args := spec.GetDockerExecCmd(globalCtx.containerCfg, p.spec.ContainerCfg, p.ContainerName, true, mountPoints...)
	args = append(args, p.spec.Image)
	return &executor.ExecuteCtx{Target: p.spec.Host, Cmd: strings.Join(args, " ")}
}

func (p *Prometheus) Remove(globalCtx *GlobalCtx) *executor.ExecuteCtx {
	args := []string{"docker rm -f", p.ContainerName}
	args = append(args, "&&", "sudo rm -rf", p.spec.DataDir, p.spec.RemoteCfgPath)
	return &executor.ExecuteCtx{Target: p.spec.Host, Cmd: strings.Join(args, " ")}
}

func (p *Prometheus) SyncConfig(globalCtx *GlobalCtx) *executor.TransferCtx {
	nodeAddr := make([]string, 0, len(p.MonitoredHosts))
	cadAddr := make([]string, 0, len(p.MonitoredHosts))
	for _, host := range p.MonitoredHosts {
		nodeAddr = append(nodeAddr, fmt.Sprintf("%s:%d", host, p.NodeExporterPort))
		cadAddr = append(cadAddr, fmt.Sprintf("%s:%d", host, p.CadvisorPort))
	}
	prometheusCfg := config.PrometheusConfig{
		NodeExporterAddress:    nodeAddr,
		CadVisorAddress:        cadAddr,
		HStreamExporterAddress: p.HStreamExporterAddr,
		AlertManagerAddress:    p.AlertManagerAddr,
	}
	cfg, err := prometheusCfg.GenConfig()
	if err != nil {
		panic(fmt.Errorf("gen prometheusCfg error: %s", err.Error()))
	}

	position := utils.ScpDir(filepath.Dir(cfg), p.spec.RemoteCfgPath)

	return &executor.TransferCtx{
		Target: p.spec.Host, Position: position,
	}
}

func (p *Prometheus) getDirs() (string, string) {
	return p.spec.RemoteCfgPath, p.spec.DataDir
}

type Grafana struct {
	spec          spec.GrafanaSpec
	ContainerName string
	DisableLogin  bool
}

func NewGrafana(graSpec spec.GrafanaSpec, disableLogin bool) *Grafana {
	return &Grafana{spec: graSpec, ContainerName: spec.GrafanaDefaultContainerName, DisableLogin: disableLogin}
}

func (g *Grafana) GetServiceName() string {
	return "grafana"
}

func (g *Grafana) Display() map[string]utils.DisplayedComponent {
	cfgDir, dataDir := g.getDirs()
	grafana := utils.DisplayedComponent{
		Name:          "Grafana",
		Host:          g.spec.Host,
		Ports:         strconv.Itoa(g.spec.Port),
		ContainerName: g.ContainerName,
		Image:         g.spec.Image,
		Paths:         strings.Join([]string{cfgDir, dataDir}, ","),
	}
	return map[string]utils.DisplayedComponent{"grafana": grafana}
}

func (g *Grafana) InitEnv(globalCtx *GlobalCtx) *executor.ExecuteCtx {
	cfgDir, dataDir := g.getDirs()
	args := append([]string{}, "sudo mkdir -p", cfgDir, dataDir,
		filepath.Join(cfgDir, "dashboards"), filepath.Join(cfgDir, "datasources"))
	return &executor.ExecuteCtx{Target: g.spec.Host, Cmd: strings.Join(args, " ")}
}

func (g *Grafana) Deploy(globalCtx *GlobalCtx) *executor.ExecuteCtx {
	mountPoints := []spec.MountPoints{
		{g.spec.RemoteCfgPath, "/etc/grafana/provisioning"},
	}
	args := spec.GetDockerExecCmd(globalCtx.containerCfg, g.spec.ContainerCfg, g.ContainerName, true, mountPoints...)
	if g.DisableLogin {
		args = append(args, "-e GF_AUTH_ANONYMOUS_ORG_ROLE=Admin",
			"-e GF_AUTH_ANONYMOUS_ENABLED=true", "-e GF_AUTH_DISABLE_LOGIN_FORM=true")
	}
	args = append(args, g.spec.Image)
	return &executor.ExecuteCtx{Target: g.spec.Host, Cmd: strings.Join(args, " ")}
}

func (g *Grafana) Remove(globalCtx *GlobalCtx) *executor.ExecuteCtx {
	args := []string{"docker rm -f", g.ContainerName}
	args = append(args, "&&", "sudo rm -rf", g.spec.DataDir, g.spec.RemoteCfgPath)
	return &executor.ExecuteCtx{Target: g.spec.Host, Cmd: strings.Join(args, " ")}
}

func (g *Grafana) SyncConfig(globalCtx *GlobalCtx) *executor.TransferCtx {
	grafanaCfg := config.GrafanaConfig{}
	cfg, err := grafanaCfg.GenConfig()
	if err != nil {
		panic(fmt.Errorf("gen grafanaCfg error: %s", err.Error()))
	}
	position := utils.ScpDir(cfg, g.spec.RemoteCfgPath)

	return &executor.TransferCtx{
		Target: g.spec.Host, Position: position,
	}
}

func (g *Grafana) getDirs() (string, string) {
	return g.spec.RemoteCfgPath, g.spec.DataDir
}

type AlertManager struct {
	spec          spec.AlertManagerSpec
	ContainerName string
	DisableLogin  bool
}

func NewAlertManager(graSpec spec.AlertManagerSpec) *AlertManager {
	return &AlertManager{spec: graSpec, ContainerName: spec.AlertManagerDefaultContainerName}
}

func (a *AlertManager) GetServiceName() string {
	return "alertManager"
}

func (a *AlertManager) Display() map[string]utils.DisplayedComponent {
	cfgDir, dataDir := a.getDirs()
	alert := utils.DisplayedComponent{
		Name:          "AlertManager",
		Host:          a.spec.Host,
		Ports:         strconv.Itoa(a.spec.Port),
		ContainerName: a.ContainerName,
		Image:         a.spec.Image,
		Paths:         strings.Join([]string{cfgDir, dataDir}, ","),
	}
	return map[string]utils.DisplayedComponent{"alertManager": alert}
}

func (a *AlertManager) InitEnv(globalCtx *GlobalCtx) *executor.ExecuteCtx {
	cfgDir, dataDir := a.getDirs()
	args := append([]string{}, "sudo mkdir -p", cfgDir, dataDir)
	return &executor.ExecuteCtx{Target: a.spec.Host, Cmd: strings.Join(args, " ")}
}

func (a *AlertManager) Deploy(globalCtx *GlobalCtx) *executor.ExecuteCtx {
	mountPoints := []spec.MountPoints{
		{a.spec.RemoteCfgPath, "/etc/alertmanager"},
	}
	args := spec.GetDockerExecCmd(globalCtx.containerCfg, a.spec.ContainerCfg, a.ContainerName, true, mountPoints...)
	args = append(args, a.spec.Image)
	return &executor.ExecuteCtx{Target: a.spec.Host, Cmd: strings.Join(args, " ")}
}

func (a *AlertManager) Remove(globalCtx *GlobalCtx) *executor.ExecuteCtx {
	args := []string{"docker rm -f", a.ContainerName}
	args = append(args, "&&", "sudo rm -rf", a.spec.DataDir, a.spec.RemoteCfgPath)
	return &executor.ExecuteCtx{Target: a.spec.Host, Cmd: strings.Join(args, " ")}
}

func (a *AlertManager) SyncConfig(globalCtx *GlobalCtx) *executor.TransferCtx {
	position := utils.ScpDir(filepath.Dir("template/alertmanager/alertmanager.yml"), a.spec.RemoteCfgPath)

	return &executor.TransferCtx{
		Target: a.spec.Host, Position: position,
	}
}

func (a *AlertManager) getDirs() (string, string) {
	return a.spec.RemoteCfgPath, a.spec.DataDir
}

type HStreamExporter struct {
	spec          spec.HStreamExporterSpec
	ContainerName string
}

func NewHStreamExporter(exporterSpec spec.HStreamExporterSpec) *HStreamExporter {
	return &HStreamExporter{spec: exporterSpec, ContainerName: spec.HStreamExporterDefaultContainerName}
}

func (h *HStreamExporter) GetServiceName() string {
	return "hstream-exporter"
}

func (h *HStreamExporter) Display() map[string]utils.DisplayedComponent {
	cfgDir, dataDir := h.getDirs()
	exporter := utils.DisplayedComponent{
		Name:          "HStreamExporter",
		Host:          h.spec.Host,
		Ports:         strconv.Itoa(h.spec.Port),
		ContainerName: h.ContainerName,
		Image:         h.spec.Image,
		Paths:         strings.Join([]string{cfgDir, dataDir}, ","),
	}
	return map[string]utils.DisplayedComponent{"hstreamExporter": exporter}
}

func (h *HStreamExporter) InitEnv(globalCtx *GlobalCtx) *executor.ExecuteCtx {
	cfgDir, dataDir := h.getDirs()
	args := append([]string{}, "sudo mkdir -p", cfgDir, dataDir)
	return &executor.ExecuteCtx{Target: h.spec.Host, Cmd: strings.Join(args, " ")}
}

func (h *HStreamExporter) Deploy(globalCtx *GlobalCtx) *executor.ExecuteCtx {
	args := spec.GetDockerExecCmd(globalCtx.containerCfg, h.spec.ContainerCfg, h.ContainerName, true)
	args = append(args, h.spec.Image)
	// FIXME: currently, only support use one http-server
	httpServer := globalCtx.HttpServerUrls[0]
	args = append(args, "hstream-exporter", "--addr", httpServer)
	return &executor.ExecuteCtx{Target: h.spec.Host, Cmd: strings.Join(args, " ")}
}

func (h *HStreamExporter) Remove(globalCtx *GlobalCtx) *executor.ExecuteCtx {
	args := []string{"docker rm -f", h.ContainerName}
	args = append(args, "&&", "sudo rm -rf", h.spec.DataDir, h.spec.RemoteCfgPath)
	return &executor.ExecuteCtx{Target: h.spec.Host, Cmd: strings.Join(args, " ")}
}

func (h *HStreamExporter) SyncConfig(globalCtx *GlobalCtx) *executor.TransferCtx {
	return nil
}

func (h *HStreamExporter) getDirs() (string, string) {
	return h.spec.RemoteCfgPath, h.spec.DataDir
}

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
	spec          spec.FilebeatSpec
	ContainerName string
}

func NewElasticSearch(esSpec spec.ElasticSearchSpec, disableSecurity bool) *ElasticSearch {
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
	mountPoints := []spec.MountPoints{}
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
	mountPoints := []spec.MountPoints{}
	args := spec.GetDockerExecCmd(globalCtx.containerCfg, k.spec.ContainerCfg, k.ContainerName, true, mountPoints...)
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

func NewFilebeat(fbSpec spec.FilebeatSpec) *Filebeat {
	return &Filebeat{
		spec:          fbSpec,
		ContainerName: spec.FilebeatDefaultContainerName,
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
	}
	if fb.spec.LocalCfgPath != "" {
		mountPoints = append(mountPoints, spec.MountPoints{Local: fb.spec.LocalCfgPath, Remote: "/usr/share/filebeat/filebeat.yml"})
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
	localCfg := fb.spec.LocalCfgPath
	if localCfg == "" {
		return nil
	}
	position := utils.ScpDir(fb.spec.LocalCfgPath, fb.spec.RemoteCfgPath)
	return &executor.TransferCtx{
		Target: fb.spec.Host, Position: position,
	}
}
