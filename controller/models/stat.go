package models

import (
	"time"

	"github.com/ICKelin/cframe/pkg/database"
	"gopkg.in/mgo.v2/bson"
)

var (
	C_STAT = "stat"
)

type Stat struct {
	database.Model `bson:",inline"`
	UserId         bson.ObjectId `json:"userId" bson:"userId"`
	EdgeName       string        `json:"edgeName" bson:"edgeName"`
	Timestamp      int64         `json:"timestamp" bson:"timestamp"`
	CPU            int64         `json:"cpu" bson:"cpu"`
	Mem            int64         `json:"mem" bson:"mem"`
	TrafficIn      int64         `json:"trafficIn" bson:"trafficIn"`
	TrafficOut     int64         `json:"trafficOut" bson:"trafficOut"`
}

type StatManager struct {
	database.ModelManager
}

func GetStatManager() *StatManager {
	return &StatManager{}
}

func (m *StatManager) AddStat(stat *Stat) error {
	stat.CreatedAt = time.Now().Unix()
	stat.UpdatedAt = time.Now().Unix()
	return m.Insert(C_STAT, stat)
}

func (m *StatManager) GetUserStat(userId bson.ObjectId, edgeName string, from int64, count, dir int) ([]*Stat, error) {
	var result []*Stat
	query := bson.M{}
	query["invalid"] = false
	query["edgeName"] = edgeName
	query["userId"] = userId
	if dir == 1 {
		query["created_at"] = bson.M{"$gt": from}
	} else {
		query["created_at"] = bson.M{"$lte": from}
	}
	err := m.C(C_STAT).Find(query).Sort("-timestamp").Limit(count).All(&result)
	return result, err
}

func (m *StatManager) RemoveOldStat() error {
	var query = bson.M{}
	query["created_at"] = bson.M{"$let": time.Now().AddDate(0, 0, -5).Unix()}
	return m.Remove(C_STAT, query)
}
