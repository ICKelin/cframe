package handler

import (
	"net/http"

	"github.com/ICKelin/cframe/pkg/auth"
	"github.com/gin-gonic/gin"
)

type ResponseBody struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

type BaseHandler struct{}

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
		Code:    0,
		Message: "success",
		Data:    data,
	}

	if err != nil {
		body.Code = 99999
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

	userInfo, err := auth.GetUserByToken(token)
	if err != nil || len(userInfo.Username) <= 0 {
		ctx.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	ctx.Set("username", userInfo.Username)
}
