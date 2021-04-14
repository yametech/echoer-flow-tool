package controller

import (
	"fmt"
	"github.com/yametech/verthandi/pkg/common"
	"github.com/yametech/verthandi/pkg/core"
	"github.com/yametech/verthandi/pkg/resource/base"
	"github.com/yametech/verthandi/pkg/store"
)

var _ Controller = &PipelineController{}

type PipelineController struct {
	stop chan struct{}
	store.IStore
	tqStop chan struct{}
	tq     *Queue
}

func NewPipelineController(stage store.IStore) *PipelineController {
	server := &PipelineController{
		stop:   make(chan struct{}),
		IStore: stage,
		tqStop: make(chan struct{}),
		tq:     &Queue{},
	}
	return server
}

func (p *PipelineController) Run() error {
	return p.recv()
}

func (p *PipelineController) Stop() error {
	p.tqStop <- struct{}{}
	p.stop <- struct{}{}
	return nil
}

func (p *PipelineController) recv() error {
	pipeLineObjs, _, err := p.List(common.DefaultNamespace, common.Pipeline, "", map[string]interface{}{}, 0, 0)
	if err != nil {
		return err
	}
	pipeLineCoder := store.GetResourceCoder(string(base.PipelineKind))
	if pipeLineCoder == nil {
		return fmt.Errorf("(%s) %s", base.PipelineKind, "coder not exist")
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
		case <-p.stop:
			pipeLineWatchChan.CloseStop() <- struct{}{}
			return nil
		case item, ok := <-pipeLineWatchChan.ResultChan():
			if !ok {
				return nil
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

func (p *PipelineController) reconcilePipeline(pipeLineObj *base.Pipeline) error {
	fmt.Println(*pipeLineObj)
	return nil
}
