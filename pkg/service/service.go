package service

import (
	"encoding/json"
	"fmt"
	"github.com/hstreamdb/deployment-tool/pkg/executor"
	"github.com/hstreamdb/deployment-tool/pkg/spec"
	"github.com/hstreamdb/deployment-tool/pkg/utils"
	"golang.org/x/exp/slices"
	"os"
	"reflect"
	"sort"
	"strconv"
	"strings"
)

type Service interface {
	InitEnv(cfg *GlobalCtx) *executor.ExecuteCtx
	Deploy(cfg *GlobalCtx) *executor.ExecuteCtx
	Stop(cfg *GlobalCtx) *executor.ExecuteCtx
	Remove(cfg *GlobalCtx) *executor.ExecuteCtx
	SyncConfig(cfg *GlobalCtx) *executor.TransferCtx
	GetServiceName() string
}

type GlobalCtx struct {
	User                 string
	KeyPath              string
	SSHPort              int
	MetaReplica          int
	EnableGrpcHs         bool
	RemoteCfgPath        string
	DataDir              string
	EnableDscpReflection bool
	containerCfg         spec.ContainerCfg

	Hosts     []string
	SeedNodes string
	// for zk: host1:2181,host2:2181
	// for rqlite: http://host1:4001,http://host2:4001
	MetaStoreUrls string
	MetaStoreType spec.MetaStoreType
	// total count of meta store instances
	MetaStoreCount int
	// for zk: use zkUrl
	HStoreConfigInMetaStore string
	// the origin meta store config file in local
	LocalMetaStoreConfigFile string
	// the origin store config file in local
	LocalHStoreConfigFile string
	// the origin server config file in local
	LocalHServerConfigFile string
	// the origin elastic search config file in local
	LocalEsConfigFile string
	HAdminInfos       []AdminInfo
	HStreamServerUrls string
	HServerEndPoints  string
	PrometheusUrls    []string
	HttpServerUrls    []string
	ServiceAddr       map[string][]string
}

func newGlobalCtx(c spec.ComponentsSpec, hosts []string) (*GlobalCtx, error) {
	metaStoreUrl, metaStoreTp, err := c.GetMetaStoreUrl()
	if err != nil {
		return nil, err
	}

	admins := getAdminInfos(c)
	if len(admins) == 0 {
		return nil, fmt.Errorf("need at least one hadmin node")
	}
	cfgInMetaStore := ""
	if !c.Global.DisableStoreNetworkCfgPath && metaStoreTp == spec.ZK {
		cfgInMetaStore = "zk:" + metaStoreUrl + spec.DefaultStoreConfigPath
	}

	hserverUrl := c.GetHServerUrl()
	httpServerUrl := c.GetHttpServerUrl()
	hserverEndpoints := c.GetHServerEndpoint()
	prometheusUrls := c.GetPrometheusAddr()
	// all service address in `host:port` form, except cadvisor and node-exporter
	serviceAddr := c.GetAddress()

	return &GlobalCtx{
		User:                 c.Global.User,
		KeyPath:              c.Global.KeyPath,
		SSHPort:              c.Global.SSHPort,
		MetaReplica:          c.Global.MetaReplica,
		EnableGrpcHs:         c.Global.EnableHsGrpc,
		EnableDscpReflection: c.Global.EnableDscpReflection,
		containerCfg:         c.Global.ContainerCfg,

		Hosts:                    hosts,
		MetaStoreUrls:            metaStoreUrl,
		MetaStoreType:            metaStoreTp,
		MetaStoreCount:           len(c.MetaStore),
		HStoreConfigInMetaStore:  cfgInMetaStore,
		LocalMetaStoreConfigFile: c.Global.MetaStoreConfigPath,
		LocalHStoreConfigFile:    c.Global.HStoreConfigPath,
		LocalHServerConfigFile:   c.Global.HServerConfigPath,
		LocalEsConfigFile:        c.Global.EsConfigPath,
		HAdminInfos:              admins,
		HStreamServerUrls:        hserverUrl,
		HServerEndPoints:         hserverEndpoints,
		PrometheusUrls:           prometheusUrls,
		HttpServerUrls:           httpServerUrl,
		ServiceAddr:              serviceAddr,
	}, nil
}

