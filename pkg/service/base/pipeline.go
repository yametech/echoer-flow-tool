package base

import (
	base2 "github.com/yametech/verthandi/pkg/api/resource/base"
	"github.com/yametech/verthandi/pkg/common"
	"github.com/yametech/verthandi/pkg/core"
	"github.com/yametech/verthandi/pkg/resource/base"
	"github.com/yametech/verthandi/pkg/service"
	"github.com/yametech/verthandi/pkg/utils"
)

type PipeLineService struct {
	service.IService
}

func NewPipeLineService(i service.IService) *PipeLineService {
	return &PipeLineService{i}
}

func (p *PipeLineService) Watch(resource, kind, version string) (chan core.IObject, chan struct{}) {
	objectChan := make(chan core.IObject, 32)
	closed := make(chan struct{})
	p.IService.Watch(common.DefaultNamespace, resource, kind, version, objectChan, closed)
	return objectChan, closed
}

func (p *PipeLineService) List() ([]*base2.RespPipeline, error) {
	unStructPipeLines, err := p.IService.List(common.DefaultNamespace, common.Pipeline, "", map[string]interface{}{}, 0, 0)
	if err != nil {
		return nil, err
	}
	pipelines := make([]*base2.RespPipeline, 0)
	err = utils.UnstructuredObjectToInstanceObj(unStructPipeLines, &pipelines)
	if err != nil {
		return nil, err
	}
	// 对每个pipeline的stage和step都进行加载
	for _, pipeline := range pipelines {
		p.ReconcilePipeLine(pipeline)
	}
	return pipelines, nil
}

func (p *PipeLineService) Create(request *base2.RequestPipeLine) error {
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
					ActionName: reqStep.ActionName,
					Data:       reqStep.Data,
					Trigger:    reqStep.Trigger,
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

func (p *PipeLineService) ReconcilePipeLine(pipeline *base2.RespPipeline) {
	for _, stageUUID := range pipeline.Spec.Stages {
		stage := base2.RespStage{}
		err := p.GetByUUID(common.DefaultNamespace, common.Stage, stageUUID, &stage)
		if err != nil {
			continue
		}
		for _, stepUUID := range stage.Spec.Steps {
			step := base2.RespStep{}

			err := p.GetByUUID(common.DefaultNamespace, common.Step, stepUUID, &step)
			if err != nil {
				continue
			}
			stage.Spec.StepsData = append(stage.Spec.StepsData, step)
		}
		pipeline.Spec.StagesData = append(pipeline.Spec.StagesData, stage)
	}
}
