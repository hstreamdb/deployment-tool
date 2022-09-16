package task

import (
	"fmt"
	ext "github.com/hstreamdb/dev-deploy/pkg/executor"
	"github.com/hstreamdb/dev-deploy/pkg/service"
	"sync"
)

type HServerClusterCtx struct {
	ctx     *service.GlobalCtx
	service []*service.HServer
}

type InitHServerEnv struct {
	HServerClusterCtx
}

func (s *InitHServerEnv) String() string {
	return "Task: init server environment"
}

func (s *InitHServerEnv) Run(executor ext.Executor) error {
	wg := sync.WaitGroup{}
	mutex := sync.Mutex{}
	var firstErr error
	wg.Add(len(s.service))
	for _, svc := range s.service {
		go func(svc *service.HServer) {
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

type SyncHServerConfig struct {
	HServerClusterCtx
}

func (s *SyncHServerConfig) String() string {
	return "Task: sync server config"
}

func (s *SyncHServerConfig) Run(executor ext.Executor) error {
	wg := sync.WaitGroup{}
	mutex := sync.Mutex{}
	var firstErr error
	wg.Add(len(s.service))
	for _, svc := range s.service {
		go func(svc *service.HServer) {
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

type StartHServerCluster struct {
	HServerClusterCtx
}

func (s *StartHServerCluster) String() string {
	return "Task: start server cluster"
}

func (s *StartHServerCluster) Run(executor ext.Executor) error {
	wg := sync.WaitGroup{}
	mutex := sync.Mutex{}
	var firstErr error
	wg.Add(len(s.service))
	for _, svc := range s.service {
		go func(svc *service.HServer) {
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

type WaitServerReady struct {
	HServerClusterCtx
}

func (w *WaitServerReady) String() string {
	return "Task: wait server ready"
}

func (w *WaitServerReady) Run(executor ext.Executor) error {
	for _, store := range w.service {
		executorCtx := store.CheckReady(w.ctx)
		target := fmt.Sprintf("%s:%d", executorCtx.Target, w.ctx.SSHPort)
		res, err := executor.Execute(target, executorCtx.Cmd)
		if err != nil {
			return fmt.Errorf("%s-%s", err.Error(), res)
		}
	}
	return nil
}

type HServerInit struct {
	HServerClusterCtx
}

func (s *HServerInit) String() string {
	return "Task: init server cluster"
}

func (s *HServerInit) Run(executor ext.Executor) error {
	server := s.service[0]
	executorCtx := server.Init(s.ctx)
	target := fmt.Sprintf("%s:%d", executorCtx.Target, s.ctx.SSHPort)
	res, err := executor.Execute(target, executorCtx.Cmd)
	if err != nil {
		return fmt.Errorf("%s-%s", err.Error(), res)
	}
	return nil
}

type RemoveHServer struct {
	HServerClusterCtx
}

func (r *RemoveHServer) String() string {
	return "Task: remove hserver"
}

func (r *RemoveHServer) Run(executor ext.Executor) error {
	wg := sync.WaitGroup{}
	mutex := sync.Mutex{}
	var firstErr error
	wg.Add(len(r.service))
	for _, svc := range r.service {
		go func(svc *service.HServer) {
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
