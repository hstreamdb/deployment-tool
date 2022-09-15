package task

import (
	"fmt"
	ext "github.com/hstreamdb/dev-deploy/pkg/executor"
	"github.com/hstreamdb/dev-deploy/pkg/service"
	"sync"
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
	wg := sync.WaitGroup{}
	mutex := sync.Mutex{}
	var firstErr error
	wg.Add(len(s.service))
	for _, svc := range s.service {
		go func(svc *service.MetaStore) {
			defer wg.Done()
			executorCtx := svc.InitEnv(s.ctx)
			target := fmt.Sprintf("%s:%d", executorCtx.Target, s.ctx.SshPort)
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

type SyncMetaStoreConfig struct {
	MetaStoreClusterCtx
}

func (s *SyncMetaStoreConfig) String() string {
	return "Task: sync meta store config"
}

func (s *SyncMetaStoreConfig) Run(executor ext.Executor) error {
	wg := sync.WaitGroup{}
	mutex := sync.Mutex{}
	var firstErr error
	wg.Add(len(s.service))
	for _, svc := range s.service {
		go func(svc *service.MetaStore) {
			defer wg.Done()
			transferCtx := svc.SyncConfig(s.ctx)
			if transferCtx == nil {
				fmt.Printf("skip %s\n", s)
				return
			}
			target := fmt.Sprintf("%s:%d", transferCtx.Target, s.ctx.SshPort)
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

type StartMetaStoreCluster struct {
	MetaStoreClusterCtx
}

func (s *StartMetaStoreCluster) String() string {
	return "Task: start meta store cluster"
}

func (s *StartMetaStoreCluster) Run(executor ext.Executor) error {
	wg := sync.WaitGroup{}
	mutex := sync.Mutex{}
	var firstErr error
	wg.Add(len(s.service))
	for _, svc := range s.service {
		go func(svc *service.MetaStore) {
			defer wg.Done()
			executorCtx := svc.Deploy(s.ctx)
			target := fmt.Sprintf("%s:%d", executorCtx.Target, s.ctx.SshPort)
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

type WaitMetaStoreReady struct {
	MetaStoreClusterCtx
}

func (w *WaitMetaStoreReady) String() string {
	return "Task: wait meta store ready"
}

func (w *WaitMetaStoreReady) Run(executor ext.Executor) error {
	for _, metaStore := range w.service {
		executorCtx := metaStore.CheckReady(w.ctx)
		target := fmt.Sprintf("%s:%d", executorCtx.Target, w.ctx.SshPort)
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
	service := m.service[0]
	executorCtx := service.StoreValue(m.Key, m.Value)
	target := fmt.Sprintf("%s:%d", executorCtx.Target, m.ctx.SshPort)
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
	service := m.service[0]
	executorCtx := service.GetValue(m.Key)
	target := fmt.Sprintf("%s:%d", executorCtx.Target, m.ctx.SshPort)
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
	wg := sync.WaitGroup{}
	mutex := sync.Mutex{}
	var firstErr error
	wg.Add(len(r.service))
	for _, svc := range r.service {
		go func(svc *service.MetaStore) {
			defer wg.Done()
			executorCtx := svc.Remove(r.ctx)
			target := fmt.Sprintf("%s:%d", executorCtx.Target, r.ctx.SshPort)
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
