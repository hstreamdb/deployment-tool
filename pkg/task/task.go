package task

import (
	"fmt"
	ext "github.com/hstreamdb/dev-deploy/pkg/executor"
	"github.com/hstreamdb/dev-deploy/pkg/service"
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
				fmt.Printf("skip %s\n", getServiceName(svc))
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
					executor.Execute(target, position.Opts)
				}
			}

		}(svc)
	}
	wg.Wait()
	return *ep.Load()
}

func getServiceName(svc service.Service) string {
	switch svc.(type) {
	case *service.HServer:
		return "HServer"
	case *service.HStore:
		return "HStore"
	case *service.MetaStore:
		return "MetaStore"
	case *service.Prometheus:
		return "Prometheus"
	case *service.Grafana:
		return "Grafana"
	}
	return ""
}
