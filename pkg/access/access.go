package access

import (
	"fmt"

	"github.com/ICKelin/cframe/pkg/etcdstorage"
)

var (
	accessPrefix = "/cloud"
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

func NewAccessManager() *AccessManager {
	return &AccessManager{}
}

func (m *AccessManager) Add(access *AccessInfo) {
	key := fmt.Sprintf("%s/%s", accessPrefix, access.CloudPlatform)
	m.storage.Set(key, access)
}

func (m *AccessManager) Del(platform CloudPlatform) {
	key := fmt.Sprintf("%s/%s", accessPrefix, platform)
	m.storage.Del(key)
}

func (m *AccessManager) GetAccessList() {
	// kvs := m.storage.List()
}
