package service

import (
	"fmt"
	"github.com/hstreamdb/deployment-tool/pkg/executor"
	"github.com/hstreamdb/deployment-tool/pkg/spec"
	"github.com/hstreamdb/deployment-tool/pkg/template/config"
	"github.com/hstreamdb/deployment-tool/pkg/utils"
	log "github.com/sirupsen/logrus"
	"os"
	"path/filepath"
	"regexp"
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
	args := append([]string{}, "mkdir -p", cfgDir, dataDir, "-m 0775")
	args = append(args, fmt.Sprintf("&& chown -R %[1]s:$(id -gn %[1]s) %[2]s %[3]s", globalCtx.User, cfgDir, dataDir))
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
		// for suse, seems should use /var/run, see: https://github.com/google/cadvisor/issues/2671#issuecomment-851346288
		{Local: "/var/run", Remote: "/var/run:ro"},
		{Local: "/sys", Remote: "/sys:ro"},
		{Local: "/var/lib/docker/", Remote: "/var/lib/docker:ro"},
		{Local: "/dev/disk/", Remote: "/dev/disk:ro"},
	}
	startCadvisor := spec.GetDockerExecCmd(globalCtx.containerCfg, m.spec.ContainerCfg, m.CadvisorContainerName, false, cardvisorMountP...)
	args = append(args, startCadvisor...)
	args = append(args, fmt.Sprintf("-p %d:8080", m.spec.CadvisorPort))
	//args = append(args, "--detach=true", "--privileged=true", "--device /dev/kmsg")
	args = append(args, "--detach=true", "--device /dev/kmsg")
	args = append(args, m.spec.CadvisorImage)
	return &executor.ExecuteCtx{Target: m.Host, Cmd: strings.Join(args, " ")}
}

func (m *MonitorSuite) Stop(globalCtx *GlobalCtx) *executor.ExecuteCtx {
	args := []string{"docker rm -f", m.NodeContainerName, m.CadvisorContainerName}
	return &executor.ExecuteCtx{Target: m.Host, Cmd: strings.Join(args, " ")}
}

func (m *MonitorSuite) Remove(globalCtx *GlobalCtx) *executor.ExecuteCtx {
	args := []string{"docker rm -f", m.NodeContainerName, m.CadvisorContainerName}
	args = append(args, "&&", "rm -rf", m.spec.DataDir, m.spec.RemoteCfgPath)
	return &executor.ExecuteCtx{Target: m.Host, Cmd: strings.Join(args, " ")}
}

func (m *MonitorSuite) SyncConfig(globalCtx *GlobalCtx) *executor.TransferCtx {
	return nil
}

func (m *MonitorSuite) getDirs() (string, string) {
	return m.spec.RemoteCfgPath, m.spec.DataDir
}

// ================================================================================
// 	BlackBox

type BlackBox struct {
	spec           spec.BlackBoxSpec
	ContainerName  string
	MonitoredHosts []string
}

func NewBlackBox(blackBoxSpec spec.BlackBoxSpec) *BlackBox {
	return &BlackBox{spec: blackBoxSpec, ContainerName: spec.BlackBoxDefaultContainerName}
}

func (b *BlackBox) GetServiceName() string {
	return "blackbox"
}

func (b *BlackBox) Display() map[string]utils.DisplayedComponent {
	cfgDir, dataDir := b.getDirs()
	blackBox := utils.DisplayedComponent{
		Name:          "BlackBox",
		Host:          b.spec.Host,
		Ports:         strconv.Itoa(b.spec.Port),
		ContainerName: b.ContainerName,
		Image:         b.spec.Image,
		Paths:         strings.Join([]string{cfgDir, dataDir}, ","),
	}
	return map[string]utils.DisplayedComponent{"blackbox": blackBox}
}

func (b *BlackBox) InitEnv(globalCtx *GlobalCtx) *executor.ExecuteCtx {
	cfgDir, dataDir := b.getDirs()
	args := append([]string{}, "mkdir -p", cfgDir, dataDir, "-m 0775")
	args = append(args, fmt.Sprintf("&& chown -R %[1]s:$(id -gn %[1]s) %[2]s %[3]s", globalCtx.User, cfgDir, dataDir))
	return &executor.ExecuteCtx{Target: b.spec.Host, Cmd: strings.Join(args, " ")}
}

