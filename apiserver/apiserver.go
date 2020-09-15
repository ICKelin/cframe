package main

import (
	"fmt"
	"net/http"

	"github.com/ICKelin/cframe/pkg/edgemanager"
	log "github.com/ICKelin/cframe/pkg/logs"
	"github.com/gin-gonic/gin"
)

type ApiServer struct {
	addr string
}

func NewApiServer(addr string) *ApiServer {
	return &ApiServer{
		addr: addr,
	}
}

func (s *ApiServer) Run() {
	eng := gin.New()
	eng.POST("/api-service/v1/edge/add", s.addEdge)
	eng.DELETE("/api-service/v1/edge/del", s.delEdge)
	eng.GET("/api-service/v1/edge/list", s.getEdgeList)
	eng.GET("/api-service/v1/topology", s.getTopology)

	eng.POST("/api-service/v1/signup", nil)
	eng.POST("/api-service/v1/signin", nil)

	eng.Run(s.addr)
}

func (s *ApiServer) addEdge(ctx *gin.Context) {
	addForm := AddEdgeForm{}
	err := ctx.BindJSON(&addForm)
	if err != nil {
		log.Error("bind add edge form fail: %v", err)
		ctx.JSON(http.StatusBadRequest, err)
		return
	}

	if len(addForm.Name) <= 0 {
		log.Error("invalid name", addForm.Name)
		ctx.JSON(http.StatusBadRequest, fmt.Errorf("invalid name"))
		return
	}

	// verify cidr format and conflict
	ok := edgemanager.VerifyCidr(addForm.Cidr)
	if !ok {
		log.Error("verify cidr fail")
		ctx.JSON(http.StatusBadRequest, fmt.Errorf("invalid cidr"))
		return
	}

	edg := &edgemanager.Edge{
		Type:     addForm.Type,
		Name:     addForm.Name,
		HostAddr: addForm.HostAddr,
		Cidr:     addForm.Cidr,
	}
	edgemanager.AddEdge(edg.Name, edg)
	ctx.JSON(http.StatusOK, nil)
}

func (s *ApiServer) delEdge(ctx *gin.Context) {
	delForm := DeleteEdgeForm{}
	err := ctx.BindJSON(&delForm)
	if err != nil {
		log.Error("bind add edge form fail: %v", err)
		ctx.JSON(http.StatusBadRequest, err)
		return
	}

	if len(delForm.Name) <= 0 {
		log.Error("invalid name", delForm.Name)
		ctx.JSON(http.StatusBadRequest, fmt.Errorf("invalid name"))
		return
	}

	edgemanager.DelEdge(delForm.Name)
	ctx.JSON(http.StatusOK, nil)
}

func (s *ApiServer) getEdgeList(ctx *gin.Context) {
	edges := edgemanager.GetEdges()
	ctx.JSON(http.StatusOK, edges)
}

type topology struct {
	EdgeNode []*edgemanager.Edge     `json:"edge_node"`
	EdgeHost []*edgemanager.EdgeHost `json:"edge_host"`
}

func (s *ApiServer) getTopology(ctx *gin.Context) {
	edges := edgemanager.GetEdges()
	hosts := edgemanager.GetEdgeHosts()
	t := &topology{
		EdgeNode: edges,
		EdgeHost: hosts,
	}
	ctx.JSON(http.StatusOK, t)
}

func (s *ApiServer) signup(ctx *gin.Context) {}

func (s *ApiServer) signin(ctx *gin.Context) {}
