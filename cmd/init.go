package main

import (
	"fmt"
	"github.com/hstreamdb/deployment-tool/embed"
	"github.com/hstreamdb/deployment-tool/pkg/utils"
	"github.com/spf13/cobra"
	"os"
	"path/filepath"
)

const (
	available800 = "8.0.0"
	available760 = "7.6.0"
)

func newInit() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init",
		Short: "Init generates a configuration file template and initializes the execution environment",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := utils.MakeDirs([]utils.DirCfg{
				{Path: "template/script", Perm: 0755},
				{Path: "template/prometheus", Perm: 0755},
				{Path: "template/prometheus_common", Perm: 0755},
				{Path: "template/blackbox", Perm: 0755},
				{Path: "template/grafana/dashboards", Perm: 0755},
				{Path: "template/grafana/datasources", Perm: 0755},
				{Path: "template/alertmanager", Perm: 0755},
				{Path: "template/filebeat", Perm: 0755},
				{Path: "template/kibana", Perm: 0755},
				{Path: "template/hstream_console", Perm: 0755},
			}); err != nil {
				return err
			}

			fileMaps := map[string]string{
				filepath.Join("config", "config.yaml"):                             "template/config.yaml",
				filepath.Join("config", "alertmanager.yml"):                        "template/alertmanager/alertmanager.yml",
				filepath.Join("config", "logdevice.config"):                        "template/logdevice.conf",
				filepath.Join("config/blackbox", "blackbox.yml"):                   "template/blackbox/blackbox.yml",
				filepath.Join("config/grafana/dashboards", "dashboard.yml"):        "template/grafana/dashboards/dashboard.yml",
				filepath.Join("config/grafana/dashboards", "hstream_monitor.json"): "template/grafana/dashboards/hstream_monitor.json",
				filepath.Join("config/grafana/dashboards", "hstream_kafka.json"):   "template/grafana/dashboards/hstream_kafka.json",
				filepath.Join("config/grafana/datasources", "datasource.yml"):      "template/grafana/datasources/datasource.yml",
				filepath.Join("config/kibana", "export_7.6.0.ndjson"):              "template/kibana/export_7.6.0.ndjson",
				filepath.Join("config/kibana", "export_8.0.0.ndjson"):              "template/kibana/export_8.0.0.ndjson",
				filepath.Join("config/prometheus", "alert.yml"):                    "template/prometheus_common/alert.yml",
				filepath.Join("config/prometheus", "cluster.yml"):                  "template/prometheus_common/cluster.yml",
				filepath.Join("config/prometheus", "disks.yml"):                    "template/prometheus_common/disks.yml",
				filepath.Join("config/prometheus", "zk.yml"):                       "template/prometheus_common/zk.yml",
			}
			return getFiles(fileMaps)
		},
	}
	return cmd
}

func getFiles(fileMap map[string]string) error {
	for k, v := range fileMap {
		if err := getFile(k, v); err != nil {
			return err
		}
	}
	return nil
}

func getFile(origin string, target string) error {
	tpl, err := embed.ReadConfig(origin)
	if err != nil {
		return fmt.Errorf("get %s file error: %s\n", origin, err.Error())
	}

	if err = os.WriteFile(target, tpl, 0664); err != nil {
		return fmt.Errorf("write %s error: %s\n", target, err.Error())
	}
	return nil
}
