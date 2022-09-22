package task

import (
	"fmt"
	ext "github.com/hstreamdb/dev-deploy/pkg/executor"
	"github.com/hstreamdb/dev-deploy/pkg/service"
)

type MetaStoreClusterCtx struct {
	ctx     *service.GlobalCtx
	service []*service.MetaStore
}

type InitMetaStoreEnv struct {
	MetaStoreClusterCtx
}

func (s *InitMetaStoreEnv) String() string {
	return "Task: init meta store environment"
}

func (s *InitMetaStoreEnv) Run(executor ext.Executor) error {
	return serviceInitEnv(executor, s.ctx, s.service)
}

type SyncMetaStoreConfig struct {
	MetaStoreClusterCtx
}

func (s *SyncMetaStoreConfig) String() string {
	return "Task: sync meta store config"
}

func (s *SyncMetaStoreConfig) Run(executor ext.Executor) error {
	return configSync(executor, s.ctx, s.service)
}

type StartMetaStoreCluster struct {
	MetaStoreClusterCtx
}

func (s *StartMetaStoreCluster) String() string {
	return "Task: start meta store cluster"
}

func (s *StartMetaStoreCluster) Run(executor ext.Executor) error {
	return serviceDeploy(executor, s.ctx, s.service)
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
	target := fmt.Sprintf("%s:%d", executorCtx.Target, m.ctx.SSHPort)
	res, err := executor.Execute(target, executorCtx.Cmd)
	if err != nil {
		return fmt.Errorf("%s-%s", err.Error(), res)
	}
	fmt.Printf("[MetaStore] Get Value: %s\n", res)
	return nil
}

type RemoveMetaStore struct {
	MetaStoreClusterCtx
}

func (r *RemoveMetaStore) String() string {
	return "Task: remove meta store"
}

func (r *RemoveMetaStore) Run(executor ext.Executor) error {
	return serviceRemove(executor, r.ctx, r.service)
}
