package config

import (
	"bytes"
	"github.com/hstreamdb/deployment-tool/embed"
	"os"
	"path/filepath"
	"text/template"
)

type HStoreConfig struct {
	ZkUrl string
}

func (m *HStoreConfig) GenConfig() (string, error) {
	ph := filepath.Join("config", "logdevice.config.tpl")
	sh, err := embed.ReadConfig(ph)
	if err != nil {
		return "", err
	}

	tpl, err := template.New("HStoreConf").Parse(string(sh))
	if err != nil {
		return "", err
	}

	content := bytes.NewBufferString("")
	if err = tpl.Execute(content, m); err != nil {
		return "", err
	}

	file := filepath.Join("template", "logdevice.conf")
	return file, os.WriteFile(file, content.Bytes(), 0664)
}
