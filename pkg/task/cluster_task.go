package task

import (
	"fmt"
	"os"

	ext "github.com/hstreamdb/deployment-tool/pkg/executor"
	"github.com/hstreamdb/deployment-tool/pkg/service"
	"github.com/hstreamdb/deployment-tool/pkg/spec"
)

type runCtx struct {
	executor  ext.Executor
	services  *service.Services
	ignoreErr bool
	err       error
}

func (r *runCtx) run(f func(executor ext.Executor, services *service.Services) error) {
	if r.err != nil {
		if !r.ignoreErr {
			return
		}

		fmt.Printf("ignore error: %s\n", r.err.Error())
	}
	r.err = f(r.executor, r.services)
}

func SetUpCluster(executor ext.Executor, services *service.Services) error {
	ctx := runCtx{executor: executor, services: services, ignoreErr: false}
	fmt.Println("============ Set up cluster with components ============")
	services.ShowAllServices()

	ctx.run(SetUpMetaStoreCluster)
	ctx.run(SetUpHAdminCluster)
	ctx.run(SetUpHStoreCluster)
	ctx.run(Bootstrap)
	ctx.run(SetUpHServerCluster)
	ctx.run(CheckClusterStatus)

	if len(services.BlackBox) != 0 {
		ctx.run(SetUpBlackBoxService)
	}

	if len(services.Prometheus) != 0 {
		ctx.run(SetUpHStreamMonitorStack)
		ctx.run(SetUpHStreamExporterService)
		ctx.run(SetUpPrometheusService)
		ctx.run(SetUpGrafanaService)
		ctx.run(SetUpAlertService)
	}

	if len(services.HStreamConsole) != 0 {
		ctx.run(SetUpHStreamConsole)
	}

	if len(services.ElasticSearch) != 0 {
		ctx.run(SetUpElasticSearch)
		ctx.run(SetUpKibana)
		ctx.run(SetUpFilebeat)
		ctx.run(SetUpVector)
	}
	return ctx.err
}

func RemoveCluster(executor ext.Executor, services *service.Services) error {
	ctx := runCtx{executor: executor, services: services, ignoreErr: true}
	if len(services.ElasticSearch) != 0 {
		ctx.run(RemoveVector)
		ctx.run(RemoveFilebeat)
		ctx.run(RemoveKibana)
		ctx.run(RemoveElasticSearch)
	}

	if len(services.HStreamConsole) != 0 {
		ctx.run(RemoveHStreamConsole)
	}

	if len(services.Prometheus) != 0 {
		ctx.run(RemoveAlertService)
		ctx.run(RemoveGrafanaService)
		ctx.run(RemovePrometheusService)
		ctx.run(RemoveHStreamExporterService)
		ctx.run(RemoveHStreamMonitorStack)
	}

	if len(services.BlackBox) != 0 {
		ctx.run(RemoveBlackBoxService)
	}

	ctx.run(RemoveHServerCluster)
	ctx.run(RemoveHStoreCluster)
	ctx.run(RemoveHAdminCluster)
	ctx.run(RemoveMetaStoreCluster)
	return ctx.err
}

