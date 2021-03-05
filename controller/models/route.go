package models

import (
	"time"

	"github.com/ICKelin/cframe/pkg/database"
	log "github.com/ICKelin/cframe/pkg/logs"
	"gopkg.in/mgo.v2/bson"
)

type Route struct {
	database.Model `bson:",inline"`
	UserId         bson.ObjectId `json:"userId" bson:"userId"`
	Name           string        `json:"name" bson:"name"`
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

func (m *RouteManager) AddRoute(userId bson.ObjectId, name, cidr, nexthop string) (*Route, error) {
	route := &Route{
		UserId:  userId,
		Cidr:    cidr,
		Name:    name,
		Nexthop: nexthop,
	}
	route.CreatedAt = time.Now().Unix()
	route.UpdatedAt = time.Now().Unix()
	route.Id = bson.NewObjectId()
	err := m.Insert(C_ROUTE, route)
	log.Debug("route %v, err %v", route, err)
	return route, err
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

func (m *RouteManager) GetUserRoutes(userId bson.ObjectId) ([]*Route, error) {
	var result []*Route
	query := bson.M{}
	query["invalid"] = false
	query["userId"] = userId
	err := m.FindAll(C_ROUTE, query, &result)
	return result, err
}

func (m *RouteManager) GetOtherRoutes(userId bson.ObjectId, curAddr string) ([]*Route, error) {
	var result []*Route
	query := bson.M{}
	query["invalid"] = false
	query["userId"] = userId
	query["nexthop"] = bson.M{"$ne": curAddr}
	err := m.FindAll(C_ROUTE, query, &result)
	return result, err
}
