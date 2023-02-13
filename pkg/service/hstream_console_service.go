package service

import (
	"fmt"
	"github.com/hstreamdb/deployment-tool/pkg/executor"
	"github.com/hstreamdb/deployment-tool/pkg/spec"
	"github.com/hstreamdb/deployment-tool/pkg/template/config"
	"github.com/hstreamdb/deployment-tool/pkg/utils"
	log "github.com/sirupsen/logrus"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type HStreamConsole struct {
	spec          spec.HStreamConsoleSpec
	ContainerName string
}

func NewHStreamConsole(id uint32, consoleSPec spec.HStreamConsoleSpec) *HStreamConsole {
	return &HStreamConsole{spec: consoleSPec, ContainerName: spec.ConsoleDefaultContainerName}
}

func (h *HStreamConsole) GetServiceName() string {
	return "hstream-console"
}

func (h *HStreamConsole) Display() map[string]utils.DisplayedComponent {
	cfgDir, dataDir := h.getDirs()
	server := utils.DisplayedComponent{
		Name:          "HStreamConsole",
		Host:          h.spec.Host,
		Ports:         strconv.Itoa(h.spec.Port),
		ContainerName: h.ContainerName,
		Image:         h.spec.Image,
		Paths:         strings.Join([]string{cfgDir, dataDir}, ","),
	}
	return map[string]utils.DisplayedComponent{"hstreamConsole": server}
}

func (h *HStreamConsole) InitEnv(globalCtx *GlobalCtx) *executor.ExecuteCtx {
	cfgDir, dataDir := h.getDirs()
	args := append([]string{}, "mkdir -p", cfgDir, dataDir, "-m 0775")
	args = append(args, fmt.Sprintf("&& chown -R %[1]s:$(id -gn %[1]s) %[2]s %[3]s", globalCtx.User, cfgDir, dataDir))
	return &executor.ExecuteCtx{Target: h.spec.Host, Cmd: strings.Join(args, " ")}
}

func (h *HStreamConsole) Deploy(globalCtx *GlobalCtx) *executor.ExecuteCtx {
	mountPoints := []spec.MountPoints{
		{h.spec.DataDir, h.spec.DataDir},
		{h.spec.RemoteCfgPath + "/application.properties", "/hstream/application.properties"},
	}
	args := spec.GetDockerExecCmd(globalCtx.containerCfg, h.spec.ContainerCfg, h.ContainerName, true, mountPoints...)
	args = append(args, h.spec.Image)
	return &executor.ExecuteCtx{Target: h.spec.Host, Cmd: strings.Join(args, " ")}
}

func (h *HStreamConsole) Stop(cfg *GlobalCtx) *executor.ExecuteCtx {
	args := []string{"docker rm -f", h.ContainerName}
	return &executor.ExecuteCtx{Target: h.spec.Host, Cmd: strings.Join(args, " ")}
}

func (h *HStreamConsole) Remove(globalCtx *GlobalCtx) *executor.ExecuteCtx {
	args := []string{"docker rm -f", h.ContainerName}
	args = append(args, "&&", "rm -rf", h.spec.DataDir,
		h.spec.RemoteCfgPath)
	return &executor.ExecuteCtx{Target: h.spec.Host, Cmd: strings.Join(args, " ")}
}

func (h *HStreamConsole) SyncConfig(globalCtx *GlobalCtx) *executor.TransferCtx {
	var prometheusUrl string
	if len(globalCtx.PrometheusUrls) == 0 {
		log.Warnf("get empty prometheus url when genarate hstream-console config file.")
	} else {
		prometheusUrl = fmt.Sprintf("http://%s", globalCtx.PrometheusUrls[0])
	}
	cfg := config.ConsoleConfig{
		Port:          h.spec.Port,
		ServerAddr:    globalCtx.HStreamServerUrls,
		EndpointAddr:  globalCtx.HServerEndPoints,
		PrometheusUrl: prometheusUrl,
	}
	genCfg, err := cfg.GenConfig()
	if err != nil {
		log.Errorf("gen ConsoleConfig error: %s", err.Error())
		os.Exit(1)
	}

	positions := []executor.Position{
		{LocalDir: genCfg, RemoteDir: filepath.Join(h.spec.RemoteCfgPath, "application.properties")},
	}

	return &executor.TransferCtx{Target: h.spec.Host, Position: positions}
}

func (h *HStreamConsole) getDirs() (string, string) {
	return h.spec.RemoteCfgPath, h.spec.DataDir
}
