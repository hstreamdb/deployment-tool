package task

import (
	"fmt"
	ext "github.com/hstreamdb/deployment-tool/pkg/executor"
	"github.com/hstreamdb/deployment-tool/pkg/service"
	"github.com/hstreamdb/deployment-tool/pkg/spec"
	"os"
)

type runCtx struct {
	executor ext.Executor
	services *service.Services
	err      error
}

func (r *runCtx) run(f func(executor ext.Executor, services *service.Services) error) {
	if r.err != nil {
		return
	}
	r.err = f(r.executor, r.services)
}

func SetUpCluster(executor ext.Executor, services *service.Services) error {
	ctx := runCtx{executor: executor, services: services}
	fmt.Println("============ Set up cluster with components ============")
	services.ShowAllServices()

	ctx.run(SetUpMetaStoreCluster)
	ctx.run(SetUpHStoreCluster)
	ctx.run(SetUpHServerCluster)
	ctx.run(CheckClusterStatus)
	ctx.run(SetUpHttpServerService)
	if len(services.Prometheus) != 0 {
		ctx.run(SetUpHStreamMonitorStack)
		ctx.run(SetUpHStreamExporterService)
		ctx.run(SetUpPrometheusService)
		ctx.run(SetUpGrafanaService)
		ctx.run(SetUpAlertService)
	}

	fmt.Printf("[DEBUG]: len(services.Filebeat) = %v\n", len(services.Filebeat))
	if len(services.Filebeat) != 0 {
		ctx.run(SetUpElasticSearch)
		ctx.run(SetUpKibana)
		ctx.run(SetUpFilebeat)
	}
	return ctx.err
}

func RemoveCluster(executor ext.Executor, services *service.Services) error {
	ctx := runCtx{executor: executor, services: services}
	if len(services.Prometheus) != 0 {
		ctx.run(RemoveAlertService)
		ctx.run(RemoveGrafanaService)
		ctx.run(RemovePrometheusService)
		ctx.run(RemoveHStreamExporterService)
		ctx.run(RemoveHStreamMonitorStack)
	}
	if len(services.Filebeat) != 0 {
		ctx.run(RemoveFilebeat)
		ctx.run(RemoveKibana)
		ctx.run(RemoveElasticSearch)
	}

	ctx.run(RemoveHttpServerService)
	ctx.run(RemoveHServerCluster)
	ctx.run(RemoveHStoreCluster)
	ctx.run(RemoveMetaStoreCluster)
	return ctx.err
}

// ==========================================================================================================

