package base

import (
	"github.com/yametech/verthandi/pkg/core"
	"github.com/yametech/verthandi/pkg/store"
	"github.com/yametech/verthandi/pkg/store/gtm"
)

/*
	pipeline是双重数组
	整条pipeline由多个Stage组成，每个Stage由多个Step组成
	Stage是串行，Step是并行
	pipeline{
		stage1->stage2->stage3
	}
	stage{
		->step1
		->step2
	}

*/

type PipelineStatus uint8

const (
	Running PipelineStatus = iota
	Finished
)

const PipelineKind core.Kind = "pipeline"

type PipelineSpec struct {
	Stages         []string `json:"stages" bson:"stages"`
	LastState      string   `json:"last_state" bson:"last_state"`
	PipelineStatus `json:"pipeline_status" bson:"pipeline_status"`
}

type Pipeline struct {
	core.Metadata `json:"metadata"`
	Spec          PipelineSpec `json:"spec"`
}

// Pipeline impl Coder
func (*Pipeline) Decode(op *gtm.Op) (core.IObject, error) {
	action := &Pipeline{}
	if err := core.ObjectToResource(op.Data, action); err != nil {
		return nil, err
	}
	return action, nil
}

func (pl *Pipeline) Clone() core.IObject {
	result := &Pipeline{}
	core.Clone(pl, result)
	return result
}

func init() {
	store.AddResourceCoder(string(PipelineKind), &Pipeline{})
}
