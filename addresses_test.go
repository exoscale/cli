package egoscale

import (
	"net"
	"testing"
)

func TestIPAddress(t *testing.T) {
	instance := &IPAddress{}
	if instance.ResourceType() != "PublicIpAddress" {
		t.Errorf("ResourceType doesn't match")
	}
}

func TestAssociateIPAddress(t *testing.T) {
	req := &AssociateIPAddress{}
	_ = req.response().(*AsyncJobResult)
	_ = req.asyncResponse().(*IPAddress)
}

func TestDisassociateIPAddress(t *testing.T) {
	req := &DisassociateIPAddress{}
	_ = req.response().(*AsyncJobResult)
	_ = req.asyncResponse().(*booleanResponse)
}

func TestListPublicIPAddresses(t *testing.T) {
	req := &ListPublicIPAddresses{}
	_ = req.response().(*ListPublicIPAddressesResponse)
}

func TestUpdateIPAddress(t *testing.T) {
	req := &UpdateIPAddress{}
	_ = req.response().(*AsyncJobResult)
	_ = req.asyncResponse().(*IPAddress)
}

func TestGetIPAddress(t *testing.T) {
	ts := newServer(response{200, jsonContentType, `
{"listpublicipaddressesresponse": {
	"count": 1,
	"publicipaddress": [
		{
			"account": "yoan.blanc@exoscale.ch",
			"allocated": "2017-12-18T08:16:56+0100",
			"domain": "yoan.blanc@exoscale.ch",
			"domainid": "17b3bc0c-8aed-4920-be3a-afdd6daf4314",
			"forvirtualnetwork": false,
			"id": "adc31802-9413-47ea-a252-266f7eadaf00",
			"ipaddress": "109.100.242.45",
			"iselastic": true,
			"isportable": false,
			"issourcenat": false,
			"isstaticnat": false,
			"issystem": false,
			"networkid": "00304a04-e7ea-4e77-a786-18bc64347bf7",
			"physicalnetworkid": "01f747f5-b445-487f-b2d7-81a5a512989e",
			"state": "Allocated",
			"tags": [],
			"zoneid": "1128bd56-b4e9-4ac6-a7b9-c715b187ce11",
			"zonename": "ch-gva-2"
		}
	]
}}`})

	defer ts.Close()

	cs := NewClient(ts.URL, "KEY", "SECRET")
	eip := &IPAddress{
		IPAddress: net.ParseIP("109.100.242.45"),
		IsElastic: true,
	}
	if err := cs.Get(eip); err != nil {
		t.Error(err)
	}

	if eip.Account != "yoan.blanc@exoscale.ch" {
		t.Errorf("Account doesn't match, got %v", eip.Account)
	}
}

func TestListIPAddress(t *testing.T) {
	ts := newServer(response{200, jsonContentType, `
		{"listpublicipaddressesresponse":{
			"count": 1,
			"publicipaddress": [
			  {
				"account": "exoscale-1",
				"allocated": "2018-04-09T10:08:08+0200",
				"domain": "exoscale-1",
				"domainid": "5b2f621e-3eb6-4a14-a315-d4d7d62f28ff",
				"forvirtualnetwork": false,
				"id": "b27adf5e-c189-4b75-b72a-4e9b25232b63",
				"ipaddress": "159.100.246.105",
				"iselastic": false,
				"isportable": false,
				"issourcenat": false,
				"isstaticnat": false,
				"issystem": false,
				"networkid": "e38bf934-1277-4e82-8cef-f3a862d9ec57",
				"physicalnetworkid": "0e2989e9-4cac-4faa-a6c6-8e43f936a76a",
				"state": "Allocated",
				"tags": [],
				"zoneid": "91e5e9e4-c9ed-4b76-bee4-427004b3baf9",
				"zonename": "ch-dk-2"
			  }
			]
		  }}
		`})

	defer ts.Close()

	cs := NewClient(ts.URL, "KEY", "SECRET")
	eip := &IPAddress{
		ZoneID: "91e5e9e4-c9ed-4b76-bee4-427004b3baf9",
	}
	ips, err := cs.List(eip)
	if err != nil {
		t.Errorf("%v", err)
	}

	if len(ips) != 1 {
		t.Errorf("IP not found")
	}

	ip := ips[0].(*IPAddress)

	if ip.ZoneID != "91e5e9e4-c9ed-4b76-bee4-427004b3baf9" {
		t.Errorf("Wrong ip address")
	}
}

func TestListIPAddressPaginate(t *testing.T) {
	ts := newServer(response{200, jsonContentType, `
		{"listpublicipaddressesresponse":{
			"count": 1,
			"publicipaddress": [
			  {
				"account": "exoscale-1",
				"allocated": "2018-04-09T10:08:08+0200",
				"domain": "exoscale-1",
				"domainid": "5b2f621e-3eb6-4a14-a315-d4d7d62f28ff",
				"forvirtualnetwork": false,
				"id": "b27adf5e-c189-4b75-b72a-4e9b25232b63",
				"ipaddress": "159.100.246.105",
				"iselastic": false,
				"isportable": false,
				"issourcenat": false,
				"isstaticnat": false,
				"issystem": false,
				"networkid": "e38bf934-1277-4e82-8cef-f3a862d9ec57",
				"physicalnetworkid": "0e2989e9-4cac-4faa-a6c6-8e43f936a76a",
				"state": "Allocated",
				"tags": [],
				"zoneid": "91e5e9e4-c9ed-4b76-bee4-427004b3baf9",
				"zonename": "ch-dk-2"
			  },
			  {
				"account": "exoscale-1",
				"allocated": "2018-04-09T10:08:08+0200",
				"domain": "exoscale-1",
				"domainid": "5b2f621e-3eb6-4a14-a315-d4d7d62f28ff",
				"forvirtualnetwork": false,
				"id": "testestetet",
				"ipaddress": "159.100.246.105",
				"iselastic": false,
				"isportable": false,
				"issourcenat": false,
				"isstaticnat": false,
				"issystem": false,
				"networkid": "e38bf934-1277-4e82-8cef-f3a862d9ec57",
				"physicalnetworkid": "0e2989e9-4cac-4faa-a6c6-8e43f936a76a",
				"state": "Allocated",
				"tags": [],
				"zoneid": "91e5e9e4-c9ed-4b76-bee4-427004b3baf9",
				"zonename": "ch-dk-2"
			  }
			]
		  }}`})

	defer ts.Close()

	cs := NewClient(ts.URL, "KEY", "SECRET")

	eip := &IPAddress{}

	req, err := eip.ListRequest()
	if err != nil {
		t.Errorf("%v", err)
	}

	cs.Paginate(req, func(i interface{}, err error) bool {
		if i.(*IPAddress).ID != "b27adf5e-c189-4b75-b72a-4e9b25232b63" {
			t.Errorf("Expected id 'b27adf5e-c189-4b75-b72a-4e9b25232b63', got %v", i.(*IPAddress).ID)
		}
		return false
	})

}

func TestListIPAddressFailure(t *testing.T) {
	ts := newServer(response{200, jsonContentType, `
	{"listpublicipaddressesresponse": {
		"count": 1,
		"publicipaddress": {}
	}`})

	defer ts.Close()

	cs := NewClient(ts.URL, "KEY", "SECRET")

	eip := &IPAddress{}

	_, err := cs.List(eip)
	if err == nil {
		t.Errorf("Expected an error, got %v", err)
	}

}
