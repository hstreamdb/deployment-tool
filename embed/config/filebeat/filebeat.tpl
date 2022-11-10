filebeat.inputs:
  - type: container
    paths:
      - "/var/lib/docker/containers/*/*.log"

processors:
  - add_docker_metadata:
      host: "unix:///var/run/docker.sock"

  - add_labels:
      labels:
        host: {{ .FilebeatHost }}

  - decode_json_fields:
      fields: ["message"]
      target: "json"
      overwrite_keys: true

output.elasticsearch:
  hosts: ["{{ .ElasticsearchHost }}:{{ .ElasticsearchPort }}"]
  indices:
    - index: "filebeat-%{[agent.version]}-%{+yyyy.MM.dd}"

logging.json: true
logging.metrics.enabled: false
