package auth

import (
	"fmt"

	"github.com/ICKelin/cframe/pkg/etcdstorage"
)

var (
	authPrefix  = "/auth"
	authManager *AuthManager
)

type UserInfo struct {
	// user access key
	AccessKey string
	// user access secret
	SecretKey string
	Username  string
	Password  string
}

type AuthManager struct {
	store *etcdstorage.Etcd
}

func NewAuthManager(store *etcdstorage.Etcd) *AuthManager {
	if authManager != nil {
		return authManager
	}
	m := &AuthManager{store}
	authManager = m
	return m
}

func (m *AuthManager) Create(accessKey, secretKey string, user *UserInfo) {
	key := fmt.Sprintf("%s/%s/%s", authPrefix, accessKey, secretKey)
	m.store.Set(key, user)
}

func (m *AuthManager) GetUserInfo(accessKey, secretKey string) (*UserInfo, error) {
	user := UserInfo{}
	key := fmt.Sprintf("%s/%s/%s", authPrefix, accessKey, secretKey)
	err := m.store.Get(key, &user)
	return &user, err
}

func GetUserInfo(accessKey, secretKey string) (*UserInfo, error) {
	if authManager == nil {
		return nil, fmt.Errorf("auth manager without initial")
	}
	return authManager.GetUserInfo(accessKey, secretKey)
}
