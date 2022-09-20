package embed

import (
	em "embed"
	"io/fs"
)

//go:embed config
var embededConfig em.FS

func ReadConfig(path string) ([]byte, error) {
	return embededConfig.ReadFile(path)
}

//go:embed config/grafana
var grafanaRoot em.FS

func GetGrafanaRoot() fs.FS {
	return grafanaRoot
}

//go:embed script
var embedScript em.FS

func ReadScript(path string) ([]byte, error) {
	return embedScript.ReadFile(path)
}
