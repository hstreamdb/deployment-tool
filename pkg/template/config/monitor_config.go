package config

import (
	"bytes"
	"github.com/hstreamdb/deployment-tool/embed"
	"os"
	"path/filepath"
	"text/template"
)

type PrometheusConfig struct {
	ClusterId              string
	NodeExporterAddress    []string
	CadVisorAddress        []string
	HStreamExporterAddress []string
	AlertManagerAddress    []string
	BlackBoxAddress        string
	BlackBoxTargets        map[string][]string
	MetaZkAddress          []string
}

func (p *PrometheusConfig) GenConfig() (string, error) {
	ph := filepath.Join("config", "prometheus", "prometheus.tpl")
	sh, err := embed.ReadConfig(ph)
	if err != nil {
		return "", err
	}

	tpl, err := template.New("Prometheus").Parse(string(sh))
	if err != nil {
		return "", err
	}

	content := bytes.NewBufferString("")
	if err = tpl.Execute(content, p); err != nil {
		return "", err
	}

	file := filepath.Join("template", "prometheus", "prometheus.yml")
	return file, os.WriteFile(file, content.Bytes(), 0664)
}
