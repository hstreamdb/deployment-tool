package config

import (
	"bytes"
	"github.com/hstreamdb/deployment-tool/embed"
	"os"
	"path/filepath"
	"text/template"
)

type KibanaConfig struct {
	KibanaHost        string
	KibanaPort        string
	ElasticSearchHost string
	ElasticSearchPort string
}

func (k *KibanaConfig) GenConfig() (string, error) {
	path := filepath.Join("config", "kibana", "kibana.tpl")
	cfg, err := embed.ReadConfig(path)
	if err != nil {
		return "", err
	}

	tpl, err := template.New("Kibana").Parse(string(cfg))
	if err != nil {
		return "", err
	}

	content := bytes.NewBufferString("")
	if err = tpl.Execute(content, k); err != nil {
		return "", err
	}

	path = filepath.Join("template", "kibana", "kibana.yml")
	return path, os.WriteFile(path, content.Bytes(), 0664)
}
