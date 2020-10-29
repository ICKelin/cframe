package handler

import (
	"context"
	"fmt"

	"github.com/ICKelin/cframe/codec/proto"
	log "github.com/ICKelin/cframe/pkg/logs"
	"github.com/gin-gonic/gin"
)

type CSPHandler struct {
	BaseHandler
	ctrlCli proto.ControllerServiceClient
}

func NewCSPHandler(userCli proto.UserServiceClient, ctrl proto.ControllerServiceClient) *CSPHandler {
	return &CSPHandler{
		BaseHandler: BaseHandler{userCli: userCli},
		ctrlCli:     ctrl,
	}
}

func (h *CSPHandler) Run(eng *gin.Engine) {
	group := eng.Group("/api-service/v1/csp")
	group.Use(h.MidAuth)
	{
		group.GET("/list", h.getCSPList)
		group.POST("/add", h.addCSP)
		group.POST("/del", h.delCSP)
	}
}

func (h *CSPHandler) getCSPList(ctx *gin.Context) {
	userId := h.GetUserId(ctx)
	log.Debug("get csp list for user %s", userId)

	reply, err := h.ctrlCli.GetCSPList(context.Background(), &proto.GetCSPListReq{
		UserId: userId,
	})
	if err != nil {
		log.Error("get csp list for %s fail: %v", userId, err)
		h.Response(ctx, nil, err)
		return
	}

	h.Response(ctx, reply.CspInfo, nil)
}

func (h *CSPHandler) addCSP(ctx *gin.Context) {
	userId := h.GetUserId(ctx)
	log.Debug("add csp for user %s", userId)
	f := AddCSPForm{}
	if ok := h.BindAndValidate(ctx, &f); !ok {
		return
	}

	reply, err := h.ctrlCli.AddCSP(context.Background(), &proto.AddCSPReq{
		UserId:       userId,
		AccessKey:    f.AccessKey,
		AccessSecret: f.AccessSecret,
		CspType:      proto.CSPType(f.CSPType),
	})

	if err != nil {
		log.Error("add csp rpc fail: %v", err)
		h.Response(ctx, nil, err)
		return
	}

	if reply.Code != 0 {
		log.Error("add csp fail: %d %s", reply.Code, reply.Message)
		h.Response(ctx, nil, fmt.Errorf("<%d>%s", reply.Code, reply.Message))
		return
	}
	h.Response(ctx, reply.Data, nil)
}

func (h *CSPHandler) delCSP(ctx *gin.Context) {
	userId := h.GetUserId(ctx)
	log.Debug("delcsp for user %s", userId)
	f := DelCSPForm{}
	if ok := h.BindAndValidate(ctx, &f); !ok {
		return
	}

	reply, err := h.ctrlCli.DelCSP(context.Background(), &proto.DelCSPReq{
		UserId:  userId,
		CspType: proto.CSPType(f.CSPType),
	})

	if err != nil {
		log.Error("del csp rpc fail: %v", err)
		h.Response(ctx, nil, err)
		return
	}

	if reply.Code != 0 {
		log.Error("del csp fail: %d %s", reply.Code, reply.Message)
		h.Response(ctx, nil, fmt.Errorf("<%d>%s", reply.Code, reply.Message))
		return
	}
	h.Response(ctx, nil, nil)
}