func (b *BlackBox) Deploy(globalCtx *GlobalCtx) *executor.ExecuteCtx {
	mountPoints := []spec.MountPoints{
		{b.spec.RemoteCfgPath, "/etc/blackbox_exporter"},
	}
	args := spec.GetDockerExecCmd(globalCtx.containerCfg, b.spec.ContainerCfg, b.ContainerName, true, mountPoints...)
	args = append(args, b.spec.Image, "--config.file=/etc/blackbox_exporter/blackbox.yml")
	return &executor.ExecuteCtx{Target: b.spec.Host, Cmd: strings.Join(args, " ")}
}

func (b *BlackBox) Stop(globalCtx *GlobalCtx) *executor.ExecuteCtx {
	args := []string{"docker rm -f", b.ContainerName}
	return &executor.ExecuteCtx{Target: b.spec.Host, Cmd: strings.Join(args, " ")}
}

func (b *BlackBox) Remove(globalCtx *GlobalCtx) *executor.ExecuteCtx {
	args := []string{"docker rm -f", b.ContainerName}
	args = append(args, "&&", "rm -rf", b.spec.DataDir, b.spec.RemoteCfgPath)
	return &executor.ExecuteCtx{Target: b.spec.Host, Cmd: strings.Join(args, " ")}
}

func (b *BlackBox) SyncConfig(globalCtx *GlobalCtx) *executor.TransferCtx {
	position := utils.ScpDir("template/blackbox/blackbox.yml", b.spec.RemoteCfgPath)

	return &executor.TransferCtx{
		Target: b.spec.Host, Position: position,
	}
}

func (b *BlackBox) getDirs() (string, string) {
	return b.spec.RemoteCfgPath, b.spec.DataDir
}

// ================================================================================
// 	Prometheus

type Prometheus struct {
	spec                spec.PrometheusSpec
	ContainerName       string
	MonitoredHosts      []string
	NodeExporterPort    int
	CadvisorPort        int
	HStreamExporterAddr []string
	AlertManagerAddr    []string
	BlackBoxAddr        string
}

func NewPrometheus(promSpec spec.PrometheusSpec, monitorSuites []*MonitorSuite,
	hstreamExporterAddr []string, alertAddr []string, blackBoxAddr string) *Prometheus {
	hosts := make([]string, 0, len(monitorSuites))
	for _, suite := range monitorSuites {
		hosts = append(hosts, suite.Host)
	}

	res := &Prometheus{
		spec:                promSpec,
		ContainerName:       spec.PrometheusDefaultContainerName,
		MonitoredHosts:      hosts,
		HStreamExporterAddr: hstreamExporterAddr,
		AlertManagerAddr:    alertAddr,
		BlackBoxAddr:        blackBoxAddr,
	}

	if len(monitorSuites) != 0 {
		res.NodeExporterPort = monitorSuites[0].spec.NodeExporterPort
		res.CadvisorPort = monitorSuites[0].spec.CadvisorPort
	}
	return res
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
	args := append([]string{}, "mkdir -p", cfgDir, dataDir, "-m 0775")
	args = append(args, fmt.Sprintf("&& chown -R %[1]s:$(id -gn %[1]s) %[2]s %[3]s", globalCtx.User, cfgDir, dataDir))
	return &executor.ExecuteCtx{Target: p.spec.Host, Cmd: strings.Join(args, " ")}
}

