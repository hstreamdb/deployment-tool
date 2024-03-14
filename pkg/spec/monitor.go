package spec

import "path"

const (
	NodeExporterDefaultImage         = "prom/node-exporter"
	NodeExporterDefaultContainerName = "deploy_node_exporter"
	CadvisorDefaultImage             = "gcr.io/cadvisor/cadvisor:v0.39.3"
	CadvisorDefaultContainerName     = "deploy_cadvisor"
	MonitorDefaultCfgDir             = "deploy/monitor"
	MonitorDefaultDataDir            = "data/monitor"

	BlackBoxDefaultContainerName = "deploy_blackbox"
	BlackBoxDefaultImage         = "prom/blackbox-exporter"
	BlackBoxDefaultCfgDir        = "deploy/blackbox"
	BlackBoxDefaultDataDir       = "data/blackbox"

	PrometheusDefaultContainerName = "deploy_prom"
	PrometheusDefaultImage         = "prom/prometheus"
	PrometheusDefaultCfgDir        = "deploy/prometheus"
	PrometheusDefaultDataDir       = "data/prometheus"

	GrafanaDefaultContainerName = "deploy_grafana"
	GrafanaDefaultImage         = "grafana/grafana-oss:main"
	GrafanaDefaultCfgDir        = "deploy/grafana"
	GrafanaDefaultDataDir       = "data/grafana"

	AlertManagerDefaultContainerName = "deploy_alert_manager"
	AlertManagerDefaultImage         = "prom/alertmanager"
	AlertManagerDefaultCfgDir        = "deploy/alertmanager"
	AlertManagerDefaultDataDir       = "data/alertmanager"

	HStreamExporterDefaultContainerName = "deploy_hstream_exporter"
	HStreamExporterDefaultImage         = "hstreamdb/hstream-exporter"
	HStreamExporterDefaultCfgDir        = "deploy/hstream-exporter"
	HStreamExporterDefaultDataDir       = "data/hstream-exporter"
)

// ================================================================================
// 	monitor spec

type MonitorSpec struct {
	NodeExporterImage string       `yaml:"node_exporter_image"`
	NodeExporterPort  int          `yaml:"node_exporter_port" default:"9100"`
	CadvisorImage     string       `yaml:"cadvisor_image"`
	CadvisorPort      int          `yaml:"cadvisor_port" default:"7000"`
	ExcludedHosts     []string     `yaml:"excluded_hosts"`
	ExtendHosts       []string     `yaml:"extend_hosts"`
	ContainerCfg      ContainerCfg `yaml:"container_config"`
	RemoteCfgPath     string
	DataDir           string
}

func (m *MonitorSpec) SetDataDir(prefix string) {
	m.DataDir = path.Join(prefix, MonitorDefaultDataDir)
}

func (m *MonitorSpec) SetDefaultImage() {
	m.NodeExporterImage = NodeExporterDefaultImage
	m.CadvisorImage = CadvisorDefaultImage
}

func (m *MonitorSpec) SetRemoteCfgPath(prefix string) {
	m.RemoteCfgPath = path.Join(prefix, MonitorDefaultCfgDir)
}

// ================================================================================
// 	blackbox

type BlackBoxSpec struct {
	Host          string       `yaml:"host"`
	SSHPort       int          `yaml:"ssh_port" default:"22"`
	Port          int          `yaml:"port" default:"9115"`
	Image         string       `yaml:"image"`
	ContainerCfg  ContainerCfg `yaml:"container_config"`
	DataDir       string
	RemoteCfgPath string
}

func (b *BlackBoxSpec) SetDataDir(prefix string) {
	b.DataDir = path.Join(prefix, BlackBoxDefaultDataDir)
}

func (b *BlackBoxSpec) SetDefaultImage() {
	b.Image = BlackBoxDefaultImage
}

func (b *BlackBoxSpec) SetRemoteCfgPath(prefix string) {
	b.RemoteCfgPath = path.Join(prefix, BlackBoxDefaultCfgDir)
}

// ================================================================================
// 	prometheus

