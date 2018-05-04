package egoscale

import (
	"testing"
)

func TestQueryAsyncJobResult(t *testing.T) {
	req := &QueryAsyncJobResult{}
	if req.name() != "queryAsyncJobResult" {
		t.Errorf("API call doesn't match")
	}
	_ = req.response().(*AsyncJobResult)
}

func TestListAsyncJobs(t *testing.T) {
	req := &ListAsyncJobs{}
	if req.name() != "listAsyncJobs" {
		t.Errorf("API call doesn't match")
	}
	_ = req.response().(*ListAsyncJobsResponse)
}
