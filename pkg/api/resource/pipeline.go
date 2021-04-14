package resource

type RequestPipeLine struct {
	Name string                     `json:"name"`
	Step [][]map[string]interface{} `json:"step"`
}
