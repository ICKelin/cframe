package handler

import (
	"encoding/base64"
	"net/http"

	"github.com/ICKelin/cframe/pkg/auth"
	log "github.com/ICKelin/cframe/pkg/logs"
	"github.com/gin-gonic/gin"
	uuid "github.com/satori/go.uuid"
)

type UserHandler struct {
	BaseHandler
}

func (h *UserHandler) Run(eng *gin.Engine) {
	group := eng.Group("/api-service/v1/user")
	group.POST("/signup", h.signup)
	group.POST("/signin", h.signin)

	group.Use(h.MidAuth)
	{
		group.GET("/profile", h.getUserInfo)
	}
}

func (h *UserHandler) signup(ctx *gin.Context) {
	f := SignupForm{}
	if ok := h.BindAndValidate(ctx, &f); !ok {
		return
	}

	// TODO: verify user exist

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

	uniq := uuid.NewV4()
	token := base64.StdEncoding.EncodeToString(uniq.Bytes())
	err = auth.SetUserToken(token, userInfo)
	if err != nil {
		log.Error("%v", err)
		ctx.JSON(http.StatusOK, err)
		return
	}
	ctx.JSON(http.StatusOK, token)
}

func (h *UserHandler) getUserInfo(ctx *gin.Context) {
	userInfo := h.GetUserInfo(ctx)
	h.Response(ctx, userInfo, nil)
}
