package spec

const (
	NodeExporterDefaultImage         = "prom/node-exporter"
	NodeExporterDefaultContainerName = "deploy_node_exporter"
	CadvisorDefaultImage             = "gcr.io/cadvisor/cadvisor:v0.39.3"
	CadvisorDefaultContainerName     = "deploy_cadvisor"
	MonitorDefaultCfgDir             = "/hstream/deploy/monitor"
	MonitorDefaultDataDir            = "/hstream/data/monitor"

	BlackBoxDefaultContainerName = "deploy_blackbox"
	BlackBoxDefaultImage         = "prom/blackbox-exporter"
	BlackBoxDefaultCfgDir        = "/hstream/deploy/blackbox"
	BlackBoxDefaultDataDir       = "/hstream/data/blackbox"

	PrometheusDefaultContainerName = "deploy_prom"
	PrometheusDefaultImage         = "prom/prometheus"
	PrometheusDefaultCfgDir        = "/hstream/deploy/prometheus"
	PrometheusDefaultDataDir       = "/hstream/data/prometheus"

	GrafanaDefaultContainerName = "deploy_grafana"
	GrafanaDefaultImage         = "grafana/grafana-oss:main"
	GrafanaDefaultCfgDir        = "/hstream/deploy/grafana"
	GrafanaDefaultDataDir       = "/hstream/data/grafana"

	AlertManagerDefaultContainerName = "deploy_alert_manager"
	AlertManagerDefaultImage         = "prom/alertmanager"
	AlertManagerDefaultCfgDir        = "/hstream/deploy/alertmanager"
	AlertManagerDefaultDataDir       = "/hstream/data/alertmanager"

	HStreamExporterDefaultContainerName = "deploy_hstream_exporter"
	HStreamExporterDefaultImage         = "hstreamdb/hstream-exporter"
	HStreamExporterDefaultCfgDir        = "/hstream/deploy/hstream-exporter"
	HStreamExporterDefaultDataDir       = "/hstream/data/hstream-exporter"
)

// ================================================================================
// 	monitor spec

type MonitorSpec struct {
	NodeExporterImage string       `yaml:"node_exporter_image"`
	NodeExporterPort  int          `yaml:"node_exporter_port" default:"9100"`
	CadvisorImage     string       `yaml:"cadvisor_image"`
	CadvisorPort      int          `yaml:"cadvisor_port" default:"7000"`
	ExcludedHosts     []string     `yaml:"excluded_hosts"`
	RemoteCfgPath     string       `yaml:"remote_config_path"`
	DataDir           string       `yaml:"data_dir"`
	ContainerCfg      ContainerCfg `yaml:"container_config"`
}

func (m *MonitorSpec) SetDefaultDataDir() {
	m.DataDir = MonitorDefaultDataDir
}

func (m *MonitorSpec) SetDefaultImage() {
	m.NodeExporterImage = NodeExporterDefaultImage
	m.CadvisorImage = CadvisorDefaultImage
}

func (m *MonitorSpec) SetDefaultRemoteCfgPath() {
	m.RemoteCfgPath = MonitorDefaultCfgDir
}

// ================================================================================
// 	blackbox

type BlackBoxSpec struct {
	Host          string       `yaml:"host"`
	SSHPort       int          `yaml:"ssh_port" default:"22"`
	Port          int          `yaml:"port" default:"9115"`
	Image         string       `yaml:"image"`
	DataDir       string       `yaml:"data_dir"`
	RemoteCfgPath string       `yaml:"remote_config_path"`
	ContainerCfg  ContainerCfg `yaml:"container_config"`
}

func (b *BlackBoxSpec) SetDefaultDataDir() {
	b.DataDir = BlackBoxDefaultDataDir
}

func (b *BlackBoxSpec) SetDefaultImage() {
	b.Image = BlackBoxDefaultImage
}

func (b *BlackBoxSpec) SetDefaultRemoteCfgPath() {
	b.RemoteCfgPath = BlackBoxDefaultCfgDir
}

