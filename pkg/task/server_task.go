package task

import (
	"fmt"
	ext "github.com/hstreamdb/deployment-tool/pkg/executor"
	"github.com/hstreamdb/deployment-tool/pkg/service"
	log "github.com/sirupsen/logrus"
)

type ServerType interface {
	Init(ctx *service.GlobalCtx) *ext.ExecuteCtx
	CheckReady(globalCtx *service.GlobalCtx) *ext.ExecuteCtx
	GetStatus(globalCtx *service.GlobalCtx) *ext.ExecuteCtx
	*service.HServer | *service.HServerKafka
}

type ServerClusterCtx[S ServerType] struct {
	ctx     *service.GlobalCtx
	service []S
}

type WaitServerReady[S ServerType] struct {
	ServerClusterCtx[S]
}

func (w *WaitServerReady[S]) String() string {
	return "Task: wait server ready"
}

func (w *WaitServerReady[S]) Run(executor ext.Executor) error {
	for _, server := range w.service {
		executorCtx := server.CheckReady(w.ctx)
		target := fmt.Sprintf("%s:%d", executorCtx.Target, w.ctx.SSHPort)
		res, err := executor.Execute(target, executorCtx.Cmd)
		if err != nil {
			return fmt.Errorf("%s-%s", err.Error(), res)
		}
	}
	return nil
}

type HServerInit[S ServerType] struct {
	ServerClusterCtx[S]
}

func (s *HServerInit[S]) String() string {
	return "Task: init server cluster"
}

func (s *HServerInit[S]) Run(executor ext.Executor) error {
	server := s.service[0]
	executorCtx := server.Init(s.ctx)
	if executorCtx == nil {
		fmt.Printf("skip init hserver")
		return nil
	}
	target := fmt.Sprintf("%s:%d", executorCtx.Target, s.ctx.SSHPort)
	_, err := executor.Execute(target, executorCtx.Cmd)
	if err != nil {
		log.Warningf("init hserver err, need to double check server status, err: %s", err.Error())
	}
	return nil
}
