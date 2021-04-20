package base

import "github.com/yametech/verthandi/pkg/core"

type RespStageSpec struct {
	PipelineUUID string   `json:"pipeline_uuid"`
	LastState    string   `json:"last_state"`
	StepsData    []RespStep `json:"steps_data"`
	Steps        []string `json:"steps"`
	Done         bool     `json:"done"`
}

type RespStage struct {
	core.Metadata `json:"metadata"`
	Spec          RespStageSpec `json:"spec"`
}
