package service

import (
	"fmt"
	"github.com/hstreamdb/deployment-tool/pkg/executor"
	"github.com/hstreamdb/deployment-tool/pkg/spec"
	"github.com/hstreamdb/deployment-tool/pkg/template/script"
	"github.com/hstreamdb/deployment-tool/pkg/utils"
	log "github.com/sirupsen/logrus"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
)

const (
	DefaultKafkaServerMonitorPort = 6570
)

type HServerKafka struct {
	serverId             uint32
	spec                 spec.HServerKafkaSpec
	Host                 string
	Port                 int
	ContainerName        string
	CheckReadyScriptPath string
	ServerConfigPath     string
	StoreConfigPath      string
}

func NewHServerKafka(id uint32, serverSpec spec.HServerKafkaSpec) *HServerKafka {
	return &HServerKafka{
		serverId:      id,
		spec:          serverSpec,
		Host:          serverSpec.Host,
		Port:          serverSpec.Port,
		ContainerName: spec.KafkaServerDefaultContainerName,
	}
}

func (h *HServerKafka) GetServiceName() string {
	return "kafkaServer"
}

func (h *HServerKafka) Display() map[string]utils.DisplayedComponent {
	cfgDir, dataDir := h.getDirs()
	HServerKafka := utils.DisplayedComponent{
		Name:          "HServerKafka",
		Host:          h.spec.Host,
		Ports:         strconv.Itoa(h.spec.Port),
		ContainerName: h.ContainerName,
		Image:         h.spec.Image,
		Paths:         strings.Join([]string{cfgDir, dataDir}, ","),
	}
	return map[string]utils.DisplayedComponent{"hserverKafka": HServerKafka}
}

func (h *HServerKafka) InitEnv(globalCtx *GlobalCtx) *executor.ExecuteCtx {
	cfgDir, dataDir := h.getDirs()
	args := append([]string{}, "mkdir -p", cfgDir, dataDir, cfgDir+"/script", "/data/crash", "-m 0775")
	args = append(args, fmt.Sprintf("&& chown -R %[1]s:$(id -gn %[1]s) %[2]s %[3]s /data/crash",
		globalCtx.User, cfgDir, dataDir))
	return &executor.ExecuteCtx{Target: h.spec.Host, Cmd: strings.Join(args, " ")}
}

func (h *HServerKafka) Deploy(globalCtx *GlobalCtx) *executor.ExecuteCtx {
	mountPoints := []spec.MountPoints{
		{"/mnt", "/mnt"},
		{"/var/run/docker.sock", "/var/run/docker.sock"},
		{"/tmp", "/tmp"},
		{h.spec.DataDir, h.spec.DataDir},
		{"/data/crash", "/data/crash"},
		{h.spec.RemoteCfgPath, h.spec.RemoteCfgPath},
	}

	args := spec.GetDockerExecCmd(globalCtx.containerCfg, h.spec.ContainerCfg, h.ContainerName, true, mountPoints...)
	serverBinPath := spec.KafkaServerDefaultBinPath
	if globalCtx.EnableGrpcHs {
		log.Errorln("grpc-hs is not supported in kafka server")
		os.Exit(1)
	}
	args = append(args, []string{h.spec.Image, serverBinPath}...)
	args = append(args, "--bind-address", "0.0.0.0")
	args = append(args, "--advertised-address", h.spec.Host)
	if len(h.spec.AdvertisedListener) != 0 {
		args = append(args, "--advertised-listeners", h.spec.AdvertisedListener)
	}

	args = append(args, fmt.Sprintf("--port %d", h.spec.Port))
	if len(h.ServerConfigPath) != 0 {
		args = append(args, "--config-path", h.ServerConfigPath)
	}

	if h.spec.Opts == nil {
		h.spec.Opts = make(map[string]string)
	}

	//if needSeedNodes(version) {
	//	if _, ok := h.spec.Opts["seed-nodes"]; !ok {
	//		h.spec.Opts["seed-nodes"] = globalCtx.SeedNodes
	//	}
	//}

	metaStoreUrl := getMetaStoreUrl(globalCtx.MetaStoreType, globalCtx.MetaStoreUrls)
	h.spec.Opts["metastore-uri"] = metaStoreUrl

	if len(h.StoreConfigPath) != 0 {
		h.spec.Opts["store-config"] = h.StoreConfigPath
	} else {
		h.spec.Opts["store-config"] = globalCtx.HStoreConfigInMetaStore
	}

	if _, ok := h.spec.Opts["server-id"]; !ok {
		h.spec.Opts["server-id"] = strconv.Itoa(int(h.serverId))
	}
	for k, v := range h.spec.Opts {
		if k == "enable-tls" {
			args = append(args, "--enable-tls")
		} else {
			args = append(args, fmt.Sprintf("--%s %s", k, v))
		}
	}
	return &executor.ExecuteCtx{Target: h.spec.Host, Cmd: strings.Join(args, " ")}
}

