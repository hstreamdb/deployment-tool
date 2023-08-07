package task

import (
	"fmt"
	ext "github.com/hstreamdb/deployment-tool/pkg/executor"
	"github.com/hstreamdb/deployment-tool/pkg/service"
)

type HServerClusterCtx struct {
	ctx     *service.GlobalCtx
	service []*service.HServer
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
	if executorCtx == nil {
		fmt.Printf("skip init hserver")
		return nil
	}
	target := fmt.Sprintf("%s:%d", executorCtx.Target, s.ctx.SSHPort)
	_, err := executor.Execute(target, executorCtx.Cmd)
	return err
}
