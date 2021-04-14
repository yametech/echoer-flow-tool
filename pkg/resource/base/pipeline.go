package base

import (
	"github.com/yametech/verthandi/pkg/core"
	"github.com/yametech/verthandi/pkg/store"
	"github.com/yametech/verthandi/pkg/store/gtm"
)

type PipelineStatus uint8

const (
	Running PipelineStatus = iota
	Waiting
	Finished
)

const PipelineKind core.Kind = "pipeline"

type PipelineSpec struct {
	Steps          [][]map[string]interface{} `json:"steps" bson:"steps"`
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
