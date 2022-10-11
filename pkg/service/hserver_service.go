package service

import (
	"fmt"
	"github.com/hstreamdb/deployment-tool/pkg/executor"
	"github.com/hstreamdb/deployment-tool/pkg/spec"
	"github.com/hstreamdb/deployment-tool/pkg/template/script"
	"github.com/hstreamdb/deployment-tool/pkg/utils"
	"path/filepath"
	"strings"
)

const (
	DefaultServerMonitorPort = 6570
)

type HServer struct {
	serverId             uint32
	spec                 spec.HServerSpec
	CheckReadyScriptPath string
}

func NewHServer(id uint32, serverSpec spec.HServerSpec) *HServer {
	return &HServer{serverId: id, spec: serverSpec}
}

func (h *HServer) InitEnv(globalCtx *GlobalCtx) *executor.ExecuteCtx {
	cfgDir, dataDir := h.getDirs()
	args := append([]string{}, "sudo mkdir -p", cfgDir, dataDir, cfgDir+"/script")
	return &executor.ExecuteCtx{Target: h.spec.Host, Cmd: strings.Join(args, " ")}
}

func (h *HServer) Deploy(globalCtx *GlobalCtx) *executor.ExecuteCtx {
	var (
		mountPoints []spec.MountPoints
		configPath  string
	)
	if len(globalCtx.LocalHServerConfigFile) != 0 {
		configPath, _ = h.getDirs()
		mountPoints = append(mountPoints, spec.MountPoints{Local: configPath, Remote: spec.ServerBinConfigPath})
	}
	args := spec.GetDockerExecCmd(globalCtx.containerCfg, h.spec.ContainerCfg, spec.ServerDefaultContainerName, true, mountPoints...)
	args = append(args, []string{h.spec.Image, spec.ServerDefaultBinPath}...)
	args = append(args, "--host", h.spec.Host)
	args = append(args, fmt.Sprintf("--port %d", h.spec.Port))
	args = append(args, "--address", h.spec.Address)
	args = append(args, fmt.Sprintf("--internal-port %d", h.spec.InternalPort))
	if len(configPath) != 0 {
		args = append(args, "--config-path", configPath)
	}

	_, version := parseImage(h.spec.Image)
	if needSeedNodes(version) {
		args = append(args, "--seed-nodes", globalCtx.SeedNodes)
	}

	if utils.CompareVersion(version, utils.Version096) > 0 {
		metaStoreUrl := getMetaStoreUrl(globalCtx.MetaStoreType, globalCtx.MetaStoreUrls)
		args = append(args, "--metastore-uri", metaStoreUrl)
	} else if utils.CompareVersion(version, utils.Version095) > 0 {
		metaStoreUrl := getMetaStoreUrl(globalCtx.MetaStoreType, globalCtx.MetaStoreUrls)
		args = append(args, "--meta-store", metaStoreUrl)
	} else {
		args = append(args, "--zkuri", globalCtx.MetaStoreUrls)
	}

	args = append(args, "--store-config", globalCtx.HStoreConfigInMetaStore)
	args = append(args, fmt.Sprintf("--server-id %d", h.serverId))
	args = append(args, "--store-log-level", h.spec.Opts.StoreLogLevel)
	args = append(args, "--log-level", h.spec.Opts.ServerLogLevel)
	args = append(args, "--compression", h.spec.Opts.Compression)
	admin := globalCtx.HadminAddress[0]
	adminInfo := strings.Split(admin, ":")
	args = append(args, fmt.Sprintf("--store-admin-host %s --store-admin-port %s", adminInfo[0], adminInfo[1]))
	return &executor.ExecuteCtx{Target: h.spec.Host, Cmd: strings.Join(args, " ")}
}

func (h *HServer) Remove(globalCtx *GlobalCtx) *executor.ExecuteCtx {
	args := []string{"docker rm -f", spec.ServerDefaultContainerName, "&& sudo rm -rf",
		h.spec.DataDir, h.spec.RemoteCfgPath}
	return &executor.ExecuteCtx{Target: h.spec.Host, Cmd: strings.Join(args, " ")}
}

func (h *HServer) SyncConfig(globalCtx *GlobalCtx) *executor.TransferCtx {
	checkReadyScript := script.HServerReadyCheckScript{Host: h.spec.Host, Port: DefaultServerMonitorPort, Timeout: 600}
	file, err := checkReadyScript.GenScript()
	if err != nil {
		panic("gen script error")
	}

	scriptName := filepath.Base(file)
	cfgDir, _ := h.getDirs()
	remoteScriptPath := filepath.Join(cfgDir, "script", scriptName)
	h.CheckReadyScriptPath = remoteScriptPath
	position := []executor.Position{
		{LocalDir: file, RemoteDir: remoteScriptPath, Opts: fmt.Sprintf("sudo chmod +x %s", remoteScriptPath)},
	}
	if len(globalCtx.LocalHServerConfigFile) != 0 {
		position = append(position, executor.Position{LocalDir: globalCtx.LocalHServerConfigFile, RemoteDir: cfgDir})
	}

	return &executor.TransferCtx{
		Target: h.spec.Host, Position: position,
	}
}

func (h *HServer) Init(ctx *GlobalCtx) *executor.ExecuteCtx {
	_, version := parseImage(h.spec.Image)
	if utils.CompareVersion(version, utils.Version090) >= 0 {
		args := []string{"docker exec -t", spec.ServerDefaultContainerName}
		args = append(args, "/usr/local/bin/hadmin", "server", "--host", h.spec.Host, "init")
		return &executor.ExecuteCtx{Target: h.spec.Host, Cmd: strings.Join(args, " ")}
	}
	return nil
}

func (h *HServer) CheckReady(globalCtx *GlobalCtx) *executor.ExecuteCtx {
	if len(h.CheckReadyScriptPath) == 0 {
		panic("empty checkReadyScriptPath")
	}

	args := []string{"/bin/bash"}
	args = append(args, h.CheckReadyScriptPath)
	return &executor.ExecuteCtx{Target: h.spec.Host, Cmd: strings.Join(args, " ")}
}

func (h *HServer) GetHost() string {
	return h.spec.Host
}

func (h *HServer) getDirs() (string, string) {
	return h.spec.RemoteCfgPath, h.spec.DataDir
}

func getMetaStoreUrl(tp spec.MetaStoreType, url string) string {
	switch tp {
	case spec.ZK:
		return "zk://" + url
	case spec.RQLITE:
		return "rq://" + url
	case spec.Unknown:
		return ""
	}
	return ""
}

func needSeedNodes(version utils.Version) bool {
	return utils.CompareVersion(version, utils.Version082) > 0 && utils.CompareVersion(version, utils.Version084) != 0
}

func parseImage(imageStr string) (string, utils.Version) {
	if !strings.Contains(imageStr, ":") {
		return imageStr, utils.Version{IsLatest: true}
	}

	fragment := strings.Split(imageStr, ":")
	image, version := fragment[0], fragment[1]
	return image, utils.CreateVersion(version)
}
