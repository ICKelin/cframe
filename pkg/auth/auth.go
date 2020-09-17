package auth

import (
	"fmt"
	"time"

	"github.com/ICKelin/cframe/pkg/etcdstorage"
)

var (
	authPrefix    = "/auth"
	tokenPrefix   = "/token"
	defaultExpire = time.Hour * 3
	authManager   *AuthManager
)

type Token struct {
	Username string
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

// create user secret key
func (m *AuthManager) CreateSecret(secretKey string, user *UserInfo) error {
	key := fmt.Sprintf("%s/%s", authPrefix, secretKey)
	return m.store.Set(key, user)
}

// set user signin token
func (m *AuthManager) SetUserToken(token string, userInfo *UserInfo) error {
	key := fmt.Sprintf("%s/%s", tokenPrefix, token)
	return m.store.SetWithExpiration(key, userInfo, defaultExpire)
}

// get user info by secret key
func (m *AuthManager) GetAuth(secretKey string) (*UserInfo, error) {
	user := UserInfo{}
	key := fmt.Sprintf("%s/%s", authPrefix, secretKey)
	err := m.store.Get(key, &user)
	return &user, err
}

// get user info by signin token
func (m *AuthManager) GetUserByToken(token string) (*UserInfo, error) {
	key := fmt.Sprintf("%s/%s", tokenPrefix, token)
	userInfo := UserInfo{}
	err := m.store.Get(key, &userInfo)
	return &userInfo, err
}

func GetAuth(secretKey string) (*UserInfo, error) {
	if authManager == nil {
		return nil, fmt.Errorf("auth manager without initial")
	}
	return authManager.GetAuth(secretKey)
}

func GetUserByToken(token string) (*UserInfo, error) {
	if authManager == nil {
		return nil, fmt.Errorf("auth manager without initial")
	}
	return authManager.GetUserByToken(token)
}

func CreateSecret(secretKey string, userInfo *UserInfo) error {
	if authManager == nil {
		return fmt.Errorf("auth manager without initial")
	}
	return authManager.CreateSecret(secretKey, userInfo)
}

func SetUserToken(token string, userInfo *UserInfo) error {
	if authManager == nil {
		return fmt.Errorf("auth manager without initial")
	}
	return authManager.SetUserToken(token, userInfo)
}
