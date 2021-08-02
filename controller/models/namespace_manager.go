package models

import (
	"encoding/json"
	"fmt"

	"github.com/ICKelin/cframe/pkg/etcdstorage"
	log "github.com/ICKelin/cframe/pkg/logs"
)

var (
	namespacePrefix = "/namespace/"
)

type Namespace struct {
	Name   string
	Secret string
}

type NamespaceManager struct {
	storage *etcdstorage.Etcd
}

func NewNamespaceManager(store *etcdstorage.Etcd) *NamespaceManager {
	return &NamespaceManager{
		storage: store,
	}
}

func (m *NamespaceManager) AddNamespace(ns *Namespace) error {
	key := fmt.Sprintf("%s%s", namespacePrefix, ns.Name)
	return m.storage.Set(key, ns)
}

func (m *NamespaceManager) DelNamespace(name string) error {
	key := fmt.Sprintf("%s%s", namespacePrefix, name)
	m.storage.Del(key)
	return nil
}

func (m *NamespaceManager) GetNamespace(name string) (*Namespace, error) {
	key := fmt.Sprintf("%s%s", namespacePrefix, name)
	ns := Namespace{}
	err := m.storage.Get(key, &ns)
	if err != nil {
		return nil, err
	}
	return &ns, nil
}

func (m *NamespaceManager) GetNamespaces() []*Namespace {
	key := fmt.Sprintf("%s", namespacePrefix)
	res, err := m.storage.List(key)
	if err != nil {
		log.Error("list %s fail: %v", namespacePrefix, err)
		return nil
	}

	nss := make([]*Namespace, 0)
	for _, val := range res {
		r := Namespace{}
		err := json.Unmarshal([]byte(val), &r)
		if err != nil {
			log.Error("unmarshal to edge fail: %v", err)
			continue
		}
		nss = append(nss, &r)
	}
	return nss
}
