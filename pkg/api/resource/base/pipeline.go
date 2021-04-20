package base

import (
	"github.com/yametech/verthandi/pkg/core"
	"github.com/yametech/verthandi/pkg/store"
	"github.com/yametech/verthandi/pkg/store/gtm"
)

type RequestPipeLine struct {
	Name  string         `json:"name"`
	Stage []RequestStage `json:"stage"`
}

type RequestStage struct {
	Steps []RequestStep `json:"steps"`
}

type RequestStep struct {
	ActionName string                 `json:"action_name" bson:"action_name"`
	Data       map[string]interface{} `json:"data" bson:"data"`
	FrontID    string                 `json:"front_id" bson:"front_id"`
	Trigger    bool                   `json:"trigger" bson:"trigger"`
}

type RespPipelineStatus uint8

const (
	Running RespPipelineStatus = iota
	Finished
)

type RespPipelineSpec struct {
	StagesData []RespStage `json:"stages_data"`
	Stages     []string    `json:"stages"`

	LastState          string `json:"last_state"`
	RespPipelineStatus `json:"pipeline_status"`
}

type RespPipeline struct {
	core.Metadata `json:"metadata"`
	Spec          RespPipelineSpec `json:"spec"`
}

const RespPipelineKind core.Kind = "resppipeline"

// Pipeline impl Coder
func (*RespPipeline) Decode(op *gtm.Op) (core.IObject, error) {
	action := &RespPipeline{}
	if err := core.ObjectToResource(op.Data, action); err != nil {
		return nil, err
	}
	return action, nil
}

func init() {
	store.AddResourceCoder(string(RespPipelineKind), &RespPipeline{})
}
