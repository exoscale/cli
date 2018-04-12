package egoscale

import (
	"net"
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

func TestGetIPAddress(t *testing.T) {
	ts := newServer(response{200, `
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

func TestGetIPAddressInvalid(t *testing.T) {
	ts := newServer(response{400, ``})

	defer ts.Close()

	cs := NewClient(ts.URL, "KEY", "SECRET")
	eip := &IPAddress{}
	if err := cs.Get(eip); err == nil {
		t.Errorf("An error was expected")
	}
}

func TestGetIPAddressMissing(t *testing.T) {
	ts := newServer(response{200, `
{"listpublicipaddressesresponse":{}}
`})

	defer ts.Close()

	cs := NewClient(ts.URL, "KEY", "SECRET")
	eip := &IPAddress{
		ID: "missing",
	}
	if err := cs.Get(eip); err == nil {
		t.Errorf("An error was expected")
	}
}

func TestGetIPAddressMultiple(t *testing.T) {
	ts := newServer(response{200, `
{"listpublicipaddressesresponse":{
	"count": 2,
	"publicipaddress": [
		{"id": "1"},
		{"id": "2"}
	]
}}
`})

	defer ts.Close()

	cs := NewClient(ts.URL, "KEY", "SECRET")
	eip := &IPAddress{
		ID: "multiple",
	}
	if err := cs.Get(eip); err == nil {
		t.Errorf("An error was expected")
	}
}

func TestGetIPAddressError(t *testing.T) {
	ts := newServer(response{400, `
{"listpublicipaddressesresponse": {
	"errorcode": 431,
	"errortext": "foo"
}}`})

	defer ts.Close()

	cs := NewClient(ts.URL, "KEY", "SECRET")
	eip := &IPAddress{
		ID: "err",
	}
	if err := cs.Get(eip); err == nil {
		t.Errorf("An error was expected")
	}
}

func TestDeleteIPAddress(t *testing.T) {
	ts := newServer(response{200, `
{"queryasyncjobresultresponse": {
	"jobid": "b1ac7d06-3320-4388-b234-43420bcb236c",
	"jobprocstatus": 0,
	"jobresult": {
		"success": true
	},
	"jobstatus": 2
}}`})

	defer ts.Close()

	cs := NewClient(ts.URL, "KEY", "SECRET")
	eip := &IPAddress{
		ID: "err",
	}
	if err := cs.Delete(eip); err != nil {
		t.Error(err)
	}
}

func TestDeleteIPAddressInvalid(t *testing.T) {
	ts := newServer(response{400, ``})

	defer ts.Close()

	cs := NewClient(ts.URL, "KEY", "SECRET")
	eip := &IPAddress{}
	if err := cs.Delete(eip); err == nil {
		t.Errorf("An error was expected")
	}
}

func TestDeleteIPAddressError(t *testing.T) {
	ts := newServer(response{400, `
{"queryasyncjobresultresponse": {
	"jobid": "b1ac7d06-3320-4388-b234-43420bcb236c",
	"jobprocstatus": 0,
	"jobresult": {
		"errorcode": 431,
		"errortext": "Only elastic IP can be released explicitly."
	},
	"jobstatus": 2
}}`})

	defer ts.Close()

	cs := NewClient(ts.URL, "KEY", "SECRET")
	eip := &IPAddress{
		ID: "err",
	}
	if err := cs.Delete(eip); err == nil {
		t.Errorf("An error was expected")
	}
}

func TestListIPAddress(t *testing.T) {
	ts := newServer(response{200, `
		{"listPublicIpAddresses":{
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
