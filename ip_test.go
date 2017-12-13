package egoscale

import (
	"log"
	"testing"
)

var async = AsyncInfo{0, 0}

func TestAssociateIpAddress(t *testing.T) {
	ts := newServer(200, `
{
	"associateipaddressresponse (m00t)": {
		"accountid": "74c1d64f-50b5-4cfa-b461-955b02a8ec99",
		"cmd": "org.apache.cloudstack.api.command.user.address.AssociateIPAddrCmd",
		"created": "2017-12-12T08:21:06+0100",
		"jobid": "f4a2b143-586b-42f7-9e64-f2f22c7bf95f",
		"jobprocstatus": 0,
		"jobresult": {
			"ipaddress": {
				"account": "yoan.blanc@exoscale.ch",
				"associated": "2017-12-12T08:21:06+0100",
				"associatednetworkid": "d48bfccc-c11f-438f-8177-9cf6a40dc4f8",
				"associatednetworkname": "defaultGuestNetwork",
				"domain": "yoan.blanc@exoscale.ch",
				"domainid": "2da0d0d3-e7b2-42ef-805d-eb2ea90ae7ef",
				"forvirtualnetwork": false,
				"id": "c89e3241-cd4e-4c1d-9ac8-eec6ffde93fb",
				"ipaddress": "159.100.251.223",
				"iselastic": true,
				"isportable": false,
				"issourcenat": false,
				"isstaticnat": false,
				"issystem": false,
				"networkid": "d48bfccc-c11f-438f-8177-9cf6a40dc4f8",
				"physicalnetworkid": "ecff7e96-d4f6-4af4-ac0f-dfaf95b39e0d",
				"state": "Associated",
				"tags": [],
				"zoneid": "381d0a95-ed4a-4ad9-b41c-b97073c1a433",
				"zonename": "ch-dk-2"
			}
		},
		"jobresultcode": 0,
		"jobresulttype": "object",
		"jobstatus": 1,
		"userid": "2c93a7ff-aea0-432d-bb3d-c367cb4dad8d"
	}
}
	`)
	defer ts.Close()

	cs := NewClient(ts.URL, "TOKEN", "SECRET")
	profile := IpAddressProfile{Zone: "fakeId"}
	ipAddress, err := cs.CreateIpAddress(profile, async)
	if err != nil {
		log.Fatal(err)
	}

	if ipAddress.IpAddress != "159.100.251.223" {
		t.Errorf("Expected the IpAddress to be created")
	}
}

func TestAssociateIpAddressBadZone(t *testing.T) {
	ts := newServer(400, `
{
	"associateipaddressresponse (bad zone)": {
		"cserrorcode": 4350,
		"errorcode": 431,
		"errortext": "bummer!",
		"uuidList": []
	}
}
	`)
	defer ts.Close()

	cs := NewClient(ts.URL, "TOKEN", "SECRET")
	profile := IpAddressProfile{Zone: "fakeId"}
	_, err := cs.CreateIpAddress(profile, async)
	if err == nil {
		t.Errorf("Expected an error to be returned, got an IpAddress")
	}
	if err.Error() != "exoscale API error 431 (internal code: 4350): bummer!" {
		t.Errorf("Expected the CloudStack API Error to be returned. Got: %s", err.Error())
	}
}

func TestDisassociateIpAddress(t *testing.T) {
	ts := newServer(200, `
{
	"disassociateipaddressresponse (w00t)": {
		"accountid": "74c1d64f-50b5-4cfa-b461-955b02a8ec99",
		"cmd": "org.apache.cloudstack.api.command.user.address.DisassociateIPAddrCmd",
		"created": "2017-12-12T09:46:56+0100",
		"jobid": "d54305ab-86c5-4460-a187-04b575b60e94",
		"jobprocstatus": 0,
		"jobresult": {
			"success": true
		},
		"jobresultcode": 0,
		"jobresulttype": "object",
		"jobstatus": 1,
		"userid": "2c93a7ff-aea0-432d-bb3d-c367cb4dad8d"
	}
}
`)
	defer ts.Close()

	cs := NewClient(ts.URL, "TOKEN", "SECRET")
	ipAddressId := "fakeId"
	err := cs.DestroyIpAddress(ipAddressId, async)
	if err != nil {
		log.Fatal(err)
	}
}

func TestDisassociateIpAddressBadId(t *testing.T) {
	ts := newServer(431, `
{
	"disassociateipaddressresponse": {
		"cserrorcode": 4350,
		"errorcode": 431,
		"errortext": "Unable to find ip address by id=42",
		"uuidList": []
	}
}
	`)
	defer ts.Close()

	cs := NewClient(ts.URL, "TOKEN", "SECRET")
	ipAddressId := "fakeId"
	err := cs.DestroyIpAddress(ipAddressId, async)
	if err == nil {
		t.Errorf("Expected an error, got nothing")
	}
}
