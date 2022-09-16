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
	// DefaultAdminApiPort TCP port on which the server listens to for admin commands, supports commands over SSL
	DefaultAdminApiPort = 6440
	// DefaultServerListenPort TCP port on which the server listens for non-SSL clients
	DefaultServerListenPort = 16111
	BootStrapCmd            = "nodes-config bootstrap --metadata-replicate-across "
)

type HStore struct {
	storeId              uint32
	spec                 spec.HStoreSpec
	ContainerName        string
	CheckReadyScriptPath string
	MountScriptPath      string
}

func NewHStore(id uint32, storeSpec spec.HStoreSpec) *HStore {
	return &HStore{storeId: id, spec: storeSpec, ContainerName: spec.StoreDefaultContainerName}
}

func (h *HStore) InitEnv(globalCtx *GlobalCtx) *executor.ExecuteCtx {
	cfgDir, dataDir := h.getDirs()
	args := append([]string{}, "sudo mkdir -p", cfgDir, dataDir, cfgDir+"/script")
	args = append(args, fmt.Sprintf("&& echo %d | tee %s", h.spec.StoreOps.Shards, filepath.Join(dataDir, "NSHARDS")))
	return &executor.ExecuteCtx{Target: h.spec.Host, Cmd: strings.Join(args, " ")}
}

func (h *HStore) MountDisk() *executor.ExecuteCtx {
	args := []string{"/bin/bash", h.MountScriptPath}
	return &executor.ExecuteCtx{Target: h.spec.Host, Cmd: strings.Join(args, " ")}
}

func (h *HStore) Deploy(globalCtx *GlobalCtx) *executor.ExecuteCtx {
	mountPoints := []spec.MountPoints{
		{"/mnt", "/mnt"},
		{h.spec.DataDir, h.spec.DataDir},
	}
	args := spec.GetDockerExecCmd(globalCtx.containerCfg, h.spec.ContainerCfg, spec.StoreDefaultContainerName, mountPoints...)
	args = append(args, []string{h.spec.Image, spec.StoreDefaultBinPath}...)
	configPath := h.spec.RemoteCfgPath
	if len(globalCtx.HStoreConfigInMetaStore) != 0 {
		configPath = globalCtx.HStoreConfigInMetaStore
	}
	args = append(args, fmt.Sprintf("--config-path %s", configPath))
	args = append(args, fmt.Sprintf("--name ld_%d", h.storeId))
	args = append(args, fmt.Sprintf("--address %s", h.spec.Host))
	args = append(args, fmt.Sprintf("--local-log-store-path %s", h.spec.DataDir))
	args = append(args, fmt.Sprintf("--num-shards %d", h.spec.StoreOps.Shards))
	var role string
	switch h.spec.Role {
	case "Both", "both":
		role = "storage,sequencer"
	case "Storage", "storage":
		role = "storage"
	case "Sequencer", "sequencer":
		role = "sequencer"
	}
	args = append(args, fmt.Sprintf("--roles %s", role))
	if h.spec.EnableAdmin {
		args = append(args, "--enable-maintenance-manager",
			"--enable-safety-check-periodic-metadata-update",
			"--maintenance-log-snapshotting",
		)
	}
	return &executor.ExecuteCtx{Target: h.spec.Host, Cmd: strings.Join(args, " ")}
}

func (h *HStore) Remove(globalCtx *GlobalCtx) *executor.ExecuteCtx {
	args := []string{"docker rm -f", spec.StoreDefaultContainerName}
	args = append(args, "&&", "sudo rm -rf", fmt.Sprintf("%s/shard*/*", h.spec.DataDir), h.spec.DataDir)
	return &executor.ExecuteCtx{Target: h.spec.Host, Cmd: strings.Join(args, " ")}
}

func (h *HStore) SyncConfig(globalCtx *GlobalCtx) *executor.TransferCtx {
	cfgDir, _ := h.getDirs()

	checkReadyScript := script.HStoreReadyCheckScript{
		Host:             h.spec.Host,
		AdminApiPort:     DefaultAdminApiPort,
		ServerListenPort: DefaultServerListenPort,
		Timeout:          20,
	}
	mountScript := script.HStoreMountDiskScript{
		Host:    h.spec.Host,
		Shard:   h.spec.StoreOps.Shards,
		Disk:    h.spec.StoreOps.Disk,
		DataDir: h.spec.DataDir,
	}
	position, err := h.syncScript(cfgDir, []script.Script{checkReadyScript, mountScript}...)
	if err != nil {
		panic("gen script error")
	}

	if len(globalCtx.HStoreConfigInMetaStore) == 0 {
		position = append(position, executor.Position{LocalDir: globalCtx.LocalHStoreConfigFile, RemoteDir: cfgDir})
	}

	return &executor.TransferCtx{Target: h.spec.Host, Position: position}
}

func (h *HStore) syncScript(cfgDir string, scpts ...script.Script) ([]executor.Position, error) {
	res := make([]executor.Position, 0, len(scpts))
	for _, scpt := range scpts {
		file, err := scpt.GenScript()
		if err != nil {
			return nil, err
		}
		scriptName := filepath.Base(file)
		remoteScriptPath := filepath.Join(cfgDir, "script", scriptName)
		switch scpt.(type) {
		case script.HStoreReadyCheckScript:
			h.CheckReadyScriptPath = remoteScriptPath
		case script.HStoreMountDiskScript:
			h.MountScriptPath = remoteScriptPath
		}
		res = append(res, executor.Position{LocalDir: file, RemoteDir: remoteScriptPath, Opts: fmt.Sprintf("sudo chmod +x %s", remoteScriptPath)})
	}
	return res, nil
}

func (h *HStore) CheckReady(globalCtx *GlobalCtx) *executor.ExecuteCtx {
	if len(h.CheckReadyScriptPath) == 0 {
		panic("empty checkReadyScriptPath")
	}

	args := []string{"/bin/bash"}
	args = append(args, h.CheckReadyScriptPath)
	return &executor.ExecuteCtx{Target: h.spec.Host, Cmd: strings.Join(args, " ")}
}

func (h *HStore) Bootstrap(globalCtx *GlobalCtx) *executor.ExecuteCtx {
	args := []string{"docker exec -t"}
	args = append(args, h.ContainerName, "hadmin store")
	args = append(args, fmt.Sprintf("--port %d", DefaultAdminApiPort))
	args = append(args, BootStrapCmd)
	args = append(args, fmt.Sprintf("node:%d", globalCtx.MetaReplica))
	return &executor.ExecuteCtx{Target: h.spec.Host, Cmd: strings.Join(args, " ")}
}

func (h *HStore) AdminStoreCmd(globalCtx *GlobalCtx, cmd string) *executor.ExecuteCtx {
	args := []string{"docker exec -t"}
	args = append(args, h.ContainerName, "hadmin store")
	args = append(args, fmt.Sprintf("--port %d", DefaultAdminApiPort))
	args = append(args, cmd)
	return &executor.ExecuteCtx{Target: h.spec.Host, Cmd: strings.Join(args, " ")}
}

func (h *HStore) AdminServerCmd(globalCtx *GlobalCtx, host string, cmd string) *executor.ExecuteCtx {
	args := []string{"docker exec -t"}
	args = append(args, h.ContainerName, "hadmin server")
	args = append(args, "--host", host, cmd)
	return &executor.ExecuteCtx{Target: h.spec.Host, Cmd: strings.Join(args, " ")}
}

func (h *HStore) IsAdmin() bool {
	return h.spec.EnableAdmin
}

func (h *HStore) getDirs() (string, string) {
	return h.spec.RemoteCfgPath, h.spec.DataDir
}
