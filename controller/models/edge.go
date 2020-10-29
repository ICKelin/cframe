package models

import (
	"time"

	"github.com/ICKelin/cframe/codec/proto"
	"github.com/ICKelin/cframe/pkg/database"
	"gopkg.in/mgo.v2/bson"
)

type EdgeInfo struct {
	database.Model `bson:",inline"`
	Name           string        `json:"name" bson:"name"`
	UserId         bson.ObjectId `json:"userId" bson:"userId"`
	CSPType        proto.CSPType `json:"csptype" bson:"csptype"`
	PublicIP       string        `json:"publicIp" bson:"puslicIp"`
	PublicPort     int32         `json:"publicPort" bson:"publicPort"`
	Cidr           string        `json:"cidr" bson:"cidr"`
	Comment        string        `json:"comment" bson:"comment"`
	ActiveAt       int64         `json:"activeAt" bson:"activeAt"`
}

var (
	C_EDGE = "edge"
)

type EdgeManager struct {
	database.ModelManager
}

func GetEdgeManager() *EdgeManager {
	return &EdgeManager{}
}

func (m *EdgeManager) AddEdge(edge *EdgeInfo) (*EdgeInfo, error) {
	edge.CreatedAt = time.Now().Unix()
	edge.UpdatedAt = time.Now().Unix()
	edge.Id = bson.NewObjectId()
	err := m.Insert(C_EDGE, edge)
	return edge, err
}

func (m *EdgeManager) GetEdgeList(userId bson.ObjectId) ([]*EdgeInfo, error) {
	var result []*EdgeInfo
	query := bson.M{}
	query["userId"] = userId
	query["invalid"] = false
	err := m.FindAll(C_EDGE, query, &result)
	return result, err
}

func (m *EdgeManager) GetEdgeByName(userId bson.ObjectId, name string) (*EdgeInfo, error) {
	var result *EdgeInfo
	query := bson.M{}
	query["userId"] = userId
	query["name"] = name
	query["invalid"] = false
	err := m.FindOne(C_EDGE, query, &result)
	return result, err
}

func (m *EdgeManager) DelEdge(userId bson.ObjectId, name string) error {
	var query = bson.M{}
	var update = bson.M{}
	query["invalid"] = false
	query["userId"] = userId
	query["name"] = name

	update["invalid"] = true
	update["invalid_at"] = time.Now().Unix()
	return m.Update(C_EDGE, query, bson.M{"$set": update})
}

func (m *EdgeManager) UpdateActive(userId bson.ObjectId, name string, tm time.Time) error {
	var query = bson.M{}
	var update = bson.M{}
	query["invalid"] = false
	query["userId"] = userId
	query["name"] = name

	update["activeAt"] = tm.Unix()
	return m.Update(C_EDGE, query, bson.M{"$set": update})
}
