package base

import "github.com/yametech/echoer-flow-tool/pkg/service"

type PipeLineService struct {
	service.IService
}

func NewPipeLineService(i service.IService) *PipeLineService {
	return &PipeLineService{i}
}
