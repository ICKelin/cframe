package edgemanager

import (
	"encoding/json"
	"strconv"
	"strings"

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

func New(store *etcdstorage.Etcd) *EdgeManager {
	if defaultEdgeManager != nil {
		return defaultEdgeManager
	}

	m := &EdgeManager{
		storage: store,
	}
	defaultEdgeManager = m
	return m
}

func (m *EdgeManager) Watch(delfunc, putfunc func(edge *Edge)) {
	chs := m.storage.Watch(edgePrefix)
	for c := range chs {
		for _, evt := range c.Events {
			log.Info("type: %v", evt.Type)
			log.Info("new: %v", evt.Kv)
			log.Info("old: %v", evt.PrevKv)
			switch evt.Type {
			case clientv3.EventTypeDelete:
				if delfunc != nil {
					edge := Edge{}
					err := json.Unmarshal(evt.PrevKv.Value, &edge)
					if err != nil {
						log.Info("json unmarshal fail: %v", err)
						continue
					}

					delfunc(&edge)
				}

			case clientv3.EventTypePut:
				if putfunc != nil {
					edge := Edge{}
					err := json.Unmarshal(evt.Kv.Value, &edge)
					if err != nil {
						log.Info("json unmarshal fail: %v", err)
						continue
					}

					putfunc(&edge)
				}
			}
		}
	}

}

func (m *EdgeManager) AddEdge(name string, edge *Edge) {
	m.storage.Set(edgePrefix+name, edge)
}

func (m *EdgeManager) DelEdge(name string) {
	m.storage.Del(edgePrefix + name)
}

func (m *EdgeManager) GetEdge(name string) *Edge {
	edg := Edge{}
	err := m.storage.Get(edgePrefix+name, &edg)
	if err != nil {
		return nil
	}
	return &edg
}

func (m *EdgeManager) GetEdges() []*Edge {
	res, err := m.storage.List(edgePrefix)
	if err != nil {
		log.Error("list %s fail: %v", edgePrefix, err)
		return nil
	}

	edges := make([]*Edge, 0)
	for _, val := range res {
		edge := Edge{}
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

func AddEdge(name string, edge *Edge) {
	if defaultEdgeManager == nil {
		return
	}
	defaultEdgeManager.AddEdge(name, edge)
}

func DelEdge(name string) {
	if defaultEdgeManager == nil {
		return
	}
	defaultEdgeManager.DelEdge(name)
}

func GetEdge(name string) *Edge {
	if defaultEdgeManager == nil {
		return nil
	}
	return defaultEdgeManager.GetEdge(name)
}

func GetEdges() []*Edge {
	if defaultEdgeManager == nil {
		return nil
	}
	return defaultEdgeManager.GetEdges()
}

func VerifyCidr(cidr string) bool {
	if defaultEdgeManager == nil {
		return false
	}

	return defaultEdgeManager.VerifyCidr(cidr)
}
