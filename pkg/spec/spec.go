package spec

import (
	"fmt"
	"github.com/creasty/defaults"
	"reflect"
	"strings"
)

const DefaultStoreConfigPath = "/logdevice.conf"

type ComponentsSpec struct {
	Global    GlobalCfg       `yaml:"global"`
	HServer   []HServerSpec   `yaml:"hserver"`
	HStore    []HStoreSpec    `yaml:"hstore"`
	MetaStore []MetaStoreSpec `yaml:"meta_store"`
}

func (c *ComponentsSpec) GetHosts() []string {
	v := reflect.Indirect(reflect.ValueOf(c))
	t := v.Type()

	globalCfgName := reflect.TypeOf(GlobalCfg{}).Name()
	res := []string{}
	for i := 0; i < t.NumField(); i++ {
		field := v.Field(i)
		if field.Type().Name() == globalCfgName {
			continue
		}
		res = append(res, getHostsInner(field)...)
	}
	return res
}

func getHostsInner(v reflect.Value) []string {
	t := v.Type()
	res := []string{}
	switch t.Kind() {
	case reflect.Slice:
		for i := 0; i < v.Len(); i++ {
			if hosts := getHostsInner(v.Index(i)); hosts != nil {
				res = append(res, hosts...)
			}
		}
	case reflect.Struct:
		host := v.FieldByName("Host").String()
		if host != "" {
			res = append(res, host)
		}
	}
	return res
}

func (c *ComponentsSpec) GetMetaStoreUrl() string {
	hosts := []string{}
	for _, spc := range c.MetaStore {
		hosts = append(hosts, spc.Host)
	}
	if len(hosts) == 0 {
		return ""
	}

	// append an empty string to help strings join
	hosts = append(hosts, "")
	url := strings.Join(hosts, ":2181,")
	url = url[:len(url)-1]
	return url
}

//func (c *ComponentsSpec) setDefault() error {
//	v := reflect.Indirect(reflect.ValueOf(c))
//	t := reflect.TypeOf(v)
//
//	for i := 0; i < t.NumField(); i++ {
//		field := v.Field(i)
//		switch field.Kind() {
//		case reflect.Struct:
//			setFieldDefault(field)
//		case reflect.Slice:
//			for j := 0; j < field.Len(); j++ {
//				setFieldDefault(field.Index(j))
//			}
//		default:
//			setFieldDefault(field)
//		}
//	}
//	return nil
//}
//
//func setFieldDefault(v reflect.Value) {
//	tp := reflect.New(v.Type())
//	tp.Elem().Set(v)
//	if err := defaults.Set(tp.Interface()); err != nil {
//		panic(err)
//	}
//	v.Set(tp.Elem())
//}

func (c *ComponentsSpec) UnmarshalYAML(unmarshal func(interface{}) error) error {
	// avoid recursion unmarshal
	type tmpSpec ComponentsSpec
	if err := unmarshal((*tmpSpec)(c)); err != nil {
		return err
	}

	// defaults.Set will initialize a list to an empty list, so if the default value is set first,
	// none of the list fields in the componentsSpec will be set correctly. that's why we have to
	// unmarshal first and then set the default value
	if err := defaults.Set(c); err != nil {
		return err
	}

	for i := 0; i < len(c.MetaStore); i++ {
		if len(c.MetaStore[i].DataDir) == 0 {
			c.MetaStore[i].DataDir = MetaStoreDefaultDataDir
		}
	}

	for i := 0; i < len(c.HStore); i++ {
		if len(c.HStore[i].DataDir) == 0 {
			c.HStore[i].DataDir = StoreDefaultDataDir
		}
	}

	for i := 0; i < len(c.HStore); i++ {
		if len(c.HStore[i].DataDir) == 0 {
			c.HStore[i].DataDir = ServerDefaultDataDir
		}
	}

	return nil
}

type GlobalCfg struct {
	User    string `yaml:"user"`
	KeyPath string `yaml:"key_path"`
	//Password     string       `yaml:"-"`
	//RemoteCfgPath string       `yaml:"remote_config_path"`
	//DataDir       string       `yaml:"data_dir"`
	SshPort          int          `yaml:"ssh_port" default:"22"`
	MetaReplica      int          `yaml:"meta_replica" default:"3"`
	HStoreConfigPath string       `yaml:"hstore_config_path"`
	ContainerCfg     ContainerCfg `yaml:"container_config"`
}

type ContainerCfg struct {
	Cpu            string `yaml:"cpu_limit"`
	Memory         string `yaml:"memory_limit"`
	RemoveWhenExit bool   `yaml:"remove_when_exit"`
	DisableRestart bool   `yaml:"disable_restart"`
}

func (c ContainerCfg) GetCmd() string {
	fmt.Printf("ContainerCfg: %+v\n", c)
	args := make([]string, 0, 4)
	if !c.DisableRestart {
		args = append(args, "--restart unless-stopped")
	}
	if c.RemoveWhenExit {
		args = append(args, "--rm")
	}
	if len(c.Cpu) != 0 {
		args = append(args, fmt.Sprintf("--cpus=%s", c.Cpu))
	}
	if len(c.Memory) != 0 {
		args = append(args, fmt.Sprintf("--memory=%s", c.Memory))
	}
	return strings.Join(args, " ")
}
