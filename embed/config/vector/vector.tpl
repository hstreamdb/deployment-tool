[api]
  enabled = true

[sources.journald]
  type = "journald"

#[sources.docker]
#  type = "docker_logs"
#  exclude_containers = ["vector"]

#[transforms.remap_fields]
#  type = "remap"
#  inputs = ["journald"]
#  source = """
#      message = .message
#      host = .host
#      container_name = .CONTAINER_NAME
#      time_stamp = .SYSLOG_TIMESTAMP
#      . = {
#        "message": message,
#        "host": host,
#        "container_name": container_name,
#        "time_stamp": time_stamp
#      }
#  """

[transforms.remap_fields]
  type = "remap"
  inputs = ["journald"]
  source = '''
    del(.CONTAINER_ID)
    del(.CONTAINER_LOG_EPOCH)
    del(.CONTAINER_LOG_ORDINAL)
    del(.CONTAINER_TAG)
    del(.PRIORITY)
    del(.SYSLOG_IDENTIFIER)
    del(._BOOT_ID)
    del(._CAP_EFFECTIVE)
    del(._CMDLINE)
    del(._EXE)
    del(._GID)
    del(._MACHINE_ID)
    del(._SELINUX_CONTEXT)
    del(._SOURCE_REALTIME_TIMESTAMP)
    del(._SYSTEMD_CGROUP)
    del(._SYSTEMD_INVOCATION_ID)
    del(._SYSTEMD_SLICE)
    del(._SYSTEMD_UNIT)
    del(._TRANSPORT)
    del(._UID)
    del(.__MONOTONIC_TIMESTAMP)
    del(.__REALTIME_TIMESTAMP)
    .host = get_env_var!("VECTOR_MACHINE_IP")
'''

[transforms.filter]
  type = "filter"
  inputs = ["remap_fields"]
  condition = '.CONTAINER_NAME != "deploy_vector"'

[sinks.elasticsearch]
  type = "elasticsearch"
  inputs = ["filter"]
  endpoint = "http://{{ .ElasticsearchHost }}:{{ .ElasticsearchPort }}"
#  bulk.action = "create"
  bulk.index = "hstream-log-%Y-%m-%d"
#  compression = "zstd"
