package edgemanager

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/ICKelin/cframe/pkg/etcdstorage"
	log "github.com/ICKelin/cframe/pkg/logs"
)

var (
	// support 3 minutes storage lease
	defaultLease    = time.Minute * 3
	recycleInterval = time.Second * 30

	defaultEdgeHostManager *EdgeHostManager
	edgeHostPrefix         = "/host"
)

type EdgeHost struct {
	IP        string `json:"ip"`
	EdgeName  string `json:"edge_name"`
	expiredIn time.Time
}

func (h *EdgeHost) String() string {
	return h.IP
}

type EdgeHostManager struct {
	store *etcdstorage.Etcd
}

func NewEdgeHostManager(store *etcdstorage.Etcd) *EdgeHostManager {
	if defaultEdgeHostManager != nil {
		return defaultEdgeHostManager
	}

	m := &EdgeHostManager{
		store: store,
	}

	defaultEdgeHostManager = m
	return defaultEdgeHostManager
}

func (m *EdgeHostManager) AddEdgeHost(edge *Edge, host *EdgeHost) {
	key := fmt.Sprintf("%s/%s/%s", edgeHostPrefix, edge.Name, host.String())
	host.EdgeName = edge.Name
	err := m.store.SetWithExpiration(key, host, defaultLease)
	if err != nil {
		log.Error("set edge host %s fail: %v", key, err)
	}
}

func (m *EdgeHostManager) DelEdgeHost(edge *Edge, host *EdgeHost) {
	key := fmt.Sprintf("%s/%s/%s", edgeHostPrefix, edge.Name, host.String())
	m.store.Del(key)
}

func (m *EdgeHostManager) GetEdgeHosts() []*EdgeHost {
	res, err := m.store.List(edgeHostPrefix)
	if err != nil {
		log.Error("list %s fail: %v", edgeHostPrefix, err)
		return nil
	}

	hosts := make([]*EdgeHost, 0)
	for _, v := range res {
		host := EdgeHost{}
		err := json.Unmarshal([]byte(v), &host)
		if err != nil {
			log.Error("unmarshal to host fail: %v", err)
			continue
		}
		hosts = append(hosts, &host)
	}
	return hosts
}

func AddedgeHost(edge *Edge, host *EdgeHost) {
	if defaultEdgeHostManager != nil {
		defaultEdgeHostManager.AddEdgeHost(edge, host)
	}
}

func DelEdgeHost(edge *Edge, host *EdgeHost) {
	if defaultEdgeHostManager != nil {
		defaultEdgeHostManager.DelEdgeHost(edge, host)
	}
}

func GetEdgeHosts() []*EdgeHost {
	if defaultEdgeHostManager != nil {
		return defaultEdgeHostManager.GetEdgeHosts()
	}
	return nil
}