func SetUpMetaStoreCluster(executor ext.Executor, services *service.Services) error {
	if len(services.MetaStore) == 0 {
		return nil
	}

	metaStoreClusterCtx := MetaStoreClusterCtx{
		ctx:     services.Global,
		service: services.MetaStore,
	}

	tasks := getStartServiceTask(metaStoreClusterCtx.ctx, metaStoreClusterCtx.service)
	tasks = append(tasks, &WaitMetaStoreReady{metaStoreClusterCtx})
	if len(metaStoreClusterCtx.ctx.HStoreConfigInMetaStore) != 0 {
		cfg, err := os.ReadFile(metaStoreClusterCtx.ctx.LocalHStoreConfigFile)
		if err != nil {
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

func RemoveMetaStoreCluster(executor ext.Executor, services *service.Services) error {
	return removeCluster(executor, services.Global, services.MetaStore)
}

func RemoveFilebeat(executor ext.Executor, services *service.Services) error {
	return removeCluster(executor, services.Global, services.Filebeat)
}

func RemoveKibana(executor ext.Executor, services *service.Services) error {
	return removeCluster(executor, services.Global, services.Kibana)
}

func RemoveElasticSearch(executor ext.Executor, services *service.Services) error {
	return removeCluster(executor, services.Global, services.ElasticSearch)
}

func SetUpHStoreCluster(executor ext.Executor, services *service.Services) error {
	if len(services.HStore) == 0 {
		return nil
	}

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

func RemoveHStoreCluster(executor ext.Executor, services *service.Services) error {
	return removeCluster(executor, services.Global, services.HStore)
}

func SetUpHServerCluster(executor ext.Executor, services *service.Services) error {
	if len(services.HServer) == 0 {
		return nil
	}

	serverClusterCtx := HServerClusterCtx{
		ctx:     services.Global,
		service: services.HServer,
	}
	tasks := getStartServiceTask(serverClusterCtx.ctx, serverClusterCtx.service)
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

func RemoveHServerCluster(executor ext.Executor, services *service.Services) error {
	return removeCluster(executor, services.Global, services.HServer)
}

func SetUpHttpServerService(executor ext.Executor, services *service.Services) error {
	return startCluster(executor, services.Global, services.HttpServer)
}

func RemoveHttpServerService(executor ext.Executor, services *service.Services) error {
	return removeCluster(executor, services.Global, services.HttpServer)
}

func SetUpHStreamMonitorStack(executor ext.Executor, services *service.Services) error {
	return startCluster(executor, services.Global, services.MonitorSuite)
}

func RemoveHStreamMonitorStack(executor ext.Executor, services *service.Services) error {
	return removeCluster(executor, services.Global, services.MonitorSuite)
}

func SetUpHStreamExporterService(executor ext.Executor, services *service.Services) error {
	return startCluster(executor, services.Global, services.HStreamExporter)
}

func RemoveHStreamExporterService(executor ext.Executor, services *service.Services) error {
	return removeCluster(executor, services.Global, services.HStreamExporter)
}

func SetUpPrometheusService(executor ext.Executor, services *service.Services) error {
	return startCluster(executor, services.Global, services.Prometheus)
}

func RemovePrometheusService(executor ext.Executor, services *service.Services) error {
	return removeCluster(executor, services.Global, services.Prometheus)
}

func SetUpGrafanaService(executor ext.Executor, services *service.Services) error {
	return startCluster(executor, services.Global, services.Grafana)
}

func RemoveGrafanaService(executor ext.Executor, services *service.Services) error {
	return removeCluster(executor, services.Global, services.Grafana)
}

func SetUpAlertService(executor ext.Executor, services *service.Services) error {
	return startCluster(executor, services.Global, services.AlertManager)
}

func RemoveAlertService(executor ext.Executor, services *service.Services) error {
	return removeCluster(executor, services.Global, services.AlertManager)
}

func SetUpElasticSearch(executor ext.Executor, services *service.Services) error {
	return startCluster(executor, services.Global, services.ElasticSearch)
}

func SetUpKibana(executor ext.Executor, services *service.Services) error {
	return startCluster(executor, services.Global, services.Kibana)
}

func SetUpFilebeat(executor ext.Executor, services *service.Services) error {
	return startCluster(executor, services.Global, services.Filebeat)
}

func startCluster[S service.Service](executor ext.Executor, ctx *service.GlobalCtx, services []S) error {
	if len(services) == 0 {
		return nil
	}

	tasks := getStartServiceTask(ctx, services)
	for _, task := range tasks {
		fmt.Println(task)
		if err := task.Run(executor); err != nil {
			return err
		}
	}

	fmt.Printf("Set up %s service success\n", services[0].GetServiceName())
	return nil
}

func removeCluster[S service.Service](executor ext.Executor, ctx *service.GlobalCtx, services []S) error {
	if len(services) == 0 {
		return nil
	}

	tasks := getRemoveServiceTask(ctx, services)
	for _, task := range tasks {
		fmt.Println(task)
		if err := task.Run(executor); err != nil {
			return err
		}
	}

	fmt.Printf("Remove %s cluster success\n", services[0].GetServiceName())
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
	if len(c.storeServices) == 0 {
		return nil
	}

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

	if len(c.serverServices) == 0 {
		return nil
	}

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
