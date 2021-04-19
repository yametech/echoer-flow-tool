package base

type FSMResp struct {
	Metadata struct {
		Name    string      `json:"name"`
		Kind    string      `json:"kind"`
		Version int         `json:"version"`
		UUID    string      `json:"uuid"`
		Labels  interface{} `json:"labels"`
	} `json:"metadata"`
	Spec struct {
		Steps []struct {
			Metadata struct {
				Name    string      `json:"name"`
				Kind    string      `json:"kind"`
				Version int         `json:"version"`
				UUID    string      `json:"uuid"`
				Labels  interface{} `json:"labels"`
			} `json:"metadata"`
			Spec struct {
				FlowID      string `json:"flow_id"`
				FlowRunUUID string `json:"flow_run_uuid"`
				ActionRun   struct {
					ActionName   string `json:"action_name"`
					ActionParams struct {
						Branch      string `json:"branch"`
						CodeType    string `json:"codeType"`
						CommitID    string `json:"commitId"`
						GitURL      string `json:"gitUrl"`
						Output      string `json:"output"`
						ProjectFile string `json:"projectFile"`
						ProjectPath string `json:"projectPath"`
						RetryCount  int    `json:"retryCount"`
						ServiceName string `json:"serviceName"`
					} `json:"action_params"`
					ReturnStateMap struct {
						FAIL    string `json:"FAIL"`
						SUCCESS string `json:"SUCCESS"`
					} `json:"return_state_map"`
					Done bool `json:"done"`
				} `json:"action_run"`
				Response struct {
					State string `json:"state"`
				} `json:"response"`
				Data            string      `json:"data"`
				RetryCount      int         `json:"retry_count"`
				GlobalVariables interface{} `json:"global_variables"`
			} `json:"spec"`
		} `json:"steps"`
		HistoryStates  []string    `json:"history_states"`
		LastState      string      `json:"last_state"`
		CurrentState   string      `json:"current_state"`
		LastEvent      string      `json:"last_event"`
		LastErr        string      `json:"last_err"`
		GlobalVariable interface{} `json:"global_variable"`
	} `json:"spec"`
}
