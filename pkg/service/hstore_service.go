package service

import (
	"fmt"
	"github.com/hstreamdb/deployment-tool/pkg/executor"
	"github.com/hstreamdb/deployment-tool/pkg/spec"
	"github.com/hstreamdb/deployment-tool/pkg/template/script"
	"github.com/hstreamdb/deployment-tool/pkg/utils"
	"path"
	"path/filepath"
	"strconv"
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
	storeId uint32
	spec    spec.HStoreSpec
	Host    string
	// FIXME: check admin port setting
	AdminPort            int
	ContainerName        string
	CheckReadyScriptPath string
	MountScriptPath      string
}

func NewHStore(id uint32, storeSpec spec.HStoreSpec) *HStore {
	return &HStore{
		storeId:       id,
		spec:          storeSpec,
		Host:          storeSpec.Host,
		ContainerName: spec.StoreDefaultContainerName,
		AdminPort:     storeSpec.AdminPort,
	}
}

func (h *HStore) GetServiceName() string {
	return "store"
}

func (h *HStore) Display() map[string]utils.DisplayedComponent {
	cfgDir, dataDir := h.getDirs()
	hstore := utils.DisplayedComponent{
		Name:          "HStore",
		Host:          h.spec.Host,
		Ports:         strconv.Itoa(h.AdminPort),
		ContainerName: h.ContainerName,
		Image:         h.spec.Image,
		Paths:         strings.Join([]string{cfgDir, dataDir}, ","),
	}
	return map[string]utils.DisplayedComponent{"hstore": hstore}
}

func (h *HStore) InitEnv(globalCtx *GlobalCtx) *executor.ExecuteCtx {
	cfgDir, dataDir := h.getDirs()
	args := append([]string{}, "sudo mkdir -p", cfgDir, dataDir, cfgDir+"/script", "/crash", "-m 0775")
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
		{"/crash", "/data/crash"},
		{h.spec.RemoteCfgPath, h.spec.RemoteCfgPath},
	}
	args := spec.GetDockerExecCmd(globalCtx.containerCfg, h.spec.ContainerCfg, h.ContainerName, true, mountPoints...)
	args = append(args, []string{h.spec.Image, spec.StoreDefaultBinPath}...)
	configPath := path.Join(h.spec.RemoteCfgPath, "logdevice.conf")
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
	args = append(args, "&&", "sudo rm -rf",
		fmt.Sprintf("%s/shard*/*", h.spec.DataDir),
		h.spec.DataDir, h.spec.RemoteCfgPath)
	return &executor.ExecuteCtx{Target: h.spec.Host, Cmd: strings.Join(args, " ")}
}

