package egoscale

import (
	"testing"
)

func TestNics(t *testing.T) {
	var _ asyncCommand = (*AddIPToNic)(nil)
	var _ asyncCommand = (*RemoveIPFromNic)(nil)
	var _ syncCommand = (*ListNics)(nil)
	var _ asyncCommand = (*ActivateIP6)(nil)
}

func TestAddIPToNic(t *testing.T) {
	req := &AddIPToNic{}
	if req.APIName() != "addIpToNic" {
		t.Errorf("API call doesn't match")
	}
	_ = req.asyncResponse().(*AddIPToNicResponse)
}

func TestRemoveIPFromNic(t *testing.T) {
	req := &RemoveIPFromNic{}
	if req.APIName() != "removeIpFromNic" {
		t.Errorf("API call doesn't match")
	}
	_ = req.asyncResponse().(*booleanAsyncResponse)
}

func TestListNics(t *testing.T) {
	req := &ListNics{}
	if req.APIName() != "listNics" {
		t.Errorf("API call doesn't match")
	}
	_ = req.response().(*ListNicsResponse)
}

func TestActivateIP6(t *testing.T) {
	req := &ActivateIP6{}
	if req.APIName() != "activateIp6" {
		t.Errorf("API call doesn't match")
	}
	_ = req.asyncResponse().(*ActivateIP6Response)
}
