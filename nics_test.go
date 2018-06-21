package egoscale

import (
	"testing"
)

func TestAddIPToNic(t *testing.T) {
	req := &AddIPToNic{}
	_ = req.response().(*AsyncJobResult)
	_ = req.asyncResponse().(*NicSecondaryIP)
}

func TestRemoveIPFromNic(t *testing.T) {
	req := &RemoveIPFromNic{}
	_ = req.response().(*AsyncJobResult)
	_ = req.asyncResponse().(*booleanResponse)
}

func TestListNicsAPIName(t *testing.T) {
	req := &ListNics{}
	_ = req.response().(*ListNicsResponse)
}

func TestActivateIP6(t *testing.T) {
	req := &ActivateIP6{}
	_ = req.response().(*AsyncJobResult)
	_ = req.asyncResponse().(*Nic)
}

func TestListNics(t *testing.T) {
	ts := newServer(response{200, jsonContentType, `
{"listnicsresponse": {
	"count": 1,
	"nic": [
		{
			"gateway": "165.150.8.1",
			"id": "fed8fa77-c27d-411c-a630-2ce787161ad6",
			"ipaddress": "165.150.8.20",
			"isdefault": true,
			"macaddress": "06:f0:f8:00:00:57",
			"netmask": "255.255.252.0",
			"networkid": "2fbc4161-da5a-4be3-b5d8-ee34270cb827",
			"virtualmachineid": "d7658121-64c8-4c50-96a7-3bb5ceeca7b2"
		}
	]
}}`})
	defer ts.Close()

	cs := NewClient(ts.URL, "KEY", "SECRET")

	nic := &Nic{
		VirtualMachineID: "d7658121-64c8-4c50-96a7-3bb5ceeca7b2",
	}
	nics, err := cs.List(nic)
	if err != nil {
		t.Error(err)
	}

	if len(nics) != 1 {
		t.Errorf("One nic was expected, got %d", len(nics))
	}

	if !nics[0].(*Nic).IsDefault {
		t.Errorf("Nic should be default")
	}
}

func TestListNicInvalid(t *testing.T) {
	ts := newServer()
	defer ts.Close()

	cs := NewClient(ts.URL, "KEY", "SECRET")

	nic := new(Nic)

	_, err := cs.List(nic)
	if err == nil {
		t.Error("An error was expected")
	}
}

func TestListNicError(t *testing.T) {
	ts := newServer(response{431, jsonContentType, `
{"listnicresponse": {
	"cserrorcode": 9999,
	"errorcode": 431,
	"errortext": "Unable to execute API command listnics due to missing parameter virtualmachineid",
	"uuidList": []
}}`})
	defer ts.Close()

	cs := NewClient(ts.URL, "KEY", "SECRET")

	nic := &Nic{
		VirtualMachineID: "1",
	}

	_, err := cs.List(nic)
	if err == nil {
		t.Error("An error was expected")
	}
}
