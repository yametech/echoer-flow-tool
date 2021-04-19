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
	}
	pipeline.GenerateVersion()

	for _, reqStage := range request.Stage {
		// 多个stage，初始化
		stage := base.Stage{}
		stage.Spec.PipelineUUID = pipeline.UUID
		stage.GenerateVersion()

		for _, reqStep := range reqStage.Steps {
			// 多个step，初始化
			step := base.Step{
				Spec: base.StepSpec{
					Type:    base.StepType(reqStep.Type),
					Data:    reqStep.Data,
					Trigger: reqStep.Trigger,
				},
				Metadata: core.Metadata{
					Kind: string(base.StepKind),
				},
			}
			step.Spec.StageUUID = stage.UUID
			step.Spec.PipelineUUID = pipeline.UUID
			step.GenerateVersion()

			stage.Spec.Steps = append(stage.Spec.Steps, step.UUID)
			_, err := p.IService.Create(common.DefaultNamespace, common.Step, &step)
			if err != nil {
				return err
			}
		}
		pipeline.Spec.Stages = append(pipeline.Spec.Stages, stage.UUID)
		_, err := p.IService.Create(common.DefaultNamespace, common.Stage, &stage)
		if err != nil {
			return err
		}

	}

	_, err := p.IService.Create(common.DefaultNamespace, common.Pipeline, pipeline)
	return err
}
