package edagemanager

import (
	"context"
	"encoding/json"
	"strconv"
	"strings"

	"github.com/ICKelin/cframe/pkg/ip"
	log "github.com/ICKelin/cframe/pkg/logs"
	"github.com/coreos/etcd/clientv3"
)

var (
	defaultEdageManager *EdageManager
	edagePrefix         = "/edages/"
)

type EdageManager struct {
	storage *EtcdStorage
}

func New(store *EtcdStorage) *EdageManager {
	if defaultEdageManager != nil {
		return defaultEdageManager
	}

	m := &EdageManager{
		storage: store,
	}
	defaultEdageManager = m
	return m
}

func (m *EdageManager) Watch(delfunc, putfunc func(edage *Edage)) {
	chs := m.storage.cli.Watch(context.Background(), edagePrefix,
		clientv3.WithPrefix(), clientv3.WithPrevKV())

	for c := range chs {
		for _, evt := range c.Events {
			log.Info("type: %v", evt.Type)
			log.Info("new: %v", evt.Kv)
			log.Info("old: %v", evt.PrevKv)
			switch evt.Type {
			case clientv3.EventTypeDelete:
				if delfunc != nil {
					edage := Edage{}
					err := json.Unmarshal(evt.PrevKv.Value, &edage)
					if err != nil {
						log.Info("json unmarshal fail: %v", err)
						continue
					}

					delfunc(&edage)
				}

			case clientv3.EventTypePut:
				if putfunc != nil {
					edage := Edage{}
					err := json.Unmarshal(evt.Kv.Value, &edage)
					if err != nil {
						log.Info("json unmarshal fail: %v", err)
						continue
					}

					putfunc(&edage)
				}
			}
		}
	}

}

func (m *EdageManager) AddEdage(name string, edage *Edage) {
	m.storage.Set(edagePrefix+name, edage)
}

func (m *EdageManager) DelEdage(name string) {
	m.storage.Del(edagePrefix + name)
}

func (m *EdageManager) GetEdage(name string) *Edage {
	edg := Edage{}
	err := m.storage.Get(edagePrefix+name, &edg)
	if err != nil {
		return nil
	}
	return &edg
}

func (m *EdageManager) GetEdages() []*Edage {
	res, err := m.storage.List(edagePrefix)
	if err != nil {
		log.Error("list %s fail: %v", edagePrefix, err)
		return nil
	}

	edages := make([]*Edage, 0)
	for _, val := range res {
		edage := Edage{}
		err := json.Unmarshal([]byte(val), &edage)
		if err != nil {
			log.Error("unmarshal to edage fail: %v", err)
			continue
		}
		edages = append(edages, &edage)
	}
	return edages
}

func (m *EdageManager) VerifyCidr(cidr string) bool {
	b := true
	return b
}

// VerifyConflict verify cidr1 and cidr2 ip range
// [bip1, eip1], [bip2, eip2]
// bip1 < bip2 < eip1
// bip1 < eip2 < eip1 or
// bip2 < bip1 < eip2
// bip2 < eip1 < eip2
func (m *EdageManager) verifyConflict(cidr1, cidr2 string) bool {
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

func AddEdage(name string, edage *Edage) {
	if defaultEdageManager == nil {
		return
	}
	defaultEdageManager.AddEdage(name, edage)
}

func DelEdage(name string) {
	if defaultEdageManager == nil {
		return
	}
	defaultEdageManager.DelEdage(name)
}

func GetEdage(name string) *Edage {
	if defaultEdageManager == nil {
		return nil
	}
	return defaultEdageManager.GetEdage(name)
}

func GetEdages() []*Edage {
	if defaultEdageManager == nil {
		return nil
	}
	return defaultEdageManager.GetEdages()
}

func VerifyCidr(cidr string) bool {
	if defaultEdageManager == nil {
		return false
	}

	return defaultEdageManager.VerifyCidr(cidr)
}
