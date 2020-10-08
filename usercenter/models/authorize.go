package models

import (
	"fmt"
	"time"

	"github.com/ICKelin/cframe/pkg/database"
	"gopkg.in/mgo.v2/bson"
)

var (
	C_AUTH = "authorize"
)

type Auth struct {
	database.Model `bson:",inline"`
	UserId         bson.ObjectId `json:"userId" bson:"userId"`
	Token          string        `json:"token" bson:"token"`
	ExpiredIn      int64         `json:"expiredIn" bson:"expiredIn"`
}

type AuthManager struct {
	database.ModelManager
}

func GetAuthManager() *AuthManager {
	return &AuthManager{}
}

func (m *AuthManager) AddAuth(authInfo *Auth) (*Auth, error) {
	authInfo.CreatedAt = time.Now().Unix()
	authInfo.UpdatedAt = time.Now().Unix()
	err := m.Insert(C_AUTH, authInfo)
	return authInfo, err
}

func (m *AuthManager) GetAuthByToken(token string) (*Auth, error) {
	var result *Auth
	var query = bson.M{}
	query["invalid"] = false
	query["token"] = token
	// query["expiredIn"] = bson.M{"$gte": time.Now().Unix()}

	err := m.FindOne(C_AUTH, query, &result)
	if result == nil {
		return nil, fmt.Errorf("not found")
	}
	return result, err
}
