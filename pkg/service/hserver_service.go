package service

import (
	"fmt"
	"github.com/hstreamdb/deployment-tool/pkg/executor"
	"github.com/hstreamdb/deployment-tool/pkg/spec"
	"github.com/hstreamdb/deployment-tool/pkg/template/script"
	"github.com/hstreamdb/deployment-tool/pkg/utils"
	"path"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

const (
	DefaultServerMonitorPort = 6570
)

type HServer struct {
	serverId             uint32
	spec                 spec.HServerSpec
	Host                 string
	Port                 int
	ContainerName        string
	CheckReadyScriptPath string
	ServerConfigPath     string
	StoreConfigPath      string
}

func NewHServer(id uint32, serverSpec spec.HServerSpec) *HServer {
	return &HServer{
		serverId:      id,
		spec:          serverSpec,
		Host:          serverSpec.Host,
		Port:          serverSpec.Port,
		ContainerName: spec.ServerDefaultContainerName,
	}
}

func (h *HServer) GetServiceName() string {
	return "server"
}

func (h *HServer) Display() map[string]utils.DisplayedComponent {
	cfgDir, dataDir := h.getDirs()
	hserver := utils.DisplayedComponent{
		Name:          "HServer",
		Host:          h.spec.Host,
		Ports:         strconv.Itoa(h.spec.Port),
		ContainerName: h.ContainerName,
		Image:         h.spec.Image,
		Paths:         strings.Join([]string{cfgDir, dataDir}, ","),
	}
	return map[string]utils.DisplayedComponent{"hserver": hserver}
}

func (h *HServer) InitEnv(globalCtx *GlobalCtx) *executor.ExecuteCtx {
	cfgDir, dataDir := h.getDirs()
	args := append([]string{}, "sudo mkdir -p", cfgDir, dataDir, cfgDir+"/script", "/data/crash", "-m 0775")
	args = append(args, fmt.Sprintf("&& sudo chown -R %[1]s:$(id -gn %[1]s) %[2]s %[3]s /data/crash",
		globalCtx.User, cfgDir, dataDir))
	return &executor.ExecuteCtx{Target: h.spec.Host, Cmd: strings.Join(args, " ")}
}

func (h *HServer) Deploy(globalCtx *GlobalCtx) *executor.ExecuteCtx {
	mountPoints := []spec.MountPoints{
		{"/mnt", "/mnt"},
		{h.spec.DataDir, h.spec.DataDir},
		{"/data/crash", "/data/crash"},
		{h.spec.RemoteCfgPath, h.spec.RemoteCfgPath},
	}

	args := spec.GetDockerExecCmd(globalCtx.containerCfg, h.spec.ContainerCfg, h.ContainerName, true, mountPoints...)
	serverBinPath := spec.ServerDefaultBinPath
	if globalCtx.EnableHsGrpc {
		serverBinPath = spec.ServerDefaultHsGrpcBinPath
	}
	args = append(args, []string{h.spec.Image, serverBinPath}...)
	_, version := parseImage(h.spec.Image)
	if utils.CompareVersion(version, utils.Version0101) > 0 {
		args = append(args, "--bind-address", h.spec.Host)
		args = append(args, "--advertised-address", h.spec.AdvertisedAddress)
	} else {
		args = append(args, "--host", h.spec.Host)
		args = append(args, "--address", h.spec.AdvertisedAddress)
		args = append(args, "--compression", h.spec.Opts.Compression)
	}

	args = append(args, fmt.Sprintf("--port %d", h.spec.Port))
	args = append(args, fmt.Sprintf("--internal-port %d", h.spec.InternalPort))
	if len(h.ServerConfigPath) != 0 {
		args = append(args, "--config-path", h.ServerConfigPath)
	}

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

	if len(h.StoreConfigPath) != 0 {
		args = append(args, "--store-config", h.StoreConfigPath)
	} else {
		args = append(args, "--store-config", globalCtx.HStoreConfigInMetaStore)
	}

	args = append(args, fmt.Sprintf("--server-id %d", h.serverId))
	args = append(args, "--store-log-level", h.spec.Opts.StoreLogLevel)
	args = append(args, "--log-level", h.spec.Opts.ServerLogLevel)
	args = append(args, fmt.Sprintf("--checkpoint-replica %d", globalCtx.MetaReplica))
	admin := globalCtx.HAdminInfos[0]
	args = append(args, fmt.Sprintf("--store-admin-host %s --store-admin-port %d", admin.Host, admin.Port))
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
		panic(fmt.Sprintf("gen script error: %s\n", err.Error()))
	}

	scriptName := filepath.Base(file)
	cfgDir, _ := h.getDirs()
	remoteScriptPath := filepath.Join(cfgDir, "script", scriptName)
	h.CheckReadyScriptPath = remoteScriptPath
	position := []executor.Position{
		{LocalDir: file, RemoteDir: remoteScriptPath, Opts: fmt.Sprintf("sudo chmod +x %s", remoteScriptPath)},
	}
	if len(globalCtx.LocalHServerConfigFile) != 0 {
		serverPath := path.Join(cfgDir, "config.yaml")
		position = append(position, executor.Position{LocalDir: globalCtx.LocalHServerConfigFile, RemoteDir: serverPath})
		h.ServerConfigPath = serverPath
	}
	if len(globalCtx.HStoreConfigInMetaStore) == 0 {
		storePath := path.Join(cfgDir, "logdevice.conf")
		position = append(position, executor.Position{LocalDir: globalCtx.LocalHStoreConfigFile, RemoteDir: storePath})
		h.StoreConfigPath = storePath
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

	args := []string{"/usr/bin/env bash"}
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
		urls := strings.ReplaceAll(url, "http://", "")
		finalUrl := strings.Split(urls, ",")[0]
		return "rq://" + finalUrl
	case spec.Unknown:
		return ""
	}
	return ""
}

func needSeedNodes(version utils.Version) bool {
	return utils.CompareVersion(version, utils.Version082) > 0 && utils.CompareVersion(version, utils.Version084) != 0
}

func parseImage(imageStr string) (string, utils.Version) {
	reg := regexp.MustCompile(".*[:v]?\\d{1,3}.\\d{1,3}.\\d{1,3}")
	if !reg.MatchString(imageStr) {
		return imageStr, utils.Version{IsLatest: true}
	}

	fragment := strings.Split(imageStr, ":")
	image, version := fragment[0], fragment[1]
	return image, utils.CreateVersion(version)
}
