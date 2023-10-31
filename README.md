# deployment-tool

This repository contains a command tool `hdt`, which can be used to set up a HStreamDB Cluster with docker.

## Quick start

### Check environment

- Start HStreamDB requires an operating system kernel version greater than at least Linux 4.14. Check with command:

  ```shell
  uname -r
  ```

- Make sure docker is installed.

- Make sure that the log-in user has `sudo` execute privileges，and configure `sudo` without password.

- For nodes which deploy `HStore` instances, mount the data disk to `/mnt/data*/`.

  - "*" Matching incremental numbers, start from zero
  - one disk should mount to one directory. e.g. if we have two data disks `/dev/vdb` and `/dev/vdc`, then `/dev/vdb` should mount to `/mnt/data0` and `/dev/vdc` should mount to `/mnt/data1`

### Installation

Binaries are available here: https://github.com/hstreamdb/deployment-tool/releases

### Generate configuration template && init local environment

```shell
./hdt init
```

The current directory structure will be as follows after running the `init` command:

```shell
├── hdt
└── template                 
    ├── config.yaml
    ├── grafana
    │   ├── dashboards
    │   └── datasources
    ├── prometheus
    └── script
```

### Update `config.yaml`

Update the `config.yaml` file with cluster-related information. The configuration in the `config.yaml` template will deploy a cluster on 3 nodes, each consisting of a `HServer` instance, a `HStore` instance, a `Meta-Store` instance and associated monitoring components. `Prometheus` 、`Grafana`  and other monitor components will be deploy on a separate node.

To use this configuration file, just update the host information of the node and the ssh key-pair path. The final configuration file may looks like:

```shell
global:
  user: "root"
  key_path: "~/.ssh/hstream.pem"
  ssh_port: 22

monitor:
  node_exporter_port: 9100
  cadvisor_port: 7000
  grafana_disable_login: true

hserver:
  - host: 172.24.47.173
  - host: 172.24.47.174
  - host: 172.24.47.175

hstore:
  - host: 172.24.47.173
    enable_admin: true
  - host: 172.24.47.174
  - host: 172.24.47.175

meta_store:
  - host: 172.24.47.173
  - host: 172.24.47.174
  - host: 172.24.47.175

prometheus:
  - host: 172.24.47.172

grafana:
  - host: 172.24.47.172
  
hstream_exporter:
  - host: 172.24.47.172
```

### Set up cluster

```shell
./hdt start 
```

### Remove cluster

```shell
./hdt remove
```

