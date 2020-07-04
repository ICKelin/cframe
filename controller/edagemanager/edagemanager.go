package edagemanager

import (
	"fmt"
	"sync"
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

type Edage struct {
	Name     string
	Comment  string
	Cidr     string
	HostAddr string
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

func (e *Edage) String() string {
	return fmt.Sprintf("name: %s hostaddr: %s cidr: %s", e.Name, e.HostAddr, e.Cidr)
}
