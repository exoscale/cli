package egoscale

import (
	"testing"
)

func TestListServiceOfferings(t *testing.T) {
	req := &ListServiceOfferings{}
	if req.name() != "listServiceOfferings" {
		t.Errorf("API call doesn't match")
	}
	_ = req.response().(*ListServiceOfferingsResponse)
}

func TestGetServiceOffering(t *testing.T) {
	ts := newServer(response{200, jsonContentType, `
{"listserviceofferingsresponse": {
	"count": 1,
	"serviceoffering": [
    {
      "id": "7c12e6df-6096-43e6-b9e4-3cb7b4e3f4c8",
      "cpunumber": 1,
      "cpuspeed": 2198,
      "created": "2013-01-25T14:21:15+0100",
      "displaytext": "Micro 512mb 1cpu",
      "domain": "",
      "domainid": "",
      "memory": 512,
      "name": "Micro",
      "storagetype": "local"
    }
  ]
}}`})

	defer ts.Close()

	cs := NewClient(ts.URL, "KEY", "SECRET")
	so := &ServiceOffering{
		Name: "Micro",
	}
	if err := cs.Get(so); err != nil {
		t.Error(err)
	}

	if so.ID != "7c12e6df-6096-43e6-b9e4-3cb7b4e3f4c8" {
		t.Errorf("ServiceOffering doesn't match expected id %q, got %v", "7c12e6df-6096-43e6-b9e4-3cb7b4e3f4c8", so.ID)
	}
}
