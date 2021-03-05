package handler

import (
	"context"
	"fmt"

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
	// h.ctrlCli.DelRoute(context.Background(), &proto.DelRouteReq{})
}
