package spec

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"gotest.tools/v3/assert"
	"testing"
)

func TestParseConfig(t *testing.T) {
	t.Parallel()
	config := ComponentsSpec{}
	err := yaml.Unmarshal([]byte(`
global:
  user: "root"
  key_path: "~/.ssh/test.pem"
  ssh_port: 22
  container_config:
    disable_restart: true
    remove_when_exit: true
  #    cpu_limit: 200
  #    memory_limit: 8G

hserver:
  - host: 10.1.0.10
    image: "hstreamdb/hstream"
    server_config:
      server_log_level: info
      store_log_level: error
#    local_config_path: $PWD/server.yaml
#    remote_config_path: "/home/deploy/hserver"
#    cpu_limit: 200
#    memory_limit: 8G
  - host: 10.1.0.11
    image: "hstreamdb/hstream"
#    local_config_path: $PWD/server.yaml
#    remote_config_path: "/home/deploy/hserver"
#    cpu_limit: 200
#    memory_limit: 8G

hstore:
  - host: 10.1.0.10
    image: "hstreamdb/hstream"
    local_config_path: $PWD/logdevice.conf
    remote_config_path: "/home/deploy/hstore"
    data_dir: "/home/deploy/data/store"
    disk: 1
    shards: 2
    role: "Both" # [Storage|Sequencer|Both]
    enable_admin: true
    container_config:
      cpu_limit: 200
      memory_limit: 8G
  - host: 10.1.0.11
    image: "hstreamdb/hstream"
    local_config_path: $PWD/logdevice.conf
    remote_config_path: "/home/deploy/hstore"
    data_dir: "/home/deploy/data/store"
    disk: 1
    shards: 2
    role: "Both" # [Storage|Sequencer|Both]
    container_config:
      cpu_limit: 200
      memory_limit: 8G
  - host: 10.1.0.12
    image: "hstreamdb/hstream"
    local_config_path: $PWD/logdevice.conf
    remote_config_path: "/home/deploy/hstore"
    data_dir: "/home/deploy/data/store"
    disk: 1
    shards: 2
    container_config:
      cpu_limit: 200
      memory_limit: 8G

hadmin:
  - host: 10.1.0.11
    image: "hstreamdb/hstream"
    meta_replica: 3
    embed: true

meta_store:
  - host: 10.1.0.13
    image: "zookeeper:3.6"
    data_dir: "/home/deploy/data/meta"
    local_config_path: $PWD/logdevice.conf
    remote_config_path: "/home/deploy/hstore"
`), &config)
	assert.NilError(t, err)

	t.Logf("%+v\n", config)
}

func TestGetContainerCfg(t *testing.T) {
	t.Parallel()
	config := ComponentsSpec{}
	err := yaml.Unmarshal([]byte(`
global:
  user: "root"
  key_path: "~/.ssh/test.pem"
  ssh_port: 22
  container_config:
    disable_restart: true
    remove_when_exit: true
    cpu_limit: 200
`), &config)
	assert.NilError(t, err)
	fmt.Printf("%+v\n", config)

	cfg := GetContainerCfg(config.Global)
	fmt.Println(cfg)
	assert.Equal(t, cfg.Cpu, "200")
	assert.Equal(t, cfg.DisableRestart, true)
	assert.Equal(t, cfg.RemoveWhenExit, true)
	assert.Equal(t, cfg.Memory, "")
}

func TestMergeContainerCfg(t *testing.T) {
	t.Parallel()
	tests := map[string]struct {
		lhs  ContainerCfg
		rhs  ContainerCfg
		want ContainerCfg
	}{
		"empty lhs": {
			lhs: ContainerCfg{},
			rhs: ContainerCfg{
				DisableRestart: false,
				Memory:         "2G",
			},
			want: ContainerCfg{
				DisableRestart: false,
				Memory:         "2G",
			},
		},
		"empty rhs": {
			lhs: ContainerCfg{
				DisableRestart: false,
				Memory:         "2G",
			},
			rhs: ContainerCfg{},
			want: ContainerCfg{
				DisableRestart: false,
				Memory:         "2G",
			},
		},
		"merge two": {
			lhs: ContainerCfg{
				Memory:         "2G",
				DisableRestart: false,
			},
			rhs: ContainerCfg{
				Memory:         "4G",
				RemoveWhenExit: false,
			},
			want: ContainerCfg{
				Memory:         "4G",
				DisableRestart: false,
				RemoveWhenExit: false,
			},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			get := MergeContainerCfg(tc.lhs, tc.rhs)
			assert.Equal(t, get, tc.want)
		})
	}
}
