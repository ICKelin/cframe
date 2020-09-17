package auth

import (
	"encoding/base64"
	"fmt"

	"github.com/ICKelin/cframe/pkg/etcdstorage"
	uuid "github.com/satori/go.uuid"
)

var (
	userPrefix  = "/user"
	userManager *UserManager
)

type UserInfo struct {
	// user access secret
	SecretKey string
	Username  string
	Password  string
}

type UserManager struct {
	store *etcdstorage.Etcd
}

func NewUserManager(store *etcdstorage.Etcd) *UserManager {
	if userManager != nil {
		return userManager
	}
	m := &UserManager{store}
	userManager = m
	return m
}

func (m *UserManager) CreateUser(username, password string) (*UserInfo, error) {
	key := fmt.Sprintf("%s/%s", userPrefix, username)
	uniq := uuid.NewV4()
	secret := base64.StdEncoding.EncodeToString(uniq.Bytes())
	userInfo := &UserInfo{
		Username:  username,
		Password:  password,
		SecretKey: secret,
	}

	err := m.store.Set(key, userInfo)
	if err != nil {
		return nil, err
	}
	return userInfo, err
}

func (m *UserManager) GetUser(username, password string) (*UserInfo, error) {
	key := fmt.Sprintf("%s/%s", userPrefix, username)
	userInfo := UserInfo{}
	m.store.Get(key, &userInfo)
	if userInfo.Password != password {
		return nil, fmt.Errorf("invalid password")
	}
	return &userInfo, nil
}

func CreateUser(username, password string) (*UserInfo, error) {
	if userManager == nil {
		return nil, fmt.Errorf("usermanager without initial")
	}

	return userManager.CreateUser(username, password)
}

func GetUser(username, password string) (*UserInfo, error) {
	if userManager == nil {
		return nil, fmt.Errorf("usermanager without initial")
	}

	return userManager.GetUser(username, password)
}
