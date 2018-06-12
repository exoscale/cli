package egoscale

import (
	"encoding/json"
	"fmt"
)

// QueryAsyncJobResult represents a query to fetch the status of async job
//
// CloudStack API: https://cloudstack.apache.org/api/apidocs-4.10/apis/queryAsyncJobResult.html
type QueryAsyncJobResult struct {
	JobID string `json:"jobid" doc:"the ID of the asychronous job"`
}

func (*QueryAsyncJobResult) name() string {
	return "queryAsyncJobResult"
}

func (*QueryAsyncJobResult) description() string {
	return "Retrieves the current status of asynchronous job."
}

func (*QueryAsyncJobResult) response() interface{} {
	return new(AsyncJobResult)
}

func (*ListAsyncJobs) name() string {
	return "listAsyncJobs"
}

func (*ListAsyncJobs) description() string {
	return "Lists all pending asynchronous jobs for the account."
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

func (a *AsyncJobResult) Error() error {
	r := new(ErrorResponse)
	if e := json.Unmarshal(*a.JobResult, r); e != nil {
		return e
	}
	return r
}