func StopCluster(executor ext.Executor, services *service.Services) error {
	ctx := runCtx{executor: executor, services: services}
	if len(services.ElasticSearch) != 0 {
		ctx.run(StopVector)
		ctx.run(StopFilebeat)
		ctx.run(StopKibana)
		ctx.run(StopElasticSearch)
	}

	if len(services.HStreamConsole) != 0 {
		ctx.run(StopHStreamConsole)
	}

	if len(services.Prometheus) != 0 {
		ctx.run(StopAlertService)
		ctx.run(StopGrafanaService)
		ctx.run(StopPrometheusService)
		ctx.run(StopHStreamExporterService)
		ctx.run(StopHStreamMonitorStack)
	}

	if len(services.BlackBox) != 0 {
		ctx.run(StopBlackBoxService)
	}

	ctx.run(StopHServerCluster)
	ctx.run(StopHStoreCluster)
	ctx.run(StopHAdminCluster)
	ctx.run(StopMetaStoreCluster)
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

func StopMetaStoreCluster(executor ext.Executor, services *service.Services) error {
	return stopCluster(executor, services.Global, services.MetaStore)
}

func SetUpHAdminCluster(executor ext.Executor, services *service.Services) error {
	return startCluster(executor, services.Global, services.HAdmin)
}

func RemoveHAdminCluster(executor ext.Executor, services *service.Services) error {
	return removeCluster(executor, services.Global, services.HAdmin)
}

func StopHAdminCluster(executor ext.Executor, services *service.Services) error {
	return stopCluster(executor, services.Global, services.HAdmin)
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

func StopHStoreCluster(executor ext.Executor, services *service.Services) error {
	return stopCluster(executor, services.Global, services.HStore)
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

func StopHServerCluster(executor ext.Executor, services *service.Services) error {
	return stopCluster(executor, services.Global, services.HServer)
}

func SetUpHStreamConsole(executor ext.Executor, services *service.Services) error {
	return startCluster(executor, services.Global, services.HStreamConsole)
}

func RemoveHStreamConsole(executor ext.Executor, services *service.Services) error {
	return removeCluster(executor, services.Global, services.HStreamConsole)
}

func StopHStreamConsole(executor ext.Executor, services *service.Services) error {
	return stopCluster(executor, services.Global, services.HStreamConsole)
}

func SetUpHStreamMonitorStack(executor ext.Executor, services *service.Services) error {
	return startCluster(executor, services.Global, services.MonitorSuite)
}

func RemoveHStreamMonitorStack(executor ext.Executor, services *service.Services) error {
	return removeCluster(executor, services.Global, services.MonitorSuite)
}

func StopHStreamMonitorStack(executor ext.Executor, services *service.Services) error {
	return stopCluster(executor, services.Global, services.MonitorSuite)
}

func SetUpHStreamExporterService(executor ext.Executor, services *service.Services) error {
	return startCluster(executor, services.Global, services.HStreamExporter)
}

func RemoveHStreamExporterService(executor ext.Executor, services *service.Services) error {
	return removeCluster(executor, services.Global, services.HStreamExporter)
}

func StopHStreamExporterService(executor ext.Executor, services *service.Services) error {
	return stopCluster(executor, services.Global, services.HStreamExporter)
}

func SetUpBlackBoxService(executor ext.Executor, services *service.Services) error {
	return startCluster(executor, services.Global, services.BlackBox)
}

func RemoveBlackBoxService(executor ext.Executor, services *service.Services) error {
	return removeCluster(executor, services.Global, services.BlackBox)
}

func StopBlackBoxService(executor ext.Executor, services *service.Services) error {
	return stopCluster(executor, services.Global, services.BlackBox)
}

func SetUpPrometheusService(executor ext.Executor, services *service.Services) error {
	return startCluster(executor, services.Global, services.Prometheus)
}

func RemovePrometheusService(executor ext.Executor, services *service.Services) error {
	return removeCluster(executor, services.Global, services.Prometheus)
}

func StopPrometheusService(executor ext.Executor, services *service.Services) error {
	return stopCluster(executor, services.Global, services.Prometheus)
}

func SetUpGrafanaService(executor ext.Executor, services *service.Services) error {
	return startCluster(executor, services.Global, services.Grafana)
}

func RemoveGrafanaService(executor ext.Executor, services *service.Services) error {
	return removeCluster(executor, services.Global, services.Grafana)
}

func StopGrafanaService(executor ext.Executor, services *service.Services) error {
	return stopCluster(executor, services.Global, services.Grafana)
}

func SetUpAlertService(executor ext.Executor, services *service.Services) error {
	return startCluster(executor, services.Global, services.AlertManager)
}

func RemoveAlertService(executor ext.Executor, services *service.Services) error {
	return removeCluster(executor, services.Global, services.AlertManager)
}

func StopAlertService(executor ext.Executor, services *service.Services) error {
	return stopCluster(executor, services.Global, services.AlertManager)
}

func SetUpElasticSearch(executor ext.Executor, services *service.Services) error {
	return startCluster(executor, services.Global, services.ElasticSearch)
}

func RemoveElasticSearch(executor ext.Executor, services *service.Services) error {
	return removeCluster(executor, services.Global, services.ElasticSearch)
}

func StopElasticSearch(executor ext.Executor, services *service.Services) error {
	return stopCluster(executor, services.Global, services.ElasticSearch)
}

func SetUpKibana(executor ext.Executor, services *service.Services) error {
	if len(services.Kibana) == 0 {
		return nil
	}

	kibanaClusterCtx := KibanaClusterCtx{
		ctx:     services.Global,
		service: services.Kibana,
	}
	tasks := getStartServiceTask(kibanaClusterCtx.ctx, kibanaClusterCtx.service)
	tasks = append(tasks, &WaitKibanaReady{kibanaClusterCtx})
	for _, task := range tasks {
		fmt.Println(task)
		if err := task.Run(executor); err != nil {
			return err
		}
	}
	return nil
}

func RemoveKibana(executor ext.Executor, services *service.Services) error {
	return removeCluster(executor, services.Global, services.Kibana)
}

func StopKibana(executor ext.Executor, services *service.Services) error {
	return stopCluster(executor, services.Global, services.Kibana)
}

func SetUpFilebeat(executor ext.Executor, services *service.Services) error {
	return startCluster(executor, services.Global, services.Filebeat)
}

func RemoveFilebeat(executor ext.Executor, services *service.Services) error {
	return removeCluster(executor, services.Global, services.Filebeat)
}

func StopFilebeat(executor ext.Executor, services *service.Services) error {
	return stopCluster(executor, services.Global, services.Filebeat)
}

func SetUpVector(executor ext.Executor, services *service.Services) error {
	return startCluster(executor, services.Global, services.Vector)
}

func RemoveVector(executor ext.Executor, services *service.Services) error {
	return removeCluster(executor, services.Global, services.Vector)
}

func StopVector(executor ext.Executor, services *service.Services) error {
	return stopCluster(executor, services.Global, services.Vector)
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

func stopCluster[S service.Service](executor ext.Executor, ctx *service.GlobalCtx, services []S) error {
	if len(services) == 0 {
		return nil
	}

	tasks := getStopServiceTask(ctx, services)
	for _, task := range tasks {
		fmt.Println(task)
		if err := task.Run(executor); err != nil {
			return err
		}
	}

	fmt.Printf("Stop %s cluster success\n", services[0].GetServiceName())
	return nil
}

func CheckClusterStatus(executor ext.Executor, services *service.Services) error {
	task := &CheckClusterStatusCtx{
		ctx:            services.Global,
		serverServices: services.HServer,
	}

	fmt.Println(task)
	if err := task.Run(executor); err != nil {
		return err
	}
	return nil
}

func Bootstrap(executor ext.Executor, services *service.Services) error {
	task := &BootstrapCtx{
		ctx: services.Global,
	}

	fmt.Println(task)
	if err := task.Run(executor); err != nil {
		return err
	}
	return nil
}
