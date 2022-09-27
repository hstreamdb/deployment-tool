package service

import (
	"fmt"
	"github.com/hstreamdb/deployment-tool/pkg/executor"
	"github.com/hstreamdb/deployment-tool/pkg/spec"
	"github.com/hstreamdb/deployment-tool/pkg/template/config"
	"golang.org/x/exp/slices"
	"sort"
	"strings"
)

type Service interface {
	InitEnv(cfg *GlobalCtx) *executor.ExecuteCtx
	Deploy(cfg *GlobalCtx) *executor.ExecuteCtx
	Remove(cfg *GlobalCtx) *executor.ExecuteCtx
	SyncConfig(cfg *GlobalCtx) *executor.TransferCtx
}

type GlobalCtx struct {
	User          string
	KeyPath       string
	SSHPort       int
	MetaReplica   int
	RemoteCfgPath string
	DataDir       string
	containerCfg  spec.ContainerCfg

	Hosts     []string
	SeedNodes string
	// for zk: host1:2181,host2:2181
	MetaStoreUrls string
	MetaStoreType spec.MetaStoreType
	// for zk: use zkUrl
	HStoreConfigInMetaStore string
	// the origin meta store config file in local
	LocalMetaStoreConfigFile string
	// the origin store config file in local
	LocalHStoreConfigFile string
	// the origin server config file in local
	LocalHServerConfigFile string
	HadminAddress          []string
}

func newGlobalCtx(c spec.ComponentsSpec, hosts []string) (*GlobalCtx, error) {
	metaStoreUrl, metaStoreTp, err := c.GetMetaStoreUrl()
	if err != nil {
		return nil, err
	}

	admins := make([]string, 0, len(c.HStore))
	for _, v := range c.HStore {
		if v.EnableAdmin {
			admins = append(admins, fmt.Sprintf("%s:%d", v.Host, v.AdminPort))
		}
	}
	if len(admins) == 0 {
		return nil, fmt.Errorf("need at least one hadmin node")
	}
	cfgInMetaStore := ""
	if !c.Global.DisableStoreNetworkCfgPath {
		cfgInMetaStore = "zk:" + metaStoreUrl + spec.DefaultStoreConfigPath
	}

	return &GlobalCtx{
		User:         c.Global.User,
		KeyPath:      c.Global.KeyPath,
		SSHPort:      c.Global.SSHPort,
		MetaReplica:  c.Global.MetaReplica,
		containerCfg: c.Global.ContainerCfg,

		Hosts:                    hosts,
		MetaStoreUrls:            metaStoreUrl,
		MetaStoreType:            metaStoreTp,
		HStoreConfigInMetaStore:  cfgInMetaStore,
		LocalMetaStoreConfigFile: c.Global.MetaStoreConfigPath,
		LocalHStoreConfigFile:    c.Global.HStoreConfigPath,
		LocalHServerConfigFile:   c.Global.HServerConfigPath,
		HadminAddress:            admins,
	}, nil
}

type Services struct {
	Global       *GlobalCtx
	MonitorSuite []*MonitorSuite
	HServer      []*HServer
	HStore       []*HStore
	MetaStore    []*MetaStore
	Prometheus   []*Prometheus
	Grafana      []*Grafana
}

func NewServices(c spec.ComponentsSpec) (*Services, error) {
	seedNodes := make([]string, 0, len(c.HServer))
	hserver := make([]*HServer, 0, len(c.HServer))
	for idx, v := range c.HServer {
		hserver = append(hserver, NewHServer(uint32(idx+1), v))
		seedNodes = append(seedNodes, fmt.Sprintf("%s:%d", v.Host, v.InternalPort))
	}

	hstore := make([]*HStore, 0, len(c.HStore))
	for idx, v := range c.HStore {
		hstore = append(hstore, NewHStore(uint32(idx+1), v))
	}

	metaStore := make([]*MetaStore, 0, len(c.MetaStore))
	for idx, v := range c.MetaStore {
		metaStore = append(metaStore, NewMetaStore(uint32(idx+1), v))
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

	proms := make([]*Prometheus, 0, len(c.Prometheus))
	for _, v := range c.Prometheus {
		proms = append(proms, NewPrometheus(v, monitorSuites))
	}

	grafana := make([]*Grafana, 0, len(c.Grafana))
	for _, v := range c.Grafana {
		grafana = append(grafana, NewGrafana(v, c.Monitor.GrafanaDisableLogin))
	}

	globalCtx, err := newGlobalCtx(c, hosts)
	if err != nil {
		return nil, err
	}

	if len(c.Global.HStoreConfigPath) == 0 {
		cfg := config.HStoreConfig{ZkUrl: "ip://" + globalCtx.MetaStoreUrls}
		configPath, err := cfg.GenConfig()
		if err != nil {
			return nil, fmt.Errorf("generate hstore config err: %s", err.Error())
		}
		globalCtx.LocalHStoreConfigFile = configPath
	}

	globalCtx.SeedNodes = strings.Join(seedNodes, ",")

	return &Services{
		Global:       globalCtx,
		MonitorSuite: monitorSuites,
		HServer:      hserver,
		HStore:       hstore,
		MetaStore:    metaStore,
		Prometheus:   proms,
		Grafana:      grafana,
	}, nil
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
