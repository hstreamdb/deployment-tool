package task

import (
	"fmt"
	ext "github.com/hstreamdb/dev-deploy/pkg/executor"
	"github.com/hstreamdb/dev-deploy/pkg/service"
	"sync"
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
	wg := sync.WaitGroup{}
	mutex := sync.Mutex{}
	var firstErr error
	wg.Add(len(s.service))
	for _, svc := range s.service {
		go func(svc *service.MonitorSuite) {
			defer wg.Done()
			executorCtx := svc.InitEnv(s.ctx)
			target := fmt.Sprintf("%s:%d", executorCtx.Target, s.ctx.SSHPort)
			res, err := executor.Execute(target, executorCtx.Cmd)
			if err != nil {
				mutex.Lock()
				if firstErr == nil {
					firstErr = fmt.Errorf("%s-%s", err.Error(), res)
				}
				mutex.Unlock()
			}
		}(svc)
	}
	wg.Wait()
	return firstErr
}

type SyncMonitorSuiteConfig struct {
	MonitorSuiteCtx
}

func (s *SyncMonitorSuiteConfig) String() string {
	return "Task: sync monitor suite config"
}

func (s *SyncMonitorSuiteConfig) Run(executor ext.Executor) error {
	wg := sync.WaitGroup{}
	mutex := sync.Mutex{}
	var firstErr error
	wg.Add(len(s.service))
	for _, svc := range s.service {
		go func(svc *service.MonitorSuite) {
			defer wg.Done()
			transferCtx := svc.SyncConfig(s.ctx)
			if transferCtx == nil {
				fmt.Printf("skip %s\n", s)
				return
			}
			target := fmt.Sprintf("%s:%d", transferCtx.Target, s.ctx.SSHPort)
			for _, position := range transferCtx.Position {
				if err := executor.Transfer(target, position.LocalDir, position.RemoteDir); err != nil {
					mutex.Lock()
					if firstErr == nil {
						firstErr = err
					}
					mutex.Unlock()
					break
				}

				if len(position.Opts) != 0 {
					executor.Execute(target, position.Opts)
				}
			}

		}(svc)
	}
	wg.Wait()
	return firstErr
}

type StartMonitorSuite struct {
	MonitorSuiteCtx
}

func (s *StartMonitorSuite) String() string {
	return "Task: start monitor suites"
}

func (s *StartMonitorSuite) Run(executor ext.Executor) error {
	wg := sync.WaitGroup{}
	mutex := sync.Mutex{}
	var firstErr error
	wg.Add(len(s.service))
	for _, svc := range s.service {
		go func(svc *service.MonitorSuite) {
			defer wg.Done()
			executorCtx := svc.Deploy(s.ctx)
			target := fmt.Sprintf("%s:%d", executorCtx.Target, s.ctx.SSHPort)
			res, err := executor.Execute(target, executorCtx.Cmd)
			if err != nil {
				mutex.Lock()
				if firstErr == nil {
					firstErr = fmt.Errorf("%s-%s", err.Error(), res)
				}
				mutex.Unlock()
			}
		}(svc)
	}
	wg.Wait()
	return firstErr
}

type RemoveMonitorSuite struct {
	MonitorSuiteCtx
}

func (r *RemoveMonitorSuite) String() string {
	return "Task: remove monitor suites"
}

func (r *RemoveMonitorSuite) Run(executor ext.Executor) error {
	wg := sync.WaitGroup{}
	mutex := sync.Mutex{}
	var firstErr error
	wg.Add(len(r.service))
	for _, svc := range r.service {
		go func(svc *service.MonitorSuite) {
			defer wg.Done()
			executorCtx := svc.Remove(r.ctx)
			target := fmt.Sprintf("%s:%d", executorCtx.Target, r.ctx.SSHPort)
			res, err := executor.Execute(target, executorCtx.Cmd)
			if err != nil {
				mutex.Lock()
				if firstErr == nil {
					firstErr = fmt.Errorf("%s-%s", err.Error(), res)
				}
				mutex.Unlock()
			}
		}(svc)
	}
	wg.Wait()
	return firstErr
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
	wg := sync.WaitGroup{}
	mutex := sync.Mutex{}
	var firstErr error
	wg.Add(len(p.service))
	for _, svc := range p.service {
		go func(svc *service.Prometheus) {
			defer wg.Done()
			executorCtx := svc.InitEnv(p.ctx)
			target := fmt.Sprintf("%s:%d", executorCtx.Target, p.ctx.SSHPort)
			res, err := executor.Execute(target, executorCtx.Cmd)
			if err != nil {
				mutex.Lock()
				if firstErr == nil {
					firstErr = fmt.Errorf("%s-%s", err.Error(), res)
				}
				mutex.Unlock()
			}
		}(svc)
	}
	wg.Wait()
	return firstErr
}

type SyncPrometheusConfig struct {
	PrometheusCtx
}

func (s *SyncPrometheusConfig) String() string {
	return "Task: sync prometheus config"
}

func (s *SyncPrometheusConfig) Run(executor ext.Executor) error {
	wg := sync.WaitGroup{}
	mutex := sync.Mutex{}
	var firstErr error
	wg.Add(len(s.service))
	for _, svc := range s.service {
		go func(svc *service.Prometheus) {
			defer wg.Done()
			transferCtx := svc.SyncConfig(s.ctx)
			if transferCtx == nil {
				fmt.Printf("skip %s\n", s)
				return
			}
			target := fmt.Sprintf("%s:%d", transferCtx.Target, s.ctx.SSHPort)
			for _, position := range transferCtx.Position {
				if err := executor.Transfer(target, position.LocalDir, position.RemoteDir); err != nil {
					mutex.Lock()
					if firstErr == nil {
						firstErr = err
					}
					mutex.Unlock()
					break
				}

				if len(position.Opts) != 0 {
					executor.Execute(target, position.Opts)
				}
			}

		}(svc)
	}
	wg.Wait()
	return firstErr
}

