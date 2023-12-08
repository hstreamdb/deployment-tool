package config

import (
	"bytes"
	"fmt"
	"github.com/hstreamdb/deployment-tool/pkg/utils"
	"os"
	"path/filepath"
	"text/template"

	"github.com/hstreamdb/deployment-tool/embed"
)

type AlertManagerConfig struct {
	Address      string
	AuthUser     string
	AuthPassword string
}

type PrometheusConfig struct {
	PromHost               string
	ClusterId              string
	NodeExporterAddress    []string
	CadVisorAddress        []string
	HStreamExporterAddress []string
	AlertManagerConfig     []AlertManagerConfig
	BlackBoxAddress        string
	BlackBoxTargets        map[string][]string
	MetaZkAddress          []string
	HStreamKafkaAddress    []string
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

	dst := filepath.Join("template", "prometheus", fmt.Sprintf("prometheus_%s", p.PromHost))

	if err = utils.CpDir("template/prometheus_common", dst); err != nil {
		return "", err
	}

	file := filepath.Join(dst, "prometheus.yml")
	return file, os.WriteFile(file, content.Bytes(), 0664)
}
