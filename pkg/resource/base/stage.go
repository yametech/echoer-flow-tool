package base

import "github.com/yametech/verthandi/pkg/core"

type StageSpec struct {
	PipelineUUID string   `json:"pipeline_uuid" bson:"pipeline_uuid"`
	LastState    string   `json:"last_state" bson:"last_state"`
	Steps        []string `json:"steps" bson:"steps"`
	Done         bool     `json:"done" bson:"done"`
}

type Stage struct {
	core.Metadata `json:"metadata"`
	Spec          StageSpec `json:"spec"`
}

func (pl *Stage) Clone() core.IObject {
	result := &Stage{}
	core.Clone(pl, result)
	return result
}
