package base

import (
	"fmt"
	"github.com/yametech/verthandi/pkg/api"
	baseService "github.com/yametech/verthandi/pkg/service/base"
)

type baseServer struct {
	*api.Server
	*baseService.PipeLineService
}

func NewBaseServer(serviceName string, server *api.Server) *baseServer {
	base := &baseServer{
		Server:          server,
		PipeLineService: baseService.NewPipeLineService(server.IService),
	}
	group := base.Group(fmt.Sprintf("/%s", serviceName))

	//base
	{
		group.GET("/pipelinewatch", base.WatchPipeLine)
		group.GET("/stepwatch", base.WatchStep)

		group.GET("/pipelines", base.ListPipeLine)
		group.POST("/pipeline", base.CreatePipeLine)
	}

	return base
}
