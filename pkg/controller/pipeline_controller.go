package controller

import (
	"encoding/json"
	"fmt"
	"github.com/r3labs/sse/v2"
	flowRunGen "github.com/yametech/go-flowrun"
	baseResource "github.com/yametech/verthandi/pkg/api/resource/base"
	"github.com/yametech/verthandi/pkg/common"
	"github.com/yametech/verthandi/pkg/core"
	"github.com/yametech/verthandi/pkg/proc"
	"github.com/yametech/verthandi/pkg/resource/base"
	"github.com/yametech/verthandi/pkg/store"
	"github.com/yametech/verthandi/pkg/utils"
	"strings"
	"time"
)

var _ Controller = &PipelineController{}

type PipelineController struct {
	store.IStore
	proc *proc.Proc
}

func NewPipelineController(store store.IStore) *PipelineController {
	server := &PipelineController{
		IStore: store,
		proc:   proc.NewProc(),
	}
	return server
}

func (p *PipelineController) Run() error {
	p.proc.Add(p.recvPipeLine)
	p.proc.Add(p.watchEchoer)
	return <-p.proc.Start()
}

func (p *PipelineController) recvPipeLine(errC chan<- error) {
	pipeLineObjs, err := p.List(common.DefaultNamespace, common.Pipeline, "", map[string]interface{}{}, 0, 0)
	if err != nil {
		errC <- err
	}
	pipeLineCoder := store.GetResourceCoder(string(base.PipelineKind))
	if pipeLineCoder == nil {
		errC <- fmt.Errorf("(%s) %s", base.PipelineKind, "coder not exist")
	}
	pipeLineWatchChan := store.NewWatch(pipeLineCoder)

	var version int64
	for _, item := range pipeLineObjs {
		pipeLineObj := &base.Pipeline{}
		if err := core.UnmarshalInterfaceToResource(&item, pipeLineObj); err != nil {
			fmt.Printf("unmarshal step error %s\n", err)
		}
		if pipeLineObj.GetResourceVersion() > version {
			version = pipeLineObj.GetResourceVersion()
		}
		go p.handlePipeline(pipeLineObj)
	}

	p.Watch2(common.DefaultNamespace, common.Pipeline, version, pipeLineWatchChan)
	fmt.Println("pipelineController start watching pipeline")
	for {
		select {
		case item, ok := <-pipeLineWatchChan.ResultChan():
			if !ok {
				errC <- fmt.Errorf("recvPipeLine watch channal close")
			}
			if item.GetUUID() == "" {
				continue
			}
			pipeLineObj := &base.Pipeline{}
			if err := core.UnmarshalInterfaceToResource(&item, pipeLineObj); err != nil {
				fmt.Printf("receive pipeline UnmarshalInterfaceToResource error %s\n", err)
				continue
			}
			go p.handlePipeline(pipeLineObj)
		}
	}
}

func (p *PipelineController) watchEchoer(errC chan<- error) {
	version := time.Now().Add(-time.Hour * 1).Unix()
	url := fmt.Sprintf("%s/watch?resource=flowrun?version=%d", common.EchoerAddr, version)

	client := sse.NewClient(url)
	fmt.Println("[Echoer] start watch")
	err := client.SubscribeRaw(func(msg *sse.Event) {
		data := baseResource.FSMResp{}
		err := json.Unmarshal(msg.Data, &data)
		if err != nil {
			fmt.Println(err.Error())
		}
		p.handleFlowRun(&data)

	})
	if err != nil {
		errC <- err
	}
}

func (p *PipelineController) handlePipeline(pipeLine *base.Pipeline) {
	for _, stageUUID := range pipeLine.Spec.Stages {
		stage := &base.Stage{}
		err := p.GetByUUID(common.DefaultNamespace, common.Stage, stageUUID, stage)
		if err != nil {
			fmt.Printf("[Control] handlePipeline get stage error %s\n", err)
			return
		}
		if stage.Spec.Done == true {
			continue
		}
		for _, stepUUID := range stage.Spec.Steps {
			step := &base.Step{}
			err := p.GetByUUID(common.DefaultNamespace, common.Step, stepUUID, step)
			if err != nil {
				fmt.Printf("[Control] handlePipeline get step error %s\n", err)
				return
			}
			if step.Spec.StepStatus == base.Initializing && step.Spec.Trigger == true {
				fmt.Printf("[Control] handlePipeline send: %s step: %s to Echoer\n", step.UUID, step.Spec.ActionName)
				go p.sendEchoer(step)
				step.Spec.StepStatus = base.Sending
				_, _, err := p.Apply(common.DefaultNamespace, common.Step, step.UUID, step, false)
				if err != nil {
					fmt.Printf("[Control] handlePipeline apply step error %s\n", err)
				}
			}
		}
		break
	}
}

