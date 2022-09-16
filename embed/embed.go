package embed

import (
	em "embed"
)

//go:embed config
var embededConfig em.FS

func ReadConfig(path string) ([]byte, error) {
	return embededConfig.ReadFile(path)
}

//go:embed script
var embedScript em.FS

func ReadScript(path string) ([]byte, error) {
	return embedScript.ReadFile(path)
}
