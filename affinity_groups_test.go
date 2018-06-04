package egoscale

import (
	"net/url"
	"testing"
)

func TestCreateAffinityGroup(t *testing.T) {
	req := &CreateAffinityGroup{}
	if req.name() != "createAffinityGroup" {
		t.Errorf("API call doesn't match")
	}
	_ = req.asyncResponse().(*AffinityGroup)
}

func TestDeleteAffinityGroup(t *testing.T) {
	req := &DeleteAffinityGroup{}
	if req.name() != "deleteAffinityGroup" {
		t.Errorf("API call doesn't match")
	}
	_ = req.asyncResponse().(*booleanResponse)
}

func TestListAffinityGroups(t *testing.T) {
	req := &ListAffinityGroups{}
	if req.name() != "listAffinityGroups" {
		t.Errorf("API call doesn't match")
	}
	_ = req.response().(*ListAffinityGroupsResponse)
}

func TestListAffinityGroupTypes(t *testing.T) {
	req := &ListAffinityGroupTypes{}
	if req.name() != "listAffinityGroupTypes" {
		t.Errorf("API call doesn't match")
	}
	_ = req.response().(*ListAffinityGroupTypesResponse)
}

func TestUpdateVMAffinityGroup(t *testing.T) {
	req := &UpdateVMAffinityGroup{}
	if req.name() != "updateVMAffinityGroup" {
		t.Errorf("API call doesn't match")
	}
	_ = req.asyncResponse().(*VirtualMachine)
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

func TestGetAffinityGroup(t *testing.T) {
	ts := newServer(response{200, jsonContentType, `
{"listaffinitygroupsresponse": {
	"affinitygroup": [
		{
			"account": "yoan.blanc@exoscale.ch",
			"description": "default anti-affinity group",
			"domain": "yoan.blanc@exoscale.ch",
			"domainid": "2da0d0d3-e7b2-42ef-805d-eb2ea90ae7ef",
			"id": "6d7bc27c-6c8d-4b6a-ae97-83f73df18667",
			"name": "default",
			"type": "host anti-affinity"
		}
	],
	"count": 1
}}`})
	defer ts.Close()

	cs := NewClient(ts.URL, "KEY", "SECRET")
	sg := &AffinityGroup{
		ID: "6d7bc27c-6c8d-4b6a-ae97-83f73df18667",
	}
	if err := cs.Get(sg); err != nil {
		t.Error(err)
	}

	if sg.Account != "yoan.blanc@exoscale.ch" {
		t.Errorf("Account doesn't match, got %v", sg.Account)
	}
}
