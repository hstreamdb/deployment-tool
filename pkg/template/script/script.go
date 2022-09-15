package script

type Script interface {
	GenScript() (string, error)
}
