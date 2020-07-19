package edagemanager

import (
	"encoding/json"
	"fmt"
	"time"

	log "github.com/ICKelin/cframe/pkg/logs"
)

var (
	// support 3 minutes storage lease
	defaultLease    = time.Minute * 3
	recycleInterval = time.Second * 30

	defaultEdageHostManager *EdageHostManager
	edageHostPrefix         = "/host"
)

type EdageHost struct {
	IP        string `json:"ip"`
	EdageName string `json:"edage_name"`
	expiredIn time.Time
}

func (h *EdageHost) String() string {
	return h.IP
}

type EdageHostManager struct {
	store *EtcdStorage
}

func NewEdageHostManager(store *EtcdStorage) *EdageHostManager {
	if defaultEdageHostManager != nil {
		return defaultEdageHostManager
	}

	m := &EdageHostManager{
		store: store,
	}

	defaultEdageHostManager = m
	return defaultEdageHostManager
}

func (m *EdageHostManager) AddEdageHost(edage *Edage, host *EdageHost) {
	key := fmt.Sprintf("%s/%s/%s", edageHostPrefix, edage.Name, host.String())
	host.EdageName = edage.Name
	err := m.store.SetWithExpiration(key, host, defaultLease)
	if err != nil {
		log.Error("set edage host %s fail: %v", key, err)
	}
}

func (m *EdageHostManager) DelEdageHost(edage *Edage, host *EdageHost) {
	key := fmt.Sprintf("%s/%s/%s", edageHostPrefix, edage.Name, host.String())
	m.store.Del(key)
}

func (m *EdageHostManager) GetEdageHosts() []*EdageHost {
	res, err := m.store.List(edageHostPrefix)
	if err != nil {
		log.Error("list %s fail: %v", edageHostPrefix, err)
		return nil
	}

	hosts := make([]*EdageHost, 0)
	for _, v := range res {
		host := EdageHost{}
		err := json.Unmarshal([]byte(v), &host)
		if err != nil {
			log.Error("unmarshal to host fail: %v", err)
			continue
		}
		hosts = append(hosts, &host)
	}
	return hosts
}

func AddedageHost(edage *Edage, host *EdageHost) {
	if defaultEdageHostManager != nil {
		defaultEdageHostManager.AddEdageHost(edage, host)
	}
}

func DelEdageHost(edage *Edage, host *EdageHost) {
	if defaultEdageHostManager != nil {
		defaultEdageHostManager.DelEdageHost(edage, host)
	}
}

func GetEdageHosts() []*EdageHost {
	if defaultEdageHostManager != nil {
		return defaultEdageHostManager.GetEdageHosts()
	}
	return nil
}
