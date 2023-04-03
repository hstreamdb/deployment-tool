package service

import (
	"fmt"
	"github.com/hstreamdb/deployment-tool/pkg/executor"
	"github.com/hstreamdb/deployment-tool/pkg/spec"
	"github.com/hstreamdb/deployment-tool/pkg/utils"
	"strconv"
	"strings"
)

type HttpServer struct {
	httpServerId  uint32
	spec          spec.HttpServerSpec
	ContainerName string
}

func NewHttpServer(id uint32, metaSpec spec.HttpServerSpec) *HttpServer {
	return &HttpServer{httpServerId: id, spec: metaSpec, ContainerName: spec.HttpServerDefaultContainerName}
}

func (h *HttpServer) GetServiceName() string {
	return "http-server"
}

func (h *HttpServer) Display() map[string]utils.DisplayedComponent {
	cfgDir, dataDir := h.getDirs()
	server := utils.DisplayedComponent{
		Name:          "HttpServer",
		Host:          h.spec.Host,
		Ports:         strconv.Itoa(h.spec.Port),
		ContainerName: h.ContainerName,
		Image:         h.spec.Image,
		Paths:         strings.Join([]string{cfgDir, dataDir}, ","),
	}
	return map[string]utils.DisplayedComponent{"httpServer": server}
}

func (h *HttpServer) InitEnv(globalCtx *GlobalCtx) *executor.ExecuteCtx {
	cfgDir, dataDir := h.getDirs()
	args := append([]string{}, "sudo mkdir -p", cfgDir, dataDir, "-m 0775")
	args = append(args, fmt.Sprintf("&& sudo chown -R %[1]s:$(id -gn %[1]s) %[2]s %[3]s", globalCtx.User, cfgDir, dataDir))
	return &executor.ExecuteCtx{Target: h.spec.Host, Cmd: strings.Join(args, " ")}
}

func (h *HttpServer) Deploy(globalCtx *GlobalCtx) *executor.ExecuteCtx {
	args := spec.GetDockerExecCmd(globalCtx.containerCfg, h.spec.ContainerCfg, h.ContainerName, true)
	args = append(args, h.spec.Image)
	args = append(args, "hstream-http-server", "-address", fmt.Sprintf("0.0.0.0:%d", h.spec.Port))
	args = append(args, "-services-url", globalCtx.HStreamServerUrls)
	return &executor.ExecuteCtx{Target: h.spec.Host, Cmd: strings.Join(args, " ")}
}

func (h *HttpServer) Stop(globalCtx *GlobalCtx) *executor.ExecuteCtx {
	args := []string{"docker rm -f", h.ContainerName}
	return &executor.ExecuteCtx{Target: h.spec.Host, Cmd: strings.Join(args, " ")}
}

func (h *HttpServer) Remove(globalCtx *GlobalCtx) *executor.ExecuteCtx {
	args := []string{"docker rm -f", h.ContainerName}
	args = append(args, "&&", "sudo rm -rf", h.spec.DataDir,
		h.spec.RemoteCfgPath)
	return &executor.ExecuteCtx{Target: h.spec.Host, Cmd: strings.Join(args, " ")}
}

func (h *HttpServer) SyncConfig(globalCtx *GlobalCtx) *executor.TransferCtx {
	return nil
}

func (h *HttpServer) getDirs() (string, string) {
	return h.spec.RemoteCfgPath, h.spec.DataDir
}