func (p *PipelineController) sendEchoer(step *base.Step) {
	if step.UUID == "" {
		step.UUID = utils.NewSUID().String()
	}

	flowRun := &flowRunGen.FlowRun{
		EchoerUrl: common.EchoerAddr,
		Name:      fmt.Sprintf("%s_%d", common.DefaultServerName, time.Now().UnixNano()),
	}
	flowRunStep := map[string]string{
		"SUCCESS": "done", "FAIL": "done",
	}
	stepName := step.Metadata.Name
	if stepName == "" {
		stepName = "pipeline"
	}

	flowRunStepName := fmt.Sprintf("%s_%s", stepName, step.UUID)
	flowRun.AddStep(flowRunStepName, flowRunStep, step.Spec.ActionName, step.Spec.Data)

	flowRunData := flowRun.Generate()
	fmt.Println(flowRunData)
	flowRun.Create(flowRunData)
}

func (p *PipelineController) handleFlowRun(f *baseResource.FSMResp) {
	flowName := strings.Split(f.Metadata.Name, "_")
	if len(flowName) != 2 || flowName[0] != common.DefaultServerName {
		return
	}

	// 遍历flow的所有step
	fmt.Printf("[Echoer] get flowrun %s\n", f.Metadata.Name)
	for _, flowStep := range f.Spec.Steps {
		//如果step未完成，就跳过等待
		if flowStep.Spec.ActionRun.Done != true {
			continue
		}
		stepUUID := strings.Split(flowStep.Metadata.Name, "_")[1]
		step := &base.Step{}
		err := p.GetByUUID(common.DefaultNamespace, common.Step, stepUUID, step)
		if err != nil {
			fmt.Printf("[Echoer] get step %s error: %s\n", stepUUID, err.Error())
			return
		}
		if flowStep.Spec.Response.State == "SUCCESS" {
			step.Spec.StepStatus = base.Finish
			fmt.Printf("[Echoer] change step: %s action: %s to success\n", step.UUID, step.Spec.ActionName)
		} else {
			step.Spec.StepStatus = base.Fail
			fmt.Printf("[Echoer] change step: %s action: %s to fail\n", step.UUID, step.Spec.ActionName)
		}
		_, _, err = p.Apply(common.DefaultNamespace, common.Step, step.UUID, step, false)
		if err != nil {
			fmt.Println("[Echoer] apply step error:", err)
		}
		p.reconcileStep(step)
	}
}

func (p *PipelineController) reconcileStep(stepObj *base.Step) {
	if stepObj.Spec.StepStatus != base.Finish {
		return
	}
	stage := &base.Stage{}
	err := p.GetByUUID(common.DefaultNamespace, common.Stage, stepObj.Spec.StageUUID, stage)
	if err != nil {
		fmt.Printf("[Echoer] reconcileStep get stage error %s\n", err)
		return
	}
	if stage.Spec.Done == true {
		p.reconcileStage(stage)
		return
	}
	stage.Spec.Done = true
	for _, stepUUID := range stage.Spec.Steps {
		if stepObj.UUID == stepUUID {
			continue
		}
		step := &base.Step{}
		err := p.GetByUUID(common.DefaultNamespace, common.Step, stepUUID, step)
		if err != nil {
			fmt.Printf("[Control] handlePipeline get step error %s\n", err)
			return
		}
		if step.Spec.StepStatus != base.Finish {
			stage.Spec.Done = false
			break
		}
	}
	_, _, err = p.Apply(common.DefaultNamespace, common.Stage, stage.UUID, stage, false)
	if err != nil {
		fmt.Printf("[Control] handlePipeline apply stage error %s\n", err)
		return
	}
	if stage.Spec.Done == true {
		p.reconcileStage(stage)
	}
}

func (p *PipelineController) reconcileStage(stageObj *base.Stage) {
	if stageObj.Spec.Done != true {
		return
	}
	pipeLine := &base.Pipeline{}
	err := p.GetByUUID(common.DefaultNamespace, common.Pipeline, stageObj.Spec.PipelineUUID, pipeLine)
	if err != nil {
		fmt.Printf("[Echoer] reconcileStage get pipeline error %s\n", err)
		return
	}
	if pipeLine.Spec.PipelineStatus == base.Finished {
		return
	}
	pipeLine.Spec.PipelineStatus = base.Finished
	for _, stageUUID := range pipeLine.Spec.Stages {
		if stageObj.UUID == stageUUID {
			continue
		}
		stage := &base.Stage{}
		err := p.GetByUUID(common.DefaultNamespace, common.Stage, stageUUID, stage)
		if err != nil {
			fmt.Printf("[Echoer] reconcileStage get Stage error %s\n", err)
			return
		}
		if stage.Spec.Done != true {
			pipeLine.Spec.PipelineStatus = base.Running
			break
		}
	}
	_, _, err = p.Apply(common.DefaultNamespace, common.Pipeline, pipeLine.UUID, pipeLine, true)
	if err != nil {
		fmt.Printf("[Control] reconcileStage apply pipeline error %s\n", err)
		return
	}
}
