package models

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/ICKelin/cframe/codec"
	"github.com/ICKelin/cframe/pkg/etcdstorage"
	log "github.com/ICKelin/cframe/pkg/logs"
	"github.com/coreos/etcd/clientv3"
)

var (
	defaultRouteManager *RouteManager
	routePrefix         = "/routes/"
)

type RouteManager struct {
	storage *etcdstorage.Etcd
}

func NewRouteManager(store *etcdstorage.Etcd) *RouteManager {
	return &RouteManager{
		storage: store,
	}
}

func (m *RouteManager) Watch(delfunc, putfunc func(appId string, route *codec.Route)) {
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

			appId := sp[2]
			switch evt.Type {
			case clientv3.EventTypeDelete:
				if delfunc != nil {
					route := codec.Route{}
					err := json.Unmarshal(evt.PrevKv.Value, &route)
					if err != nil {
						log.Info("json unmarshal fail: %v", err)
						continue
					}

					delfunc(appId, &route)
				}

			case clientv3.EventTypePut:
				if putfunc != nil {
					route := codec.Route{}
					err := json.Unmarshal(evt.Kv.Value, &route)
					if err != nil {
						log.Info("json unmarshal fail: %v", err)
						continue
					}

					putfunc(appId, &route)
				}
			}
		}
	}

}

func (m *RouteManager) AddRoute(appId, route *codec.Route) error {
	key := fmt.Sprintf("%s%s/%s", routePrefix, appId, route.Name)
	return m.storage.Set(key, route)
}

func (m *RouteManager) DelRoute(appId, name string) error {
	key := fmt.Sprintf("%s%s/%s", routePrefix, appId, name)
	m.storage.Del(key)
	return nil
}

func (m *RouteManager) GetRoutes(appId string) []*codec.Route {
	key := fmt.Sprintf("%s%s", routePrefix, appId)
	res, err := m.storage.List(key)
	if err != nil {
		log.Error("list %s fail: %v", edgePrefix, err)
		return nil
	}

	routes := make([]*codec.Route, 0)
	for _, val := range res {
		r := codec.Route{}
		err := json.Unmarshal([]byte(val), &r)
		if err != nil {
			log.Error("unmarshal to edge fail: %v", err)
			continue
		}
		routes = append(routes, &r)
	}
	return routes
}
