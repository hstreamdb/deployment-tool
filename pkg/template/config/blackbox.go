package config

import (
	"github.com/hstreamdb/deployment-tool/embed"
	"os"
	"path/filepath"
)

type BlackBoxConfig struct{}

func (b *BlackBoxConfig) GenConfig() (string, error) {
	path := filepath.Join("config", "blackbox", "blackbox.yml")
	cfg, err := embed.ReadConfig(path)
	if err != nil {
		return "", err
	}

	path = filepath.Join("template", "blackbox", "blackbox.yml")
	return path, os.WriteFile(path, cfg, 0664)
}
