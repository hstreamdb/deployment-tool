#
# ** THIS IS AN AUTO-GENERATED FILE **
#

# Default Kibana configuration for docker target
server.host: "{{ .KibanaHost }}"
server.port: {{ .KibanaPort }}

elasticsearch.hosts:
    ["http://{{ .ElasticSearchHost }}:{{ .ElasticSearchPort }}"]
