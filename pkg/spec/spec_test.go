package spec

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"gotest.tools/v3/assert"
	"testing"
)

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

func TestUpdateComponentSpecWithGlobal(t *testing.T) {
	t.Parallel()
	globalCfg := GlobalCfg{
		User:                "root",
		KeyPath:             "aaa",
		SSHPort:             22,
		MetaReplica:         3,
		MetaStoreConfigPath: "meta_store_config_path",
		HStoreConfigPath:    "store_config_path",
		HServerConfigPath:   "server_config_path",
		ContainerCfg: ContainerCfg{
			Cpu:            "200",
			Memory:         "8G",
			RemoveWhenExit: true,
			DisableRestart: true,
		},
	}

	cmpSpec := &ComponentsSpec{
		HServer: []HServerSpec{
			{
				Host:    "127.0.0.1",
				Image:   "hstreamdb/hstream:v0.8.4",
				SSHPort: 21,
				ContainerCfg: ContainerCfg{
					Cpu: "100",
				},
			},
			{
				Host:    "127.0.0.2",
				Address: "127.1.1.1",
				ContainerCfg: ContainerCfg{
					Cpu:            "300",
					Memory:         "10G",
					RemoveWhenExit: false,
					DisableRestart: false,
				},
			},
		},
		HStore: []HStoreSpec{
			{
				DataDir: "abc",
			},
		},
	}

	err := updateComponentSpecWithGlobal(globalCfg, cmpSpec)
	assert.NilError(t, err)
	for _, server := range cmpSpec.HServer {
		fmt.Printf("%+v\n", server)
	}
	for _, store := range cmpSpec.HStore {
		fmt.Printf("%+v\n", store)
	}
}
