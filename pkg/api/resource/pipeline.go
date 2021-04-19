package resource

type RequestPipeLine struct {
	Name  string         `json:"name"`
	Stage []RequestStage `json:"stage"`
}

type RequestStep struct {
	Type    uint8                  `json:"type" bson:"type"`
	Data    map[string]interface{} `json:"data" bson:"data"`
	Trigger bool                   `json:"trigger" bson:"trigger"`
}

type RequestStage struct {
	Steps []RequestStep `json:"steps"`
}