func (h *HServerKafka) Stop(globalCtx *GlobalCtx) *executor.ExecuteCtx {
	args := []string{"docker rm -f", spec.KafkaServerDefaultContainerName}
	return &executor.ExecuteCtx{Target: h.spec.Host, Cmd: strings.Join(args, " ")}
}

func (h *HServerKafka) Remove(globalCtx *GlobalCtx) *executor.ExecuteCtx {
	args := []string{"docker rm -f", spec.KafkaServerDefaultContainerName, "&& rm -rf",
		h.spec.DataDir, h.spec.RemoteCfgPath}
	return &executor.ExecuteCtx{Target: h.spec.Host, Cmd: strings.Join(args, " ")}
}

func (h *HServerKafka) SyncConfig(globalCtx *GlobalCtx) *executor.TransferCtx {
	checkReadyScript := script.HServerReadyCheckScript{Host: h.spec.Host, Port: h.spec.Port, Timeout: 600}
	file, err := checkReadyScript.GenScript()
	if err != nil {
		log.Errorf("gen script error: %s", err.Error())
		os.Exit(1)
	}

	scriptName := filepath.Base(file)
	cfgDir, _ := h.getDirs()
	remoteScriptPath := filepath.Join(cfgDir, "script", scriptName)
	h.CheckReadyScriptPath = remoteScriptPath
	position := []executor.Position{
		{LocalDir: file, RemoteDir: remoteScriptPath, Opts: fmt.Sprintf("chmod +x %s", remoteScriptPath)},
	}
	if len(globalCtx.LocalHServerKafkaConfigFile) != 0 {
		serverPath := path.Join(cfgDir, "config.yaml")
		position = append(position, executor.Position{LocalDir: globalCtx.LocalHServerKafkaConfigFile, RemoteDir: serverPath})
		h.ServerConfigPath = serverPath
	}
	if len(globalCtx.HStoreConfigInMetaStore) == 0 {
		storePath := path.Join(cfgDir, "logdevice.conf")
		position = append(position, executor.Position{LocalDir: globalCtx.LocalHStoreConfigFile, RemoteDir: storePath})
		h.StoreConfigPath = storePath
	}
	if _, ok := h.spec.Opts["tls-key-path"]; ok {
		tlsKeyPath := path.Join(cfgDir, path.Base(h.spec.Opts["tls-key-path"]))
		position = append(position, executor.Position{LocalDir: h.spec.Opts["tls-key-path"], RemoteDir: tlsKeyPath})
		h.spec.Opts["tls-key-path"] = tlsKeyPath
	}
	if _, ok := h.spec.Opts["tls-cert-path"]; ok {
		certPath := path.Join(cfgDir, path.Base(h.spec.Opts["tls-cert-path"]))
		position = append(position, executor.Position{LocalDir: h.spec.Opts["tls-cert-path"], RemoteDir: certPath})
		h.spec.Opts["tls-cert-path"] = certPath
	}
	if _, ok := h.spec.Opts["tls-ca-path"]; ok {
		caPath := path.Join(cfgDir, path.Base(h.spec.Opts["tls-ca-path"]))
		position = append(position, executor.Position{LocalDir: h.spec.Opts["tls-ca-path"], RemoteDir: caPath})
		h.spec.Opts["tls-ca-path"] = caPath
	}

	return &executor.TransferCtx{
		Target: h.spec.Host, Position: position,
	}
}

func (h *HServerKafka) Init(ctx *GlobalCtx) *executor.ExecuteCtx {
	args := []string{"docker exec -t", spec.KafkaServerDefaultContainerName}
	if _, ok := h.spec.Opts["enable-tls"]; ok {
		if _, ok = h.spec.Opts["tls-ca-path"]; !ok {
			log.Errorf("tls-ca-path should not be empty when enable tls set.")
			os.Exit(1)
		}

		args = append(args, "/usr/local/bin/hstream", "--host", h.spec.Host)
		args = append(args, "--port", fmt.Sprintf("%d", h.spec.Port))
		args = append(args, "--tls-ca", h.spec.Opts["tls-ca-path"], "init")
	} else {
		args = append(args, "/usr/local/bin/hstream", "--host", h.spec.Host)
		args = append(args, "--port", fmt.Sprintf("%d", h.spec.Port), "init")
	}
	return &executor.ExecuteCtx{Target: h.spec.Host, Cmd: strings.Join(args, " ")}
}

func (h *HServerKafka) CheckReady(globalCtx *GlobalCtx) *executor.ExecuteCtx {
	if len(h.CheckReadyScriptPath) == 0 {
		log.Error("empty HServerKafka checkReadyScriptPath")
		os.Exit(1)
	}

	args := []string{"/usr/bin/env bash"}
	args = append(args, h.CheckReadyScriptPath)
	return &executor.ExecuteCtx{Target: h.spec.Host, Cmd: strings.Join(args, " ")}
}

func (h *HServerKafka) GetStatus(globalCtx *GlobalCtx) *executor.ExecuteCtx {
	return nil
}

func (h *HServerKafka) GetHost() string {
	return h.spec.Host
}

func (h *HServerKafka) getDirs() (string, string) {
	return h.spec.RemoteCfgPath, h.spec.DataDir
}
