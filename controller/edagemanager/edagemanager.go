package edagemanager

import (
	"log"
	"strconv"
	"strings"
	"sync"

	"github.com/ICKelin/cframe/pkg/ip"
)

var (
	defaultEdageManager *EdageManager
)

type EdageManager struct {
	// edages info
	// key: Edage.Name
	// val: &Edage{}
	edageMu sync.Mutex
	edages  map[string]*Edage
}

func New() *EdageManager {
	if defaultEdageManager != nil {
		return defaultEdageManager
	}

	m := &EdageManager{
		edages: make(map[string]*Edage),
	}
	defaultEdageManager = m
	return m
}

func (m *EdageManager) AddEdage(name string, edage *Edage) {
	m.edageMu.Lock()
	defer m.edageMu.Unlock()
	m.edages[name] = edage
}

func (m *EdageManager) DelEdage(name string) {
	m.edageMu.Lock()
	defer m.edageMu.Unlock()
	delete(m.edages, name)
}

func (m *EdageManager) GetEdage(name string) *Edage {
	m.edageMu.Lock()
	defer m.edageMu.Unlock()
	return m.edages[name]
}

func (m *EdageManager) GetEdages() map[string]*Edage {
	m.edageMu.Lock()
	defer m.edageMu.Unlock()
	return m.edages
}

func (m *EdageManager) VerifyCidr(cidr string) bool {
	m.edageMu.Lock()
	defer m.edageMu.Unlock()
	for _, e := range m.edages {
		if m.verifyConflict(cidr, e.Cidr) == false {
			return false
		}
	}

	return true
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
		log.Println("[E] invalid cidr format: ", cidr1)
		return false
	}

	ip1, sprefix1 := sp[0], sp[1]
	prefix1, err := strconv.Atoi(sprefix1)
	if err != nil {
		log.Println("[E] invalid cidr format: ", cidr1)
		return false
	}

	ipv41, err := ip.ParseIP4(ip1)
	if err != nil {
		log.Println("[E] invalid cidr format: ", cidr1)
		return false
	}

	sp = strings.Split(cidr2, "/")
	if len(sp) != 2 {
		log.Println("[E] invalid cidr format: ", cidr2)
		return false
	}

	ip2, sprefix2 := sp[0], sp[1]
	prefix2, err := strconv.Atoi(sprefix2)
	if err != nil {
		log.Println("[E] invalid cidr format: ", cidr2)
		return false
	}

	ipv42, err := ip.ParseIP4(ip2)
	if err != nil {
		log.Println("[E] invalid cidr format: ", cidr2)
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