func (p *Prometheus) Deploy(globalCtx *GlobalCtx) *executor.ExecuteCtx {
	mountPoints := []spec.MountPoints{
		{p.spec.RemoteCfgPath, "/etc/prometheus"},
		{p.spec.DataDir, "/prometheus"},
	}
	args := spec.GetDockerExecCmd(globalCtx.containerCfg, p.spec.ContainerCfg, p.ContainerName, true, mountPoints...)
	// FIXME: set user to root to make sure files under /prometheus can be created and written correctly.
	// Refer to: https://github.com/prometheus/prometheus/issues/5976
	// Find another way to give correct permission to /prometheus
	args = append(args, "--user $(id -u)", p.spec.Image)
	args = append(args, fmt.Sprintf("--storage.tsdb.retention.time=%s", p.spec.RetentionTime))
	args = append(args, "--config.file=/etc/prometheus/prometheus.yml")
	return &executor.ExecuteCtx{Target: p.spec.Host, Cmd: strings.Join(args, " ")}
}

func (p *Prometheus) Stop(globalCtx *GlobalCtx) *executor.ExecuteCtx {
	args := []string{"docker rm -f", p.ContainerName}
	return &executor.ExecuteCtx{Target: p.spec.Host, Cmd: strings.Join(args, " ")}
}

func (p *Prometheus) Remove(globalCtx *GlobalCtx) *executor.ExecuteCtx {
	args := []string{"docker rm -f", p.ContainerName}
	args = append(args, "&&", "rm -rf", p.spec.DataDir, p.spec.RemoteCfgPath)
	return &executor.ExecuteCtx{Target: p.spec.Host, Cmd: strings.Join(args, " ")}
}

func (p *Prometheus) SyncConfig(globalCtx *GlobalCtx) *executor.TransferCtx {
	allServiceAddr := globalCtx.ServiceAddr

	nodeAddr := make([]string, 0, len(p.MonitoredHosts))
	cadAddr := make([]string, 0, len(p.MonitoredHosts))
	for _, host := range p.MonitoredHosts {
		node := fmt.Sprintf("%s:%d", host, p.NodeExporterPort)
		cad := fmt.Sprintf("%s:%d", host, p.CadvisorPort)
		nodeAddr = append(nodeAddr, node)
		cadAddr = append(cadAddr, cad)
	}

	allServiceAddr["node-exporter"] = nodeAddr
	allServiceAddr["cadvisor"] = cadAddr
	metaZkAddress := []string{}
	if globalCtx.MetaStoreType == spec.ZK {
		metaZkAddress = allServiceAddr["meta_store"]
		reg := regexp.MustCompile("(.*):.*")
		for i, addr := range metaZkAddress {
			metaZkAddress[i] = reg.ReplaceAllString(addr, "$1:7070")
		}
		delete(allServiceAddr, "meta_store")
	}
	prometheusCfg := config.PrometheusConfig{
		ClusterId:              globalCtx.ClusterId,
		NodeExporterAddress:    nodeAddr,
		CadVisorAddress:        cadAddr,
		HStreamExporterAddress: p.HStreamExporterAddr,
		AlertManagerAddress:    p.AlertManagerAddr,
		BlackBoxAddress:        p.BlackBoxAddr,
		BlackBoxTargets:        allServiceAddr,
		MetaZkAddress:          metaZkAddress,
	}
	cfg, err := prometheusCfg.GenConfig()
	if err != nil {
		log.Errorf("gen prometheusCfg error: %s", err.Error())
		os.Exit(1)
	}

	position := utils.ScpDir(filepath.Dir(cfg), p.spec.RemoteCfgPath)

	return &executor.TransferCtx{
		Target: p.spec.Host, Position: position,
	}
}

func (p *Prometheus) getDirs() (string, string) {
	return p.spec.RemoteCfgPath, p.spec.DataDir
}

// ================================================================================
// 	Grafana

type Grafana struct {
	spec          spec.GrafanaSpec
	ContainerName string
}

func NewGrafana(graSpec spec.GrafanaSpec) *Grafana {
	return &Grafana{spec: graSpec, ContainerName: spec.GrafanaDefaultContainerName}
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
	args := append([]string{}, "mkdir -p", cfgDir, dataDir,
		filepath.Join(cfgDir, "dashboards"), filepath.Join(cfgDir, "datasources"), "-m 0775")
	args = append(args, fmt.Sprintf("&& chown -R %[1]s:$(id -gn %[1]s) %[2]s %[3]s", globalCtx.User, cfgDir, dataDir))
	return &executor.ExecuteCtx{Target: g.spec.Host, Cmd: strings.Join(args, " ")}
}

