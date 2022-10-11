package service

import (
	"fmt"
	"github.com/hstreamdb/deployment-tool/pkg/executor"
	"github.com/hstreamdb/deployment-tool/pkg/spec"
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

func (m *HttpServer) InitEnv(globalCtx *GlobalCtx) *executor.ExecuteCtx {
	cfgDir, dataDir := m.getDirs()
	args := append([]string{}, "sudo mkdir -p", cfgDir, dataDir)
	return &executor.ExecuteCtx{Target: m.spec.Host, Cmd: strings.Join(args, " ")}
}

func (m *HttpServer) Deploy(globalCtx *GlobalCtx) *executor.ExecuteCtx {
	args := spec.GetDockerExecCmd(globalCtx.containerCfg, m.spec.ContainerCfg, spec.HttpServerDefaultContainerName, true)
	args = append(args, m.spec.Image)
	args = append(args, "hstream-http-server", "-address", fmt.Sprintf("0.0.0.0:%d", m.spec.Port))
	args = append(args, "-services-url", globalCtx.HStreamServerUrls)
	return &executor.ExecuteCtx{Target: m.spec.Host, Cmd: strings.Join(args, " ")}
}

func (m *HttpServer) Remove(globalCtx *GlobalCtx) *executor.ExecuteCtx {
	args := []string{"docker rm -f", spec.HttpServerDefaultContainerName}
	args = append(args, "&&", "sudo rm -rf", m.spec.DataDir,
		m.spec.RemoteCfgPath)
	return &executor.ExecuteCtx{Target: m.spec.Host, Cmd: strings.Join(args, " ")}
}

func (m *HttpServer) SyncConfig(globalCtx *GlobalCtx) *executor.TransferCtx {
	return nil
}

func (m *HttpServer) getDirs() (string, string) {
	return m.spec.RemoteCfgPath, m.spec.DataDir
}
