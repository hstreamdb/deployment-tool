package task

import (
	"fmt"
	ext "github.com/hstreamdb/dev-deploy/pkg/executor"
	"github.com/hstreamdb/dev-deploy/pkg/service"
	"github.com/hstreamdb/dev-deploy/pkg/spec"
	"os"
)

func SetUpCluster(executor ext.Executor, services *service.Services) error {
	if err := SetUpMetaStoreCluster(executor, services); err != nil {
		return err
	}
	if err := SetUpHStoreCluster(executor, services); err != nil {
		return err
	}
	if err := SetUpHServerCluster(executor, services); err != nil {
		return err
	}

	if err := CheckClusterStatus(executor, services); err != nil {
		return err
	}

	return nil
}

func SetUpHServerCluster(executor ext.Executor, services *service.Services) error {
	serverClusterCtx := HServerClusterCtx{
		ctx:     services.Global,
		service: services.HServer,
	}
	tasks := append([]Task{}, &InitHServerEnv{serverClusterCtx})
	tasks = append(tasks, &SyncHServerConfig{serverClusterCtx})
	tasks = append(tasks, &StartHServerCluster{serverClusterCtx})
	tasks = append(tasks, &WaitServerReady{serverClusterCtx})
	tasks = append(tasks, &HServerInit{serverClusterCtx})
	for _, task := range tasks {
		fmt.Println(task)
		if err := task.Run(executor); err != nil {
			return err
		}
	}
	return nil
}

func SetUpHStoreCluster(executor ext.Executor, services *service.Services) error {
	storeClusterCtx := HStoreClusterCtx{
		ctx:     services.Global,
		service: services.HStore,
	}
	tasks := append([]Task{}, &InitStoreEnv{storeClusterCtx})
	tasks = append(tasks, &SyncStoreConfig{storeClusterCtx})
	tasks = append(tasks, &MountDisk{storeClusterCtx})
	tasks = append(tasks, &StartStoreCluster{storeClusterCtx})
	tasks = append(tasks, &WaitStoreReady{storeClusterCtx})
	tasks = append(tasks, &BootStrap{storeClusterCtx})
	for _, task := range tasks {
		fmt.Println(task)
		if err := task.Run(executor); err != nil {
			fmt.Printf("task %s err: %+v\n", task, err)
			return err
		}
	}

	fmt.Println("Set up HStore cluster success")
	return nil
}

func SetUpMetaStoreCluster(executor ext.Executor, services *service.Services) error {
	metaStoreClusterCtx := MetaStoreClusterCtx{
		ctx:     services.Global,
		service: services.MetaStore,
	}

	tasks := append([]Task{}, &InitMetaStoreEnv{metaStoreClusterCtx})
	tasks = append(tasks, &SyncMetaStoreConfig{metaStoreClusterCtx})
	tasks = append(tasks, &StartMetaStoreCluster{metaStoreClusterCtx})
	tasks = append(tasks, &WaitMetaStoreReady{metaStoreClusterCtx})
	if len(metaStoreClusterCtx.ctx.HStoreConfigInMetaStore) != 0 {
		cfg, err := os.ReadFile(metaStoreClusterCtx.ctx.LocalHStoreConfigFile)
		if err != nil {
			fmt.Printf("%+v\n", err)
			return fmt.Errorf("can't read local store config file, path: %s, err: %s\n",
				metaStoreClusterCtx.ctx.LocalHStoreConfigFile, err.Error())
		}
		tasks = append(tasks, &MetaStoreStoreValue{Key: spec.DefaultStoreConfigPath, Value: string(cfg), MetaStoreClusterCtx: metaStoreClusterCtx})
		tasks = append(tasks, &MetaStoreGetValue{Key: spec.DefaultStoreConfigPath, MetaStoreClusterCtx: metaStoreClusterCtx})
	}

	for _, task := range tasks {
		fmt.Println(task)
		if err := task.Run(executor); err != nil {
			return err
		}
	}

	fmt.Println("Set up meta store cluster success")
	return nil
}

func RemoveCluster(executor ext.Executor, services *service.Services) error {
	if err := RemoveMetaStoreCluster(executor, services); err != nil {
		return err
	}
	if err := RemoveHStoreCluster(executor, services); err != nil {
		return err
	}
	if err := RemoveHServerCluster(executor, services); err != nil {
		return err
	}

	return nil
}

func RemoveMetaStoreCluster(executor ext.Executor, services *service.Services) error {
	metaStoreClusterCtx := MetaStoreClusterCtx{
		ctx:     services.Global,
		service: services.MetaStore,
	}

	tasks := append([]Task{}, &RemoveMetaStore{metaStoreClusterCtx})
	for _, task := range tasks {
		fmt.Println(task)
		if err := task.Run(executor); err != nil {
			return err
		}
	}

	fmt.Println("Remove meta store cluster success")
	return nil
}

func RemoveHStoreCluster(executor ext.Executor, services *service.Services) error {
	storeClusterCtx := HStoreClusterCtx{
		ctx:     services.Global,
		service: services.HStore,
	}

	tasks := append([]Task{}, &RemoveStore{storeClusterCtx})
	for _, task := range tasks {
		fmt.Println(task)
		if err := task.Run(executor); err != nil {
			return err
		}
	}

	fmt.Println("Remove store cluster success")
	return nil
}

func RemoveHServerCluster(executor ext.Executor, services *service.Services) error {
	serverClusterCtx := HServerClusterCtx{
		ctx:     services.Global,
		service: services.HServer,
	}

	tasks := append([]Task{}, &RemoveHServer{serverClusterCtx})
	for _, task := range tasks {
		fmt.Println(task)
		if err := task.Run(executor); err != nil {
			return err
		}
	}

	fmt.Println("Remove server cluster success")
	return nil
}

type ClusterCtx struct {
	ctx            *service.GlobalCtx
	serverServices []*service.HServer
	storeServices  []*service.HStore
}

type CheckClusterStats struct {
	ClusterCtx
}

func (c *CheckClusterStats) String() string {
	return "Task: check cluster status"
}

func (c *CheckClusterStats) Run(executor ext.Executor) error {
	var adminStore *service.HStore
	for _, store := range c.storeServices {
		if store.IsAdmin() {
			adminStore = store
			break
		}
	}

	executorCtx := adminStore.AdminStoreCmd(c.ctx, "status")
	target := fmt.Sprintf("%s:%d", executorCtx.Target, c.ctx.SSHPort)
	res, err := executor.Execute(target, executorCtx.Cmd)
	if err != nil {
		return fmt.Errorf("%s-%s", err.Error(), res)
	}
	fmt.Printf("=== HStore Status ===\n%s\n", res)

	executorCtx = adminStore.AdminServerCmd(c.ctx, c.serverServices[0].GetHost(), "status")
	res, err = executor.Execute(target, executorCtx.Cmd)
	if err != nil {
		return fmt.Errorf("%s-%s", err.Error(), res)
	}
	fmt.Printf("=== HServer Status ===\n%s\n", res)
	return nil
}

func CheckClusterStatus(executor ext.Executor, services *service.Services) error {
	clusterCtx := ClusterCtx{
		ctx:            services.Global,
		serverServices: services.HServer,
		storeServices:  services.HStore,
	}
	task := &CheckClusterStats{clusterCtx}
	if err := task.Run(executor); err != nil {
		return err
	}
	return nil
}
