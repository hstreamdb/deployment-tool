package task

import (
	"fmt"
	ext "github.com/hstreamdb/deployment-tool/pkg/executor"
	"github.com/hstreamdb/deployment-tool/pkg/service"
	log "github.com/sirupsen/logrus"
)

type BootstrapCtx struct {
	ctx *service.GlobalCtx
}

func (b *BootstrapCtx) String() string {
	return "Task: bootstrap hstore cluster"
}

func (b *BootstrapCtx) Run(executor ext.Executor) error {
	for _, admin := range b.ctx.HAdminInfos {
		executorCtx := service.Bootstrap(b.ctx, admin)
		target := fmt.Sprintf("%s:%d", executorCtx.Target, b.ctx.SSHPort)
		res, err := executor.Execute(target, executorCtx.Cmd)
		if err != nil {
			log.Errorf("bootstrap with %s error: %s, res: %s", target, err, res)
			continue
		}
		return nil
	}
	return fmt.Errorf("bootstrap error")
}

type CheckClusterStatusCtx struct {
	ctx            *service.GlobalCtx
	serverServices []*service.HServer
}

func (c *CheckClusterStatusCtx) String() string {
	return "Task: check cluster status"
}

func (c *CheckClusterStatusCtx) Run(executor ext.Executor) error {
	success := false
	for _, admin := range c.ctx.HAdminInfos {
		executorCtx := service.AdminStoreCmd(c.ctx, admin, "status")
		target := fmt.Sprintf("%s:%d", executorCtx.Target, c.ctx.SSHPort)
		res, err := executor.Execute(target, executorCtx.Cmd)
		if err != nil {
			log.Errorf("get store status with %s error: %s, res: %s", target, err, res)
			continue
		}
		success = true
		fmt.Printf("=== HStore Status ===\n%s\n", res)
		break
	}

	if !success {
		return fmt.Errorf("can't get store status")
	}

	if len(c.serverServices) == 0 {
		return nil
	}

	for _, admin := range c.ctx.HAdminInfos {
		executorCtx := service.AdminServerCmd(c.ctx, admin, c.serverServices[0].Host, c.serverServices[0].Port, "status")
		target := fmt.Sprintf("%s:%d", executorCtx.Target, c.ctx.SSHPort)
		res, err := executor.Execute(target, executorCtx.Cmd)
		if err != nil {
			log.Errorf("get server status with %s error: %s, res: %s", target, err, res)
			continue
		}
		fmt.Printf("=== HServer Status ===\n%s\n", res)
		return nil
	}

	return fmt.Errorf("can't get server status")
}
