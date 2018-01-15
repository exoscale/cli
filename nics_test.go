package egoscale

import (
	"testing"
)

func TestNics(t *testing.T) {
	var _ AsyncCommand = (*AddIPToNic)(nil)
	var _ AsyncCommand = (*RemoveIPFromNic)(nil)
	var _ Command = (*ListNics)(nil)
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
