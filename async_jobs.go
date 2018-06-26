package egoscale

import (
	"encoding/json"
	"fmt"
)

// AsyncJobResult represents an asynchronous job result
type AsyncJobResult struct {
	AccountID       string           `json:"accountid"`
	Cmd             string           `json:"cmd"`
	Created         string           `json:"created"`
	JobInstanceID   string           `json:"jobinstanceid,omitempty"`
	JobInstanceType string           `json:"jobinstancetype,omitempty"`
	JobProcStatus   int              `json:"jobprocstatus"`
	JobResult       *json.RawMessage `json:"jobresult"`
	JobResultCode   int              `json:"jobresultcode"`
	JobResultType   string           `json:"jobresulttype"`
	JobStatus       JobStatusType    `json:"jobstatus"`
	UserID          string           `json:"userid"`
	JobID           string           `json:"jobid"`
}

func (a *AsyncJobResult) Error() error {
	r := new(ErrorResponse)
	if e := json.Unmarshal(*a.JobResult, r); e != nil {
		return e
	}
	return r
}

// QueryAsyncJobResult represents a query to fetch the status of async job
type QueryAsyncJobResult struct {
	JobID string `json:"jobid" doc:"the ID of the asychronous job"`
	_     bool   `name:"queryAsyncJobResult" description:"Retrieves the current status of asynchronous job."`
}

func (*QueryAsyncJobResult) response() interface{} {
	return new(AsyncJobResult)
}

// ListAsyncJobs list the asynchronous jobs
type ListAsyncJobs struct {
	Account     string `json:"account,omitempty" doc:"list resources by account. Must be used with the domainId parameter."`
	DomainID    string `json:"domainid,omitempty" doc:"list only resources belonging to the domain specified"`
	IsRecursive *bool  `json:"isrecursive,omitempty" doc:"defaults to false, but if true, lists all resources from the parent specified by the domainId till leaves."`
	Keyword     string `json:"keyword,omitempty" doc:"List by keyword"`
	ListAll     *bool  `json:"listall,omitempty" doc:"If set to false, list only resources belonging to the command's caller; if set to true - list resources that the caller is authorized to see. Default value is false"`
	Page        int    `json:"page,omitempty"`
	PageSize    int    `json:"pagesize,omitempty"`
	StartDate   string `json:"startdate,omitempty" doc:"the start date of the async job"`
	_           bool   `name:"listAsyncJobs" description:"Lists all pending asynchronous jobs for the account."`
}

// ListAsyncJobsResponse represents a list of job results
type ListAsyncJobsResponse struct {
	Count     int              `json:"count"`
	AsyncJobs []AsyncJobResult `json:"asyncjobs"`
}

func (*ListAsyncJobs) response() interface{} {
	return new(ListAsyncJobsResponse)
}

//Response return response of AsyncJobResult from a given type
func (a *AsyncJobResult) Response(i interface{}) error {
	if a.JobStatus == Failure {
		return a.Error()
	}

	var err error
	if a.JobStatus == Success {
		m := map[string]json.RawMessage{}
		err = json.Unmarshal(*(a.JobResult), &m)

		if err == nil {
			if len(m) > 1 {
				err = json.Unmarshal(*(a.JobResult), i)
			} else if len(m) == 1 {
				for k := range m {
					if k == "success" {
						err = json.Unmarshal(*(a.JobResult), i)
					}
					if e := json.Unmarshal(m[k], i); e != nil {
						return e
					}
				}
			} else {
				return fmt.Errorf("empty response")
			}
		}
	}

	return err
}
