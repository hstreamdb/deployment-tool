package executor

type Executor interface {
	Execute(target, cmd string) (string, error)
	Transfer(target, localPath, remotePath string) error
	Close()
}

type Cmd string

type ExecuteCtx struct {
	Target string
	Cmd    string
}

type Position struct {
	LocalDir  string
	RemoteDir string
	Opts      string
}

type TransferCtx struct {
	Target   string
	Position []Position
}
