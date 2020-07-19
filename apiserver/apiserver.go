package main

import (
	"fmt"
	"net/http"

	"github.com/ICKelin/cframe/pkg/edagemanager"
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
	eng.POST("/api-service/v1/edage/add", s.addEdage)
	eng.POST("/api-service/v1/edage/del", s.delEdage)
	eng.GET("/api-service/v1/edage/list", s.getEdageList)
	eng.GET("/api-service/v1/topology", s.getTopology)
	eng.Run(s.addr)
}

func (s *ApiServer) addEdage(ctx *gin.Context) {
	addForm := AddEdageForm{}
	err := ctx.BindJSON(&addForm)
	if err != nil {
		log.Error("bind add edage form fail: %v", err)
		ctx.JSON(http.StatusBadRequest, err)
		return
	}

	if len(addForm.Name) <= 0 {
		log.Error("invalid name", addForm.Name)
		ctx.JSON(http.StatusBadRequest, fmt.Errorf("invalid name"))
		return
	}

	// verify cidr format and conflict
	ok := edagemanager.VerifyCidr(addForm.Cidr)
	if !ok {
		log.Error("verify cidr fail")
		ctx.JSON(http.StatusBadRequest, fmt.Errorf("invalid cidr"))
		return
	}

	edg := &edagemanager.Edage{
		Name:     addForm.Name,
		HostAddr: addForm.HostAddr,
		Cidr:     addForm.Cidr,
	}
	edagemanager.AddEdage(edg.Name, edg)
	ctx.JSON(http.StatusOK, nil)
}

func (s *ApiServer) delEdage(ctx *gin.Context) {
	delForm := DeleteEdageForm{}
	err := ctx.BindJSON(&delForm)
	if err != nil {
		log.Error("bind add edage form fail: %v", err)
		ctx.JSON(http.StatusBadRequest, err)
		return
	}

	if len(delForm.Name) <= 0 {
		log.Error("invalid name", delForm.Name)
		ctx.JSON(http.StatusBadRequest, fmt.Errorf("invalid name"))
		return
	}

	edagemanager.DelEdage(delForm.Name)
	ctx.JSON(http.StatusOK, nil)
}

func (s *ApiServer) getEdageList(ctx *gin.Context) {
	edages := edagemanager.GetEdages()
	ctx.JSON(http.StatusOK, edages)
}

type topology struct {
	EdageNode []*edagemanager.Edage     `json:"edage_node"`
	EdageHost []*edagemanager.EdageHost `json:"edage_host"`
}

func (s *ApiServer) getTopology(ctx *gin.Context) {
	edages := edagemanager.GetEdages()
	hosts := edagemanager.GetEdageHosts()
	t := &topology{
		EdageNode: edages,
		EdageHost: hosts,
	}
	ctx.JSON(http.StatusOK, t)
}
