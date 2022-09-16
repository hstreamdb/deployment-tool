package task

import (
	"fmt"
	ext "github.com/hstreamdb/dev-deploy/pkg/executor"
	"github.com/hstreamdb/dev-deploy/pkg/service"
	"sync"
)

type HStoreClusterCtx struct {
	ctx     *service.GlobalCtx
	service []*service.HStore
}

type InitStoreEnv struct {
	HStoreClusterCtx
}

func (s *InitStoreEnv) String() string {
	return "Task: init store environment"
}

func (s *InitStoreEnv) Run(executor ext.Executor) error {
	wg := sync.WaitGroup{}
	mutex := sync.Mutex{}
	var firstErr error
	wg.Add(len(s.service))
	for _, svc := range s.service {
		go func(svc *service.HStore) {
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

type SyncStoreConfig struct {
	HStoreClusterCtx
}

func (s *SyncStoreConfig) String() string {
	return "Task: sync store config"
}

func (s *SyncStoreConfig) Run(executor ext.Executor) error {
	wg := sync.WaitGroup{}
	mutex := sync.Mutex{}
	var firstErr error
	wg.Add(len(s.service))
	for _, svc := range s.service {
		go func(svc *service.HStore) {
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

type StartStoreCluster struct {
	HStoreClusterCtx
}

func (s *StartStoreCluster) String() string {
	return "Task: start store cluster"
}

func (s *StartStoreCluster) Run(executor ext.Executor) error {
	wg := sync.WaitGroup{}
	mutex := sync.Mutex{}
	var firstErr error
	wg.Add(len(s.service))
	for _, svc := range s.service {
		go func(svc *service.HStore) {
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

type WaitStoreReady struct {
	HStoreClusterCtx
}

func (w *WaitStoreReady) String() string {
	return "Task: wait store ready"
}

func (w *WaitStoreReady) Run(executor ext.Executor) error {
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

type BootStrap struct {
	HStoreClusterCtx
}

func (b *BootStrap) String() string {
	return "Task: bootstrap"
}

func (b *BootStrap) Run(executor ext.Executor) error {
	var adminStore *service.HStore
	for _, store := range b.service {
		if store.IsAdmin() {
			adminStore = store
			break
		}
	}

	executorCtx := adminStore.Bootstrap(b.ctx)
	target := fmt.Sprintf("%s:%d", executorCtx.Target, b.ctx.SSHPort)
	res, err := executor.Execute(target, executorCtx.Cmd)
	if err != nil {
		return fmt.Errorf("%s-%s", err.Error(), res)
	}
	return nil
}

type RemoveStore struct {
	HStoreClusterCtx
}

func (r *RemoveStore) String() string {
	return "Task: remove store"
}

func (r *RemoveStore) Run(executor ext.Executor) error {
	wg := sync.WaitGroup{}
	mutex := sync.Mutex{}
	var firstErr error
	wg.Add(len(r.service))
	for _, svc := range r.service {
		go func(svc *service.HStore) {
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

type MountDisk struct {
	HStoreClusterCtx
}

func (m *MountDisk) String() string {
	return "Task: mount disk"
}

func (m *MountDisk) Run(executor ext.Executor) error {
	wg := sync.WaitGroup{}
	mutex := sync.Mutex{}
	var firstErr error
	wg.Add(len(m.service))
	for _, svc := range m.service {
		go func(svc *service.HStore) {
			defer wg.Done()
			executorCtx := svc.MountDisk()
			target := fmt.Sprintf("%s:%d", executorCtx.Target, m.ctx.SSHPort)
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
