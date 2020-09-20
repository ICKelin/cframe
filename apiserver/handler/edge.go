package handler

import (
	"fmt"

	"github.com/ICKelin/cframe/pkg/edgemanager"
	log "github.com/ICKelin/cframe/pkg/logs"
	"github.com/gin-gonic/gin"
)

type EdgeHandler struct {
	BaseHandler
}

func (h *EdgeHandler) Run(eng *gin.Engine) {
	group := eng.Group("/api-service/v1/edge")
	group.Use(h.MidAuth)
	{
		group.POST("/add", h.addEdge)
		group.POST("/del", h.delEdge)
		group.GET("/list", h.getEdgeList)
		group.GET("/topology", h.getTopology)
	}

}

func (h *EdgeHandler) addEdge(ctx *gin.Context) {
	username := ctx.GetString("username")
	addForm := AddEdgeForm{}
	if ok := h.BindAndValidate(ctx, &addForm); !ok {
		return
	}

	// verify cidr format and conflict
	ok := edgemanager.VerifyCidr(addForm.Cidr)
	if !ok {
		log.Error("verify cidr fail")
		h.Response(ctx, nil, fmt.Errorf("cidr conflict"))
		return
	}

	edg := &edgemanager.Edge{
		Type:     addForm.Type,
		Name:     addForm.Name,
		HostAddr: addForm.HostAddr,
		Cidr:     addForm.Cidr,
	}
	edgemanager.AddEdge(username, edg.Name, edg)
	h.Response(ctx, nil, nil)
}

func (h *EdgeHandler) delEdge(ctx *gin.Context) {
	username := ctx.GetString("username")

	delForm := DeleteEdgeForm{}
	if ok := h.BindAndValidate(ctx, &delForm); !ok {
		return
	}

	edgemanager.DelEdge(username, delForm.Name)
	h.Response(ctx, nil, nil)
}

func (h *EdgeHandler) getEdgeList(ctx *gin.Context) {
	username := ctx.GetString("username")

	edges := edgemanager.GetEdges(username)
	h.Response(ctx, edges, nil)
}

type topology struct {
	EdgeNode []*edgemanager.Edge     `json:"edge_node"`
	EdgeHost []*edgemanager.EdgeHost `json:"edge_host"`
}

func (h *EdgeHandler) getTopology(ctx *gin.Context) {
	username := h.GetUsername(ctx)
	log.Debug("get topology for user %s", username)

	edges := edgemanager.GetEdges(username)
	hosts := edgemanager.GetEdgeHosts()
	t := &topology{
		EdgeNode: edges,
		EdgeHost: hosts,
	}
	h.Response(ctx, t, nil)
}