type Services struct {
	Global          *GlobalCtx
	MonitorSuite    []*MonitorSuite
	HServer         []*HServer
	HStore          []*HStore
	HAdmin          []*HAdmin
	MetaStore       []*MetaStore
	HStreamConsole  []*HStreamConsole
	BlackBox        []*BlackBox
	Prometheus      []*Prometheus
	Grafana         []*Grafana
	AlertManager    []*AlertManager
	HStreamExporter []*HStreamExporter
	HttpServer      []*HttpServer
	ElasticSearch   []*ElasticSearch
	Kibana          []*Kibana
	Filebeat        []*Filebeat
}

func NewServices(c spec.ComponentsSpec) (*Services, error) {
	seedNodes := make([]string, 0, len(c.HServer))
	hserver := make([]*HServer, 0, len(c.HServer))
	for idx, v := range c.HServer {
		hserver = append(hserver, NewHServer(uint32(idx+1), v))
		seedNodes = append(seedNodes, fmt.Sprintf("%s:%d", v.Host, v.InternalPort))
	}

	hadmin := make([]*HAdmin, 0, len(c.HAdmin))
	for idx, v := range c.HAdmin {
		hadmin = append(hadmin, NewHAdmin(uint32(idx+1), v))
	}

	hstore := make([]*HStore, 0, len(c.HStore))
	for idx, v := range c.HStore {
		hstore = append(hstore, NewHStore(uint32(idx+1), v))
	}

	metaStore := make([]*MetaStore, 0, len(c.MetaStore))
	for idx, v := range c.MetaStore {
		metaStore = append(metaStore, NewMetaStore(uint32(idx+1), v))
	}

	hstreamConsole := make([]*HStreamConsole, 0, len(c.HStreamConsole))
	for idx, v := range c.HStreamConsole {
		hstreamConsole = append(hstreamConsole, NewHStreamConsole(uint32(idx+1), v))
	}

	hosts := c.GetHosts()
	sort.Strings(hosts)
	hosts = slices.Compact(hosts)
	monitorSuites := make([]*MonitorSuite, 0, len(hosts))
	excludedHosts := getExcludedMonitorHosts(c)
	for _, host := range hosts {
		if slices.Contains(excludedHosts, host) {
			continue
		}
		monitorSuites = append(monitorSuites, NewMonitorSuite(host, c.Monitor))
	}

	httpServer := make([]*HttpServer, 0, len(c.HttpServer))
	for idx, v := range c.HttpServer {
		httpServer = append(httpServer, NewHttpServer(uint32(idx+1), v))
	}

	hstreamExporter := make([]*HStreamExporter, 0, len(c.HStreamExporter))
	for _, v := range c.HStreamExporter {
		hstreamExporter = append(hstreamExporter, NewHStreamExporter(v))
	}

	blackBox := make([]*BlackBox, 0, len(c.BlackBox))
	for _, v := range c.BlackBox {
		blackBox = append(blackBox, NewBlackBox(v))
	}

	alertManager := make([]*AlertManager, 0, len(c.AlertManager))
	for _, v := range c.AlertManager {
		alertManager = append(alertManager, NewAlertManager(v))
	}

	blackBoxAddr := ""
	if len(blackBox) != 0 {
		blackBoxAddr = fmt.Sprintf("%s:%d", blackBox[0].spec.Host, blackBox[0].spec.Port)
	}
	proms := make([]*Prometheus, 0, len(c.Prometheus))
	for _, v := range c.Prometheus {
		proms = append(proms, NewPrometheus(v, monitorSuites, c.GetHStreamExporterAddr(), c.GetAlertManagerAddr(), blackBoxAddr))
	}

	grafana := make([]*Grafana, 0, len(c.Grafana))
	for _, v := range c.Grafana {
		grafana = append(grafana, NewGrafana(v, c.Monitor.GrafanaDisableLogin))
	}

	elasticSearch := make([]*ElasticSearch, 0, len(c.ElasticSearch))
	for _, v := range c.ElasticSearch {
		elasticSearch = append(elasticSearch, NewElasticSearch(v))
	}

	kibana := make([]*Kibana, 0, len(c.Kibana))
	if len(elasticSearch) != 0 {
		for _, v := range c.Kibana {
			kibana = append(kibana, NewKibana(v, elasticSearch[0].spec.Host, elasticSearch[0].spec.Port))
		}
	}

	filebeat := make([]*Filebeat, 0, len(c.Filebeat))
	if len(elasticSearch) != 0 {
		for _, v := range c.Filebeat {
			filebeat = append(filebeat, NewFilebeat(v,
				elasticSearch[0].spec.Host, strconv.Itoa(elasticSearch[0].spec.Port),
				kibana[0].spec.Host, strconv.Itoa(kibana[0].spec.Port),
			))
		}
	}

	globalCtx, err := newGlobalCtx(c, hosts)
	if err != nil {
		return nil, err
	}

	configPath, err := updateStoreConfig(globalCtx)
	if err != nil {
		return nil, fmt.Errorf("update store config file err: %s", err.Error())
	}
	globalCtx.LocalHStoreConfigFile = configPath

	globalCtx.SeedNodes = strings.Join(seedNodes, ",")

	return &Services{
		Global:          globalCtx,
		MonitorSuite:    monitorSuites,
		HServer:         hserver,
		HAdmin:          hadmin,
		HStore:          hstore,
		MetaStore:       metaStore,
		HStreamConsole:  hstreamConsole,
		BlackBox:        blackBox,
		Prometheus:      proms,
		Grafana:         grafana,
		AlertManager:    alertManager,
		HStreamExporter: hstreamExporter,
		HttpServer:      httpServer,
		ElasticSearch:   elasticSearch,
		Kibana:          kibana,
		Filebeat:        filebeat,
	}, nil
}

