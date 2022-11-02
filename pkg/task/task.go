package task

import (
	"fmt"
	ext "github.com/hstreamdb/deployment-tool/pkg/executor"
	"github.com/hstreamdb/deployment-tool/pkg/service"
	"sync"
	"sync/atomic"
)

type Task interface {
	Run(executor ext.Executor) error
}

type basicExecuteTask uint8

const (
	InitEnv basicExecuteTask = iota + 1
	Deploy
	Remove
)

func serviceInitEnv[S service.Service](executor ext.Executor, ctx *service.GlobalCtx, services []S) error {
	return parallelRun(executor, ctx, services, InitEnv)
}

func serviceDeploy[S service.Service](executor ext.Executor, ctx *service.GlobalCtx, services []S) error {
	return parallelRun(executor, ctx, services, Deploy)
}

func serviceRemove[S service.Service](executor ext.Executor, ctx *service.GlobalCtx, services []S) error {
	return parallelRun(executor, ctx, services, Remove)
}

func parallelRun[S service.Service](executor ext.Executor, ctx *service.GlobalCtx, services []S, tp basicExecuteTask) error {
	wg := sync.WaitGroup{}
	var firstErr error
	ep := atomic.Pointer[error]{}
	ep.Store(&firstErr)
	wg.Add(len(services))
	for _, svc := range services {
		go func(svc S) {
			defer wg.Done()
			var executorCtx *ext.ExecuteCtx
			switch tp {
			case InitEnv:
				executorCtx = svc.InitEnv(ctx)
			case Deploy:
				executorCtx = svc.Deploy(ctx)
			case Remove:
				executorCtx = svc.Remove(ctx)
			}
			target := fmt.Sprintf("%s:%d", executorCtx.Target, ctx.SSHPort)
			res, err := executor.Execute(target, executorCtx.Cmd)
			if err != nil && *ep.Load() == nil {
				e := fmt.Errorf("%s-%s", err.Error(), res)
				ep.Store(&e)
			}
		}(svc)
	}
	wg.Wait()
	return *ep.Load()
}

func configSync[S service.Service](executor ext.Executor, ctx *service.GlobalCtx, services []S) error {
	wg := sync.WaitGroup{}
	var firstErr error
	ep := atomic.Pointer[error]{}
	ep.Store(&firstErr)
	wg.Add(len(services))
	for _, svc := range services {
		go func(svc service.Service) {
			defer wg.Done()
			transferCtx := svc.SyncConfig(ctx)
			if transferCtx == nil {
				fmt.Printf("skip sync config for %s\n", svc.GetServiceName())
				return
			}
			target := fmt.Sprintf("%s:%d", transferCtx.Target, ctx.SSHPort)
			for _, position := range transferCtx.Position {
				if err := executor.Transfer(target, position.LocalDir, position.RemoteDir); err != nil {
					if *ep.Load() == nil {
						ep.Store(&err)
					}
					break
				}

				if len(position.Opts) != 0 {
					if _, err := executor.Execute(target, position.Opts); err != nil {
						if *ep.Load() == nil {
							ep.Store(&err)
						}
						break
					}
				}
			}

		}(svc)
	}
	wg.Wait()
	return *ep.Load()
}

type initEnvTask[S service.Service] struct {
	serviceName string
	ctx         *service.GlobalCtx
	services    []S
}

func (i *initEnvTask[S]) String() string {
	return fmt.Sprintf("Task: init %s environment\n", i.serviceName)
}

func (i *initEnvTask[S]) Run(executor ext.Executor) error {
	return serviceInitEnv(executor, i.ctx, i.services)
}

type configSyncTask[S service.Service] struct {
	serviceName string
	ctx         *service.GlobalCtx
	services    []S
}

func (c *configSyncTask[S]) String() string {
	return fmt.Sprintf("Task: sync %s config\n", c.serviceName)
}

func (c *configSyncTask[S]) Run(executor ext.Executor) error {
	return configSync(executor, c.ctx, c.services)
}

type serviceDeployTask[S service.Service] struct {
	serviceName string
	ctx         *service.GlobalCtx
	services    []S
}

func (s *serviceDeployTask[S]) String() string {
	return fmt.Sprintf("Task: start %s cluster\n", s.serviceName)
}

func (s *serviceDeployTask[S]) Run(executor ext.Executor) error {
	return serviceDeploy(executor, s.ctx, s.services)
}

type removeServiceTask[S service.Service] struct {
	serviceName string
	ctx         *service.GlobalCtx
	services    []S
}

func (r *removeServiceTask[S]) String() string {
	return fmt.Sprintf("Task: remove %s\n", r.serviceName)
}

func (r *removeServiceTask[S]) Run(executor ext.Executor) error {
	return serviceRemove(executor, r.ctx, r.services)
}

func getStartServiceTask[S service.Service](ctx *service.GlobalCtx, services []S) []Task {
	serviceName := services[0].GetServiceName()
	tasks := make([]Task, 0, 3)
	tasks = append(tasks, &initEnvTask[S]{serviceName: serviceName, ctx: ctx, services: services})
	tasks = append(tasks, &configSyncTask[S]{serviceName: serviceName, ctx: ctx, services: services})
	tasks = append(tasks, &serviceDeployTask[S]{serviceName: serviceName, ctx: ctx, services: services})
	return tasks
}

func getRemoveServiceTask[S service.Service](ctx *service.GlobalCtx, services []S) []Task {
	serviceName := services[0].GetServiceName()
	return []Task{&removeServiceTask[S]{serviceName: serviceName, ctx: ctx, services: services}}
}