type PrometheusSpec struct {
	Host            string `yaml:"host"`
	SSHPort         int    `yaml:"ssh_port" default:"22"`
	Port            int    `yaml:"port" default:"9090"`
	Image           string `yaml:"image"`
	RetentionTime   string `yaml:"retention_time" default:"15d"`
	BlackBoxConfigs struct {
		Address string `yaml:"address"`
	} `yaml:"blackbox_exporter_configs"`
	HStreamExporterConfigs []struct {
		Address string `yaml:"address"`
	} `yaml:"hstream_exporter_configs"`
	AlertManagerConfigs []struct {
		Address      string `yaml:"address"`
		AuthUser     string `yaml:"auth_user"`
		AuthPassword string `yaml:"auth_password"`
	} `yaml:"alertmanager_configs"`
	ContainerCfg  ContainerCfg `yaml:"container_config"`
	DataDir       string
	RemoteCfgPath string
}

func (p *PrometheusSpec) SetDataDir(prefix string) {
	p.DataDir = path.Join(prefix, PrometheusDefaultDataDir)
}

func (p *PrometheusSpec) SetDefaultImage() {
	p.Image = PrometheusDefaultImage
}

func (p *PrometheusSpec) SetRemoteCfgPath(prefix string) {
	p.RemoteCfgPath = path.Join(prefix, PrometheusDefaultCfgDir)
}

// ================================================================================
// 	grafana

type GrafanaSpec struct {
	Host          string            `yaml:"host"`
	SSHPort       int               `yaml:"ssh_port" default:"22"`
	Port          int               `yaml:"port" default:"3000"`
	Image         string            `yaml:"image"`
	DisableLogin  bool              `yaml:"disable_login"`
	Options       map[string]string `yaml:"option"`
	ContainerCfg  ContainerCfg      `yaml:"container_config"`
	DataDir       string
	RemoteCfgPath string
}

func (g *GrafanaSpec) SetDataDir(prefix string) {
	g.DataDir = path.Join(prefix, GrafanaDefaultDataDir)
}

func (g *GrafanaSpec) SetDefaultImage() {
	g.Image = GrafanaDefaultImage
}

func (g *GrafanaSpec) SetRemoteCfgPath(prefix string) {
	g.RemoteCfgPath = path.Join(prefix, GrafanaDefaultCfgDir)
}

// ================================================================================
// 	alert-manager

type AlertManagerSpec struct {
	Host          string       `yaml:"host"`
	SSHPort       int          `yaml:"ssh_port" default:"22"`
	Port          int          `yaml:"port" default:"9093"`
	Image         string       `yaml:"image"`
	AuthUser      string       `yaml:"auth_user"`
	AuthPassword  string       `yaml:"auth_password"`
	ContainerCfg  ContainerCfg `yaml:"container_config"`
	DataDir       string
	RemoteCfgPath string
}

func (a *AlertManagerSpec) SetDataDir(prefix string) {
	a.DataDir = path.Join(prefix, AlertManagerDefaultDataDir)
}

func (a *AlertManagerSpec) SetDefaultImage() {
	a.Image = AlertManagerDefaultImage
}

func (a *AlertManagerSpec) SetRemoteCfgPath(prefix string) {
	a.RemoteCfgPath = path.Join(prefix, AlertManagerDefaultCfgDir)
}

// ================================================================================
// 	hstream-exporter

type HStreamExporterSpec struct {
	Host          string       `yaml:"host"`
	SSHPort       int          `yaml:"ssh_port" default:"22"`
	Port          int          `yaml:"port" default:"9250"`
	Image         string       `yaml:"image"`
	LogLevel      string       `yaml:"log_level" default:"info"`
	ServerAddress string       `yaml:"server_address"`
	ContainerCfg  ContainerCfg `yaml:"container_config"`
	DataDir       string
	RemoteCfgPath string
}

func (g *HStreamExporterSpec) SetDataDir(prefix string) {
	g.DataDir = path.Join(prefix, HStreamExporterDefaultDataDir)
}

func (g *HStreamExporterSpec) SetDefaultImage() {
	g.Image = HStreamExporterDefaultImage
}

func (g *HStreamExporterSpec) SetRemoteCfgPath(prefix string) {
	g.RemoteCfgPath = path.Join(prefix, HStreamExporterDefaultCfgDir)
}
