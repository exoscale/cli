package egoscale

import (
	"testing"
	"time"
)

func TestListZones(t *testing.T) {
	req := &ListZones{}
	if req.APIName() != "listZones" {
		t.Errorf("API call doesn't match")
	}
	_ = req.response().(*ListZonesResponse)
}

func TestListZone(t *testing.T) {
	ts := newServer(response{200, `
{"listzonesresponse": {
	"count": 4,
	"zone": [
		{
			"allocationstate": "Enabled",
			"dhcpprovider": "VirtualRouter",
			"id": "1747ef5e-5451-41fd-9f1a-58913bae9702",
			"localstorageenabled": true,
			"name": "ch-gva-2",
			"networktype": "Basic",
			"securitygroupsenabled": true,
			"tags": [],
			"zonetoken": "f9a2983b-42e5-3b12-ae74-0b1f54cd6204"
		},
		{
			"allocationstate": "Enabled",
			"dhcpprovider": "VirtualRouter",
			"id": "381d0a95-ed4a-4ad9-b41c-b97073c1a433",
			"localstorageenabled": true,
			"name": "ch-dk-2",
			"networktype": "Basic",
			"securitygroupsenabled": true,
			"tags": [],
			"zonetoken": "23a24359-121a-38af-a938-e225c97c397b"
		},
		{
			"allocationstate": "Enabled",
			"dhcpprovider": "VirtualRouter",
			"id": "b0fcd72f-47ad-4779-a64f-fe4de007ec72",
			"localstorageenabled": true,
			"name": "at-vie-1",
			"networktype": "Basic",
			"securitygroupsenabled": true,
			"tags": [],
			"zonetoken": "a2a8345d-7daa-3316-8d90-5b8e49706764"
		},
		{
			"allocationstate": "Enabled",
			"dhcpprovider": "VirtualRouter",
			"id": "de88c980-78f6-467c-a431-71bcc88e437f",
			"localstorageenabled": true,
			"name": "de-fra-1",
			"networktype": "Basic",
			"securitygroupsenabled": true,
			"tags": [],
			"zonetoken": "c4bdb9f2-c28d-36a3-bbc5-f91fc69527e6"
		}
	]
}}`})
	defer ts.Close()

	cs := NewClient(ts.URL, "KEY", "SECRET")

	zone := new(Zone)
	zones, err := cs.List(zone)
	if err != nil {
		t.Error(err)
	}

	if len(zones) != 4 {
		t.Errorf("Four zones were expected, got %d", len(zones))
	}

	if zones[2].(Zone).Name != "at-vie-1" {
		t.Errorf("Expected VIE1 to be third, got %#v", zones[2])
	}
}

func TestListZoneTwoPages(t *testing.T) {
	ts := newServer(response{200, `
{"listzonesresponse": {
	"count": 4,
	"zone": [
		{
			"allocationstate": "Enabled",
			"dhcpprovider": "VirtualRouter",
			"id": "1747ef5e-5451-41fd-9f1a-58913bae9702",
			"localstorageenabled": true,
			"name": "ch-gva-2",
			"networktype": "Basic",
			"securitygroupsenabled": true,
			"tags": [],
			"zonetoken": "f9a2983b-42e5-3b12-ae74-0b1f54cd6204"
		},
		{
			"allocationstate": "Enabled",
			"dhcpprovider": "VirtualRouter",
			"id": "381d0a95-ed4a-4ad9-b41c-b97073c1a433",
			"localstorageenabled": true,
			"name": "ch-dk-2",
			"networktype": "Basic",
			"securitygroupsenabled": true,
			"tags": [],
			"zonetoken": "23a24359-121a-38af-a938-e225c97c397b"
		}
	]
}}`}, response{200, `
{"listzonesresponse": {
	"count": 4,
	"zone": [
		{
			"allocationstate": "Enabled",
			"dhcpprovider": "VirtualRouter",
			"id": "b0fcd72f-47ad-4779-a64f-fe4de007ec72",
			"localstorageenabled": true,
			"name": "at-vie-1",
			"networktype": "Basic",
			"securitygroupsenabled": true,
			"tags": [],
			"zonetoken": "a2a8345d-7daa-3316-8d90-5b8e49706764"
		},
		{
			"allocationstate": "Enabled",
			"dhcpprovider": "VirtualRouter",
			"id": "de88c980-78f6-467c-a431-71bcc88e437f",
			"localstorageenabled": true,
			"name": "de-fra-1",
			"networktype": "Basic",
			"securitygroupsenabled": true,
			"tags": [],
			"zonetoken": "c4bdb9f2-c28d-36a3-bbc5-f91fc69527e6"
		}
	]
}}`}, response{200, `
{"listzonesresponse": {
	"count": 4
}}`})
	defer ts.Close()

	cs := NewClient(ts.URL, "KEY", "SECRET")
	cs.PageSize = 2

	zone := new(Zone)
	zones, err := cs.List(zone)
	if err != nil {
		t.Error(err)
	}

	if len(zones) != 4 {
		t.Errorf("Four zones were expected, got %d", len(zones))
	}
}

func TestListZoneError(t *testing.T) {
	ts := newServer(response{200, `
{"listzonesresponse": {
	"count": 4,
	"zone": [
		{
			"allocationstate": "Enabled",
			"dhcpprovider": "VirtualRouter",
			"id": "1747ef5e-5451-41fd-9f1a-58913bae9702",
			"localstorageenabled": true,
			"name": "ch-gva-2",
			"networktype": "Basic",
			"securitygroupsenabled": true,
			"tags": [],
			"zonetoken": "f9a2983b-42e5-3b12-ae74-0b1f54cd6204"
		},
		{
			"allocationstate": "Enabled",
			"dhcpprovider": "VirtualRouter",
			"id": "381d0a95-ed4a-4ad9-b41c-b97073c1a433",
			"localstorageenabled": true,
			"name": "ch-dk-2",
			"networktype": "Basic",
			"securitygroupsenabled": true,
			"tags": [],
			"zonetoken": "23a24359-121a-38af-a938-e225c97c397b"
		}
	]
}}`}, response{400, `
{"listzonesresponse": {
	"cserrorcode": 9999,
	"errorcode": 431,
	"errortext": "Unable to execute API command listzones",
	"uuidList": []
}}`})
	defer ts.Close()

	cs := NewClient(ts.URL, "KEY", "SECRET")
	cs.PageSize = 2

	zone := new(Zone)
	_, err := cs.List(zone)
	if err == nil {
		t.Error("An error was expected")
	}
}

func TestListZoneTimeout(t *testing.T) {
	ts := newSleepyServer(time.Second, 200, `
{"listzonesresponse": {
	"count": 4
}}`)
	defer ts.Close()

	cs := NewClientWithTimeout(ts.URL, "KEY", "SECRET", time.Millisecond)

	zone := new(Zone)
	_, err := cs.List(zone)
	if err == nil {
		t.Errorf("An error was expected")
	}
}