func (s *Services) ShowAllServices() {
	v := reflect.Indirect(reflect.ValueOf(s))
	t := v.Type()

	showedComponents := make(map[string][]utils.DisplayedComponent)
	for i := 0; i < t.NumField(); i++ {
		field := v.Field(i)
		if field.Type().Kind() != reflect.Slice {
			continue
		}

		for j := 0; j < field.Len(); j++ {
			service := field.Index(j)
			if service.Type().Kind() != reflect.Ptr {
				panic(fmt.Sprintf("Show all services error, unexpected service kind: %s", service.String()))
			}

			fn := service.MethodByName("Display")
			res := fn.Call(nil)
			displayedComponent := res[0].Interface().(map[string]utils.DisplayedComponent)
			for k, c := range displayedComponent {
				if _, ok := showedComponents[k]; !ok {
					showedComponents[k] = []utils.DisplayedComponent{}
				}
				showedComponents[k] = append(showedComponents[k], c)
			}
		}
	}
	utils.ShowComponents(showedComponents)
}

// updateStoreConfig update hstore config file and write the updated config
// file to template/logdevice.conf
func updateStoreConfig(ctx *GlobalCtx) (string, error) {
	configPath := "template/logdevice.conf"
	content, err := os.ReadFile(configPath)
	if err != nil {
		return "", err
	}
	cfg := &storeCfg{}
	if err = json.Unmarshal(content, cfg); err != nil {
		return "", err
	}
	cfg.updateLogReplic(ctx.MetaReplica)

	if cfg.Zookeeper != nil && cfg.Rqlite != nil {
		return "", fmt.Errorf("can't set both zookeeper and rqlite fields in config file")
	}

	switch ctx.MetaStoreType {
	case spec.ZK:
		cfg.Zookeeper = map[string]interface{}{
			"zookeeper_uri": "ip://" + ctx.MetaStoreUrls,
			"timeout":       "30s",
		}
		cfg.Rqlite = nil
	case spec.RQLITE:
		urls := strings.ReplaceAll(ctx.MetaStoreUrls, "http://", "")
		url := strings.Split(urls, ",")[0]
		cfg.Rqlite = map[string]interface{}{
			"rqlite_uri": "ip://" + url,
		}
		cfg.Zookeeper = nil
	}
	res, err := json.MarshalIndent(cfg, "", "\t")
	if err != nil {
		return "", err
	}

	if err = os.WriteFile(configPath, res, 0755); err != nil {
		return "", err
	}
	return configPath, nil
}

