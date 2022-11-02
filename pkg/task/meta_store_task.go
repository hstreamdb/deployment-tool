package task

import (
	"fmt"
	ext "github.com/hstreamdb/deployment-tool/pkg/executor"
	"github.com/hstreamdb/deployment-tool/pkg/service"
)

type MetaStoreClusterCtx struct {
	ctx     *service.GlobalCtx
	service []*service.MetaStore
}

type WaitMetaStoreReady struct {
	MetaStoreClusterCtx
}

func (w *WaitMetaStoreReady) String() string {
	return "Task: wait meta store ready"
}

func (w *WaitMetaStoreReady) Run(executor ext.Executor) error {
	for _, metaStore := range w.service {
		executorCtx := metaStore.CheckReady(w.ctx)
		if executorCtx == nil {
			fmt.Printf("skip wait meta store ready for %s\n", metaStore.GetServiceName())
			return nil
		}
		target := fmt.Sprintf("%s:%d", executorCtx.Target, w.ctx.SSHPort)
		res, err := executor.Execute(target, executorCtx.Cmd)
		if err != nil {
			return fmt.Errorf("%s-%s", err.Error(), res)
		}
	}
	return nil
}

type MetaStoreStoreValue struct {
	MetaStoreClusterCtx
	Key   string
	Value string
}

func (m *MetaStoreStoreValue) String() string {
	return "Task: store value to meta store"
}

func (m *MetaStoreStoreValue) Run(executor ext.Executor) error {
	svc := m.service[0]
	executorCtx := svc.StoreValue(m.Key, m.Value)
	if executorCtx == nil {
		fmt.Printf("skip store value to metastore")
		return nil
	}
	target := fmt.Sprintf("%s:%d", executorCtx.Target, m.ctx.SSHPort)
	res, err := executor.Execute(target, executorCtx.Cmd)
	if err != nil {
		return fmt.Errorf("%s-%s", err.Error(), res)
	}
	return nil
}

type MetaStoreGetValue struct {
	MetaStoreClusterCtx
	Key string
}

func (m *MetaStoreGetValue) String() string {
	return "Task: get value from meta store"
}

func (m *MetaStoreGetValue) Run(executor ext.Executor) error {
	svc := m.service[0]
	executorCtx := svc.GetValue(m.Key)
	if executorCtx == nil {
		fmt.Printf("skip get value from metastore")
		return nil
	}
	target := fmt.Sprintf("%s:%d", executorCtx.Target, m.ctx.SSHPort)
	res, err := executor.Execute(target, executorCtx.Cmd)
	if err != nil {
		return fmt.Errorf("%s-%s", err.Error(), res)
	}
	fmt.Printf("[MetaStore] Get Value: %s\n", res)
	return nil
}
