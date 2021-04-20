package base

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/yametech/verthandi/pkg/api"
	"github.com/yametech/verthandi/pkg/api/resource/base"
	base2 "github.com/yametech/verthandi/pkg/api/resource/base"
	"github.com/yametech/verthandi/pkg/common"
	"github.com/yametech/verthandi/pkg/utils"
	"io"
	"net/http"
)

func (s *baseServer) ListPipeLine(g *gin.Context) {
	pipelines, err := s.PipeLineService.List()
	if err != nil {
		api.RequestParamsError(g, "get pipelines error", err)
		return
	}
	g.JSON(http.StatusOK, pipelines)
}

func (s *baseServer) CreatePipeLine(g *gin.Context) {
	rawData, err := g.GetRawData()
	if err != nil {
		api.RequestParamsError(g, "get rawData error", err)
		return
	}
	request := &base.RequestPipeLine{}
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

func (s *baseServer) WatchPipeLine(g *gin.Context) {
	version := g.DefaultQuery("version", "0")
	objectChan, closed := s.PipeLineService.Watch(common.Pipeline, string(base2.RespPipelineKind), version)
	streamEndEvent := "STREAM_END"

	g.Stream(func(w io.Writer) bool {
		select {
		case <-g.Writer.CloseNotify():
			closed <- struct{}{}
			close(closed)
			g.SSEvent("", streamEndEvent)
			return false
		case object, ok := <-objectChan:
			if !ok {
				g.SSEvent("", streamEndEvent)
				return false
			}
			pipeLine := &base2.RespPipeline{}
			err := utils.UnstructuredObjectToInstanceObj(object, pipeLine)
			if err != nil {
				fmt.Println("pipeline watch Unmarshal err", err)
				return true
			}
			s.PipeLineService.ReconcilePipeLine(pipeLine)
			g.SSEvent("", pipeLine)
		}
		return true
	},
	)
}

func (s *baseServer) WatchStep(g *gin.Context) {
	version := g.DefaultQuery("version", "0")
	objectChan, closed := s.PipeLineService.Watch(common.Step, string(base2.RespStepKind), version)
	streamEndEvent := "STREAM_END"

	g.Stream(func(w io.Writer) bool {
		select {
		case <-g.Writer.CloseNotify():
			closed <- struct{}{}
			close(closed)
			g.SSEvent("", streamEndEvent)
			return false
		case object, ok := <-objectChan:
			if !ok {
				g.SSEvent("", streamEndEvent)
				return false
			}
			g.SSEvent("", object)
		}
		return true
	},
	)
}
