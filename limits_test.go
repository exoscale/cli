package egoscale

import (
	"testing"
)

func TestListResourceLimits(t *testing.T) {
	req := &ListResourceLimits{}
	if req.name() != "listResourceLimits" {
		t.Errorf("API call doesn't match")
	}
	_ = req.response().(*ListResourceLimitsResponse)
}

func TestUpdateResourceLimit(t *testing.T) {
	req := &UpdateResourceLimit{}
	if req.name() != "updateResourceLimit" {
		t.Errorf("API call doesn't match")
	}
	_ = req.response().(*UpdateResourceLimitResponse)
}

func TestGetAPILimit(t *testing.T) {
	req := &GetAPILimit{}
	if req.name() != "getAPILimit" {
		t.Errorf("API call doesn't match")
	}
	_ = req.response().(*GetAPILimitResponse)
}
