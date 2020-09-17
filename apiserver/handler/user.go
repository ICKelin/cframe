package handler

import (
	"net/http"

	"github.com/ICKelin/cframe/pkg/auth"
	log "github.com/ICKelin/cframe/pkg/logs"
	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	BaseHandler
}

func (h *UserHandler) Run(eng *gin.Engine) {
	eng.POST("/api-service/v1/signup", h.signup)
	eng.POST("/api-service/v1/signin", h.signin)
}

func (h *UserHandler) signup(ctx *gin.Context) {
	f := SignupForm{}
	if ok := h.BindAndValidate(ctx, &f); !ok {
		return
	}

	// create user info
	user, err := auth.CreateUser(f.Username, f.Password)
	if err != nil {
		log.Error("%v", err)
		ctx.JSON(http.StatusOK, err)
		return
	}

	// create secret key
	err = auth.CreateSecret(user.SecretKey, user)
	if err != nil {
		log.Error("%v", err)
		ctx.JSON(http.StatusOK, err)
		return
	}
	ctx.JSON(http.StatusOK, user)
}

func (h *UserHandler) signin(ctx *gin.Context) {
	f := SigninForm{}
	if ok := h.BindAndValidate(ctx, &f); !ok {
		return
	}

	userInfo, err := auth.GetUser(f.Username, f.Password)
	if err != nil {
		log.Error("%v", err)
		ctx.JSON(http.StatusOK, err)
		return
	}

	token := ""
	err = auth.SetUserToken(token, userInfo)
	if err != nil {
		log.Error("%v", err)
		ctx.JSON(http.StatusOK, err)
		return
	}
	ctx.JSON(http.StatusOK, token)
}
