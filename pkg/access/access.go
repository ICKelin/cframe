package access

import (
	"encoding/json"
	"fmt"

	"github.com/ICKelin/cframe/pkg/etcdstorage"
)

var (
	accessPrefix   = "/cloud"
	defaultManager *AccessManager
)

type CloudPlatform string

type AccessManager struct {
	storage *etcdstorage.Etcd
}

type AccessInfo struct {
	CloudPlatform CloudPlatform
	AccessKey     string
	AccessSecret  string
}

func NewAccessManager(storage *etcdstorage.Etcd) *AccessManager {
	if defaultManager != nil {
		return defaultManager
	}
	m := &AccessManager{
		storage: storage,
	}
	defaultManager = m
	return m
}

func (m *AccessManager) Add(access *AccessInfo) {
	key := fmt.Sprintf("%s/%s", accessPrefix, access.CloudPlatform)
	m.storage.Set(key, access)
}

func (m *AccessManager) Del(platform CloudPlatform) {
	key := fmt.Sprintf("%s/%s", accessPrefix, platform)
	m.storage.Del(key)
}

func (m *AccessManager) GetAccessList() ([]*AccessInfo, error) {
	kvs, err := m.storage.List(accessPrefix)
	if err != nil {
		return nil, err
	}
	accessList := make([]*AccessInfo, 0)
	for _, v := range kvs {
		a := AccessInfo{}
		err := json.Unmarshal([]byte(v), &a)
		if err != nil {
			return nil, err
		}
		accessList = append(accessList, &a)
	}
	return accessList, nil
}

func Add(access *AccessInfo) {
	if defaultManager != nil {
		defaultManager.Add(access)
	}
}

func Del(platform CloudPlatform) {
	if defaultManager != nil {
		defaultManager.Del(platform)
	}
}

func GetAccessList() ([]*AccessInfo, error) {
	if defaultManager != nil {
		return defaultManager.GetAccessList()
	}
	return nil, fmt.Errorf("access manager without initial")
}
