package egoscale

import (
	"net/url"
	"testing"
)

func TestAffinityGroups(t *testing.T) {
	var _ asyncCommand = (*CreateAffinityGroup)(nil)
	var _ asyncCommand = (*DeleteAffinityGroup)(nil)
	var _ syncCommand = (*ListAffinityGroupTypes)(nil)
	var _ syncCommand = (*ListAffinityGroups)(nil)
	var _ asyncCommand = (*UpdateVMAffinityGroup)(nil)
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

func TestUpdateVMOnBeforeSend(t *testing.T) {
	req := &UpdateVMAffinityGroup{}
	params := url.Values{}

	if err := req.onBeforeSend(&params); err != nil {
		t.Error(err)
	}

	if _, ok := params["affinitygroupids"]; !ok {
		t.Errorf("affinitygroupids should have been set")
	}
}
