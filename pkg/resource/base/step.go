package base

type StepType uint8

const (
	CI StepType = iota
	CD
)

type Step struct {
	Type    StepType               `json:"type" bson:"type"`
	Data    map[string]interface{} `json:"data" bson:"data"`
	Trigger bool                   `json:"trigger" bson:"trigger"`
	Done    bool                   `json:"done" bson:"done"`
}

type Stage struct {
	Steps []Step `json:"steps"`
	Done  bool   `json:"done" bson:"done"`
}
