package config

import (
	"bytes"
	"fmt"
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

func (f *FilebeatConfig) GenConfig() (string, error) {
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
	if err = tpl.Execute(content, f); err != nil {
		return "", err
	}

	path = filepath.Join("template", "filebeat", fmt.Sprintf("filebeat_%s_filebeat.yml", f.FilebeatHost))
	return path, os.WriteFile(path, content.Bytes(), 0664)
}
