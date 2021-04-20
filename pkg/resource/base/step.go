package base

import (
	"github.com/yametech/verthandi/pkg/core"
	"github.com/yametech/verthandi/pkg/store"
	"github.com/yametech/verthandi/pkg/store/gtm"
)

const StepKind core.Kind = "step"

type StepStatus uint8

const (
	Initializing StepStatus = iota
	Sending
	Fail
	Finish
)

type StepSpec struct {
	StageUUID    string                 `json:"stage_uuid" bson:"stage_uuid"`
	PipelineUUID string                 `json:"pipeline_uuid" bson:"pipeline_uuid"`
	ActionName   string                 `json:"action_name" bson:"action_name"`
	Data         map[string]interface{} `json:"data" bson:"data"`
	Trigger      bool                   `json:"trigger" bson:"trigger"`
	StepStatus   `json:"step_status" bson:"step_status"`
}

type Step struct {
	core.Metadata `json:"metadata"`
	Spec          StepSpec `json:"spec"`
}

// Pipeline impl Coder
func (*Step) Decode(op *gtm.Op) (core.IObject, error) {
	step := &Step{}
	if err := core.ObjectToResource(op.Data, step); err != nil {
		return nil, err
	}
	return step, nil
}

func (pl *Step) Clone() core.IObject {
	result := &Step{}
	core.Clone(pl, result)
	return result
}

func init() {
	store.AddResourceCoder(string(StepKind), &Step{})
}
