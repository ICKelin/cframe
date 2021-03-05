package routemanager

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/ICKelin/cframe/pkg/etcdstorage"
	log "github.com/ICKelin/cframe/pkg/logs"
	"github.com/coreos/etcd/clientv3"
)

var (
	defaultRouteManager *RouteManager
	routePrefix         = "/route/"
)

type RouteManager struct {
	storage *etcdstorage.Etcd
}

func New(store *etcdstorage.Etcd) *RouteManager {
	if defaultRouteManager != nil {
		return defaultRouteManager
	}

	m := &RouteManager{
		storage: store,
	}
	defaultRouteManager = m
	return m
}

func (m *RouteManager) Watch(delfunc, putfunc func(userId string, route *Route)) {
	chs := m.storage.Watch(routePrefix)
	for c := range chs {
		for _, evt := range c.Events {
			log.Info("type: %v", evt.Type)
			log.Info("new: %v", evt.Kv)
			log.Info("old: %v", evt.PrevKv)
			sp := strings.Split(string(evt.Kv.Key), "/")

			if len(sp) < 3 {
				log.Warn("unsupported key value")
				continue
			}

			userId := sp[2]
			switch evt.Type {
			case clientv3.EventTypeDelete:
				if delfunc != nil {
					route := Route{}
					err := json.Unmarshal(evt.PrevKv.Value, &route)
					if err != nil {
						log.Info("json unmarshal fail: %v", err)
						continue
					}

					delfunc(userId, &route)
				}

			case clientv3.EventTypePut:
				if putfunc != nil {
					route := Route{}
					err := json.Unmarshal(evt.Kv.Value, &route)
					if err != nil {
						log.Info("json unmarshal fail: %v", err)
						continue
					}

					putfunc(userId, &route)
				}
			}
		}
	}

}

func (m *RouteManager) AddRoute(username, name string, route *Route) error {
	key := fmt.Sprintf("%s%s/%s", routePrefix, username, name)
	return m.storage.Set(key, route)
}

func (m *RouteManager) DelRoute(username, name string) error {
	key := fmt.Sprintf("%s%s/%s", routePrefix, username, name)
	m.storage.Del(key)
	return nil
}

func AddRoute(username, name string, route *Route) error {
	if defaultRouteManager == nil {
		return fmt.Errorf("route manager nil")
	}
	return defaultRouteManager.AddRoute(username, name, route)
}

func DelRoute(username, name string) error {
	if defaultRouteManager == nil {
		return fmt.Errorf("route manager nil")
	}
	return defaultRouteManager.DelRoute(username, name)
}
