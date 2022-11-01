package executor

import (
	log "github.com/sirupsen/logrus"
)

type DebugExecutor struct {
	User         string
	Password     string
	IdentityFile string
}

func NewDebugExecutor(user, password, keyPath string) *DebugExecutor {
	log.Debugf("create executor, user: %s, password: %s, identityFile: %s\n", user, password, keyPath)
	return &DebugExecutor{
		User:         user,
		Password:     password,
		IdentityFile: keyPath,
	}
}

func (d *DebugExecutor) Execute(address, cmd string) (string, error) {
	log.Debugf("Execute [%s]: %s\n", address, cmd)
	return "", nil
}

func (d *DebugExecutor) Close() {

}

func (d *DebugExecutor) Transfer(address, localPath, remotePath string) error {
	log.Debugf("Scp [%s] %s@%s:%s\n", localPath, d.User, address, remotePath)
	return nil
}
