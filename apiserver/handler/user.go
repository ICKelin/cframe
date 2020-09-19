package handler

import (
	"encoding/base64"

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
		h.Response(ctx, nil, err)
		return
	}

	// create secret key
	err = auth.CreateSecret(user.SecretKey, user)
	if err != nil {
		log.Error("%v", err)
		h.Response(ctx, nil, err)
		return
	}
	h.Response(ctx, user, err)
}

type SignInResponse struct {
	Token    string
	Username string
}

func (h *UserHandler) signin(ctx *gin.Context) {
	f := SigninForm{}
	if ok := h.BindAndValidate(ctx, &f); !ok {
		return
	}

	userInfo, err := auth.GetUser(f.Username, f.Password)
	if err != nil {
		log.Error("%v", err)
		h.Response(ctx, nil, err)
		return
	}

	uniq := uuid.NewV4()
	token := base64.StdEncoding.EncodeToString(uniq.Bytes())
	err = auth.SetUserToken(token, userInfo)
	if err != nil {
		log.Error("%v", err)
		h.Response(ctx, nil, err)
		return
	}

	signinResp := &SignInResponse{
		Token:    token,
		Username: f.Username,
	}
	h.Response(ctx, signinResp, nil)
}

func (h *UserHandler) getUserInfo(ctx *gin.Context) {
	userInfo := h.GetUserInfo(ctx)
	h.Response(ctx, userInfo, nil)
}
