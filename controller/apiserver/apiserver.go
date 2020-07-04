package apiserver

import (
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
	eng.POST("/api-service/v1/edage/add", s.addEdage)
	eng.POST("/api-service/v1/edage/del", s.delEdage)
	eng.GET("/api-service/v1/edage/list", s.getEdageList)

	eng.Run(s.addr)
}

func (s *ApiServer) addEdage(ctx *gin.Context)     {}
func (s *ApiServer) delEdage(ctx *gin.Context)     {}
func (s *ApiServer) getEdageList(ctx *gin.Context) {}
