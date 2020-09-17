package main

import (
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

	edgeHandler := &handler.EdgeHandler{}
	userHandler := &handler.UserHandler{}

	edgeHandler.Run(eng)
	userHandler.Run(eng)

	eng.Run(s.addr)
}
