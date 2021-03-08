package handler

import (
	"context"
	"fmt"
	"os"

	"github.com/ICKelin/cframe/codec/proto"
	log "github.com/ICKelin/cframe/pkg/logs"
	"github.com/gin-gonic/gin"
)

type RouteHandler struct {
	BaseHandler
	ctrlCli proto.ControllerServiceClient
}

func NewRouteHandler(userCli proto.UserServiceClient, ctrl proto.ControllerServiceClient) *RouteHandler {
	return &RouteHandler{
		BaseHandler: BaseHandler{userCli: userCli},
		ctrlCli:     ctrl,
	}
}

func (h *RouteHandler) Run(eng *gin.Engine) {
	group := eng.Group("/api-service/v1/route")
	group.Use(h.MidAuth)
	{
		group.GET("/list", h.getRouteList)
		group.POST("/add", h.addRoute)
		group.POST("/del", h.delRoute)
		group.GET("/topology", h.getTopolocy)
	}
}

func (h *RouteHandler) getRouteList(ctx *gin.Context) {
	userId := h.GetUserId(ctx)
	log.Debug("get routes for user %s", userId)
	reply, err := h.ctrlCli.GetUserRoutes(context.Background(), &proto.GetUserRoutesReq{
		UserId: userId,
	})

	if err != nil {
		log.Error("get routes rpc fail: %v", err)
		h.Response(ctx, nil, err)
		return
	}

	if reply.Code != 0 {
		log.Error("get routes fail: %d %s", reply.Code, reply.Message)
		h.Response(ctx, nil, fmt.Errorf("<%d>%s", reply.Code, reply.Message))
		return
	}
	h.Response(ctx, reply, nil)
}

func (h *RouteHandler) addRoute(ctx *gin.Context) {
	userId := h.GetUserId(ctx)
	log.Debug("add route for user %s", userId)
	f := AddRouteForm{}
	if ok := h.BindAndValidate(ctx, &f); !ok {
		return
	}

	reply, err := h.ctrlCli.AddRoute(context.Background(), &proto.AddRouteReq{
		UserId:  userId,
		Name:    f.Name,
		Cidr:    f.Cidr,
		Nexthop: f.Nexthop,
	})

	if err != nil {
		log.Error("add route rpc fail: %v", err)
		h.Response(ctx, nil, err)
		return
	}

	if reply.Code != 0 {
		log.Error("add route fail: %d %s", reply.Code, reply.Message)
		h.Response(ctx, nil, fmt.Errorf("<%d>%s", reply.Code, reply.Message))
		return
	}

	log.Debug("add route reply: %v", reply)
	h.Response(ctx, reply, nil)
}

func (h *RouteHandler) delRoute(ctx *gin.Context) {
	userId := h.GetUserId(ctx)
	log.Debug("del route for user %s", userId)
	f := DelRouteForm{}
	if ok := h.BindAndValidate(ctx, &f); !ok {
		return
	}

	reply, err := h.ctrlCli.DelRoute(context.Background(), &proto.DelRouteReq{
		Id:     f.Id,
		UserId: userId,
	})

	if err != nil {
		log.Error("del route rpc fail: %v", err)
		h.Response(ctx, nil, err)
		return
	}

	if reply.Code != 0 {
		log.Error("del route fail: %d %s", reply.Code, reply.Message)
		h.Response(ctx, nil, fmt.Errorf("<%d>%s", reply.Code, reply.Message))
		return
	}

	log.Debug("del route reply: %v", reply)
	h.Response(ctx, reply, nil)
}

func (h *RouteHandler) getTopolocy(ctx *gin.Context) {
	userId := h.GetUserId(ctx)
	log.Debug("get topolocy for user %s", userId)
	edgeReply, err := h.ctrlCli.GetEdgeList(context.Background(), &proto.GetEdgeListReq{
		UserId: userId,
	})
	if err != nil {
		log.Error("get edge for %s fail: %v", userId, err)
		h.Response(ctx, nil, err)
		return
	}

	routeReply, err := h.ctrlCli.GetUserRoutes(context.Background(), &proto.GetUserRoutesReq{
		UserId: userId,
	})

	if err != nil {
		log.Error("get routes for %s fail: %v", userId, err)
		h.Response(ctx, nil, err)
		return
	}

	topology := generateTopolocy(edgeReply.Edges, routeReply.Routes)
	log.Debug("user %s topolocy %s", userId, topology)
	fp, err := os.Open(fmt.Sprintf("%s.dot", userId))
	if err != nil {
		log.Error("create dot file for user fail: %v", err)
		return
	}
	defer fp.Close()
	fp.Write([]byte(topology))

	h.Response(ctx, topology, nil)
}

func generateTopolocy(edges []*proto.EdgeInfo, routes []*proto.Route) string {
	// generate edges gateway link
	content := ""
	edgesGw := make(map[string]string)
	edgesRoute := make(map[string][]*proto.Route)

	// edge gateway link each other
	for i := 0; i < len(edges); i++ {
		content += fmt.Sprintf(`"%s" [style=filled;];`, edges[i].Name)
		for j := i + 1; j < len(edges); j++ {
			content += fmt.Sprintf(`"%s" -> "%s" [dir=none];`, edges[i].Name, edges[j].Name)
		}

		addr := fmt.Sprintf("%s:%d", edges[i].PublicIP, edges[i].PublicPort)
		edgesGw[addr] = edges[i].Name
	}

	// routes link to edge gateway
	for i := 0; i < len(routes); i++ {
		edgeName := edgesGw[routes[i].Nexthop]
		if len(edgeName) <= 0 {
			content += fmt.Sprintf(`"%s";`, routes[i].Cidr)
		} else {
			content += fmt.Sprintf(`"%s" [fillcolor=white, styled=dotted];`, routes[i].Cidr)
			content += fmt.Sprintf(`"%s" -> "%s"[arrowhead=vee];`, routes[i].Cidr, edgesGw[routes[i].Nexthop])
			edgesRoute[edgeName] = append(edgesRoute[edgeName], routes[i])
		}
	}

	// create subgraph
	subcount := 0
	for name, routes := range edgesRoute {
		subgraph := `subgraph cluster_%d { style=dashed; "%s";%s};`
		subcontent := ""
		for _, r := range routes {
			subcontent += fmt.Sprintf(`"%s";`, r.Cidr)
		}
		content += fmt.Sprintf(subgraph, subcount, name, subcontent)
		subcount += 1
	}

	rs := `digraph topology{
		rand=LR;
		%s}`
	return fmt.Sprintf(rs, content)
}
