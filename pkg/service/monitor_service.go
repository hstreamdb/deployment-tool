package service

import (
	"fmt"
	"github.com/hstreamdb/dev-deploy/pkg/executor"
	"github.com/hstreamdb/dev-deploy/pkg/spec"
	"github.com/hstreamdb/dev-deploy/pkg/template/config"
	"github.com/hstreamdb/dev-deploy/pkg/utils"
	"path/filepath"
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

func (m *MonitorSuite) InitEnv(globalCtx *GlobalCtx) *executor.ExecuteCtx {
	cfgDir, dataDir := m.getDirs()
	args := append([]string{}, "sudo mkdir -p", cfgDir, dataDir)
	return &executor.ExecuteCtx{Target: m.Host, Cmd: strings.Join(args, " ")}
}

func (m *MonitorSuite) Deploy(globalCtx *GlobalCtx) *executor.ExecuteCtx {
	startNodeExporter := spec.GetDockerExecCmd(globalCtx.containerCfg, m.spec.ContainerCfg, m.NodeContainerName, true)
	args := append(startNodeExporter, m.spec.NodeExporterImage, "&&")

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
	spec             spec.PrometheusSpec
	ContainerName    string
	MonitoredHosts   []string
	NodeExporterPort int
	CadvisorPort     int
}

func NewPrometheus(promSpec spec.PrometheusSpec, monitorSuites []*MonitorSuite) *Prometheus {
	hosts := make([]string, 0, len(monitorSuites))
	for _, suite := range monitorSuites {
		hosts = append(hosts, suite.Host)
	}
	return &Prometheus{
		spec:             promSpec,
		ContainerName:    spec.PrometheusDefaultContainerName,
		MonitoredHosts:   hosts,
		NodeExporterPort: monitorSuites[0].spec.NodeExporterPort,
		CadvisorPort:     monitorSuites[0].spec.CadvisorPort,
	}
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
		NodeExporterAddress: nodeAddr,
		CadVisorAddress:     cadAddr,
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