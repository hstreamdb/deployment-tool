package service

import (
	"fmt"
	"github.com/hstreamdb/dev-deploy/pkg/executor"
	"github.com/hstreamdb/dev-deploy/pkg/spec"
	"github.com/hstreamdb/dev-deploy/pkg/template/script"
	"path/filepath"
	"regexp"
	"strings"
)

const (
	DefaultMetaStoreMonitorPort = 2181
)

type MetaStore struct {
	metaStoreId          uint32
	spec                 spec.MetaStoreSpec
	ContainerName        string
	CheckReadyScriptPath string
}

func NewMetaStore(id uint32, metaSpec spec.MetaStoreSpec) *MetaStore {
	return &MetaStore{metaStoreId: id, spec: metaSpec, ContainerName: spec.MetaStoreDefaultContainerName}
}

func (m *MetaStore) InitEnv(globalCtx *GlobalCtx) *executor.ExecuteCtx {
	cfgDir, dataDir := m.getDirs()
	args := append([]string{}, "sudo mkdir -p", cfgDir, dataDir, cfgDir+"/script")
	return &executor.ExecuteCtx{Target: m.spec.Host, Cmd: strings.Join(args, " ")}
}

func (m *MetaStore) Deploy(globalCtx *GlobalCtx) *executor.ExecuteCtx {
	mountPoints := []spec.MountPoints{
		{"/mnt", "/mnt"},
		{m.spec.DataDir, "/data"},
	}
	args := spec.GetDockerExecCmd(globalCtx.containerCfg, m.spec.ContainerCfg, spec.MetaStoreDefaultContainerName, mountPoints...)
	args = append(args, zkEnvArgs(m.metaStoreId, globalCtx.MetaStoreUrls)...)
	image := m.spec.Image
	if image == "" {
		image = spec.MetaStoreDefaultImage
	}
	args = append(args, image)
	return &executor.ExecuteCtx{Target: m.spec.Host, Cmd: strings.Join(args, " ")}
}

func zkEnvArgs(idx uint32, metaStoreUrls string) []string {
	if metaStoreUrls == "" {
		panic("metaStoreUrls should not be empty")
	}

	reg, err := regexp.Compile(":2181,?")
	if err != nil {
		panic(err)
	}
	urls := reg.Split(metaStoreUrls, -1)

	//urls := strings.Split(metaStoreUrls, ":2181,")
	fmt.Printf("url: %+v\n", urls)
	zkUrls := make([]string, 0, len(urls))
	for i, url := range urls {
		if url == "" {
			if i != len(urls)-1 {
				panic(fmt.Sprintf("invalid metaStoreUrls %s", metaStoreUrls))
			}
			continue
		}

		// according to https://hub.docker.com/_/zookeeper, ZOO_MY_ID must between 1 and 255
		zkUrls = append(zkUrls, fmt.Sprintf("server.%d=%s:2888:3888;2181", i+1, url))
	}
	fmt.Printf("zkurls: %+v\n", zkUrls)
	zooServers := strings.Join(zkUrls, " ")

	return []string{
		fmt.Sprintf("-e ZOO_MY_ID=%d", idx),
		fmt.Sprintf("-e ZOO_SERVERS=\"%s\"", zooServers),
		"-e ZOO_CFG_EXTRA=\"metricsProvider.className=org.apache.zookeeper.metrics.prometheus.PrometheusMetricsProvider metricsProvider.httpPort=7070\"",
	}
}

func (m *MetaStore) Remove(globalCtx *GlobalCtx) *executor.ExecuteCtx {
	args := []string{"docker rm -f", spec.MetaStoreDefaultContainerName}
	args = append(args, "&&", "sudo rm -rf", m.spec.DataDir)
	return &executor.ExecuteCtx{Target: m.spec.Host, Cmd: strings.Join(args, " ")}
}

func (m *MetaStore) SyncConfig(globalCtx *GlobalCtx) *executor.TransferCtx {
	checkReadyScript := script.MetaStoreReadyCheckScript{Host: m.spec.Host, Port: DefaultMetaStoreMonitorPort, Timeout: 20}
	file, err := checkReadyScript.GenScript()
	if err != nil {
		panic("gen script error")
	}

	scriptName := filepath.Base(file)
	cfgDir, _ := m.getDirs()
	remoteScriptPath := filepath.Join(cfgDir, "script", scriptName)
	m.CheckReadyScriptPath = remoteScriptPath
	position := []executor.Position{
		{LocalDir: file, RemoteDir: remoteScriptPath, Opts: fmt.Sprintf("sudo chmod +x %s", remoteScriptPath)},
	}
	if len(m.spec.LocalCfgPath) != 0 {
		position = append(position, executor.Position{LocalDir: m.spec.LocalCfgPath, RemoteDir: cfgDir})
	}

	return &executor.TransferCtx{
		Target: m.spec.Host, Position: position,
	}
}

func (m *MetaStore) CheckReady(globalCtx *GlobalCtx) *executor.ExecuteCtx {
	if len(m.CheckReadyScriptPath) == 0 {
		panic("empty checkReadyScriptPath")
	}

	args := []string{"/bin/bash"}
	args = append(args, m.CheckReadyScriptPath)
	return &executor.ExecuteCtx{Target: m.spec.Host, Cmd: strings.Join(args, " ")}
}

func (m *MetaStore) StoreValue(key, value string) *executor.ExecuteCtx {
	args := []string{"docker exec -t"}
	args = append(args, m.ContainerName, "zkCli.sh", "create")
	if !strings.HasPrefix(key, "/") {
		key = "/" + key
	}
	args = append(args, key, fmt.Sprintf("'%s'", value))
	return &executor.ExecuteCtx{Target: m.spec.Host, Cmd: strings.Join(args, " ")}
}

func (m *MetaStore) GetValue(key string) *executor.ExecuteCtx {
	args := []string{"docker exec -t"}
	args = append(args, m.ContainerName, "zkCli.sh", "get")
	if !strings.HasPrefix(key, "/") {
		key = "/" + key
	}
	args = append(args, key)
	return &executor.ExecuteCtx{Target: m.spec.Host, Cmd: strings.Join(args, " ")}
}

func (m *MetaStore) getDirs() (string, string) {
	if len(m.spec.RemoteCfgPath) == 0 {
		m.spec.RemoteCfgPath = spec.MetaStoreDefaultCfgDir
	}
	if len(m.spec.DataDir) == 0 {
		m.spec.DataDir = spec.MetaStoreDefaultDataDir
	}
	return m.spec.RemoteCfgPath, m.spec.DataDir
}