func (g *Grafana) Deploy(globalCtx *GlobalCtx) *executor.ExecuteCtx {
	mountPoints := []spec.MountPoints{
		{g.spec.RemoteCfgPath, "/etc/grafana/provisioning"},
		{g.spec.DataDir, "/var/lib/grafana"},
	}
	args := spec.GetDockerExecCmd(globalCtx.containerCfg, g.spec.ContainerCfg, g.ContainerName, true, mountPoints...)

	if g.spec.Options == nil {
		g.spec.Options = make(map[string]string)
	}

	if g.spec.DisableLogin {
		args = append(args, "-e GF_AUTH_ANONYMOUS_ORG_ROLE=Admin",
			"-e GF_AUTH_ANONYMOUS_ENABLED=true", "-e GF_AUTH_DISABLE_LOGIN_FORM=true")
	}

	for k, v := range g.spec.Options {
		args = append(args, fmt.Sprintf("-e %s=%s", k, v))
	}
	args = append(args, "--user $(id -u)", g.spec.Image)
	return &executor.ExecuteCtx{Target: g.spec.Host, Cmd: strings.Join(args, " ")}
}

func (g *Grafana) Stop(globalCtx *GlobalCtx) *executor.ExecuteCtx {
	args := []string{"docker rm -f", g.ContainerName}
	return &executor.ExecuteCtx{Target: g.spec.Host, Cmd: strings.Join(args, " ")}
}

func (g *Grafana) Remove(globalCtx *GlobalCtx) *executor.ExecuteCtx {
	args := []string{"docker rm -f", g.ContainerName}
	args = append(args, "&&", "rm -rf", g.spec.DataDir, g.spec.RemoteCfgPath)
	return &executor.ExecuteCtx{Target: g.spec.Host, Cmd: strings.Join(args, " ")}
}

func (g *Grafana) SyncConfig(globalCtx *GlobalCtx) *executor.TransferCtx {
	position := utils.ScpDir("template/grafana", g.spec.RemoteCfgPath)

	return &executor.TransferCtx{
		Target: g.spec.Host, Position: position,
	}
}

func (g *Grafana) getDirs() (string, string) {
	return g.spec.RemoteCfgPath, g.spec.DataDir
}

// ================================================================================
// 	AlertManager

type AlertManager struct {
	spec          spec.AlertManagerSpec
	ContainerName string
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
	args := append([]string{}, "mkdir -p", cfgDir, dataDir, "-m 0775")
	args = append(args, fmt.Sprintf("&& chown -R %[1]s:$(id -gn %[1]s) %[2]s %[3]s", globalCtx.User, cfgDir, dataDir))
	return &executor.ExecuteCtx{Target: a.spec.Host, Cmd: strings.Join(args, " ")}
}

func (a *AlertManager) Deploy(globalCtx *GlobalCtx) *executor.ExecuteCtx {
	mountPoints := []spec.MountPoints{
		{a.spec.RemoteCfgPath, "/etc/alertmanager"},
	}
	args := spec.GetDockerExecCmd(globalCtx.containerCfg, a.spec.ContainerCfg, a.ContainerName, true, mountPoints...)
	args = append(args, a.spec.Image, "--config.file=/etc/alertmanager/alertmanager.yml")
	if len(a.spec.AuthUser) != 0 && len(a.spec.AuthPassword) != 0 {
		args = append(args, "--web.config.file=/etc/alertmanager/web.yaml")
	}
	return &executor.ExecuteCtx{Target: a.spec.Host, Cmd: strings.Join(args, " ")}
}

func (a *AlertManager) Stop(globalCtx *GlobalCtx) *executor.ExecuteCtx {
	args := []string{"docker rm -f", a.ContainerName}
	return &executor.ExecuteCtx{Target: a.spec.Host, Cmd: strings.Join(args, " ")}
}

