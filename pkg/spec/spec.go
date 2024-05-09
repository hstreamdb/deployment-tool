package spec

import (
	"fmt"
	"github.com/creasty/defaults"
	log "github.com/sirupsen/logrus"
	"os"
	"reflect"
	"strings"
)

const DefaultStoreConfigPath = "/logdevice.conf"

var (
	globalCfgTypeName = reflect.TypeOf(GlobalCfg{}).Name()
)

// ComponentsSpec map config.yaml to a struct
type ComponentsSpec struct {
	Global          GlobalCfg             `yaml:"global"`
	Monitor         MonitorSpec           `yaml:"monitor"`
	HServer         []HServerSpec         `yaml:"hserver"`
	HStore          []HStoreSpec          `yaml:"hstore"`
	HAdmin          []HAdminSpec          `yaml:"hadmin"`
	MetaStore       []MetaStoreSpec       `yaml:"meta_store"`
	HStreamConsole  []HStreamConsoleSpec  `yaml:"hstream_console"`
	BlackBox        []BlackBoxSpec        `yaml:"blackBox"`
	Prometheus      []PrometheusSpec      `yaml:"prometheus"`
	Grafana         []GrafanaSpec         `yaml:"grafana"`
	AlertManager    []AlertManagerSpec    `yaml:"alertmanager"`
	HStreamExporter []HStreamExporterSpec `yaml:"hstream_exporter"`
	ElasticSearch   []ElasticSearchSpec   `yaml:"elasticsearch"`
	Kibana          []KibanaSpec          `yaml:"kibana"`
	Filebeat        []FilebeatSpec        `yaml:"filebeat"`
	Vector          []VectorSpec          `yaml:"vector"`
}