func (h *HStore) SyncConfig(globalCtx *GlobalCtx) *executor.TransferCtx {
	cfgDir, _ := h.getDirs()

	mountScript := script.HStoreMountDiskScript{
		Host:    h.spec.Host,
		Shard:   h.spec.StoreOps.Shards,
		Disk:    h.spec.StoreOps.Disk,
		DataDir: h.spec.DataDir,
	}
	checkReadyScript := script.HStoreReadyCheckScript{
		Host:         h.spec.Host,
		AdminApiPort: h.AdminPort,
		Timeout:      600,
	}

	position, err := h.syncScript(cfgDir, []script.Script{mountScript, checkReadyScript}...)
	if err != nil {
		panic(fmt.Sprintf("gen script error: %s", err))
	}

	if len(globalCtx.HStoreConfigInMetaStore) == 0 {
		position = append(position, executor.Position{LocalDir: globalCtx.LocalHStoreConfigFile, RemoteDir: path.Join(cfgDir, "logdevice.conf")})
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

	args := []string{"/usr/bin/env bash", h.CheckReadyScriptPath}
	return &executor.ExecuteCtx{Target: h.spec.Host, Cmd: strings.Join(args, " ")}
}

func (h *HStore) IsAdmin() bool {
	return h.spec.EnableAdmin
}

func (h *HStore) getDirs() (string, string) {
	return h.spec.RemoteCfgPath, h.spec.DataDir
}

// ============================================================================================

type HAdmin struct {
	storeId              uint32
	spec                 spec.HAdminSpec
	Host                 string
	AdminPort            int
	ContainerName        string
	CheckReadyScriptPath string
}

func NewHAdmin(id uint32, adminSpec spec.HAdminSpec) *HAdmin {
	return &HAdmin{
		storeId:       id,
		spec:          adminSpec,
		Host:          adminSpec.Host,
		ContainerName: spec.AdminDefaultContainerName,
		AdminPort:     adminSpec.AdminPort,
	}
}

func (h *HAdmin) GetServiceName() string {
	return "admin"
}

func (h *HAdmin) Display() map[string]utils.DisplayedComponent {
	cfgDir, dataDir := h.getDirs()
	admin := utils.DisplayedComponent{
		Name:          "HAdmin",
		Host:          h.spec.Host,
		Ports:         strconv.Itoa(h.AdminPort),
		ContainerName: h.ContainerName,
		Image:         h.spec.Image,
		Paths:         strings.Join([]string{cfgDir, dataDir}, ","),
	}
	return map[string]utils.DisplayedComponent{"hadmin": admin}
}

func (h *HAdmin) InitEnv(globalCtx *GlobalCtx) *executor.ExecuteCtx {
	cfgDir, dataDir := h.getDirs()
	args := append([]string{}, "sudo mkdir -p", cfgDir, dataDir, "/crash", "-m 0775")
	return &executor.ExecuteCtx{Target: h.spec.Host, Cmd: strings.Join(args, " ")}
}

func (h *HAdmin) Deploy(globalCtx *GlobalCtx) *executor.ExecuteCtx {
	mountPoints := []spec.MountPoints{
		{h.spec.DataDir, h.spec.DataDir},
		{"/crash", "/data/crash"},
		{h.spec.RemoteCfgPath, h.spec.RemoteCfgPath},
	}
	args := spec.GetDockerExecCmd(globalCtx.containerCfg, h.spec.ContainerCfg, h.ContainerName, true, mountPoints...)
	args = append(args, h.spec.Image, spec.AdminDefaultBinPath)
	configPath := path.Join(h.spec.RemoteCfgPath, "logdevice.conf")
	if len(globalCtx.HStoreConfigInMetaStore) != 0 {
		configPath = globalCtx.HStoreConfigInMetaStore
	}
	args = append(args, fmt.Sprintf("--config-path %s", configPath))
	args = append(args, fmt.Sprintf("--admin-port %d", h.AdminPort))
	args = append(args, "--enable-maintenance-manager",
		"--enable-safety-check-periodic-metadata-update",
		"--maintenance-log-snapshotting",
	)
	return &executor.ExecuteCtx{Target: h.spec.Host, Cmd: strings.Join(args, " ")}
}

func (h *HAdmin) Remove(globalCtx *GlobalCtx) *executor.ExecuteCtx {
	args := []string{"docker rm -f", spec.AdminDefaultContainerName}
	args = append(args, "&&", "sudo rm -rf",
		h.spec.DataDir, h.spec.RemoteCfgPath)
	return &executor.ExecuteCtx{Target: h.spec.Host, Cmd: strings.Join(args, " ")}
}

func (h *HAdmin) SyncConfig(globalCtx *GlobalCtx) *executor.TransferCtx {
	cfgDir, _ := h.getDirs()

	if len(globalCtx.HStoreConfigInMetaStore) == 0 {
		position := append([]executor.Position{},
			executor.Position{LocalDir: globalCtx.LocalHStoreConfigFile, RemoteDir: path.Join(cfgDir, "logdevice.conf")})
		return &executor.TransferCtx{Target: h.spec.Host, Position: position}
	}
	return nil
}

func (h *HAdmin) getDirs() (string, string) {
	return h.spec.RemoteCfgPath, h.spec.DataDir
}

// ==========================================================================================

type AdminInfo struct {
	Host          string
	Port          int
	ContainerName string
}

func Bootstrap(globalCtx *GlobalCtx, adminCtx AdminInfo) *executor.ExecuteCtx {
	args := []string{"docker exec -t"}
	args = append(args, adminCtx.ContainerName, "hadmin store")
	args = append(args, fmt.Sprintf("--port %d", adminCtx.Port))
	args = append(args, BootStrapCmd)
	args = append(args, fmt.Sprintf("node:%d", globalCtx.MetaReplica))
	return &executor.ExecuteCtx{Target: adminCtx.Host, Cmd: strings.Join(args, " ")}
}

func AdminStoreCmd(globalCtx *GlobalCtx, adminCtx AdminInfo, cmd string) *executor.ExecuteCtx {
	args := []string{"docker exec -t"}
	args = append(args, adminCtx.ContainerName, "hadmin store")
	args = append(args, fmt.Sprintf("--port %d", adminCtx.Port))
	args = append(args, cmd)
	return &executor.ExecuteCtx{Target: adminCtx.Host, Cmd: strings.Join(args, " ")}
}

func AdminServerCmd(globalCtx *GlobalCtx, adminCtx AdminInfo, serverHost string,
	serverPort int, cmd string) *executor.ExecuteCtx {
	args := []string{"docker exec -t"}
	args = append(args, adminCtx.ContainerName, "hadmin server")
	args = append(args, "--host", serverHost)
	args = append(args, fmt.Sprintf("--port %d", serverPort))
	args = append(args, cmd)
	return &executor.ExecuteCtx{Target: adminCtx.Host, Cmd: strings.Join(args, " ")}
}
