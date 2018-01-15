package egoscale

import (
	"testing"
)

func TestResourceLimits(t *testing.T) {
	var _ Command = (*ListResourceLimits)(nil)
}

func TestListResourceLimits(t *testing.T) {
	req := &ListResourceLimits{}
	if req.APIName() != "listResourceLimits" {
		t.Errorf("API call doesn't match")
	}
	_ = req.response().(*ListResourceLimitsResponse)
}
