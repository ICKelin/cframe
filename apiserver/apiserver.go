package main

import (
	"net/http"

	"github.com/ICKelin/cframe/apiserver/handler"
	"github.com/gin-gonic/gin"
)

type ApiServer struct {
	addr string
}

func NewApiServer(addr string) *ApiServer {
	return &ApiServer{
		addr: addr,
	}
}

func (s *ApiServer) Run() {
	eng := gin.New()
	eng.Use(MidCORS())

	edgeHandler := &handler.EdgeHandler{}
	userHandler := &handler.UserHandler{}

	edgeHandler.Run(eng)
	userHandler.Run(eng)

	eng.Run(s.addr)
}

func MidCORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Credentials", "true")
		c.Header("Access-Control-Allow-Methods", "*")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With, Access-Token, Platform, App-Version, Device-Model, System-Version, Language, Longitude, Latitude, App-Key, AppKey")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}
