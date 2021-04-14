package api

import (
	"github.com/gin-gonic/gin"
	"github.com/yametech/echoer-flow-tool/pkg/service"
)

type Extends interface {
	Name() string
}

type Server struct {
	Extends
	*gin.Engine
	service.IService
}

func NewServer(p service.IService) *Server {
	return &Server{
		Engine:   gin.Default(),
		IService: p,
	}
}
