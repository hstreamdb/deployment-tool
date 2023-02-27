# hstream-console monitor port
server.port = {{ .Port }}

# hstream server address
plain.hstream.privateAddress = {{ .ServerAddr }}
# endpoint info
plain.hstream.publicAddress = {{ .EndpointAddr }}

# prometheus service address
monitor.prometheus.url = {{ .PrometheusUrl }}