package config

import (
	"bytes"
	"github.com/hstreamdb/deployment-tool/embed"
	"os"
	"path/filepath"
	"text/template"
)

type FilebeatConfig struct {
	FilebeatHost      string
	ElasticsearchHost string
	ElasticsearchPort string
}

func (fbCfg *FilebeatConfig) GenConfig() (string, error) {
	path := filepath.Join("config", "filebeat", "filebeat.tpl")
	cfg, err := embed.ReadConfig(path)
	if err != nil {
		return "", err
	}

	tpl, err := template.New("Filebeat").Parse(string(cfg))
	if err != nil {
		return "", err
	}

	content := bytes.NewBufferString("")
	if err = tpl.Execute(content, fbCfg); err != nil {
		return "", err
	}

	path = filepath.Join("template", "filebeat", "filebeat.yml")
	return path, os.WriteFile(path, content.Bytes(), 0664)
}
