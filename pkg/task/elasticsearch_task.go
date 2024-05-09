package task

import (
	"fmt"

	ext "github.com/hstreamdb/deployment-tool/pkg/executor"
	"github.com/hstreamdb/deployment-tool/pkg/service"
)

type EsClusterCtx struct {
	ctx     *service.GlobalCtx
	service []*service.ElasticSearch
}

type ConfigLogIndex struct {
	EsClusterCtx
}

func (e *EsClusterCtx) String() string {
	return "Task: config elasticsearch log index"
}

func (e *EsClusterCtx) Run(executor ext.Executor) error {
	es := e.service[0]
	executorCtx := es.ConfigLogIndex(e.ctx)
	target := fmt.Sprintf("%s:%d", executorCtx.Target, e.ctx.SSHPort)
	res, err := executor.Execute(target, executorCtx.Cmd)
	if err != nil {
		return fmt.Errorf("%s-%s", err.Error(), res)
	}
	return nil
}
