package task

import (
	ext "github.com/hstreamdb/deployment-tool/pkg/executor"
	"github.com/hstreamdb/deployment-tool/pkg/service"
)

type MonitorSuiteCtx struct {
	ctx     *service.GlobalCtx
	service []*service.MonitorSuite
}

type InitMonitorSuiteEnv struct {
	MonitorSuiteCtx
}

func (s *InitMonitorSuiteEnv) String() string {
	return "Task: init monitor suite environment"
}

func (s *InitMonitorSuiteEnv) Run(executor ext.Executor) error {
	return serviceInitEnv(executor, s.ctx, s.service)
}

type SyncMonitorSuiteConfig struct {
	MonitorSuiteCtx
}

func (s *SyncMonitorSuiteConfig) String() string {
	return "Task: sync monitor suite config"
}

func (s *SyncMonitorSuiteConfig) Run(executor ext.Executor) error {
	return configSync(executor, s.ctx, s.service)
}

type StartMonitorSuite struct {
	MonitorSuiteCtx
}

func (s *StartMonitorSuite) String() string {
	return "Task: start monitor suites"
}

func (s *StartMonitorSuite) Run(executor ext.Executor) error {
	return serviceDeploy(executor, s.ctx, s.service)
}

type RemoveMonitorSuite struct {
	MonitorSuiteCtx
}

func (r *RemoveMonitorSuite) String() string {
	return "Task: remove monitor suites"
}

func (r *RemoveMonitorSuite) Run(executor ext.Executor) error {
	return serviceRemove(executor, r.ctx, r.service)
}

// ================================================================================

type PrometheusCtx struct {
	ctx     *service.GlobalCtx
	service []*service.Prometheus
}

type InitPrometheus struct {
	PrometheusCtx
}

func (p *InitPrometheus) String() string {
	return "Task: init prometheus environment"
}

func (p *InitPrometheus) Run(executor ext.Executor) error {
	return serviceInitEnv(executor, p.ctx, p.service)
}

type SyncPrometheusConfig struct {
	PrometheusCtx
}

func (s *SyncPrometheusConfig) String() string {
	return "Task: sync prometheus config"
}

func (s *SyncPrometheusConfig) Run(executor ext.Executor) error {
	return configSync(executor, s.ctx, s.service)
}

type StartPrometheus struct {
	PrometheusCtx
}

func (s *StartPrometheus) String() string {
	return "Task: start prometheus"
}

func (s *StartPrometheus) Run(executor ext.Executor) error {
	return serviceDeploy(executor, s.ctx, s.service)
}

type RemovePrometheus struct {
	PrometheusCtx
}

func (r *RemovePrometheus) String() string {
	return "Task: remove prometheus"
}

func (r *RemovePrometheus) Run(executor ext.Executor) error {
	return serviceRemove(executor, r.ctx, r.service)
}

// ================================================================================

type GrafanaCtx struct {
	ctx     *service.GlobalCtx
	service []*service.Grafana
}

type InitGrafana struct {
	GrafanaCtx
}

func (p *InitGrafana) String() string {
	return "Task: init grafana environment"
}

func (p *InitGrafana) Run(executor ext.Executor) error {
	return serviceInitEnv(executor, p.ctx, p.service)
}

type SyncGrafanaConfig struct {
	GrafanaCtx
}

func (s *SyncGrafanaConfig) String() string {
	return "Task: sync grafana config"
}

func (s *SyncGrafanaConfig) Run(executor ext.Executor) error {
	return configSync(executor, s.ctx, s.service)
}

type StartGrafana struct {
	GrafanaCtx
}

func (s *StartGrafana) String() string {
	return "Task: start grafana"
}

func (s *StartGrafana) Run(executor ext.Executor) error {
	return serviceDeploy(executor, s.ctx, s.service)
}

type RemoveGrafana struct {
	GrafanaCtx
}

func (r *RemoveGrafana) String() string {
	return "Task: remove grafana"
}

func (r *RemoveGrafana) Run(executor ext.Executor) error {
	return serviceRemove(executor, r.ctx, r.service)
}

// ================================================================================

type HStreamExporterCtx struct {
	ctx     *service.GlobalCtx
	service []*service.HStreamExporter
}

type InitHStreamExporter struct {
	HStreamExporterCtx
}

func (p *InitHStreamExporter) String() string {
	return "Task: init hstream-exporter environment"
}

func (p *InitHStreamExporter) Run(executor ext.Executor) error {
	return serviceInitEnv(executor, p.ctx, p.service)
}

type SyncHStreamExporterConfig struct {
	HStreamExporterCtx
}

func (s *SyncHStreamExporterConfig) String() string {
	return "Task: sync hstream-exporter config"
}

func (s *SyncHStreamExporterConfig) Run(executor ext.Executor) error {
	return configSync(executor, s.ctx, s.service)
}

type StartHStreamExporter struct {
	HStreamExporterCtx
}

func (s *StartHStreamExporter) String() string {
	return "Task: start hstream-exporter"
}

func (s *StartHStreamExporter) Run(executor ext.Executor) error {
	return serviceDeploy(executor, s.ctx, s.service)
}

type RemoveHStreamExporter struct {
	HStreamExporterCtx
}

func (r *RemoveHStreamExporter) String() string {
	return "Task: remove hstream-exporter"
}

func (r *RemoveHStreamExporter) Run(executor ext.Executor) error {
	return serviceRemove(executor, r.ctx, r.service)
}
