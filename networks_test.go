package egoscale

import (
	"testing"
)

func TestNetworks(t *testing.T) {
	var _ Taggable = (*Network)(nil)
	var _ syncCommand = (*CreateNetwork)(nil)
	var _ asyncCommand = (*DeleteNetwork)(nil)
	var _ syncCommand = (*ListNetworks)(nil)
	var _ asyncCommand = (*RestartNetwork)(nil)
	var _ asyncCommand = (*UpdateNetwork)(nil)
}

func TestNetwork(t *testing.T) {
	instance := &Network{}
	if instance.ResourceType() != "Network" {
		t.Errorf("ResourceType doesn't match")
	}
}

func TestListNetworks(t *testing.T) {
	req := &ListNetworks{}
	if req.APIName() != "listNetworks" {
		t.Errorf("API call doesn't match")
	}
	_ = req.response().(*ListNetworksResponse)
}

func TestCreateNetwork(t *testing.T) {
	req := &CreateNetwork{}
	if req.APIName() != "createNetwork" {
		t.Errorf("API call doesn't match")
	}
	_ = req.response().(*CreateNetworkResponse)
}

func TestRestartNetwork(t *testing.T) {
	req := &RestartNetwork{}
	if req.APIName() != "restartNetwork" {
		t.Errorf("API call doesn't match")
	}
	_ = req.asyncResponse().(*RestartNetworkResponse)
}

func TestUpdateNetwork(t *testing.T) {
	req := &UpdateNetwork{}
	if req.APIName() != "updateNetwork" {
		t.Errorf("API call doesn't match")
	}
	_ = req.asyncResponse().(*UpdateNetworkResponse)
}

func TestDeleteNetwork(t *testing.T) {
	req := &DeleteNetwork{}
	if req.APIName() != "deleteNetwork" {
		t.Errorf("API call doesn't match")
	}
	_ = req.asyncResponse().(*booleanAsyncResponse)
}
