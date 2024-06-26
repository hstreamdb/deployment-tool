global:
  scrape_interval: 15s
  evaluation_interval: 30s

  external_labels:
    monitor: "hstream-monitor"

{{ if .AlertManagerConfig }}
{{- range .AlertManagerConfig }}
alerting:
  alertmanagers:
    - static_configs:
        - targets:
          - '{{ .Address }}'
      {{- if .AuthUser }}
      basic_auth:
        username: '{{ .AuthUser }}'
        password: '{{ .AuthPassword }}'
      {{- end }}
{{- end }}
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
        cluster_id: '{{ .ClusterId }}'
    relabel_configs:
      - source_labels: [__address__]
        target_label: instance
        separator: ':'
        regex: '(.*):.*'
        replacement: "${1}"
      - source_labels: []
        target_label: source
        replacement: hstream
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
        cluster_id: '{{ .ClusterId }}'
    relabel_configs:
      - source_labels: [__address__]
        target_label: instance
        separator: ':'
        regex: '(.*):.*'
        replacement: "${1}"
      - source_labels: []
        target_label: source
        replacement: hstream
{{- end }}

{{- if .MetaZkAddress }}
  - job_name: "meta_zk_task"
    scrape_interval: 30s
    static_configs:
    - targets:
      {{- range .MetaZkAddress }}
      - '{{.}}'
      {{- end }}
      labels:
        group: 'meta_store'
        cluster_id: '{{ .ClusterId }}'
    relabel_configs:
      - source_labels: [__address__]
        target_label: instance
        separator: ':'
        regex: '(.*):.*'
        replacement: "${1}"
      - source_labels: []
        target_label: source
        replacement: hstream
{{- end }}

{{- if .BlackBoxAddress }}
{{ $clusterId := .ClusterId }}
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
          cluster_id: '{{ $clusterId }}'
    {{- end }}
    {{- end }}
    relabel_configs:
      - source_labels: [__address__]
        target_label: __param_target
      - source_labels: [__param_target]
        target_label: instance
      - target_label: __address__
        replacement: {{ .BlackBoxAddress }}
      - source_labels: []
        target_label: source
        replacement: hstream
{{- end }}

{{- if .HStreamExporterAddress }}
  - job_name: "hstream_metrics"
    scrape_interval: 5s
    static_configs:
    - targets:
      {{- range .HStreamExporterAddress }}
      - '{{.}}'
      {{- end }}
      labels:
        group: 'hstream-exporter'
        cluster_id: '{{ .ClusterId }}'
    relabel_configs:
      - source_labels: []
        target_label: source
        replacement: hstream
{{- end }}

{{- if .HStreamKafkaAddress }}
  - job_name: "hstream_kafka_metrics"
    scrape_interval: 5s
    static_configs:
    - targets:
      {{- range .HStreamKafkaAddress }}
      - '{{.}}'
      {{- end }}
      labels:
        group: 'hstream-kafka'
        cluster_id: '{{ .ClusterId }}'
    relabel_configs:
      - source_labels: [__address__]
        target_label: instance
        separator: ':'
        regex: '(.*):.*'
        replacement: "${1}"
      - source_labels: []
        target_label: source
        replacement: hstream
{{- end }}

{{- if .HStoreMonitorAddress }}
  - job_name: "hstore_metrics"
    scrape_interval: 5s
    static_configs:
    - targets:
      {{- range .HStoreMonitorAddress }}
      - '{{.}}'
      {{- end }}
      labels:
        group: 'hstore'
        cluster_id: '{{ .ClusterId }}'
    relabel_configs:
      - source_labels: [__address__]
        target_label: instance
        separator: ':'
        regex: '(.*):.*'
        replacement: "${1}"
      - source_labels: []
        target_label: source
        replacement: hstream
{{- end }}