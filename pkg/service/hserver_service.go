package service

import (
	"fmt"
	"github.com/hstreamdb/dev-deploy/pkg/executor"
	"github.com/hstreamdb/dev-deploy/pkg/spec"
	"github.com/hstreamdb/dev-deploy/pkg/template/script"
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
	if len(h.spec.LocalCfgPath) != 0 {
		configPath, _ = h.getDirs()
		mountPoints = append(mountPoints, spec.MountPoints{Local: configPath, Remote: spec.ServerBinConfigPath})
	}
	args := spec.GetDockerExecCmd(globalCtx.containerCfg, h.spec.ContainerCfg, spec.ServerDefaultContainerName, mountPoints...)
	image := h.spec.Image
	if image == "" {
		image = spec.ServerDefaultImage
	}
	args = append(args, []string{image, spec.ServerDefaultBinPath}...)
	args = append(args, fmt.Sprintf("--host %s", h.spec.Host))
	args = append(args, fmt.Sprintf("--port %d", h.spec.Port))
	address := h.spec.Address
	if len(address) == 0 {
		address = h.spec.Host
	}
	args = append(args, fmt.Sprintf("--address %s", address))
	args = append(args, fmt.Sprintf("--internal-port %d", h.spec.InternalPort))
	args = append(args, "--seed-nodes", globalCtx.SeedNodes)
	if len(configPath) != 0 {
		args = append(args, fmt.Sprintf("--config-path %s", configPath))
	}
	args = append(args, fmt.Sprintf("--zkuri %s", globalCtx.MetaStoreUrls))
	args = append(args, fmt.Sprintf("--store-config %s", globalCtx.HStoreConfigInMetaStore))
	args = append(args, fmt.Sprintf("--server-id %d", h.serverId))
	args = append(args, fmt.Sprintf("--store-log-level %s", h.spec.Opts.StoreLogLevel))
	args = append(args, fmt.Sprintf("--log-level %s", h.spec.Opts.ServerLogLevel))
	args = append(args, fmt.Sprintf("--compression %s", h.spec.Opts.Compression))
	admin := globalCtx.HadminAddress[0]
	adminInfo := strings.Split(admin, ":")
	args = append(args, fmt.Sprintf("--store-admin-host %s --store-admin-port %s", adminInfo[0], adminInfo[1]))
	return &executor.ExecuteCtx{Target: h.spec.Host, Cmd: strings.Join(args, " ")}
}

func (h *HServer) Remove(globalCtx *GlobalCtx) *executor.ExecuteCtx {
	args := []string{"docker rm -f", spec.ServerDefaultContainerName}
	return &executor.ExecuteCtx{Target: h.spec.Host, Cmd: strings.Join(args, " ")}
}

func (h *HServer) SyncConfig(globalCtx *GlobalCtx) *executor.TransferCtx {
	checkReadyScript := script.HServerReadyCheckScript{Host: h.spec.Host, Port: DefaultServerMonitorPort, Timeout: 20}
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
	if len(h.spec.LocalCfgPath) != 0 {
		position = append(position, executor.Position{LocalDir: h.spec.LocalCfgPath, RemoteDir: cfgDir})
	}

	return &executor.TransferCtx{
		Target: h.spec.Host, Position: position,
	}
}

func (h *HServer) Init(ctx *GlobalCtx) *executor.ExecuteCtx {
	args := []string{"docker exec -t", spec.ServerDefaultContainerName}
	args = append(args, "/usr/local/bin/hadmin", "server", "--host", h.spec.Host, "init")
	return &executor.ExecuteCtx{Target: h.spec.Host, Cmd: strings.Join(args, " ")}
}

func (h *HServer) CheckReady(globalCtx *GlobalCtx) *executor.ExecuteCtx {
	if len(h.CheckReadyScriptPath) == 0 {
		panic("empty checkReadyScriptPath")
	}

	args := []string{"/bin/bash"}
	args = append(args, h.CheckReadyScriptPath)
	return &executor.ExecuteCtx{Target: h.spec.Host, Cmd: strings.Join(args, " ")}
}

func (m *HServer) getDirs() (string, string) {
	if len(m.spec.RemoteCfgPath) == 0 {
		m.spec.RemoteCfgPath = spec.ServerDefaultConfigPath
	}
	if len(m.spec.DataDir) == 0 {
		m.spec.DataDir = spec.ServerDefaultDataDir
	}
	return m.spec.RemoteCfgPath, m.spec.DataDir
}
