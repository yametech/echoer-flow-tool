package controller

import (
	"encoding/json"
	"fmt"
	"github.com/r3labs/sse/v2"
	"github.com/yametech/verthandi/pkg/common"
	"github.com/yametech/verthandi/pkg/core"
	"github.com/yametech/verthandi/pkg/proc"
	"github.com/yametech/verthandi/pkg/resource/base"
	"github.com/yametech/verthandi/pkg/store"
	"time"
)

var _ Controller = &PipelineController{}

type PipelineController struct {
	store.IStore
	proc *proc.Proc
}

func NewPipelineController(stage store.IStore) *PipelineController {
	server := &PipelineController{
		IStore: stage,
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
	pipeLineObjs, _, err := p.List(common.DefaultNamespace, common.Pipeline, "", map[string]interface{}{}, 0, 0)
	if err != nil {
		errC <- err
	}
	pipeLineCoder := store.GetResourceCoder(string(base.PipelineKind))
	if pipeLineCoder == nil {
		errC <- fmt.Errorf("(%s) %s", base.PipelineKind, "coder not exist")
	}
	pipeLineWatchChan := store.NewWatch(pipeLineCoder)

	version := int64(0)
	for _, item := range pipeLineObjs {
		pipeLineObj := &base.Pipeline{}
		if err := core.UnmarshalInterfaceToResource(&item, pipeLineObj); err != nil {
			fmt.Printf("[ERROR] unmarshal pipeLine error %s\n", err)
		}
		if pipeLineObj.GetResourceVersion() > version {
			version = pipeLineObj.GetResourceVersion()
		}
		if err := p.reconcilePipeline(pipeLineObj); err != nil {
			fmt.Printf("[ERROR] reconcilePipeline error: %s", err.Error())
		}
	}

	p.Watch2(common.DefaultNamespace, common.Pipeline, version, pipeLineWatchChan)

	for {
		select {

		case item, ok := <-pipeLineWatchChan.ResultChan():
			if !ok {
				errC <- fmt.Errorf("reconcilePipeline recv watch channal close")
			}
			if item.GetName() == "" {
				continue
			}
			pipeLineObj := &base.Pipeline{}
			if err := core.UnmarshalInterfaceToResource(&item, pipeLineObj); err != nil {
				fmt.Printf("[ERROR] receive pipeLine UnmarshalInterfaceToResource error %s\n", err)
				continue
			}
			if err := p.reconcilePipeline(pipeLineObj); err != nil {
				fmt.Printf("[ERROR] reconcilePipeline error: %s", err.Error())
			}
		}
	}
}

func (p *PipelineController) WatchEchoer(errC chan<- error) {
	version := time.Now().Add(-time.Hour * 1).Unix()
	url := fmt.Sprintf("%s/watch?resource=flowrun?version=%d", common.EchoerAddr, version)

	client := sse.NewClient(url)
	err := client.SubscribeRaw(func(msg *sse.Event) {
		data := make(map[string]interface{}, 0)
		err := json.Unmarshal(msg.Data, &data)
		if err != nil {
			fmt.Println(err.Error())
		}
		//TODO: watch echoer 拿到data后进行处理

	})
	if err != nil {
		errC <- err
	}
}

func (p *PipelineController) reconcilePipeline(pipeLineObj *base.Pipeline) error {
	//TODO: 检测到pipeline后，要发给FSM
	fmt.Println(*pipeLineObj)
	return nil
}
