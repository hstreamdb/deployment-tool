package spec

import (
	"fmt"
	"github.com/creasty/defaults"
	"reflect"
	"strings"
)

const DefaultStoreConfigPath = "/logdevice.conf"

var (
	globalCfgTypeName = reflect.TypeOf(GlobalCfg{}).Name()
)

type ComponentsSpec struct {
	Global          GlobalCfg             `yaml:"global"`
	Monitor         MonitorSpec           `yaml:"monitor"`
	HServer         []HServerSpec         `yaml:"hserver"`
	HStore          []HStoreSpec          `yaml:"hstore"`
	MetaStore       []MetaStoreSpec       `yaml:"meta_store"`
	Prometheus      []PrometheusSpec      `yaml:"prometheus"`
	Grafana         []GrafanaSpec         `yaml:"grafana"`
	AlertManager    []AlertManagerSpec    `yaml:"alertmanager"`
	HStreamExporter []HStreamExporterSpec `yaml:"hstream_exporter"`
	HttpServer      []HttpServerSpec      `yaml:"http_server"`
	ElasticSearch   []ElasticSearchSpec   `yaml:"elasticsearch"`
	Kibana          []KibanaSpec          `yaml:"kibana"`
	Filebeat        []FilebeatSpec        `yaml:"filebeat"`
}

