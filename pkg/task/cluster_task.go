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

	return ctx.err
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
	ctx := runCtx{executor: executor, services: services}
	if len(services.Prometheus) != 0 {
		ctx.run(RemoveAlertService)
		ctx.run(RemoveGrafanaService)
		ctx.run(RemovePrometheusService)
		ctx.run(RemoveHStreamExporterService)
		ctx.run(RemoveHStreamMonitorStack)
	}
	ctx.run(RemoveHttpServerService)
	ctx.run(RemoveHServerCluster)
	ctx.run(RemoveHStoreCluster)
	ctx.run(RemoveMetaStoreCluster)
	return ctx.err
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

func SetUpHttpServerService(executor ext.Executor, services *service.Services) error {
	httpServerCtx := HttpServerCtx{
		ctx:     services.Global,
		service: services.HttpServer,
	}

	tasks := append([]Task{}, &InitHttpServerEnv{httpServerCtx})
	tasks = append(tasks, &SyncHttpServerConfig{httpServerCtx})
	tasks = append(tasks, &StartHttpServer{httpServerCtx})
	for _, task := range tasks {
		fmt.Println(task)
		if err := task.Run(executor); err != nil {
			return err
		}
	}

	fmt.Println("Set up http-server success")
	return nil
}

func RemoveHttpServerService(executor ext.Executor, services *service.Services) error {
	httpServerCtx := HttpServerCtx{
		ctx:     services.Global,
		service: services.HttpServer,
	}

	tasks := append([]Task{}, &RemoveHttpServer{httpServerCtx})
	for _, task := range tasks {
		fmt.Println(task)
		if err := task.Run(executor); err != nil {
			return err
		}
	}

	fmt.Println("Remove http-server success")
	return nil
}

func SetUpHStreamMonitorStack(executor ext.Executor, services *service.Services) error {
	monitorSuiteCtx := MonitorSuiteCtx{
		ctx:     services.Global,
		service: services.MonitorSuite,
	}

	tasks := append([]Task{}, &InitMonitorSuiteEnv{monitorSuiteCtx})
	tasks = append(tasks, &SyncMonitorSuiteConfig{monitorSuiteCtx})
	tasks = append(tasks, &StartMonitorSuite{monitorSuiteCtx})
	for _, task := range tasks {
		fmt.Println(task)
		if err := task.Run(executor); err != nil {
			return err
		}
	}

	fmt.Println("Set up monitor stack success")
	return nil
}

func RemoveHStreamMonitorStack(executor ext.Executor, services *service.Services) error {
	monitorSuiteCtx := MonitorSuiteCtx{
		ctx:     services.Global,
		service: services.MonitorSuite,
	}

	tasks := append([]Task{}, &RemoveMonitorSuite{monitorSuiteCtx})
	for _, task := range tasks {
		fmt.Println(task)
		if err := task.Run(executor); err != nil {
			return err
		}
	}

	fmt.Println("Remove monitor stack success")
	return nil
}

func SetUpHStreamExporterService(executor ext.Executor, services *service.Services) error {
	hstreamExporterCtx := HStreamExporterCtx{
		ctx:     services.Global,
		service: services.HStreamExporter,
	}

	tasks := append([]Task{}, &InitHStreamExporter{hstreamExporterCtx})
	tasks = append(tasks, &SyncHStreamExporterConfig{hstreamExporterCtx})
	tasks = append(tasks, &StartHStreamExporter{hstreamExporterCtx})
	for _, task := range tasks {
		fmt.Println(task)
		if err := task.Run(executor); err != nil {
			return err
		}
	}

	fmt.Println("Set up hstream-exporter service success")
	return nil
}

func RemoveHStreamExporterService(executor ext.Executor, services *service.Services) error {
	hstreamExporterCtx := HStreamExporterCtx{
		ctx:     services.Global,
		service: services.HStreamExporter,
	}

	tasks := append([]Task{}, &RemoveHStreamExporter{hstreamExporterCtx})
	for _, task := range tasks {
		fmt.Println(task)
		if err := task.Run(executor); err != nil {
			return err
		}
	}

	fmt.Println("Remove hstream-exporter service success")
	return nil
}

func SetUpPrometheusService(executor ext.Executor, services *service.Services) error {
	prometheusCtx := PrometheusCtx{
		ctx:     services.Global,
		service: services.Prometheus,
	}

	tasks := append([]Task{}, &InitPrometheus{prometheusCtx})
	tasks = append(tasks, &SyncPrometheusConfig{prometheusCtx})
	tasks = append(tasks, &StartPrometheus{prometheusCtx})
	for _, task := range tasks {
		fmt.Println(task)
		if err := task.Run(executor); err != nil {
			return err
		}
	}

	fmt.Println("Set up prometheus service success")
	return nil
}

func RemovePrometheusService(executor ext.Executor, services *service.Services) error {
	prometheusCtx := PrometheusCtx{
		ctx:     services.Global,
		service: services.Prometheus,
	}

	tasks := append([]Task{}, &RemovePrometheus{prometheusCtx})
	for _, task := range tasks {
		fmt.Println(task)
		if err := task.Run(executor); err != nil {
			return err
		}
	}

	fmt.Println("Remove prometheus service success")
	return nil
}

func SetUpGrafanaService(executor ext.Executor, services *service.Services) error {
	grafanaCtx := GrafanaCtx{
		ctx:     services.Global,
		service: services.Grafana,
	}

	tasks := append([]Task{}, &InitGrafana{grafanaCtx})
	tasks = append(tasks, &SyncGrafanaConfig{grafanaCtx})
	tasks = append(tasks, &StartGrafana{grafanaCtx})
	for _, task := range tasks {
		fmt.Println(task)
		if err := task.Run(executor); err != nil {
			return err
		}
	}

	fmt.Println("Set up grafana service success")
	return nil
}

func RemoveGrafanaService(executor ext.Executor, services *service.Services) error {
	grafanaCtx := GrafanaCtx{
		ctx:     services.Global,
		service: services.Grafana,
	}

	tasks := append([]Task{}, &RemoveGrafana{grafanaCtx})
	for _, task := range tasks {
		fmt.Println(task)
		if err := task.Run(executor); err != nil {
			return err
		}
	}

	fmt.Println("Remove grafana service success")
	return nil
}

func SetUpAlertService(executor ext.Executor, services *service.Services) error {
	alertCtx := AlertManagerCtx{
		ctx:     services.Global,
		service: services.AlertManager,
	}

	tasks := append([]Task{}, &InitAlertManager{alertCtx})
	tasks = append(tasks, &SyncAlertManagerConfig{alertCtx})
	tasks = append(tasks, &StartAlertManager{alertCtx})
	for _, task := range tasks {
		fmt.Println(task)
		if err := task.Run(executor); err != nil {
			return err
		}
	}

	fmt.Println("Set up alertManager service success")
	return nil
}

func RemoveAlertService(executor ext.Executor, services *service.Services) error {
	alertCtx := AlertManagerCtx{
		ctx:     services.Global,
		service: services.AlertManager,
	}

	tasks := append([]Task{}, &RemoveAlertManager{alertCtx})
	for _, task := range tasks {
		fmt.Println(task)
		if err := task.Run(executor); err != nil {
			return err
		}
	}

	fmt.Println("Remove alertManager success")
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
