package main

import (
	"net/http"

	"github.com/ICKelin/cframe/apiserver/handler"
	"github.com/ICKelin/cframe/codec/proto"
	log "github.com/ICKelin/cframe/pkg/logs"
	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"
)

type ApiServer struct {
	addr           string
	userCenterAddr string
	ctrlAddr       string
}

func NewApiServer(addr, usercenter, ctrl string) *ApiServer {
	return &ApiServer{
		addr:           addr,
		userCenterAddr: usercenter,
		ctrlAddr:       ctrl,
	}
}

func (s *ApiServer) Run() {
	eng := gin.New()
	eng.Use(MidCORS())

	userCli, err := createUserServiceCli(s.userCenterAddr)
	if err != nil {
		log.Error("create user service client fail: %v", err)
		return
	}

	ctrlCli, err := createCtrlCli(s.ctrlAddr)
	if err != nil {
		log.Error("create controller client fail: %v", err)
		return
	}

	edgeHandler := handler.NewEdgeHandler(userCli, ctrlCli)
	userHandler := handler.NewUserHandler(userCli)
	cspHandler := handler.NewCSPHandler(userCli, ctrlCli)
	routeHandler := handler.NewRouteHandler(userCli, ctrlCli)

	edgeHandler.Run(eng)
	userHandler.Run(eng)
	cspHandler.Run(eng)
	routeHandler.Run(eng)

	eng.Static("/public", "./static")

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

func createUserServiceCli(remote string) (proto.UserServiceClient, error) {
	conn, err := grpc.Dial(remote, grpc.WithInsecure())
	if err != nil {
		return nil, err
	}
	cli := proto.NewUserServiceClient(conn)
	return cli, nil
}

func createCtrlCli(remote string) (proto.ControllerServiceClient, error) {
	conn, err := grpc.Dial(remote, grpc.WithInsecure())
	if err != nil {
		return nil, err
	}
	cli := proto.NewControllerServiceClient(conn)
	return cli, nil
}
