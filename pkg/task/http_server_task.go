package task

import (
	ext "github.com/hstreamdb/deployment-tool/pkg/executor"
	"github.com/hstreamdb/deployment-tool/pkg/service"
)

type HttpServerCtx struct {
	ctx     *service.GlobalCtx
	service []*service.HttpServer
}

type InitHttpServerEnv struct {
	HttpServerCtx
}

func (s *InitHttpServerEnv) String() string {
	return "Task: init http-server environment"
}

func (s *InitHttpServerEnv) Run(executor ext.Executor) error {
	return serviceInitEnv(executor, s.ctx, s.service)
}

type SyncHttpServerConfig struct {
	HttpServerCtx
}

func (s *SyncHttpServerConfig) String() string {
	return "Task: sync http-server config"
}

func (s *SyncHttpServerConfig) Run(executor ext.Executor) error {
	return configSync(executor, s.ctx, s.service)
}

type StartHttpServer struct {
	HttpServerCtx
}

func (s *StartHttpServer) String() string {
	return "Task: start http-server"
}

func (s *StartHttpServer) Run(executor ext.Executor) error {
	return serviceDeploy(executor, s.ctx, s.service)
}

type RemoveHttpServer struct {
	HttpServerCtx
}

func (r *RemoveHttpServer) String() string {
	return "Task: remove http-server"
}

func (r *RemoveHttpServer) Run(executor ext.Executor) error {
	return serviceRemove(executor, r.ctx, r.service)
}
