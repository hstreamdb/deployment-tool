#
# ** THIS IS AN AUTO-GENERATED FILE **
#

# Default Kibana configuration for docker target
server.host: "{{ .KibanaHost }}"
server.port: {{ .KibanaPort }}

server.shutdownTimeout: "5s"
elasticsearch.hosts:
    ["http://{{ .ElasticSearchHost }}:{{ .ElasticSearchPort }}"]
monitoring.ui.container.elasticsearch.enabled: true

xpack.security.enabled: false
xpack.security.authc:
    anonymous:
        roles: kibana_admin