func (c *ComponentsSpec) GetHosts() []string {
	v := reflect.Indirect(reflect.ValueOf(c))
	t := v.Type()

	res := []string{}
	for i := 0; i < t.NumField(); i++ {
		field := v.Field(i)
		if field.Type().Name() == globalCfgTypeName {
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
		host := v.FieldByName("Host")
		if !host.IsValid() {
			return res
		}
		res = append(res, host.String())
	}
	return res
}

func (c *ComponentsSpec) GetMetaStoreUrl() (string, MetaStoreType, error) {
	hosts := []string{}
	for _, spc := range c.MetaStore {
		hosts = append(hosts, spc.Host)
	}
	if len(hosts) == 0 {
		return "", Unknown, nil
	}

	var url string
	tp := GetMetaStoreType(c.MetaStore[0].Image)
	switch tp {
	case ZK:
		url = getZkUrl(c.MetaStore)
	case RQLITE:
		url = getRqliteUrl(c.MetaStore)
	case Unknown:
		return "", Unknown, fmt.Errorf("unknown meta store type")
	}
	return url, tp, nil
}

func getZkUrl(metaStore []MetaStoreSpec) string {
	hosts := []string{}
	for _, spc := range metaStore {
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

func getRqliteUrl(metaStore []MetaStoreSpec) string {
	hosts := []string{}
	for _, spec := range metaStore {
		hosts = append(hosts, fmt.Sprintf("http://%s:%d", spec.Host, spec.Port))
	}
	return strings.Join(hosts, ",")
}

func (c *ComponentsSpec) GetHServerUrl() string {
	hosts := []string{}
	for _, spec := range c.HServer {
		hosts = append(hosts, fmt.Sprintf("%s:%d", spec.Host, spec.Port))
	}
	return strings.Join(hosts, ",")
}

func (c *ComponentsSpec) GetHttpServerUrl() []string {
	hosts := []string{}
	for _, spec := range c.HttpServer {
		hosts = append(hosts, fmt.Sprintf("%s:%d", spec.Host, spec.Port))
	}
	return hosts
}

func (c *ComponentsSpec) GetHStreamExporterAddr() []string {
	hosts := []string{}
	for _, spec := range c.HStreamExporter {
		hosts = append(hosts, fmt.Sprintf("%s:%d", spec.Host, spec.Port))
	}
	return hosts
}

func (c *ComponentsSpec) GetAlertManagerAddr() []string {
	hosts := []string{}
	for _, spec := range c.AlertManager {
		hosts = append(hosts, fmt.Sprintf("%s:%d", spec.Host, spec.Port))
	}
	return hosts
}

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

	if err := updateComponentSpecWithGlobal(c.Global, c); err != nil {
		return err
	}

	if len(c.Monitor.NodeExporterImage) == 0 {
		c.Monitor.NodeExporterImage = NodeExporterDefaultImage
	}
	if len(c.Monitor.CadvisorImage) == 0 {
		c.Monitor.CadvisorImage = CadvisorDefaultImage
	}

	return nil
}

type GlobalCfg struct {
	User                       string       `yaml:"user"`
	KeyPath                    string       `yaml:"key_path"`
	SSHPort                    int          `yaml:"ssh_port" default:"22"`
	MetaReplica                int          `yaml:"meta_replica" default:"3"`
	MetaStoreConfigPath        string       `yaml:"meta_store_config_path"`
	HStoreConfigPath           string       `yaml:"hstore_config_path"`
	HServerConfigPath          string       `yaml:"hserver_config_path"`
	DisableStoreNetworkCfgPath bool         `yaml:"disable_store_network_config_path"`
	ContainerCfg               ContainerCfg `yaml:"container_config"`
}

type ContainerCfg struct {
	Cpu            string `yaml:"cpu_limit"`
	Memory         string `yaml:"memory_limit"`
	RemoveWhenExit bool   `yaml:"remove_when_exit"`
	DisableRestart bool   `yaml:"disable_restart"`
}

func (c ContainerCfg) GetCmd() string {
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

type MetaStoreType uint

const (
	ZK MetaStoreType = iota
	RQLITE
	Unknown
)

func (m MetaStoreType) String() string {
	switch m {
	case ZK:
		return "zookeeper"
	case RQLITE:
		return "rqlite"
	case Unknown:
		return "unknown"
	}
	return ""
}

func GetMetaStoreType(image string) MetaStoreType {
	if strings.Contains(image, "zookeeper") {
		return ZK
	} else if strings.Contains(image, "rqlite") {
		return RQLITE
	}
	return Unknown
}

func updateComponentSpecWithGlobal(globalCfg GlobalCfg, data interface{}) error {
	v := reflect.ValueOf(data).Elem()
	t := v.Type()

	var err error
	for i := 0; i < t.NumField(); i++ {
		if err = updateComponent(globalCfg, v.Field(i)); err != nil {
			return err
		}
	}
	return nil
}

func updateComponent(cfg GlobalCfg, field reflect.Value) error {
	if skipUpdate(field) {
		return nil
	}

	switch field.Kind() {
	case reflect.Slice:
		for i := 0; i < field.Len(); i++ {
			if err := updateComponent(cfg, field.Index(i)); err != nil {
				return err
			}
		}
	case reflect.Struct:
		if field.Type().Name() == "ContainerCfg" {
			newCfg := MergeContainerCfg(cfg.ContainerCfg, field.Interface().(ContainerCfg))
			field.Set(reflect.ValueOf(newCfg))
			return nil
		}

		ref := reflect.New(field.Type())
		ref.Elem().Set(field)
		if err := updateComponentSpecWithGlobal(cfg, ref.Interface()); err != nil {
			return err
		}
		field.Set(ref.Elem())
	case reflect.Ptr:
		if err := updateComponent(cfg, field.Elem()); err != nil {
			return err
		}
	}

	if field.Kind() != reflect.Struct {
		return nil
	}

	for j := 0; j < field.NumField(); j++ {
		switch field.Type().Field(j).Name {
		case "SSHPort":
			if field.Field(j).Int() != 0 {
				continue
			}
			field.Field(j).Set(reflect.ValueOf(cfg.SSHPort))
		case "DataDir":
			if len(field.Field(j).String()) != 0 {
				continue
			}
			ref := reflect.New(field.Type())
			ref.Elem().Set(field)
			fn := ref.MethodByName("SetDefaultDataDir")
			fn.Call(nil)
			field.Field(j).Set(ref.Elem().FieldByName("DataDir"))
		case "RemoteCfgPath":
			if len(field.Field(j).String()) != 0 {
				continue
			}
			ref := reflect.New(field.Type())
			ref.Elem().Set(field)
			fn := ref.MethodByName("SetDefaultRemoteCfgPath")
			fn.Call(nil)
			field.Field(j).Set(ref.Elem().FieldByName("RemoteCfgPath"))
		case "Image":
			if len(field.Field(j).String()) != 0 {
				continue
			}
			ref := reflect.New(field.Type())
			ref.Elem().Set(field)
			fn := ref.MethodByName("SetDefaultImage")
			fn.Call(nil)
			field.Field(j).Set(ref.Elem().FieldByName("Image"))
		case "Address":
			if len(field.Field(j).String()) != 0 {
				continue
			}
			host := reflect.Indirect(field).FieldByName("Host").String()
			field.Field(j).Set(reflect.ValueOf(host))
		}
	}
	return nil
}

func skipUpdate(field reflect.Value) bool {
	tp := field.Type().Name()
	return tp == globalCfgTypeName
}
