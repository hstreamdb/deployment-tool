package main

import (
	"bytes"
	"fmt"
	"github.com/hstreamdb/deployment-tool/embed"
	"github.com/hstreamdb/deployment-tool/pkg/utils"
	"github.com/spf13/cobra"
	"os"
	"path/filepath"
	"text/template"
)

func newInit() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init",
		Short: "Init generates a configuration file template and initializes the execution environment",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := utils.MakeDirs([]utils.DirCfg{
				{Path: "template/script", Perm: 0755},
				{Path: "template/prometheus", Perm: 0755},
				{Path: "template/grafana/dashboards", Perm: 0755},
				{Path: "template/grafana/datasources", Perm: 0755},
				{Path: "template/alertmanager", Perm: 0755},
				{Path: "template/filebeat", Perm: 0755},
				{Path: "template/kibana", Perm: 0755},
			}); err != nil {
				return err
			}

			configFile := filepath.Join("config", "config.yaml")
			if err := getFile(configFile, "template/config.yaml"); err != nil {
				return err
			}

			alertManagerFile := filepath.Join("config", "alertmanager.yml")
			if err := getFile(alertManagerFile, "template/alertmanager/alertmanager.yml"); err != nil {
				return err
			}

			kibanaFile := filepath.Join("config", "kibana", "export.ndjson")
			if err := getFile(kibanaFile, "template/kibana/export.ndjson"); err != nil {
				return err
			}
			return nil
		},
	}
	return cmd
}

func getFile(fp string, target string) error {
	tpl, err := embed.ReadConfig(fp)
	if err != nil {
		return fmt.Errorf("get %s template error: %s\n", fp, err.Error())
	}
	cfg, err := template.New("DefaultConfig").Parse(string(tpl))
	if err != nil {
		return fmt.Errorf("render %s template error: %s\n", fp, err.Error())
	}

	content := bytes.NewBufferString("")
	if err := cfg.Execute(content, nil); err != nil {
		return err
	}

	if err := os.WriteFile(target, content.Bytes(), 0664); err != nil {
		return fmt.Errorf("write %s error: %s\n", target, err.Error())
	}
	return nil
}
