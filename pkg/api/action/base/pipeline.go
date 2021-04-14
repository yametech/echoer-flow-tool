package base

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/yametech/verthandi/pkg/api"
	apiResource "github.com/yametech/verthandi/pkg/api/resource"
	"net/http"
)

func (s *baseServer) ListPipeLine(g *gin.Context) {
	g.JSON(http.StatusOK, "")
}

func (s *baseServer) CreatePipeLine(g *gin.Context) {
	rawData, err := g.GetRawData()
	if err != nil {
		api.RequestParamsError(g, "get rawData error", err)
		return
	}
	request := &apiResource.RequestPipeLine{}
	if err := json.Unmarshal(rawData, request); err != nil {
		api.RequestParamsError(g, "unmarshal json error", err)
		return
	}
	err = s.PipeLineService.Create(request)
	if err != nil {
		api.RequestParamsError(g, "get rawData error", err)
		return
	}
	g.JSON(http.StatusOK, request)
}
