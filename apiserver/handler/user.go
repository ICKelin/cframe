package handler

import (
	"context"
	"fmt"

	"github.com/ICKelin/cframe/codec/proto"
	log "github.com/ICKelin/cframe/pkg/logs"
	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	BaseHandler
}

func NewUserHandler(userCli proto.UserServiceClient) *UserHandler {
	return &UserHandler{
		BaseHandler: BaseHandler{userCli: userCli},
	}
}

func (h *UserHandler) Run(eng *gin.Engine) {
	group := eng.Group("/api-service/v1/user")
	group.POST("/signup", h.signup)
	group.POST("/signin", h.signin)
}

func (h *UserHandler) signup(ctx *gin.Context) {
	f := SignupForm{}
	if ok := h.BindAndValidate(ctx, &f); !ok {
		return
	}

	req := &proto.AddUserReq{
		UserName: f.Username,
		Password: f.Password,
		Email:    f.Email,
		About:    f.About,
	}
	reply, err := h.userCli.AddUser(context.Background(), req)
	if err != nil {
		log.Error("add user rpc fail: %v", err)
		h.Response(ctx, nil, err)
		return
	}

	if reply.Code != 0 {
		log.Error("add user fail: %d %s", reply.Code, reply.Message)
		h.Response(ctx, nil, fmt.Errorf("<%d>%s", reply.Code, reply.Message))
		return
	}
	h.Response(ctx, reply.UserInfo, nil)
}

func (h *UserHandler) signin(ctx *gin.Context) {
	f := SigninForm{}
	if ok := h.BindAndValidate(ctx, &f); !ok {
		return
	}

	req := &proto.AuthorizeReq{
		Username: f.Username,
		Password: f.Password,
	}
	reply, err := h.userCli.Authorize(context.Background(), req)
	if err != nil {
		log.Error("auth rpc fail: %v", err)
		h.Response(ctx, nil, err)
		return
	}

	if reply.Code != 0 {
		log.Error("auth fail: %d %s", reply.Code, reply.Message)
		h.Response(ctx, nil, fmt.Errorf("<%d>%s", reply.Code, reply.Message))
		return
	}
	h.Response(ctx, reply.Data, nil)
}
