{
    "server_settings": {
        "enable-nodes-configuration-manager": "true",
        "use-nodes-configuration-manager-nodes-configuration": "true",
        "enable-node-self-registration": "true",
        "enable-cluster-maintenance-state-machine": "true",
        "append-store-durability": "memory"
    },
    "client_settings": {
        "enable-nodes-configuration-manager": "true",
        "use-nodes-configuration-manager-nodes-configuration": "true",
        "admin-client-capabilities": "true"
    },
    "cluster": "logdevice",
    "internal_logs": {
        "config_log_deltas": {
            "replicate_across": {
                "node": 3
            }
        },
        "config_log_snapshots": {
            "replicate_across": {
                "node": 3
            }
        },
        "event_log_deltas": {
            "replicate_across": {
                "node": 3
            }
        },
        "event_log_snapshots": {
            "replicate_across": {
                "node": 3
            }
        },
        "maintenance_log_deltas": {
            "replicate_across": {
                "node": 3
            }
        },
        "maintenance_log_snapshots": {
            "replicate_across": {
                "node": 3
            }
        }
    },
    "metadata_logs": {
        "nodeset": [],
        "replicate_across": {
            "node": 3
        }
    },
    "zookeeper": {
        "zookeeper_uri": "{{ .ZkUrl }}",
        "timeout": "30s"
    }
}
