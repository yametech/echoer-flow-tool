package base

import (
	apiResource "github.com/yametech/verthandi/pkg/api/resource"
	"github.com/yametech/verthandi/pkg/common"
	"github.com/yametech/verthandi/pkg/core"
	"github.com/yametech/verthandi/pkg/resource/base"
	"github.com/yametech/verthandi/pkg/service"
)

type PipeLineService struct {
	service.IService
}

func NewPipeLineService(i service.IService) *PipeLineService {
	return &PipeLineService{i}
}

func (p *PipeLineService) Create(request *apiResource.RequestPipeLine) error {
	pipeline := &base.Pipeline{
		Metadata: core.Metadata{
			Name: request.Name,
			Kind: string(base.PipelineKind),
		},
		Spec: base.PipelineSpec{
			Steps: request.Step,
		},
	}

	pipeline.GenerateVersion()
	_, err := p.IService.Create(common.DefaultNamespace, common.Pipeline, pipeline)
	return err
}