// FIXME: construct a struct to parse `replicate_across` field in
// internal_logs and metadata_logs
// storeCfg map a hstore config file to a struct
type storeCfg struct {
	ServerSettings map[string]interface{} `json:"server_settings,omitempty"`
	ClientSettings map[string]interface{} `json:"client_settings,omitempty"`
	Cluster        string                 `json:"cluster,omitempty"`
	InternalLogs   map[string]interface{} `json:"internal_logs,omitempty"`
	MetadataLogs   map[string]interface{} `json:"metadata_logs,omitempty"`
	Zookeeper      map[string]interface{} `json:"zookeeper,omitempty"`
	Rqlite         map[string]interface{} `json:"rqlite,omitempty"`
}

// FIXME: will panic if no replicate_across or node field exist.
func (s *storeCfg) updateLogReplic(replica int) {
	cfgValue := reflect.Indirect(reflect.ValueOf(s))
	for j := 0; j < cfgValue.NumField(); j++ {
		switch cfgValue.Type().Field(j).Name {
		case "InternalLogs":
			field := cfgValue.Field(j)
			v := reflect.Indirect(field)
			for item := v.MapRange(); item.Next(); {
				logCfg := item.Value().Elem()
				if logCfg.Kind() != reflect.Map {
					continue
				}
				replicateCfg := logCfg.MapIndex(reflect.ValueOf("replicate_across")).Elem()
				replicateCfg.SetMapIndex(reflect.ValueOf("node"), reflect.ValueOf(replica))
			}
			cfgValue.Field(j).Set(v)
		case "MetadataLogs":
			field := cfgValue.Field(j)
			v := reflect.Indirect(field)
			replicateCfg := v.MapIndex(reflect.ValueOf("replicate_across")).Elem()
			replicateCfg.SetMapIndex(reflect.ValueOf("node"), reflect.ValueOf(replica))
			cfgValue.Field(j).Set(v)
		}
	}
}

// getExcludedMonitorHosts get the hosts of all nodes which don't need to deploy
// a monitoring stack.
func getExcludedMonitorHosts(c spec.ComponentsSpec) []string {
	res := []string{}

	for _, host := range c.Monitor.ExcludedHosts {
		res = append(res, host)
	}
	//for _, sp := range c.Prometheus {
	//	res = append(res, sp.Host)
	//}
	//for _, sp := range c.Grafana {
	//	res = append(res, sp.Host)
	//}
	sort.Strings(res)
	return slices.Compact(res)
}

func getAdminInfos(c spec.ComponentsSpec) []AdminInfo {
	infos := []AdminInfo{}
	for _, v := range c.HAdmin {
		infos = append(infos, AdminInfo{
			Host:          v.Host,
			Port:          v.Port,
			ContainerName: spec.AdminDefaultContainerName,
		})
	}

	for _, v := range c.HStore {
		if v.EnableAdmin {
			infos = append(infos, AdminInfo{
				Host:          v.Host,
				Port:          v.Port,
				ContainerName: spec.StoreDefaultContainerName,
			})
		}
	}
	return infos
}
