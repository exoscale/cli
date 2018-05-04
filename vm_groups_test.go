package egoscale

import (
	"testing"
)

func TestListInstanceGroups(t *testing.T) {
	req := &ListInstanceGroups{}
	if req.name() != "listInstanceGroups" {
		t.Errorf("API call doesn't match")
	}
	_ = req.response().(*ListInstanceGroupsResponse)
}

func TestCreateInstanceGroup(t *testing.T) {
	req := &CreateInstanceGroup{}
	if req.name() != "createInstanceGroup" {
		t.Errorf("API call doesn't match")
	}
	_ = req.response().(*InstanceGroup)
}

func TestUpdateInstanceGroup(t *testing.T) {
	req := &UpdateInstanceGroup{}
	if req.name() != "updateInstanceGroup" {
		t.Errorf("API call doesn't match")
	}
	_ = req.response().(*InstanceGroup)
}

func TestDeleteInstanceGroup(t *testing.T) {
	req := &DeleteInstanceGroup{}
	if req.name() != "deleteInstanceGroup" {
		t.Errorf("API call doesn't match")
	}
	_ = req.response().(*booleanResponse)
}