func (a *AlertManager) Remove(globalCtx *GlobalCtx) *executor.ExecuteCtx {
	args := []string{"docker rm -f", a.ContainerName}
	args = append(args, "&&", "rm -rf", a.spec.DataDir, a.spec.RemoteCfgPath)
	return &executor.ExecuteCtx{Target: a.spec.Host, Cmd: strings.Join(args, " ")}
}

func (a *AlertManager) SyncConfig(globalCtx *GlobalCtx) *executor.TransferCtx {
	if len(a.spec.AuthUser) != 0 && len(a.spec.AuthPassword) != 0 {
		basicAuth := promBasicAuthConfig{Username: a.spec.AuthUser, Password: a.spec.AuthPassword}
		_, err := basicAuth.GenConfig("alertManager")
		if err != nil {
			log.Errorf("gen alertManager auth config error: %s", err.Error())
			os.Exit(1)
		}
	}
	position := utils.ScpDir("template/alertmanager", a.spec.RemoteCfgPath)

	return &executor.TransferCtx{
		Target: a.spec.Host, Position: position,
	}
}

func (a *AlertManager) getDirs() (string, string) {
	return a.spec.RemoteCfgPath, a.spec.DataDir
}

// ================================================================================
// 	HStreamExporter

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
	args := append([]string{}, "mkdir -p", cfgDir, dataDir, "-m 0775")
	args = append(args, fmt.Sprintf("&& chown -R %[1]s:$(id -gn %[1]s) %[2]s %[3]s", globalCtx.User, cfgDir, dataDir))
	return &executor.ExecuteCtx{Target: h.spec.Host, Cmd: strings.Join(args, " ")}
}

func (h *HStreamExporter) Deploy(globalCtx *GlobalCtx) *executor.ExecuteCtx {
	args := spec.GetDockerExecCmd(globalCtx.containerCfg, h.spec.ContainerCfg, h.ContainerName, true)
	args = append(args, h.spec.Image)
	args = append(args, "hstream-exporter", "--addr", "hstream://"+h.spec.ServerAddress)
	args = append(args, fmt.Sprintf("--listen-addr 0.0.0.0:%d", h.spec.Port))
	args = append(args, fmt.Sprintf("--log-level %s", h.spec.LogLevel))
	return &executor.ExecuteCtx{Target: h.spec.Host, Cmd: strings.Join(args, " ")}
}

func (h *HStreamExporter) Stop(globalCtx *GlobalCtx) *executor.ExecuteCtx {
	args := []string{"docker rm -f", h.ContainerName}
	return &executor.ExecuteCtx{Target: h.spec.Host, Cmd: strings.Join(args, " ")}
}

func (h *HStreamExporter) Remove(globalCtx *GlobalCtx) *executor.ExecuteCtx {
	args := []string{"docker rm -f", h.ContainerName}
	args = append(args, "&&", "rm -rf", h.spec.DataDir, h.spec.RemoteCfgPath)
	return &executor.ExecuteCtx{Target: h.spec.Host, Cmd: strings.Join(args, " ")}
}

func (h *HStreamExporter) SyncConfig(globalCtx *GlobalCtx) *executor.TransferCtx {
	return nil
}

func (h *HStreamExporter) getDirs() (string, string) {
	return h.spec.RemoteCfgPath, h.spec.DataDir
}

// ================================================================================
// 	Utils

type promBasicAuthConfig struct {
	Username string
	Password string
}

const BasicAuthTemplate = `
basic_auth_users:
  %s: "%s"
`

const (
	AlertManagerAuthCfgPath = "template/alertmanager/web.yaml"
)

func (p *promBasicAuthConfig) GenConfig(component string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(p.Password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}

	authCfg := fmt.Sprintf(BasicAuthTemplate, p.Username, string(hashedPassword))
	content, err := yaml.Marshal(authCfg)
	if err != nil {
		return "", err
	}

	switch component {
	case "alertManager":
		return AlertManagerAuthCfgPath, os.WriteFile(AlertManagerAuthCfgPath, content, 0664)
	default:
		return "", fmt.Errorf("un-supported component: %s", component)
	}
}
