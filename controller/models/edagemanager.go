package models

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/ICKelin/cframe/codec"
	"github.com/ICKelin/cframe/pkg/etcdstorage"
	"github.com/ICKelin/cframe/pkg/ip"
	log "github.com/ICKelin/cframe/pkg/logs"
	"github.com/coreos/etcd/clientv3"
)

var (
	defaultEdgeManager *EdgeManager
	edgePrefix         = "/edges/"
)

type EdgeManager struct {
	storage *etcdstorage.Etcd
}

func NewEdgeManager(store *etcdstorage.Etcd) *EdgeManager {
	return &EdgeManager{
		storage: store,
	}
}

func (m *EdgeManager) Watch(delfunc, putfunc func(appId string, edge *codec.Edge)) {
	chs := m.storage.Watch(edgePrefix)
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
					edge := codec.Edge{}
					err := json.Unmarshal(evt.PrevKv.Value, &edge)
					if err != nil {
						log.Info("json unmarshal fail: %v", err)
						continue
					}

					delfunc(appId, &edge)
				}

			case clientv3.EventTypePut:
				if putfunc != nil {
					edge := codec.Edge{}
					err := json.Unmarshal(evt.Kv.Value, &edge)
					if err != nil {
						log.Info("json unmarshal fail: %v", err)
						continue
					}

					putfunc(appId, &edge)
				}
			}
		}
	}

}

func (m *EdgeManager) AddEdge(appId string, edge *codec.Edge) {
	key := fmt.Sprintf("%s%s/%s", edgePrefix, appId, edge.Name)
	e := m.storage.Set(key, edge)
	if e != nil {
		log.Error("add edge fail: %v", e)
	}
}

func (m *EdgeManager) DelEdge(appId, name string) {
	key := fmt.Sprintf("%s%s/%s", edgePrefix, appId, name)
	m.storage.Del(key)
}

func (m *EdgeManager) GetEdge(appId, name string) *codec.Edge {
	key := fmt.Sprintf("%s%s/%s", edgePrefix, appId, name)
	edg := codec.Edge{}
	err := m.storage.Get(key, &edg)
	if err != nil {
		return nil
	}
	return &edg
}

func (m *EdgeManager) GetEdges(appId string) []*codec.Edge {
	key := fmt.Sprintf("%s%s", edgePrefix, appId)
	res, err := m.storage.List(key)
	if err != nil {
		log.Error("list %s fail: %v", edgePrefix, err)
		return nil
	}

	edges := make([]*codec.Edge, 0)
	for _, val := range res {
		edge := codec.Edge{}
		err := json.Unmarshal([]byte(val), &edge)
		if err != nil {
			log.Error("unmarshal to edge fail: %v", err)
			continue
		}
		edges = append(edges, &edge)
	}
	return edges
}

func (m *EdgeManager) VerifyCidr(cidr string) bool {
	b := true
	return b
}

// VerifyConflict verify cidr1 and cidr2 ip range
// [bip1, eip1], [bip2, eip2]
// bip1 < bip2 < eip1
// bip1 < eip2 < eip1 or
// bip2 < bip1 < eip2
// bip2 < eip1 < eip2
func (m *EdgeManager) verifyConflict(cidr1, cidr2 string) bool {
	sp := strings.Split(cidr1, "/")
	if len(sp) != 2 {
		log.Error("invalid cidr format: %s", cidr1)
		return false
	}

	ip1, sprefix1 := sp[0], sp[1]
	prefix1, err := strconv.Atoi(sprefix1)
	if err != nil {
		log.Error("invalid cidr format: %s", cidr1)
		return false
	}

	ipv41, err := ip.ParseIP4(ip1)
	if err != nil {
		log.Error("invalid cidr format: %s", cidr1)
		return false
	}

	sp = strings.Split(cidr2, "/")
	if len(sp) != 2 {
		log.Error("invalid cidr format: %s", cidr2)
		return false
	}

	ip2, sprefix2 := sp[0], sp[1]
	prefix2, err := strconv.Atoi(sprefix2)
	if err != nil {
		log.Error("invalid cidr format: %s", cidr2)
		return false
	}

	ipv42, err := ip.ParseIP4(ip2)
	if err != nil {
		log.Error("invalid cidr format: %s", cidr2)
		return false
	}

	ipn1 := ip.IP4Net{
		IP:        ipv41,
		PrefixLen: uint(prefix1),
	}
	ipn2 := ip.IP4Net{
		IP:        ipv42,
		PrefixLen: uint(prefix2),
	}

	return !ipn1.Overlaps(ipn2)
}