func (c *ComponentsSpec) UnmarshalYAML(unmarshal func(interface{}) error) error {
	// avoid recursion unmarshal
	type tmpSpec ComponentsSpec
	if err := unmarshal((*tmpSpec)(c)); err != nil {
		return err
	}

	if c.Global.EnableKafka {
		for i := 0; i < len(c.HServer); i++ {
			if defaults.CanUpdate(c.HServer[i].Port) {
				c.HServer[i].Port = 9092
			}
		}

		for i := 0; i < len(c.HStreamConsole); i++ {
			c.HStreamConsole[i].UseKafkaConsole = true
		}
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

	checkConflictAdminPort(c.HStore, c.HAdmin)
	return nil
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

// GetAddress return all services address. e.g. {hserver: [127.0.0.1:6570, 127.0.0.2:6570]}
func (c *ComponentsSpec) GetAddress() map[string][]string {
	v := reflect.Indirect(reflect.ValueOf(c))
	t := v.Type()

	res := map[string][]string{}
	for i := 0; i < t.NumField(); i++ {
		field := v.Field(i)
		tag := t.Field(i).Tag
		service := tag.Get("yaml")
		if field.Type().Name() == globalCfgTypeName {
			continue
		}
		res[service] = getAddrInner(field)
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

func (c *ComponentsSpec) GetHServerUrl() string {
	hosts := []string{}
	for _, spec := range c.HServer {
		hosts = append(hosts, fmt.Sprintf("%s:%d", spec.Host, spec.Port))
	}
	return strings.Join(hosts, ",")
}

func (c *ComponentsSpec) GetHServerEndpoint() string {
	endpoints := []string{}
	for _, spec := range c.HServer {
		endpoints = append(endpoints, fmt.Sprintf("%s:%d", spec.Host, spec.Port))
		if len(spec.AdvertisedListener) != 0 {
			listeners := strings.Split(spec.AdvertisedListener, ",")
			for _, listener := range listeners {
				parts := strings.Split(listener, "hstream://")
				if len(parts) != 2 {
					log.Errorf("invalied advertised listener: %s", listener)
					os.Exit(1)
				}
				endpoints = append(endpoints, parts[1])
			}
		}
	}
	return strings.Join(endpoints, ",")
}

func (c *ComponentsSpec) GetHServerMonitorEndpoint() []string {
	if !c.Global.EnableKafka {
		log.Errorf("Cannot get monitor endpoints when disable kafka")
		os.Exit(1)
	}

	endpoints := make([]string, 0, len(c.HServer))
	for _, v := range c.HServer {
		endpoints = append(endpoints, fmt.Sprintf("%s:%d", v.Host, v.MonitorPort))
	}
	return endpoints
}

func (c *ComponentsSpec) GetHStreamExporterAddr() []string {
	hosts := []string{}
	for _, spec := range c.HStreamExporter {
		hosts = append(hosts, fmt.Sprintf("%s:%d", spec.Host, spec.Port))
	}
	return hosts
}

func (c *ComponentsSpec) GetPrometheusAddr() []string {
	hosts := []string{}
	for _, spec := range c.Prometheus {
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

// GetMetaStoreType check docker image and return the proper meta store type
func GetMetaStoreType(image string) MetaStoreType {
	if strings.Contains(image, "zookeeper") {
		return ZK
	} else if strings.Contains(image, "rqlite") {
		return RQLITE
	}
	return Unknown
}

// =================================================================================

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

func getAddrInner(v reflect.Value) []string {
	t := v.Type()
	res := []string{}
	switch t.Kind() {
	case reflect.Slice:
		for i := 0; i < v.Len(); i++ {
			if hosts := getAddrInner(v.Index(i)); hosts != nil {
				res = append(res, hosts...)
			}
		}
	case reflect.Struct:
		host := v.FieldByName("Host")
		if !host.IsValid() {
			return res
		}
		port := v.FieldByName("Port")
		if !port.IsValid() {
			return res
		}
		res = append(res, fmt.Sprintf("%s:%d", host.String(), port.Int()))
	}
	return res
}

func getZkUrl(metaStore []MetaStoreSpec) string {
	hosts := []string{}
	for _, spc := range metaStore {
		hosts = append(hosts, fmt.Sprintf("%s:%d", spc.Host, spc.Port))
	}
	return strings.Join(hosts, ",")
}

func getRqliteUrl(metaStore []MetaStoreSpec) string {
	hosts := []string{}
	for _, spec := range metaStore {
		hosts = append(hosts, fmt.Sprintf("http://%s:%d", spec.Host, spec.Port))
	}
	return strings.Join(hosts, ",")
}

func checkConflictAdminPort(store []HStoreSpec, admin []HAdminSpec) {
	if len(admin) == 0 {
		return
	}

	adminAddress := make(map[string]struct{})
	for _, v := range store {
		if v.EnableAdmin {
			addr := fmt.Sprintf("%s:%d", v.Host, v.Port)
			adminAddress[addr] = struct{}{}
		}
	}

	for _, v := range admin {
		addr := fmt.Sprintf("%s:%d", v.Host, v.Port)
		if _, ok := adminAddress[addr]; ok {
			log.Errorf("there is a store instance monitor on %s:%d, use another admin port for hadmin",
				v.Host, v.Port)
			os.Exit(1)
		}
	}
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

		// Update meta store default port values according to image type
		if field.Type().Name() == "MetaStoreSpec" && field.FieldByName("Port").Int() == 0 {
			image := field.FieldByName("Image").String()
			if strings.Contains(image, "rqlite") {
				field.FieldByName("Port").SetInt(4001)
			} else {
				field.FieldByName("Port").SetInt(2181)
			}
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
			fn := ref.MethodByName("SetDataDir")
			fn.Call([]reflect.Value{reflect.ValueOf(cfg.HStreamPathPrefix)})
			field.Field(j).Set(ref.Elem().FieldByName("DataDir"))
		case "RemoteCfgPath":
			if len(field.Field(j).String()) != 0 {
				continue
			}
			ref := reflect.New(field.Type())
			ref.Elem().Set(field)
			fn := ref.MethodByName("SetRemoteCfgPath")
			fn.Call([]reflect.Value{reflect.ValueOf(cfg.HStreamPathPrefix)})
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
		case "AdvertisedAddress":
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
	tp := reflect.TypeOf(field).Name()
	return tp == globalCfgTypeName
}
