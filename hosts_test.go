package egoscale

import (
	"testing"
)

func TestListHosts(t *testing.T) {
	req := &ListHosts{}
	_ = req.response().(*ListHostsResponse)
}
