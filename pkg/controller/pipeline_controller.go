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
	p.proc.Add(p.recv)
	p.proc.Add(p.WatchEchoer)
	return <-p.proc.Start()
}

func (p *PipelineController) recv(errC chan<- error) {
	stepObjs, err := p.List(common.DefaultNamespace, common.Step, "", map[string]interface{}{}, 0, 0)
	if err != nil {
		errC <- err
	}
	stepCoder := store.GetResourceCoder(string(base.StepKind))
	if stepCoder == nil {
		errC <- fmt.Errorf("(%s) %s", base.PipelineKind, "coder not exist")
	}
	stepWatchChan := store.NewWatch(stepCoder)

	var version int64
	for _, item := range stepObjs {
		stepObj := &base.Step{}
		if err := core.UnmarshalInterfaceToResource(&item, stepObj); err != nil {
			fmt.Printf("unmarshal step error %s\n", err)
		}
		if stepObj.GetResourceVersion() > version {
			version = stepObj.GetResourceVersion()
		}
		go p.handleStep(stepObj)
	}

	p.Watch2(common.DefaultNamespace, common.Step, version, stepWatchChan)
	fmt.Println("pipelineController start watching step")
	for {
		select {
		case item, ok := <-stepWatchChan.ResultChan():
			if !ok {
				errC <- fmt.Errorf("handleStep recv watch channal close")
			}
			if item.GetUUID() == "" {
				continue
			}
			stepObj := &base.Step{}
			if err := core.UnmarshalInterfaceToResource(&item, stepObj); err != nil {
				fmt.Printf("receive step UnmarshalInterfaceToResource error %s\n", err)
				continue
			}
			go p.handleStep(stepObj)
		}
	}
}

func (p *PipelineController) WatchEchoer(errC chan<- error) {
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
		go p.handleEchoer(&data)

	})
	if err != nil {
		errC <- err
	}
}

func (p *PipelineController) handleStep(step *base.Step) {
	switch step.Spec.StepStatus {
	case base.Initializing:
		if !step.Spec.Trigger {
			return
		}
		fmt.Printf("[Control] handleStep send %s to Echoer\n", step.UUID)
		p.sendEchoer(step)
		step.Spec.StepStatus = base.Sending
		_, _, err := p.Apply(common.DefaultNamespace, common.Step, step.UUID, step)
		if err != nil {
			fmt.Printf("[Control] handleStep apply step error %s\n", err)
		}
	case base.Finish:
		fmt.Printf("[Control] step %s finish start reconcileStage %s\n", step.UUID, step.Spec.StageUUID)
		go p.reconcileStage(step.Spec.StageUUID)
	}
}

func (p *PipelineController) reconcileStage(stageUUID string) {
	stage := &base.Stage{}
	err := p.GetByUUID(common.DefaultNamespace, common.Stage, stageUUID, stage)
	if err != nil {
		fmt.Printf("[Control] reconcileStage get stage %s error %s\n", stageUUID, err)
		return
	}
	// 如果stage已经完成，直接检查pipeline
	if stage.Spec.Done {
		p.reconcilePipeline(stage.Spec.PipelineUUID)
		return
	}
	stage.Spec.Done = true
	for _, stepStr := range stage.Spec.Steps {
		step := &base.Step{}
		err := p.GetByUUID(common.DefaultNamespace, common.Step, stepStr, step)
		if err != nil {
			fmt.Printf("[Control] reconcileStage get step error %s\n", err)
			return
		}
		stage.Spec.LastState = fmt.Sprintf("%s-%s", "step", step.UUID)
		if step.Spec.StepStatus != base.Finish {
			stage.Spec.Done = false
			break
		}
	}
	_, _, err = p.Apply(common.DefaultNamespace, common.Stage, stage.UUID, stage)
	if err != nil {
		fmt.Printf("[Control] reconcileStage apply stage error %s\n", err)
	}
	p.reconcilePipeline(stage.Spec.PipelineUUID)
}

func (p *PipelineController) reconcilePipeline(pipeLineUUID string) {
	pipeLine := &base.Pipeline{}
	err := p.GetByUUID(common.DefaultNamespace, common.Pipeline, pipeLineUUID, pipeLine)
	if err != nil {
		fmt.Printf("[Control] reconcilePipeline get pipeLine error %s\n", err)
		return
	}
	if pipeLine.Spec.PipelineStatus == base.Finished {
		return
	}

	pipeLine.Spec.PipelineStatus = base.Finished
	for _, stageStr := range pipeLine.Spec.Stages {
		stage := &base.Stage{}
		err := p.GetByUUID(common.DefaultNamespace, common.Stage, stageStr, stage)
		if err != nil {
			fmt.Printf("[Control] reconcilePipeline get stage error %s\n", err)
			return
		}
		pipeLine.Spec.LastState = stage.Spec.LastState
		if stage.Spec.Done != true {
			pipeLine.Spec.PipelineStatus = base.Running
			break
		}
	}
	_, _, err = p.Apply(common.DefaultNamespace, common.Pipeline, pipeLine.UUID, pipeLine)
	if err != nil {
		fmt.Printf("[Control] reconcilePipeline apply pipeLine error %s\n", err)
	}
}

func (p *PipelineController) sendEchoer(step *base.Step) {
	if step.UUID == "" {
		step.UUID = utils.NewSUID().String()
	}

	flowRun := &flowRunGen.FlowRun{
		EchoerUrl: common.EchoerAddr,
		Name:      fmt.Sprintf("%s_%d", common.DefaultServerName, time.Now().Unix()),
	}
	flowRunStep := map[string]string{
		"SUCCESS": "done", "FAIL": "done",
	}
	stepName := step.Metadata.Name
	if stepName == "" {
		stepName = "pipeline"
	}

	flowRunStepName := fmt.Sprintf("%s_%s", stepName, step.UUID)
	step.Spec.Data["retryCount"] = 15
	flowRun.AddStep(flowRunStepName, flowRunStep, step.Spec.ActionName, step.Spec.Data)

	flowRunData := flowRun.Generate()
	fmt.Println(flowRunData)
	flowRun.Create(flowRunData)
}

func (p *PipelineController) handleEchoer(f *baseResource.FSMResp) {
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
		} else {
			step.Spec.StepStatus = base.Fail
		}
		_, _, err = p.Apply(common.DefaultNamespace, common.Step, step.UUID, step)
		if err != nil {
			fmt.Println("[Echoer] apply step error:", err)
		}
	}
}

//func echoerPost(flowRunData []byte) {
//	url := fmt.Sprintf("%s/flowrun", common.EchoerAddr)
//	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(flowRunData))
//	req.Header.Set("Content-Type", "application/json")
//	client := &http.Client{}
//	resp, err := client.Do(req)
//	if err != nil {
//		fmt.Println("[Control] echoerPost error", err)
//		return
//	}
//	defer resp.Body.Close()
//	if resp.StatusCode != 200 {
//		body, _ := ioutil.ReadAll(resp.Body)
//		fmt.Printf("[Control] echoerPost status error return msg: %s\n", body)
//	}
//}
