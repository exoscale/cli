package egoscale

import (
	"testing"
)

func TestAffinityGroups(t *testing.T) {
	var _ AsyncCommand = (*CreateAffinityGroup)(nil)
	var _ AsyncCommand = (*DeleteAffinityGroup)(nil)
	var _ Command = (*ListAffinityGroupTypes)(nil)
	var _ Command = (*ListAffinityGroups)(nil)
	var _ AsyncCommand = (*UpdateVMAffinityGroup)(nil)
}

func TestCreateAffinityGroup(t *testing.T) {
	req := &CreateAffinityGroup{}
	if req.APIName() != "createAffinityGroup" {
		t.Errorf("API call doesn't match")
	}
	_ = req.asyncResponse().(*CreateAffinityGroupResponse)
}

func TestDeleteAffinityGroup(t *testing.T) {
	req := &DeleteAffinityGroup{}
	if req.APIName() != "deleteAffinityGroup" {
		t.Errorf("API call doesn't match")
	}
	_ = req.asyncResponse().(*booleanAsyncResponse)
}

func TestListAffinityGroups(t *testing.T) {
	req := &ListAffinityGroups{}
	if req.APIName() != "listAffinityGroups" {
		t.Errorf("API call doesn't match")
	}
	_ = req.response().(*ListAffinityGroupsResponse)
}

func TestListAffinityGroupTypes(t *testing.T) {
	req := &ListAffinityGroupTypes{}
	if req.APIName() != "listAffinityGroupTypes" {
		t.Errorf("API call doesn't match")
	}
	_ = req.response().(*ListAffinityGroupTypesResponse)
}

func TestUpdateVMAffinityGroup(t *testing.T) {
	req := &UpdateVMAffinityGroup{}
	if req.APIName() != "updateVMAffinityGroup" {
		t.Errorf("API call doesn't match")
	}
	_ = req.asyncResponse().(*UpdateVMAffinityGroupResponse)
}