// ================================================================================
// 	prometheus

type PrometheusSpec struct {
	Host          string       `yaml:"host"`
	SSHPort       int          `yaml:"ssh_port" default:"22"`
	Port          int          `yaml:"port" default:"9090"`
	Image         string       `yaml:"image"`
	DataDir       string       `yaml:"data_dir"`
	RetentionTime string       `yaml:"retention_time" default:"15d"`
	RemoteCfgPath string       `yaml:"remote_config_path"`
	ContainerCfg  ContainerCfg `yaml:"container_config"`
}

func (p *PrometheusSpec) SetDefaultDataDir() {
	p.DataDir = PrometheusDefaultDataDir
}

func (p *PrometheusSpec) SetDefaultImage() {
	p.Image = PrometheusDefaultImage
}

func (p *PrometheusSpec) SetDefaultRemoteCfgPath() {
	p.RemoteCfgPath = PrometheusDefaultCfgDir
}

// ================================================================================
// 	grafana

type GrafanaSpec struct {
	Host          string            `yaml:"host"`
	SSHPort       int               `yaml:"ssh_port" default:"22"`
	Port          int               `yaml:"port" default:"3000"`
	Image         string            `yaml:"image"`
	DataDir       string            `yaml:"data_dir"`
	DisableLogin  bool              `yaml:"disable_login"`
	RemoteCfgPath string            `yaml:"remote_config_path"`
	Options       map[string]string `yaml:"option"`
	ContainerCfg  ContainerCfg      `yaml:"container_config"`
}

func (g *GrafanaSpec) SetDefaultDataDir() {
	g.DataDir = GrafanaDefaultDataDir
}

func (g *GrafanaSpec) SetDefaultImage() {
	g.Image = GrafanaDefaultImage
}

func (g *GrafanaSpec) SetDefaultRemoteCfgPath() {
	g.RemoteCfgPath = GrafanaDefaultCfgDir
}

// ================================================================================
// 	alert-manager

type AlertManagerSpec struct {
	Host          string       `yaml:"host"`
	SSHPort       int          `yaml:"ssh_port" default:"22"`
	Port          int          `yaml:"port" default:"9093"`
	Image         string       `yaml:"image"`
	DataDir       string       `yaml:"data_dir"`
	RemoteCfgPath string       `yaml:"remote_config_path"`
	AuthUser      string       `yaml:"auth_user"`
	AuthPassword  string       `yaml:"auth_password"`
	ContainerCfg  ContainerCfg `yaml:"container_config"`
}

func (a *AlertManagerSpec) SetDefaultDataDir() {
	a.DataDir = AlertManagerDefaultDataDir
}

func (a *AlertManagerSpec) SetDefaultImage() {
	a.Image = AlertManagerDefaultImage
}

func (a *AlertManagerSpec) SetDefaultRemoteCfgPath() {
	a.RemoteCfgPath = AlertManagerDefaultCfgDir
}

// ================================================================================
// 	hstream-exporter

type HStreamExporterSpec struct {
	Host          string       `yaml:"host"`
	SSHPort       int          `yaml:"ssh_port" default:"22"`
	Port          int          `yaml:"port" default:"9250"`
	Image         string       `yaml:"image"`
	LogLevel      string       `yaml:"log_level" default:"info"`
	DataDir       string       `yaml:"data_dir"`
	RemoteCfgPath string       `yaml:"remote_config_path"`
	ServerAddress string       `yaml:"server_address"`
	ContainerCfg  ContainerCfg `yaml:"container_config"`
}

func (g *HStreamExporterSpec) SetDefaultDataDir() {
	g.DataDir = HStreamExporterDefaultDataDir
}

func (g *HStreamExporterSpec) SetDefaultImage() {
	g.Image = HStreamExporterDefaultImage
}

func (g *HStreamExporterSpec) SetDefaultRemoteCfgPath() {
	g.RemoteCfgPath = HStreamExporterDefaultCfgDir
}
