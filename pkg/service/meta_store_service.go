package service

import (
	"fmt"
	"github.com/hstreamdb/deployment-tool/pkg/executor"
	"github.com/hstreamdb/deployment-tool/pkg/spec"
	"github.com/hstreamdb/deployment-tool/pkg/template/script"
	"github.com/hstreamdb/deployment-tool/pkg/utils"
	log "github.com/sirupsen/logrus"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

type MetaStore struct {
	metaStoreId          uint32
	spec                 spec.MetaStoreSpec
	metaStoreType        spec.MetaStoreType
	ContainerName        string
	CheckReadyScriptPath string
}

func NewMetaStore(id uint32, metaSpec spec.MetaStoreSpec) *MetaStore {
	return &MetaStore{
		metaStoreId:   id,
		spec:          metaSpec,
		metaStoreType: spec.GetMetaStoreType(metaSpec.Image),
		ContainerName: spec.MetaStoreDefaultContainerName,
	}
}

func (m *MetaStore) GetServiceName() string {
	return "meta store"
}

func (m *MetaStore) Display() map[string]utils.DisplayedComponent {
	cfgDir, dataDir := m.getDirs()
	store := utils.DisplayedComponent{
		Name:          "MetaStore",
		Host:          m.spec.Host,
		Ports:         strconv.Itoa(m.spec.Port),
		ContainerName: m.ContainerName,
		Image:         m.spec.Image,
		Paths:         strings.Join([]string{cfgDir, dataDir}, ","),
	}
	return map[string]utils.DisplayedComponent{"metaStore": store}
}

func (m *MetaStore) InitEnv(globalCtx *GlobalCtx) *executor.ExecuteCtx {
	cfgDir, dataDir := m.getDirs()
	args := append([]string{}, "mkdir -p", cfgDir, dataDir, dataDir+"/data", dataDir+"/datalog", cfgDir+"/script", "-m 0775")
	args = append(args, fmt.Sprintf("&& chown -R %[1]s:$(id -gn %[1]s) %[2]s %[3]s", globalCtx.User, cfgDir, dataDir))
	return &executor.ExecuteCtx{Target: m.spec.Host, Cmd: strings.Join(args, " ")}
}

func (m *MetaStore) Deploy(globalCtx *GlobalCtx) *executor.ExecuteCtx {
	mountPoints := []spec.MountPoints{
		{"/mnt", "/mnt"},
		{m.spec.DataDir + "/data", "/data"},
		{m.spec.DataDir + "/datalog", "/datalog"},
	}
	args := spec.GetDockerExecCmd(globalCtx.containerCfg, m.spec.ContainerCfg, spec.MetaStoreDefaultContainerName, true, mountPoints...)
	switch globalCtx.MetaStoreType {
	case spec.ZK:
		args = append(args, zkEnvArgs(m.metaStoreId, globalCtx.MetaStoreUrls, m.spec.Port)...)
		args = append(args, m.spec.Image)
	case spec.RQLITE:
		args = append(args, m.spec.Image, "rqlited")
		args = append(args, fmt.Sprintf("-node-id %d", m.metaStoreId))
		args = append(args, fmt.Sprintf("-http-addr=%s:%d", m.spec.Host, m.spec.Port))
		args = append(args, fmt.Sprintf("-raft-addr=%s:%d", m.spec.Host, m.spec.RaftPort))
		args = append(args, fmt.Sprintf("-bootstrap-expect %d", globalCtx.MetaStoreCount))
		args = append(args, "-join", globalCtx.MetaStoreUrls, "/data")
	}

	return &executor.ExecuteCtx{Target: m.spec.Host, Cmd: strings.Join(args, " ")}
}

func zkEnvArgs(idx uint32, metaStoreUrls string, port int) []string {
	if metaStoreUrls == "" {
		log.Error("metaStoreUrls should not be empty")
		os.Exit(1)
	}

	reg, err := regexp.Compile(fmt.Sprintf(":%d,?", port))
	if err != nil {
		log.Errorf("regexp compile err: %s", err.Error())
		os.Exit(1)
	}
	urls := reg.Split(metaStoreUrls, -1)
	zkUrls := make([]string, 0, len(urls))
	for i, url := range urls {
		if url == "" {
			if i != len(urls)-1 {
				log.Errorf(fmt.Sprintf("invalid metaStoreUrls %s", metaStoreUrls))
				os.Exit(1)
			}
			continue
		}

		// according to https://hub.docker.com/_/zookeeper, ZOO_MY_ID must between 1 and 255
		zkUrls = append(zkUrls, fmt.Sprintf("server.%d=%s:2888:3888;%d", i+1, url, port))
	}

	zooServers := strings.Join(zkUrls, " ")
	return []string{
		fmt.Sprintf("-e ZOO_MY_ID=%d", idx),
		fmt.Sprintf("-e ZOO_SERVERS=\"%s\"", zooServers),
		fmt.Sprintf("-e ZOO_CLIENT_PORT=%d", port),
		"-e ZOO_CFG_EXTRA=\"metricsProvider.className=org.apache.zookeeper.metrics.prometheus.PrometheusMetricsProvider metricsProvider.httpPort=7070\"",
	}
}

func (m *MetaStore) Stop(globalCtx *GlobalCtx) *executor.ExecuteCtx {
	args := []string{"docker rm -f", spec.MetaStoreDefaultContainerName}
	return &executor.ExecuteCtx{Target: m.spec.Host, Cmd: strings.Join(args, " ")}
}

func (m *MetaStore) Remove(globalCtx *GlobalCtx) *executor.ExecuteCtx {
	args := []string{"docker rm -f", spec.MetaStoreDefaultContainerName}
	args = append(args, "&&", "rm -rf", m.spec.DataDir,
		m.spec.RemoteCfgPath)
	return &executor.ExecuteCtx{Target: m.spec.Host, Cmd: strings.Join(args, " ")}
}

func (m *MetaStore) SyncConfig(globalCtx *GlobalCtx) *executor.TransferCtx {
	if m.metaStoreType == spec.RQLITE {
		return nil
	}

	checkReadyScript := script.MetaStoreReadyCheckScript{Host: m.spec.Host, Port: m.spec.Port, Timeout: 600}
	file, err := checkReadyScript.GenScript()
	if err != nil {
		log.Error("gen metastore check ready script error")
		os.Exit(1)
	}

	scriptName := filepath.Base(file)
	cfgDir, _ := m.getDirs()
	remoteScriptPath := filepath.Join(cfgDir, "script", scriptName)
	m.CheckReadyScriptPath = remoteScriptPath
	position := []executor.Position{
		{LocalDir: file, RemoteDir: remoteScriptPath, Opts: fmt.Sprintf("chmod +x %s", remoteScriptPath)},
	}

	if len(globalCtx.LocalMetaStoreConfigFile) != 0 {
		position = append(position, executor.Position{LocalDir: globalCtx.LocalMetaStoreConfigFile, RemoteDir: cfgDir})
	}

	return &executor.TransferCtx{
		Target: m.spec.Host, Position: position,
	}
}

func (m *MetaStore) CheckReady(globalCtx *GlobalCtx) *executor.ExecuteCtx {
	if m.metaStoreType == spec.RQLITE {
		return nil
	}

	if len(m.CheckReadyScriptPath) == 0 {
		log.Error("empty meta store checkReadyScriptPath")
		os.Exit(1)
	}

	args := []string{"/usr/bin/env bash"}
	args = append(args, m.CheckReadyScriptPath)
	return &executor.ExecuteCtx{Target: m.spec.Host, Cmd: strings.Join(args, " ")}
}

func (m *MetaStore) StoreValue(key, value string) *executor.ExecuteCtx {
	if m.metaStoreType != spec.ZK {
		panic("currently only spport store value to zk.")
	}

	args := []string{"docker exec -t"}
	args = append(args, m.ContainerName, "zkCli.sh", fmt.Sprintf("-server localhost:%d", m.spec.Port), "create")
	if !strings.HasPrefix(key, "/") {
		key = "/" + key
	}
	args = append(args, key, fmt.Sprintf("'%s'", value))
	return &executor.ExecuteCtx{Target: m.spec.Host, Cmd: strings.Join(args, " ")}
}

func (m *MetaStore) RemoveThenStore(key, value string) *executor.ExecuteCtx {
	if m.metaStoreType != spec.ZK {
		panic("currently only spport store value to zk.")
	}

	args := []string{"docker exec -t"}
	if !strings.HasPrefix(key, "/") {
		key = "/" + key
	}
	args = append(args, m.ContainerName, fmt.Sprintf("zkCli.sh -server localhost:%d delete  %s || true", m.spec.Port, key))
	args = append(args, "&& docker exec -t", m.ContainerName,
		fmt.Sprintf("zkCli.sh -server localhost:%d create ", m.spec.Port), key, fmt.Sprintf("'%s'", value))
	return &executor.ExecuteCtx{Target: m.spec.Host, Cmd: strings.Join(args, " ")}
}

func (m *MetaStore) GetValue(key string) *executor.ExecuteCtx {
	if m.metaStoreType != spec.ZK {
		panic("currently only spport get value to zk.")
	}

	args := []string{"docker exec -t"}
	args = append(args, m.ContainerName, "zkCli.sh", fmt.Sprintf("-server localhost:%d", m.spec.Port), "get")
	if !strings.HasPrefix(key, "/") {
		key = "/" + key
	}
	args = append(args, key)
	return &executor.ExecuteCtx{Target: m.spec.Host, Cmd: strings.Join(args, " ")}
}

func (m *MetaStore) getDirs() (string, string) {
	return m.spec.RemoteCfgPath, m.spec.DataDir
}
