package egoscale

import (
	"testing"
)

func TestAddressess(t *testing.T) {
	var _ Taggable = (*IPAddress)(nil)
	var _ asyncCommand = (*AssociateIPAddress)(nil)
	var _ asyncCommand = (*DisassociateIPAddress)(nil)
	var _ syncCommand = (*ListPublicIPAddresses)(nil)
	var _ asyncCommand = (*UpdateIPAddress)(nil)
}

func TestIPAddress(t *testing.T) {
	instance := &IPAddress{}
	if instance.ResourceType() != "PublicIpAddress" {
		t.Errorf("ResourceType doesn't match")
	}
}

func TestAssociateIPAddress(t *testing.T) {
	req := &AssociateIPAddress{}
	if req.APIName() != "associateIpAddress" {
		t.Errorf("API call doesn't match")
	}
	_ = req.asyncResponse().(*AssociateIPAddressResponse)
}

func TestDisassociateIPAddress(t *testing.T) {
	req := &DisassociateIPAddress{}
	if req.APIName() != "disassociateIpAddress" {
		t.Errorf("API call doesn't match")
	}
	_ = req.asyncResponse().(*booleanAsyncResponse)
}

func TestListPublicIPAddresses(t *testing.T) {
	req := &ListPublicIPAddresses{}
	if req.APIName() != "listPublicIpAddresses" {
		t.Errorf("API call doesn't match")
	}
	_ = req.response().(*ListPublicIPAddressesResponse)
}

func TestUpdateIPAddress(t *testing.T) {
	req := &UpdateIPAddress{}
	if req.APIName() != "updateIpAddress" {
		t.Errorf("API call doesn't match")
	}
	_ = req.asyncResponse().(*UpdateIPAddressResponse)
}
