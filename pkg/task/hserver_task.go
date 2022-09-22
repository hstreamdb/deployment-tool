package task

import (
	"fmt"
	ext "github.com/hstreamdb/dev-deploy/pkg/executor"
	"github.com/hstreamdb/dev-deploy/pkg/service"
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
	return serviceInitEnv(executor, s.ctx, s.service)
}

type SyncHServerConfig struct {
	HServerClusterCtx
}

func (s *SyncHServerConfig) String() string {
	return "Task: sync server config"
}

func (s *SyncHServerConfig) Run(executor ext.Executor) error {
	return configSync(executor, s.ctx, s.service)
}

type StartHServerCluster struct {
	HServerClusterCtx
}

func (s *StartHServerCluster) String() string {
	return "Task: start server cluster"
}

func (s *StartHServerCluster) Run(executor ext.Executor) error {
	return serviceDeploy(executor, s.ctx, s.service)
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
	return serviceRemove(executor, r.ctx, r.service)
}
