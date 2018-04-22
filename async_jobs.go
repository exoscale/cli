package egoscale

// QueryAsyncJobResult represents a query to fetch the status of async job
//
// CloudStack API: https://cloudstack.apache.org/api/apidocs-4.10/apis/queryAsyncJobResult.html
type QueryAsyncJobResult struct {
	JobID string `json:"jobid" doc:"the ID of the asychronous job"`
}

// name returns the CloudStack API command name
func (*QueryAsyncJobResult) name() string {
	return "queryAsyncJobResult"
}

func (*QueryAsyncJobResult) response() interface{} {
	return new(QueryAsyncJobResultResponse)
}

// name returns the CloudStack API command name
func (*ListAsyncJobs) name() string {
	return "listAsyncJobs"
}

func (*ListAsyncJobs) response() interface{} {
	return new(ListAsyncJobsResponse)
}
