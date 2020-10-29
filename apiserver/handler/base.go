package handler

import (
	"context"
	"net/http"

	"github.com/ICKelin/cframe/codec/proto"
	log "github.com/ICKelin/cframe/pkg/logs"
	"github.com/gin-gonic/gin"
)

const (
	CODE_SUCC = 20000
	CODE_FAIL = 50000
)

type ResponseBody struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

type BaseHandler struct {
	userCli proto.UserServiceClient
}

func (h *BaseHandler) BindAndValidate(ctx *gin.Context, obj interface{}) bool {
	err := ctx.BindJSON(obj)
	if err != nil {
		h.Response(ctx, nil, err)
		return false
	}
	return true
}

func (h *BaseHandler) Response(ctx *gin.Context, data interface{}, err error) {
	body := &ResponseBody{
		Code:    CODE_SUCC,
		Message: "success",
		Data:    data,
	}

	if err != nil {
		body.Code = CODE_FAIL
		body.Message = err.Error()
		body.Data = nil
	}

	ctx.JSON(http.StatusOK, body)
}

func (h *BaseHandler) MidAuth(ctx *gin.Context) {
	token := ctx.GetHeader("Access-Token")
	if len(token) <= 0 {
		ctx.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	reply, err := h.userCli.GetUserByToken(context.Background(), &proto.GetUserByTokenReq{
		Token: token,
	})
	if err != nil {
		log.Error("get user by token %s fail %s", token, err)
		ctx.AbortWithStatus(http.StatusUnauthorized)
		return
	}
	if reply.Code != 0 {
		log.Error("get user by token %s fail: %d %s", token, reply.Code, reply.Message)
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, reply.Message)
		return
	}

	ctx.Set("username", reply.Data.UserName)
	ctx.Set("userId", reply.Data.UserId)
}

func (h *BaseHandler) GetUsername(ctx *gin.Context) string {
	obj, _ := ctx.Get("username")
	username, _ := obj.(string)
	return username
}

func (h *BaseHandler) GetUserId(ctx *gin.Context) string {
	obj, _ := ctx.Get("userId")
	userId, _ := obj.(string)
	return userId
}
