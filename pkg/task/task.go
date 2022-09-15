package task

import (
	ext "github.com/hstreamdb/dev-deploy/pkg/executor"
)

type Task interface {
	Run(executor ext.Executor) error
}
