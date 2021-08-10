package models

import (
	"encoding/json"
	"fmt"

	"github.com/ICKelin/cframe/codec"
	"github.com/ICKelin/cframe/pkg/etcdstorage"
	log "github.com/ICKelin/cframe/pkg/logs"
)

var (
	cspPrefix = "/csps/"
)

type CSPManagr struct {
	storage *etcdstorage.Etcd
}

func NewCSPManager(store *etcdstorage.Etcd) *CSPManagr {
	return &CSPManagr{
		storage: store,
	}
}

func (m *CSPManagr) AddCSP(namespace, name string, csp *codec.CSPInfo) error {
	key := fmt.Sprintf("%s%s/%s", cspPrefix, namespace, name)
	return m.storage.Set(key, csp)
}

func (m *CSPManagr) GetCSP(namespace, name string) (*codec.CSPInfo, error) {
	key := fmt.Sprintf("%s%s/%s", cspPrefix, namespace, name)
	var csp codec.CSPInfo
	err := m.storage.Get(key, &csp)
	if err != nil {
		return nil, err
	}
	return &csp, nil
}

func (m *CSPManagr) DelCSP(namespace, name string) error {
	key := fmt.Sprintf("%s%s/%s", cspPrefix, namespace, name)
	m.storage.Del(key)
	return nil
}

func (m *CSPManagr) GetCSPList(namespace string) []*codec.CSPInfo {
	key := fmt.Sprintf("%s%s", cspPrefix, namespace)
	res, err := m.storage.List(key)
	if err != nil {
		log.Error("list %s fail: %v", key, err)
		return nil
	}

	csps := make([]*codec.CSPInfo, 0)
	for _, val := range res {
		r := codec.CSPInfo{}
		err := json.Unmarshal([]byte(val), &r)
		if err != nil {
			log.Error("unmarshal to edge fail: %v", err)
			continue
		}
		csps = append(csps, &r)
	}
	return csps
}
