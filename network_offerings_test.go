package egoscale

import (
	"testing"
)

func TestNetworkOfferings(t *testing.T) {
	var _ Command = (*ListNetworkOfferings)(nil)
	var _ Command = (*UpdateNetworkOffering)(nil)
}

func TestListNetworkOfferings(t *testing.T) {
	req := &ListNetworkOfferings{}
	if req.name() != "listNetworkOfferings" {
		t.Errorf("API call doesn't match")
	}
	_ = req.response().(*ListNetworkOfferingsResponse)
}

func TestUpdateNetworkOffering(t *testing.T) {
	req := &UpdateNetworkOffering{}
	if req.name() != "updateNetworkOffering" {
		t.Errorf("API call doesn't match")
	}
	_ = req.response().(*UpdateNetworkOfferingResponse)
}
