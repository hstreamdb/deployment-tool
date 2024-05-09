package config

import (
	"bytes"
	"fmt"
	"github.com/hstreamdb/deployment-tool/embed"
	"os"
	"path/filepath"
	"text/template"
)

type VectorConfig struct {
	VectorHost        string
	ElasticsearchHost string
	ElasticsearchPort string
}

func (v *VectorConfig) GenConfig() (string, error) {
	path := filepath.Join("config", "vector", "vector.tpl")
	cfg, err := embed.ReadConfig(path)
	if err != nil {
		return "", err
	}

	tpl, err := template.New("Vector").Parse(string(cfg))
	if err != nil {
		return "", err
	}

	content := bytes.NewBufferString("")
	if err = tpl.Execute(content, v); err != nil {
		return "", err
	}

	path = filepath.Join("template", "vector", fmt.Sprintf("vector_%s_vector.toml", v.VectorHost))
	return path, os.WriteFile(path, content.Bytes(), 0664)
}
