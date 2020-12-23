package models

import (
	"time"

	"github.com/ICKelin/cframe/pkg/database"
	"gopkg.in/mgo.v2/bson"
)

type Route struct {
	database.Model `bson:",inline"`
	EdgeId         bson.ObjectId `json:"edgeId" bson:"edgeId"`
	Cidr           string        `json:"cidr" bson:"cidr"`
	Nexthop        string        `json:"nexthop" bson:"nexthop"`
}

var (
	C_ROUTE = "route"
)

type RouteManager struct {
	database.ModelManager
}

func GetRouteManager() *RouteManager {
	return &RouteManager{}
}

func (m *RouteManager) GetEdgeRoutes(edgeId bson.ObjectId) ([]*Route, error) {
	var result []*Route
	query := bson.M{}
	query["edgeId"] = edgeId
	query["invalid"] = false
	err := m.FindAll(C_ROUTE, query, &result)
	return result, err
}

func (m *RouteManager) AddRoute(edgeId bson.ObjectId, cidr, nexthop string) (*Route, error) {
	rotue := &Route{
		EdgeId:  edgeId,
		Cidr:    cidr,
		Nexthop: nexthop,
	}
	rotue.CreatedAt = time.Now().Unix()
	rotue.UpdatedAt = time.Now().Unix()
	rotue.Id = bson.NewObjectId()
	err := m.Insert(C_ROUTE, rotue)
	return rotue, err
}

func (m *RouteManager) DelRoute(id bson.ObjectId) error {
	var query = bson.M{}
	var update = bson.M{}
	query["invalid"] = false
	query["_id"] = id

	update["invalid"] = true
	update["invalid_at"] = time.Now().Unix()
	return m.Update(C_ROUTE, query, bson.M{"$set": update})
}
