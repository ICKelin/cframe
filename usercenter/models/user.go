package models

import (
	"fmt"
	"time"

	"github.com/ICKelin/cframe/pkg/database"
	"gopkg.in/mgo.v2/bson"
)

var (
	C_USER = "user"
)

type User struct {
	database.Model `bson:",inline"`
	Username       string `json:"username" bson:"username"`
	Password       string `json:"-" bson:"password"`
	Secret         string `json:"secret" bson:"secret"`
	Email          string `json:"email" bson:"email"`
	About          string `json:"about" bson:"about"`
}

type UserManager struct {
	database.ModelManager
}

func GetUserManager() *UserManager {
	return &UserManager{}
}

func (m *UserManager) CreateUser(user *User) (*User, error) {
	user.CreatedAt = time.Now().Unix()
	user.UpdatedAt = time.Now().Unix()
	user.Id = bson.NewObjectId()
	err := m.Insert(C_USER, user)
	if err != nil {
		return nil, err
	}

	return user, err
}

func (m *UserManager) GetUserBySecret(secret string) (*User, error) {
	var result *User
	var query = bson.M{}
	query["secret"] = secret
	query["invalid"] = false

	err := m.FindOne(C_USER, query, &result)
	if result == nil {
		return nil, fmt.Errorf("not found")
	}
	return result, err
}

func (m *UserManager) GetUserByName(name string) (*User, error) {
	var result *User
	var query = bson.M{}
	query["username"] = name
	query["invalid"] = false

	err := m.FindOne(C_USER, query, &result)
	if result == nil {
		return nil, fmt.Errorf("not found")
	}
	return result, err
}
func (m *UserManager) GetUserByEmail(email string) (*User, error) {
	var result *User
	var query = bson.M{}
	query["email"] = email
	query["invalid"] = false

	err := m.FindOne(C_USER, query, &result)
	if result == nil {
		return nil, fmt.Errorf("not found")
	}
	return result, err
}
func (m *UserManager) GetUserById(userId bson.ObjectId) (*User, error) {
	var result *User
	var query = bson.M{}
	query["_id"] = userId
	query["invalid"] = false

	err := m.FindOne(C_USER, query, &result)
	if result == nil {
		return nil, fmt.Errorf("not found")
	}
	return result, err
}

func (m *UserManager) VerifyUser(username, password string) (*User, error) {
	var result *User
	var query = bson.M{}
	query["username"] = username
	query["password"] = password
	query["invalid"] = false

	err := m.FindOne(C_USER, query, &result)
	if result == nil {
		return nil, fmt.Errorf("not found")
	}
	return result, err
}
