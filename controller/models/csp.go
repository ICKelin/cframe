package models

import (
	"fmt"
	"time"

	"github.com/ICKelin/cframe/codec/proto"
	"github.com/ICKelin/cframe/pkg/database"
	"gopkg.in/mgo.v2/bson"
)

type CSP struct {
	database.Model `bson:",inline"`
	UserId         bson.ObjectId `json:"userId" bson:"userId"`
	CSPType        proto.CSPType `json:"identify" bson:"identify"`
	AccessKey      string        `json:"accessKey" bson:"accessKey"`
	SecretKey      string        `json:"secretKey" bson:"secretKey"`
}

var (
	C_CSP = "csp"
)

type CSPManager struct {
	database.ModelManager
}

func GetCSPManager() *CSPManager {
	return &CSPManager{}
}

func (m *CSPManager) AddCSP(csp *CSP) error {
	csp.CreatedAt = time.Now().Unix()
	csp.UpdatedAt = time.Now().Unix()
	return m.Insert(C_CSP, csp)
}

func (m *CSPManager) GetCSP(userId bson.ObjectId, id proto.CSPType) (*CSP, error) {
	var result *CSP
	query := bson.M{}
	query["userId"] = userId
	query["identify"] = id
	query["invalid"] = false
	err := m.FindOne(C_CSP, query, &result)
	if result == nil {
		return nil, fmt.Errorf("not found")
	}
	return result, err
}

func (m *CSPManager) GetCSPList(userId bson.ObjectId) ([]*CSP, error) {
	var result []*CSP
	query := bson.M{}
	query["userId"] = userId
	query["invalid"] = false
	err := m.FindAll(C_CSP, query, &result)
	return result, err
}

func (m *CSPManager) DelCSP(userId bson.ObjectId, csptype proto.CSPType) error {
	var query = bson.M{}
	var update = bson.M{}
	query["invalid"] = false
	query["userId"] = userId
	query["identify"] = csptype

	update["invalid"] = true
	update["invalid_at"] = time.Now().Unix()
	return m.Update(C_CSP, query, bson.M{"$set": update})
}
