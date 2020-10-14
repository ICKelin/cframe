package handler

import (
	"context"
	"fmt"

	"github.com/ICKelin/cframe/codec/proto"
	log "github.com/ICKelin/cframe/pkg/logs"
	"github.com/gin-gonic/gin"
)

type EdgeHandler struct {
	BaseHandler
	ctrlCli proto.ControllerServiceClient
}

func NewEdgeHandler(userCli proto.UserServiceClient, ctrl proto.ControllerServiceClient) *EdgeHandler {
	return &EdgeHandler{
		BaseHandler: BaseHandler{userCli: userCli},
		ctrlCli:     ctrl,
	}
}

func (h *EdgeHandler) Run(eng *gin.Engine) {
	group := eng.Group("/api-service/v1/edge")
	group.Use(h.MidAuth)
	{
		group.POST("/add", h.addEdge)
		group.GET("/list", h.getEdgeList)
		group.POST("/del", h.delEdge)
		group.POST("/stat", h.getEdgeStat)
	}
}

func (h *EdgeHandler) addEdge(ctx *gin.Context) {
	userId := h.GetUserId(ctx)
	log.Debug("add edge list for user %s", userId)

	addForm := AddEdgeForm{}
	if ok := h.BindAndValidate(ctx, &addForm); !ok {
		return
	}

	reply, err := h.ctrlCli.AddEdge(context.Background(), &proto.AddEdgeReq{
		UserId:     userId,
		Name:       addForm.Name,
		PublicIP:   addForm.PublicIP,
		Cidr:       addForm.Cidr,
		PublicPort: addForm.PublicPort,
		CspType:    proto.CSPType(addForm.CSPType),
		Comment:    addForm.Comment,
	})
	if err != nil {
		log.Error("add edge rpc fail: %v", err)
		h.Response(ctx, nil, err)
		return
	}

	if reply.Code != 0 {
		log.Error("add edge fail: %d %s", reply.Code, reply.Message)
		h.Response(ctx, nil, fmt.Errorf("<%d>%s", reply.Code, reply.Message))
		return
	}
	h.Response(ctx, reply.Data, nil)
}

func (h *EdgeHandler) getEdgeList(ctx *gin.Context) {
	userId := h.GetUserId(ctx)
	log.Debug("get edge list for user %s", userId)

	reply, err := h.ctrlCli.GetEdgeList(context.Background(), &proto.GetEdgeListReq{
		UserId: userId,
	})
	if err != nil {
		log.Error("add edge rpc fail: %v", err)
		h.Response(ctx, nil, err)
		return
	}

	if reply.Code != 0 {
		log.Error("add edge fail: %d %s", reply.Code, reply.Message)
		h.Response(ctx, nil, fmt.Errorf("<%d>%s", reply.Code, reply.Message))
		return
	}
	h.Response(ctx, reply.Edges, nil)
}

func (h *EdgeHandler) delEdge(ctx *gin.Context) {
	userId := h.GetUserId(ctx)
	log.Debug("del edge list for user %s", userId)
	f := DelEdgeForm{}
	if ok := h.BindAndValidate(ctx, &f); !ok {
		return
	}

	reply, err := h.ctrlCli.DelEdge(context.Background(), &proto.DelEdgeReq{
		UserId:   userId,
		EdgeName: f.Name,
	})
	if err != nil {
		log.Error("add edge rpc fail: %v", err)
		h.Response(ctx, nil, err)
		return
	}

	if reply.Code != 0 {
		log.Error("add edge fail: %d %s", reply.Code, reply.Message)
		h.Response(ctx, nil, fmt.Errorf("<%d>%s", reply.Code, reply.Message))
		return
	}
	h.Response(ctx, nil, nil)
}

func (h *EdgeHandler) getEdgeStat(ctx *gin.Context) {
	userId := h.GetUserId(ctx)
	log.Debug("get edge stat for user %s", userId)
	f := GetStatForm{}
	if ok := h.BindAndValidate(ctx, &f); !ok {
		return
	}

	reply, err := h.ctrlCli.GetStat(context.Background(), &proto.GetStatReq{
		UserId:    userId,
		EdgeName:  f.Name,
		From:      f.From,
		Count:     f.Count,
		Direction: f.Direction,
	})

	if err != nil {
		log.Error("get edge stat rpc fail: %v", err)
		h.Response(ctx, nil, err)
		return
	}

	if reply.Code != 0 {
		log.Error("get edge stat fail: %d %s", reply.Code, reply.Message)
		h.Response(ctx, nil, fmt.Errorf("<%d>%s", reply.Code, reply.Message))
		return
	}
	h.Response(ctx, reply.Stats, nil)
}
