global:
  scrape_interval: 15s
  evaluation_interval: 30s

  external_labels:
    monitor: "hstream-monitor"

alerting:
  alertmanagers:
    - static_configs:
        - targets:
          {{- range .AlertManagerAddress }}
          - '{{.}}'
          {{- end }}

rule_files:
  - "disks.yml"
  - "cluster.yml"
  - "zk.yml"
  - "alert.yml"

scrape_configs:
{{- if .NodeExporterAddress }}
  - job_name: "node_exporter_task"
    scrape_interval: 30s
    static_configs:
    - targets:
      {{- range .NodeExporterAddress }}
      - '{{.}}'
      {{- end }}
      labels:
        group: 'node_exporter'
    relabel_configs:
      - source_labels: [__address__]
        target_label: instance
        separator: ':'
        regex: '(.*):.*'
        replacement: "${1}"
{{- end }}

{{- if .CadVisorAddress }}
  - job_name: "cadvisor_task"
    scrape_interval: 30s
    static_configs:
    - targets:
      {{- range .CadVisorAddress }}
      - '{{.}}'
      {{- end }}
      labels:
        group: 'cadvisor'
    relabel_configs:
      - source_labels: [__address__]
        target_label: instance
        separator: ':'
        regex: '(.*):.*'
        replacement: "${1}"
{{- end }}

{{- if .BlackBoxAddress }}
  - job_name: "blackbox"
    scrape_interval: 30s
    metrics_path: /probe
    params:
      module: [tcp_connect]
    static_configs:
    {{- range $key, $value := .BlackBoxTargets }}
    {{- if $value }}
      - targets:
        {{- range $value }}
        - '{{.}}'
        {{- end }}
        labels:
          group: '{{ $key }}'
    {{- end }}
    {{- end }}
    relabel_configs:
      - source_labels: [__address__]
        target_label: __param_target
      - source_labels: [__param_target]
        target_label: instance
      - target_label: __address__
        replacement: {{ .BlackBoxAddress }}
{{- end }}

  - job_name: "hstream_metrics"
    scrape_interval: 5s
    static_configs:
    - targets:
      {{- range .HStreamExporterAddress }}
      - '{{.}}'
      {{- end }}