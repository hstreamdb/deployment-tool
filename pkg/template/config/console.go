package config

import (
	"bytes"
	"github.com/hstreamdb/deployment-tool/embed"
	"os"
	"path/filepath"
	"text/template"
)

type ConsoleConfig struct {
	Port          int
	ServerAddr    string
	EndpointAddr  string
	PrometheusUrl string
}

func (c *ConsoleConfig) GenConfig() (string, error) {
	path := filepath.Join("config", "hstream_console", "application.properties.tpl")
	cfg, err := embed.ReadConfig(path)
	if err != nil {
		return "", err
	}

	tpl, err := template.New("Console").Parse(string(cfg))
	if err != nil {
		return "", err
	}

	content := bytes.NewBufferString("")
	if err = tpl.Execute(content, c); err != nil {
		return "", err
	}

	path = filepath.Join("template", "hstream_console", "application.properties")
	return path, os.WriteFile(path, content.Bytes(), 0664)
}
