package base

import (
	"github.com/yametech/verthandi/pkg/core"
	"github.com/yametech/verthandi/pkg/store"
	"github.com/yametech/verthandi/pkg/store/gtm"
)

const RespStepKind core.Kind = "respstep"

type RespStepStatus uint8

const (
	Initializing RespStepStatus = iota
	Sending
	Fail
	Finish
)

type RespStepSpec struct {
	StageUUID      string                 `json:"stage_uuid" bson:"stage_uuid"`
	PipelineUUID   string                 `json:"pipeline_uuid" bson:"pipeline_uuid"`
	ActionName     string                 `json:"action_name" bson:"action_name"`
	Data           map[string]interface{} `json:"data" bson:"data"`
	Trigger        bool                   `json:"trigger" bson:"trigger"`
	RespStepStatus `json:"step_status" bson:"step_status"`
}

type RespStep struct {
	core.Metadata `json:"metadata"`
	Spec          RespStepSpec `json:"spec"`
}

// Pipeline impl Coder
func (*RespStep) Decode(op *gtm.Op) (core.IObject, error) {
	step := &RespStep{}
	if err := core.ObjectToResource(op.Data, step); err != nil {
		return nil, err
	}
	return step, nil
}

func init() {
	store.AddResourceCoder(string(RespStepKind), &RespStep{})
}