type StartPrometheus struct {
	PrometheusCtx
}

func (s *StartPrometheus) String() string {
	return "Task: start prometheus"
}

func (s *StartPrometheus) Run(executor ext.Executor) error {
	wg := sync.WaitGroup{}
	mutex := sync.Mutex{}
	var firstErr error
	wg.Add(len(s.service))
	for _, svc := range s.service {
		go func(svc *service.Prometheus) {
			defer wg.Done()
			executorCtx := svc.Deploy(s.ctx)
			target := fmt.Sprintf("%s:%d", executorCtx.Target, s.ctx.SSHPort)
			res, err := executor.Execute(target, executorCtx.Cmd)
			if err != nil {
				mutex.Lock()
				if firstErr == nil {
					firstErr = fmt.Errorf("%s-%s", err.Error(), res)
				}
				mutex.Unlock()
			}
		}(svc)
	}
	wg.Wait()
	return firstErr
}

type RemovePrometheus struct {
	PrometheusCtx
}

func (r *RemovePrometheus) String() string {
	return "Task: remove prometheus"
}

func (r *RemovePrometheus) Run(executor ext.Executor) error {
	wg := sync.WaitGroup{}
	mutex := sync.Mutex{}
	var firstErr error
	wg.Add(len(r.service))
	for _, svc := range r.service {
		go func(svc *service.Prometheus) {
			defer wg.Done()
			executorCtx := svc.Remove(r.ctx)
			target := fmt.Sprintf("%s:%d", executorCtx.Target, r.ctx.SSHPort)
			res, err := executor.Execute(target, executorCtx.Cmd)
			if err != nil {
				mutex.Lock()
				if firstErr == nil {
					firstErr = fmt.Errorf("%s-%s", err.Error(), res)
				}
				mutex.Unlock()
			}
		}(svc)
	}
	wg.Wait()
	return firstErr
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
	wg := sync.WaitGroup{}
	mutex := sync.Mutex{}
	var firstErr error
	wg.Add(len(p.service))
	for _, svc := range p.service {
		go func(svc *service.Grafana) {
			defer wg.Done()
			executorCtx := svc.InitEnv(p.ctx)
			target := fmt.Sprintf("%s:%d", executorCtx.Target, p.ctx.SSHPort)
			res, err := executor.Execute(target, executorCtx.Cmd)
			if err != nil {
				mutex.Lock()
				if firstErr == nil {
					firstErr = fmt.Errorf("%s-%s", err.Error(), res)
				}
				mutex.Unlock()
			}
		}(svc)
	}
	wg.Wait()
	return firstErr
}

type SyncGrafanaConfig struct {
	GrafanaCtx
}

func (s *SyncGrafanaConfig) String() string {
	return "Task: sync grafana config"
}

func (s *SyncGrafanaConfig) Run(executor ext.Executor) error {
	wg := sync.WaitGroup{}
	mutex := sync.Mutex{}
	var firstErr error
	wg.Add(len(s.service))
	for _, svc := range s.service {
		go func(svc *service.Grafana) {
			defer wg.Done()
			transferCtx := svc.SyncConfig(s.ctx)
			if transferCtx == nil {
				fmt.Printf("skip %s\n", s)
				return
			}
			target := fmt.Sprintf("%s:%d", transferCtx.Target, s.ctx.SSHPort)
			for _, position := range transferCtx.Position {
				if err := executor.Transfer(target, position.LocalDir, position.RemoteDir); err != nil {
					mutex.Lock()
					if firstErr == nil {
						firstErr = err
					}
					mutex.Unlock()
					break
				}

				if len(position.Opts) != 0 {
					executor.Execute(target, position.Opts)
				}
			}

		}(svc)
	}
	wg.Wait()
	return firstErr
}

type StartGrafana struct {
	GrafanaCtx
}

func (s *StartGrafana) String() string {
	return "Task: start grafana"
}

func (s *StartGrafana) Run(executor ext.Executor) error {
	wg := sync.WaitGroup{}
	mutex := sync.Mutex{}
	var firstErr error
	wg.Add(len(s.service))
	for _, svc := range s.service {
		go func(svc *service.Grafana) {
			defer wg.Done()
			executorCtx := svc.Deploy(s.ctx)
			target := fmt.Sprintf("%s:%d", executorCtx.Target, s.ctx.SSHPort)
			res, err := executor.Execute(target, executorCtx.Cmd)
			if err != nil {
				mutex.Lock()
				if firstErr == nil {
					firstErr = fmt.Errorf("%s-%s", err.Error(), res)
				}
				mutex.Unlock()
			}
		}(svc)
	}
	wg.Wait()
	return firstErr
}

type RemoveGrafana struct {
	GrafanaCtx
}

func (r *RemoveGrafana) String() string {
	return "Task: remove grafana"
}

func (r *RemoveGrafana) Run(executor ext.Executor) error {
	wg := sync.WaitGroup{}
	mutex := sync.Mutex{}
	var firstErr error
	wg.Add(len(r.service))
	for _, svc := range r.service {
		go func(svc *service.Grafana) {
			defer wg.Done()
			executorCtx := svc.Remove(r.ctx)
			target := fmt.Sprintf("%s:%d", executorCtx.Target, r.ctx.SSHPort)
			res, err := executor.Execute(target, executorCtx.Cmd)
			if err != nil {
				mutex.Lock()
				if firstErr == nil {
					firstErr = fmt.Errorf("%s-%s", err.Error(), res)
				}
				mutex.Unlock()
			}
		}(svc)
	}
	wg.Wait()
	return firstErr
}
