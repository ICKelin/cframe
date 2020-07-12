package edagemanager

import (
	"fmt"
	"sync"
	"time"

	log "github.com/ICKelin/cframe/pkg/logs"
)

var (
	// support 3 minutes storage lease
	defaultLease    = time.Minute * 3
	recycleInterval = time.Second * 30

	defaultEdageHostManager *EdageHostManager
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
	mu    sync.Mutex
	table map[string]*EdageHost
}

func NewEdageHostManager() *EdageHostManager {
	if defaultEdageHostManager != nil {
		return defaultEdageHostManager
	}

	m := &EdageHostManager{
		table: make(map[string]*EdageHost),
	}

	defaultEdageHostManager = m
	go m.recycle()
	return defaultEdageHostManager
}

func (m *EdageHostManager) recycle() {
	tick := time.NewTicker(recycleInterval)
	defer tick.Stop()
	for range tick.C {
		log.Info("recycle checking")
		m.mu.Lock()
		for key, val := range m.table {
			if val.expiredIn.Before(time.Now()) {
				log.Info("delete expiration edage host: %s", key)
				delete(m.table, key)
			}
		}
		m.mu.Unlock()
	}
}

func (m *EdageHostManager) AddEdageHost(edage *Edage, host *EdageHost) {
	m.mu.Lock()
	defer m.mu.Unlock()
	key := fmt.Sprintf("%s-%s", edage.Name, host.String())
	host.expiredIn = time.Now().Add(defaultLease)
	host.EdageName = edage.Name
	log.Info("set edage host info %s", key)
	m.table[key] = host
}

func (m *EdageHostManager) DelEdageHost(edage *Edage, host *EdageHost) {
	m.mu.Lock()
	defer m.mu.Unlock()
	key := fmt.Sprintf("%s-%s", edage.Name, host.String())
	delete(m.table, key)
}

func (m *EdageHostManager) GetEdageHosts() []*EdageHost {
	m.mu.Lock()
	defer m.mu.Unlock()
	results := make([]*EdageHost, 0)
	for _, val := range m.table {
		results = append(results, val)
	}
	return results
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
