package edagemanager

import (
	"strconv"
	"strings"

	"github.com/ICKelin/cframe/pkg/ip"
	log "github.com/ICKelin/cframe/pkg/logs"
)

var (
	defaultEdageManager *EdageManager
)

type EdageManager struct {
	storage IStorage
}

func New() *EdageManager {
	if defaultEdageManager != nil {
		return defaultEdageManager
	}

	m := &EdageManager{
		storage: NewMemStorage(),
	}
	defaultEdageManager = m
	return m
}

func (m *EdageManager) AddEdage(name string, edage *Edage) {
	m.storage.Set(name, edage)
}

func (m *EdageManager) DelEdage(name string) {
	m.storage.Del(name)
}

func (m *EdageManager) GetEdage(name string) *Edage {
	return m.storage.Get(name)
}

func (m *EdageManager) GetEdages() map[string]*Edage {
	return m.storage.List()
}

func (m *EdageManager) VerifyCidr(cidr string) bool {
	b := true
	m.storage.Range(func(key string, edage *Edage) bool {
		b = m.verifyConflict(cidr, edage.Cidr)
		return b
	})
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

func GetEdages() map[string]*Edage {
	if defaultEdageManager == nil {
		return nil
	}
	return defaultEdageManager.GetEdages()
}
