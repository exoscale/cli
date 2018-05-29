package egoscale

import (
	"testing"
)

func TestListHosts(t *testing.T) {
	req := &ListHosts{}
	if req.name() != "listHosts" {
		t.Errorf("API call doesn't match")
	}
	_ = req.response().(*ListHostsResponse)
}
