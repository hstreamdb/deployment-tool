package config

import (
	"bytes"
	"fmt"
	"github.com/hstreamdb/deployment-tool/embed"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

type PrometheusConfig struct {
	NodeExporterAddress    []string
	CadVisorAddress        []string
	HStreamExporterAddress []string
	AlertManagerAddress    []string
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
	for _, p := range []string{"cluster.yml", "disks.yml", "zk.yml", "alert.yml"} {
		path := filepath.Join("template", "prometheus", p)
		content, err := embed.ReadConfig(filepath.Join("config", "prometheus", p))
		if err != nil {
			return "", err
		}
		if err = os.WriteFile(path, content, 0664); err != nil {
			return "", err
		}
	}
	return file, os.WriteFile(file, content.Bytes(), 0664)
}

type GrafanaConfig struct{}

func (g *GrafanaConfig) GenConfig() (string, error) {
	grafanaRoot := embed.GetGrafanaRoot()

	if err := fs.WalkDir(grafanaRoot, "config/grafana", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			fmt.Printf("visit %s error: %s\n", path, err.Error())
			return err
		}
		if !d.IsDir() {
			paths := strings.Split(path, "/")
			n := len(paths)
			content, err := embed.ReadConfig(path)
			if err != nil {
				return err
			}
			target := filepath.Join("template", "grafana", paths[n-2], paths[n-1])
			if err = os.WriteFile(target, content, 0664); err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		return "", err
	}
	return filepath.Join("template", "grafana"), nil
}
