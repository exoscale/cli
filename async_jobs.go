package egoscale

import (
	"encoding/json"
)

// QueryAsyncJobResultRequest represents a query to fetch the status of async job
type QueryAsyncJobResultRequest struct {
	JobId string `json:"jobid"`
}

// Command returns the CloudStack API command
func (req *QueryAsyncJobResultRequest) Command() string {
	return "queryAsyncJobResult"
}

// QueryASyncJobResultResponse represents the current status of an asynchronous job
type QueryAsyncJobResultResponse struct {
	AccountId       string           `json:"accountid"`
	Cmd             string           `json:"cmd"`
	Created         string           `json:"created"`
	JobInstanceId   string           `json:"jobinstanceid"`
	JobInstanceType string           `json:"jobinstancetype"`
	JobProcStatus   int              `json:"jobprocstatus"`
	JobResult       *json.RawMessage `json:"jobresult"`
	JobResultCode   int              `json:"jobresultcode"`
	JobResultType   string           `json:"jobresulttype"`
	JobStatus       JobStatusType    `json:"jobstatus"`
	UserId          string           `json:"userid"`
	JobId           string           `json:"jobid"`
}
