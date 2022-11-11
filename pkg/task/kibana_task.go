package task

import (
	"fmt"
	ext "github.com/hstreamdb/deployment-tool/pkg/executor"
	"github.com/hstreamdb/deployment-tool/pkg/service"
)

type KibanaClusterCtx struct {
	ctx     *service.GlobalCtx
	service []*service.Kibana
}

type WaitKibanaReady struct {
	KibanaClusterCtx
}

func (w *WaitKibanaReady) String() string {
	return "Task: wait kibana ready"
}

func (w *WaitKibanaReady) Run(executor ext.Executor) error {
	for _, v := range w.service {
		executorCtx := v.CheckReady()
		target := fmt.Sprintf("%s:%d", executorCtx.Target, v.GetSSHHost())
		res, err := executor.Execute(target, executorCtx.Cmd)
		if err != nil {
			return fmt.Errorf("%s-%s", err.Error(), res)
		}
	}
	return nil
}
