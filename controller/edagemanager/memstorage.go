package edagemanager

import (
	"sync"
)

type MemStorage struct {
	// edages info
	// key: Edage.Name
	// val: &Edage{}
	edageMu sync.Mutex
	edages  map[string]*Edage
	mu      sync.Mutex
	table   map[string]*Edage
}

func NewMemStorage() *MemStorage {
	return &MemStorage{
		table: make(map[string]*Edage),
	}
}

func (m *MemStorage) Set(key string, edage *Edage) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.table[key] = edage
}

func (m *MemStorage) Get(key string) *Edage {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.table[key]
}

func (m *MemStorage) Del(key string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.table, key)
}

func (m *MemStorage) List() map[string]*Edage {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.table
}

func (m *MemStorage) Range(funcCall rangeFunc) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for key, val := range m.table {
		ok := funcCall(key, val)
		if !ok {
			return
		}
	}
}
